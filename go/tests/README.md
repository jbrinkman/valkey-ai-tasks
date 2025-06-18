# Testing Directory Structure

This directory contains the testing infrastructure for the valkey-ai-tasks project.

## Directory Structure

- `integration/`: Contains integration tests that verify the interaction between different components of the system
- `mocks/`: Contains mock implementations of interfaces used in testing
- `utils/`: Contains utility functions and helpers for testing

## Dependencies

This project uses the following testing libraries:

- [Testify](https://github.com/stretchr/testify): A toolkit with common assertions and mocks that is compatible with the standard library

## Running Tests

To run all tests:

```bash
cd /path/to/valkey-ai-tasks/go
go test ./tests/...
```

To run specific test categories:

```bash
# Run integration tests only
go test ./tests/integration/...

# Run a specific test file
go test ./tests/integration/specific_test.go
```

To run tests with verbose output:

```bash
go test -v ./tests/...
```

## Writing Tests

When adding new tests:

1. Place integration tests in the appropriate subdirectory under `integration/`
2. Add mock implementations to the `mocks/` directory
3. Use Testify's assertion and mock packages for testing
4. Use utility functions from `utils/` for common testing operations
5. Follow Go testing best practices and naming conventions

### Example Test

```go
func TestSomething(t *testing.T) {
    // Use the test utilities
    ctx, req, cancel := utils.SetupTest(t)
    defer utils.CleanupTest(t, cancel)
    
    // Your test code here
    result := SomeFunction()
    
    // Use testify assertions
    req.Equal(expected, result)
    req.NoError(err)
}
```
