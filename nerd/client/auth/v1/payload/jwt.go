package v1payload

//GetJWTOutput is output when a JWT is requested
type GetJWTOutput struct {
	Token string `json:"token"`
}

//GetWorkerJWTOutput is output when a worker JWT (JWT + RefreshToken) is requested
type GetWorkerJWTOutput struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

//RefreshWorkerJWTInput is input for refreshing a JWT
type RefreshWorkerJWTInput struct {
	Token  string `json:"jwt"`
	Secret string `json:"secret"`
}

//RefreshWorkerJWTOutput is output when a JWT refresh is requested
type RefreshWorkerJWTOutput struct {
	Token string `json:"token"`
}
