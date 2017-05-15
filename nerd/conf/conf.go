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
	"sync"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

//location is the file location of the config file.
var location string

//conf is an in-memory representation of the config file.
var conf *Config

var mut *sync.Mutex

func init() {
	mut = &sync.Mutex{}
}

//Config is the structure that describes how the config file looks.
type Config struct {
	Auth            AuthConfig           `json:"auth"`
	Credentials     CredentialsConfig    `json:"credentials"`
	EnableLogging   bool                 `json:"enable_logging"`
	CurrentProject  CurrentProjectConfig `json:"current_project"`
	NerdAPIEndpoint string               `json:"nerd_api_endpoint"`
}

//AuthConfig contains config details with respect to the authentication server.
type AuthConfig struct {
	APIEndpoint      string `json:"api_endpoint"`
	PublicKey        string `json:"public_key"`
	ClientID         string `json:"client_id"`
	OAuthSuccessURL  string `json:"oauth_success_url"`
	OAuthLocalServer string `json:"nerd_oauth_localserver"`
}

//CredentialsConfig contains oauth and jwt credentials
type CredentialsConfig struct {
	OAuth OAuthConfig `json:"oauth,omitempty"`
	JWT   JWTConfig   `json:"jwt,omitempty"`
}

//OAuthConfig contians oauth credentials
type OAuthConfig struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiration   time.Time `json:"expiration"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
}

//JWTConfig contains JWT credentials
type JWTConfig struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

//CurrentProjectConfig contains details of the current working project.
type CurrentProjectConfig struct {
	Name      string `json:"current_project"`
	AWSRegion string `json:"aws_region"`
}

//Defaults provides the default for when the config file misses certain fields.
func Defaults() *Config {
	return &Config{
		Auth: AuthConfig{
			APIEndpoint:      "http://auth.nerdalize.com",
			OAuthLocalServer: "localhost:9876",
			OAuthSuccessURL:  "https://cloud.nerdalize.com",
			ClientID:         "GuoeRJLYOXzVa9ydPjKi83lCctWtXpNHuiy46Yux",
			PublicKey: `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY-----`,
		},
		EnableLogging: false,
		CurrentProject: CurrentProjectConfig{
			Name:      "projectx",
			AWSRegion: "eu-west-1",
		},
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

//read reads the config either from memory or from disk for the first time.
func read() (*Config, error) {
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

//write writes the conf variable to disk.
func write() error {
	loc, err := GetLocation()
	if err != nil {
		return errors.Wrap(err, "failed to get config location")
	}
	c, err := read()
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

//Read reads the config file
func Read() (*Config, error) {
	mut.Lock()
	defer mut.Unlock()
	return read()
}

//WriteJWT writes the JWT to the config file
func WriteJWT(jwt, refreshToken string) error {
	mut.Lock()
	defer mut.Unlock()
	c, err := read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	c.Credentials.JWT.Token = jwt
	c.Credentials.JWT.RefreshToken = refreshToken
	return write()
}

//WriteOAuth writes oauth credentials to the config file
func WriteOAuth(accessToken, refreshToken string, expiration time.Time, scope, tokenType string) error {
	mut.Lock()
	defer mut.Unlock()
	c, err := read()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	c.Credentials.OAuth.AccessToken = accessToken
	c.Credentials.OAuth.RefreshToken = refreshToken
	c.Credentials.OAuth.Expiration = expiration
	c.Credentials.OAuth.Scope = scope
	c.Credentials.OAuth.TokenType = tokenType
	return write()
}
