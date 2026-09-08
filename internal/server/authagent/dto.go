package authagent

type tokenRequest struct {
	Profile string `json:"profile"`
}

type tokenResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}
