package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

var (
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("240")).
			Padding(0, 1)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	statusValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	statusSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				SetString(" | ")
)

// StatusBar renders a status bar showing tool info, provider, type, and current filter
type StatusBar struct {
	width        int
	toolInfo     terraform.TerraformInfo
	version      string
	provider     string
	resourceType string
	filter       string
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int) StatusBar {
	return StatusBar{
		width: width,
	}
}

// SetWidth updates the status bar width
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetToolInfo updates the tool information
func (s *StatusBar) SetToolInfo(toolInfo terraform.TerraformInfo, version string) {
	s.toolInfo = toolInfo
	s.version = version
}

// SetProvider updates the current provider
func (s *StatusBar) SetProvider(provider string) {
	s.provider = provider
}

// SetResourceType updates the current resource type
func (s *StatusBar) SetResourceType(resourceType string) {
	s.resourceType = resourceType
}

// SetFilter updates the current filter
func (s *StatusBar) SetFilter(filter string) {
	s.filter = filter
}

// Render renders the status bar
func (s StatusBar) Render() string {
	var parts []string

	// Tool and version
	if s.toolInfo.Tool != "" {
		toolText := statusKeyStyle.Render(s.toolInfo.Tool)
		if s.version != "" {
			toolText += " " + statusValueStyle.Render(s.version)
		}
		parts = append(parts, toolText)
	}

	// Provider
	if s.provider != "" {
		providerText := statusKeyStyle.Render("provider") + "=" + statusValueStyle.Render(s.provider)
		parts = append(parts, providerText)
	}

	// Resource type
	if s.resourceType != "" {
		typeText := statusKeyStyle.Render("type") + "=" + statusValueStyle.Render(s.resourceType)
		parts = append(parts, typeText)
	}

	// Filter
	if s.filter != "" {
		filterText := statusKeyStyle.Render("filter") + ": " + statusValueStyle.Render(fmt.Sprintf(`"%s"`, s.filter))
		parts = append(parts, filterText)
	}

	if len(parts) == 0 {
		parts = append(parts, statusValueStyle.Render("Loading..."))
	}

	// Join parts with separator
	content := ""
	for i, part := range parts {
		if i > 0 {
			content += statusSeparatorStyle.Render()
		}
		content += part
	}

	// Apply overall style and fit to width
	styled := statusStyle.Width(s.width).Render(content)
	return styled
}