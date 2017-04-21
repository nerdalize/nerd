package provider

import (
  "crypto/ecdsa"
  "fmt"
  "html"
  "log"
  "net/http"
  "net/url"
  // "time"

  "github.com/Sirupsen/logrus"
  "github.com/nerdalize/nerd/nerd/client"
  "github.com/nerdalize/nerd/nerd/client/credentials"
  "github.com/nerdalize/nerd/nerd/conf"
  // "github.com/nerdalize/nerd/nerd/payload"
  "github.com/pkg/errors"
  "github.com/skratchdot/open-golang/open"
)

//OAuthAPI provides nerdalize credentials by opening a browser to login to the nerdalize Auth API
type OAuthAPI struct {
  *Basis

  Client *client.AuthAPIClient
  //UserPassProvider is a function that provides the OAuthAPI provider with a username and password. This could for example be a function
  //that reads from stdin.
  UserPassProvider func() (string, string, error)
}

//NewOAuthAPI creates a new OAuthAPI provider.
func NewOAuthAPI(c *client.AuthAPIClient) *OAuthAPI {
  return &OAuthAPI{
    Basis: &Basis{
      ExpireWindow: DefaultExpireWindow,
    },
    Client: c,
  }
}

// RetrieveWithoutKey retrieves the access_tokens without a pkey
func (p *OAuthAPI) RetrieveWithoutKey() (*credentials.NerdAPIValue, error) {
  config, err := conf.Read()
  if err != nil {
    return nil, errors.Wrap(err, "failed to read nerd config file")
  }

  serverChan, err := spawnLocalServer(config)

  if err != nil {
    return nil, errors.Wrap(err, "failed to spawn local server")
  }

  logrus.Info("Opening browser...")

  // TODO: Generate a random key to send back and forth
  errOpen := open.Run("http://localhost:9876/oauth?state=bla") // TODO: Make configurable
  if errOpen != nil {
    panic(errOpen)
  }

  oauthResponse := <-serverChan

  if oauthResponse.err != nil {
    return nil, errors.Wrap(err, "failed to login through oauth")
  }
  logrus.Info("Done here...")

  // Fetch the OAuth tokens.

  tokens, err := p.Client.GetOAuthToken(oauthResponse.code)
  if err != nil {
    return nil, errors.Wrap(err, "failed to fetch tokens")
  }
  err = conf.WriteNerdTokens(tokens)
  if err != nil {
    return nil, errors.Wrap(err, "failed to write nerd token to config")
  }

  return &credentials.NerdAPIValue{
    NerdToken: tokens.AccessToken,
  }, nil
}

//Retrieve retrieves the token from the authentication server.
func (p *OAuthAPI) Retrieve(_ *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
  return p.RetrieveWithoutKey()
}

type response struct {
  err  error
  code string
}

// spawnLocalServer returns a http server to guide the login flow
func spawnLocalServer(config *conf.Config) (chan response, error) {
  serverDone := make(chan response)

  logrus.WithFields(logrus.Fields{
    "apiUrl": config.Auth.APIEndpoint,
  }).Info("Spawned local server")

  go func() {
    http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
      // Should verify received state here. r.URL.Query().Get("state") === stored_state

      fmt.Fprintf(w, "Hello, %q %q", html.EscapeString(r.URL.Path), r.URL.RawQuery)
      serverDone <- response{
        code: r.URL.Query().Get("code"),
        err:  nil,
      }
    })

    http.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {

      u, err := url.Parse(config.Auth.APIEndpoint)
      if err != nil {
        log.Fatal(err)
      }
      u.Path = "/o/authorize"
      q := u.Query()
      q.Set("state", r.URL.Query().Get("state"))
      q.Set("client_id", config.Auth.ClientID)
      q.Set("response_type", "code")
      q.Set("redirect_uri", "http://"+config.Auth.OAuthLocalserver+"/oauth/callback") // TODO: Make configurable
      u.RawQuery = q.Encode()

      w.Header().Set("Location", u.String())
      w.WriteHeader(http.StatusFound)
    })

    log.Fatal(http.ListenAndServe(":9876", nil))
  }()

  return serverDone, nil
}
