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
// If chunks is not nil then Response must be ignored
// @Deprecated
var SendRequest func(ctx context.Context,
	request *Request, timeout time.Duration) (res *Response, chunks <-chan []byte, chunksError *error, err error)

// SendRequest2 used by router and app, sends a message to a given queue
// err is not nil -> NATS-related error occured before or during reading the first response packet. Sections and secError are nil, res must be ignored
// sections not nil ->
//  - res must be ignored
//  - sections must be read to the end
//  - non-nil secError when sections are closed means NATS-related error during reading sections or an error came with IResultSenderCloseable.Close()
// sections is nil -> res and err only should be used as the result. secError is nil
// `err` and `*secError` can be a wrapped ErrTimeoutExpired (Checked as errors.Is(err, ErrTimeoutExpired))
var SendRequest2 func(ctx context.Context,
	request Request, timeout time.Duration) (res Response, sections <-chan ISection, secError *error, err error)

// RequestHandler used by app
var RequestHandler func(ctx context.Context, sender interface{}, request Request)

// SendResponse used by app
var SendResponse func(ctx context.Context, sender interface{}, response Response)

// SendParallelResponse s.e.
// @Deprecated
// If chunks is not nil they must be readed by implementation to the end
// Chunks must be closed by sender
// response is valid when chunks finishes or nil
// chunkError is set by implementation when it could not send chunk
var SendParallelResponse func(ctx context.Context, sender interface{}, chunks <-chan []byte, chunkError *error)

// SendParallelResponse2 ???
// Result of Close
var SendParallelResponse2 func(ctx context.Context, sender interface{}) (rsender IResultSenderClosable)

// MetricSerialRequestCnt s.e.
// @Depecated
var MetricSerialRequestCnt uint64

// MetricSerialRequestDurNs s.e.
// @Depecated
var MetricSerialRequestDurNs uint64

// MetricCntSerialRequest  number of serial requests
var MetricCntSerialRequest func(ctx context.Context) uint64

// MetricDurSerialRequest duration of serial requests
var MetricDurSerialRequest func(ctx context.Context) uint64
