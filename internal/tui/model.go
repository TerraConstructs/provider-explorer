package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/terraconstructs/provider-explorer/internal/config"
	"github.com/terraconstructs/provider-explorer/internal/schema"
)

type AppState int

const (
	StateProviderSelection AppState = iota
	StateResourceTypeSelection
	StateResourceDetailSelection
	StateResourceDetail
)

type Model struct {
	state             AppState
	workingDir        string
	providers         []config.ProviderInfo
	selectedProvider  int
	allSchemas        *schema.ProviderSchema
	
	resourceTypes     []string
	selectedResourceType int
	
	resources         []string
	selectedResource  int
	
	currentSchema     *schema.ResourceSchema
	
	width  int
	height int
	
	err error
}

func NewModel(workingDir string) Model {
	return Model{
		state:           StateProviderSelection,
		workingDir:      workingDir,
		selectedProvider: 0,
		selectedResourceType: 0,
		selectedResource: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		detectProvidersCmd(m.workingDir),
		loadAllSchemasCmd(m.workingDir),
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m.handleBackNavigation()
		}
		
		switch m.state {
		case StateProviderSelection:
			return m.updateProviderSelection(msg)
		case StateResourceTypeSelection:
			return m.updateResourceTypeSelection(msg)
		case StateResourceDetailSelection:
			return m.updateResourceDetailSelection(msg)
		case StateResourceDetail:
			return m.updateResourceDetail(msg)
		}
		
	case ProvidersDetectedMsg:
		m.providers = msg.providers
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
		
	case AllSchemasLoadedMsg:
		m.allSchemas = msg.schema
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
	}
	
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit"
	}
	
	switch m.state {
	case StateProviderSelection:
		return m.renderProviderSelection()
	case StateResourceTypeSelection:
		return m.renderResourceTypeSelection()
	case StateResourceDetailSelection:
		return m.renderResourceDetailSelection()
	case StateResourceDetail:
		return m.renderResourceDetail()
	}
	
	return "Loading..."
}

func (m Model) handleBackNavigation() (Model, tea.Cmd) {
	switch m.state {
	case StateProviderSelection:
		return m, tea.Quit
	case StateResourceTypeSelection:
		m.state = StateProviderSelection
		return m, nil
	case StateResourceDetailSelection:
		m.state = StateResourceTypeSelection
		return m, nil
	case StateResourceDetail:
		m.state = StateResourceDetailSelection
		return m, nil
	}
	return m, nil
}