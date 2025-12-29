English | [简体中文](README-ZH.md)

# easyGit

🚀 A modern interactive Git CLI that streamlines your daily Git workflow. Make commits, manage branches, handle tags, and more - all through an intuitive terminal interface.

**🔧 Cross-Platform Support: Linux | macOS | Windows**

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

## ✨ Core Features

### Commit & Push
- **push-all** - Stage and push all changed files with one command
- **push-selected** - Interactively select which files to commit
- **set-push-config** - Save default remote and branch preferences

### Branch & Merge Management
- **merge** - Merge any branch into current branch through interactive selection
- **branch-delete** - Delete local or remote branches safely
- **cherry-pick** - Pick commits from other branches with visual selection

### Tag Operations
- **tag-create** - Create and push tags with semantic versioning support
- **tag-delete** - Remove tags locally and remotely

### Repository Control
- **reset** - Reset to any commit with visual commit history
- **init** - Initialize easyGit configuration for a repository
- **update** - Update easyGit to the latest version

### Customization
- **set-language** - Switch between English and Chinese
- **--language flag** - Override language per-command

## 📦 Installation

easyGit supports **Linux, macOS, and Windows**. Choose your platform below:

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
easyGit push-all

# Push selected files in the working directory
easyGit push-selected

# Initialize easyGit configuration
easyGit init
```

### Advanced Operations

```bash
# Create and push a tag
easyGit tag-create

# Delete a tag
easyGit tag-delete

# Delete local or remote branches
easyGit branch-delete

# Merge selected branch into current branch
easyGit merge

# Cherry-pick commits
easyGit cherry-pick

# Reset to selected commit
easyGit reset

# Configure push settings (remote and branch)
easyGit set-push-config

# Set interface language
easyGit set-language

# Update easyGit
easyGit update
```

## 📖 Command Reference

| Command            | Description                                    |
| ------------------ | ---------------------------------------------- |
| `push-all`         | Stage and commit all changed files             |
| `push-selected`    | Interactively select files to stage and commit |
| `tag-create`       | Create and push tags with semantic versioning  |
| `tag-delete`       | Delete tags locally and remotely               |
| `branch-delete`    | Delete local or remote branches                |
| `merge`            | Merge selected branch into current branch      |
| `cherry-pick`      | Cherry-pick commits from other branches        |
| `reset`            | Reset repository to selected commit            |
| `init`             | Initialize easyGit configuration               |
| `set-push-config`  | Configure default push remote(s) and branch    |
| `set-language`     | Set default language (en/zh)                   |
| `update`           | Update easyGit to latest version               |
| `version`          | Show current version information               |

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

- [Go](https://github.com/golang/go) - The Go programming language
- [Cobra](https://github.com/spf13/cobra) - Powerful CLI applications
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Powerful TUI framework
- [Huh](https://github.com/charmbracelet/huh) - Interactive terminal forms
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UI

---

**Made with ❤️ by [KevinYouu](https://github.com/KevinYouu)**
