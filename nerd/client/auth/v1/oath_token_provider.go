package v1auth

//OAuthTokenProvider is capable of providing a oauth access token.  When IsExpired return false
//the in-memory token will be used to prevent from calling Retrieve for each API call.
type OAuthTokenProvider interface {
	IsExpired() bool
	Retrieve() (string, error)
}

//StaticOAuthTokenProvider is a simple oauth token provider that always returns the same token.
type StaticOAuthTokenProvider struct {
	Token string
}

//NewStaticOAuthTokenProvider creates a new StaticOAuthTokenProvider for the given token.
func NewStaticOAuthTokenProvider(token string) *StaticOAuthTokenProvider {
	return &StaticOAuthTokenProvider{token}
}

//IsExpired always returns false.
func (s *StaticOAuthTokenProvider) IsExpired() bool {
	return false
}

//Retrieve always returns the given token.
func (s *StaticOAuthTokenProvider) Retrieve() (string, error) {
	return s.Token, nil
}
