/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"context"
	"time"
)

// PostRequest s.e.
// Implementation should call ResponseHandler when Response arrives or timeout happens
var PostRequest func(ctx context.Context, request *Request, callbackData interface{}, timeout time.Duration)

// ResponseHandler s.e.
// nil response means timeout
var ResponseHandler func(ctx context.Context, response *Response, callbackData interface{})

// RequestHandler s.e.
var RequestHandler func(ctx context.Context, sender interface{}, request *Request)

// PostResponse s.e.
// response or chunk may be
var PostResponse func(ctx context.Context, sender interface{}, response *Response, chunk []byte)
