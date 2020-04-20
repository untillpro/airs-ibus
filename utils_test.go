/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	expected1 = map[string]interface{}{
		"fld1": "fld1Val",
	}
	expected2 = map[string]interface{}{
		"fld2": "fld2Val",
	}
	expected3 = map[string]interface{}{
		"fld3": "fld3Val",
	}
	expected4 = map[string]interface{}{
		"fld4": "fld4Val",
	}
	expectedTotal = map[string]interface{}{
		"total": float64(1),
	}
)

func TestBasicUsage(t *testing.T) {
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	var chunksErr error
	go func() {
		rs.ObjectSection("secObj", []string{"meta"}, expectedTotal)
		rs.StartMapSection("secMap", []string{"classifier", "2"})
		rs.SendElement("id1", expected1)
		rs.SendElement("id2", expected2)
		chunks <- []byte{} // should be skipped
		rs.StartArraySection("secArr", []string{"classifier", "4"})
		rs.SendElement("", "arrEl1")
		rs.SendElement("", "arrEl2")
		rs.StartMapSection("deps", []string{"classifier", "3"})
		rs.SendElement("id3", expected3)
		rs.SendElement("id4", expected4)
		close(chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)

	section := <-sections
	secObj := section.(IObjectSection)
	require.Equal(t, "secObj", secObj.Type())
	require.Equal(t, []string{"meta"}, secObj.Path())
	valMap := map[string]interface{}{}
	require.Nil(t, json.Unmarshal(secObj.Value(), &valMap))
	require.Equal(t, expectedTotal, valMap)

	section = <-sections
	secMap := section.(IMapSection)
	require.Equal(t, "secMap", secMap.Type())
	require.Equal(t, []string{"classifier", "2"}, secMap.Path())
	name, value, ok := secMap.Next()
	require.True(t, ok)
	require.Equal(t, "id1", name)
	valMap = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(value, &valMap))
	require.Equal(t, expected1, valMap)
	name, value, ok = secMap.Next()
	require.True(t, ok)
	require.Equal(t, "id2", name)
	valMap = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(value, &valMap))
	require.Equal(t, expected2, valMap)
	name, value, ok = secMap.Next()
	require.False(t, ok)
	require.Empty(t, name)
	require.Nil(t, value)

	section = <-sections
	secArr := section.(IArraySection)
	require.Equal(t, "secArr", secArr.Type())
	require.Equal(t, []string{"classifier", "4"}, secArr.Path())
	value, ok = secArr.Next()
	require.True(t, ok)
	val := ""
	require.Nil(t, json.Unmarshal(value, &val))
	require.Equal(t, "arrEl1", val)
	value, ok = secArr.Next()
	require.True(t, ok)
	val = ""
	require.Nil(t, json.Unmarshal(value, &val))
	require.Equal(t, "arrEl2", val)
	value, ok = secArr.Next()
	require.False(t, ok)
	require.Nil(t, value)

	section = <-sections
	secMap = section.(IMapSection)
	require.Equal(t, "deps", secMap.Type())
	require.Equal(t, []string{"classifier", "3"}, secMap.Path())
	name, value, ok = secMap.Next()
	require.True(t, ok)
	require.Equal(t, "id3", name)
	valMap = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(value, &valMap))
	require.Equal(t, expected3, valMap)
	name, value, ok = secMap.Next()
	require.True(t, ok)
	require.Equal(t, "id4", name)
	valMap = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(value, &valMap))
	require.Equal(t, expected4, valMap)
	name, value, ok = secMap.Next()
	require.False(t, ok)
	require.Empty(t, name)
	require.Nil(t, value)

	_, ok = <-sections
	require.False(t, ok)
}

func TestPanicOnConvertToISections(t *testing.T) {
	chunks := make(chan []byte)
	var chunksErr error
	go func() {
		chunks <- []byte{255} // unknown bus packet type
		chunks <- []byte{0}   // to test read out the channel
		close(chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)
	_, ok := <-sections
	require.False(t, ok)
	require.NotNil(t, chunksErr)
}

func TestMapElementRawBytes(t *testing.T) {
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	var chunksErr error
	go func() {
		elementJSONBytes, err := json.Marshal(&expected3)
		require.Nil(t, err)
		rs.StartMapSection("deps", []string{"classifier", "3"})
		rs.SendElement("id3", elementJSONBytes)
		close(chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)

	section := <-sections
	secMap := section.(IMapSection)
	require.Equal(t, "deps", secMap.Type())
	require.Equal(t, []string{"classifier", "3"}, secMap.Path())
	name, value, ok := secMap.Next()
	require.True(t, ok)
	require.Equal(t, "id3", name)
	valMap := map[string]interface{}{}
	require.Nil(t, json.Unmarshal(value, &valMap))
	require.Equal(t, expected3, valMap)
	name, value, ok = secMap.Next()
	require.False(t, ok)
	require.Empty(t, name)
	require.Nil(t, value)

	_, ok = <-sections
	require.False(t, ok)
}

func TestObjectElementRawBytes(t *testing.T) {
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	var chunksErr error
	go func() {
		valJSONBytes, err := json.Marshal(&expectedTotal)
		require.Nil(t, err)
		rs.ObjectSection("secObj", []string{"meta"}, valJSONBytes)
		close(chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)

	section := <-sections
	secObj := section.(IObjectSection)
	require.Equal(t, "secObj", secObj.Type())
	require.Equal(t, []string{"meta"}, secObj.Path())
	valMap := map[string]interface{}{}
	require.Nil(t, json.Unmarshal(secObj.Value(), &valMap))
	require.Equal(t, expectedTotal, valMap)

	_, ok := <-sections
	require.False(t, ok)
}

func TestElementsErrors(t *testing.T) {
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	require.NotNil(t, rs.SendElement("", func() {}))
	require.NotNil(t, rs.ObjectSection("", nil, func() {}))
}

func TestCreateResponse(t *testing.T) {
	r := CreateResponse(1, "test")
	require.Equal(t, 1, r.StatusCode)
	require.Equal(t, "test", string(r.Data))
	require.Empty(t, r.ContentType)
}

func TestCreateErrorResponse(t *testing.T) {
	r := CreateErrorResponse(1, errors.New("test"))
	require.Equal(t, 1, r.StatusCode)
	require.Equal(t, "test", string(r.Data))
	require.Equal(t, "plain/text", r.ContentType)
}

func TestStopOnChannelCloseOnElement(t *testing.T) {
	var chunksErr error

	// close on element name
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	go func() {
		rs.StartMapSection("1", []string{"2"})
		chunks <- []byte{byte(BusPacketElement)}
		close(chunks)
	}()
	sections := BytesToSections(chunks, &chunksErr)
	testStopOnChannelCloseOnElement(t, sections)

	// close on element value
	chunks = make(chan []byte)
	rs = NewResultSender(chunks)
	go func() {
		rs.StartMapSection("1", []string{"2"})
		chunks <- []byte{byte(BusPacketElement)}
		chunks <- []byte("element name")
		close(chunks)
	}()
	sections = BytesToSections(chunks, &chunksErr)
	testStopOnChannelCloseOnElement(t, sections)
}

func TestStopOnChannelCloseOnObjectValue(t *testing.T) {
	var chunksErr error

	// close on element name
	chunks := make(chan []byte)
	go func() {
		chunks <- []byte{byte(BusPacketSectionObject)}
		chunks <- []byte("1")
		chunks <- []byte{0}
		close(chunks)
	}()
	sections := BytesToSections(chunks, &chunksErr)

	_, ok := <-sections
	require.False(t, ok)

}

func testStopOnChannelCloseOnElement(t *testing.T, sections chan ISection) {
	sec := <-sections
	secMap := sec.(IMapSection)
	require.Equal(t, "1", secMap.Type())
	require.Equal(t, []string{"2"}, secMap.Path())
	_, _, ok := secMap.Next()
	require.False(t, ok)

	_, ok = <-sections
	require.False(t, ok)
}

func TestStopOnChannelClose(t *testing.T) {
	testStopOnChanneClose(t, BusPacketSectionMap)
	testStopOnChanneClose(t, BusPacketSectionArray)
	testStopOnChanneClose(t, BusPacketSectionObject)
}

func testStopOnChanneClose(t *testing.T, bpt BusPacketType) {
	var chunksErr error
	chunks := make(chan []byte)
	go func() {
		chunks <- []byte{byte(bpt)}
		close(chunks)
	}()
	sections := BytesToSections(chunks, &chunksErr)

	_, ok := <-sections
	require.False(t, ok)

	chunks = make(chan []byte)
	go func() {
		chunks <- []byte{byte(bpt)}
		chunks <- []byte("secMap")
		close(chunks)
	}()
	sections = BytesToSections(chunks, &chunksErr)

	_, ok = <-sections
	require.False(t, ok)
}
