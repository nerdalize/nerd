//Package conf gives the CLI access to the nerd config file. By default this config file is
//~/.nerd/config.json, but the location can be changed using SetLocation().
//
//All read and write operation to the config file should go through the Read() and Write() functions.
//This way we can keep an in-memory representation of the config (in the global conf variable) for fast read.
package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

//location is the file location of the config file.
var location string

//conf is an in-memory representation of the config file.
var conf *Config

//Config is the structure that describes how the config file looks.
type Config struct {
	Auth             AuthConfig          `json:"auth"`
	EnableLogging    bool                `json:"enable_logging"`
	CurrentProject   string              `json:"current_project"`
	AuthTokens       payload.OAuthTokens `json:"auth_tokens"`
	NerdToken        string              `json:"nerd_access_token"`
	NerdRefreshToken string              `json:"nerd_refresh_token"`
	NerdAPIEndpoint  string              `json:"nerd_api_endpoint"`
}

//AuthConfig contains config details with respect to authentication.
type AuthConfig struct {
	APIEndpoint      string `json:"api_endpoint"`
	PublicKey        string `json:"public_key"`
	ClientID         string `json:"client_id"`
	OAuthSuccessUrl  string `json:"oauth_success_url"`
	OAuthLocalserver string `json:"nerd_oauth_localserver"`
}

//Defaults provides the default for when the config file misses certain fields.
func Defaults() *Config {
	return &Config{
		Auth: AuthConfig{
			APIEndpoint:      "http://auth.nerdalize.com",
			OAuthLocalserver: "localhost:9876",
			OAuthSuccessUrl:  "https://cloud.nerdalize.com",
			PublicKey: `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY-----`,
		},
		EnableLogging:   false,
		CurrentProject:  "6de308f4-face-11e6-bc64-92361f002671",
		NerdAPIEndpoint: "https://batch.nerdalize.com/v1",
	}
}

//SetLocation sets the location of the config file.
func SetLocation(file string) error {
	if file == "" {
		return SetDefaultLocation()
	}
	location = file
	return nil
}

//SetDefaultLocation sets the location to ~/.nerd/config.json
func SetDefaultLocation() error {
	dir, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "failed to get home dir")
	}
	location = filepath.Join(dir, ".nerd", "config.json")
	return nil
}

//GetLocation gets the location and sets it to default it is unset.
func GetLocation() (string, error) {
	if location == "" {
		err := SetDefaultLocation()
		if err != nil {
			return "", errors.Wrap(err, "failed to set default config location")
		}
	}
	return location, nil
}

//Read reads the config either from memory or from disk for the first time.
func Read() (*Config, error) {
	if conf != nil {
		return conf, nil
	}
	loc, err := GetLocation()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config location")
	}
	content, err := ioutil.ReadFile(loc)
	if err != nil && os.IsNotExist(err) {
		return Defaults(), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to open config file")
	}
	conf = Defaults()
	err = json.Unmarshal(content, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config file")
	}
	return conf, nil
}

//Write writes the conf variable to disk.
func Write() error {
	loc, err := GetLocation()
	if err != nil {
		return errors.Wrap(err, "failed to get config location")
	}
	c, err := Read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	f, err := os.Create(loc)
	if err != nil {
		return errors.Wrap(err, "failed to create/open config file")
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(c)
	if err != nil {
		return errors.Wrap(err, "failed to encode json")
	}
	return nil
}

//WriteNerdTokens sets the nerd token and calls Write() to write to disk.
func WriteNerdTokens(tokens *payload.OAuthTokens) error {
	c, err := Read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	c.AuthTokens = *tokens
	return Write()
}
