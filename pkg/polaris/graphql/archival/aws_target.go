package archival

import (
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
)

// AWSTarget holds the result of an AWS target list operation. AWS target is
// referred to as Amazon S3 in the RSC UI.
type AWSTarget struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Cluster struct {
		ID uuid.UUID `json:"id"`
	} `json:"cluster"`
	Status       string `json:"status"`
	TargetType   string `json:"targetType"`
	SyncStatus   string `json:"syncStatus"`
	CloudAccount struct {
		ID uuid.UUID `json:"cloudAccountId"`
	} `json:"cloudAccount"`
	Bucket          string     `json:"bucket"`
	Region          aws.Region `json:"region"`
	StorageClass    string     `json:"storageClass"`
	RetrivalTier    string     `json:"awsRetrievalTier"`
	EncryptionType  string     `json:"encryptionType"`
	KMSMasterKeyID  string     `json:"kmsMasterKeyId"`
	ComputeSettings *struct {
		SubnetID        string `json:"subnetId"`
		SecurityGroupID string `json:"securityGroupId"`
		VPCID           string `json:"vpcId"`
		ProxySettings   *struct {
			PortNumber  int    `json:"portNumber"`
			Protocol    string `json:"protocol"`
			ProxyServer string `json:"proxyServer"`
			Username    string `json:"username"`
		} `json:"proxySettings"`
	} `json:"computeSettings"`
	IsConsolidationEnabled bool `json:"isConsolidationEnabled"`
	ProxySettings          *struct {
		PortNumber  int    `json:"portNumber"`
		Protocol    string `json:"protocol"`
		ProxyServer string `json:"proxyServer"`
		Username    string `json:"username"`
	} `json:"proxySettings"`
	BypassProxy          bool `json:"bypassProxy"`
	ImmutabilitySettings *struct {
		LockDurationDays int `json:"LockDurationDays"`
	} `json:"immutabilitySettings"`
	S3Endpoint  string `json:"s3Endpoint"`
	KMSEndpoint string `json:"kmsEndpoint"`
}

func (r AWSTarget) Validate() bool {
	// There's no filter for AWS targets, so we filter manually here.
	return r.TargetType == "AWS" && r.CloudAccount.ID != uuid.Nil
}

func (r AWSTarget) ListQuery(cursor string, filters []ListTargetFilter) (string, any) {
	return targetsQuery, struct {
		After   string             `json:"after,omitempty"`
		Filters []ListTargetFilter `json:"filters,omitempty"`
	}{After: cursor, Filters: filters}
}

// CreateAWSTargetParams holds the parameters for an AWS target create
// operation. AWS target is referred to as Amazon S3 in the RSC UI.
type CreateAWSTargetParams struct {
	Name                   string                         `json:"name"`
	ClusterID              uuid.UUID                      `json:"clusterUuid"`
	CloudAccountID         uuid.UUID                      `json:"cloudAccountId"`
	BucketName             string                         `json:"bucketName"`
	Region                 aws.Region                     `json:"region"`
	StorageClass           string                         `json:"storageClass"`
	RetrievalTier          string                         `json:"awsRetrievalTier,omitempty"`
	KMSMasterKeyID         string                         `json:"kmsMasterKeyId,omitempty"`
	RSAKey                 string                         `json:"rsaKey,omitempty"`
	EncryptionPassword     string                         `json:"encryptionPassword,omitempty"`
	CloudComputeSettings   *AWSTargetCloudComputeSettings `json:"cloudComputeSettings,omitempty"`
	IsConsolidationEnabled bool                           `json:"isConsolidationEnabled"`
	ProxySettings          *AWSTargetProxySettings        `json:"proxySettings,omitempty"`
	BypassProxy            bool                           `json:"bypassProxy"`
	ComputeProxySettings   *AWSTargetProxySettings        `json:"computeProxySettings,omitempty"`
	ImmutabilitySettings   *AWSTargetImmutabilitySettings `json:"immutabilitySettings,omitempty"`
	S3Endpoint             string                         `json:"s3Endpoint,omitempty"`
	KMSEndpoint            string                         `json:"kmsEndpoint,omitempty"`
}

// CreateAWSTargetResult holds the result of an AWS target create operation.
// AWS target is referred to as Amazon S3 in the RSC UI.
type CreateAWSTargetResult struct {
	ID string `json:"id"`
}

func (CreateAWSTargetResult) CreateQuery(createParams CreateAWSTargetParams) (string, any) {
	return createAwsTargetQuery, createParams
}

func (r CreateAWSTargetResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.ID)
}

// UpdateAWSTargetParams holds the parameters for an AWS target update
// operation. AWS target is referred to as Amazon S3 in the RSC UI.
type UpdateAWSTargetParams struct {
	Name                   string                         `json:"name,omitempty"`
	CloudAccountID         uuid.UUID                      `json:"cloudAccountId,omitempty"`
	StorageClass           string                         `json:"storageClass,omitempty"`
	RetrievalTier          string                         `json:"awsRetrievalTier,omitempty"`
	CloudComputeSettings   *AWSTargetCloudComputeSettings `json:"cloudComputeSettings,omitempty"`
	IsConsolidationEnabled bool                           `json:"isConsolidationEnabled,omitempty"`
	ProxySettings          *AWSTargetProxySettings        `json:"proxySettings,omitempty"`
	BypassProxy            bool                           `json:"bypassProxy,omitempty"`
	ComputeProxySettings   *AWSTargetProxySettings        `json:"computeProxySettings,omitempty"`
	ImmutabilitySettings   *AWSTargetImmutabilitySettings `json:"immutabilitySettings,omitempty"`
	S3Endpoint             string                         `json:"s3Endpoint,omitempty"`
	KMSEndpoint            string                         `json:"kmsEndpoint,omitempty"`
	ComputeSettingsID      string                         `json:"computeSettingsId,omitempty"`
	IAMPairID              string                         `json:"iamPairId,omitempty"`
}

// UpdateAWSTargetResult holds the result of an AWS target update operation.
// AWS target is referred to as Amazon S3 in the RSC UI.
type UpdateAWSTargetResult CreateAWSTargetResult

func (UpdateAWSTargetResult) UpdateQuery(targetID uuid.UUID, updateParams UpdateAWSTargetParams) (string, any) {
	return updateAwsTargetQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateAWSTargetParams
	}{ID: targetID, UpdateAWSTargetParams: updateParams}
}

func (r UpdateAWSTargetResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.ID)
}

// AWSTargetCloudComputeSettings holds the cloud compute settings for an AWS
// target.
type AWSTargetCloudComputeSettings struct {
	VPCID           string `json:"vpcId"`
	SubnetID        string `json:"subnetId"`
	SecurityGroupID string `json:"securityGroupId"`
}

// AWSTargetProxySettings holds the proxy settings for an AWS target.
type AWSTargetProxySettings struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ProxyServer string `json:"proxyServer"`
	Protocol    string `json:"protocol"`
	PortNumber  int    `json:"portNumber"`
}

// AWSTargetImmutabilitySettings holds the immutability settings for an AWS
// target.
type AWSTargetImmutabilitySettings struct {
	LockDurationDays int `json:"lockDurationDays"`
}
