/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"context"
	"time"
)

// SendRequest used by router and app, sends a message to a given queue
// If err is not nil res and chunks are nil
// If err is nil and chunks is not nil chunks must be read to the end
// Non-nil chunksError when chunks are closed means an error in chunks
// `err` and `error` can be a wrapped ErrTimeoutExpired (Checked as errors.Is(err, ErrTimeoutExpired))
var SendRequest func(ctx context.Context,
	request *Request, timeout time.Duration) (res *Response, err error, chunks <-chan []byte, chunksError *error)

// RequestHandler used by app
var RequestHandler func(ctx context.Context, sender interface{}, request Request)

// SendResponse used by app
// If chunks is not nil they must be readed by implementation to the end
// Chunks must be closed by sender
// Non-nil chunksError when chunks are closed means an error in chunks
var SendResponse func(ctx context.Context, sender interface{}, response Response, chunks <-chan []byte, chunksError *error)
