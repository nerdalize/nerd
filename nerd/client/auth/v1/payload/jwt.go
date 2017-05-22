package v1payload

//GetJWTOutput is output when a JWT is requested
type GetJWTOutput struct {
	Token string `json:"token"`
}

type GetWorkerJWTOutput struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type RefreshWorkerJWTInput struct {
	Token  string `json:"jwt"`
	Secret string `json:"secret"`
}

type RefreshWorkerJWTOutput struct {
	Token string `json:"token"`
}
