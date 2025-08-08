// Copyright (c) 2024 Anchore, Inc.
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file contains code derived from github.com/anchore/bubbly
// Original source: https://github.com/anchore/bubbly/tree/main/bubbles/internal/testutil/run_model.go

package testutil

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
)

func RunModel(_ testing.TB, m tea.Model, iterations int, message tea.Msg) string {
	if iterations == 0 {
		iterations = 1
	}
	m.Init()
	var cmd tea.Cmd = func() tea.Msg {
		return message
	}

	for i := 0; cmd != nil && i < iterations; i++ {
		msgs := flatten(cmd())
		var nextCmds []tea.Cmd
		var next tea.Cmd
		for _, msg := range msgs {
			fmt.Printf("Message: %+v %+v\n", reflect.TypeOf(msg), msg)
			m, next = m.Update(msg)
			nextCmds = append(nextCmds, next)
		}
		cmd = tea.Batch(nextCmds...)
	}
	return m.View()
}

func flatten(p tea.Msg) (msgs []tea.Msg) {
	if p == nil {
		return nil
	}
	if reflect.TypeOf(p).Name() == "batchMsg" {
		partials := extractBatchMessages(p)
		for _, m := range partials {
			msgs = append(msgs, flatten(m)...)
		}
	} else {
		msgs = []tea.Msg{p}
	}
	return msgs
}

func extractBatchMessages(m tea.Msg) (ret []tea.Msg) {
	sliceMsgType := reflect.SliceOf(reflect.TypeOf(tea.Cmd(nil)))
	value := reflect.ValueOf(m) // note: this is technically unaddressable

	// make our own instance that is addressable
	valueCopy := reflect.New(value.Type()).Elem()
	valueCopy.Set(value)

	cmds := reflect.NewAt(sliceMsgType, unsafe.Pointer(valueCopy.UnsafeAddr())).Elem()
	for i := 0; i < cmds.Len(); i++ {
		item := cmds.Index(i)
		r := item.Call(nil)
		ret = append(ret, r[0].Interface().(tea.Msg))
	}
	return ret
}