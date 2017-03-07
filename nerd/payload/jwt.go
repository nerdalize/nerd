package payload

import jwt "github.com/dgrijalva/jwt-go"

type NerdClaims struct {
	*jwt.StandardClaims
	ProjectID string `json:"proj,omitempty"`
}
