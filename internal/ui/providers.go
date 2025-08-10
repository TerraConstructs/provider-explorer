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
	providerTitleStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Bold(true).
				Foreground(lipgloss.Color("212"))

	providerItemStyle = lipgloss.NewStyle().PaddingLeft(4)

	providerSelectedItemStyle = lipgloss.NewStyle().
					PaddingLeft(2).
					Foreground(lipgloss.Color("170")).
					Bold(true)
)

// ProviderItem represents a provider in the list
type ProviderItem struct {
	name   string
	schema *tfjson.ProviderSchema
}

// FilterValue implements list.Item
func (i ProviderItem) FilterValue() string { return i.name }

// Title returns the provider name
func (i ProviderItem) Title() string { return i.name }

// Description returns a description of the provider
func (i ProviderItem) Description() string {
	resourceCount := len(i.schema.ResourceSchemas)
	dataSourceCount := len(i.schema.DataSourceSchemas)
	functionCount := len(i.schema.Functions)

	desc := ""
	if resourceCount > 0 {
		desc += fmt.Sprintf("%d resources", resourceCount)
	}
	if dataSourceCount > 0 {
		if desc != "" {
			desc += ", "
		}
		desc += fmt.Sprintf("%d data sources", dataSourceCount)
	}
	if functionCount > 0 {
		if desc != "" {
			desc += ", "
		}
		desc += fmt.Sprintf("%d functions", functionCount)
	}
	return desc
}

// providerDelegate is a custom delegate for provider items
type providerDelegate struct{}

func (d providerDelegate) Height() int                               { return 2 }
func (d providerDelegate) Spacing() int                              { return 1 }
func (d providerDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d providerDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ProviderItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.Title(), i.Description())

	if index == m.Index() {
		fmt.Fprint(w, providerSelectedItemStyle.Render("> "+str))
	} else {
		fmt.Fprint(w, providerItemStyle.Render("  "+str))
	}
}

// ProvidersModel manages the providers list
type ProvidersModel struct {
	list     list.Model
	width    int
	height   int
	focused  bool
	schemas  *tfjson.ProviderSchemas
	selected string
}

// NewProvidersModel creates a new providers model
func NewProvidersModel(width, height int) ProvidersModel {
	l := list.New([]list.Item{}, providerDelegate{}, width, height)
	l.Title = "Providers"
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = providerTitleStyle
	l.SetShowHelp(false)

	return ProvidersModel{
		list:   l,
		width:  width,
		height: height,
	}
}

// SetSize updates the model size
func (m *ProvidersModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SetSchemas updates the provider schemas and rebuilds the list
func (m *ProvidersModel) SetSchemas(schemas *tfjson.ProviderSchemas) {
	m.schemas = schemas

	// Build provider items
	var items []list.Item
	var keys []string
	for key := range schemas.Schemas {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		items = append(items, ProviderItem{
			name:   key,
			schema: schemas.Schemas[key],
		})
	}

	m.list.SetItems(items)
}

// Focus sets focus on the providers list
func (m *ProvidersModel) Focus() {
	m.focused = true
}

// Blur removes focus from the providers list
func (m *ProvidersModel) Blur() {
	m.focused = false
}

// Focused returns whether the providers list is focused
func (m ProvidersModel) Focused() bool {
	return m.focused
}

// SelectedProvider returns the currently selected provider name and schema
func (m ProvidersModel) SelectedProvider() (string, *tfjson.ProviderSchema) {
	if item, ok := m.list.SelectedItem().(ProviderItem); ok {
		return item.name, item.schema
	}
	return "", nil
}

// Update handles messages for the providers model
func (m ProvidersModel) Update(msg tea.Msg) (ProvidersModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

// View renders the providers model
func (m ProvidersModel) View() string {
	return m.list.View()
}
