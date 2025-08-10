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

	statusHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")). // Brighter than regular values but dimmer than keys
			Italic(true)                       // Subtle visual distinction
)

// StatusBar renders a status bar showing tool info, provider, type, and current filter
type StatusBar struct {
	width           int
	toolInfo        terraform.TerraformInfo
	version         string
	provider        string
	resourceType    string
	filter          string
	copyMessage     string
	copyMessageType string // "success" or "error"
	helpText        string
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

// SetCopyStatus updates the copy status message
func (s *StatusBar) SetCopyStatus(message, messageType string) {
	s.copyMessage = message
	s.copyMessageType = messageType
}

// ClearCopyStatus clears the copy status message
func (s *StatusBar) ClearCopyStatus() {
	s.copyMessage = ""
	s.copyMessageType = ""
}

// SetHelpText updates the help text
func (s *StatusBar) SetHelpText(helpText string) {
	s.helpText = helpText
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

	// Build left side (existing info)
	leftContent := ""
	for i, part := range parts {
		if i > 0 {
			leftContent += statusSeparatorStyle.Render()
		}
		leftContent += part
	}

	// Build center content (copy status)
	centerContent := ""
	if s.copyMessage != "" {
		var copyStyle lipgloss.Style
		if s.copyMessageType == "success" {
			copyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Green
				Bold(true)
		} else {
			copyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")). // Red
				Bold(true)
		}
		centerContent = copyStyle.Render(s.copyMessage)
	}

	// Build right content (help text)
	rightContent := ""
	if s.helpText != "" {
		rightContent = statusHelpStyle.Render(s.helpText)
	}

	// Calculate content widths
	leftWidth := lipgloss.Width(leftContent)
	centerWidth := lipgloss.Width(centerContent)
	rightWidth := lipgloss.Width(rightContent)

	// Use fixed positioning approach to prevent layout shifting
	contentWidth := s.width - 2 // Account for padding

	var finalContent string

	// Check if we have enough space for all content
	minRequiredWidth := leftWidth + centerWidth + rightWidth + 2 // Minimum spacing

	if contentWidth >= minRequiredWidth {
		// Create a buffer of the full content width
		buffer := make([]rune, contentWidth)
		for i := range buffer {
			buffer[i] = ' '
		}

		// Place left content at start
		leftRunes := []rune(leftContent)
		copy(buffer[0:], leftRunes)

		// Place right content at end (fixed position)
		if rightWidth > 0 {
			rightRunes := []rune(rightContent)
			rightStart := contentWidth - rightWidth
			if rightStart >= leftWidth+2 { // Ensure no overlap
				copy(buffer[rightStart:], rightRunes)
			}
		}

		// Place center content at true center. Prioritize visibility: overlay even if it overlaps
		// with left or right details so transient messages (like copy status) are always seen.
		if centerWidth > 0 {
			centerStart := (contentWidth - centerWidth) / 2
			if centerStart < 0 {
				centerStart = 0
			}
			if centerStart+centerWidth > contentWidth {
				centerStart = contentWidth - centerWidth
				if centerStart < 0 {
					centerStart = 0
				}
			}
			centerRunes := []rune(centerContent)
			copy(buffer[centerStart:], centerRunes)
		}

		finalContent = string(buffer)
	} else {
		// Not enough space - fallback to simple layout prioritizing center content
		if centerContent != "" && rightContent != "" {
			// Show center message and right help, skip left details if needed
			finalContent = centerContent + " | " + rightContent
		} else if centerContent != "" {
			// Show left info and center message
			finalContent = leftContent + " " + centerContent
		} else if rightContent != "" {
			// Show left info and right help
			finalContent = leftContent + " | " + rightContent
		} else {
			finalContent = leftContent
		}
	}

	// Apply overall style and fit to width
	styled := statusStyle.Width(s.width).Render(finalContent)
	return styled
}
