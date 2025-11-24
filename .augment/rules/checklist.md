---
type: "always_apply"
---

# Code Review Checklist

Use this checklist when reviewing or writing code for the Rubrik Polaris SDK for Go.

## Documentation

- [ ] All exported types have documentation comments
- [ ] All exported functions have documentation comments
- [ ] All exported methods have documentation comments
- [ ] All exported constants have documentation comments
- [ ] Documentation comments start with the name of the item
- [ ] Documentation comments use complete sentences with proper punctuation (periods, commas, etc.)

## Naming Conventions

- [ ] All acronyms are fully uppercase (AWS, CDM, API, ID, UUID, URL, HTTP, JSON, VPC, IAM, ARN, EKS, RDS, EC2)
- [ ] Type names follow Go conventions (PascalCase for exported, camelCase for unexported)
- [ ] Variable names are descriptive and meaningful
- [ ] Function names clearly describe their action
- [ ] Package names should be short and all lower case.

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

## Code Quality

- [ ] Code follows Go formatting standards (gofmt/goimports)
- [ ] Errors are handled explicitly (no ignored errors)
- [ ] Context is used for cancellation and timeouts
- [ ] Functions are focused and single-purpose
- [ ] No unnecessary complexity

## Testing

- [ ] New functionality has appropriate tests
- [ ] Tests follow existing patterns in the codebase
- [ ] Tests use meaningful names
- [ ] Tests cover error cases
- [ ] Table-driven tests are preferred when not too complex

## Examples by Category

### ✅ Correct Examples

#### Documentation
```go
// CloudAccount represents an AWS cloud account in RSC.
type CloudAccount struct {
    ID uuid.UUID
}

// AddAccount adds a new AWS cloud account to RSC.
func AddAccount(ctx context.Context) error {
    return nil
}
```

#### Acronyms
```go
type AWSAPI struct {
    CloudAccountID uuid.UUID
    VPCID         string
    IAMARN        string
}

type CDMCluster struct {
    ID uuid.UUID
}
```

#### GraphQL Query
```graphql
query RubrikPolarisSDKRequest($cloudAccountId: String!) {
  result: allAwsCloudAccounts(cloudAccountId: $cloudAccountId) {
    id
    name
  }
}
```

#### Region Types
```go
func AddExocompute(ctx context.Context, region aws.Region) error {
    regionEnum := region.ToRegionEnum()
    // use regionEnum
}

type Config struct {
    Region aws.Region
}
```

### ❌ Incorrect Examples

#### Missing Documentation
```go
type CloudAccount struct {
    ID uuid.UUID
}

func AddAccount(ctx context.Context) error {
    return nil
}
```

#### Wrong Acronym Capitalization
```go
type AwsApi struct {
    CloudAccountId uuid.UUID
    VpcId         string
    IamArn        string
}

type CdmCluster struct {
    Id uuid.UUID
}
```

#### Wrong GraphQL Query
```graphql
query GetAwsAccounts($cloudAccountId: String!) {
  allAwsCloudAccounts(cloudAccountId: $cloudAccountId) {
    id
    name
  }
}
```

#### String Instead of Region Type
```go
func AddExocompute(ctx context.Context, region string) error {
    // No type safety
}

type Config struct {
    Region string
}
```

## Pre-Commit Checklist

Before committing code:

1. [ ] Run `go fmt ./...`
2. [ ] Run `go vet ./...`
3. [ ] Run `go generate ./...` (if GraphQL queries were modified)
4. [ ] Run tests: `go test ./...`
5. [ ] Review changes against this checklist
6. [ ] Verify all new code has documentation
7. [ ] Verify all acronyms are uppercase
8. [ ] Verify region types are used instead of strings

## Code Review Checklist (For Reviewers)

When reviewing code, check for:

### Standards Compliance
- [ ] All exported items have proper documentation
- [ ] Acronyms are fully uppercase (AWS, CDM, DNS, NTP, VM, VPC, ID, etc.)
- [ ] Naming is consistent with existing code
- [ ] Error messages use proper formatting (`%w` for wrapping, `%s` for strings)

### GraphQL Compliance
- [ ] `.graphql` files use `RubrikPolarisSDKRequest` query/mutation name
- [ ] Results are aliased to `result:`
- [ ] Generated `queries.go` files are included if `.graphql` files changed
- [ ] No manual edits to generated files

### Region Type Safety
- [ ] Region parameters use typed regions (not strings)
- [ ] Region struct fields use typed regions (not strings)
- [ ] Invalid regions are handled properly

### Testing & Quality
- [ ] New functionality has tests
- [ ] Tests follow existing patterns
- [ ] Error cases are tested
- [ ] Code passes CI checks (vet, fmt, staticcheck)

### Completeness
- [ ] All related files are updated
- [ ] No unnecessary code changes
- [ ] Removed code is intentional
- [ ] PR description explains the changes

**See [reviewer-guide.md](./reviewer-guide.md) for detailed review guidelines.**

## Common Mistakes to Avoid

1. **Using `Id` instead of `ID`** - Always use uppercase for acronyms
2. **Forgetting to run `go generate`** - Required after modifying `.graphql` files
3. **Manually editing `queries.go`** - These files are auto-generated
4. **Using `string` for regions** - Always use typed region types
5. **Missing documentation** - All exported items need doc comments
6. **Wrong query name** - Must be `RubrikPolarisSDKRequest`
7. **Missing result alias** - Query result must be aliased to `result`

