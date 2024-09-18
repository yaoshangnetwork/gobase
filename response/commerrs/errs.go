package commerrs

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var _ error = (*APIError)(nil)

func (e *APIError) Error() string {
	return e.Message
}
