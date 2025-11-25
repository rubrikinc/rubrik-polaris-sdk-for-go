---
type: "agent_requested"
description: "Rules when adding or working with regions in any of the public cloud providers"
---

# Cloud Region Type Guidelines

**Rule**: ALWAYS use strongly-typed region types (`aws.Region`, `azure.Region`, `gcp.Region`) instead of strings for cloud provider regions.

## Region Type Packages

| Provider | Package | Type |
|----------|---------|------|
| **AWS** | `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws` | `aws.Region` |
| **Azure** | `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure` | `azure.Region` |
| **GCP** | `github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp` | `gcp.Region` |

## Common Operations

All region types support the same operations:

```go
// Parse from string (use RegionFromAny for flexibility)
region := aws.RegionFromName("us-east-1")      // Parse native format
region := aws.RegionFromRegionEnum("US_EAST_1") // Parse GraphQL enum format
region := aws.RegionFromAny("us-east-1")        // Try all formats

// Convert to GraphQL enum
regionEnum := region.ToRegionEnum()             // Standard conversion
nativeEnum := region.ToNativeRegionEnum()       // Native enum (AWS/Azure only)

// Validate
if region == aws.RegionUnknown {
    return fmt.Errorf("invalid region")
}
```

## Usage Examples

### ✅ Correct - Use Region Types

```go
// Function parameter
func AddExocompute(ctx context.Context, region aws.Region) error {
    if region == aws.RegionUnknown {
        return fmt.Errorf("invalid region")
    }
    regionEnum := region.ToRegionEnum()
    // ...
}

// Struct field
type Config struct {
    Region aws.Region
}
```

### ❌ Incorrect - Don't Use Strings

```go
// DON'T: String parameter
func AddExocompute(ctx context.Context, region string) error { ... }

// DON'T: String field
type Config struct {
    Region string  // Should be aws.Region
}
```

## Special Values

- `aws.RegionUnknown`, `azure.RegionUnknown`, `gcp.RegionUnknown` - Invalid region
- `aws.RegionSource`, `azure.RegionSource` - Same as source (for replication)
