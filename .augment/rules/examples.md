# Code Examples and Anti-Patterns

This document provides comprehensive examples of correct and incorrect code patterns for the Rubrik Polaris SDK for Go.

## Table of Contents

1. [Documentation Examples](#documentation-examples)
2. [Acronym Capitalization Examples](#acronym-capitalization-examples)
3. [GraphQL Query Examples](#graphql-query-examples)
4. [Region Type Examples](#region-type-examples)
5. [Complete Code Examples](#complete-code-examples)

## Documentation Examples

### ✅ Correct: Exported Type with Documentation

```go
// CloudAccount represents an AWS cloud account in Rubrik Security Cloud.
// It contains the account metadata and enabled features.
type CloudAccount struct {
    // ID is the Rubrik cloud account identifier.
    ID uuid.UUID
    
    // NativeID is the AWS account ID.
    NativeID string
    
    // Name is the friendly name of the cloud account.
    Name string
    
    // Features contains the list of enabled features for this account.
    Features []Feature
}
```

### ✅ Correct: Exported Function with Documentation

```go
// AddAccount adds a new AWS cloud account to Rubrik Security Cloud.
// It returns the cloud account ID on success or an error if the operation fails.
//
// Example:
//   id, err := api.AddAccount(ctx, AccountID("123456789012"), Name("Production"))
//   if err != nil {
//       return err
//   }
func (a API) AddAccount(ctx context.Context, opts ...OptionFunc) (uuid.UUID, error) {
    // implementation
}
```

### ❌ Incorrect: Missing Documentation

```go
type CloudAccount struct {
    ID       uuid.UUID
    NativeID string
    Name     string
    Features []Feature
}

func (a API) AddAccount(ctx context.Context, opts ...OptionFunc) (uuid.UUID, error) {
    // implementation
}
```

## Acronym Capitalization Examples

### ✅ Correct: Uppercase Acronyms

```go
// Struct with proper acronym capitalization
type AWSCloudAccountConfig struct {
    CloudAccountID uuid.UUID
    VPCID         string
    SubnetIDs     []string
    IAMARN        string
    KMSKeyID      string
    S3BucketURL   string
    HTTPEndpoint  string
    JSONConfig    string
}

// Function with proper acronym capitalization
func (a API) GetAWSAccountByID(ctx context.Context, cloudAccountID uuid.UUID) (*AWSAccount, error) {
    // implementation
}

// Constants with proper acronym capitalization (type prefix pattern)
const (
    AWSFeatureEC2  = "EC2"
    AWSFeatureRDS  = "RDS"
    AWSFeatureS3   = "S3"
    AWSFeatureEKS  = "EKS"
)

// CDM-related types
type CDMCluster struct {
    ID      uuid.UUID
    URL     string
    Version string
}
```

### ❌ Incorrect: Mixed Case Acronyms

```go
// Wrong: Mixed case acronyms
type AwsCloudAccountConfig struct {
    CloudAccountId uuid.UUID
    VpcId         string
    SubnetIds     []string
    IamArn        string
    KmsKeyId      string
    S3BucketUrl   string
    HttpEndpoint  string
    JsonConfig    string
}

func (a API) GetAwsAccountById(ctx context.Context, id uuid.UUID) (*AwsAccount, error) {
    // implementation
}

const (
    AwsEc2Feature  = "EC2"
    AwsRdsFeature  = "RDS"
    AwsS3Feature   = "S3"
    AwsEksFeature  = "EKS"
)

type CdmCluster struct {
    Id      uuid.UUID
    CdmUrl  string
    CdmVersion string
}
```

## GraphQL Query Examples

### ✅ Correct: Query with Standard Name and Result Alias

**File**: `pkg/polaris/graphql/aws/queries/get_cloud_account.graphql`

```graphql
query RubrikPolarisSDKRequest($cloudAccountId: UUID!) {
  result: awsCloudAccount(cloudAccountId: $cloudAccountId) {
    id
    nativeId
    accountName
    features {
      feature
      status
      regions
    }
  }
}
```

### ✅ Correct: Query with Multiple Parameters

**File**: `pkg/polaris/graphql/aws/queries/list_exocompute_configs.graphql`

```graphql
query RubrikPolarisSDKRequest(
  $cloudAccountId: UUID!
  $region: AwsRegion!
  $feature: CloudAccountFeature!
) {
  result: awsExocomputeConfigs(
    cloudAccountId: $cloudAccountId
    region: $region
    feature: $feature
  ) {
    configId
    vpcId
    subnets {
      subnetId
      availabilityZone
    }
  }
}
```

### ✅ Correct: Mutation with Extrapolated Input Fields

**File**: `pkg/polaris/graphql/aws/queries/update_cloud_account.graphql`

```graphql
mutation RubrikPolarisSDKRequest($cloudAccountId: UUID!, $name: String!) {
  result: updateAwsCloudAccount(input: {
    cloudAccountId: $cloudAccountId,
    name: $name,
  }) {
    id
    status
  }
}
```

**Note**: Always extrapolate input fields as individual parameters instead of using complex input types. This avoids generating unnecessary Go types and keeps the API simpler.

### ❌ Incorrect: Using Complex Input Type

```graphql
mutation RubrikPolarisSDKRequest($input: UpdateAwsCloudAccountInput!) {
  result: updateAwsCloudAccount(input: $input) {
    id
    status
  }
}
```

**Why this is wrong**: Using complex input types generates unnecessary Go types and makes the API more complex.

### ❌ Incorrect: Custom Query Name

```graphql
query GetAwsCloudAccount($cloudAccountId: UUID!) {
  awsCloudAccount(cloudAccountId: $cloudAccountId) {
    id
    nativeId
  }
}
```

### ❌ Incorrect: Missing Result Alias

```graphql
query RubrikPolarisSDKRequest($cloudAccountId: UUID!) {
  awsCloudAccount(cloudAccountId: $cloudAccountId) {
    id
    nativeId
  }
}
```

## Region Type Examples

### ✅ Correct: Using Region Types

```go
import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// ExocomputeConfig uses typed region
type ExocomputeConfig struct {
    CloudAccountID uuid.UUID
    Region        aws.Region
    VPCID         string
    SubnetIDs     []string
}

// AddExocompute accepts typed region parameter
func (a API) AddExocompute(ctx context.Context, config ExocomputeConfig) error {
    // Convert to GraphQL enum when needed
    regionEnum := config.Region.ToRegionEnum()
    
    // Use in GraphQL query
    // ...
    
    return nil
}

// ParseRegionFromUser validates and converts user input
func ParseRegionFromUser(input string) (aws.Region, error) {
    region := aws.RegionFromAny(input)
    if region == aws.RegionUnknown {
        return aws.RegionUnknown, fmt.Errorf("invalid AWS region: %s", input)
    }
    return region, nil
}

// GetSupportedRegions returns a list of regions
func GetSupportedRegions() []aws.Region {
    return []aws.Region{
        aws.RegionUsEast1,
        aws.RegionUsEast2,
        aws.RegionUsWest1,
        aws.RegionUsWest2,
        aws.RegionEuWest1,
        aws.RegionEuCentral1,
    }
}
```

### ❌ Incorrect: Using Strings for Regions

```go
// Wrong: Using string for region
type ExocomputeConfig struct {
    CloudAccountID uuid.UUID
    Region        string  // Should be aws.Region
    VPCID         string
    SubnetIDs     []string
}

// Wrong: Accepting string for region
func (a API) AddExocompute(ctx context.Context, config ExocomputeConfig) error {
    // No type safety, can pass any string
    return nil
}

// Wrong: No validation
func ParseRegionFromUser(input string) (string, error) {
    // Just returns the string without validation
    return input, nil
}

// Wrong: Returning strings
func GetSupportedRegions() []string {
    return []string{
        "us-east-1",
        "us-east-2",
        "us-west-1",
    }
}
```

## Complete Code Examples

### ✅ Correct: Complete Feature Implementation

```go
package aws

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// ExocomputeConfig represents the configuration for AWS Exocompute.
type ExocomputeConfig struct {
    // CloudAccountID is the Rubrik cloud account identifier.
    CloudAccountID uuid.UUID
    
    // Region is the AWS region where Exocompute will be deployed.
    Region aws.Region
    
    // VPCID is the VPC identifier.
    VPCID string
    
    // SubnetIDs contains the subnet identifiers for Exocompute.
    SubnetIDs []string
}

// AddExocompute adds Exocompute configuration to the specified AWS cloud account.
// It returns an error if the operation fails.
func (a API) AddExocompute(ctx context.Context, config ExocomputeConfig) error {
    a.log.Print(log.Trace)
    
    // Validate configuration
    if config.CloudAccountID == uuid.Nil {
        return fmt.Errorf("cloud account ID is required")
    }
    if config.Region == aws.RegionUnknown {
        return fmt.Errorf("valid region is required")
    }
    if config.VPCID == "" {
        return fmt.Errorf("VPC ID is required")
    }
    
    // Convert region to GraphQL enum
    regionEnum := config.Region.ToRegionEnum()
    
    // Call GraphQL mutation
    err := aws.Wrap(a.client).AddExocomputeConfig(
        ctx,
        config.CloudAccountID,
        regionEnum,
        config.VPCID,
        config.SubnetIDs,
    )
    if err != nil {
        return fmt.Errorf("failed to add exocompute: %s", err)
    }
    
    return nil
}
```
