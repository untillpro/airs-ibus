/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package ibus

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicUsage(t *testing.T) {
	chunks := make(chan []byte)
	rs := NewResultSender(chunks)
	var chunksErr error
	go func() {
		rs.StartMapSection("secMap", []string{"classifier", "2"})
		rs.SendElement("id1", map[string]interface{}{
			"fld1": "fld1Val",
		})
		rs.SendElement("id2", map[string]interface{}{
			"fld2": "fld2Val",
		})
		rs.StartArraySection("secArr", []string{"classifier", "4"})
		rs.SendElement("", "arrEl1")
		rs.SendElement("", "arrEl2")
		rs.ObjectSection("secObj", []string{"meta"}, map[string]interface{}{
			"total": 1,
		})
		rs.StartMapSection("deps", []string{"classifier", "3"})
		rs.SendElement("id3", map[string]interface{}{
			"fld3": "fld3Val",
		})
		rs.SendElement("id4", map[string]interface{}{
			"fld4": "fld4Val",
		})
		close(chunks)
	}()

	sections := BytesToSections(chunks, &chunksErr)
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
	secObj := section.(IObjectSection)
	require.Equal(t, "secObj", secObj.Type())
	require.Equal(t, []string{"meta"}, secObj.Path())
	valMap = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(secObj.Value(), &valMap))
	require.Equal(t, map[string]interface{}{"total": float64(1)}, valMap)

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
