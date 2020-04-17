/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicUsage(t *testing.T) {
	chunks := make(chan []byte)
	rsi := &ResultSenderImpl{Chunks: chunks}
	var chunksErr error
	go func() {
		rsi.StartMapSection("secMap", []string{"classifier", "2"})
		rsi.SendElement("id1", map[string]interface{}{
			"fld1": "fld1Val",
		})
		rsi.SendElement("id2", map[string]interface{}{
			"fld2": "fld2Val",
		})
		rsi.StartArraySection("secArr", []string{"classifier", "4"})
		rsi.SendElement("", "arrEl1")
		rsi.SendElement("", "arrEl2")
		rsi.ObjectSection("secObj", []string{"meta"}, map[string]interface{}{
			"total": 1,
		})
		rsi.StartMapSection("deps", []string{"classifier", "3"})
		rsi.SendElement("id3", map[string]interface{}{
			"fld3": "fld3Val",
		})
		rsi.SendElement("id4", map[string]interface{}{
			"fld4": "fld4Val",
		})
		close(rsi.Chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)
	buf := bytes.NewBufferString("")
	WithTimeout(func() {
		section := <-sections
		secMap := section.(IMapSection)
		require.Equal(t, "secMap", secMap.Type())
		require.Equal(t, []string{"classifier", "2"}, secMap.Path())
		name, value, ok := secMap.Next()
		require.True(t, ok)
		require.Equal(t, "id1", name)
		valMap := map[string]interface{}{}
		require.Nil(t, json.Unmarshal(value, &valMap))
		require.Equal(t, map[string]interface{}{"fld1": "fld1Val"}, valMap)
		name, value, ok = secMap.Next()
		require.True(t, ok)
		require.Equal(t, "id2", name)
		valMap = map[string]interface{}{}
		require.Nil(t, json.Unmarshal(value, &valMap))
		require.Equal(t, map[string]interface{}{"fld2": "fld2Val"}, valMap)
		name, value, ok = secMap.Next()
		require.False(t, ok)
		require.Empty(t, name)
		require.Nil(t, value)
		section.ToJSON(buf)
		require.Equal(t, `{"type":"secMap","path":["classifier","2"],"elements":{"id1":{"fld1":"fld1Val"},"id2":{"fld2":"fld2Val"}}}`, string(buf.Bytes()))
		buf.Reset()

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
		section.ToJSON(buf)
		require.Equal(t, `{"type":"secArr","path":["classifier","4"],"elements":["arrEl1","arrEl2"]}`, string(buf.Bytes()))
		buf.Reset()

		section = <-sections
		secObj := section.(IObjectSection)
		require.Equal(t, "secObj", secObj.Type())
		require.Equal(t, []string{"meta"}, secObj.Path())
		valMap = map[string]interface{}{}
		require.Nil(t, json.Unmarshal(secObj.Value(), &valMap))
		require.Equal(t, map[string]interface{}{"total": float64(1)}, valMap)
		section.ToJSON(buf)
		require.Equal(t, `{"type":"secObj","path":["meta"],"elements":{"total":1}}`, string(buf.Bytes()))
		buf.Reset()

		section = <-sections
		secMap = section.(IMapSection)
		require.Equal(t, "deps", secMap.Type())
		require.Equal(t, []string{"classifier", "3"}, secMap.Path())
		name, value, ok = secMap.Next()
		require.True(t, ok)
		require.Equal(t, "id3", name)
		valMap = map[string]interface{}{}
		require.Nil(t, json.Unmarshal(value, &valMap))
		require.Equal(t, map[string]interface{}{"fld3": "fld3Val"}, valMap)
		name, value, ok = secMap.Next()
		require.True(t, ok)
		require.Equal(t, "id4", name)
		valMap = map[string]interface{}{}
		require.Nil(t, json.Unmarshal(value, &valMap))
		require.Equal(t, map[string]interface{}{"fld4": "fld4Val"}, valMap)
		name, value, ok = secMap.Next()
		require.False(t, ok)
		require.Empty(t, name)
		require.Nil(t, value)
		section.ToJSON(buf)
		require.Equal(t, `{"type":"deps","path":["classifier","3"],"elements":{"id3":{"fld3":"fld3Val"},"id4":{"fld4":"fld4Val"}}}`, string(buf.Bytes()))
		buf.Reset()

		_, ok = <-sections
		require.False(t, ok)
	}, 1000)
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
	WithTimeout(func() {
		_, ok := <-sections
		require.False(t, ok)
	}, 1000)
	require.NotNil(t, chunksErr)
}

func WithTimeout(f func(), timeoutMS int64) bool {
	timeout := time.After(time.Duration(timeoutMS) * time.Millisecond)
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-timeout:
		return false
	case <-done:
		return true
	}
}
