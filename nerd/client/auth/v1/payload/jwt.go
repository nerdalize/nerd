package v1payload

//GetJWTOutput is output when a JWT is requested
type GetJWTOutput struct {
	Token string `json:"token"`
}
