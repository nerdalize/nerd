package cmd

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
)

const (
	authorizeEndpoint = "o/authorize"
)

//Login command
type Login struct {
	Config string `long:"config-src" default:"oidc" default-mask:"" description:"type of configuration to use (from env, endpoint, or oidc)"`
	*command
}

//LoginFactory returns a factory method for the join command
func LoginFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &Login{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &ConfOpts{}, flags.PassAfterNonOption, "nerd login")

	t, ok := cmd.advancedOpts.(*ConfOpts)
	if !ok {
		return nil
	}
	t.ConfigFile = cmd.setConfig
	t.SessionFile = cmd.setSession
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *Login) Execute(args []string) (err error) {
	if os.Getenv("NERD_ENV") == "staging" {
		cmd.config = conf.StagingDefaults()
	}
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return renderConfigError(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.Logger(),
	})
	randomState := randomString(32)
	doneCh := make(chan response)
	svr := &http.Server{
		Addr: cmd.config.Auth.OAuthLocalServer,
	}
	defer svr.Close()
	go spawnServer(svr, cmd.config.Auth, randomState, doneCh)

	err = open.Run(fmt.Sprintf("http://%s/oauth?state=%s", cmd.config.Auth.OAuthLocalServer, randomState))
	if err != nil {
		cmd.out.Info(fmt.Sprintf("Failed to open browser window, access the login at http://%s/oauth?state=%s", cmd.config.Auth.OAuthLocalServer, randomState))
	}

	oauthResponse := <-doneCh
	if oauthResponse.err != nil {
		return renderConfigError(oauthResponse.err, "failed to do oauth login")
	}

	out, err := authOpsClient.GetOAuthCredentials(oauthResponse.code, cmd.config.Auth.SecureClientID, cmd.config.Auth.SecureClientSecret, fmt.Sprintf("http://%s/oauth/callback", cmd.config.Auth.OAuthLocalServer))
	if err != nil {
		return renderConfigError(err, "failed to get oauth credentials")
	}

	expiration := time.Unix(time.Now().Unix()+int64(out.ExpiresIn), 0)
	err = cmd.session.WriteOAuth(out.AccessToken, out.RefreshToken, out.IDToken, expiration, out.Scope, out.TokenType)
	if err != nil {
		return renderConfigError(err, "failed to write oauth tokens to config")
	}

	client := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             cmd.Logger(),
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.SecureClientID, cmd.config.Auth.SecureClientSecret, cmd.session),
	})
	list, err := client.ListProjects()
	if err != nil {
		return renderServiceError(err, "cannot list projects")
	}

	if len(list.Projects) == 0 {
		cmd.out.Info("Successful login, but you don't have any cluster. Please contact mayday@nerdalize.com.")
		return nil
	}
	var projectSlug string
	// c := populator.Client{
	// 	Secret:       cmd.config.Auth.SecureClientSecret,
	// 	ID:           cmd.config.Auth.SecureClientID,
	// 	IDPIssuerURL: cmd.config.Auth.IDPIssuerURL,
	// }
	for _, project := range list.Projects {
		projectSlug = project.Nk
		// @TO UPDATE
		// err = setProject(&c, cmd.opts.KubeConfig, cmd.opts.Config, project, cmd.outputter.Logger)
		// if err != nil {
		// 	projectSlug = ""
		// 	continue
		// }
		break
	}
	if projectSlug == "" {
		cmd.out.Info("Successful login, but it seems that there is a connection problem to your cluster(s). Please try again, and if the problem persists, contact mayday@nerdalize.com.")
		return nil
	}

	err = cmd.session.WriteProject(projectSlug, conf.DefaultAWSRegion)
	if err != nil {
		return renderConfigError(err, "cannot write session file")
	}

	cmd.out.Info("Successful login. You can now start a job with the `nerd job run` command.")
	return nil
}

type response struct {
	err  error
	code string
}

//spawnServer spawns a local server http server to guide the login flow
func spawnServer(svr *http.Server, cfg conf.AuthConfig, randomState string, doneCh chan response) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		if randomState != r.URL.Query().Get("state") {
			doneCh <- response{
				code: "",
				err:  errors.Errorf("oauth state does not match provided state (expected: %v, actual: %v)", randomState, r.URL.Query().Get("state")),
			}
			return
		}

		w.Header().Set("Location", cfg.OAuthSuccessURL)
		w.WriteHeader(http.StatusFound)

		doneCh <- response{
			code: r.URL.Query().Get("code"),
			err:  nil,
		}
	})

	mux.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {
		base, err := url.Parse(cfg.APIEndpoint)
		if err != nil {
			doneCh <- response{
				code: "",
				err:  errors.Wrapf(err, "failed to parse auth api endpoint (%v)", cfg.APIEndpoint),
			}
			return
		}
		path, err := url.Parse(authorizeEndpoint)
		if err != nil {
			doneCh <- response{
				code: "",
				err:  errors.Wrapf(err, "failed to parse authorize endpoint (%v)", authorizeEndpoint),
			}
			return
		}

		resolved := base.ResolveReference(path)
		q := resolved.Query()
		q.Set("state", r.URL.Query().Get("state"))
		q.Set("client_id", cfg.SecureClientID)
		q.Set("response_type", "code")
		q.Set("scope", "openid email group api")
		q.Set("redirect_uri", fmt.Sprintf("http://%s/oauth/callback", cfg.OAuthLocalServer))
		resolved.RawQuery = q.Encode()

		w.Header().Set("Location", resolved.String())
		w.WriteHeader(http.StatusFound)
	})
	svr.Handler = mux
	err := svr.ListenAndServe()
	if err != nil {
		doneCh <- response{
			code: "",
			err:  errors.Wrap(err, "failed to spawn local server"),
		}
	}
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

//randomString returns a random string of `n` characters
func randomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// Description returns long-form help text
func (cmd *Login) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *Login) Synopsis() string { return "Start a new authorized session." }

// Usage shows usage
func (cmd *Login) Usage() string { return "nerd login [OPTIONS]" }
