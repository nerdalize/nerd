package payload

import jwt "github.com/dgrijalva/jwt-go"

//NerdClaims hold nerdalize specific jwt claims
type NerdClaims struct {
  *jwt.StandardClaims
  ProjectID string `json:"proj,omitempty"`
}

// OAuthTokens represents the OAuth access tokens
type OAuthTokens struct {
  AccessToken  string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  ExpiresIn    string `json:"expires_in"`
  Scope        string `json:"scope,omitempty"`
  TokenType    string `json:"token_type,omitempty"`
}
