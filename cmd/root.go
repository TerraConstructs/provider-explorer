package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/terraconstructs/provider-explorer/internal/config"
	"github.com/terraconstructs/provider-explorer/internal/terraform"
	"github.com/terraconstructs/provider-explorer/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "provider-explorer [path]",
	Short: "Interactive Terraform provider resource explorer",
	Long: `Provider Explorer is a TUI application that helps you explore Terraform provider
resources and their schemas. It provides fuzzy search and interactive navigation
through providers, resource types, and detailed resource schemas.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	workingDir := "."
	if len(args) > 0 {
		workingDir = args[0]
	}

	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if !config.HasTerraformConfig(absPath) {
		return fmt.Errorf("no Terraform configuration found in %s", absPath)
	}

	// Check if we have a valid cache first
	if !terraform.HasValidProviderCache(absPath) {
		// Only prompt for init if no valid cache exists
		if err := config.InitTerraformDirectory(absPath); err != nil {
			return err
		}
    }

	// Change to the working directory so the UI can load schemas
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Create model with default terminal size (will be updated by tea.WindowSizeMsg)
	model := ui.NewModel(80, 24)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}
