# Testing Directory Structure

This directory contains the testing infrastructure for the valkey-ai-tasks application.

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
6. For testing notes functionality, ensure you test with various Markdown formats and special characters

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

## Testing Notes Functionality

The system supports Markdown-formatted notes for both plans and tasks. When testing notes functionality:

### Integration Tests

Integration tests for notes functionality are located in:

- `integration/plan_repository_suite_test.go`: Tests for plan notes
- `integration/task_repository_suite_test.go`: Tests for task notes

These tests cover:

1. **Creating and updating notes**: Testing that notes can be created and updated correctly
2. **Retrieving notes**: Testing that notes can be retrieved correctly
3. **Error handling**: Testing behavior when attempting to update or retrieve notes for non-existent entities
4. **Special characters**: Testing that notes with special characters, emojis, and Unicode are handled correctly
5. **Markdown formatting**: Testing that various Markdown formatting elements are preserved

### Example Notes Test

```go
func (suite *TaskRepositorySuite) TestUpdateTaskNotes() {
    // Create a task
    task := models.Task{
        ID:          uuid.New().String(),
        PlanID:     "test-plan-id",
        Title:      "Test Task",
        Description: "Test Description",
        Status:     "pending",
        Priority:   "medium",
    }
    
    err := suite.taskRepo.Create(context.Background(), task)
    suite.NoError(err)
    
    // Update notes
    markdownNotes := "# Task Notes\n\nThis is a **bold** statement.\n\n```go\nfunc example() {\n  fmt.Println(\"Hello\")\n}\n```"
    err = suite.taskRepo.UpdateNotes(context.Background(), task.ID, markdownNotes)
    suite.NoError(err)
    
    // Get notes and verify
    notes, err := suite.taskRepo.GetNotes(context.Background(), task.ID)
    suite.NoError(err)
    suite.Equal(markdownNotes, notes)
}
```
