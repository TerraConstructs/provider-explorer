package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

var (
	typeTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	typeItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	typeSelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	typeDisabledItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("240")).
				Strikethrough(true)
)

// ResourceType represents the type of resource
type ResourceType int

const (
	DataSourcesType ResourceType = iota
	ResourcesType
	EphemeralResourcesType
	ProviderFunctionsType
)

// TypeItem represents a resource type in the picker
type TypeItem struct {
	name     string
	resType  ResourceType
	enabled  bool
	count    int
}

// FilterValue implements list.Item
func (i TypeItem) FilterValue() string { return i.name }

// Title returns the type name
func (i TypeItem) Title() string { return i.name }

// Description returns a description with count
func (i TypeItem) Description() string {
	if i.count == 0 {
		return "No items available"
	}
	return fmt.Sprintf("%d items", i.count)
}

// typeDelegate is a custom delegate for type items
type typeDelegate struct{}

func (d typeDelegate) Height() int                               { return 1 }
func (d typeDelegate) Spacing() int                              { return 0 }
func (d typeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d typeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(TypeItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s (%s)", i.Title(), i.Description())

	if !i.enabled {
		fmt.Fprint(w, typeDisabledItemStyle.Render("  "+str))
	} else if index == m.Index() {
		fmt.Fprint(w, typeSelectedItemStyle.Render("> "+str))
	} else {
		fmt.Fprint(w, typeItemStyle.Render("  "+str))
	}
}

// TypesModel manages the resource type picker
type TypesModel struct {
	list     list.Model
	width    int
	height   int
	focused  bool
	toolInfo terraform.TerraformInfo
	version  string
}

// NewTypesModel creates a new types model
func NewTypesModel(width, height int) TypesModel {
	// Create initial items (will be updated when provider is selected)
	items := []list.Item{
		TypeItem{name: "Data Sources", resType: DataSourcesType, enabled: true, count: 0},
		TypeItem{name: "Resources", resType: ResourcesType, enabled: true, count: 0},
		TypeItem{name: "Ephemeral Resources", resType: EphemeralResourcesType, enabled: false, count: 0},
		TypeItem{name: "Provider Functions", resType: ProviderFunctionsType, enabled: false, count: 0},
	}

	l := list.New(items, typeDelegate{}, width, height)
	l.Title = "Type"
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = typeTitleStyle
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return TypesModel{
		list:   l,
		width:  width,
		height: height,
	}
}

// SetSize updates the model size
func (m *TypesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SetToolInfo updates tool information for feature gating
func (m *TypesModel) SetToolInfo(toolInfo terraform.TerraformInfo, version string) {
	m.toolInfo = toolInfo
	m.version = version
	m.updateFeatureSupport()
}

// SetCounts updates the counts for each type based on the selected provider
func (m *TypesModel) SetCounts(dataSourceCount, resourceCount, ephemeralCount, functionCount int) {
	items := []list.Item{
		TypeItem{
			name:    "Data Sources",
			resType: DataSourcesType,
			enabled: true,
			count:   dataSourceCount,
		},
		TypeItem{
			name:    "Resources", 
			resType: ResourcesType,
			enabled: true,
			count:   resourceCount,
		},
		TypeItem{
			name:    "Ephemeral Resources",
			resType: EphemeralResourcesType,
			enabled: m.supportsEphemeralResources(),
			count:   ephemeralCount,
		},
		TypeItem{
			name:    "Provider Functions",
			resType: ProviderFunctionsType,
			enabled: m.supportsProviderFunctions(),
			count:   functionCount,
		},
	}

	m.list.SetItems(items)
}

// Focus sets focus on the types list
func (m *TypesModel) Focus() {
	m.focused = true
}

// Blur removes focus from the types list
func (m *TypesModel) Blur() {
	m.focused = false
}

// Focused returns whether the types list is focused
func (m TypesModel) Focused() bool {
	return m.focused
}

// SelectedType returns the currently selected resource type
func (m TypesModel) SelectedType() (ResourceType, bool) {
	if item, ok := m.list.SelectedItem().(TypeItem); ok {
		return item.resType, item.enabled
	}
	return DataSourcesType, false
}

// MoveToEnabledItem moves selection to the next enabled item if current is disabled
func (m *TypesModel) MoveToEnabledItem() {
	current := m.list.Index()
	items := m.list.Items()

	// If current item is enabled, nothing to do
	if current < len(items) {
		if item, ok := items[current].(TypeItem); ok && item.enabled {
			return
		}
	}

	// Find next enabled item
	for i := 0; i < len(items); i++ {
		if item, ok := items[i].(TypeItem); ok && item.enabled {
			m.list.Select(i)
			return
		}
	}
}

// updateFeatureSupport updates which features are supported
func (m *TypesModel) updateFeatureSupport() {
	// Get current items and update their enabled status
	currentItems := m.list.Items()
	newItems := make([]list.Item, len(currentItems))

	for i, item := range currentItems {
		if typeItem, ok := item.(TypeItem); ok {
			switch typeItem.resType {
			case EphemeralResourcesType:
				typeItem.enabled = m.supportsEphemeralResources()
			case ProviderFunctionsType:
				typeItem.enabled = m.supportsProviderFunctions()
			}
			newItems[i] = typeItem
		} else {
			newItems[i] = item
		}
	}

	m.list.SetItems(newItems)
	m.MoveToEnabledItem()
}

// supportsEphemeralResources checks if ephemeral resources are supported
func (m TypesModel) supportsEphemeralResources() bool {
	if m.toolInfo.Tool == "" || m.version == "" {
		return false
	}
	return m.toolInfo.SupportsFeature(terraform.EphemeralResources, m.version)
}

// supportsProviderFunctions checks if provider functions are supported
func (m TypesModel) supportsProviderFunctions() bool {
	if m.toolInfo.Tool == "" || m.version == "" {
		return false
	}
	return m.toolInfo.SupportsFeature(terraform.ProviderFunctions, m.version)
}

// Update handles messages for the types model
func (m TypesModel) Update(msg tea.Msg) (TypesModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	
	return m, cmd
}

// View renders the types model
func (m TypesModel) View() string {
	return m.list.View()
}