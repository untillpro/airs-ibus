package ibus

import (
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

func readSection(ch <-chan []byte, kind SectionKind) *Section {
	sectionTypeBytes := []byte{}
	if kind != SectionKindObject {
		ok := false
		sectionTypeBytes, ok = <-ch
		if !ok {
			return nil
		}
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
	return &Section{
		SectionType: string(sectionTypeBytes),
		Path:        path,
		SectionKind: kind,
	}
}

// BytesToSections s.e.
func BytesToSections(ch <-chan []byte, chunksErr *error) (sections <-chan ISection) {
	return nil
}

// DecodedChan converts chan of []bytes to chan of interface{}
// interface{} could be one of following: []byte, &ibus.Section, &ibus.Element
func DecodedChan(ch <-chan []byte, chunksErr *error) (res chan interface{}, chunksErrRes *error) {
	chunksErrRes = chunksErr
	res = make(chan interface{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				err := fmt.Errorf("panic on channel []byte -> interface{}\n%s", stackTrace)
				*chunksErrRes = err
				for range ch {
				}
			}
			close(res)
		}()
		var currentSection *Section
		for chunk := range ch {
			if len(chunk) > 1 {
				res <- chunk
				continue
			}
			switch BusPacketType(chunk[0]) {
			case BusPacketSectionArray:
				currentSection = readSection(ch, SectionKindArray)
				if currentSection != nil {
					res <- currentSection
				} else {
					return
				}
			case BusPacketSectionMap:
				currentSection = readSection(ch, SectionKindMap)
				if currentSection != nil {
					res <- currentSection
				} else {
					return
				}
			case BusPacketSectionObject:
				currentSection = readSection(ch, SectionKindObject)
				if currentSection != nil {
					res <- currentSection
				} else {
					return
				}
				valueBytes, ok := <-ch
				if !ok {
					return
				}
				el := &Element{
					Name:  "",
					Value: valueBytes,
				}
				res <- el
			case BusPacketElement:
				nameBytes := []byte{}
				if currentSection.SectionKind != SectionKindArray {
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
				el := &Element{
					Name:  string(nameBytes),
					Value: valueBytes,
				}
				res <- el
			default:
				panic("unknown bus packet type: " + string(chunk[0]))
			}
		}
	}()
	return
}
