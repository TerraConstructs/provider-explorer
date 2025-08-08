package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/terraconstructs/provider-explorer/internal/config"
	"github.com/terraconstructs/provider-explorer/internal/schema"
	"github.com/terraconstructs/provider-explorer/internal/terraform"
)

type ProvidersDetectedMsg struct {
	providers []config.ProviderInfo
	err       error
}

type AllSchemasLoadedMsg struct {
	schema *schema.ProviderSchema
	err    error
}

func loadAllSchemasCmd(workingDir string) tea.Cmd {
	return func() tea.Msg {
		// Get the complete schema
		completeSchema, err := terraform.FetchAllProviderSchemas(workingDir)
		return AllSchemasLoadedMsg{
			schema: completeSchema,
			err:    err,
		}
	}
}

func detectProvidersCmd(workingDir string) tea.Cmd {
	return func() tea.Msg {
		providers, err := config.GetInstalledProviders(workingDir)
		return ProvidersDetectedMsg{
			providers: providers,
			err:       err,
		}
	}
}