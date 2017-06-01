package command

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/jwt"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
	"github.com/restic/chunker"
)

type stdoutkw struct{}

//Write writes a key to stdout.
func (kw *stdoutkw) Write(k string) (err error) {
	// _, err = fmt.Fprintf(os.Stdout, "%v\n", k)
	logrus.Info(k)
	return nil
}

//NewClient creates a new batch Client.
func NewClient(c *conf.Config, session *conf.Session, outputter *format.Outputter) (*v1batch.Client, error) {
	key, err := jwt.ParseECDSAPublicKeyFromPemBytes([]byte(c.Auth.PublicKey))
	if err != nil {
		return nil, errors.Wrap(err, "ECDSA Public Key is invalid")
	}
	base, err := url.Parse(c.NerdAPIEndpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "nerd endpoint '%v' is not a valid URL", c.NerdAPIEndpoint)
	}
	authbase, err := url.Parse(c.Auth.APIEndpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", c.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: outputter.Logger,
	})
	authTokenClient := v1auth.NewTokenClient(v1auth.TokenClientConfig{
		Base:   authbase,
		Logger: outputter.Logger,
	})
	return v1batch.NewClient(v1batch.ClientConfig{
		JWTProvider: v1batch.NewChainedJWTProvider(
			jwt.NewEnvProvider(key, session, authTokenClient),
			jwt.NewConfigProvider(key, session, authTokenClient),
			jwt.NewAuthAPIProvider(key, session, v1auth.NewClient(v1auth.ClientConfig{
				Base:               authbase,
				Logger:             outputter.Logger,
				OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, c.Auth.ClientID, session),
			})),
		),
		Base:   base,
		Logger: outputter.Logger,
	}), nil
}

//UserPassProvider prompts the username and password on stdin.
func UserPassProvider(ui cli.Ui) func() (string, string, error) {
	return func() (string, string, error) {
		ui.Info("Please enter your Nerdalize username and password.")
		user, err := ui.Ask("Username: ")
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read username")
		}
		pass, err := ui.AskSecret("Password: ")
		if err != nil {
			return "", "", errors.Wrap(err, "failed to read password")
		}
		return user, pass, nil
	}
}

//ErrorCauser returns the error that is one level up in the error chain.
func ErrorCauser(err error) error {
	type causer interface {
		Cause() error
	}

	if err2, ok := err.(causer); ok {
		err = err2.Cause()
	}
	return err
}

//batchErr returns a human-readble error message for batch HTTPErrors
func batchErr(err *v1batch.HTTPError) error {
	switch err.StatusCode {
	case http.StatusUnprocessableEntity:
		if len(err.Err.Fields) > 0 {
			return fmt.Errorf("Validation error: %v", err.Err.Fields)
		}
	case http.StatusNotFound:
		return fmt.Errorf("The specified resource does not exist")
	}
	return fmt.Errorf("unknown server error (%v)", err.StatusCode)
}

//HandleError handles the way errors are presented to the user.
func HandleError(err error) error {
	if errors.Cause(err) == oauth.ErrTokenRevoked {
		return fmt.Errorf("Your login session has expired. Please login using 'nerd login'")
	}
	if errors.Cause(err) == oauth.ErrTokenUnset {
		return fmt.Errorf("You are not logged in. Please login using 'nerd login'")
	}
	if herr, ok := errors.Cause(err).(*v1batch.HTTPError); ok {
		return batchErr(herr)
	}
	return err
}

//ProgressBar creates a new CLI progess bar and adds input from the progressCh to the bar.
func ProgressBar(w io.Writer, total int64, progressCh <-chan int64, doneCh chan<- struct{}) {
	bar := pb.New64(total)
	bar.SetUnits(pb.U_BYTES)
	bar.Output = w
	bar.Start()
	for elem := range progressCh {
		bar.Add64(elem)
	}
	bar.Finish()
	doneCh <- struct{}{}
}

//Chunker is a wrapper of the restic/chunker library, to make it compatible with the v1data.Chunker interface.
type Chunker struct {
	cr *chunker.Chunker
}

//NewChunker returns a new Chunker
func NewChunker(pol uint64, r io.Reader) *Chunker {
	return &Chunker{
		cr: chunker.New(r, chunker.Pol(pol)),
	}
}

//Next wraps the restic/chunker Next call.
func (c *Chunker) Next() (data []byte, length uint, err error) {
	buf := make([]byte, chunker.MaxSize)
	chunk, err := c.cr.Next(buf)
	if err != nil {
		return []byte{}, 0, err
	}
	return chunk.Data, chunk.Length, nil
}
