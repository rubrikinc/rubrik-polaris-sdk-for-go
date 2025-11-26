---
type: "always_apply"
---

# Augment AI Guidelines for Rubrik Polaris SDK for Go

This directory contains guidelines and coding standards for the Rubrik Polaris SDK for Go project.

## Core Principles

**Project-Specific Standards** (beyond standard Go conventions):

1. **Type Safety**: Use typed region types (`aws.Region`, `azure.Region`, `gcp.Region`) instead of strings
2. **GraphQL Standards**: All queries use `query RubrikPolarisSDKRequest` with result aliased to `result`
3. **Code Generation**: Never manually edit generated `queries.go` files

For general Go conventions (formatting, documentation, naming, error handling, acronym capitalization), refer to:
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) (especially the Initialisms section)

## Guidelines Documents
### [Coding Standards](./coding-standards.md)
General Go coding conventions, documentation requirements, and best practices.

### [GraphQL Guidelines](./graphql-guidelines.md)
Standards for GraphQL queries:
- Query naming: `query RubrikPolarisSDKRequest`
- Result aliasing: `result:`
- Input field extrapolation (no complex input types)
- Code generation workflow with `go generate ./...`

### [Region Types](./region-types.md)
Cloud provider region handling:
- Use `aws.Region`, `azure.Region`, `gcp.Region` types
- Parse with `RegionFromName()`, `RegionFromAny()`
- Convert with `.ToRegionEnum()`, `.ToNativeRegionEnum()`, etc.
- Handle `RegionUnknown` validation

### [Reviewer Guide](./reviewer-guide.md)
Code review checklist and feedback guidelines for ensuring quality and consistency.

### [Checklist](./checklist.md)
Quick reference checklist for common coding tasks and reviews.

## Workflow

### Adding GraphQL Queries
1. Create/edit `.graphql` file in `queries/` subdirectory
2. Use `query RubrikPolarisSDKRequest` with `result:` alias
3. Run `go generate ./...` from repository root
4. Verify generated `queries.go` file

### Working with Regions
1. Import region package (aws/azure/gcp)
2. Use region constants or parse with `RegionFromName()`
3. Convert to GraphQL enum with appropriate method
4. Validate against `RegionUnknown`

## Enforcement

These guidelines are enforced through:
1. Code review (see [reviewer-guide.md](./reviewer-guide.md))
2. Augment AI when making code suggestions
3. CI/CD linting and static analysis

