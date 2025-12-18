# Contributing to Guardian-Log

Thank you for your interest in contributing! This guide will help you get started.

## Ways to Contribute

- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve documentation
- 🔧 Submit code changes

## Getting Started

### 1. Fork the Repository

Click "Fork" on GitHub to create your copy.

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/guardian-log.git
cd guardian-log
```

### 3. Set Up Development Environment

```bash
# Install dependencies
make install

# Start development servers
make dev-backend   # Terminal 1
make dev-frontend  # Terminal 2

# Access at http://localhost:5173
```

See [Development Guide](docs/development/GUIDE.md) for details.

## Making Changes

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bugfix
```

### 2. Make Your Changes

- Write clean, readable code
- Follow existing code style
- Add comments where helpful
- Update documentation if needed

### 3. Test Your Changes

```bash
make test       # Run tests
make lint       # Run linters
make build      # Build to verify
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "Add feature: description"
```

**Commit message format:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test changes
- `chore:` Build/tooling changes

### 5. Push and Create Pull Request

```bash
git push origin feature/my-feature
```

Then create a Pull Request on GitHub.

## Pull Request Guidelines

- **Clear title** - Describe what the PR does
- **Description** - Explain why and how
- **Link issues** - Reference related issues
- **Tests pass** - All CI checks must pass
- **Small PRs** - Focus on one thing

## Code Style

### Go
- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable names
- Add godoc comments for exported items

### TypeScript/React
- Use TypeScript for type safety
- Functional components with hooks
- Descriptive component names
- Format with Prettier

## Testing

```bash
# Go tests
go test ./...

# With coverage
go test -cover ./...

# Frontend tests (when available)
cd web && npm test
```

## Documentation

- Update docs for new features
- Keep docs in sync with code
- Use clear, concise language
- Include code examples

## Reporting Bugs

**Include:**
- Description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, versions, etc.)
- Logs if relevant

## Suggesting Features

**Include:**
- Clear description of the feature
- Use case / problem it solves
- Possible implementation approach
- Alternatives considered

## Questions?

- Check [existing issues](https://github.com/OWNER/guardian-log/issues)
- Start a [discussion](https://github.com/OWNER/guardian-log/discussions)
- Read the [docs](docs/)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Guardian-Log! 🎉
