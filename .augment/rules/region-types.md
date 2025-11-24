# Cloud Region Type Guidelines

This document defines the standards for handling cloud provider regions in the Rubrik Polaris SDK for Go.

## Region Type Enforcement

**Rule**: When working with cloud provider locations or regions (AWS, Azure, GCP), ALWAYS use strongly-typed region types instead of strings, even when the underlying GraphQL API expects a string.

### Why Use Region Types?

1. **Type Safety**: Catch invalid regions at compile time
2. **Consistency**: Ensure region values are valid across the codebase
3. **Documentation**: Region types provide clear documentation of valid values
4. **Conversion**: Region types handle conversion between different GraphQL enum formats

## Available Region Types

### AWS Regions

**Package**: `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws`

**Type**: `aws.Region`

**Usage**:

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// Define a region
region := aws.RegionUsEast1

// Parse from string
region := aws.RegionFromName("us-east-1")
region := aws.RegionFromRegionEnum("US_EAST_1")
region := aws.RegionFromAny("us-east-1") // Tries all formats

// Get region information
name := region.Name()                    // "us-east-1"
displayName := region.DisplayName()      // "US East (N. Virginia)"

// Convert to GraphQL enum types
regionEnum := region.ToRegionEnum()           // For AwsRegion GraphQL enum
nativeEnum := region.ToNativeRegionEnum()     // For AwsNativeRegion GraphQL enum
```

**Common AWS Regions**:
- `aws.RegionUsEast1`
- `aws.RegionUsEast2`
- `aws.RegionUsWest1`
- `aws.RegionUsWest2`
- `aws.RegionEuWest1`
- `aws.RegionEuCentral1`
- `aws.RegionApSouthEast1`
- etc.

### Azure Regions

**Package**: `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure`

**Type**: `azure.Region`

**Usage**:

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// Define a region
region := azure.RegionEastUS

// Parse from string
region := azure.RegionFromName("eastus")
region := azure.RegionFromRegionEnum("EAST_US")
region := azure.RegionFromAny("eastus")

// Get region information
name := region.Name()                         // "eastus"
displayName := region.DisplayName()           // "East US"
regionalName := region.RegionalDisplayName()  // "(US) East US"

// Convert to GraphQL enum types
regionEnum := region.ToRegionEnum()                    // For AzureRegion
cloudAccountEnum := region.ToCloudAccountRegionEnum()  // For AzureCloudAccountRegion
nativeEnum := region.ToNativeRegionEnum()              // For AzureNativeRegion
```

**Common Azure Regions**:
- `azure.RegionEastUS`
- `azure.RegionEastUS2`
- `azure.RegionWestUS`
- `azure.RegionWestEurope`
- `azure.RegionNorthEurope`
- etc.

### GCP Regions

**Package**: `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp`

**Type**: `gcp.Region`

**Usage**:

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// Define a region
region := gcp.RegionUSCentral1

// Parse from string
region := gcp.RegionFromName("us-central1")
region := gcp.RegionFromRegionEnum("US_CENTRAL1")
region := gcp.RegionFromAny("us-central1")

// Get region information
name := region.Name()                // "us-central1"
displayName := region.DisplayName()  // "us-central1 (Iowa, North America)"

// Convert to GraphQL enum type
regionEnum := region.ToRegionEnum()  // For GcpRegion GraphQL enum
```

**Common GCP Regions**:
- `gcp.RegionUSCentral1`
- `gcp.RegionUSEast1`
- `gcp.RegionUSWest1`
- `gcp.RegionEuropeWest1`
- `gcp.RegionAsiaSouthEast1`
- etc.

## Examples

### ✅ Correct Usage

```go
import (
    "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// Function parameter uses region type
func AddExocompute(ctx context.Context, region aws.Region, vpcID string) error {
    // Convert to appropriate GraphQL enum when needed
    regionEnum := region.ToRegionEnum()
    
    // Use in GraphQL query
    // ...
}

// Struct field uses region type
type ExocomputeConfig struct {
    Region aws.Region
    VPCID  string
}

// Parse user input
func ParseRegionInput(input string) (aws.Region, error) {
    region := aws.RegionFromAny(input)
    if region == aws.RegionUnknown {
        return aws.RegionUnknown, fmt.Errorf("invalid region: %s", input)
    }
    return region, nil
}
```

### ❌ Incorrect Usage

```go
// DON'T: Using string for region
func AddExocompute(ctx context.Context, region string, vpcID string) error {
    // No type safety, can pass invalid values
}

// DON'T: String field for region
type ExocomputeConfig struct {
    Region string  // Should be aws.Region
    VPCID  string
}

// DON'T: Accepting string without validation
func ParseRegionInput(input string) (string, error) {
    // No validation of region value
    return input, nil
}
```

## Working with Multiple Regions

```go
// Slice of regions
regions := []aws.Region{
    aws.RegionUsEast1,
    aws.RegionUsWest2,
    aws.RegionEuWest1,
}

// Convert to GraphQL enum slice
regionEnums := make([]aws.RegionEnum, len(regions))
for i, region := range regions {
    regionEnums[i] = region.ToRegionEnum()
}
```

## Special Region Values

Each cloud provider has special region constants:

- `aws.RegionUnknown` - Unknown or invalid region
- `aws.RegionSource` - Same as source (for replication)
- `azure.RegionUnknown` - Unknown or invalid region
- `azure.RegionSource` - Same as source (for replication)
- `gcp.RegionUnknown` - Unknown or invalid region

