package ui

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
)

var (
	entityTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	entityItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	entitySelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	filterFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	hintUnfocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")) // Dimmed gray for unfocused state
)

// EntityItem represents an entity (resource, data source, etc.) in the list
type EntityItem struct {
	name   string
	schema *tfjson.Schema
}

// FilterValue implements list.Item
func (i EntityItem) FilterValue() string { return i.name }

// Title returns the entity name
func (i EntityItem) Title() string { return i.name }

// Description returns a description of the entity
func (i EntityItem) Description() string {
	if i.schema == nil || i.schema.Block == nil {
		return "No schema available"
	}

	attrCount := len(i.schema.Block.Attributes)
	blockCount := len(i.schema.Block.NestedBlocks)

	desc := ""
	if attrCount > 0 {
		desc += fmt.Sprintf("%d attributes", attrCount)
	}
	if blockCount > 0 {
		if desc != "" {
			desc += ", "
		}
		desc += fmt.Sprintf("%d blocks", blockCount)
	}

	if desc == "" {
		desc = "Empty schema"
	}

	return desc
}

// entityDelegate is a custom delegate for entity items
type entityDelegate struct{}

func (d entityDelegate) Height() int                               { return 2 }
func (d entityDelegate) Spacing() int                              { return 1 }
func (d entityDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d entityDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(EntityItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.Title(), i.Description())

	if index == m.Index() {
		fmt.Fprint(w, entitySelectedItemStyle.Render("> "+str))
	} else {
		fmt.Fprint(w, entityItemStyle.Render("  "+str))
	}
}

// EntitiesModel manages the entities list
type EntitiesModel struct {
	list         list.Model
	width        int
	height       int
	focused      bool
	filterFocused bool
	currentType  ResourceType
	provider     string
	providerSchema *tfjson.ProviderSchema
}

// NewEntitiesModel creates a new entities model
func NewEntitiesModel(width, height int) EntitiesModel {
	l := list.New([]list.Item{}, entityDelegate{}, width, height)
	l.Title = "Entities"
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = entityTitleStyle
	l.SetShowHelp(false)

	return EntitiesModel{
		list:   l,
		width:  width,
		height: height,
	}
}

// SetSize updates the model size
func (m *EntitiesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SetProvider updates the provider and rebuilds the list based on current type
func (m *EntitiesModel) SetProvider(providerName string, schema *tfjson.ProviderSchema) {
	m.provider = providerName
	m.providerSchema = schema
	m.rebuildList()
}

// SetType updates the resource type and rebuilds the list
func (m *EntitiesModel) SetType(resType ResourceType) {
	m.currentType = resType
	m.rebuildList()
	
	// Update title based on type
	switch resType {
	case DataSourcesType:
		m.list.Title = "Data Sources"
	case ResourcesType:
		m.list.Title = "Resources"
	case EphemeralResourcesType:
		m.list.Title = "Ephemeral Resources"
	case ProviderFunctionsType:
		m.list.Title = "Provider Functions"
	}
}

// rebuildList rebuilds the entity list based on current provider and type
func (m *EntitiesModel) rebuildList() {
	if m.providerSchema == nil {
		m.list.SetItems([]list.Item{})
		return
	}

	var items []list.Item
	var keys []string

	switch m.currentType {
	case DataSourcesType:
		for key := range m.providerSchema.DataSourceSchemas {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			items = append(items, EntityItem{
				name:   key,
				schema: m.providerSchema.DataSourceSchemas[key],
			})
		}

	case ResourcesType:
		for key := range m.providerSchema.ResourceSchemas {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			items = append(items, EntityItem{
				name:   key,
				schema: m.providerSchema.ResourceSchemas[key],
			})
		}

	case EphemeralResourcesType:
		for key := range m.providerSchema.EphemeralResourceSchemas {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			items = append(items, EntityItem{
				name:   key,
				schema: m.providerSchema.EphemeralResourceSchemas[key],
			})
		}

	case ProviderFunctionsType:
		for key := range m.providerSchema.Functions {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			// Functions don't have schema like resources, create a placeholder
			items = append(items, EntityItem{
				name:   key,
				schema: nil, // Functions have different structure
			})
		}
	}

	m.list.SetItems(items)
}

// Focus sets focus on the entities list
func (m *EntitiesModel) Focus() {
	m.focused = true
}

// Blur removes focus from the entities list
func (m *EntitiesModel) Blur() {
	m.focused = false
	m.filterFocused = false
}

// Focused returns whether the entities list is focused
func (m EntitiesModel) Focused() bool {
	return m.focused
}

// StartFiltering starts filter mode
func (m *EntitiesModel) StartFiltering() {
	if !m.focused {
		return
	}
	m.filterFocused = true
	// Note: The actual filter activation happens when "/" is sent to the list in Update()
}

// StopFiltering stops filter mode
func (m *EntitiesModel) StopFiltering() {
	m.filterFocused = false
	if m.list.FilterState() == list.Filtering {
		m.list.ResetFilter()
	}
}

// IsFilterFocused returns whether the filter is currently focused
func (m EntitiesModel) IsFilterFocused() bool {
	return m.filterFocused
}

// HasAppliedFilter returns whether there's a filter currently applied
func (m EntitiesModel) HasAppliedFilter() bool {
	return m.list.FilterState() == list.FilterApplied
}

// GetCurrentFilter returns the current filter value
func (m EntitiesModel) GetCurrentFilter() string {
	return m.list.FilterInput.Value()
}

// SelectedEntity returns the currently selected entity
func (m EntitiesModel) SelectedEntity() (string, *tfjson.Schema) {
	if item, ok := m.list.SelectedItem().(EntityItem); ok {
		return item.name, item.schema
	}
	return "", nil
}

// Update handles messages for the entities model
func (m EntitiesModel) Update(msg tea.Msg) (EntitiesModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	// Handle special keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "/":
			// Start filtering mode
			m.StartFiltering()
			// Send the key to the list to start filtering
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case "esc":
			// Cancel filtering mode (reset to original list) 
			if m.filterFocused {
				m.StopFiltering()
				return m, nil
			} else if m.list.FilterState() == list.FilterApplied {
				// Clear applied filter
				m.list.ResetFilter()
				return m, nil
			}
		case "enter":
			// Apply current filter (exit filter mode but keep filtered results)
			if m.filterFocused {
				m.filterFocused = false
				// Send enter to the list to accept the filter
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	
	return m, cmd
}

// View renders the entities model
func (m EntitiesModel) View() string {
	view := m.list.View()
	
	// Always show instruction text to maintain consistent height
	if m.focused && m.filterFocused {
		hint := filterFocusedStyle.Render("(filtering - Enter=apply, Esc=cancel)")
		view += "\n" + hint
	} else if m.focused {
		var hint string
		if m.list.FilterState() == list.FilterApplied {
			hint = "press / to filter, esc to clear"
		} else {
			hint = "press / to filter"
		}
		view += "\n" + hint
	} else {
		// Show dimmed instruction text when unfocused to maintain consistent height
		var hint string
		if m.list.FilterState() == list.FilterApplied {
			hint = hintUnfocusedStyle.Render("press / to filter, esc to clear")
		} else {
			hint = hintUnfocusedStyle.Render("press / to filter")
		}
		view += "\n" + hint
	}
	
	return view
}