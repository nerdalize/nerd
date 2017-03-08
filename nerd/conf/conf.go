package conf

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var location string
var conf *Config

type Config struct {
	Auth            AuthConfig `json:"auth"`
	CurrentProject  string     `json:"current_project"`
	NerdToken       string     `json:"nerd_token"`
	NerdAPIEndpoint string     `json:"nerd_api_endpoint"`
}

type AuthConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	PublicKey   string `json:"public_key"`
}

func Defaults() *Config {
	return &Config{
		Auth: AuthConfig{
			APIEndpoint: "https://auth.nerdalize.com",
			PublicKey: `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY-----`,
		},
		CurrentProject:  "6de308f4-face-11e6-bc64-92361f002671",
		NerdAPIEndpoint: "https://platform.nerdalize.net",
	}
}

func SetLocation(file string) error {
	if file == "" {
		return SetDefaultLocation()
	}
	location = file
	return nil
}

func SetDefaultLocation() error {
	dir, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "failed to get home dir")
	}
	location = filepath.Join(dir, ".nerd", "config.json")
	return nil
}

func GetLocation() (string, error) {
	if location == "" {
		err := SetDefaultLocation()
		if err != nil {
			return "", errors.Wrap(err, "failed to set default config location")
		}
	}
	return location, nil
}

func Read() (*Config, error) {
	if conf != nil {
		return conf, nil
	}
	loc, err := GetLocation()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config location")
	}
	content, err := ioutil.ReadFile(loc)
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

func Write() error {
	loc, err := GetLocation()
	if err != nil {
		return errors.Wrap(err, "failed to get config location")
	}
	c, err := Read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	data, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "failed to encode json")
	}
	err = ioutil.WriteFile(loc, data, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write to config file")
	}
	return nil
}

func WriteNerdToken(token string) error {
	c, err := Read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	c.NerdToken = token
	return Write()
}
