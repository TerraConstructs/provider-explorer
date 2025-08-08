# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: CLI entry points and Cobra commands.
- `internal/config/`: Terraform/OpenTofu config detection and validation.
- `internal/terraform/`: Provider discovery, schema loading, caching, binary detection.
- `internal/schema/`: Types and helpers for provider schemas.
- `internal/ui/`: Bubble Tea models, views, styles, and transformations.
- `ui_test/`: Teatest integration flows and snapshot baselines.
- `testdata/`: Minimal fixtures (e.g., `schemas/aws_min.json`).
- `fixtures/`: Example Terraform configs for demos/tests.

## Build, Test, and Development Commands
- `make build`: Build binary via GoReleaser.
- `make install`: Install to Go bin (`$GOBIN`).
- `make fmt`: Format code (gofmt/goimports).
- `make test`: All tests (unit + integration).
- `make test-unit`: Unit tests only.
- `make test-integration`: Teatest integration tests.
- `make test-short`: Quicker subset for iteration.
- `make test-coverage`: Tests with coverage report.
- `make clean` / `make help`: Clean artifacts / list targets.
Note: During development, do not run the TUI manually; validate behavior through tests.

## Coding Style & Naming Conventions
- Language: Go 1.24.4; use `gofmt`/`goimports` (run `make fmt`).
- Packages: lowercase, no underscores; files use descriptive names.
- Exported identifiers: `CamelCase`; unexported: `lowerCamelCase`.
- Tests: `*_test.go`; functions `TestXxx` with clear, behavior-driven names.
- Keep functions small and cohesive; avoid one-letter names except loop indices.

## Testing Guidelines
- Frameworks: `teatest` (integration), `go-snaps` (snapshots), `testify` (assertions).
- Patterns: Always use `teatest.WaitFor(...)`; set `lipgloss.SetColorProfile(0)`.
- Model setup: `ui.NewModelWithSchemas(ps, width, height)`; load schemas via helpers.
- Snapshots: Review diffs in `__snapshots__/`; update only when changes are intentional. Never edit snapshot files manually. To update, run tests with go-snaps env var:
  - `UPDATE_SNAPS=true go test ./...` (update changed snapshots)
  - `UPDATE_SNAPS=always go test ./...` (force create/update, even on CI)
  - `UPDATE_SNAPS=clean go test ./...` (remove obsolete snapshots)
- Quick runs: `make test-short`; full suite: `make test-integration`.

## Commit & Pull Request Guidelines
- Commits: Clear, imperative subject (e.g., "Add filter exact-match test").
- PRs: Describe behavior changes, link issues, summarize test impact, attach relevant snapshot diffs.
- Requirements: All tests green; run `make fmt`; update README/CLAUDE.md if workflows or UX change.

## Security & Configuration Tips
- Providers and schemas may touch local env; never commit secrets.
- Schema cache location: `~/.resource-cache/`.
- Tooling auto-detects Terraform/OpenTofu; ensure binaries are on `PATH`.
