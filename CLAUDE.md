# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Building the Project
- **Local Development Build**: `./build.sh` - Builds fastGit with version injection using git tag or dev version
- **Standard Go Build**: `go build -o fastGit ./cmd/fastgit` - Basic build without version injection
- **Release Build**: `goreleaser release --clean` - Creates cross-platform releases (triggered by git tags)

### Testing
- **Run Tests**: `go test ./...` - Runs all tests across the project
- **Run Specific Test**: `go test ./internal/i18n` - Runs tests for specific package
- **Benchmark Tests**: `go test -bench=. ./internal/i18n` - Runs performance benchmarks

### Dependencies
- **Install Dependencies**: `go mod tidy` - Cleans up and installs required dependencies
- **Update Dependencies**: `go get -u ./...` - Updates all dependencies

## Project Architecture

### Core Structure
fastGit is a CLI tool written in Go that simplifies Git operations through an interactive TUI (Terminal User Interface). The project follows a clean modular architecture:

### Key Directories
- `cmd/fastgit/` - Main application entry point and CLI commands
  - `main.go` - Application entry point with language detection
  - `root.go` - Root command setup and subcommand registration
  - `commands/` - Individual command implementations (push_all.go, merge.go, etc.)
- `internal/` - Internal packages (not exported)
  - `gitcmd/` - Core Git operation implementations
  - `i18n/` - Internationalization support (English/Chinese)
  - `form/` - TUI form components (input, select, multi-select)
  - `spinner/` - Loading animations and progress indicators
  - `theme/` - Visual styling and color themes
  - `config/` - Configuration management
  - `command/` - Command execution utilities

### Architecture Patterns
- **Command Pattern**: Each Git operation is encapsulated as a separate command in `cmd/fastgit/commands/`
- **MVC-like Separation**: Clear separation between CLI commands, business logic (internal packages), and data models
- **Dependency Injection**: Commands depend on internal packages rather than direct implementations
- **Internationalization**: All UI strings go through the i18n system with language detection

### Core Technologies
- **CLI Framework**: Cobra for command-line interface structure
- **TUI Framework**: Bubbletea for terminal user interface
- **Form Components**: Huh for interactive forms
- **Styling**: Lipgloss for terminal styling and theming
- **Git Operations**: Direct git command execution through os/exec

### Key Features
- Interactive file selection for Git operations
- Multi-step Git workflows with progress tracking
- Bilingual support (English/Chinese) with automatic detection
- Modern TUI with themes, animations, and progress indicators
- Cross-platform support (Linux, macOS, Windows)

### Command Flow
1. `main.go` detects language from environment or flags
2. `root.go` sets up command structure with translated descriptions
3. Individual commands in `commands/` use internal packages for:
   - Git operations via `internal/gitcmd`
   - UI components via `internal/form` and `internal/spinner`
   - Configuration via `internal/config`

### Configuration
- Uses SQLite for storing configuration and usage data
- Supports version injection during build process
- Theme customization through `internal/theme` package

## Development Notes

### Language Support
- All user-facing strings must use `i18n.T()` function
- Translation keys are stored in `internal/i18n/en.go` and `internal/i18n/zh.go`
- Language can be set via command-line flags (`-l`, `--language`) or detected automatically

### Git Operations
- Git commands are executed through the `internal/gitcmd` package
- Each Git operation (push, merge, reset, etc.) has its own implementation file
- Commands support both interactive and non-interactive modes

### Testing Strategy
- Unit tests exist for i18n functionality with benchmarks
- Performance testing ensures translation speed (< 10ms for 10000 translations)
- Test files follow Go conventions with `_test.go` suffix

### UI/UX Principles
- All interactive elements use modern TUI components
- Progress tracking for long-running operations
- Consistent theming and color schemes
- Accessibility through keyboard navigation

## Common Development Tasks

### Adding a New Command
1. Create command file in `cmd/fastgit/commands/`
2. Implement command function returning `*cobra.Command`
3. Add translation keys in `internal/i18n/en.go` and `internal/i18n/zh.go`
4. Register command in `cmd/fastgit/root.go`
5. Implement core logic in appropriate `internal/` package

### Modifying Git Operations
- Core Git logic is in `internal/gitcmd/` directory
- Each operation has its own file (e.g., `pushAll.go`, `merge.go`)
- Follow existing patterns for error handling and user feedback

### Updating UI/Theme
- Modify `internal/theme/theme.go` for visual changes
- Form components are in `internal/form/`
- Spinner animations are in `internal/spinner/`

### Release Process
- Tag a commit with version number (e.g., `v1.2.3`)
- GitHub Actions trigger GoReleaser automatically
- Cross-platform binaries are created and attached to release