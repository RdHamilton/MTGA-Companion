# Contributing to MTGA-Companion

Thank you for your interest in contributing to MTGA-Companion! This document provides guidelines and instructions for contributing to the project.

## Getting Started

### Prerequisites

- **Go 1.23+** - The project requires Go 1.23.12 or later
- **Git** - For version control
- **MTG Arena** - For testing (optional, but helpful)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/MTGA-Companion.git
   cd MTGA-Companion
   ```

3. **Add the upstream repository**:
   ```bash
   git remote add upstream https://github.com/ramonehamilton/MTGA-Companion.git
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Verify your setup**:
   ```bash
   ./scripts/dev.sh check
   ./scripts/test.sh
   ```

## Development Workflow

### Making Changes

1. **Create a new branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following the coding standards below

3. **Test your changes**:
   ```bash
   # Run all checks
   ./scripts/dev.sh check
   
   # Run tests
   ./scripts/test.sh
   
   # Run tests with coverage
   ./scripts/test.sh coverage
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Add: brief description of your change"
   ```

   Use clear, descriptive commit messages. Prefix with:
   - `Add:` for new features
   - `Fix:` for bug fixes
   - `Update:` for updates to existing features
   - `Refactor:` for code refactoring
   - `Docs:` for documentation changes

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** on GitHub

### Code Standards

#### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use `gofumpt` for stricter formatting (if available)
- Run `go vet` to catch common mistakes
- Follow Go best practices and idioms

The project includes scripts to help with this:

```bash
# Format code
./scripts/dev.sh fmt

# Run all checks
./scripts/dev.sh check
```

#### Code Organization

- Keep functions focused and small
- Write clear, descriptive names
- Add comments for exported functions and types
- Write tests for new functionality

#### Testing

- Write tests for new features and bug fixes
- Aim for good test coverage
- Use table-driven tests where appropriate
- Run tests with race detection:
  ```bash
  ./scripts/test.sh race
  ```

### Project Structure

```
MTGA-Companion/
├── cmd/
│   └── mtga-companion/      # Application entry point
├── internal/
│   ├── mtga/
│   │   └── logreader/       # Log reading and parsing
│   └── storage/             # Database and persistence
├── pkg/                     # Public libraries (future)
├── scripts/                 # Development scripts
└── examples/                # Example code
```

- `cmd/` - Application entry points
- `internal/` - Private application code
- `pkg/` - Public libraries (if any)
- `scripts/` - Development and build scripts

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Ensure all checks pass**:
   ```bash
   ./scripts/dev.sh check
   ./scripts/test.sh
   ```

3. **Update documentation** if you've changed functionality

### Pull Request Guidelines

- **Keep PRs focused** - One feature or fix per PR
- **Write clear descriptions** - Explain what and why
- **Reference issues** - Link to related issues if applicable
- **Add tests** - Include tests for new functionality
- **Update docs** - Update README or other docs if needed

### PR Review

- All PRs will be reviewed before merging
- Be open to feedback and suggestions
- Address review comments promptly
- Keep discussions constructive and respectful

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **Description** - Clear description of the bug
- **Steps to reproduce** - How to trigger the bug
- **Expected behavior** - What should happen
- **Actual behavior** - What actually happens
- **Environment** - OS, Go version, MTGA version
- **Logs** - Any relevant error messages or logs

### Feature Requests

For feature requests, please include:

- **Use case** - Why this feature would be useful
- **Proposed solution** - How you envision it working
- **Alternatives** - Other approaches you've considered

## Development Scripts

The project includes helpful scripts in the `scripts/` directory:

### `scripts/dev.sh`

Development workflow commands:

```bash
./scripts/dev.sh          # Run all checks and build
./scripts/dev.sh fmt      # Format code
./scripts/dev.sh vet      # Run go vet
./scripts/dev.sh lint     # Run golangci-lint
./scripts/dev.sh check    # Run fmt, vet, and lint
./scripts/dev.sh build    # Build the application
```

### `scripts/test.sh`

Testing commands:

```bash
./scripts/test.sh              # Run tests with race detection
./scripts/test.sh unit         # Run unit tests
./scripts/test.sh coverage     # Generate coverage report
./scripts/test.sh race         # Run with race detection
./scripts/test.sh verbose      # Verbose output
./scripts/test.sh bench        # Run benchmarks
```

## Questions?

- **Open an issue** for questions or discussions
- **Check existing issues** to see if your question has been answered
- Be patient - this is a small project maintained in spare time

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Be patient with questions

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or inflammatory comments
- Personal attacks
- Any other unprofessional conduct

## License

By contributing, you agree that your contributions will be licensed under the same MIT License that covers the project.

## Thank You!

Contributions of all kinds are welcome - code, documentation, bug reports, feature suggestions, and more. Thank you for helping make MTGA-Companion better!


