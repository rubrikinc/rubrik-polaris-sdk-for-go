# Coding Standards

## General

- Read existing code before editing — follow adjacent patterns
- All exported functions and types require doc comments
- Functions follow Single Responsibility Principle
- Inject dependencies to enable testing with fakes
- Avoid code duplication — use existing methods and stdlib before writing custom
  implementations
- MIT license header required on all source files (copyright Rubrik, Inc.)

## Go-Specific

- Make code changes before adding import statements (IDE may remove unused ones)
- Use `secret.String` (from `pkg/polaris/graphql/core/secret`) for credentials
  and tokens — never bare `string`
- Use typed regions (`aws.Region`, etc.) — never `string` for regions

## Testing

- CamelCase test function names
- Prefer fakes over mocks for dependencies
- Table-driven tests when complexity is manageable
- Integration tests require `TEST_INTEGRATION=1` env var
- Use `internal/assert/` for test assertions and `internal/handler/` for mock
  HTTP servers
