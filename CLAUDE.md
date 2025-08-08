# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

provider-explorer CLI is a Golang TUI (Terminal User Interface) application that helps developers work with Terraform provider resource schemas. The application uses the Charm Bracelet Bubbletea framework for the TUI and provides fuzzy autocomplete functionality.

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

## Testing Strategy

**CRITICAL: Testing is the ONLY way to validate this TUI application. Manual execution is PROHIBITED during development.**

This application uses comprehensive snapshot-based testing and teatest integration testing to ensure UI reliability and prevent visual regressions.

### Mandatory Development Workflow

The development process MUST follow this strict sequence:

1. **Define Workflow Changes**: Clearly specify what UI behavior changes are required
2. **Review Existing Tests**: Identify which teatest workflows need updates based on changes
3. **Update Test Workflows**: Modify/add teatest integration tests in `ui_test/` directory
4. **Review Snapshots**: Examine snapshot changes to ensure no unexpected visual regressions
5. **Build Application**: Run `make build` to ensure compilation succeeds
6. **Format Code**: Run `make fmt` to format all Go code
7. **NEVER RUN MANUALLY**: Do not attempt to run `./provider-explorer` in the shell

### Snapshot-Based Testing

The application uses two types of snapshot testing:

#### 1. Component-Level Snapshots (`internal/ui/tree/model_test.go`)
- Tests individual tree component rendering with `snaps.MatchSnapshot(t, got)`
- Captures tree structure, selection indicators, scrolling, and hidden node behavior
- Uses `testutil.RunModel()` helper for component isolation

#### 2. Full UI Baseline Snapshots (`ui_test/ui_snapshots_baseline_test.go`) 
- Captures complete application UI output for visual regression detection
- Tests multiple screen sizes: 120x30, 80x24, 160x40
- Includes navigation views and tree views across different dimensions
- Critical for detecting layout issues and responsive behavior changes

### Teatest Integration Testing Rules

**31 comprehensive integration tests** in `ui_test/` directory cover all UI workflows:

#### Mandatory Patterns:
- **ALWAYS use `teatest.WaitFor()`** - Never read output directly
- **Proper async handling** - Wait for UI state changes before proceeding
- **Full model initialization** - Use `ui.NewModelWithSchemas(ps, width, height)`
- **Consistent setup** - Load schemas with `ui.LoadProvidersSchemaFromFile()`
- **Color profile stability** - Set `lipgloss.SetColorProfile(0)` for consistent output

#### Key Test Categories:

**Navigation Tests:**
- `ui_snapshots_baseline_test.go`: Complete navigation workflows
- `debug_navigation_test.go`: Navigation edge cases and debugging
- `tree_navigation_test.go`: Tree view navigation patterns

**Filtering Tests:**
- 8+ filtering tests covering exact matches, step-by-step filtering, AWS-specific filtering
- `filter_test.go`, `filter_exact_test.go`, `filter_with_enter_test.go`, etc.
- All use `teatest.WaitFor()` with proper byte content validation

**Flow Tests:**
- `simple_flow_test.go`: Basic user workflows without export
- `flow_test.go`: Complete workflows including transformations  
- `minimal_flow_test.go`: Minimal viable user paths

**Export/Transformation Tests:**
- `export_test.go`: HCL transformation functionality (unit tests, not teatest)
- Validates arguments → variables and attributes → outputs conversions

#### Critical Anti-Patterns - NEVER DO:
- ❌ Manual application execution (`./provider-explorer`)
- ❌ Direct output reading without `teatest.WaitFor()`
- ❌ Incomplete model initialization
- ❌ Ignoring async operations
- ❌ Removing or skipping integration tests
- ❌ Testing without proper schema fixtures

#### Teatest Best Practices:
```go
// ✅ CORRECT: Always wait for UI state
teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
    return bytes.Contains(b, []byte("press / to filter"))
}, teatest.WithDuration(5*time.Second))

// ❌ WRONG: Direct output reading
output := tm.Output()
buf := make([]byte, 8192)
n, _ := output.Read(buf) // This is unreliable for UI state
```

### Test Execution

Use these commands for different test scenarios:

```bash
# All tests (including integration)
make test

# Unit tests only (fast, skips teatest)  
make test-unit

# Integration tests only (teatest)
make test-integration

# Quick tests for development
make test-short

# Tests with coverage report
make test-coverage
```

### Snapshot Updates

When legitimate UI changes occur, snapshot files will need updates:
- Review changes carefully in `__snapshots__/` directories
- Ensure changes align with intended modifications
- Commit updated snapshots alongside code changes

### Test Data

Tests use minimal fixtures in `testdata/schemas/aws_min.json` containing:
- `aws_instance` resource with required/optional attributes
- `aws_s3_bucket` resource for filtering tests
- Consistent schema structure for reproducible testing

This testing approach ensures complete UI coverage while maintaining development velocity through reliable automated validation.
