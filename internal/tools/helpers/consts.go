package helpers

//nolint:gochecknoglobals // error patterns must be shared across function calls for performance
var (
	safePrefixes = []string{
		"decode response:",
		"read response body:",
		"parse base url:",
		"create request:",
		"marshal request body:",
		"execute request:",
		"get token:",
	}

	safeContains = []string{
		"unsupported protocol scheme",
		"unprocessable entity",
	}
)
