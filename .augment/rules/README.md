# Augment AI Guidelines for Rubrik Polaris SDK for Go

This directory contains guidelines and coding standards for the Rubrik Polaris SDK for Go project, specifically formatted for Augment AI to understand and enforce.

## Overview

These guidelines ensure consistency, type safety, and maintainability across the codebase. All contributors and AI assistants should follow these standards when making changes to the project.

## Guidelines Documents

### 1. [Coding Standards](./coding-standards.md)

Defines general coding conventions including:
- **Documentation Requirements**: All types and functions must have documentation comments
- **Acronym Capitalization**: All acronyms must be fully uppercase (e.g., `AWS`, `CDM`, `API`, `ID`)
- General Go best practices

**Key Rules**:
- ✅ `type AWSAPI struct` not `type AwsApi struct`
- ✅ `CloudAccountID uuid.UUID` not `CloudAccountId uuid.UUID`
- ✅ Every exported type and function must have a doc comment

### 2. [GraphQL Guidelines](./graphql-guidelines.md)

Defines standards for working with GraphQL queries:
- **Query Naming**: All queries must use `query RubrikPolarisSDKRequest` with result aliased to `result`
- **Input Handling**: Always extrapolate input fields as individual parameters (not using complex input types)
- **Code Generation**: Queries are auto-generated from `.graphql` files using `go generate ./...`
- **Never manually edit** `queries.go` files

**Key Rules**:
- ✅ Create/edit `.graphql` files in `queries/` subdirectories
- ✅ Extrapolate input fields as individual parameters
- ✅ Run `go generate ./...` from repository root after changes
- ❌ Never manually edit generated `queries.go` files
- ❌ Never manually create query variables in Go code
- ❌ Never use complex input types (e.g., `$input: UpdateAwsCloudAccountInput!`)

**Example**:
```graphql
query RubrikPolarisSDKRequest($cloudAccountId: String) {
  result: allAwsCloudAccounts(cloudAccountId: $cloudAccountId) {
    id
    name
  }
}
```

### 3. [Region Types](./region-types.md)

Defines standards for handling cloud provider regions:
- **Type Safety**: Always use region types (`aws.Region`, `azure.Region`, `gcp.Region`) instead of strings
- **Conversion**: Region types handle conversion to GraphQL enum formats
- **Validation**: Region types provide compile-time validation

**Key Rules**:
- ✅ Use `aws.Region`, `azure.Region`, or `gcp.Region` types
- ✅ Parse strings using `RegionFromName()`, `RegionFromAny()`, etc.
- ❌ Never use `string` type for region parameters or struct fields

**Example**:
```go
// Correct
func AddExocompute(ctx context.Context, region aws.Region) error {
    regionEnum := region.ToRegionEnum()
    // ...
}

// Incorrect
func AddExocompute(ctx context.Context, region string) error {
    // No type safety
}
```

### 4. [Reviewer Guide](./reviewer-guide.md)

Comprehensive guide for Augment AI to act as a code reviewer:
- **Review Checklist**: Systematic approach to reviewing code
- **Common Issues**: Patterns from historical code reviews
- **Feedback Format**: How to provide constructive feedback
- **Severity Levels**: Critical, Important, Minor, Suggestion

**When to Use**:
- Reviewing pull requests
- Providing feedback on code changes
- Ensuring code quality and consistency

**Key Focus Areas**:
- Code standards compliance
- GraphQL query standards
- Region type enforcement
- Error handling
- Testing coverage
- Generated files validation

## Quick Reference

### When Adding New Code

1. **Add documentation comments** to all exported types and functions
2. **Use uppercase acronyms** (AWS, CDM, API, ID, UUID, etc.)
3. **Use region types** instead of strings for cloud regions
4. **Follow GraphQL query conventions** when adding queries

### When Adding/Modifying GraphQL Queries

1. Create or edit `.graphql` file in appropriate `queries/` directory
2. Use `query RubrikPolarisSDKRequest` as the query name
3. Alias the result to `result`
4. Run `go generate ./...` from repository root
5. Verify the generated `queries.go` file

### When Working with Cloud Regions

1. Import the appropriate region package:
   - `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws`
   - `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure`
   - `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp`
2. Use region constants or parse from strings using `RegionFromName()` (e.g., "us-east-2")
3. Convert to the appropriate GraphQL enum type using the correct conversion function:
   - **AWS**: `.ToRegionEnum()` for `AwsRegion`, `.ToNativeRegionEnum()` for `AwsNativeRegion`
   - **Azure**: `.ToRegionEnum()` for `AzureRegion`, `.ToCloudAccountRegionEnum()` for `AzureCloudAccountRegion`, `.ToNativeRegionEnum()` for `AzureNativeRegion`
   - **GCP**: `.ToRegionEnum()` for `GcpRegion`

## Common Patterns

### Documentation Comment Pattern

```go
// TypeName describes what this type represents.
type TypeName struct {
    // Field documentation
    Field string
}

// MethodName performs a specific action and returns the result.
func (t TypeName) MethodName(ctx context.Context) error {
    // implementation
}
```

### Region Handling Pattern

#### AWS Region Examples

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// Parse region from user input (e.g., "us-east-2")
region := aws.RegionFromName(userInput)
if region == aws.RegionUnknown {
    return fmt.Errorf("invalid region: %s", userInput)
}

// Example 1: Convert to AwsRegion enum (most common)
regionEnum := region.ToRegionEnum()
err := addExocompute(ctx, cloudAccountID, regionEnum, vpcID)

// Example 2: Convert to AwsNativeRegion enum (for native protection)
nativeRegionEnum := region.ToNativeRegionEnum()
err := enableNativeProtection(ctx, cloudAccountID, nativeRegionEnum)

// Example 3: Using region constants directly
region := aws.RegionUsEast1
regionEnum := region.ToRegionEnum()

// Example 4: Get region information
name := region.Name()                    // "us-east-1"
displayName := region.DisplayName()      // "US East (N. Virginia)"
```

#### Azure Region Examples

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// Parse region from user input (e.g., "eastus")
region := azure.RegionFromName(userInput)
if region == azure.RegionUnknown {
    return fmt.Errorf("invalid region: %s", userInput)
}

// Example 1: Convert to AzureRegion enum (most common)
regionEnum := region.ToRegionEnum()
err := addExocompute(ctx, subscriptionID, regionEnum)

// Example 2: Convert to AzureCloudAccountRegion enum (for cloud account operations)
cloudAccountRegionEnum := region.ToCloudAccountRegionEnum()
err := updateCloudAccount(ctx, cloudAccountID, cloudAccountRegionEnum)

// Example 3: Convert to AzureNativeRegion enum (for native protection)
nativeRegionEnum := region.ToNativeRegionEnum()
err := enableNativeProtection(ctx, subscriptionID, nativeRegionEnum)

// Example 4: Get region information
name := region.Name()                         // "eastus"
displayName := region.DisplayName()           // "East US"
regionalName := region.RegionalDisplayName()  // "(US) East US"
```

#### GCP Region Examples

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// Parse region from user input (e.g., "us-central1")
region := gcp.RegionFromName(userInput)
if region == gcp.RegionUnknown {
    return fmt.Errorf("invalid region: %s", userInput)
}

// Example 1: Convert to GcpRegion enum
regionEnum := region.ToRegionEnum()
err := addExocompute(ctx, projectID, regionEnum)

// Example 2: Using region constants
region := gcp.RegionUSCentral1
regionEnum := region.ToRegionEnum()

// Example 3: Get region information
name := region.Name()                // "us-central1"
displayName := region.DisplayName()  // "us-central1 (Iowa, North America)"
```

### GraphQL Query Pattern

File: `pkg/polaris/graphql/aws/queries/my_query.graphql`
```graphql
query RubrikPolarisSDKRequest($param: String!) {
  result: myGraphQLOperation(param: $param) {
    field1
    field2
  }
}
```

Then run: `go generate ./...`

## Enforcement

These guidelines should be enforced by:
1. Code review
2. Augment AI when making code suggestions or changes
3. Linting and static analysis tools where applicable

## Questions?

If you're unsure about how to apply these guidelines, refer to existing code in the repository for examples, or consult the detailed guideline documents in this directory.

