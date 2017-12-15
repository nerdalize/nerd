package command

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
)

const (
	authorizeEndpoint = "o/authorize"
)

//Login command
type Login struct {
	*command
}

//LoginFactory returns a factory method for the join command
func LoginFactory() (cli.Command, error) {
	comm, err := newCommand("nerd login", "Start a new authorized session.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Login{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Login) DoRun(args []string) error {
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return HandleError(errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint))
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.outputter.Logger,
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
		return HandleError(errors.Wrap(err, "Failed to open browser window. Please see github.com/nerdalize/nerd for alternative ways of authenticating."))
	}

	oauthResponse := <-doneCh
	if oauthResponse.err != nil {
		return HandleError(errors.Wrap(oauthResponse.err, "failed to do oauth login"))
	}

	out, err := authOpsClient.GetOAuthCredentials(oauthResponse.code, cmd.config.Auth.ClientID, fmt.Sprintf("http://%s/oauth/callback", cmd.config.Auth.OAuthLocalServer))
	if err != nil {
		return HandleError(errors.Wrap(err, "failed to get oauth credentials"))
	}

	expiration := time.Unix(time.Now().Unix()+int64(out.ExpiresIn), 0)
	err = cmd.session.WriteOAuth(out.AccessToken, out.RefreshToken, expiration, out.Scope, out.TokenType)
	if err != nil {
		return HandleError(errors.Wrap(err, "failed to write oauth tokens to config"))
	}
	cmd.ui.Info("Successful login. You can now select a project using 'nerd project'")
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
		q.Set("client_id", cfg.ClientID)
		q.Set("response_type", "code")
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
