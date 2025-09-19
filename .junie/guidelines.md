# Development Guidelines for Gobee Project

This document provides essential information for developers working on the Gobee project. It covers build/configuration instructions, testing procedures, and development practices.

## Project Structure

The Gobee project is organized as a multi-module Go project. Each directory at the root level represents a separate Go module with its own `go.mod` file:

- `lvan`: Core functionality and tools
- `sign`: Signing utilities
- `realm`: Authentication and authorization
- `cache`: Caching mechanisms
- `memshare`: Memory sharing utilities
- `reflectparam`: Reflection utilities
- And others...

## Build/Configuration Instructions

### Setting Up the Development Environment

1. **Go Version**: The project requires Go 1.21 or later.

2. **Module Structure**: Each module has its own `go.mod` file. When working on a specific module, navigate to its directory before running Go commands.

3. **Dependencies**: Dependencies are managed through the `go.mod` files in each module. Use `go mod tidy` in the module directory to ensure dependencies are up to date.

4. **Building Modules**:
   - Navigate to the module directory: `cd <module_name>`
   - Build the module: `go build ./...`
   - For executables, build the specific command: `go build ./cmd/<command_name>`

5. **Windows-Specific Tools**:
   - Some tools in the `lvan/tools` directory use batch files (`.bat`) for Windows-specific functionality.
   - The `lsmp.bat` script in `lvan/tools/lsmp` is used for processing and packaging data files.

## Testing Information

### Running Tests

1. **Standard Tests**:
   - Navigate to the module directory: `cd <module_name>`
   - Run all tests: `go test ./...`
   - Run tests with verbose output: `go test -v ./...`
   - Run a specific test: `go test -v -run <TestName>`

2. **Test Coverage**:
   - Generate test coverage: `go test -cover ./...`
   - Detailed coverage report: `go test -coverprofile=coverage.out ./...`
   - View coverage in browser: `go tool cover -html=coverage.out`

### Adding New Tests

1. **Test File Naming**:
   - Test files should be named with the `_test.go` suffix.
   - Place test files in the same package as the code being tested.

2. **Test Function Naming**:
   - Test functions should be named `Test<FunctionName>` for unit tests.
   - Use `Benchmark<FunctionName>` for benchmarks.
   - Use `Example<FunctionName>` for examples.

3. **Test Patterns**:
   - The project uses table-driven tests extensively.
   - For complex tests, use subtests with `t.Run()`.
   - Use `testing/quick` for property-based testing where appropriate.

### Example Test

Here's a simple example of a test for the `sign` module:

```go
package sign

import (
	"testing"
)

func TestSimpleMd5(t *testing.T) {
	// Create a test struct with sign tags
	testStruct := &struct {
		Name      string `sign:"yes"`
		Value     int    `sign:"yes"`
		Signature string `sign:"access"`
	}{
		Name:  "TestValue",
		Value: 123,
	}

	// Apply the MD5 signature
	Md5Apply("test_key", testStruct)

	// Verify that the signature field was set
	if testStruct.Signature == "" {
		t.Error("Signature field was not set")
	}

	// Get the expected signature using the Md5 function
	expected := Md5("test_key", testStruct)

	// Verify that the signature matches the expected value
	if testStruct.Signature != expected {
		t.Errorf("Expected signature %s, got %s", expected, testStruct.Signature)
	}

	t.Logf("Successfully verified MD5 signature: %s", testStruct.Signature)
}
```

To run this test:
```
cd sign
go test -v -run TestSimpleMd5
```

## Additional Development Information

### Code Style

1. **Go Conventions**:
   - Follow standard Go code style and conventions.
   - Use `gofmt` or `goimports` to format code.
   - Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines.

2. **Error Handling**:
   - Use explicit error checking.
   - Return errors rather than using panic (except in truly exceptional cases).
   - Use meaningful error messages.

3. **Documentation**:
   - Document public functions and types with godoc-compatible comments.
   - Include examples where appropriate.

### Project-Specific Practices

1. **Module Independence**:
   - Modules should be as independent as possible.
   - Cross-module dependencies should be minimized.

2. **Testing**:
   - Write tests for all new functionality.
   - Use table-driven tests for comprehensive test coverage.
   - Consider property-based testing for complex functions.

3. **Windows Compatibility**:
   - The project is primarily developed for Windows.
   - Use Windows-style paths (backslashes) in Windows-specific code.
   - For cross-platform code, use `filepath.Join()` instead of hardcoded path separators.

4. **Batch Files**:
   - Some tools use batch files for Windows-specific functionality.
   - Follow the existing patterns when modifying or creating new batch files.

### Debugging

1. **Logging**:
   - The project uses a custom logging package in `lvan/pkg/logger`.
   - Use appropriate log levels for different types of messages.

2. **Testing Tools**:
   - Use `go test -v` for verbose test output.
   - For complex issues, use `t.Logf()` to output debug information during tests.

## Conclusion

This document provides a starting point for development on the Gobee project. As the project evolves, this document should be updated to reflect changes in the development process and best practices.