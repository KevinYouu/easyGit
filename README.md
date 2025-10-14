English | [简体中文](README-CN.md)

# fastGit

🚀 A modern CLI tool that simplifies Git operations with an interactive terminal user interface. Supports Linux, macOS, and Windows.

> This project uses its own features to commit code - dogfooding at its finest!

![fastGit Demo](assets/fast-git.gif)

## ✨ Features

- 🎯 **Interactive Git Operations** - Select files, branches, and commits with a beautiful TUI
- 🌍 **Bilingual Support** - English and Chinese with automatic language detection
- 🎨 **Modern Interface** - Beautiful themes, animations, and progress indicators
- ⚡ **Fast & Efficient** - Streamlined workflows for common Git operations
- 🔧 **Cross-Platform** - Works on Linux, macOS, and Windows

## 📦 Installation

### Prerequisites

- [Git](https://git-scm.com/) must be installed on your system

### Quick Install (Recommended)

```bash
# Linux/macOS
curl -sSL https://raw.githubusercontent.com/KevinYouu/fastGit/main/install.sh | bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/KevinYouu/fastGit/main/install.sh | bash
```

```powershell
# Windows
iwr -useb https://raw.githubusercontent.com/KevinYouu/fastGit/main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/KevinYouu/fastGit.git
cd fastGit
./build.sh
```

## 🚀 Quick Start

### Basic Usage

```bash
# Push all changes in the working directory
fastGit pa

# Push selected files in the working directory
fastGit ps

# Check repository status
fastGit s
```

### Advanced Operations

```bash
# Create and push a tag
fastGit t

# Delete a tag
fastGit td

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

## 📖 Command Reference

| Command | Alias | Description |
|---------|-------|-------------|
| `push-all` | `pa` | Stage and commit all changed files |
| `push-selected` | `ps` | Interactively select files to stage and commit |
| `status` | `s` | Show repository status with enhanced UI |
| `tag` | `t` | Create and push tags with semantic versioning |
| `tag-delete` | `td` | Delete tags locally and remotely |
| `merge` | `m` | Merge selected branch into current branch |
| `cherry-pick` | `cp` | Cherry-pick commits from other branches |
| `reset` | `rs` | Reset repository to selected commit |
| `remotes` | `rv` | List all remote repositories |
| `init` | - | Initialize fastGit configuration |
| `update` | - | Update fastGit to latest version |
| `version` | - | Show current version information |

### Language Support

fastGit automatically detects your system language or you can specify it manually:

```bash
# Force English
fastGit --language en pa

# Force Chinese
fastGit --language zh pa
```

## 🛠️ Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/KevinYouu/fastGit.git
cd fastGit

# Install dependencies
go mod tidy

# Build with version injection
./build.sh

# Or build manually
go build -o fastGit ./cmd/fastgit
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
fastGit/
├── cmd/fastgit/          # CLI commands and entry point
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
