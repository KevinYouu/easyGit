English | [简体中文](README-ZH.md)

# easyGit

🚀 A modern interactive Git CLI that streamlines your daily Git workflow. Make commits, manage branches, handle tags, and more - all through an intuitive terminal interface. Supports Linux/macOS/Windows

> This project uses its own features to commit code - dogfooding at its finest!

![easyGit Demo](assets/easygit.gif)

## 📖 What is easyGit?

**easyGit is designed for developers who love Git but want to work faster.**

Instead of typing multiple Git commands or remembering complex flags, easyGit provides:

- ⚡ **One-Command Workflows** - Push all changes or select specific files with a single command
- 🎯 **Interactive Selection** - Choose branches, commits, tags, and files through visual menus
- 🔧 **Smart Configuration** - Save your push preferences (remote/branch) and language settings
- 🌍 **Multilingual** - Full English and Chinese support with auto-detection
- 🎨 **Beautiful Interface** - Clear visual feedback with modern TUI components
- 💻 **Cross-Platform** - Works seamlessly on Linux, macOS, and Windows

## 📖 Design Philosophy

**easyGit is a Git workflow enhancement tool, not a replacement for Git commands.**

We focus on:

- 🎯 **Integrating multi-step workflows** - Combining multiple Git commands into one interactive flow
- 🧠 **Simplifying complex commands** - Providing friendly UIs for hard-to-remember parameters
- 🚀 **Optimizing repetitive work** - Batch operations and smart defaults
- ✨ **Enhancing user experience** - Beautiful TUI with clear visual feedback

📚 Read more: [Design Philosophy](docs/DESIGN_PHILOSOPHY.md) | [设计哲学 (中文)](docs/DESIGN_PHILOSOPHY_ZH.md)

## 📦 Installation

### Prerequisites

- [Git](https://git-scm.com/) must be installed on your system

### Quick Install (Recommended)

**Linux / macOS:**

```bash
# Using curl
curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
```

**Windows:**

```powershell
# Using PowerShell
iwr -useb https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1 | iex
```

### Build from Source

Works on all platforms with Go installed:

```bash
git clone https://github.com/KevinYouu/easyGit.git
cd easyGit
./build.sh
```

## 🚀 Quick Start

### Basic Usage

```bash
# Push all changes in the working directory
easyGit push-all  # or: easyGit pa

# Push selected files in the working directory
easyGit push-selected  # or: easyGit ps

# Initialize easyGit configuration (usually auto-initialized on first run)
easyGit init
```

### Advanced Operations

```bash
# Create and push a tag
easyGit tag-create  # or: easyGit tc

# Delete a tag
easyGit tag-delete  # or: easyGit td

# Delete local or remote branches
easyGit branch-delete  # or: easyGit bd

# Merge selected branch into current branch
easyGit merge  # or: easyGit m

# Cherry-pick commits
easyGit cherry-pick  # or: easyGit cp

# Reset to selected commit
easyGit reset  # or: easyGit rs

# Configure push settings (remote and branch)
easyGit set-push-config

# Set interface language
easyGit set-language

# Update easyGit
easyGit update
```

## 📖 Command Reference

| Command           | Alias | Description                                    |
| ----------------- | ----- | ---------------------------------------------- |
| `push-all`        | `pa`  | Stage and commit all changed files             |
| `push-selected`   | `ps`  | Interactively select files to stage and commit |
| `tag-create`      | `tc`  | Create and push tags with semantic versioning  |
| `tag-delete`      | `td`  | Delete tags locally and remotely               |
| `branch-delete`   | `bd`  | Delete local or remote branches                |
| `merge`           | `m`   | Merge selected branch into current branch      |
| `cherry-pick`     | `cp`  | Cherry-pick commits from other branches        |
| `reset`           | `rs`  | Reset repository to selected commit            |
| `init`            | -     | Initialize easyGit configuration               |
| `set-push-config` | -     | Configure default push remote(s) and branch    |
| `set-language`    | -     | Set default language (en/zh)                   |
| `update`          | -     | Update easyGit to latest version               |
| `version`         | `v`   | Show current version information               |

### Language Support

easyGit automatically detects your system language or you can specify it manually:

```bash
# Force English
easyGit --language en push-all

# Force Chinese
easyGit --language zh push-all
```

## 🛠️ Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/KevinYouu/easyGit.git
cd easyGit

# Install dependencies
go mod tidy

# Build with version injection
./build.sh

# Or build manually
go build -o easyGit ./cmd/easygit
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/i18n

# Run benchmarks
go test -bench=. ./internal/i18n
```

### Project Structure

```
easyGit/
├── cmd/easygit/          # CLI commands and entry point
│   ├── main.go          # Application entry point
│   ├── root.go          # Root command setup
│   └── commands/        # Individual command implementations
├── internal/            # Internal packages
│   ├── gitcmd/          # Git operation implementations
│   ├── i18n/            # Internationalization
│   ├── form/            # TUI form components
│   ├── spinner/         # Loading animations
│   ├── theme/           # Visual styling
│   └── config/          # Configuration management
└── docs/                # Documentation
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new features
- Update documentation as needed
- Ensure all commands support both English and Chinese

## 📄 License

This project is licensed under the [LICENSE](LICENSE) file.

## 🙏 Acknowledgments

Built with these amazing open source projects:

### Core Dependencies

- [Go](https://github.com/golang/go) - The Go programming language
- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Powerful TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components for Bubbletea
- [Huh](https://github.com/charmbracelet/huh) - Interactive terminal forms
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UI
- [SQLite](https://gitlab.com/cznic/sqlite) - Pure Go SQLite implementation for configuration storage
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - Terminal utilities
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) - Text processing and i18n support

---

**Made with ❤️ by [KevinYouu](https://github.com/KevinYouu)**
