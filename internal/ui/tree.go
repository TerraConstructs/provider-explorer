package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/terraconstructs/provider-explorer/internal/ui/tree"
)

var (
	treeTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))
)

// ViewMode represents whether we're viewing Arguments or Attributes
type ViewMode string

const (
	ArgumentsMode  ViewMode = "Arguments"
	AttributesMode ViewMode = "Attributes"
)

// SchemaTreeModel manages the schema tree view using our forked tree component
type SchemaTreeModel struct {
	treeModel    tree.Model
	width        int
	height       int
	focused      bool
	mode         ViewMode
	entity       string
	schema       *tfjson.Schema
	nodePathMap  map[string][]string // maps node IDs to paths
	pathToNodeID map[string]string   // reverse mapping
	currentIndex int                 // for generating unique node IDs
	nodeIsBlock  map[string]bool     // track node type for cascade selection
}

// NewSchemaTreeModel creates a new schema tree model
func NewSchemaTreeModel(width, height int) SchemaTreeModel {
	treeModel := tree.NewModel()
	treeModel.SetHeight(height - 2) // Reserve space for title and instructions

	return SchemaTreeModel{
		treeModel:    treeModel,
		width:        width,
		height:       height,
		mode:         ArgumentsMode,
		nodePathMap:  make(map[string][]string),
		pathToNodeID: make(map[string]string),
		nodeIsBlock:  make(map[string]bool),
	}
}

// SetSize updates the model size
func (m *SchemaTreeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.treeModel.SetHeight(height - 2) // Reserve space for title and instructions
}

// SetSchema updates the schema and rebuilds the tree
func (m *SchemaTreeModel) SetSchema(entityName string, schema *tfjson.Schema) {
	m.entity = entityName
	m.schema = schema
	m.rebuildTree()
}

// ToggleMode switches between Arguments and Attributes view
func (m *SchemaTreeModel) ToggleMode() {
	if m.mode == ArgumentsMode {
		m.mode = AttributesMode
	} else {
		m.mode = ArgumentsMode
	}
	m.rebuildTree()
}

// GetMode returns the current view mode
func (m SchemaTreeModel) GetMode() ViewMode {
	return m.mode
}

// rebuildTree rebuilds the tree based on current mode and schema
func (m *SchemaTreeModel) rebuildTree() {
	if m.schema == nil || m.schema.Block == nil {
		return
	}

	// Clear existing mappings
	m.nodePathMap = make(map[string][]string)
	m.pathToNodeID = make(map[string]string)
	m.currentIndex = 0
	m.nodeIsBlock = make(map[string]bool)

	// Create new tree model
	m.treeModel = tree.NewModel()
	m.treeModel.SetHeight(m.height - 2)

	switch m.mode {
	case ArgumentsMode:
		// Show required and optional attributes
		for name, attr := range m.schema.Block.Attributes {
			if !attr.Computed { // Arguments are non-computed
				path := []string{name}
				nodeID := m.generateNodeID()
				schemaNode := tree.NewAttributeNode(nodeID, name, attr, path)
				m.nodePathMap[nodeID] = path
				m.pathToNodeID[m.pathKey(path)] = nodeID
				m.treeModel.Add("", nodeID, schemaNode)
				m.nodeIsBlock[nodeID] = false
			}
		}

		// Add nested blocks (these are also arguments)
		for name, block := range m.schema.Block.NestedBlocks {
			path := []string{name}
			m.addBlockNodes("", name, block.Block, path)
		}

	case AttributesMode:
		// Show computed attributes
		for name, attr := range m.schema.Block.Attributes {
			if attr.Computed {
				path := []string{name}
				nodeID := m.generateNodeID()
				schemaNode := tree.NewAttributeNode(nodeID, name, attr, path)
				m.nodePathMap[nodeID] = path
				m.pathToNodeID[m.pathKey(path)] = nodeID
				m.treeModel.Add("", nodeID, schemaNode)
				m.nodeIsBlock[nodeID] = false
			}
		}

		// Computed nested blocks are less common but possible
		for name, block := range m.schema.Block.NestedBlocks {
			path := []string{name}
			m.addBlockNodes("", name, block.Block, path)
		}
	}
}

// addBlockNodes recursively adds block nodes and their children
func (m *SchemaTreeModel) addBlockNodes(parentID, name string, block *tfjson.SchemaBlock, path []string) {
	nodeID := m.generateNodeID()
	schemaNode := tree.NewBlockNode(nodeID, name, block, path)
	m.nodePathMap[nodeID] = path
	m.pathToNodeID[m.pathKey(path)] = nodeID
	m.treeModel.Add(parentID, nodeID, schemaNode)
	m.nodeIsBlock[nodeID] = true

	// Add attributes from the nested block
	for attrName, attr := range block.Attributes {
		childPath := append(path, attrName)
		childNodeID := m.generateNodeID()
		childSchemaNode := tree.NewAttributeNode(childNodeID, attrName, attr, childPath)
		m.nodePathMap[childNodeID] = childPath
		m.pathToNodeID[m.pathKey(childPath)] = childNodeID
		m.treeModel.Add(nodeID, childNodeID, childSchemaNode)
		m.nodeIsBlock[childNodeID] = false
	}

	// Add nested blocks recursively
	for blockName, nestedBlock := range block.NestedBlocks {
		childPath := append(path, blockName)
		m.addBlockNodes(nodeID, blockName, nestedBlock.Block, childPath)
	}
}

// toggleNodeSelectionCascade toggles selection for a node and all its descendants.
// For block nodes, this will select/deselect all descendant blocks and attributes.
func (m *SchemaTreeModel) toggleNodeSelectionCascade(nodeID string) {
	desired := !m.treeModel.IsSelected(nodeID)
	base, ok := m.nodePathMap[nodeID]
	if !ok {
		// Fallback to simple toggle if path is unknown
		m.treeModel.ToggleSelection(nodeID)
		return
	}
	// Set selection for this node
	m.treeModel.SetSelection(nodeID, desired)
	// Apply to all descendants
	for id, p := range m.nodePathMap {
		if id == nodeID {
			continue
		}
		if isDescendantPath(base, p) {
			m.treeModel.SetSelection(id, desired)
		}
	}
}

func isDescendantPath(prefix, candidate []string) bool {
	if len(candidate) <= len(prefix) {
		return false
	}
	for i := range prefix {
		if candidate[i] != prefix[i] {
			return false
		}
	}
	return true
}

// generateNodeID generates a unique node ID
func (m *SchemaTreeModel) generateNodeID() string {
	id := fmt.Sprintf("node_%d", m.currentIndex)
	m.currentIndex++
	return id
}

// pathKey creates a string key from a path for map lookups
func (m *SchemaTreeModel) pathKey(path []string) string {
	return strings.Join(path, ".")
}

// Focus sets focus on the tree
func (m *SchemaTreeModel) Focus() {
	m.focused = true
}

// Blur removes focus from the tree
func (m *SchemaTreeModel) Blur() {
	m.focused = false
}

// Focused returns whether the tree is focused
func (m SchemaTreeModel) Focused() bool {
	return m.focused
}

// GetSelectedPaths returns the paths of all selected nodes
func (m SchemaTreeModel) GetSelectedPaths() [][]string {
	selectedNodeIDs := m.treeModel.GetSelectedNodes()
	var paths [][]string
	for _, nodeID := range selectedNodeIDs {
		if path, exists := m.nodePathMap[nodeID]; exists {
			paths = append(paths, path)
		}
	}
	return paths
}

// ClearSelection clears all selected nodes
func (m *SchemaTreeModel) ClearSelection() {
	m.treeModel.ClearSelection()
}

// SelectAll selects all visible nodes
func (m *SchemaTreeModel) SelectAll() {
	m.treeModel.SelectAll()
}

// MoveDown moves selection/viewport down in the tree
func (m *SchemaTreeModel) MoveDown() {
	m.treeModel.MoveDown()
}

// MoveUp moves selection/viewport up in the tree
func (m *SchemaTreeModel) MoveUp() {
	m.treeModel.MoveUp()
}

// MovePageDown moves down by one page
func (m *SchemaTreeModel) MovePageDown() {
	m.treeModel.MovePageDown()
}

// MovePageUp moves up by one page
func (m *SchemaTreeModel) MovePageUp() {
	m.treeModel.MovePageUp()
}

// Update handles messages for the tree model
func (m SchemaTreeModel) Update(msg tea.Msg) (SchemaTreeModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	// Handle our custom key bindings
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case " ": // Space to toggle selection
			currentNode := m.treeModel.GetCurrentNode()
			if currentNode != "" {
				if m.nodeIsBlock[currentNode] {
					m.toggleNodeSelectionCascade(currentNode)
				} else {
					m.treeModel.ToggleSelection(currentNode)
				}
			}
			return m, nil
		case "j": // Navigate down
			m.treeModel.MoveDown()
			return m, nil
		case "k": // Navigate up
			m.treeModel.MoveUp()
			return m, nil
		case "pgdown":
			m.treeModel.MovePageDown()
			return m, nil
		case "pgup":
			m.treeModel.MovePageUp()
			return m, nil
		case "ctrl+a": // Toggle select all / clear
			visible := m.treeModel.GetVisibleNodes()
			selected := m.treeModel.GetSelectedNodes()
			if len(visible) > 0 && len(selected) >= len(visible) {
				m.ClearSelection()
			} else {
				m.SelectAll()
			}
			return m, nil
		case "escape": // Clear selection when in tree view
			m.ClearSelection()
			return m, nil
		case "a": // Toggle arguments/attributes
			m.ToggleMode()
			return m, nil
		}
	}

	// Forward to tree model
	var cmd tea.Cmd
	newTreeModel, cmd := m.treeModel.Update(msg)
	m.treeModel = newTreeModel.(tree.Model)

	return m, cmd
}

// View renders the tree model
func (m SchemaTreeModel) View() string {
	title := treeTitleStyle.Render(fmt.Sprintf("Schema (%s)", m.mode))

	// Calculate available height for tree content (total - title - instructions)
	availableHeight := m.height - 2
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Ensure tree model has the right height
	m.treeModel.SetHeight(availableHeight)
	// Clamp viewport/cursor to ensure cursor visibility within content area
	m.treeModel.Clamp()

	treeView := m.treeModel.View()

	// Add selection info at bottom
	selectedPaths := m.GetSelectedPaths()
	var instructions string
	if len(selectedPaths) > 0 {
		instructions = fmt.Sprintf("Selected: %d nodes (press 'e' to export, esc to clear)", len(selectedPaths))
	} else {
		instructions = "↑/↓ or j/k to navigate, space to select, ctrl+a to select all, 'a' to toggle mode"
	}

	return fmt.Sprintf("%s\n%s\n%s", title, treeView, instructions)
}
