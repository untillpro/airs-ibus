package ibus

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
)

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

// NewResultSender creates IResultSender instance
func NewResultSender(chunks chan []byte) IResultSender {
	return &ResultSender{chunks: chunks}
}

func readSection(ch <-chan []byte, kind SectionKind, prevSection *sectionData) (nextSection *sectionData) {
	closeSection(prevSection)
	sectionTypeBytes, ok := <-ch
	if !ok {
		return nil
	}
	pathElem, ok := <-ch
	if !ok {
		return nil
	}
	path := []string{}
	for ok && pathElem[0] != 0 {
		path = append(path, string(pathElem))
		pathElem, ok = <-ch
	}
	return &sectionData{
		sectionType: string(sectionTypeBytes),
		path:        path,
		sectionKind: kind,
		elems:       make(chan *element),
	}
}

type sectionData struct {
	sectionType string
	path        []string
	sectionKind SectionKind
	elems       chan *element
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

type sectionDataMap struct {
	*sectionData
}

type sectionDataObject struct {
	*sectionData
	valueGot bool
}

func (s *sectionDataArray) Next() (value []byte, ok bool) {
	elem, ok := <-s.elems
	if !ok {
		return
	}
	return elem.value, true
}

func (s *sectionDataMap) Next() (name string, value []byte, ok bool) {
	elem, ok := <-s.elems
	if !ok {
		return
	}
	return elem.name, elem.value, true
}

func (s *sectionDataObject) Value() []byte {
	if s.valueGot {
		return nil
	}
	elem, ok := <-s.elems
	if !ok {
		return nil
	}
	s.valueGot = true
	return elem.value
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
	bytes, err := rsi.marshalElem(element)
	if err != nil {
		return err
	}
	rsi.startSection(BusPacketSectionObject, sectionType, path)
	rsi.sendElementBytes("", bytes)
	return nil

}

// StartMapSection s.e.
func (rsi *ResultSender) StartMapSection(sectionType string, path []string) {
	rsi.startSection(BusPacketSectionMap, sectionType, path)
}

// SendElement s.e.
// if element is []byte then send it as is. Note: if the element will be read by airs-router then JSON malformation is possible. Caller must take care of this.
func (rsi *ResultSender) SendElement(name string, element interface{}) (err error) {
	bytes, err := rsi.marshalElem(element)
	if err != nil {
		return err
	}
	rsi.sendElementBytes(name, bytes)
	return nil
}

func (rsi *ResultSender) sendElementBytes(name string, elemBytes []byte) {
	rsi.chunks <- []byte{byte(BusPacketElement)}
	if !rsi.skipName {
		rsi.chunks <- []byte(name)
	}
	rsi.chunks <- elemBytes
}

func (rsi *ResultSender) marshalElem(element interface{}) ([]byte, error) {
	if bytesRaw, ok := element.([]byte); ok {
		return bytesRaw, nil
	}
	bytesJSON, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	return bytesJSON, nil
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
// MapSection or ArraySection received -> caller must call Next() until !ok even if elements are not needed
func BytesToSections(ch <-chan []byte, chunksErr *error) (sections chan ISection) {
	log.Println("BytesToSections start ")
	sections = make(chan ISection)
	go func() {
		var currentSection *sectionData
		counter := 0
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				err := fmt.Errorf("panic on channel []byte -> interface{}:%s\n%s", r, stackTrace)
				*chunksErr = err
				for range ch {
				}
			}
			closeSection(currentSection)
			close(sections)
			_, ok := <-ch
			log.Println("BytesToSections counter " + strconv.Itoa(counter) + ", ok:" + strconv.FormatBool(ok))
		}()
		ok := false
		for chunk := range ch {
			if len(chunk) == 0 {
				continue
			}
			switch BusPacketType(chunk[0]) {
			case BusPacketSectionMap:
				if currentSection = readSection(ch, SectionKindMap, currentSection); currentSection == nil {
					log.Println("BytesToSections 1")
					return
				}
				sections <- &sectionDataMap{currentSection}
			case BusPacketElement:
				nameBytes := []byte{}
				if currentSection.sectionKind != SectionKindArray {
					if nameBytes, ok = <-ch; !ok {
						log.Println("BytesToSections 2")
						return
					}
					log.Println("!!!BytesToSections elem:" + string(nameBytes))
				}
				valueBytes, ok := <-ch
				if !ok {
					log.Println("BytesToSections 3")
					return
				}
				if currentSection != nil {
					currentSection.elems <- &element{string(nameBytes), valueBytes}
				}
			case BusPacketSectionArray:
				if currentSection = readSection(ch, SectionKindArray, currentSection); currentSection == nil {
					log.Println("BytesToSections 4")
					return
				}
				sections <- &sectionDataArray{currentSection}
			case BusPacketSectionObject:
				if currentSection = readSection(ch, SectionKindObject, currentSection); currentSection == nil {
					log.Println("BytesToSections 5")
					return
				}
				sections <- &sectionDataObject{currentSection, false}
			default:
				panic("unexpected bus packet type: " + string(chunk[0]))
			}
			counter++
		}
	}()
	return
}

func closeSection(sec *sectionData) {
	if sec != nil {
		close(sec.elems)
	}
}
