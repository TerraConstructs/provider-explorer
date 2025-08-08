package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EE6FF8"))

	dimmedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87"))
)

func (m Model) renderProviderSelection() string {
	if len(m.providers) == 0 {
		return titleStyle.Render("Provider Explorer") + "\n\n" +
			"Loading providers..." + "\n\n" +
			dimmedStyle.Render("Press q to quit")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Select Provider"))
	b.WriteString("\n\n")

	for i, provider := range m.providers {
		cursor := " "
		if i == m.selectedProvider {
			cursor = selectedStyle.Render(">")
		}

		line := fmt.Sprintf("%s %s", cursor, provider.Name)
		if provider.Source != "" {
			line += dimmedStyle.Render(fmt.Sprintf(" (%s)", provider.Source))
		}
		if provider.Version != "" {
			line += dimmedStyle.Render(fmt.Sprintf(" %s", provider.Version))
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("↑/↓: navigate • enter: select • q: quit"))

	return b.String()
}

func (m Model) renderResourceTypeSelection() string {
	if len(m.resourceTypes) == 0 {
		return titleStyle.Render("Loading Resource Types") + "\n\n" +
			"Loading..." + "\n\n" +
			dimmedStyle.Render("Press esc to go back • q to quit")
	}

	var b strings.Builder
	currentProvider := m.providers[m.selectedProvider]
	b.WriteString(titleStyle.Render(fmt.Sprintf("Resource Types - %s", currentProvider.Name)))
	b.WriteString("\n\n")

	for i, resourceType := range m.resourceTypes {
		cursor := " "
		if i == m.selectedResourceType {
			cursor = selectedStyle.Render(">")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, resourceType))
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

func (m Model) renderResourceDetailSelection() string {
	if len(m.resources) == 0 {
		return titleStyle.Render("Loading Resources") + "\n\n" +
			"Loading..." + "\n\n" +
			dimmedStyle.Render("Press esc to go back • q to quit")
	}

	var b strings.Builder
	currentProvider := m.providers[m.selectedProvider]
	resourceType := m.resourceTypes[m.selectedResourceType]
	
	b.WriteString(titleStyle.Render(fmt.Sprintf("%s - %s", currentProvider.Name, resourceType)))
	b.WriteString("\n\n")

	for i, resource := range m.resources {
		cursor := " "
		if i == m.selectedResource {
			cursor = selectedStyle.Render(">")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, resource))
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

func (m Model) renderResourceDetail() string {
	if m.currentSchema == nil {
		return titleStyle.Render("Loading Resource Detail") + "\n\n" +
			"Loading..." + "\n\n" +
			dimmedStyle.Render("Press esc to go back • q to quit")
	}

	var b strings.Builder
	currentProvider := m.providers[m.selectedProvider]
	resource := m.resources[m.selectedResource]
	
	b.WriteString(titleStyle.Render(fmt.Sprintf("%s - %s", currentProvider.Name, resource)))
	b.WriteString("\n\n")

	if m.currentSchema.Block.Description != "" {
		b.WriteString(m.currentSchema.Block.Description)
		b.WriteString("\n\n")
	}

	b.WriteString("Arguments:\n")
	for name, attr := range m.currentSchema.Block.Attributes {
		if attr.Required || attr.Optional {
			marker := ""
			if attr.Required {
				marker = selectedStyle.Render("*")
			}
			b.WriteString(fmt.Sprintf("  %s%s", marker, name))
			if attr.Description != "" {
				b.WriteString(dimmedStyle.Render(fmt.Sprintf(" - %s", attr.Description)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\nAttributes:\n")
	for name, attr := range m.currentSchema.Block.Attributes {
		if attr.Computed {
			b.WriteString(fmt.Sprintf("  %s", name))
			if attr.Description != "" {
				b.WriteString(dimmedStyle.Render(fmt.Sprintf(" - %s", attr.Description)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimmedStyle.Render("esc: back • q: quit"))

	return b.String()
}

func fuzzyFilter(input string, choices []string) []string {
	if input == "" {
		return choices
	}

	matches := fuzzy.Find(input, choices)
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.Str
	}
	return result
}