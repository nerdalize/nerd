package client

import (
  "fmt"
  "html"
  "log"
  "net/http"
  "net/url"

  "github.com/Sirupsen/logrus"
  "github.com/nerdalize/nerd/nerd/conf"
  "github.com/pkg/errors"
  "github.com/skratchdot/open-golang/open"
)

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
      q.Set("redirect_uri", "http://localhost:9876/oauth/callback")
      u.RawQuery = q.Encode()

      w.Header().Set("Location", u.String())
      w.WriteHeader(http.StatusFound)
    })

    log.Fatal(http.ListenAndServe(":9876", nil))
  }()

  return serverDone, nil
}

// OAuthLogin Spawns a server, and tries to login the user via a browser -> Returns the oauth code to fetch the token with.
func OAuthLogin() (string, error) {

  logrus.SetFormatter(&logrus.JSONFormatter{})
  config, err := conf.Read()
  if err != nil {
    return "", errors.Wrap(err, "failed to read nerd config file")
  }

  logrus.Info("Opening browser for authentication...")
  serverChan, err := spawnLocalServer(config)

  if err != nil {
    return "", errors.Wrap(err, "failed to spawn local server")
  }

  logrus.Info("Opening browser...")

  errOpen := open.Run("http://localhost:9876/oauth?state=bla")
  if errOpen != nil {
    panic(errOpen)
  }

  logrus.Info("Done here...")
  response := <-serverChan
  return response.code, response.err
}
