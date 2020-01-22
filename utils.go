package ibus

func CreateResponse(code int, message string) *Response {
	return &Response{
		StatusCode: code,
		Data:       []byte(message),
	}
}
