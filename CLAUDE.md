# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Get Resource CLI is a Golang TUI (Terminal User Interface) application that helps developers work with Terraform provider resource schemas. The application uses the Charm Bracelet Bubbletea framework for the TUI and provides fuzzy autocomplete functionality.

## Architecture

The application follows a multi-step workflow:

1. **Provider Discovery**: Detects Terraform/OpenTofu configuration and ensures providers are initialized
2. **Schema Caching**: Loads provider schemas from cache (~/.resource-cache/) or fetches via `terraform providers schema`
3. **Interactive Navigation**: Provides a TUI with fuzzy matching to navigate providers → resource types → specific resources
4. **Transformations**: Offers built-in transformations like "convert to HCL"

### Key Components

- **Main Command**: Takes a path parameter (defaults to PWD) and validates Terraform configuration
- **Init Subcommand**: Initializes directories with popular provider configurations (AWS, GCP, Azure)
- **TUI Navigation**: Four main sections per provider:
  - Data Sources
  - Ephemeral Resources  
  - Provider Functions
  - Resources
- **Schema Representation**: Converts nested provider schemas to displayable string format
- **HCL Transformations**: 
  - Arguments → Terraform variable blocks
  - Attributes → Terraform output blocks

## Development Commands

```bash
# Build using GoReleaser (recommended)
make build

# Install to Go bin
make install

# Format code
make fmt

# Run tests
make test

# Clean build artifacts
make clean

# Show all available targets
make help

# Direct Go commands (alternative)
go build -o provider-explorer .
go test ./...
go mod tidy
```

## Usage

The CLI takes an optional path argument (defaults to current directory):

```bash
# Use current directory
./provider-explorer

# Use specific directory with Terraform configuration
./provider-explorer ./fixtures/example-aws

# Get help
./provider-explorer --help
```

The application requires a valid Terraform configuration in the specified directory.

## Key Dependencies

Based on the README, the project uses:
- **github.com/charmbracelet/bubbletea**: For the TUI framework
- Standard Go libraries for Terraform/OpenTofu integration
- JSON handling for provider schema parsing

## Cache Management

The application maintains a provider schema cache in `~/.resource-cache/` to avoid repeated `terraform providers schema` calls.

## Terraform Integration

The CLI integrates directly with Terraform/OpenTofu by:
- Detecting available flavors (Terraform vs OpenTofu, defaulting to tofu)
- Running `init` command via `os.exec` to download providers
- Invoking `terraform providers schema` to get JSON schema data