//go:generate go run ../queries_gen.go cloudcluster

// Copyright 2021 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package cloudcluster

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC CloudCluster API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Azure API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ClusterCreateValidations represents the valid cluster create validations.
type ClusterCreateValidations string

const (
	NoChecks                   ClusterCreateValidations = "NO_CHECKS"
	AllChecks                  ClusterCreateValidations = "ALL_CHECKS"
	ClusterNameCheck           ClusterCreateValidations = "CLUSTER_NAME_CHECK"
	DNSServersCheck            ClusterCreateValidations = "DNS_SERVERS_CHECK"
	NTPServersCheck            ClusterCreateValidations = "NTP_SERVERS_CHECK"
	NodeCountCheck             ClusterCreateValidations = "NODE_COUNT_CHECK"
	ObjectStoreCheck           ClusterCreateValidations = "OBJECT_STORE_CHECK"
	ImmutabilityCheck          ClusterCreateValidations = "IMMUTABILITY_CHECK"
	AWSInstanceProfileCheck    ClusterCreateValidations = "AWS_INSTANCE_PROFILE_CHECK"
	AWSNetworkConfigCheck      ClusterCreateValidations = "AWS_NETWORK_CONFIG_CHECK"
	AzureVMImageCheck          ClusterCreateValidations = "AZURE_VM_IMAGE_CHECK"
	AzureAvailabilityZoneCheck ClusterCreateValidations = "AZURE_AVAILABILITY_ZONE_CHECK"
	AzureQuotaCheck            ClusterCreateValidations = "AZURE_QUOTA_CHECK"
	AzureMICheck               ClusterCreateValidations = "AZURE_MI_CHECK"
	CloudAccountCheck          ClusterCreateValidations = "CLOUD_ACCOUNT_CHECK"
	GCPNetworkConfigCheck      ClusterCreateValidations = "GCP_NETWORK_CONFIG_CHECK"
	GCPSerivceAccountCheck     ClusterCreateValidations = "GCP_SERVICE_ACCOUNT_CHECK"
	GCPInstanceLabelKeyCheck   ClusterCreateValidations = "GCP_INSTANCE_LABEL_KEY_CHECK"
	GCPClusterNameLengthCheck  ClusterCreateValidations = "GCP_CLUSTER_NAME_LENGTH_CHECK"
)

type ClusterProduct string

const (
	CDM          ClusterProduct = "CDM"
	CLOUD_DIRECT ClusterProduct = "CLOUD_DIRECT"
	DATOS        ClusterProduct = "DATOS"
	POLARIS      ClusterProduct = "POLARIS"
)

type ClusterProductType string

const (
	CLOUD      ClusterProductType = "Cloud"
	RSC        ClusterProductType = "Polaris"
	EXOCOMPUTE ClusterProductType = "ExoCompute"
	ONPREM     ClusterProductType = "OnPrem"
	ROBO       ClusterProductType = "Robo"
	UNKNOWN    ClusterProductType = "Unknown"
)

type ClusterStatus string

const (
	ClusterConnected    ClusterStatus = "Connected"
	ClusterDisconnected ClusterStatus = "Disconnected"
	ClusterInitializing ClusterStatus = "Initializing"
)

type ClusterSystemStatus string

const (
	ClusterSystemStatusOK      ClusterSystemStatus = "OK"
	ClusterSystemStatusWARNING ClusterSystemStatus = "WARNING"
	ClusterSystemStatusFATAL   ClusterSystemStatus = "FATAL"
)

type ClusterFilter struct {
	ID              []string              `json:"id"`
	Name            []string              `json:"name"`
	Type            []ClusterProductType  `json:"type"`
	ConnectionState []ClusterStatus       `json:"connectionState"`
	SystemStatus    []ClusterSystemStatus `json:"systemStatus"`
	ProductType     []ClusterProduct      `json:"productType"`
}

// ClusterSortByEnum represents the valid sort by values.
type ClusterSortBy string

const (
	SortByEstimatedRunway  ClusterSortBy = "ESTIMATED_RUNWAY"
	SortByInstalledVersion ClusterSortBy = "INSTALLED_VERSION"
	SortByClusterName      ClusterSortBy = "ClusterName"
	SortByClusterType      ClusterSortBy = "ClusterType"
	SortByClusterLocation  ClusterSortBy = "CLUSTER_LOCATION"
	SortByRegisteredAt     ClusterSortBy = "RegisteredAt"
)

// VmConfigType represents the valid VM config types.
type VmConfigType string

// VmConfigType values.
const (
	CCVmConfigStandard   VmConfigType = "STANDARD"
	CCVmConfigDense      VmConfigType = "DENSE"
	CCVmConfigExtraDense VmConfigType = "EXTRA_DENSE"
)

// AwsEsConfigInput represents the input for creating an AWS ES config.
type AwsEsConfigInput struct {
	BucketName         string `json:"bucketName"`
	ShouldCreateBucket bool   `json:"shouldCreateBucket"`
	EnableImmutability bool   `json:"enableImmutability"`
	EnableObjectLock   bool   `json:"enableObjectLock"`
}

// AzureManagedIdentityName represents the input for creating an Azure managed identity.
type AzureManagedIdentityName struct {
	ClientID      string `json:"clientId"`
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
}

// AzureEsConfigInput represents the input for creating an Azure ES config.
type AzureEsConfigInput struct {
	ContainerName         string                   `json:"containerName"`
	EnableImmutability    bool                     `json:"enableImmutability"`
	EndpointSuffix        string                   `json:"endpointSuffix,omitempty"`
	ManagedIdentity       AzureManagedIdentityName `json:"managedIdentity"`
	ResourceGroup         string                   `json:"resourceGroup"`
	ShouldCreateContainer bool                     `json:"shouldCreateContainer"`
	StorageAccount        string                   `json:"storageAccount"`
	StorageSecret         secret.String            `json:"storageSecret,omitempty"`
}

// GcpEsConfigInput represents the input for creating a GCP ES config.
type GcpEsConfigInput struct {
	BucketName         string `json:"bucketName"`
	Region             string `json:"region"`
	ShouldCreateBucket bool   `json:"shouldCreateBucket"`
}

// OciEsConfigInput represents the input for creating an OCI ES config.
type OciEsConfigInput struct {
	AccessKey    string `json:"accessKey"`
	BucketName   string `json:"bucketName"`
	OciNamespace string `json:"ociNamespace"`
	SecretKey    string `json:"secretKey"`
}

// CcpJobStatus represents the valid job statuses.
type CcpJobStatus string

const (
	CcpJobStatusInitializing               CcpJobStatus = "INITIALIZING"
	CcpJobStatusNodeCreate                 CcpJobStatus = "NODE_CREATE"
	CcpJobStatusNodeConnectionVerification CcpJobStatus = "NODE_CONNECTION_VERIFICATION"
	CcpJobStatusNodeInfoExtraction         CcpJobStatus = "NODE_INFO_EXTRACTION"
	CcpJobStatusBootstrapping              CcpJobStatus = "BOOTSTRAPPING"
	CcpJobStatusRotateToken                CcpJobStatus = "ROTATE_TOKEN"
	CcpJobStatusFailed                     CcpJobStatus = "FAILED"
	CcpJobStatusCompleted                  CcpJobStatus = "COMPLETED"
	CcpJobStatusInvalid                    CcpJobStatus = "INVALID"
)

// CcpJobType represents the valid job types.
type CcpJobType string

const (
	CcpJobTypeClusterCreate                   CcpJobType = "CLUSTER_CREATE"
	CcpJobTypeClusterDelete                   CcpJobType = "CLUSTER_DELETE"
	CcpJobTypeAddNode                         CcpJobType = "ADD_NODE"
	CcpJobTypeRemoveNode                      CcpJobType = "REMOVE_NODE"
	CcpJobTypeReplaceNode                     CcpJobType = "REPLACE_NODE"
	CcpJobTypeClusterRecover                  CcpJobType = "CLUSTER_RECOVER"
	CcpJobTypeClusterOps                      CcpJobType = "CLUSTER_OPS"
	CcpJobTypeMigrateNodes                    CcpJobType = "MIGRATE_NODES"
	CcpJobTypeMigrateClusterToManagedIdentity CcpJobType = "MIGRATE_CLUSTER_TO_MANAGED_IDENTITY"
	CcpJobTypeManualAddNodes                  CcpJobType = "MANUAL_ADD_NODES"
)

// ClusterTimezoneType represents the valid cluster timezones.
type ClusterTimezoneType string

const (
	CLUSTER_TIMEZONE_UNSPECIFIED                    ClusterTimezoneType = "CLUSTER_TIMEZONE_UNSPECIFIED"
	CLUSTER_TIMEZONE_AFRICA_JOHANNESBURG            ClusterTimezoneType = "AFRICA_JOHANNESBURG"
	CLUSTER_TIMEZONE_AFRICA_LAGOS                   ClusterTimezoneType = "AFRICA_LAGOS"
	CLUSTER_TIMEZONE_AFRICA_NAIROBI                 ClusterTimezoneType = "AFRICA_NAIROBI"
	CLUSTER_TIMEZONE_AMERICA_ANCHORAGE              ClusterTimezoneType = "AMERICA_ANCHORAGE"
	CLUSTER_TIMEZONE_AMERICA_ARAGUAINA              ClusterTimezoneType = "AMERICA_ARAGUAINA"
	CLUSTER_TIMEZONE_AMERICA_ATIKOKAN               ClusterTimezoneType = "AMERICA_ATIKOKAN"
	CLUSTER_TIMEZONE_AMERICA_BARBADOS               ClusterTimezoneType = "AMERICA_BARBADOS"
	CLUSTER_TIMEZONE_AMERICA_BOGOTA                 ClusterTimezoneType = "AMERICA_BOGOTA"
	CLUSTER_TIMEZONE_AMERICA_CHICAGO                ClusterTimezoneType = "AMERICA_CHICAGO"
	CLUSTER_TIMEZONE_AMERICA_COSTA_RICA             ClusterTimezoneType = "AMERICA_COSTA_RICA"
	CLUSTER_TIMEZONE_AMERICA_DENVER                 ClusterTimezoneType = "AMERICA_DENVER"
	CLUSTER_TIMEZONE_AMERICA_LOS_ANGELES            ClusterTimezoneType = "AMERICA_LOS_ANGELES"
	CLUSTER_TIMEZONE_AMERICA_MEXICO_CITY            ClusterTimezoneType = "AMERICA_MEXICO_CITY"
	CLUSTER_TIMEZONE_AMERICA_NEW_YORK               ClusterTimezoneType = "AMERICA_NEW_YORK"
	CLUSTER_TIMEZONE_AMERICA_NORONHA                ClusterTimezoneType = "AMERICA_NORONHA"
	CLUSTER_TIMEZONE_AMERICA_PANAMA                 ClusterTimezoneType = "AMERICA_PANAMA"
	CLUSTER_TIMEZONE_AMERICA_PHOENIX                ClusterTimezoneType = "AMERICA_PHOENIX"
	CLUSTER_TIMEZONE_AMERICA_SANTIAGO               ClusterTimezoneType = "AMERICA_SANTIAGO"
	CLUSTER_TIMEZONE_AMERICA_ST_JOHNS               ClusterTimezoneType = "AMERICA_ST_JOHNS"
	CLUSTER_TIMEZONE_AMERICA_TORONTO                ClusterTimezoneType = "AMERICA_TORONTO"
	CLUSTER_TIMEZONE_AMERICA_VANCOUVER              ClusterTimezoneType = "AMERICA_VANCOUVER"
	CLUSTER_TIMEZONE_ASIA_BANGKOK                   ClusterTimezoneType = "ASIA_BANGKOK"
	CLUSTER_TIMEZONE_ASIA_DHAKA                     ClusterTimezoneType = "ASIA_DHAKA"
	CLUSTER_TIMEZONE_ASIA_DUBAI                     ClusterTimezoneType = "ASIA_DUBAI"
	CLUSTER_TIMEZONE_ASIA_HONG_KONG                 ClusterTimezoneType = "ASIA_HONG_KONG"
	CLUSTER_TIMEZONE_ASIA_KARACHI                   ClusterTimezoneType = "ASIA_KARACHI"
	CLUSTER_TIMEZONE_ASIA_KATHMANDU                 ClusterTimezoneType = "ASIA_KATHMANDU"
	CLUSTER_TIMEZONE_ASIA_KOLKATA                   ClusterTimezoneType = "ASIA_KOLKATA"
	CLUSTER_TIMEZONE_ASIA_MAGADAN                   ClusterTimezoneType = "ASIA_MAGADAN"
	CLUSTER_TIMEZONE_ASIA_SINGAPORE                 ClusterTimezoneType = "ASIA_SINGAPORE"
	CLUSTER_TIMEZONE_ASIA_TOKYO                     ClusterTimezoneType = "ASIA_TOKYO"
	CLUSTER_TIMEZONE_ATLANTIC_CAPE_VERDE            ClusterTimezoneType = "ATLANTIC_CAPE_VERDE"
	CLUSTER_TIMEZONE_AUSTRALIA_ADELAIDE             ClusterTimezoneType = "AUSTRALIA_ADELAIDE"
	CLUSTER_TIMEZONE_AUSTRALIA_BRISBANE             ClusterTimezoneType = "AUSTRALIA_BRISBANE"
	CLUSTER_TIMEZONE_AUSTRALIA_PERTH                ClusterTimezoneType = "AUSTRALIA_PERTH"
	CLUSTER_TIMEZONE_AUSTRALIA_SYDNEY               ClusterTimezoneType = "AUSTRALIA_SYDNEY"
	CLUSTER_TIMEZONE_EUROPE_AMSTERDAM               ClusterTimezoneType = "EUROPE_AMSTERDAM"
	CLUSTER_TIMEZONE_EUROPE_ATHENS                  ClusterTimezoneType = "EUROPE_ATHENS"
	CLUSTER_TIMEZONE_EUROPE_LONDON                  ClusterTimezoneType = "EUROPE_LONDON"
	CLUSTER_TIMEZONE_EUROPE_MOSCOW                  ClusterTimezoneType = "EUROPE_MOSCOW"
	CLUSTER_TIMEZONE_PACIFIC_AUCKLAND               ClusterTimezoneType = "PACIFIC_AUCKLAND"
	CLUSTER_TIMEZONE_PACIFIC_HONOLULU               ClusterTimezoneType = "PACIFIC_HONOLULU"
	CLUSTER_TIMEZONE_PACIFIC_MIDWAY                 ClusterTimezoneType = "PACIFIC_MIDWAY"
	CLUSTER_TIMEZONE_UTC                            ClusterTimezoneType = "UTC"
	CLUSTER_TIMEZONE_AFRICA_ABIDJAN                 ClusterTimezoneType = "AFRICA_ABIDJAN"
	CLUSTER_TIMEZONE_AFRICA_ALGIERS                 ClusterTimezoneType = "AFRICA_ALGIERS"
	CLUSTER_TIMEZONE_AFRICA_BISSAU                  ClusterTimezoneType = "AFRICA_BISSAU"
	CLUSTER_TIMEZONE_AFRICA_CEUTA                   ClusterTimezoneType = "AFRICA_CEUTA"
	CLUSTER_TIMEZONE_AFRICA_MAPUTO                  ClusterTimezoneType = "AFRICA_MAPUTO"
	CLUSTER_TIMEZONE_AFRICA_MONROVIA                ClusterTimezoneType = "AFRICA_MONROVIA"
	CLUSTER_TIMEZONE_AFRICA_NDJAMENA                ClusterTimezoneType = "AFRICA_NDJAMENA"
	CLUSTER_TIMEZONE_AFRICA_SAO_TOME                ClusterTimezoneType = "AFRICA_SAO_TOME"
	CLUSTER_TIMEZONE_AFRICA_TRIPOLI                 ClusterTimezoneType = "AFRICA_TRIPOLI"
	CLUSTER_TIMEZONE_AFRICA_TUNIS                   ClusterTimezoneType = "AFRICA_TUNIS"
	CLUSTER_TIMEZONE_AFRICA_WINDHOEK                ClusterTimezoneType = "AFRICA_WINDHOEK"
	CLUSTER_TIMEZONE_AMERICA_ADAK                   ClusterTimezoneType = "AMERICA_ADAK"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_BUENOS_AIRES ClusterTimezoneType = "AMERICA_ARGENTINA_BUENOS_AIRES"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_CATAMARCA    ClusterTimezoneType = "AMERICA_ARGENTINA_CATAMARCA"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_CORDOBA      ClusterTimezoneType = "AMERICA_ARGENTINA_CORDOBA"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_JUJUY        ClusterTimezoneType = "AMERICA_ARGENTINA_JUJUY"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_LA_RIOJA     ClusterTimezoneType = "AMERICA_ARGENTINA_LA_RIOJA"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_MENDOZA      ClusterTimezoneType = "AMERICA_ARGENTINA_MENDOZA"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_RIO_GALLEGOS ClusterTimezoneType = "AMERICA_ARGENTINA_RIO_GALLEGOS"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_SALTA        ClusterTimezoneType = "AMERICA_ARGENTINA_SALTA"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_SAN_JUAN     ClusterTimezoneType = "AMERICA_ARGENTINA_SAN_JUAN"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_SAN_LUIS     ClusterTimezoneType = "AMERICA_ARGENTINA_SAN_LUIS"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_TUCUMAN      ClusterTimezoneType = "AMERICA_ARGENTINA_TUCUMAN"
	CLUSTER_TIMEZONE_AMERICA_ARGENTINA_USHUAIA      ClusterTimezoneType = "AMERICA_ARGENTINA_USHUAIA"
	CLUSTER_TIMEZONE_AMERICA_BAHIA                  ClusterTimezoneType = "AMERICA_BAHIA"
	CLUSTER_TIMEZONE_AMERICA_BELEM                  ClusterTimezoneType = "AMERICA_BELEM"
	CLUSTER_TIMEZONE_AMERICA_BELIZE                 ClusterTimezoneType = "AMERICA_BELIZE"
	CLUSTER_TIMEZONE_AMERICA_BOA_VISTA              ClusterTimezoneType = "AMERICA_BOA_VISTA"
	CLUSTER_TIMEZONE_AMERICA_BOISE                  ClusterTimezoneType = "AMERICA_BOISE"
	CLUSTER_TIMEZONE_AMERICA_CAMBRIDGE_BAY          ClusterTimezoneType = "AMERICA_CAMBRIDGE_BAY"
	CLUSTER_TIMEZONE_AMERICA_CAMPO_GRANDE           ClusterTimezoneType = "AMERICA_CAMPO_GRANDE"
	CLUSTER_TIMEZONE_AMERICA_CANCUN                 ClusterTimezoneType = "AMERICA_CANCUN"
	CLUSTER_TIMEZONE_AMERICA_CARACAS                ClusterTimezoneType = "AMERICA_CARACAS"
	CLUSTER_TIMEZONE_AMERICA_CAYENNE                ClusterTimezoneType = "AMERICA_CAYENNE"
	CLUSTER_TIMEZONE_AMERICA_CHIHUAHUA              ClusterTimezoneType = "AMERICA_CHIHUAHUA"
	CLUSTER_TIMEZONE_AMERICA_CIUDAD_JUAREZ          ClusterTimezoneType = "AMERICA_CIUDAD_JUAREZ"
	CLUSTER_TIMEZONE_AMERICA_CUIABA                 ClusterTimezoneType = "AMERICA_CUIABA"
	CLUSTER_TIMEZONE_AMERICA_DANMARKSHAVN           ClusterTimezoneType = "AMERICA_DANMARKSHAVN"
	CLUSTER_TIMEZONE_AMERICA_DAWSON                 ClusterTimezoneType = "AMERICA_DAWSON"
	CLUSTER_TIMEZONE_AMERICA_DAWSON_CREEK           ClusterTimezoneType = "AMERICA_DAWSON_CREEK"
	CLUSTER_TIMEZONE_AMERICA_DETROIT                ClusterTimezoneType = "AMERICA_DETROIT"
	CLUSTER_TIMEZONE_AMERICA_EDMONTON               ClusterTimezoneType = "AMERICA_EDMONTON"
	CLUSTER_TIMEZONE_AMERICA_EIRUNEPE               ClusterTimezoneType = "AMERICA_EIRUNEPE"
	CLUSTER_TIMEZONE_AMERICA_EL_SALVADOR            ClusterTimezoneType = "AMERICA_EL_SALVADOR"
	CLUSTER_TIMEZONE_AMERICA_FORT_NELSON            ClusterTimezoneType = "AMERICA_FORT_NELSON"
	CLUSTER_TIMEZONE_AMERICA_FORTALEZA              ClusterTimezoneType = "AMERICA_FORTALEZA"
	CLUSTER_TIMEZONE_AMERICA_GLACE_BAY              ClusterTimezoneType = "AMERICA_GLACE_BAY"
	CLUSTER_TIMEZONE_AMERICA_GOOSE_BAY              ClusterTimezoneType = "AMERICA_GOOSE_BAY"
	CLUSTER_TIMEZONE_AMERICA_GRAND_TURK             ClusterTimezoneType = "AMERICA_GRAND_TURK"
	CLUSTER_TIMEZONE_AMERICA_GUATEMALA              ClusterTimezoneType = "AMERICA_GUATEMALA"
	CLUSTER_TIMEZONE_AMERICA_GUAYAQUIL              ClusterTimezoneType = "AMERICA_GUAYAQUIL"
	CLUSTER_TIMEZONE_AMERICA_GUYANA                 ClusterTimezoneType = "AMERICA_GUYANA"
	CLUSTER_TIMEZONE_AMERICA_HALIFAX                ClusterTimezoneType = "AMERICA_HALIFAX"
	CLUSTER_TIMEZONE_AMERICA_HAVANA                 ClusterTimezoneType = "AMERICA_HAVANA"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_INDIANAPOLIS   ClusterTimezoneType = "AMERICA_INDIANA_INDIANAPOLIS"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_KNOX           ClusterTimezoneType = "AMERICA_INDIANA_KNOX"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_MARENGO        ClusterTimezoneType = "AMERICA_INDIANA_MARENGO"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_PETERSBURG     ClusterTimezoneType = "AMERICA_INDIANA_PETERSBURG"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_TELL_CITY      ClusterTimezoneType = "AMERICA_INDIANA_TELL_CITY"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_VEVAY          ClusterTimezoneType = "AMERICA_INDIANA_VEVAY"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_VINCENNES      ClusterTimezoneType = "AMERICA_INDIANA_VINCENNES"
	CLUSTER_TIMEZONE_AMERICA_INDIANA_WINAMAC        ClusterTimezoneType = "AMERICA_INDIANA_WINAMAC"
	CLUSTER_TIMEZONE_AMERICA_INUVIK                 ClusterTimezoneType = "AMERICA_INUVIK"
	CLUSTER_TIMEZONE_AMERICA_IQALUIT                ClusterTimezoneType = "AMERICA_IQALUIT"
	CLUSTER_TIMEZONE_AMERICA_JAMAICA                ClusterTimezoneType = "AMERICA_JAMAICA"
	CLUSTER_TIMEZONE_AMERICA_JUNEAU                 ClusterTimezoneType = "AMERICA_JUNEAU"
	CLUSTER_TIMEZONE_AMERICA_KENTUCKY_LOUISVILLE    ClusterTimezoneType = "AMERICA_KENTUCKY_LOUISVILLE"
	CLUSTER_TIMEZONE_AMERICA_KENTUCKY_MONTICELLO    ClusterTimezoneType = "AMERICA_KENTUCKY_MONTICELLO"
	CLUSTER_TIMEZONE_AMERICA_LA_PAZ                 ClusterTimezoneType = "AMERICA_LA_PAZ"
	CLUSTER_TIMEZONE_AMERICA_LIMA                   ClusterTimezoneType = "AMERICA_LIMA"
	CLUSTER_TIMEZONE_AMERICA_MACEIO                 ClusterTimezoneType = "AMERICA_MACEIO"
	CLUSTER_TIMEZONE_AMERICA_MANAGUA                ClusterTimezoneType = "AMERICA_MANAGUA"
	CLUSTER_TIMEZONE_AMERICA_MANAUS                 ClusterTimezoneType = "AMERICA_MANAUS"
	CLUSTER_TIMEZONE_AMERICA_MARTINIQUE             ClusterTimezoneType = "AMERICA_MARTINIQUE"
	CLUSTER_TIMEZONE_AMERICA_MATAMOROS              ClusterTimezoneType = "AMERICA_MATAMOROS"
	CLUSTER_TIMEZONE_AMERICA_MENOMINEE              ClusterTimezoneType = "AMERICA_MENOMINEE"
	CLUSTER_TIMEZONE_AMERICA_MERIDA                 ClusterTimezoneType = "AMERICA_MERIDA"
	CLUSTER_TIMEZONE_AMERICA_METLAKATLA             ClusterTimezoneType = "AMERICA_METLAKATLA"
	CLUSTER_TIMEZONE_AMERICA_MIQUELON               ClusterTimezoneType = "AMERICA_MIQUELON"
	CLUSTER_TIMEZONE_AMERICA_MONCTON                ClusterTimezoneType = "AMERICA_MONCTON"
	CLUSTER_TIMEZONE_AMERICA_MONTERREY              ClusterTimezoneType = "AMERICA_MONTERREY"
	CLUSTER_TIMEZONE_AMERICA_MONTEVIDEO             ClusterTimezoneType = "AMERICA_MONTEVIDEO"
	CLUSTER_TIMEZONE_AMERICA_NOME                   ClusterTimezoneType = "AMERICA_NOME"
	CLUSTER_TIMEZONE_AMERICA_NORTH_DAKOTA_BEULAH    ClusterTimezoneType = "AMERICA_NORTH_DAKOTA_BEULAH"
	CLUSTER_TIMEZONE_AMERICA_NORTH_DAKOTA_CENTER    ClusterTimezoneType = "AMERICA_NORTH_DAKOTA_CENTER"
	CLUSTER_TIMEZONE_AMERICA_NORTH_DAKOTA_NEW_SALEM ClusterTimezoneType = "AMERICA_NORTH_DAKOTA_NEW_SALEM"
	CLUSTER_TIMEZONE_AMERICA_OJINAGA                ClusterTimezoneType = "AMERICA_OJINAGA"
	CLUSTER_TIMEZONE_AMERICA_PARAMARIBO             ClusterTimezoneType = "AMERICA_PARAMARIBO"
	CLUSTER_TIMEZONE_AMERICA_PORTO_VELHO            ClusterTimezoneType = "AMERICA_PORTO_VELHO"
	CLUSTER_TIMEZONE_AMERICA_PUERTO_RICO            ClusterTimezoneType = "AMERICA_PUERTO_RICO"
	CLUSTER_TIMEZONE_AMERICA_PUNTA_ARENAS           ClusterTimezoneType = "AMERICA_PUNTA_ARENAS"
	CLUSTER_TIMEZONE_AMERICA_RANKIN_INLET           ClusterTimezoneType = "AMERICA_RANKIN_INLET"
	CLUSTER_TIMEZONE_AMERICA_RECIFE                 ClusterTimezoneType = "AMERICA_RECIFE"
	CLUSTER_TIMEZONE_AMERICA_REGINA                 ClusterTimezoneType = "AMERICA_REGINA"
	CLUSTER_TIMEZONE_AMERICA_RESOLUTE               ClusterTimezoneType = "AMERICA_RESOLUTE"
	CLUSTER_TIMEZONE_AMERICA_RIO_BRANCO             ClusterTimezoneType = "AMERICA_RIO_BRANCO"
	CLUSTER_TIMEZONE_AMERICA_SANTAREM               ClusterTimezoneType = "AMERICA_SANTAREM"
	CLUSTER_TIMEZONE_AMERICA_SAO_PAULO              ClusterTimezoneType = "AMERICA_SAO_PAULO"
	CLUSTER_TIMEZONE_AMERICA_SCORESBYSUND           ClusterTimezoneType = "AMERICA_SCORESBYSUND"
	CLUSTER_TIMEZONE_AMERICA_SITKA                  ClusterTimezoneType = "AMERICA_SITKA"
	CLUSTER_TIMEZONE_AMERICA_SWIFT_CURRENT          ClusterTimezoneType = "AMERICA_SWIFT_CURRENT"
	CLUSTER_TIMEZONE_AMERICA_TEGUCIGALPA            ClusterTimezoneType = "AMERICA_TEGUCIGALPA"
	CLUSTER_TIMEZONE_AMERICA_THULE                  ClusterTimezoneType = "AMERICA_THULE"
	CLUSTER_TIMEZONE_AMERICA_TIJUANA                ClusterTimezoneType = "AMERICA_TIJUANA"
	CLUSTER_TIMEZONE_AMERICA_WHITEHORSE             ClusterTimezoneType = "AMERICA_WHITEHORSE"
	CLUSTER_TIMEZONE_AMERICA_WINNIPEG               ClusterTimezoneType = "AMERICA_WINNIPEG"
	CLUSTER_TIMEZONE_AMERICA_YAKUTAT                ClusterTimezoneType = "AMERICA_YAKUTAT"
)

// CloudClusterProvisionInfo represents the cloud cluster provision info.
type CloudClusterProvisionInfo struct {
	Progress  int          `json:"progress"`
	JobStatus CcpJobStatus `json:"jobStatus"`
	JobType   CcpJobType   `json:"jobType"`
	Vendor    string       `json:"vendor"`
}

type CloudClusterStorageConfig struct {
	LocationName           string `json:"locationName"`
	LocationID             string `json:"locationId"`
	IsImmutable            bool   `json:"isImmutable"`
	IsUsingManagedIdentity bool   `json:"isUsingManagedIdentity"`
}

type CloudClusterCloudInfo struct {
	Name                   string                    `json:"name"`
	Region                 string                    `json:"region"`
	RegionID               string                    `json:"regionId"`
	NetworkName            string                    `json:"networkName"`
	NativeCloudAccountName string                    `json:"nativeCloudAccountName"`
	Vendor                 string                    `json:"vendor"`
	NativeCloudAccountID   string                    `json:"nativeCloudAccountId"`
	CloudAccount           string                    `json:"cloudAccount"`
	StorageConfig          CloudClusterStorageConfig `json:"storageConfig"`
}

type CloudClusterNode struct {
	BrikID          string `json:"brikId"`
	IpAddress       string `json:"ipAddress"`
	NeedsInspection bool   `json:"needsInspection"`
	CpuCores        int    `json:"cpuCores,omitempty"`
	Ram             int64  `json:"ram,omitempty"`
	ClusterID       string `json:"clusterId"`
	NetworkSpeed    string `json:"networkSpeed,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	ID              string `json:"id"`
}

type CloudClusterNodeConnection struct {
	Edges []struct {
		Node CloudClusterNode `json:"node"`
	} `json:"edges"`
}

// CloudCluster represents the cloud cluster.
type CloudCluster struct {
	ID            uuid.UUID                  `json:"id"`
	Name          string                     `json:"name"`
	ProvisionInfo CloudClusterProvisionInfo  `json:"ccprovisionInfo"`
	CloudInfo     CloudClusterCloudInfo      `json:"cloudInfo,omitempty"`
	ClusterNodes  CloudClusterNodeConnection `json:"clusterNodeConnection"`
	ProductType   ClusterProduct             `json:"productType"`
	Timezone      ClusterTimezoneType        `json:"timezone"`
	Version       string                     `json:"version"`
}

// AllCloudClusters returns all cloud clusters.
func (a API) AllCloudClusters(ctx context.Context, first int, after string, filter ClusterFilter, sortBy ClusterSortBy, sortOrder core.SortOrder) ([]CloudCluster, error) {
	a.log.Print(log.Trace)

	query := allClustersConnectionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		First     int            `json:"first"`
		After     string         `json:"after,omitempty"`
		Filter    ClusterFilter  `json:"filter"`
		SortBy    ClusterSortBy  `json:"sortBy"`
		SortOrder core.SortOrder `json:"sortOrder"`
	}{First: first, After: after, Filter: filter, SortBy: sortBy, SortOrder: sortOrder})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Edges []struct {
					Node CloudCluster `json:"node"`
				} `json:"edges"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var clusters []CloudCluster
	for _, edge := range payload.Data.Result.Edges {
		clusters = append(clusters, edge.Node)
	}

	return clusters, nil
}

// CloudClusterInstanceProperties represents the cloud cluster instance properties.
type CloudClusterInstanceProperties struct {
	InstanceType       string `json:"instanceType"`
	Vendor             string `json:"vendor"`
	VcpuCount          int    `json:"vcpuCount"`
	MemoryGib          int    `json:"memoryGib"`
	CapacityTb         int    `json:"capacityTb"`
	ProcessorType      string `json:"processorType"`
	VmType             string `json:"vmType"`
	InstanceTypeString string `json:"instanceTypeString"`
}

// CloudClusterInstancePropertiesRequest represents the request for cloud cluster instance properties.
type CloudClusterInstancePropertiesRequest struct {
	CloudVendor  string `json:"cloudVendor"`
	InstanceType string `json:"instanceType"`
}

// CloudClusterInstanceProperties returns the cloud cluster instance properties.
func (a API) CloudClusterInstanceProperties(ctx context.Context, input CloudClusterInstancePropertiesRequest) (CloudClusterInstanceProperties, error) {
	a.log.Print(log.Trace)

	query := cloudClusterInstancePropertiesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CloudClusterInstancePropertiesRequest `json:"input"`
	}{Input: input})

	if err != nil {
		return CloudClusterInstanceProperties{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				InstanceProperties CloudClusterInstanceProperties `json:"instanceProperties"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudClusterInstanceProperties{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.InstanceProperties, nil
}

// CloudClusterDnsServers represents the cloud cluster DNS servers.
type CloudClusterDnsServers struct {
	Servers []string `json:"servers"`
	Domains []string `json:"domains"`
}

// CloudClusterDnsServers returns the cloud cluster DNS servers.
func (a API) CloudClusterDnsServers(ctx context.Context, clusterID uuid.UUID) (CloudClusterDnsServers, error) {
	a.log.Print(log.Trace)

	query := clusterDnsServersQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID uuid.UUID `json:"clusterUuid"`
	}{ClusterID: clusterID})

	if err != nil {
		return CloudClusterDnsServers{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result CloudClusterDnsServers `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudClusterDnsServers{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

type NTPSymmetricKey struct {
	Key     string `json:"key"`
	KeyId   string `json:"keyId"`
	KeyType string `json:"keyType"`
}

// CloudClusterNtpServers represents the cloud cluster NTP servers.
type CloudClusterNtpServers struct {
	Server       string          `json:"server"`
	SymmetricKey NTPSymmetricKey `json:"symmetricKey,omitempty"`
}

type GetClusterNtpServersInput struct {
	ClusterID uuid.UUID `json:"clusterId"`
}

// CloudClusterNtpServers returns the cloud cluster NTP servers.
func (a API) CloudClusterNtpServers(ctx context.Context, clusterID uuid.UUID) ([]CloudClusterNtpServers, error) {
	a.log.Print(log.Trace)

	query := clusterNtpServersQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID string `json:"id"`
	}{ClusterID: clusterID.String()})

	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Data []CloudClusterNtpServers `json:"data"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Data, nil
}

type IpmiInfo struct {
	IsAvailable bool `json:"isAvailable"`
	UsesHttps   bool `json:"usesHttps"`
	UsesIkvm    bool `json:"usesIkvm"`
}

// CloudClusterSettings represents the cloud cluster settings.
type CloudClusterSettings struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Status      ClusterStatus `json:"status"`
	GeoLocation string        `json:"geoLocation"`
	Timezone    string        `json:"timezone"`
	IpmiInfo    IpmiInfo      `json:"ipmiInfo,omitempty"`
}

// CloudClusterSettings returns the cloud cluster settings.
func (a API) CloudClusterSettings(ctx context.Context, clusterID uuid.UUID) (CloudClusterSettings, error) {
	a.log.Print(log.Trace)

	query := clusterSettingsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID uuid.UUID `json:"id"`
	}{ClusterID: clusterID})

	if err != nil {
		return CloudClusterSettings{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result CloudClusterSettings `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudClusterSettings{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

type UpdateClusterDnsServersAndSearchDomainsInput struct {
	ClusterID      uuid.UUID `json:"clusterId"`
	DnsServers     []string  `json:"dnsServers"`
	SearchDomains  []string  `json:"searchDomains"`
	IsUsingDefault bool      `json:"isUsingDefault"`
}

// UpdateClusterDnsServersAndSearchDomains updates the cloud cluster DNS servers and search domains.
func (a API) UpdateClusterDnsServersAndSearchDomains(ctx context.Context, input UpdateClusterDnsServersAndSearchDomainsInput) error {
	a.log.Print(log.Trace)

	query := updateCusterDnsAndSearchDomainsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input UpdateClusterDnsServersAndSearchDomainsInput `json:"input"`
	}{Input: input})

	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, errors.New("failed to update DNS servers and search domains"))
	}

	return nil
}
