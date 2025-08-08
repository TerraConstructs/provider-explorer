package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/terraconstructs/provider-explorer/internal/schema"
)

func (m Model) updateProviderSelection(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedProvider > 0 {
			m.selectedProvider--
		}
	case "down", "j":
		if m.selectedProvider < len(m.providers)-1 {
			m.selectedProvider++
		}
	case "enter":
		if len(m.providers) > 0 && m.allSchemas != nil {
			m.state = StateResourceTypeSelection
			m.resourceTypes = []string{"Data Sources", "Resources"}
			m.selectedResourceType = 0
		}
	}
	return m, nil
}

func (m Model) updateResourceTypeSelection(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedResourceType > 0 {
			m.selectedResourceType--
		}
	case "down", "j":
		if m.selectedResourceType < len(m.resourceTypes)-1 {
			m.selectedResourceType++
		}
	case "enter":
		if len(m.resourceTypes) > 0 && m.allSchemas != nil {
			provider := m.providers[m.selectedProvider]
			
			// Find the provider in the schema - it might have a full name like registry.terraform.io/hashicorp/aws
			var providerKey string
			for key := range m.allSchemas.ProviderSchemas {
				if strings.HasSuffix(key, "/"+provider.Name) || key == provider.Name {
					providerKey = key
					break
				}
			}
			
			if providerKey == "" {
				return m, nil
			}
			
			resourceType := m.resourceTypes[m.selectedResourceType]
			var resources []string
			
			switch resourceType {
			case "Data Sources":
				for name := range m.allSchemas.ProviderSchemas[providerKey].DataSourceSchemas {
					resources = append(resources, name)
				}
			case "Resources":
				for name := range m.allSchemas.ProviderSchemas[providerKey].ResourceSchemas {
					resources = append(resources, name)
				}
			}
			
			sort.Strings(resources)
			m.resources = resources
			m.selectedResource = 0
			m.state = StateResourceDetailSelection
		}
	}
	return m, nil
}

func (m Model) updateResourceDetailSelection(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedResource > 0 {
			m.selectedResource--
		}
	case "down", "j":
		if m.selectedResource < len(m.resources)-1 {
			m.selectedResource++
		}
	case "enter":
		if len(m.resources) > 0 && m.allSchemas != nil {
			provider := m.providers[m.selectedProvider]
			resourceName := m.resources[m.selectedResource]
			resourceType := m.resourceTypes[m.selectedResourceType]
			
			// Find the provider key in the schema
			var providerKey string
			for key := range m.allSchemas.ProviderSchemas {
				if strings.HasSuffix(key, "/"+provider.Name) || key == provider.Name {
					providerKey = key
					break
				}
			}
			
			if providerKey == "" {
				return m, nil
			}
			
			var resourceSchema *schema.ResourceSchema
			
			if strings.Contains(resourceType, "Data") {
				if ds, exists := m.allSchemas.ProviderSchemas[providerKey].DataSourceSchemas[resourceName]; exists {
					resourceSchema = &ds
				}
			} else {
				if rs, exists := m.allSchemas.ProviderSchemas[providerKey].ResourceSchemas[resourceName]; exists {
					resourceSchema = &rs
				}
			}
			
			if resourceSchema != nil {
				m.currentSchema = resourceSchema
				m.state = StateResourceDetail
			}
		}
	}
	return m, nil
}

func (m Model) updateResourceDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	return m, nil
}