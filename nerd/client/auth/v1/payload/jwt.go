package v1payload

//GetJWTOutput is output when a JWT is requested
type GetJWTOutput struct {
	Token string `json:"token"`
}

type GetWorkerJWTOutput struct {
	WorkerJWT
}

type RefreshWorkerJWTInput struct {
	WorkerJWT
}

type RefreshWorkerJWTOutput struct {
	WorkerJWT
}

type WorkerJWT struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}
