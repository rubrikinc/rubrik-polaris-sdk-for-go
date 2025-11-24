# Rubrik Polaris SDK for Go - Coding Standards

This document defines the coding standards and conventions for the Rubrik Polaris SDK for Go project.

## Documentation Requirements

### All Types and Functions Must Be Documented

**Rule**: Every exported type, struct, interface, function, method, and constant MUST have a documentation comment.

**Format**:
- Documentation comments should start with the name of the item being documented
- Use complete sentences with proper punctuation
- For complex types or functions, provide usage examples when appropriate

**Examples**:

✅ **Correct**:
```go
// CloudAccount represents an AWS cloud account in RSC.
type CloudAccount struct {
    ID       uuid.UUID
    NativeID string
    Name     string
}

// AddAccount adds a new AWS cloud account to RSC and returns the cloud account ID.
func (a API) AddAccount(ctx context.Context, account AccountFunc) (uuid.UUID, error) {
    // implementation
}
```

❌ **Incorrect**:
```go
// No documentation comment
type CloudAccount struct {
    ID       uuid.UUID
    NativeID string
}

func (a API) AddAccount(ctx context.Context, account AccountFunc) (uuid.UUID, error) {
    // implementation
}
```

### Internal/Private Items

While not strictly required, internal types and functions should also be documented when their purpose is not immediately obvious from the name.

## Acronym Capitalization

**Rule**: All acronyms in struct names, type names, variable names, and function names MUST be fully uppercase.

**Common Acronyms**:
- `CDM` not `Cdm`
- `AWS` not `Aws`
- `GCP` not `Gcp`
- `API` not `Api`
- `ID` not `Id`
- `UUID` not `Uuid`
- `URL` not `Url`
- `HTTP` not `Http`
- `JSON` not `Json`
- `VPC` not `Vpc`
- `IAM` not `Iam`
- `ARN` not `Arn`
- `EKS` not `Eks`
- `RDS` not `Rds`
- `S3` not `s3`
- `EC2` not `Ec2`

**Examples**:

✅ **Correct**:
```go
type AWSAPI struct {
    CloudAccountID uuid.UUID
    VPCID         string
    IAMARN        string
}

type CDMCluster struct {
    ID   uuid.UUID
    Name string
}

func (a API) GetAWSAccount(ctx context.Context, id uuid.UUID) (*AWSAccount, error) {
    // implementation
}
```

❌ **Incorrect**:
```go
type AwsApi struct {
    CloudAccountId uuid.UUID
    VpcId         string
    IamArn        string
}

type CdmCluster struct {
    Id   uuid.UUID
    Name string
}

func (a API) GetAwsAccount(ctx context.Context, id uuid.UUID) (*AwsAccount, error) {
    // implementation
}
```

## General Go Conventions

- Follow standard Go formatting (use `gofmt` or `goimports`)
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Handle errors explicitly; never ignore errors
- Use context.Context for cancellation and timeouts
- Prefer composition over inheritance

