package v1payload

//GetOAuthCredentialsInput is input for getting oauth credentials
type GetOAuthCredentialsInput struct {
	Code        string `url:"code"`
	ClientID    string `url:"client_id"`
	RedirectURI string `url:"redirect_uri"`
	GrantType   string `url:"grant_type"`
}

//GetOAuthCredentialsOutput is output when getting oauth credentials
type GetOAuthCredentialsOutput struct {
	OAuthCredentials
}

//RefreshOAuthCredentialsInput is input for refreshing oauth credentials
type RefreshOAuthCredentialsInput struct {
	RefreshToken string `url:"refresh_token"`
	ClientID     string `url:"client_id"`
	GrantType    string `url:"grant_type"`
}

//RefreshOAuthCredentialsOutput is output when refreshing oauth credentials
type RefreshOAuthCredentialsOutput struct {
	OAuthCredentials
}

// OAuthCredentials represents the OAuth access tokens
type OAuthCredentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}
