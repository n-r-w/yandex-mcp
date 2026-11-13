package authagent

type tokenRequest struct {
	TokenType string `json:"token_type"`
	Profile   string `json:"profile"`
}

type tokenResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}
