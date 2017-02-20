package credentials

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
)

// ProviderName is the name of the credentials provider.
// TODO: Rename JWT to NerdToken
const JWTHomeLocation = `~/.nerd/jwt`
const NerdTokenPermissions = 0644

// Provider satisfies the credentials.Provider interface, and is a client to
// retrieve credentials from an arbitrary endpoint.
type NerdAPIProvider struct {
	// The date/time when to expire on
	expiration time.Time

	// If set will be used by IsExpired to determine the current time.
	// Defaults to time.Now if CurrentTime is not set.  Available for testing
	// to be able to mock out the current time.
	CurrentTime func() time.Time

	// TODO: include Client field to mock it for testing
}

// NewCredentialsClient returns a Credentials wrapper for retrieving credentials
// from an arbitrary endpoint concurrently. The client will request the
// func NewNerdalizeCredentials() *credentials.Credentials {
// 	return credentials.NewCredentials(&Provider{staticCreds: false})
// }

// IsExpired returns true if the credentials retrieved are expired, or not yet
// retrieved.
func (p *NerdAPIProvider) IsExpired() bool {
	if p.CurrentTime == nil {
		p.CurrentTime = time.Now
	}
	return p.expiration.Before(p.CurrentTime())
}

func promptUserPass() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your Nerdalize username and password.")
	fmt.Print("Username: ")
	user, err := reader.ReadString('\n')
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read username")
	}
	fmt.Print("Password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read password")
	}
	return strings.Replace(user, "\n", "", 1), string(pass), nil
}

// TODO: Move this to separate client?
func getNerdToken(user, pass string) (string, error) {
	type body struct {
		token string `json:"token"`
	}
	b := &body{}
	url := "thatacher_url"
	s := sling.New().Get(url)
	req, err := s.Request()
	if err != nil {
		return "", errors.Wrapf(err, "failed to create request to fetch nerd token (%v)", url)
	}
	req.SetBasicAuth(user, pass)
	_, err = s.Do(req, b, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to request nerd token (%v)", url)
	}
	if b.token == "" {
		return "", errors.Errorf("failed to read nerd token from response body (%v)", url)
	}
	return b.token, nil
}

func saveNerdToken(token string) error {
	err := ioutil.WriteFile(JWTHomeLocation, []byte(token), NerdTokenPermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to write nerd token to '%v'", JWTHomeLocation)
	}
	return nil
}

// Retrieve will attempt to request the credentials from the endpoint the Provider
// was configured for. And error will be returned if the retrieval fails.
func (p *NerdAPIProvider) Retrieve() (*NerdClaims, error) {
	user, pass, err := promptUserPass()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get username or password")
	}
	token, err := getNerdToken(user, pass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nerd token for username and password")
	}
	err = saveNerdToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save nerd token")
	}
	claims, err := DecodeToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retreive claims from nerd token")
	}
	p.expiration = time.Unix(claims.ExpiresAt, 0)
	return claims, nil
}
