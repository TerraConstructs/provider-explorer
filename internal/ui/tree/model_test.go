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
// This file contains test code derived from github.com/anchore/bubbly
// Original source: https://github.com/anchore/bubbly/tree/main/bubbles/tree/model_test.go

package tree

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"

	"github.com/terraconstructs/provider-explorer/internal/ui/tree/testutil"
)

var _ VisibleModel = (*dummyViewer)(nil)

type dummyViewer struct {
	hidden bool
	state  string
}

func (d dummyViewer) IsVisible() bool {
	return !d.hidden
}

func (d dummyViewer) Init() tea.Cmd {
	return nil
}

func (d dummyViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return d, nil
}

func (d dummyViewer) View() string {
	return d.state
}

func TestModel_View(t *testing.T) {

	tests := []struct {
		name       string
		taskGen    func(testing.TB) Model
		iterations int
	}{
		{
			name: "simple case with selection indicators",
			taskGen: func(tb testing.TB) Model {
				subject := NewModel()

				// └─ a
				//    └─ a-a

				require.NoError(t, subject.Add("", "a", dummyViewer{state: "a"}))
				require.NoError(t, subject.Add("a", "a-a", dummyViewer{state: "a-a"}))

				return subject
			},
		},
		{
			name: "sibling branches with selection",
			taskGen: func(tb testing.TB) Model {
				subject := NewModel()

				// ├─ a
				// │  ├─ a-a
				// │  └─ a-b
				// └─ b
				//    └─ b-a

				require.NoError(t, subject.Add("", "a", dummyViewer{state: "a"}))
				require.NoError(t, subject.Add("a", "a-a", dummyViewer{state: "a-a"}))
				require.NoError(t, subject.Add("a", "a-b", dummyViewer{state: "a-b"}))
				require.NoError(t, subject.Add("", "b", dummyViewer{state: "b"}))
				require.NoError(t, subject.Add("b", "b-a", dummyViewer{state: "b-a"}))

				// Select some nodes
				subject.ToggleSelection("a")
				subject.ToggleSelection("b-a")

				return subject
			},
		},
		{
			name: "scrolling test with height limit",
			taskGen: func(tb testing.TB) Model {
				subject := NewModel()
				subject.SetHeight(5) // Limit display height

				// Create a tall tree that requires scrolling
				require.NoError(t, subject.Add("", "root1", dummyViewer{state: "root1"}))
				require.NoError(t, subject.Add("", "root2", dummyViewer{state: "root2"}))
				require.NoError(t, subject.Add("", "root3", dummyViewer{state: "root3"}))
				require.NoError(t, subject.Add("", "root4", dummyViewer{state: "root4"}))
				require.NoError(t, subject.Add("", "root5", dummyViewer{state: "root5"}))
				require.NoError(t, subject.Add("", "root6", dummyViewer{state: "root6"}))
				require.NoError(t, subject.Add("", "root7", dummyViewer{state: "root7"}))

				return subject
			},
		},
		{
			name: "hidden nodes",
			taskGen: func(tb testing.TB) Model {
				subject := NewModel()

				// └─ a
				//    └─ a-a

				require.NoError(t, subject.Add("", "a", dummyViewer{state: "a"}))
				require.NoError(t, subject.Add("a", "a-a", dummyViewer{state: "a-a"})) // shown as a leaf instead of a fork
				require.NoError(t, subject.Add("a", "a-b", dummyViewer{state: "a-b", hidden: true}))
				require.NoError(t, subject.Add("", "b", dummyViewer{state: "b", hidden: true}))
				require.NoError(t, subject.Add("b", "b-a", dummyViewer{state: "b-a"})) // gets pruned entirely

				return subject
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m tea.Model = tt.taskGen(t)
			tsk, ok := m.(Model)
			require.True(t, ok)
			got := testutil.RunModel(t, tsk, tt.iterations, nil)
			t.Log(got)
			snaps.MatchSnapshot(t, got)
		})
	}
}

func TestModel_Selection(t *testing.T) {
	subject := NewModel()
	
	require.NoError(t, subject.Add("", "a", dummyViewer{state: "node-a"}))
	require.NoError(t, subject.Add("", "b", dummyViewer{state: "node-b"}))
	
	// Initially nothing selected
	require.False(t, subject.IsSelected("a"))
	require.False(t, subject.IsSelected("b"))
	require.Empty(t, subject.GetSelectedNodes())
	
	// Select node a
	subject.ToggleSelection("a")
	require.True(t, subject.IsSelected("a"))
	require.False(t, subject.IsSelected("b"))
	
	selected := subject.GetSelectedNodes()
	require.Len(t, selected, 1)
	require.Contains(t, selected, "a")
	
	// Select node b
	subject.ToggleSelection("b")
	require.True(t, subject.IsSelected("a"))
	require.True(t, subject.IsSelected("b"))
	
	selected = subject.GetSelectedNodes()
	require.Len(t, selected, 2)
	require.Contains(t, selected, "a")
	require.Contains(t, selected, "b")
	
	// Deselect node a
	subject.ToggleSelection("a")
	require.False(t, subject.IsSelected("a"))
	require.True(t, subject.IsSelected("b"))
	
	selected = subject.GetSelectedNodes()
	require.Len(t, selected, 1)
	require.Contains(t, selected, "b")
}

func TestModel_Scrolling(t *testing.T) {
	subject := NewModel()
	subject.SetHeight(3) // Very small height to test scrolling
	
	// Add enough nodes to require scrolling
	for i := 0; i < 10; i++ {
		require.NoError(t, subject.Add("", fmt.Sprintf("node%d", i), dummyViewer{state: fmt.Sprintf("Node %d", i)}))
	}
	
	// Test initial view (should show first 3 items)
	initialView := subject.View()
	require.Contains(t, initialView, "Node 0")
	require.Contains(t, initialView, "Node 1") 
	require.Contains(t, initialView, "Node 2")
	require.NotContains(t, initialView, "Node 3")
	
	// Test scrolling down - need to get the updated model
	updatedModel, _ := subject.Update(tea.KeyMsg{Type: tea.KeyDown})
	scrolledSubject := updatedModel.(Model)
	scrolledView := scrolledSubject.View()
	require.NotContains(t, scrolledView, "Node 0")
	require.Contains(t, scrolledView, "Node 1")
	require.Contains(t, scrolledView, "Node 2")
	require.Contains(t, scrolledView, "Node 3")
}