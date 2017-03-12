package payload

import jwt "github.com/dgrijalva/jwt-go"

//NerdClaims hold nerdalize specific jwt claims
type NerdClaims struct {
	*jwt.StandardClaims
	ProjectID string `json:"proj,omitempty"`
}
