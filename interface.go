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
// If chunks is not nil chunks must be read to the end
// Non-nil chunksError when chunks are closed means an error in chunks
// `err` and `error` can be a wrapped ErrTimeoutExpired (Checked as errors.Is(err, ErrTimeoutExpired))
var SendRequest func(ctx context.Context,
	request *Request, timeout time.Duration) (res *Response, chunks <-chan []byte, chunksError *error, err error)

// RequestHandler used by app
var RequestHandler func(ctx context.Context, sender interface{}, request Request)

// SendResponse used by app
var SendResponse func(ctx context.Context, sender interface{}, response Response)

// SendParallelResponse s.e.
// If chunks is not nil they must be readed by implementation to the end
// Chunks must be closed by sender
// Non-nil chunksError when chunks are closed means an error in chunks
var SendParallelResponse func(ctx context.Context, sender interface{}, chunks <-chan []byte, chunksError *error)

// MetricSerialRequestCnt s.e.
var MetricSerialRequestCnt uint64

// MetricSerialRequestDurNs s.e.
var MetricSerialRequestDurNs uint64
