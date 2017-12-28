package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd"
	"github.com/pkg/errors"
)

const (
	//DefaultAWSRegion can be used to set the project region
	DefaultAWSRegion = "eu-west-1"
)

//SessionSnapshot is a snapshot of the session file
type SessionSnapshot struct {
	OAuth   OAuth   `json:"oauth,omitempty"`
	JWT     JWT     `json:"jwt,omitempty"`
	Project Project `json:"project,omitempty"`
}

//RequireProjectID returns the current project name from the session snapshot or error with ErrProjectIDNotSet
func (ss *SessionSnapshot) RequireProjectID() (name string, err error) {
	name = ss.Project.Name
	if ss.OAuth.AccessToken == "" {
		return "", nerd.ErrTokenUnset
	}

	if name == "" {
		return "", nerd.ErrProjectIDNotSet
	}
	return name, nil
}

//OAuth contains oauth credentials
type OAuth struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token"`
	Expiration   time.Time `json:"expiration,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

//JWT contains JWT credentials
type JWT struct {
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

//Project contains details of the current working project.
type Project struct {
	Name      string `json:"name,omitempty"`
	AWSRegion string `json:"aws_region,omitempty"`
}

//Session is an object capable of reading and writing the session file
type Session struct {
	location string
	m        *sync.Mutex
}

//_ is used to make sure Session implements SessionInterface
var _ SessionInterface = &Session{}

//NewSession creates a new Session
func NewSession(loc string) *Session {
	return &Session{
		location: loc,
		m:        &sync.Mutex{},
	}
}

//GetDefaultSessionLocation sets the location to ~/.nerd/session.json
func GetDefaultSessionLocation() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "failed to find home dir")
	}
	return filepath.Join(dir, ".nerd", "session.json"), nil
}

//SessionInterface is the interface of Session
type SessionInterface interface {
	Read() (*SessionSnapshot, error)
	WriteJWT(jwt, refreshToken string) error
	WriteOAuth(accessToken, refreshToken, idToken string, expiration time.Time, scope, tokenType string) error
	WriteProject(project, awsRegion string) error
}

//readFile reads the session file
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

//write writes a SessionSnapshot to the session file
func (s *Session) write(ss *SessionSnapshot) error {
	f, err := os.Create(s.location)
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

//Read returns a snapshot of the session file
func (s *Session) Read() (*SessionSnapshot, error) {
	s.m.Lock()
	defer s.m.Unlock()
	ss, err := s.readFile()
	if err != nil {
		return nil, err
	}
	return ss, nil
}

//WriteJWT writes the jwt object to the session file
func (s *Session) WriteJWT(jwt, refreshToken string) error {
	s.m.Lock()
	defer s.m.Unlock()
	ss, err := s.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.JWT.Token = jwt
	ss.JWT.RefreshToken = refreshToken
	err = s.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}

//WriteOAuth writes the oauth object to the session file
func (s *Session) WriteOAuth(accessToken, refreshToken, idToken string, expiration time.Time, scope, tokenType string) error {
	s.m.Lock()
	defer s.m.Unlock()
	ss, err := s.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.OAuth.AccessToken = accessToken
	ss.OAuth.RefreshToken = refreshToken
	ss.OAuth.IDToken = idToken
	ss.OAuth.Expiration = expiration
	ss.OAuth.Scope = scope
	ss.OAuth.TokenType = tokenType
	err = s.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}

//WriteProject writes the project object to the session file
func (s *Session) WriteProject(name, awsRegion string) error {
	s.m.Lock()
	defer s.m.Unlock()
	ss, err := s.readFile()
	if err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	ss.Project.Name = name
	ss.Project.AWSRegion = awsRegion
	err = s.write(ss)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}
