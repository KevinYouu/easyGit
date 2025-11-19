# Project Overview

`fastGit` is a modern Command Line Interface (CLI) tool designed to simplify common Git operations through an interactive terminal user interface (TUI). It aims to provide a fast, efficient, and user-friendly experience for developers working with Git. The project supports Linux, macOS, and Windows, offering bilingual support (English and Chinese) with automatic language detection.

## Key Features:
- **Interactive Git Operations**: Utilizes a TUI to select files, branches, and commits.
- **Bilingual Support**: Available in English and Chinese.
- **Modern Interface**: Features themes, animations, and progress indicators.
- **Cross-Platform**: Compatible with Linux, macOS, and Windows.

## Technologies Used:
- **Go**: The primary programming language.
- **Cobra**: For building robust CLI applications.
- **Bubbletea**: A powerful framework for building terminal user interfaces.
- **Huh**: For interactive terminal forms within the TUI.
- **Lipgloss**: For styling and visual definitions in the terminal UI.

## Architecture:
The project is structured with a clear separation of concerns:
- **`cmd/fastgit/`**: Contains the main application entry point, root command setup, and implementations for individual CLI commands.
- **`internal/`**: Houses various internal packages, including:
    - **`gitcmd/`**: Implementations of Git operations.
    - **`i18n/`**: Internationalization support for bilingual features.
    - **`form/`**: Components for TUI forms.
    - **`spinner/`**: Loading animations.
    - **`theme/`**: Visual styling definitions.
    - **`config/`**: Configuration management.

# Building and Running

## Prerequisites:
- Git must be installed on your system.
- Go environment for building from source.

## Build from Source:

To build `fastGit` from its source code, follow these steps:

```bash
# Clone the repository
git clone https://github.com/KevinYouu/fastGit.git
cd fastGit

# Install dependencies
go mod tidy

# Build using the provided script (recommended for version injection)
./build.sh

# Alternatively, build manually
go build -o fastGit ./cmd/fastgit
```

## Running Tests:

To ensure the integrity and functionality of `fastGit`, you can run the provided tests:

```bash
# Run all tests in the project
go test ./...

# Run tests for a specific package (e.g., the internationalization package)
go test ./internal/i18n

# Run benchmarks for a specific package
go test -bench=. ./internal/i18n
```

## Basic Usage:

Once built, you can use `fastGit` with its various commands and aliases:

```bash
# Push all changes in the working directory
fastGit pa

# Push selected files in the working directory
fastGit ps

# Check repository status
fastGit s
```

## Advanced Operations:

```bash
# Create and push a tag
fastGit t

# Delete a tag
fastGit td

# Delete a branch
fastGit bd

# Merge selected branch into current branch
fastGit m

# Cherry-pick commits
fastGit cp

# Reset to selected commit
fastGit rs

# View remote repositories
fastGit rv

# Initialize fastGit configuration
fastGit init

# Update fastGit
fastGit update
```

## Language Support:

`fastGit` automatically detects your system language, but you can override it:

```bash
# Force English
fastGit --language en pa

# Force Chinese
fastGit --language zh pa
```

# Development Conventions

When contributing to `fastGit` or developing new features, please adhere to the following guidelines:

- **Go Conventions**: Follow standard Go conventions and best practices for code style and structure.
- **Testing**: Always add comprehensive tests for new features and bug fixes to maintain code quality and prevent regressions.
- **Documentation**: Keep the documentation up-to-date with any changes in features, commands, or development processes.
- **Bilingual Support**: Ensure that all new commands and features support both English and Chinese for a consistent user experience.

## 📖 Command Reference

| Command | Alias | Description |
|---------|-------|-------------|
| `push-all` | `pa` | Stage and commit all changed files |
| `push-selected` | `ps` | Interactively select files to stage and commit |
| `status` | `s` | Show repository status with enhanced UI |
| `tag` | `t` | Create and push tags with semantic versioning |
| `tag-delete` | `td` | Delete tags locally and remotely |
| `branch-delete` | `bd` | Delete branches locally and remotely |
| `merge` | `m` | Merge selected branch into current branch |
| `cherry-pick` | `cp` | Cherry-pick commits from other branches |
| `reset` | `rs` | Reset repository to selected commit |
| `remotes` | `rv` | List all remote repositories |
| `init` | - | Initialize fastGit configuration |
| `update` | - | Update fastGit to latest version |
| `version` | - | Show current version information |

# Further Exploration:

For more detailed information on specific commands or internal workings, refer to the `docs/` directory and individual package documentations within the `internal/` directory. For example, `docs/UI_ENHANCEMENT.md` might contain details about UI improvements.