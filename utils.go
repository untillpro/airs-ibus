/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

// CreateResponse creates *Response with given status code and string data
func CreateResponse(code int, message string) Response {
	return Response{
		StatusCode: code,
		Data:       []byte(message),
	}
}

// CreateErrorResponse creates *Response with given status code, error message and ContentType "plain/text"
func CreateErrorResponse(code int, err error) Response {
	return Response{
		StatusCode:  code,
		Data:        []byte(err.Error()),
		ContentType: "plain/text",
	}
}
