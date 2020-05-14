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

// ISection s.e.
type ISection interface {
	Type() string
}

// IDataSection s.e.
type IDataSection interface {
	ISection
	Path() []string
}

// IObjectSection s.e.
type IObjectSection interface {
	IDataSection
	// Caller MUST call Value() even if it does not need the value
	// note: second and further Value() calls will return nil
	Value() []byte
}

// IArraySection s.e.
// Caller MUST call Next() until !ok
type IArraySection interface {
	IDataSection
	Next() (value []byte, ok bool)
}

// IMapSection s.e.
// Caller MUST call Next() until !ok
type IMapSection interface {
	IDataSection
	Next() (name string, value []byte, ok bool)
}

// IResultSender used by ParallelFunction
type IResultSender interface {
	// Must be called before first Send*
	// Can be called multiple times - each time new section started
	// Section path may NOT include data from database, only constants should be used
	StartArraySection(sectionType string, path []string)
	StartMapSection(sectionType string, path []string)
	ObjectSection(sectionType string, path []string, element interface{}) (err error)

	// For reading journal
	// StartBinarySection(sectionType string, path []string)

	// element should be "marshallable" by json package
	// Send* can be called multiple times per array
	// name is ignored for Array section
	// For reading journal
	// if element is []byte then it will be sent sent as is. Note: JSON malformation is possible for airs-router's http client. Caller must take care of this.
	SendElement(name string, element interface{}) (err error)
}
