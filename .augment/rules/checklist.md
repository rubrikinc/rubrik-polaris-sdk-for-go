---
type: "always_apply"
description: "Example description"
---

# Code Review Checklist

Use this checklist when reviewing or writing code for the Rubrik Polaris SDK for Go.

## GraphQL Queries

- [ ] Query uses `query RubrikPolarisSDKRequest` as the query name
- [ ] Query result is aliased to `result`
- [ ] Input fields are extrapolated as individual parameters (not using complex input types)
- [ ] Query is in a `.graphql` file in the appropriate `queries/` subdirectory
- [ ] `go generate ./...` has been run after adding/modifying `.graphql` files
- [ ] No manual edits to generated `queries.go` files
- [ ] No manual query variables created in Go code

## Region Handling

- [ ] Cloud regions use typed region types (aws.Region, azure.Region, gcp.Region)
- [ ] No string types used for region parameters
- [ ] No string types used for region struct fields
- [ ] Region parsing uses appropriate `RegionFrom*()` functions
- [ ] Region conversion to GraphQL enums uses `.ToRegionEnum()` or similar methods
- [ ] Invalid regions are handled (check for `RegionUnknown`)

## Pre-Commit Checklist

Before committing code:

1. [ ] Run `go fmt ./...`
2. [ ] Run `go vet ./...`
3. [ ] Run `go generate ./...` (if GraphQL queries were modified)
4. [ ] Run tests: `go test ./...`
5. [ ] Verify region types are used instead of strings (project-specific)

## Code Review Checklist (For Reviewers)

When reviewing code, check for:

### GraphQL Compliance
- [ ] `.graphql` files use `RubrikPolarisSDKRequest` query/mutation name
- [ ] Results are aliased to `result:`
- [ ] Generated `queries.go` files are included if `.graphql` files changed
- [ ] No manual edits to generated files

### Region Type Safety
- [ ] Region parameters use typed regions (not strings)
- [ ] Region struct fields use typed regions (not strings)
- [ ] Invalid regions are handled properly

**See [reviewer-guide.md](./reviewer-guide.md) for detailed review guidelines and [examples.md](./examples.md) for code examples.**

## Common Mistakes to Avoid

1. **Forgetting to run `go generate`** - Required after modifying `.graphql` files
2. **Manually editing `queries.go`** - These files are auto-generated
3. **Using `string` for regions** - Always use typed region types
4. **Missing documentation** - All exported items need doc comments
5. **Wrong query name** - Must be `RubrikPolarisSDKRequest`
6. **Missing result alias** - Query result must be aliased to `result`