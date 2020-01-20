/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"context"
	"time"
)

// SendRequest used by router. sends a message to a given queue
// chunks must be readed to the end
var SendRequest func(ctx context.Context, request *Request, callbackData interface{}, timeout time.Duration) (res Response, chunks <-chan []byte)

// RequestHandler used by app
var RequestHandler func(ctx context.Context, sender interface{}, request *Request)

// PostResponse used by app
// Only first answer must have not nil Response
// Empty chunk means end-of-data
// First answer with chunk means next chunks follow
// chunks must be readed by implementation to the end
// chunks must be closed by sender
var PostResponse func(ctx context.Context, sender interface{}, response Response, chunks <-chan []byte)
