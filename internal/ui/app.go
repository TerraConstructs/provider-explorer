package ui

import (
    "github.com/charmbracelet/bubbles/help"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    tfjson "github.com/hashicorp/terraform-json"
    "github.com/terraconstructs/provider-explorer/internal/terraform"
    "strings"
)

var (
	// Border styles for focus indication
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("14")). // Bright cyan
				Padding(0, 1)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")). // Dark gray
				Padding(0, 1)
)

// FocusArea represents which UI component is currently focused
type FocusArea int

const (
	FocusProviders FocusArea = iota
	FocusTypes
	FocusEntities
	FocusTree
)

// AppStage represents the current stage of the application flow
type AppStage int

const (
    StageLoading AppStage = iota
    StageProviderSelect
    StageTypeSelect
    StageEntityBrowse
    StageTreeView
    StageExportResult
)

// schemaLoadedMsg is sent when schemas are loaded
type schemaLoadedMsg struct {
	schemas     *tfjson.ProviderSchemas
	toolInfo    terraform.TerraformInfo
	version     string
	err         error
}

// exportRequestMsg is sent when user requests export
type exportRequestMsg struct{}

// Model represents the main application model
type Model struct {
	// UI Components
	providers ProvidersModel
	types     TypesModel
	entities  EntitiesModel
	tree      SchemaTreeModel
	status    StatusBar
	help      help.Model
	keys      KeyMap

	// Layout
	width  int
	height int

	// State
	stage   AppStage
	focus   FocusArea
	schemas *tfjson.ProviderSchemas
	toolInfo terraform.TerraformInfo
	version  string

	// Current selections
	selectedProvider string
	selectedType     ResourceType
	selectedEntity   string

	// Export result
	exportResult string
	showHelp     bool

	// Export (attributes) name prompt state
	exportNamePrompt bool
	exportName       string
}

// NewModel creates a new application model
func NewModel(width, height int) Model {
	// Calculate layout dimensions
	leftWidth := width / 3
	rightWidth := width - leftWidth
	topHeight := 8  // Types picker height
	bottomHeight := height - topHeight - 3 // 3 for status and margins

    return Model{
        providers: NewProvidersModel(leftWidth, height-3),
        types:     NewTypesModel(rightWidth, topHeight),
        entities:  NewEntitiesModel(rightWidth, bottomHeight/2),
        tree:      NewSchemaTreeModel(rightWidth, bottomHeight/2),
        status:    NewStatusBar(width),
        help:      help.New(),
        keys:      DefaultKeyMap(),
        width:     width,
        height:    height,
        stage:     StageLoading,
        focus:     FocusProviders,
    }
}

// loadSchemaCmd loads the provider schemas
func loadSchemaCmd(workingDir string) tea.Cmd {
	return func() tea.Msg {
		schemaWithVersion, err := terraform.FetchAllProviderSchemas(workingDir)
		if err != nil {
			return schemaLoadedMsg{err: err}
		}
		
		var version string
		if schemaWithVersion.VersionInfo != nil {
			version = schemaWithVersion.VersionInfo.Version
		}
		
		return schemaLoadedMsg{
			schemas:  schemaWithVersion.Schemas,
			toolInfo: schemaWithVersion.TfInfo,
			version:  version,
		}
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return loadSchemaCmd(".")
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Handle export name prompt input first
    if m.exportNamePrompt {
        if km, ok := msg.(tea.KeyMsg); ok {
            switch km.String() {
            case "esc":
                m.exportNamePrompt = false
                return m, nil
            case "enter":
                if strings.TrimSpace(m.exportName) == "" {
                    // Block empty input; stay in dialog
                    return m, nil
                }
                // Proceed with attributes export using provided instance name
                entityName, entitySchema := m.entities.SelectedEntity()
                if entitySchema != nil {
                    selected := m.filteredSelectedPaths(m.tree.GetSelectedPaths())
                    m.exportResult = ConvertSelectedAttributesToHCLOutputs(entityName, entitySchema, m.selectedProvider, m.exportName, selected)
                    m.stage = StageExportResult
                    m.tree.Blur()
                }
                m.exportNamePrompt = false
                return m, nil
            case "backspace", "ctrl+h":
                if len(m.exportName) > 0 {
                    m.exportName = m.exportName[:len(m.exportName)-1]
                }
                return m, nil
            default:
                // Append runes for normal characters
                if km.Type == tea.KeyRunes {
                    m.exportName += string(km.Runes)
                    return m, nil
                }
            }
        }
        // While prompt is active, don't forward to other components
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

    case schemaLoadedMsg:
        if msg.err != nil {
            // TODO: Handle error better
            return m, tea.Quit
        }
        m.schemas = msg.schemas
        m.toolInfo = msg.toolInfo
        m.version = msg.version
        
        // Update components with loaded data
        m.providers.SetSchemas(msg.schemas)
        m.types.SetToolInfo(msg.toolInfo, msg.version)
        m.status.SetToolInfo(msg.toolInfo, msg.version)
        
        // Check if only one provider exists - auto-select it
        if len(msg.schemas.Schemas) == 1 {
            // Extract the single provider name and schema
            var providerName string
            var providerSchema *tfjson.ProviderSchema
            for name, schema := range msg.schemas.Schemas {
                providerName = name
                providerSchema = schema
                break
            }
            
            // Update application state with selected provider
            m.selectedProvider = providerName
            m.status.SetProvider(providerName)
            m.types.SetCounts(
                len(providerSchema.DataSourceSchemas),
                len(providerSchema.ResourceSchemas),
                len(providerSchema.EphemeralResourceSchemas),
                len(providerSchema.Functions),
            )
            
            // Skip provider selection and go directly to type selection
            m.stage = StageTypeSelect
            m.focus = FocusTypes
            m.providers.Blur()
            m.types.Focus()
        } else {
            // Multiple providers - show provider selection stage
            m.stage = StageProviderSelect
            m.focus = FocusProviders
            m.providers.Focus()
        }

    case tea.KeyMsg:
        // Direct tree navigation keys when tree is focused
        if m.stage == StageTreeView && m.focus == FocusTree {
            switch msg.String() {
            case "j", "down":
                m.tree.MoveDown()
                return m, nil
            case "k", "up":
                m.tree.MoveUp()
                return m, nil
            case "pgdown":
                m.tree.treeModel.MovePageDown()
                return m, nil
            case "pgup":
                m.tree.treeModel.MovePageUp()
                return m, nil
            }
        }
        switch msg.String() {
            case "ctrl+c", "q":
                return m, tea.Quit
            case "?":
                m.showHelp = !m.showHelp
                return m, nil
		case "tab":
			return m, m.handleTabNavigation()
		case "enter":
			// If entities are focused and filtering, let them handle enter
			if m.focus == FocusEntities && m.entities.IsFilterFocused() {
				// Forward to entities component
				var cmd tea.Cmd
				m.entities, cmd = m.entities.Update(msg)
				return m, cmd
			}
			return m, m.handleEnter()
		case "esc":
			// If entities are focused and have filter state, let them handle escape
			if m.focus == FocusEntities && (m.entities.IsFilterFocused() || m.entities.HasAppliedFilter()) {
				// Forward to entities component
				var cmd tea.Cmd
				m.entities, cmd = m.entities.Update(msg)
				return m, cmd
			}
			return m, m.handleEscape()
		case "e":
			if m.stage == StageTreeView && m.focus == FocusTree {
				return m, m.handleExport()
			}
		case "c":
			if m.stage == StageExportResult && m.exportResult != "" {
				return m, m.handleCopy()
			}
		}
	}

	// Forward messages to focused component
	switch m.focus {
	case FocusProviders:
		var cmd tea.Cmd
		m.providers, cmd = m.providers.Update(msg)
		cmds = append(cmds, cmd)
	case FocusTypes:
		var cmd tea.Cmd
		m.types, cmd = m.types.Update(msg)
		cmds = append(cmds, cmd)
	case FocusEntities:
		var cmd tea.Cmd
		m.entities, cmd = m.entities.Update(msg)
		cmds = append(cmds, cmd)
		
		// Update status with filter
		m.status.SetFilter(m.entities.GetCurrentFilter())
	case FocusTree:
		var cmd tea.Cmd
		m.tree, cmd = m.tree.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateLayout updates component sizes based on current dimensions and stage
func (m *Model) updateLayout() {
	m.status.SetWidth(m.width)
	
	if m.stage == StageTreeView {
		// Tree view: 2-pane layout (entities + tree)
		halfWidth := (m.width / 2) - 2 // Account for borders
		fullHeight := m.height - 2 // Reserve space for status bar
		
		// Ensure minimum sizes
		if halfWidth < 10 { halfWidth = 10 }
		if fullHeight < 10 { fullHeight = 10 }
		
		m.entities.SetSize(halfWidth, fullHeight)
		m.tree.SetSize(halfWidth, fullHeight)
	} else {
		// Navigation view: providers+types stacked on left, entities on right (50-50 split)
		// Account for borders (2 chars per side = 4 chars total per border)
		leftWidth := (m.width / 2) - 4
		rightWidth := (m.width / 2) - 4
		fullHeight := m.height - 4 // Reserve space for status bar
		
		// Calculate heights for left column - account for list titles and vertical spacing
		typesHeight := 8 // Types list (4 items + title + borders)
		// Account for spacing between stacked components and list titles
		verticalSpacing := 1 // Space between providers and types panes
		providersHeight := fullHeight - typesHeight - verticalSpacing
		
		// The entities height should match the actual combined left column height
		// Left column = providersHeight + typesHeight + spacing between them
		entitiesHeight := providersHeight + typesHeight + verticalSpacing
		
		// Ensure minimum sizes - but account for list titles needing space
		if leftWidth < 10 { leftWidth = 10 }
		if rightWidth < 10 { rightWidth = 10 }
		if providersHeight < 8 { providersHeight = 8 } // Need space for title + content
		if typesHeight < 6 { typesHeight = 6 } // Need space for title + items
		if entitiesHeight < 10 { entitiesHeight = 10 }
		
		m.providers.SetSize(leftWidth, providersHeight)
		m.types.SetSize(leftWidth, typesHeight)
		m.entities.SetSize(rightWidth, entitiesHeight) // Match left column total height
		
		// Tree gets minimal size when not shown  
		m.tree.SetSize(rightWidth, 10)
	}
}

// handleTabNavigation handles tab key for focus cycling
func (m *Model) handleTabNavigation() tea.Cmd {
	// Blur current
	switch m.focus {
	case FocusProviders:
		m.providers.Blur()
	case FocusTypes:
		m.types.Blur()
	case FocusEntities:
		m.entities.Blur()
	case FocusTree:
		m.tree.Blur()
	}

	// Move to next focus based on stage
	switch m.stage {
	case StageProviderSelect:
		m.focus = FocusProviders
	case StageTypeSelect:
		if m.focus == FocusProviders {
			m.focus = FocusTypes
		} else {
			m.focus = FocusProviders
		}
	case StageEntityBrowse:
		switch m.focus {
		case FocusProviders:
			m.focus = FocusTypes
		case FocusTypes:
			m.focus = FocusEntities
		case FocusEntities:
			m.focus = FocusProviders
		}
	case StageTreeView:
		if m.focus == FocusEntities {
			m.focus = FocusTree
		} else {
			m.focus = FocusEntities
		}
	}

	// Focus new component
	switch m.focus {
	case FocusProviders:
		m.providers.Focus()
	case FocusTypes:
		m.types.Focus()
	case FocusEntities:
		m.entities.Focus()
	case FocusTree:
		m.tree.Focus()
	}

	return nil
}

// handleEnter handles enter key for selection
func (m *Model) handleEnter() tea.Cmd {
	switch m.focus {
	case FocusProviders:
		if providerName, providerSchema := m.providers.SelectedProvider(); providerName != "" {
			m.selectedProvider = providerName
			m.stage = StageTypeSelect
			m.focus = FocusTypes
			
			// Update status and types with provider info
			m.status.SetProvider(providerName)
			m.types.SetCounts(
				len(providerSchema.DataSourceSchemas),
				len(providerSchema.ResourceSchemas),
				len(providerSchema.EphemeralResourceSchemas),
				len(providerSchema.Functions),
			)
			
			// Focus types and blur providers
			m.providers.Blur()
			m.types.Focus()
		}
		
	case FocusTypes:
		if resourceType, enabled := m.types.SelectedType(); enabled {
			m.selectedType = resourceType
			m.stage = StageEntityBrowse
			m.focus = FocusEntities
			
			// Update entities with selected provider and type
			if providerName, providerSchema := m.providers.SelectedProvider(); providerName != "" {
				m.entities.SetProvider(providerName, providerSchema)
				m.entities.SetType(resourceType)
				
				// Update status
				typeName := ""
				switch resourceType {
				case DataSourcesType:
					typeName = "Data Sources"
				case ResourcesType:
					typeName = "Resources"
				case EphemeralResourcesType:
					typeName = "Ephemeral Resources"
				case ProviderFunctionsType:
					typeName = "Provider Functions"
				}
				m.status.SetResourceType(typeName)
			}
			
			// Focus entities and blur types
			m.types.Blur()
			m.entities.Focus()
		}
		
	case FocusEntities:
		if entityName, entitySchema := m.entities.SelectedEntity(); entityName != "" {
			m.selectedEntity = entityName
			m.tree.SetSchema(entityName, entitySchema)
			
			// Transition to tree view stage
			m.stage = StageTreeView
			m.focus = FocusTree
			
			// Update layout for new stage (this resizes components appropriately)
			m.updateLayout()
			
			// Focus tree and blur all navigation components
			m.providers.Blur()
			m.types.Blur()
			m.entities.Blur()
			m.tree.Focus()
		}
	}
	
	return nil
}

// handleEscape handles escape key for going back
func (m *Model) handleEscape() tea.Cmd {
	switch m.stage {
	case StageTypeSelect:
		m.stage = StageProviderSelect
		m.focus = FocusProviders
		m.types.Blur()
		m.providers.Focus()
		m.status.SetProvider("")
		m.status.SetResourceType("")
		
	case StageEntityBrowse:
		m.stage = StageTypeSelect
		m.focus = FocusTypes
		m.entities.Blur()
		m.types.Focus()
		
	case StageTreeView:
		if m.focus == FocusTree {
			// Switch back to entity list in tree view
			m.focus = FocusEntities
			m.tree.Blur()
			m.entities.Focus()
		} else {
			// Go back to navigation mode
			m.stage = StageEntityBrowse
			m.focus = FocusEntities
			m.tree.Blur()
			m.entities.Focus()
		}
		
	case StageExportResult:
		m.stage = StageTreeView
		m.focus = FocusTree
		m.tree.Focus()
		m.exportResult = ""
	}
	
	return nil
}

// handleExport handles export request
func (m *Model) handleExport() tea.Cmd {
	selectedPaths := m.tree.GetSelectedPaths()
	if len(selectedPaths) == 0 {
		return nil
	}

    // Get current entity schema
    _, entitySchema := m.entities.SelectedEntity()
	if entitySchema == nil {
		return nil
	}

	// Generate HCL based on tree mode
	switch m.tree.GetMode() {
	case ArgumentsMode:
		// Export only selected argument attributes
		m.exportResult = ConvertSelectedArgumentsToHCLVariables(entitySchema, m.filteredSelectedPaths(selectedPaths))
		m.stage = StageExportResult
		m.tree.Blur()
		return nil
	case AttributesMode:
		// Prompt for resource instance name before exporting
		m.exportNamePrompt = true
		m.exportName = "main"
		return nil
	}

	return nil
}

// handleCopy handles copy to clipboard
func (m *Model) handleCopy() tea.Cmd {
	return func() tea.Msg {
		if err := CopyToClipboard(m.exportResult); err != nil {
			// TODO: Show error message
			return nil
		}
		// TODO: Show success message
		return nil
	}
}

// GetSchemas returns the loaded schemas (for testing)
func (m Model) GetSchemas() *tfjson.ProviderSchemas {
	return m.schemas
}

// View renders the application
func (m Model) View() string {
    if m.showHelp {
        return m.help.View(m.keys)
    }

    if m.stage == StageLoading {
        return m.renderLoadingView()
    }

    if m.stage == StageExportResult {
        return m.renderExportView()
    }

    if m.stage == StageTreeView {
        base := m.renderTreeView()
        if m.exportNamePrompt {
            return lipgloss.JoinVertical(lipgloss.Top, base, m.renderExportNameDialog())
        }
        return base
    }

	return m.renderNavigationView()
}

// renderLoadingView renders a minimal loading screen while schemas are fetched
func (m Model) renderLoadingView() string {
    // Simple centered-ish loading message plus status bar
    msg := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("212")).
        Render("Loading provider schemas…")

    sub := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Render("Please wait while schemas are loaded from cache or Terraform")

    content := lipgloss.JoinVertical(lipgloss.Top, msg, sub)
    // Add status bar at bottom (shows Loading… by default)
    statusView := m.status.Render()
    return lipgloss.JoinVertical(lipgloss.Top, content, statusView)
}

// renderNavigationView renders providers+types stacked on left, entities on right
func (m Model) renderNavigationView() string {
	// Use exact same calculations as updateLayout for consistency
	leftWidth := (m.width / 2) - 4
	rightWidth := (m.width / 2) - 4
	fullHeight := m.height - 4 // Reserve space for status bar
	
	// Calculate heights matching updateLayout logic
	typesHeight := 8
	verticalSpacing := 1
	providersHeight := fullHeight - typesHeight - verticalSpacing
	entitiesHeight := providersHeight + typesHeight + verticalSpacing
	
	// Ensure minimum sizes
	if leftWidth < 10 { leftWidth = 10 }
	if rightWidth < 10 { rightWidth = 10 }
	if providersHeight < 8 { providersHeight = 8 }
	if typesHeight < 6 { typesHeight = 6 }
	if entitiesHeight < 10 { entitiesHeight = 10 }

	// Providers view with focus border and explicit size enforcement
	providersView := m.providers.View()
	if m.focus == FocusProviders {
		providersView = focusedBorderStyle.Width(leftWidth).Height(providersHeight).Render(providersView)
	} else {
		providersView = unfocusedBorderStyle.Width(leftWidth).Height(providersHeight).Render(providersView)
	}

	// Types view with focus border and explicit size enforcement
	typesView := m.types.View()
	if m.focus == FocusTypes {
		typesView = focusedBorderStyle.Width(leftWidth).Height(typesHeight).Render(typesView)
	} else {
		typesView = unfocusedBorderStyle.Width(leftWidth).Height(typesHeight).Render(typesView)
	}

	// Entities view with focus border and explicit size enforcement
	entitiesView := m.entities.View()
	if m.focus == FocusEntities {
		entitiesView = focusedBorderStyle.Width(rightWidth).Height(entitiesHeight).Render(entitiesView)
	} else {
		entitiesView = unfocusedBorderStyle.Width(rightWidth).Height(entitiesHeight).Render(entitiesView)
	}

	// Left column: stack providers and types vertically
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, providersView, typesView)

	// Main content - join left column and entities horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, entitiesView)

	// Add status bar at bottom
	statusView := m.status.Render()

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Top, mainContent, statusView)
}

// renderTreeView renders the 2-pane tree view layout (entities + tree)
func (m Model) renderTreeView() string {
    // Use exact same calculations as updateLayout for consistency
    halfWidth := (m.width / 2) - 2 // Base outer pane width (will include borders)
    // Leave extra headroom to guarantee the top border is visible across terminals.
    // Status bar consumes 1 line; subtract 2 for borders + 1 safety line.
    fullHeight := m.height - 3
	
	// Ensure minimum sizes
	if halfWidth < 10 { halfWidth = 10 }
	if fullHeight < 10 { fullHeight = 10 }

    // Compute inner (content) dimensions accounting for border + padding
    // Border contributes 1 on each side; style uses Padding(0,1) so padX=1, padY=0
    innerWidth := halfWidth - 2 /*borders*/ - 2 /*paddingX*/
    if innerWidth < 1 { innerWidth = 1 }
    innerHeight := fullHeight - 2 /*borders*/ - 0 /*paddingY*/
    if innerHeight < 1 { innerHeight = 1 }

    // Ensure components use inner dimensions so titles and borders align
    m.entities.SetSize(innerWidth, innerHeight)
    m.tree.SetSize(innerWidth, innerHeight)

    // Left pane - entities with focus border and explicit size enforcement
    entitiesView := m.entities.View()
	if m.focus == FocusEntities {
		entitiesView = focusedBorderStyle.Width(halfWidth).Height(fullHeight).Render(entitiesView)
	} else {
		entitiesView = unfocusedBorderStyle.Width(halfWidth).Height(fullHeight).Render(entitiesView)
	}

    // Right pane - tree with focus border and explicit size enforcement
    treeView := m.tree.View()
	if m.focus == FocusTree {
		treeView = focusedBorderStyle.Width(halfWidth).Height(fullHeight).Render(treeView)
	} else {
		treeView = unfocusedBorderStyle.Width(halfWidth).Height(fullHeight).Render(treeView)
	}

	// Join entities and tree horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, entitiesView, treeView)

	// Add status bar at bottom
	statusView := m.status.Render()

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Top, mainContent, statusView)
}

// renderExportView renders the export result view
func (m Model) renderExportView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("Exported HCL")

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(m.exportResult)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press 'c' to copy to clipboard, 'esc' to return")

	return lipgloss.JoinVertical(lipgloss.Top, title, content, hint)
}

// filteredSelectedPaths enforces hierarchical selection: if a parent path is not selected,
// all of its children are considered unselected in the result.
func (m Model) filteredSelectedPaths(paths [][]string) [][]string {
    set := make(map[string]struct{}, len(paths))
    for _, p := range paths {
        set[strings.Join(p, ".")] = struct{}{}
    }
    var out [][]string
    for _, p := range paths {
        okAnc := true
        for i := 1; i < len(p); i++ {
            if _, ok := set[strings.Join(p[:i], ".")]; !ok {
                okAnc = false
                break
            }
        }
        if okAnc {
            out = append(out, p)
        }
    }
    return out
}

// renderExportNameDialog renders a simple centered input dialog for the instance name.
func (m Model) renderExportNameDialog() string {
    boxWidth := 50
    title := lipgloss.NewStyle().Bold(true).Render("Export Outputs: Resource Instance Name")
    prompt := "Enter instance name (default 'main'):"
    input := m.exportName
    if input == "" {
        input = ""
    }
    hint := "Enter to confirm • Esc to cancel"
    body := lipgloss.JoinVertical(lipgloss.Top,
        title,
        prompt,
        "> "+input,
        lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(hint),
    )
    panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(boxWidth).Render(body)
    // Center-ish by adding vertical spacing; exact centering not trivial in plain text
    return panel
}
