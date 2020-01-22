package ibus

// CreateResponse creates *Response with given status code and string data
func CreateResponse(code int, message string) *Response {
	return &Response{
		StatusCode: code,
		Data:       []byte(message),
	}
}