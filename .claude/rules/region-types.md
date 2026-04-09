# Region Type Safety

ALWAYS use strongly-typed region types instead of strings for cloud provider
regions.

| Provider | Type | Package |
|----------|------|---------|
| AWS | `aws.Region` | `pkg/polaris/graphql/regions/aws` |
| Azure | `azure.Region` | `pkg/polaris/graphql/regions/azure` |
| GCP | `gcp.Region` | `pkg/polaris/graphql/regions/gcp` |

## Parse

```go
region := aws.RegionFromName("us-east-1")      // Native format
region := aws.RegionFromRegionEnum("US_EAST_1") // GraphQL enum
region := aws.RegionFromAny("us-east-1")        // Try all formats
```

## Convert

```go
regionEnum := region.ToRegionEnum()        // Standard GraphQL enum
nativeEnum := region.ToNativeRegionEnum()  // AWS/Azure native enum
```

## Validate

```go
if region == aws.RegionUnknown {
    return fmt.Errorf("invalid region")
}
```

## Special Values

- `RegionUnknown` — invalid/unrecognized region
- `RegionSource` — "same as source" for replication (AWS/Azure only)
