# Copilot Instructions for ghq

## Language & Build

- This is a Go project. Use `go build` to build and `go test ./...` to run tests.

## Code Quality Checks

Before committing, always run the following in order:

1. `goimports -w .` — format code and organize imports
2. `go vet ./...` — check for common errors
3. `staticcheck ./...` — run static analysis

All three must pass with no errors before pushing.

## Testing

- Run `go test ./...` to execute all tests.
- Ensure all existing tests pass after making changes.
