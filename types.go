/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

// HTTPMethod s.e.
// see const.go/HTTPMethodGET...
type HTTPMethod int

// Request s.e.
type Request struct {
	Method HTTPMethod

	QueueID string

	// Always 0 for non-party queues
	WSID int64
	// Calculated from PartitionDividend, 0 for non-party queues
	PartitionNumber int

	Header map[string][]string

	// Part of URL which follows: queue alias in non-party queues, part dividend in partitioned queues
	Resource string

	// Part of URL which follows ? (URL.Query())
	Query map[string][]string

	// Content of http.Request JSON-parsed Body
	Body []byte

	// attachment-name => attachment-id
	// Must be non-null
	Attachments map[string]string
}

// Response s.e.
type Response struct {
	ContentType string
	StatusCode  int
	Data        []byte
}

// BusPacketType s.e.
type BusPacketType int

// SectionKind int
type SectionKind int

const (
	// BusPacketSectionMap s.e.
	BusPacketSectionMap BusPacketType = iota
	// BusPacketElement s.e.
	BusPacketElement
	// BusPacketSectionArray s.e.
	BusPacketSectionArray
	// BusPacketSectionObject s.e.
	BusPacketSectionObject
)

const (
	// SectionKindUnspecified s.e.
	SectionKindUnspecified SectionKind = iota
	// SectionKindMap s.e.
	SectionKindMap
	// SectionKindArray s.e.
	SectionKindArray
	// SectionKindObject s.e.
	SectionKindObject
)

// Section s.e.
type Section struct {
	SectionType string
	Path        []string
	SectionKind SectionKind
}

// Element s.e.
type Element struct {
	Name  string
	Value []byte
}
