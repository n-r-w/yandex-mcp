package authagent

type tokenRequest struct {
	Version int    `json:"version"`
	Profile string `json:"profile"`
}

type tokenResponse struct {
	Version int    `json:"version"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
