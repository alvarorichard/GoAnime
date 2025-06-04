# Developer's Guide

Welcome to the GoAnime development guide! This document outlines the development
workflow, coding standards, and best practices for contributing to the GoAnime
project.

## Table of Contents

- [Branching Strategy](#branching-strategy)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Quality Assurance](#quality-assurance)
- [Testing](#testing)
- [Build Process](#build-process)
- [Contributing Guidelines](#contributing-guidelines)

## Branching Strategy

### Branch Structure

- **`main`**: Production-ready code. This branch should always be stable and deployable.
- **`dev`**: Development branch where all features are integrated before merging
  to main.
- **`feature/*`**: Feature branches for new functionality (e.g.,
  `feature/anime-search`, `feature/discord-integration`).
- **`bugfix/*`**: Bug fix branches (e.g., `bugfix/player-crash`, `bugfix/episode-loading`).
- **`hotfix/*`**: Critical fixes that need to go directly to main.

### Important Rules

⚠️ **NEVER commit directly to the `main` branch!**

- All changes must go through the `dev` branch first
- Create feature branches from `dev`
- Merge feature branches back to `dev`
- Only merge `dev` to `main` after thorough testing

## Development Workflow

### 1. Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/alvarorichard/GoAnime.git
cd GoAnime

# Switch to dev branch
git checkout dev

# Install dependencies (if using Nix)
nix-shell

# Or install Go dependencies
go mod tidy
```

### 2. Creating a New Feature

```bash
# Start from the latest dev branch
git checkout dev
git pull origin dev

# Create a new feature branch
git checkout -b feature/your-feature-name

# Make your changes...
# Commit your changes
git add .
git commit -m "feat: add your feature description"

# Push to your feature branch
git push origin feature/your-feature-name
```

### 3. Submitting Changes

1. Create a Pull Request from your feature branch to `dev`
2. Ensure all checks pass (tests, linting, formatting)
3. Request code review from maintainers
4. Address any feedback
5. Merge to `dev` after approval

## Code Standards

### Go Formatting

We use the standard Go formatter. **Always run `go fmt` before committing:**

```bash
# Format all Go files in the project
go fmt ./...

# Or format specific files
go fmt internal/player/player.go
```

### Code Style Guidelines

1. **Follow Go conventions**: Use `gofmt`, follow Go naming conventions
2. **Documentation**: Add comments for exported functions and types
3. **Error handling**: Always handle errors appropriately
4. **Package organization**: Keep packages focused and cohesive

### Code Documentation (Recommended)

While not mandatory, it's **highly recommended** to document your code with
meaningful comments:

```go
// Good: Descriptive comment explaining the purpose
// ParseEpisodeNumber extracts the episode number from a filename or URL.
// It supports various naming conventions like "Episode 01", "Ep1", "S01E05", etc.
// Returns -1 if no episode number is found.
func ParseEpisodeNumber(input string) int {
    // Implementation details...
}

// Good: Comment explaining complex logic
// Check if the anime is already in our local database
// to avoid unnecessary API calls
if exists := checkLocalCache(animeID); exists {
    return getCachedAnime(animeID)
}

// Avoid: Obvious comments
// i++ // increment i by 1
```

**Documentation Guidelines:**

- Document **what** the code does, not **how** it does it
- Explain **why** certain decisions were made
- Document edge cases and expected inputs/outputs
- Use proper Go doc comments for exported functions
- Comment complex algorithms or business logic

### Example Code Style

```go
// Package player provides video playback functionality for GoAnime.
package player

import (
    "context"
    "fmt"
    "log"
)

// Player represents a video player instance.
type Player struct {
    config Config
    state  State
}

// NewPlayer creates a new player instance with the given configuration.
func NewPlayer(config Config) (*Player, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &Player{
        config: config,
        state:  StateIdle,
    }, nil
}
```

## Quality Assurance

### Automated Code Quality Checks

We use automated quality verification bots and tools:

1. **Go Linting**: `golangci-lint` for comprehensive code analysis
2. **Code Coverage**: Maintain minimum test coverage
3. **Security Scanning**: Automated vulnerability detection
4. **Dependency Updates**: Automated dependency security updates

### Pre-commit Checks

Before committing, ensure your code passes:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test ./...

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

### CI/CD Pipeline

Our continuous integration pipeline automatically:

- Runs `go fmt` checks
- Executes linting with `golangci-lint`
- Runs the full test suite
- Checks code coverage
- Performs security scans
- Builds for multiple platforms

## Testing

### Testing Strategy

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete user workflows

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/player/

# Run tests with verbose output
go test -v ./...
```

### Test File Organization

- Test files should be named `*_test.go`
- Place tests in the same package as the code being tested
- Use table-driven tests for multiple test cases

### Example Test

```go
func TestPlayer_Play(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid video file",
            input:   "test.mp4",
            wantErr: false,
        },
        {
            name:    "invalid file",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewPlayer(DefaultConfig())
            err := p.Play(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Play() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Build Process

### Local Development Build

```bash
# Build for current platform
go build -o bin/goanime cmd/goanime/main.go

# Run the application
./bin/goanime
```

### Cross-Platform Builds

Use the provided build scripts:

```bash
# Linux build
./build/buildlinux.sh

# Linux with SQLite
./build/buildlinux-with-sqlite.sh

# macOS build
./build/buildmacos.sh

# Windows build
./build/buildwindows.sh
```

## Contributing Guidelines

### Code Review Process

1. **Self-review**: Review your own code before submitting
2. **Peer review**: At least one maintainer must review and approve
3. **Automated checks**: All CI checks must pass
4. **Testing**: Include tests for new functionality

### Commit Message Format

Use conventional commit format:

```bash
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```bash
feat(player): add video quality selection
fix(api): handle network timeout errors
docs(readme): update installation instructions
```

### Pull Request Guidelines

1. **Title**: Use descriptive titles following conventional commit format
2. **Description**: Explain what changes you made and why
3. **Testing**: Describe how you tested your changes
4. **Breaking Changes**: Clearly mark any breaking changes
5. **Screenshots**: Include screenshots for UI changes

### Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help newcomers learn and contribute
- Focus on the code, not the person

## Development Tools

### Project Structure

Understanding the GoAnime project structure will help you navigate and contribute
effectively:

**Key Directories Explained:**

- **`cmd/`**: Contains the main application entry points. In Go projects, this is
  where executable commands live.
- **`internal/`**: Private application code that cannot be imported by external
  packages. This ensures encapsulation.
- **`internal/api/`**: Handles all external API communications (anime databases,
  streaming services, skip times)
- **`internal/player/`**: Core video player functionality with platform-specific
  implementations (Unix/Windows)
- **`internal/models/`**: Data structures used throughout the application
  (anime, skip times, URLs)
- **`internal/tracking/`**: Manages watch history, progress tracking with
  CGO/no-CGO variants
- **`internal/discord/`**: Discord Rich Presence integration for showing what
  you're watching
- **`internal/playback/`**: Media playback logic for different content types
  (movies, series)
- **`internal/appflow/`**: Application workflow and data flow management
- **`build/`**: Platform-specific build scripts for Linux, macOS, and Windows
- **`test/`**: Comprehensive test suite organized by functionality
- **Nix files**: Development environment configuration (`flake.nix`, `shell.nix`,
  etc.)

**Development Flow:**

1. **API Layer** (`internal/api/`) fetches anime data and episode information
2. **Models** (`internal/models/`) structure the data (anime info, skip times,
   URLs)
3. **Appflow** (`internal/appflow/`) manages data flow through the application
4. **Playback** (`internal/playback/`) handles media logic based on content type
5. **Player** (`internal/player/`) manages video playback with platform-specific
   code
6. **Tracking** (`internal/tracking/`) records user progress and watch history
7. **Discord** (`internal/discord/`) updates Rich Presence status

### Recommended Tools

- **IDE**: VS Code with Go extension, GoLand
- **Linting**: golangci-lint
- **Testing**: Go's built-in testing framework
- **Debugging**: Delve debugger
- **Version Control**: Git with conventional commits

### VS Code Configuration

Create `.vscode/settings.json`:

```json
{
    "go.formatTool": "gofmt",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

## Getting Help

- **Documentation**: Check existing docs in the `/docs` folder
- **Issues**: Search existing GitHub issues
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Don't hesitate to ask for help during review

Remember: Good code is written for humans to read, and only incidentally for
computers to execute!

---

Happy coding! 🚀
