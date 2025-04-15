# Contributing to Go AI Agent

Thank you for your interest in contributing to the Go AI Agent project! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/ai-agent.git`
3. Create a new branch: `git checkout -b feature/your-feature-name`

## Development Workflow

1. Make your changes
2. Follow the code style guidelines in CLAUDE.md
3. Run tests to ensure your changes don't break existing functionality: `go test ./...`
4. Commit your changes with a clear commit message
5. Push to your branch
6. Open a pull request

## Pull Request Process

1. Ensure your code follows the style guidelines
2. Update documentation if necessary
3. Make sure all tests pass
4. The PR should reference any related issues

## Style Guidelines

- Format code using `gofmt` or `go fmt ./...`
- Group imports: standard library, then third-party, then local packages
- Always check errors and provide context in error messages
- Use camelCase for variables, PascalCase for exported functions/types
- Document exported functions and types with proper godoc format

## Testing

All new features or bug fixes should include tests. Run tests with:

```bash
go test ./...
```

## License

By contributing to this project, you agree that your contributions will be licensed under the project's MIT License.