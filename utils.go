package ibus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime/debug"
)

// CreateResponse creates *Response with given status code and string data
// TODO Why pointer is used here?
func CreateResponse(code int, message string) *Response {
	return &Response{
		StatusCode: code,
		Data:       []byte(message),
	}
}

// CreateErrorResponse creates *Response with given status code, error message and ContentType "plain/text"
// TODO Why pointer is used here?
func CreateErrorResponse(code int, err error) *Response {
	return &Response{
		StatusCode:  code,
		Data:        []byte(err.Error()),
		ContentType: "plain/text",
	}
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
		sectionType: string(sectionTypeBytes),
		path:        path,
		sectionKind: kind,
	}
}

type sectionData struct {
	sectionType string
	path        []string
	sectionKind SectionKind
	elems       []*element
	currentElem int
}

// ToJSON encodes the section to JSON and writes it to `buf`
func (s *sectionData) ToJSON(buf *bytes.Buffer) {
	buf.WriteString("{")
	if len(s.Type()) > 0 {
		buf.WriteString(fmt.Sprintf(`"type":"%s",`, s.Type()))
	}
	if len(s.Path()) > 0 {
		buf.WriteString(`"path":[`)
		for _, path := range s.Path() {
			buf.WriteString(fmt.Sprintf(`"%s",`, path))
		}
		buf.Truncate(buf.Len() - 1)
		buf.WriteString("],")
	}
	if len(s.elems) > 0 {
		buf.WriteString(`"elements":`)
		finalizer := ""
		if s.sectionKind == SectionKindArray {
			buf.WriteString("[")
			finalizer = "]"
		} else if s.sectionKind != SectionKindObject {
			buf.WriteString("{")
			finalizer = "}"
		}
		for _, elem := range s.elems {
			if len(elem.Name) > 0 {
				buf.WriteString(fmt.Sprintf(`"%s":`, elem.Name))
			}
			buf.Write(elem.Value)
			buf.WriteString(",")
		}
		buf.Truncate(buf.Len() - 1)
		buf.WriteString(finalizer)
	}
	buf.WriteString("}")
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
	if len(s.elems) <= s.currentElem {
		return "", nil, false
	}
	name = s.elems[s.currentElem].Name
	value = s.elems[s.currentElem].Value
	ok = true
	s.currentElem++
	return
}

func (s *sectionData) Value() []byte {
	return s.elems[0].Value
}

type element struct {
	Name  string
	Value []byte
}

func sendSection(s *sectionData, sections chan ISection) {
	if s == nil {
		return
	}
	switch s.sectionKind {
	case SectionKindArray:
		sections <- IArraySection(&sectionDataArray{s})
	case SectionKindMap:
		sections <- IMapSection(s)
	}
}

// ResultSenderImpl s.e.
type ResultSenderImpl struct {
	Chunks   chan []byte
	skipName bool
}

// StartArraySection s.e.
func (rsi *ResultSenderImpl) StartArraySection(sectionType string, path []string) {
	rsi.startSection(BusPacketSectionArray, sectionType, path)
	rsi.skipName = true
}

// ObjectSection s.e.
func (rsi *ResultSenderImpl) ObjectSection(sectionType string, path []string, element interface{}) (err error) {
	bytes, err := json.Marshal(element)
	if err != nil {
		return err
	}
	rsi.startSection(BusPacketSectionObject, sectionType, path)
	rsi.Chunks <- bytes
	return nil

}

// StartMapSection s.e.
func (rsi *ResultSenderImpl) StartMapSection(sectionType string, path []string) {
	rsi.startSection(BusPacketSectionMap, sectionType, path)
}

// SendElement s.e.
func (rsi *ResultSenderImpl) SendElement(name string, element interface{}) (err error) {
	bytes, err := json.Marshal(element)
	if err != nil {
		return err
	}
	rsi.Chunks <- []byte{byte(BusPacketElement)}
	if !rsi.skipName {
		rsi.Chunks <- []byte(name)
	}
	rsi.Chunks <- bytes
	return nil
}

func (rsi *ResultSenderImpl) startSection(packetType BusPacketType, sectionType string, path []string) {
	rsi.Chunks <- []byte{byte(packetType)}
	rsi.Chunks <- []byte(sectionType)
	for _, p := range path {
		rsi.Chunks <- []byte(p)
	}
	rsi.Chunks <- []byte{0}
	rsi.skipName = false
}

// BytesToSections converts chan []byte to chan ISection
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
		var currentSection *sectionData
		for chunk := range ch {
			switch BusPacketType(chunk[0]) {
			case BusPacketSectionMap:
				sendSection(currentSection, sections)
				currentSection = readSection(ch, SectionKindMap)
				if currentSection == nil {
					return
				}
			case BusPacketElement:
				nameBytes := []byte{}
				if currentSection.sectionKind != SectionKindArray {
					ok := false
					nameBytes, ok = <-ch
					if !ok {
						return
					}
				}
				valueBytes, ok := <-ch
				if !ok {
					return
				}
				currentSection.elems = append(currentSection.elems, &element{Name: string(nameBytes), Value: valueBytes})
			case BusPacketSectionArray:
				sendSection(currentSection, sections)
				currentSection = readSection(ch, SectionKindArray)
				if currentSection == nil {
					return
				}
			case BusPacketSectionObject:
				sendSection(currentSection, sections)
				objectSection := readSection(ch, SectionKindObject)
				if objectSection == nil {
					return
				}
				valueBytes, ok := <-ch
				if !ok {
					return
				}
				// will send immediately to not to wait for next packet
				objectSection.elems = append(objectSection.elems, &element{Value: valueBytes})
				sections <- objectSection
				currentSection = nil
			default:
				panic("unknown bus packet type: " + string(chunk[0]))
			}
		}
		sendSection(currentSection, sections)
	}()
	return
}
