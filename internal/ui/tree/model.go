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
// Original source: https://github.com/anchore/bubbly/tree/main/bubbles/tree/model.go

package tree

import (
    "errors"
    "strings"
    "sync"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/scylladb/go-set/strset"
)

var _ tea.Model = (*Model)(nil)

type Model struct {
	roots    []string
	nodes    map[string]VisibleModel
	children map[string][]string
	parents  map[string]string
	lock     *sync.RWMutex

	// Selection and scrolling support (extensions)
    selected    map[string]bool // tracks which nodes are selected
	viewport    int             // current scroll position
	height      int             // available display height
	totalHeight int             // total content height
    cursor      int             // current cursor position (visible line index)

	// formatting options
	Margin                    string
	Indent                    string
	Fork                      string
	Branch                    string
	Leaf                      string
	Padding                   string
	VerticalPadMultilineNodes bool
	RootsWithoutPrefix        bool
}

func NewModel() Model {
	return Model{
		nodes:    make(map[string]VisibleModel),
		children: make(map[string][]string),
		parents:  make(map[string]string),
		selected: make(map[string]bool), // extension
		lock:     &sync.RWMutex{},

		// formatting options
		Margin:                    "",
		Indent:                    "   ",
		Branch:                    "│  ",
		Fork:                      "├──",
		Leaf:                      "└──",
		Padding:                   "",
		VerticalPadMultilineNodes: false,
		RootsWithoutPrefix:        false,
	}
}

// SetHeight sets the available display height for scrolling (extension)
func (m *Model) SetHeight(height int) {
	m.height = height
}

// ToggleSelection toggles the selection state of a node (extension)
func (m *Model) ToggleSelection(id string) {
    m.lock.Lock()
    defer m.lock.Unlock()
    m.selected[id] = !m.selected[id]
}

// SetSelection explicitly sets the selection state of a node (extension)
func (m *Model) SetSelection(id string, selected bool) {
    m.lock.Lock()
    defer m.lock.Unlock()
    m.selected[id] = selected
}

// IsSelected returns whether a node is selected (extension)
func (m Model) IsSelected(id string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.selected[id]
}

// GetSelectedNodes returns all selected node IDs (extension)
func (m Model) GetSelectedNodes() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	var selected []string
	for id, isSelected := range m.selected {
		if isSelected {
			selected = append(selected, id)
		}
	}
	return selected
}

// GetVisibleNodes returns all visible node IDs in order (extension)
func (m Model) GetVisibleNodes() []string {
    m.lock.RLock()
    defer m.lock.RUnlock()
    return m.getVisibleNodesUnsafe()
}

// collectVisibleNodes recursively collects visible node IDs
func (m Model) collectVisibleNodes(id string, observed *strset.Set, result *[]string) {
	if observed.Has(id) {
		return
	}
	observed.Add(id)
	
	node := m.nodes[id]
	if node.IsVisible() {
		*result = append(*result, id)
		
		// Add children
		for _, childID := range m.children[id] {
			m.collectVisibleNodes(childID, observed, result)
		}
	}
}

// GetCurrentNode returns the node ID at the cursor position (extension)
func (m Model) GetCurrentNode() string {
    visibleNodes := m.GetVisibleNodes()
    if m.cursor >= 0 && m.cursor < len(visibleNodes) {
        return visibleNodes[m.cursor]
    }
    return ""
}

// getVisibleNodesUnsafe returns visible nodes without acquiring locks.
// Callers must ensure appropriate synchronization.
func (m Model) getVisibleNodesUnsafe() []string {
    var visibleNodes []string
    observed := strset.New()
    for _, id := range m.roots {
        m.collectVisibleNodes(id, observed, &visibleNodes)
    }
    return visibleNodes
}

// SelectAll selects all visible nodes (extension)
func (m *Model) SelectAll() {
    m.lock.Lock()
    defer m.lock.Unlock()
    // Avoid deadlock: compute visible nodes without taking another lock
    visibleNodes := m.getVisibleNodesUnsafe()
    for _, nodeID := range visibleNodes {
        m.selected[nodeID] = true
    }
}

// ClearSelection clears all selected nodes (extension)
func (m *Model) ClearSelection() {
	m.lock.Lock()
	defer m.lock.Unlock()
	
	m.selected = make(map[string]bool)
}

func (m *Model) Add(parent string, id string, model VisibleModel) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if id == "" {
		return errors.New("id cannot be empty")
	}

	m.nodes[id] = model
	if parent != "" {
		m.children[parent] = append(m.children[parent], id)
		m.parents[id] = parent
	} else {
		m.roots = append(m.roots, id)
	}

	return nil
}

func (m *Model) Remove(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.nodes, id)
	delete(m.children, id)
	delete(m.parents, id)
	delete(m.selected, id) // extension: clean up selection state
	
	for _, children := range m.children {
		for i, child := range children {
			if child == id {
				m.children[child] = append(children[:i], children[i+1:]...)
			}
		}
	}

	for i, node := range m.roots {
		if node == id {
			m.roots = append(m.roots[:i], m.roots[i+1:]...)
		}
	}
}

func (m Model) Init() tea.Cmd {
    return nil
}

// MoveDown moves the cursor/viewport down by one line
func (m *Model) MoveDown() {
    visibleNodes := m.getVisibleNodesUnsafe()
    if len(visibleNodes) == 0 {
        return
    }
    // Move cursor first
    if m.cursor < len(visibleNodes)-1 {
        m.cursor++
    }
    // Scroll only to keep cursor in view
    if m.height > 0 && m.cursor >= m.viewport+m.height {
        m.viewport = m.cursor - m.height + 1
    }
}

// MoveUp moves the cursor/viewport up by one line
func (m *Model) MoveUp() {
    // Move cursor first
    if m.cursor > 0 {
        m.cursor--
    }
    // Scroll only to keep cursor in view
    if m.height > 0 && m.cursor < m.viewport {
        m.viewport = m.cursor
    }
}

// MovePageDown moves the cursor down by one page and scrolls to keep it visible
func (m *Model) MovePageDown() {
    visible := m.getVisibleNodesUnsafe()
    if len(visible) == 0 || m.height <= 0 {
        // Fallback to single-line move if no paging context
        m.MoveDown()
        return
    }
    target := m.cursor + m.height
    if target > len(visible)-1 {
        target = len(visible) - 1
    }
    m.cursor = target
    // Place cursor at bottom of viewport
    if m.cursor >= m.viewport+m.height {
        m.viewport = m.cursor - m.height + 1
    }
}

// MovePageUp moves the cursor up by one page and scrolls to keep it visible
func (m *Model) MovePageUp() {
    visible := m.getVisibleNodesUnsafe()
    if len(visible) == 0 || m.height <= 0 {
        // Fallback to single-line move if no paging context
        m.MoveUp()
        return
    }
    target := m.cursor - m.height
    if target < 0 {
        target = 0
    }
    m.cursor = target
    // Ensure cursor is at top of viewport
    if m.cursor < m.viewport {
        m.viewport = m.cursor
    }
}

// Clamp ensures viewport and cursor are within valid bounds relative to
// the current visible nodes and height. It keeps the cursor visible.
func (m *Model) Clamp() {
    visible := m.getVisibleNodesUnsafe()
    n := len(visible)
    if n == 0 {
        m.viewport = 0
        m.cursor = 0
        return
    }
    if m.height < 0 {
        m.height = 0
    }
    if m.cursor < 0 {
        m.cursor = 0
    }
    if m.cursor > n-1 {
        m.cursor = n - 1
    }
    // viewport must be in [0, maxStart]
    maxStart := n - m.height
    if maxStart < 0 {
        maxStart = 0
    }
    if m.viewport < 0 {
        m.viewport = 0
    }
    if m.viewport > maxStart {
        m.viewport = maxStart
    }
    // Keep cursor within viewport window
    if m.height > 0 {
        if m.cursor < m.viewport {
            m.viewport = m.cursor
        } else if m.cursor >= m.viewport+m.height {
            m.viewport = m.cursor - m.height + 1
        }
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds tea.Cmd
	
	// Create a copy to allow mutations
	newModel := m
	
	// Handle keyboard input for scrolling and selection (extensions)
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k": // Scroll up by one line when height is set
            // Primary behavior: scroll viewport up
            if newModel.height > 0 && newModel.viewport > 0 {
                newModel.viewport--
            } else if newModel.cursor > 0 { // fallback cursor move
                newModel.cursor--
                if newModel.cursor < newModel.viewport {
                    newModel.viewport--
                }
            }
        case "down", "j": // Scroll down by one line when height is set
            visibleNodes := newModel.GetVisibleNodes()
            if newModel.height > 0 {
                // Only scroll if there are more lines below the viewport
                if newModel.viewport+newModel.height < len(visibleNodes) {
                    newModel.viewport++
                }
                // Keep cursor within bounds of visible content
                if newModel.cursor < newModel.viewport {
                    newModel.cursor = newModel.viewport
                }
            } else {
                // Fallback cursor navigation when no height constraint
                if newModel.cursor < len(visibleNodes)-1 {
                    newModel.cursor++
                }
            }
        case "pgup":
            newModel.viewport -= newModel.height / 2
            if newModel.viewport < 0 {
                newModel.viewport = 0
            }
        case "pgdown":
            newModel.viewport += newModel.height / 2
            // View() will handle bounds checking
        }
    }
	
	// Update child nodes
	for id := range newModel.nodes {
		model, cmd := newModel.nodes[id].Update(msg)
		if cmd != nil {
			cmds = tea.Batch(cmds, cmd)
		}
		newModel.nodes[id] = model.(VisibleModel)
	}

	return newModel, cmds
}

func (m Model) View() string {
	sb := strings.Builder{}
	observed := strset.New()
	currentLine := 0 // Track current line for cursor highlighting

	for i, id := range m.roots {
		ret, _ := m.renderNodeWithCursor(i, id, observed, 0, []bool{m.isLastElement(i, m.roots)}, &currentLine)
		if len(ret) > 0 {
			sb.WriteString(ret)
		}
	}

	// Calculate total height from rendered content
	fullContent := strings.TrimRight(sb.String(), "\n")
	if fullContent == "" {
		return ""
	}
	
	contentLines := strings.Split(fullContent, "\n")
	// Note: We can't mutate m.totalHeight here since this is a value receiver
	// The scrolling logic will need to be updated to work with the current content
	
	// Apply scrolling if needed (extension)
	if m.height > 0 && len(contentLines) > m.height {
		start := m.viewport
		end := start + m.height
		if end > len(contentLines) {
			end = len(contentLines)
		}
		if start >= 0 && start < len(contentLines) {
			contentLines = contentLines[start:end]
		}
		fullContent = strings.Join(contentLines, "\n")
	}

	// optionally add a margin to the left of the entire tree
	if m.Margin != "" {
		lines := strings.Split(fullContent, "\n")
		sb = strings.Builder{}
		for i, line := range lines {
			sb.WriteString(m.Margin)
			sb.WriteString(line)
			if i != len(lines)-1 {
				sb.WriteString("\n")
			}
		}
		return sb.String()
	}

	return fullContent
}

// renderNode is the original method for compatibility
func (m Model) renderNode(siblingIdx int, id string, observed *strset.Set, depth int, path []bool) string {
	currentLine := 0
	result, _ := m.renderNodeWithCursor(siblingIdx, id, observed, depth, path, &currentLine)
	return result
}

// renderNodeWithCursor renders a node and tracks cursor position for highlighting
func (m Model) renderNodeWithCursor(siblingIdx int, id string, observed *strset.Set, depth int, path []bool, currentLine *int) (string, int) {
	if observed.Has(id) {
		return "", 0
	}

	observed.Add(id)

	node := m.nodes[id]

	if !node.IsVisible() {
		return "", 0
	}

	prefix := strings.Builder{}

	// handle indentation and prefixes for each level
	for i := 0; i < depth; i++ {
		if m.RootsWithoutPrefix && i == 0 {
			prefix.WriteString(m.Padding)
			continue
		}
		if path[i] {
			prefix.WriteString(m.Indent)
		} else {
			prefix.WriteString(m.Branch)
		}
		prefix.WriteString(m.Padding)
	}

	// determine the correct prefix (fork or leaf)
	if m.RootsWithoutPrefix && depth > 0 || !m.RootsWithoutPrefix {
		prefix.WriteString(m.forkOrLeaf(siblingIdx, id))
	}

	sb := strings.Builder{}

    // add the node's view with selection indicator (extension)
    current := node.View()
    linesRendered := 0
    if len(current) > 0 {
        // Add selection checkbox prefix
        selectionPrefix := "[ ] "
        if m.IsSelected(id) {
            selectionPrefix = "[x] "
        }
        
        // Add cursor highlighting if this is the current line
        isCursorLine := *currentLine == m.cursor
        cursorMarker := "  "
        if isCursorLine {
            cursorMarker = "> "
        }
        fullLine := cursorMarker + selectionPrefix + current
        if isCursorLine {
            // Make the entire cursor line bold for visibility
            fullLine = lipgloss.NewStyle().Bold(true).Render(fullLine)
        }
        current = fullLine
		
		sb.WriteString(m.prefixLines(current, prefix.String(), m.hasChildren(id)))
		sb.WriteString("\n")
		linesRendered = 1
		*currentLine++
	}

	// process all children
	for i, childID := range m.children[id] {
		_, ok := m.nodes[childID]
		if ok && !observed.Has(childID) {
			newPath := append([]bool(nil), path...)
			newPath = append(newPath, m.isLastElement(i, m.children[id]))
			childResult, childLines := m.renderNodeWithCursor(i, childID, observed, depth+1, newPath, currentLine)
			sb.WriteString(childResult)
			linesRendered += childLines
		}
	}

	return sb.String(), linesRendered
}

func (m Model) isLastElement(idx int, siblings []string) bool {
	// check if this is the last visible element in the list of siblings
	for i := idx + 1; i < len(siblings); i++ {
		if m.nodes[siblings[i]].IsVisible() {
			return false
		}
	}
	return true
}

func (m Model) hasChildren(id string) bool {
	// check if there are any children that are visible
	for _, childID := range m.children[id] {
		if m.nodes[childID].IsVisible() {
			return true
		}
	}
	return false
}

func (m Model) forkOrLeaf(siblingIdx int, id string) string {
	if parent, exists := m.parents[id]; exists {
		// index relative to the parent's "children" list
		if m.isLastElement(siblingIdx, m.children[parent]) {
			return m.Leaf
		}
		return m.Fork
	}

	// index relative to the root nodes
	if m.isLastElement(siblingIdx, m.roots) {
		return m.Leaf
	}
	return m.Fork
}

func (m Model) prefixLines(input, prefix string, hasChildren bool) string {
	lines := strings.Split(strings.TrimRight(input, "\n"), "\n")
	sb := strings.Builder{}
	nextPrefix := strings.ReplaceAll(prefix, m.Fork, m.Branch)
	nextPrefix = strings.ReplaceAll(nextPrefix, m.Leaf, m.Indent)

	doPadding := m.VerticalPadMultilineNodes && len(lines) > 1

	for i, line := range lines {
		if i == 0 {
			sb.WriteString(prefix)
		} else {
			sb.WriteString(nextPrefix)
		}
		sb.WriteString(line)
		if doPadding || i != len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	if doPadding {
		sb.WriteString(nextPrefix)
		if hasChildren {
			sb.WriteString(m.Branch)
		}
	}

	return sb.String()
}
