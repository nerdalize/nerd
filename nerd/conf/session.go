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

const (
	//DefaultAWSRegion can be used to set the project region
	DefaultAWSRegion = "eu-west-1"
)

//Config is the structure that describes how the config file looks.
type SessionSnapshot struct {
	OAuth   OAuth   `json:"oauth,omitempty"`
	JWT     JWT     `json:"jwt,omitempty"`
	Project Project `json:"project,omitempty"`
}

//OAuthConfig contians oauth credentials
type OAuth struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiration   time.Time `json:"expiration,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

//JWTConfig contains JWT credentials
type JWT struct {
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

//ProjectConfig contains details of the current working project.
type Project struct {
	Name      string `json:"name"`
	AWSRegion string `json:"aws_region"`
}

type Session struct {
	location string
	m        *sync.Mutex
}

var _ SessionInterface = &Session{}

func NewSession(loc string) *Session {
	return &Session{
		location: loc,
		m:        &sync.Mutex{},
	}
}

//GetDefaultLocation sets the location to ~/.nerd/session.json
func GetDefaultSessionLocation() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "failed to find home dir")
	}
	return filepath.Join(dir, ".nerd", "session.json"), nil
}

type SessionInterface interface {
	Read() (*SessionSnapshot, error)
	WriteJWT(jwt, refreshToken string) error
	WriteOAuth(accessToken, refreshToken string, expiration time.Time, scope, tokenType string) error
	WriteProject(project, awsRegion string) error
}

func (s *Session) readFile() (*SessionSnapshot, error) {
	ss := &SessionSnapshot{}
	content, err := ioutil.ReadFile(s.location)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open config file")
	}
	err = json.Unmarshal(content, ss)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config file")
	}
	return ss, nil
}

func (c *Session) write(ss *SessionSnapshot) error {
	f, err := os.Create(c.location)
	if err != nil {
		return errors.Wrap(err, "failed to create/open config file")
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(ss)
	if err != nil {
		return errors.Wrap(err, "failed to encode json")
	}
	return nil
}

func (c *Session) Read() (*SessionSnapshot, error) {
	c.m.Lock()
	defer c.m.Unlock()
	ss, err := c.readFile()
	if err != nil {
		return nil, err
	}
	return ss, nil
}

func (c *Session) WriteJWT(jwt, refreshToken string) error {
	c.m.Lock()
	defer c.m.Unlock()
	ss, err := c.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.JWT.Token = jwt
	ss.JWT.RefreshToken = refreshToken
	err = c.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}

func (c *Session) WriteOAuth(accessToken, refreshToken string, expiration time.Time, scope, tokenType string) error {
	c.m.Lock()
	defer c.m.Unlock()
	ss, err := c.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.OAuth.AccessToken = accessToken
	ss.OAuth.RefreshToken = refreshToken
	ss.OAuth.Expiration = expiration
	ss.OAuth.Scope = scope
	ss.OAuth.TokenType = tokenType
	err = c.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}

func (c *Session) WriteProject(name, awsRegion string) error {
	c.m.Lock()
	defer c.m.Unlock()
	ss, err := c.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.Project.Name = name
	ss.Project.AWSRegion = awsRegion
	err = c.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}
