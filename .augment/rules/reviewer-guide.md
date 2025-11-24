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

### 1. Code Standards Compliance

#### Documentation
- [ ] All exported types have documentation comments starting with the type name
- [ ] All exported functions have documentation comments starting with the function name
- [ ] All exported methods have documentation comments
- [ ] Documentation uses complete sentences with proper punctuation
- [ ] Complex functions include usage examples where appropriate

**Common Issues**:
- Missing documentation on new types or functions
- Documentation that doesn't start with the item name
- Incomplete or unclear documentation

**Example Feedback**:
```
❌ Missing documentation:
type CloudAccount struct { ... }

✅ Should be:
// CloudAccount represents an AWS cloud account in RSC.
type CloudAccount struct { ... }
```

#### Naming Conventions
- [ ] All acronyms are fully uppercase (AWS, CDM, API, ID, UUID, VPC, IAM, ARN, EKS, RDS, EC2, etc.)
- [ ] Type names follow Go conventions (PascalCase for exported)
- [ ] Variable names are descriptive and meaningful
- [ ] Function names clearly describe their action

**Common Issues**:
- Mixed case acronyms (e.g., `AwsApi` instead of `AWSAPI`)
- Inconsistent naming (e.g., `Id` instead of `ID`)
- Unclear variable names

**Example Feedback**:
```
❌ Incorrect acronym capitalization:
type AwsCloudAccount struct {
    VpcId string
}

✅ Should be:
type AWSCloudAccount struct {
    VPCID string
}
```

### 2. GraphQL Query Standards

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

Note: After adding/modifying .graphql files, run `go generate ./...` from the repository root.
```

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

### 4. Code Quality

#### Error Handling
- [ ] All errors are handled explicitly (no ignored errors)
- [ ] Errors are wrapped appropriately (see guidance below)
- [ ] Error messages are descriptive and helpful
- [ ] Error messages start with lowercase (Go convention)

**Error Wrapping Guidance**:
- Use `%w` when the caller needs to take action on the specific error type (e.g., `graphql.ErrNotFound`, transient errors)
- Use `%s` when the error is an implementation detail and wrapping would create unnecessary coupling between caller and implementation
- When in doubt, prefer `%s` to avoid exposing implementation details as part of the interface

**Example Feedback**:
```
❌ Ignored error:
data, _ := json.Marshal(obj)

✅ Should handle error and hide implementation details:
data, err := json.Marshal(obj)
if err != nil {
    return fmt.Errorf("failed to marshal object: %s", err)
}
```

#### Function Design
- [ ] Functions are focused and single-purpose
- [ ] Functions use `context.Context` for cancellation and timeouts
- [ ] Functions have reasonable complexity (not too long or nested)
- [ ] Repeated code is extracted into helper functions

#### Consistency
- [ ] Code follows existing patterns in the codebase
- [ ] Similar functionality uses similar approaches
- [ ] Naming is consistent with related code

### 5. Testing

#### Test Coverage
- [ ] New functionality has appropriate tests
- [ ] Tests follow existing patterns in the codebase
- [ ] Tests use meaningful names (e.g., `TestFunctionName_Scenario`)
- [ ] Tests cover error cases
- [ ] Tests are deterministic (no flaky tests)
- [ ] Table-driven tests are preferred when they can be implemented without too much complexity

**Example Feedback**:
```
Please add tests for the new AddExocompute function. Consider testing:
- Successful exocompute addition
- Invalid region handling
- Error cases (e.g., invalid cloud account ID)

Example test structure (table-driven preferred):
func TestAddExocompute(t *testing.T) {
    tests := []struct {
        name    string
        config  ExocomputeConfig
        wantErr bool
    }{
        {"success", validConfig, false},
        {"invalid region", invalidRegionConfig, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

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

### 7. CI/CD Compliance

#### Build Pipeline
- [ ] Code passes `go vet ./...`
- [ ] Code passes `gofmt` (no formatting issues)
- [ ] Code passes `staticcheck`
- [ ] All tests pass
- [ ] No race conditions (if applicable)

**Example Feedback**:
```
Please run the following locally before pushing:
  go fmt ./...
  go vet ./...
  go run honnef.co/go/tools/cmd/staticcheck@v0.6.1 ./...
  go test ./...
```

## Review Patterns from Past PRs

Based on historical code reviews, pay special attention to:

1. **Acronym Consistency**: Frequently, `Vm` → `VM`, `Cdm` → `CDM`, `Id` → `ID`, `Dns` → `DNS`, `Ntp` → `NTP`

2. **GraphQL Query Names**: Ensure mutation/query names in `.graphql` files use `RubrikPolarisSDKRequest`

3. **Error Message Format**: Use `%w` for error wrapping, `%s` for string formatting

4. **Struct Field Naming**: Ensure consistency (e.g., `DNSNameServers` not `DnsNameServers`)

5. **Removed Code**: Check if large deletions are intentional (e.g., moving code to another file)

6. **Test Error Messages**: Use `%s` instead of `%v` for error formatting in tests

7. **Whitespace**: Remove unnecessary blank lines, ensure consistent spacing

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
**🟡 Naming Convention**: Acronym capitalization

❌ Current code:
type AwsVmConfig struct {
    CdmVersion string
    VpcId      string
}

✅ Suggested fix:
type AWSVMConfig struct {
    CDMVersion string
    VPCID      string
}

Reason: All acronyms should be fully uppercase per project coding standards.
```

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

