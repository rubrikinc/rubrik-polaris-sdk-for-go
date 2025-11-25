---
type: "agent_requested"
description: "When reviewing code for a PR or acting as a reviewer for another developer"
---

# Code Review Guide for Augment AI

This document provides comprehensive guidelines for Augment AI to act as a code reviewer for the Rubrik Polaris SDK for Go project.

## Role as Reviewer

When acting as a code reviewer, Augment should:
1. **Be thorough but constructive** - Point out issues clearly but suggest solutions
2. **Focus on code quality** - Ensure code follows project standards and best practices
3. **Check for consistency** - Verify code matches existing patterns in the codebase
4. **Validate completeness** - Ensure all necessary changes are included
5. **Verify testing** - Check that appropriate tests are included or updated

## Review Checklist

**Note**: For general Go conventions (documentation, naming, formatting, error handling), refer to [Effective Go](https://go.dev/doc/effective_go). This checklist focuses on **project-specific** standards.

### 1. Project-Specific Standards

#### GraphQL Query Standards

#### Query Structure
- [ ] Query uses `query RubrikPolarisSDKRequest` or `mutation RubrikPolarisSDKRequest`
- [ ] Query result is aliased to `result:`
- [ ] Input fields are extrapolated as individual parameters (not using complex input types)
- [ ] Query is in a `.graphql` file in the appropriate `queries/` subdirectory
- [ ] Generated `queries.go` file has been updated (run `go generate ./...`)
- [ ] No manual edits to generated `queries.go` files

**Common Issues**:
- Custom query names instead of `RubrikPolarisSDKRequest`
- Missing `result:` alias
- Using complex input types instead of extrapolating fields
- Manually editing generated files
- Forgetting to run `go generate ./...`

**Example Feedback**:
```
❌ Wrong query name:
query GetAwsAccounts($id: String!) {
  allAwsCloudAccounts(id: $id) { ... }
}

✅ Should be:
query RubrikPolarisSDKRequest($id: String!) {
  result: allAwsCloudAccounts(id: $id) { ... }
}

❌ Using complex input type:
mutation RubrikPolarisSDKRequest($input: UpdateAwsCloudAccountInput!) {
  result: updateAwsCloudAccount(input: $input) { ... }
}

✅ Should extrapolate fields:
mutation RubrikPolarisSDKRequest($cloudAccountId: String!, $name: String!) {
  result: updateAwsCloudAccount(input: {
    cloudAccountId: $cloudAccountId,
    name: $name,
  }) { ... }
}
```
Note: After adding/modifying .graphql files, run `go generate ./...` from the repository root.

### 3. Region Type Enforcement

#### Type Safety
- [ ] Cloud regions use typed region types (`aws.Region`, `azure.Region`, `gcp.Region`)
- [ ] No string types used for region parameters
- [ ] No string types used for region struct fields
- [ ] Region parsing uses appropriate `RegionFrom*()` functions
- [ ] Region conversion to GraphQL enums uses `.ToRegionEnum()` or similar methods
- [ ] Invalid regions are handled (check for `RegionUnknown`)

**Common Issues**:
- Using `string` instead of region types
- Not validating region input
- Not handling `RegionUnknown` case

**Example Feedback**:
```
❌ Using string for region:
func AddExocompute(ctx context.Context, region string) error { ... }

✅ Should use region type:
func AddExocompute(ctx context.Context, region aws.Region) error {
    if region == aws.RegionUnknown {
        return fmt.Errorf("invalid region")
    }
    regionEnum := region.ToRegionEnum()
    // ...
}
```

### 4. Code Quality and Testing

**Note**: For general Go error handling, function design, and best practices, see [Effective Go](https://go.dev/doc/effective_go).
**Note**: For general Go testing conventions, see [Effective Go - Testing](https://go.dev/doc/effective_go#testing).

#### Project-Specific Test Requirements
- [ ] New functionality has appropriate tests
- [ ] Tests follow existing patterns in the codebase
- [ ] Table-driven tests are preferred when they can be implemented without too much complexity

### 6. Generated Files

#### Code Generation
- [ ] If `.graphql` files were added/modified, `go generate ./...` was run
- [ ] Generated `queries.go` files are included in the PR
- [ ] No manual edits to generated files
- [ ] Generated files match the `.graphql` source files

**Example Feedback**:
```
It looks like you added a new GraphQL query file but didn't run `go generate ./...`.
Please run:
  go generate ./...
from the repository root and commit the generated queries.go file.
```

#### Build Pipeline
- [ ] Code passes `staticcheck`
- [ ] All tests pass

## Review Patterns from Past PRs

Based on historical code reviews, pay special attention to:
1. **GraphQL Query Names**: Ensure mutation/query names in `.graphql` files use `RubrikPolarisSDKRequest`
2. **Removed Code**: Check if large deletions are intentional (e.g., moving code to another file)
3. **Test Error Messages**: Use `%s` instead of `%v` for error formatting in tests
4. **Whitespace**: Remove unnecessary blank lines, ensure consistent spacing

## Providing Feedback

### Feedback Format

Use this format for review comments:

```markdown
**[Category]**: [Issue Description]

❌ Current code:
[code snippet]

✅ Suggested fix:
[corrected code snippet]

Reason: [Brief explanation]
```

### Severity Levels

- **🔴 Critical**: Must be fixed (e.g., security issues, broken functionality)
- **🟡 Important**: Should be fixed (e.g., violates coding standards, potential bugs)
- **🔵 Minor**: Nice to have (e.g., style improvements, optimization opportunities)
- **💡 Suggestion**: Optional improvement (e.g., alternative approaches)

### Example Review Comments

```markdown
**🔴 GraphQL Query**: Missing result alias

❌ Current code:
query RubrikPolarisSDKRequest($id: UUID!) {
  awsCloudAccount(id: $id) { ... }
}

✅ Suggested fix:
query RubrikPolarisSDKRequest($id: UUID!) {
  result: awsCloudAccount(id: $id) { ... }
}

Reason: All GraphQL queries must alias the result to `result:` per project standards.
```

## Final Checklist

Before approving a PR, verify:
- [ ] All coding standards are followed
- [ ] GraphQL queries follow conventions
- [ ] Region types are used correctly
- [ ] Tests are included and pass
- [ ] Documentation is complete
- [ ] Generated files are up-to-date
- [ ] No unnecessary code changes
- [ ] Error handling is proper
- [ ] Code is consistent with existing patterns

