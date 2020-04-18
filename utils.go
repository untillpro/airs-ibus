package ibus

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
)

// CreateResponse creates *Response with given status code and string data
// TODO Why pointer is used here?
func CreateResponse(code int, message string) Response {
	return Response{
		StatusCode: code,
		Data:       []byte(message),
	}
}

// CreateErrorResponse creates *Response with given status code, error message and ContentType "plain/text"
// TODO Why pointer is used here?
func CreateErrorResponse(code int, err error) Response {
	return Response{
		StatusCode:  code,
		Data:        []byte(err.Error()),
		ContentType: "plain/text",
	}
}

// NewResultSender creates IResultSender instance
func NewResultSender(chunks chan []byte) IResultSender {
	return &ResultSender{chunks: chunks}
}

func readSection(ch <-chan []byte, kind SectionKind) *sectionData {
	sectionTypeBytes := []byte{}
	ok := false
	sectionTypeBytes, ok = <-ch
	if !ok {
		return nil
	}
	pathElem, ok := <-ch
	if !ok {
		return nil
	}
	path := []string{}
	for pathElem[0] != 0 {
		path = append(path, string(pathElem))
		pathElem, ok = <-ch
		if !ok {
			break
		}
	}
	return &sectionData{
		sectionType:      string(sectionTypeBytes),
		path:             path,
		sectionKind:      kind,
		ch:               ch,
		nextPacketTypeCh: make(chan BusPacketType),
	}
}

type sectionData struct {
	sectionType      string
	path             []string
	sectionKind      SectionKind
	nextPacketTypeCh chan BusPacketType
	ch               <-chan []byte
	objValue         []byte
}

func (s *sectionData) Type() string {
	return s.sectionType
}

func (s *sectionData) Path() []string {
	return s.path
}

type sectionDataArray struct {
	*sectionData
}

func (s *sectionDataArray) Next() (value []byte, ok bool) {
	_, value, ok = s.sectionData.Next()
	return
}

func (s *sectionData) Next() (name string, value []byte, ok bool) {
	chunk, okCh := <-s.ch
	if !okCh {
		s.nextPacketTypeCh <- BusPacketSectionUnspecified
		return
	}
	if BusPacketType(chunk[0]) != BusPacketElement {
		s.nextPacketTypeCh <- BusPacketType(chunk[0])
		return
	}
	nameBytes := []byte{}
	if s.sectionKind != SectionKindArray {
		nameBytes, okCh = <-s.ch
		if !okCh {
			s.nextPacketTypeCh <- BusPacketSectionUnspecified
			return
		}
	}
	valueBytes, okCh := <-s.ch
	if !okCh {
		s.nextPacketTypeCh <- BusPacketSectionUnspecified
		return
	}
	return string(nameBytes), valueBytes, true
}

func (s *sectionData) Value() []byte {
	return s.objValue
}

type element struct {
	name  string
	value []byte
}

// ResultSender s.e.
type ResultSender struct {
	chunks   chan []byte
	skipName bool
}

// StartArraySection s.e.
func (rsi *ResultSender) StartArraySection(sectionType string, path []string) {
	rsi.startSection(BusPacketSectionArray, sectionType, path)
	rsi.skipName = true
}

// ObjectSection s.e.
func (rsi *ResultSender) ObjectSection(sectionType string, path []string, element interface{}) (err error) {
	bytes, err := json.Marshal(element)
	if err != nil {
		return err
	}
	rsi.startSection(BusPacketSectionObject, sectionType, path)
	rsi.chunks <- bytes
	return nil

}

// StartMapSection s.e.
func (rsi *ResultSender) StartMapSection(sectionType string, path []string) {
	rsi.startSection(BusPacketSectionMap, sectionType, path)
}

// SendElement s.e.
// if element is []byte then send it as is
func (rsi *ResultSender) SendElement(name string, element interface{}) (err error) {
	bytes, err := json.Marshal(element)
	if err != nil {
		return err
	}
	rsi.chunks <- []byte{byte(BusPacketElement)}
	if !rsi.skipName {
		rsi.chunks <- []byte(name)
	}
	rsi.chunks <- bytes
	return nil
}

func (rsi *ResultSender) startSection(packetType BusPacketType, sectionType string, path []string) {
	rsi.chunks <- []byte{byte(packetType)}
	rsi.chunks <- []byte(sectionType)
	for _, p := range path {
		rsi.chunks <- []byte(p)
	}
	rsi.chunks <- []byte{0}
	rsi.skipName = false
}

// BytesToSections converts chan []byte to chan ISection
// Caller should not process chan ISection by >1 goroutine (Elements and Sections will be mixed up)
// Cons of not to collect elements:
// - caller side should not process chan ISection by >1 goroutine
// - MapSection or ArraySection received -> caller must call Next() until !ok even caller does not need elements
func BytesToSections(ch <-chan []byte, chunksErr *error) (sections chan ISection) {
	sections = make(chan ISection)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				err := fmt.Errorf("panic on channel []byte -> interface{}\n%s", stackTrace)
				*chunksErr = err
				for range ch {
				}
			}
			close(sections)
		}()
		nextBusPacketType := BusPacketSectionUnspecified
		for {
			if nextBusPacketType == BusPacketSectionUnspecified {
				chunk, ok := <-ch
				if !ok {
					break
				}
				nextBusPacketType = BusPacketType(chunk[0])
			}
			nextBusPacketType = processChunk(ch, nextBusPacketType, sections)
		}
	}()
	return
}

func processChunk(ch <-chan []byte, bpt BusPacketType, sections chan ISection) (nextPacketType BusPacketType) {
	switch bpt {
	case BusPacketSectionMap:
		sec := readSection(ch, SectionKindMap)
		if sec == nil {
			return
		}
		sections <- sec
		return <-sec.nextPacketTypeCh
	case BusPacketSectionArray:
		sec := readSection(ch, SectionKindArray)
		if sec == nil {
			return
		}
		sections <- &sectionDataArray{sec}
		return <-sec.nextPacketTypeCh
	case BusPacketSectionObject:
		sec := readSection(ch, SectionKindObject)
		if sec == nil {
			return
		}
		ok := false
		sec.objValue, ok = <-ch
		if !ok {
			return
		}
		sections <- sec
	default:
		panic("unepected bus packet type: " + string(bpt))
	}
	return
}
