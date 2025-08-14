# Contributing to FL-GO

Thank you for your interest in contributing to FL-GO! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Docker (optional)
- Node.js and npm (for web UI development)

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/fl-go.git
   cd fl-go
   ```

2. **Setup development environment**
   ```bash
   make setup-dev
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   make install-web-deps
   ```

4. **Run tests to verify setup**
   ```bash
   make test
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Follow the [coding standards](#coding-standards)
- Write tests for new functionality
- Update documentation as needed

### 3. Run Tests and Checks

```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-security

# Run linting
make lint

# Format code
make format
```

### 4. Commit Your Changes

Follow the [conventional commits](https://www.conventionalcommits.org/) format:

```bash
git commit -m "feat: add new federated learning algorithm"
git commit -m "fix: resolve connection timeout issue"
git commit -m "docs: update API documentation"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Write meaningful comments for exported functions
- Keep functions small and focused
- Use meaningful variable names

### Code Organization

- Place new packages in the `pkg/` directory
- Add new commands in the `cmd/` directory
- Update examples in the `examples/` directory
- Document changes in the `docs/` directory

### Testing

- Write unit tests for all new functionality
- Aim for at least 80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"normal case", "input", "expected"},
        {"edge case", "", ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := NewFeature(tt.input)
            if result != tt.expected {
                t.Errorf("got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

## Documentation

### Code Documentation

- Document all exported functions and types
- Use [godoc](https://godoc.org/) format
- Include examples for complex functions

### User Documentation

- Update relevant documentation in `docs/`
- Add examples in `examples/`
- Update README.md if needed

## Pull Request Guidelines

### Before Submitting

1. **Ensure all tests pass**
   ```bash
   make test
   ```

2. **Check code quality**
   ```bash
   make lint
   make format
   ```

3. **Update documentation**
   - Add/update relevant docs
   - Update examples if needed
   - Update README if needed

4. **Test your changes**
   - Test locally
   - Test with different configurations
   - Test edge cases

### Pull Request Template

Use the following template for your PR:

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Documentation
- [ ] Code documentation updated
- [ ] User documentation updated
- [ ] Examples updated

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Tests pass locally
- [ ] Documentation updated
```

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Environment information**
   - OS and version
   - Go version
   - FL-GO version

2. **Steps to reproduce**
   - Clear, step-by-step instructions
   - Minimal example if possible

3. **Expected vs actual behavior**
   - What you expected to happen
   - What actually happened

4. **Additional context**
   - Logs and error messages
   - Configuration files
   - Screenshots if relevant

### Feature Requests

When requesting features, please include:

1. **Problem description**
   - What problem are you trying to solve?
   - Why is this important?

2. **Proposed solution**
   - How should this work?
   - Any specific requirements?

3. **Use cases**
   - Who would benefit from this?
   - How would they use it?

## Release Process

### Versioning

FL-GO follows [semantic versioning](https://semver.org/):

- **Major version**: Breaking changes
- **Minor version**: New features, backward compatible
- **Patch version**: Bug fixes, backward compatible

### Release Checklist

Before releasing:

1. **Update version**
   - Update version in relevant files
   - Update CHANGELOG.md

2. **Run full test suite**
   ```bash
   make test
   make test-performance
   ```

3. **Build and test**
   ```bash
   make build
   make docker-build
   ```

4. **Create release**
   - Create git tag
   - Write release notes
   - Upload binaries

## Getting Help

- **Documentation**: Check the `docs/` directory
- **Examples**: Look at `examples/` directory
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## License

By contributing to FL-GO, you agree that your contributions will be licensed under the same license as the project.
