# Contributing to Valkey MCP Task Management Server

Thank you for considering contributing to the Valkey MCP Task Management Server! This document outlines the process and guidelines for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct. Please be respectful, inclusive, and considerate in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment as described in the README.md
4. Create a new branch for your changes

## Development Workflow

1. Make your changes in your feature branch
2. Add tests for your changes
3. Ensure all tests pass with `make test`
4. Update documentation as needed
5. Commit your changes using [Conventional Commits](#conventional-commits)
6. Push your branch and submit a pull request

## Pull Request Process

1. Ensure your PR includes tests for any new functionality
2. Update the README.md or documentation with details of changes if applicable
3. The PR should work with the CI/CD pipeline without errors
4. A maintainer will review your PR and may request changes
5. Once approved, a maintainer will merge your PR

## Conventional Commits

This project strictly follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. All commit messages MUST adhere to this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Changes that don't affect code meaning (formatting, etc.)
- `refactor`: Code changes that neither fix bugs nor add features
- `perf`: Performance improvements
- `test`: Adding or correcting tests
- `build`: Changes to build system or dependencies
- `ci`: Changes to CI configuration
- `chore`: Other changes that don't modify source or test files

### Breaking Changes

Breaking changes MUST be indicated by:
- Adding an exclamation mark after the type/scope: `feat!: introduce breaking API change`
- OR including `BREAKING CHANGE:` in the footer: 
  ```
  feat: new feature
  
  BREAKING CHANGE: breaks compatibility with previous versions
  ```

### Examples

```
feat(storage): add support for Valkey Cluster

Adds the ability to connect to Valkey Cluster for improved scalability.
```

```
fix: correct task ordering logic

Fixes issue #42 where tasks were not properly ordered after reordering.
```

```
docs: update API documentation with new endpoints
```

```
feat!: redesign task API

BREAKING CHANGE: The task API has been completely redesigned to improve usability.
Previous endpoints are no longer supported.
```

## Testing Guidelines

- Write tests for all new features and bug fixes
- Maintain or improve test coverage with each PR
- Use table-driven tests when appropriate
- Integration tests should be in the `tests/integration` directory
- Unit tests should be in the same package as the code they test

## Code Style

- Follow standard Go coding conventions
- Use `gofmt` to format your code
- Follow the recommendations in [Effective Go](https://golang.org/doc/effective_go)
- Run `golint` and `go vet` to catch common issues

## Documentation

- Update documentation when changing functionality
- Use clear, concise language
- Include examples where appropriate
- Document exported functions, types, and constants

## Issue Reporting

- Use the issue tracker to report bugs or request features
- Check existing issues before creating a new one
- Provide detailed steps to reproduce bugs
- Include relevant information like OS, Go version, etc.

## Releasing

The project uses [Semantic Release](https://github.com/semantic-release/semantic-release) to automatically determine the next version number based on conventional commits:

- `fix:` commits increment the patch version (1.0.0 → 1.0.1)
- `feat:` commits increment the minor version (1.0.0 → 1.1.0)
- Breaking changes increment the major version (1.0.0 → 2.0.0)

## License

By contributing to this project, you agree that your contributions will be licensed under the project's BSD-3-Clause License.
