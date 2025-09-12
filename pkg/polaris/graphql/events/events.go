//go:generate go run ../queries_gen.go events

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

// Package core provides a low-level interface to core GraphQL queries provided
// by the Polaris platform. E.g., task chains and enum definitions.
package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Azure API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ActivityObjectType represents the type of an object.
type ActivityObjectType string

const (
	ObjectTypeOpenstackEnvironment            ActivityObjectType = "OPENSTACK_ENVIRONMENT"
	ObjectTypeActiveDirectoryDomain           ActivityObjectType = "ACTIVE_DIRECTORY_DOMAIN"
	ObjectTypeK8sCluster                      ActivityObjectType = "K8S_CLUSTER"
	ObjectTypeOrionThreatHunt                 ActivityObjectType = "ORION_THREAT_HUNT"
	ObjectTypeM365BackupStorageGroup          ActivityObjectType = "M365_BACKUP_STORAGE_GROUP"
	ObjectTypeKuprNamespace                   ActivityObjectType = "KuprNamespace"
	ObjectTypeSalesforceObject                ActivityObjectType = "SALESFORCE_OBJECT"
	ObjectTypeOlvmHost                        ActivityObjectType = "OLVM_HOST"
	ObjectTypeOracleRac                       ActivityObjectType = "OracleRac"
	ObjectTypeMysqldbInstance                 ActivityObjectType = "MYSQLDB_INSTANCE"
	ObjectTypeAzureSqlManagedInstanceDatabase ActivityObjectType = "AzureSqlManagedInstanceDatabase"
	ObjectTypeHypervVm                        ActivityObjectType = "HypervVm"
	ObjectTypeDataLocation                    ActivityObjectType = "DataLocation"
	ObjectTypeK8sLabel                        ActivityObjectType = "K8S_LABEL"
	ObjectTypeIdpEntraId                      ActivityObjectType = "IDP_ENTRA_ID"
	ObjectTypeCloudDirectNasNamespace         ActivityObjectType = "CLOUD_DIRECT_NAS_NAMESPACE"
	ObjectTypeCloudDirectNasExport            ActivityObjectType = "CLOUD_DIRECT_NAS_EXPORT"
	ObjectTypeHypervScvmm                     ActivityObjectType = "HypervScvmm"
	ObjectTypeO365Onedrive                    ActivityObjectType = "O365Onedrive"
	ObjectTypeHdfs                            ActivityObjectType = "Hdfs"
	ObjectTypeCassandraSource                 ActivityObjectType = "CASSANDRA_SOURCE"
	ObjectTypeAppFlows                        ActivityObjectType = "AppFlows"
	ObjectTypeAwsNativeS3Bucket               ActivityObjectType = "AWS_NATIVE_S3_BUCKET"
	ObjectTypeIdpAws                          ActivityObjectType = "IDP_AWS"
	ObjectTypeSapHanaSystem                   ActivityObjectType = "SapHanaSystem"
	ObjectTypeKuprCluster                     ActivityObjectType = "KuprCluster"
	ObjectTypeGcpNativeProject                ActivityObjectType = "GcpNativeProject"
	ObjectTypeAwsNativeRdsInstance            ActivityObjectType = "AwsNativeRdsInstance"
	ObjectTypeO365Organization                ActivityObjectType = "O365Organization"
	ObjectTypeK8sVirtualMachine               ActivityObjectType = "K8S_VIRTUAL_MACHINE"
	ObjectTypeGoogleWorkspaceOrgUnit          ActivityObjectType = "GOOGLE_WORKSPACE_ORG_UNIT"
	ObjectTypeCloudAccount                    ActivityObjectType = "CLOUD_ACCOUNT"
	ObjectTypeVolumeGroup                     ActivityObjectType = "VolumeGroup"
	ObjectTypeMongoCollection                 ActivityObjectType = "MONGO_COLLECTION"
	ObjectTypeRubrikEc2Instance               ActivityObjectType = "RubrikEc2Instance"
	ObjectTypeM365BackupStorageOrg            ActivityObjectType = "M365_BACKUP_STORAGE_ORG"
	ObjectTypeM365BackupStorageMailbox        ActivityObjectType = "M365_BACKUP_STORAGE_MAILBOX"
	ObjectTypeO365Group                       ActivityObjectType = "O365Group"
	ObjectTypeObjectProtection                ActivityObjectType = "ObjectProtection"
	ObjectTypePrincipalServiceAccount         ActivityObjectType = "PRINCIPAL_SERVICE_ACCOUNT"
	ObjectTypeNutanixVm                       ActivityObjectType = "NutanixVm"
	ObjectTypeIdpOnPremAd                     ActivityObjectType = "IDP_ON_PREM_AD"
	ObjectTypeGoogleWorkspaceOrganization     ActivityObjectType = "GOOGLE_WORKSPACE_ORGANIZATION"
	ObjectTypeReplicationPair                 ActivityObjectType = "REPLICATION_PAIR"
	ObjectTypeIntelFeed                       ActivityObjectType = "INTEL_FEED"
	ObjectTypeAtlassianSite                   ActivityObjectType = "ATLASSIAN_SITE"
	ObjectTypeCloudDirectNasShare             ActivityObjectType = "CLOUD_DIRECT_NAS_SHARE"
	ObjectTypeMongodbSource                   ActivityObjectType = "MONGODB_SOURCE"
	ObjectTypePrincipalGroup                  ActivityObjectType = "PRINCIPAL_GROUP"
	ObjectTypePostgresDbCluster               ActivityObjectType = "POSTGRES_DB_CLUSTER"
	ObjectTypeHypervServer                    ActivityObjectType = "HypervServer"
	ObjectTypeO365SharePointList              ActivityObjectType = "O365SharePointList"
	ObjectTypeVcenter                         ActivityObjectType = "Vcenter"
	ObjectTypeAzureNativeDisk                 ActivityObjectType = "AzureNativeDisk"
	ObjectTypePublicCloudMachineInstance      ActivityObjectType = "PublicCloudMachineInstance"
	ObjectTypePrincipalGpo                    ActivityObjectType = "PRINCIPAL_GPO"
	ObjectTypeO365SharePointDrive             ActivityObjectType = "O365SharePointDrive"
	ObjectTypeShareFileset                    ActivityObjectType = "ShareFileset"
	ObjectTypeMongodbCollection               ActivityObjectType = "MONGODB_COLLECTION"
	ObjectTypeAwsNativeEc2Instance            ActivityObjectType = "AwsNativeEc2Instance"
	ObjectTypeLdap                            ActivityObjectType = "Ldap"
	ObjectTypeWebhook                         ActivityObjectType = "WEBHOOK"
	ObjectTypeAwsNativeDynamodbTable          ActivityObjectType = "AWS_NATIVE_DYNAMODB_TABLE"
	ObjectTypeUpgrade                         ActivityObjectType = "Upgrade"
	ObjectTypeLinuxHost                       ActivityObjectType = "LinuxHost"
	ObjectTypeO365Mailbox                     ActivityObjectType = "O365Mailbox"
	ObjectTypeD365Organization                ActivityObjectType = "D365_ORGANIZATION"
	ObjectTypeAwsNativeAccount                ActivityObjectType = "AwsNativeAccount"
	ObjectTypeAzureStorageAccount             ActivityObjectType = "AZURE_STORAGE_ACCOUNT"
	ObjectTypeVmwareHost                      ActivityObjectType = "VMWARE_HOST"
	ObjectTypePolarisAccount                  ActivityObjectType = "PolarisAccount"
	ObjectTypeAzureSqlManagedInstance         ActivityObjectType = "AzureSqlManagedInstance"
	ObjectTypeAzureNativeVm                   ActivityObjectType = "AzureNativeVm"
	ObjectTypeAzureDevopsProject              ActivityObjectType = "AZURE_DEVOPS_PROJECT"
	ObjectTypeD365Metadata                    ActivityObjectType = "D365_METADATA"
	ObjectTypeOrganization                    ActivityObjectType = "ORGANIZATION"
	ObjectTypeO365Team                        ActivityObjectType = "O365Team"
	ObjectTypeOlvmManager                     ActivityObjectType = "OLVM_MANAGER"
	ObjectTypeCloudNativeVm                   ActivityObjectType = "CloudNativeVm"
	ObjectTypeWindowsFileset                  ActivityObjectType = "WindowsFileset"
	ObjectTypeStorageArray                    ActivityObjectType = "StorageArray"
	ObjectTypeDb2Database                     ActivityObjectType = "Db2Database"
	ObjectTypeAuthDomain                      ActivityObjectType = "AuthDomain"
	ObjectTypeCertificateManagement           ActivityObjectType = "CERTIFICATE_MANAGEMENT"
	ObjectTypeMssql                           ActivityObjectType = "Mssql"
	ObjectTypeNasFileset                      ActivityObjectType = "NAS_FILESET"
	ObjectTypeActiveDirectoryForest           ActivityObjectType = "ACTIVE_DIRECTORY_FOREST"
	ObjectTypePrincipalAssumableIdentity      ActivityObjectType = "PRINCIPAL_ASSUMABLE_IDENTITY"
	ObjectTypeCloudNativeVirtualMachine       ActivityObjectType = "CloudNativeVirtualMachine"
	ObjectTypeCloudDirectNasSystem            ActivityObjectType = "CLOUD_DIRECT_NAS_SYSTEM"
	ObjectTypeAzureAdDirectory                ActivityObjectType = "AZURE_AD_DIRECTORY"
	ObjectTypeVmwareVm                        ActivityObjectType = "VmwareVm"
	ObjectTypeJobInstance                     ActivityObjectType = "JobInstance"
	ObjectTypeSlaDomain                       ActivityObjectType = "SlaDomain"
	ObjectTypeAwsEventType                    ActivityObjectType = "AwsEventType"
	ObjectTypeMongodbDatabase                 ActivityObjectType = "MONGODB_DATABASE"
	ObjectTypeNasSystem                       ActivityObjectType = "NasSystem"
	ObjectTypeExocompute                      ActivityObjectType = "Exocompute"
	ObjectTypeGoogleWorkspaceUser             ActivityObjectType = "GOOGLE_WORKSPACE_USER"
	ObjectTypeAzureDevopsOrganization         ActivityObjectType = "AZURE_DEVOPS_ORGANIZATION"
	ObjectTypeSmbDomain                       ActivityObjectType = "SmbDomain"
	ObjectTypeNutanixPrismCentral             ActivityObjectType = "NUTANIX_PRISM_CENTRAL"
	ObjectTypeD365DataverseTable              ActivityObjectType = "D365_DATAVERSE_TABLE"
	ObjectTypeO365Calendar                    ActivityObjectType = "O365Calendar"
	ObjectTypeK8sProtectionSet                ActivityObjectType = "K8S_PROTECTION_SET"
	ObjectTypePrincipalExternalAccount        ActivityObjectType = "PRINCIPAL_EXTERNAL_ACCOUNT"
	ObjectTypeMongoSource                     ActivityObjectType = "MONGO_SOURCE"
	ObjectTypeSalesforceMetadata              ActivityObjectType = "SALESFORCE_METADATA"
	ObjectTypePrincipalExternalPrincipal      ActivityObjectType = "PRINCIPAL_EXTERNAL_PRINCIPAL"
	ObjectTypeAwsNativeEbsVolume              ActivityObjectType = "AwsNativeEbsVolume"
	ObjectTypeActiveDirectoryDomainController ActivityObjectType = "ACTIVE_DIRECTORY_DOMAIN_CONTROLLER"
	ObjectTypeCassandraColumnFamily           ActivityObjectType = "CASSANDRA_COLUMN_FAMILY"
	ObjectTypeGcpCloudSqlInstance             ActivityObjectType = "GCP_CLOUD_SQL_INSTANCE"
	ObjectTypeOauthToken                      ActivityObjectType = "OAUTH_TOKEN"
	ObjectTypeEncryptionManagement            ActivityObjectType = "ENCRYPTION_MANAGEMENT"
	ObjectTypeManagedVolume                   ActivityObjectType = "ManagedVolume"
	ObjectTypeVcd                             ActivityObjectType = "Vcd"
	ObjectTypePrincipalPublic                 ActivityObjectType = "PRINCIPAL_PUBLIC"
	ObjectTypeAzureSqlDatabase                ActivityObjectType = "AzureSqlDatabase"
	ObjectTypeDb2Instance                     ActivityObjectType = "Db2Instance"
	ObjectTypeOlvmVirtualMachine              ActivityObjectType = "OLVM_VIRTUAL_MACHINE"
	ObjectTypeLinuxFileset                    ActivityObjectType = "LinuxFileset"
	ObjectTypeGoogleWorkspaceSharedDrive      ActivityObjectType = "GOOGLE_WORKSPACE_SHARED_DRIVE"
	ObjectTypeO365Site                        ActivityObjectType = "O365Site"
	ObjectTypeNasHost                         ActivityObjectType = "NasHost"
	ObjectTypeCrossAccountPair                ActivityObjectType = "CROSS_ACCOUNT_PAIR"
	ObjectTypeUser                            ActivityObjectType = "User"
	ObjectTypeK8sNamespaceV2                  ActivityObjectType = "K8S_NAMESPACE_V2"
	ObjectTypeSalesforceOrganization          ActivityObjectType = "SALESFORCE_ORGANIZATION"
	ObjectTypeRubrikEbsVolume                 ActivityObjectType = "RubrikEbsVolume"
	ObjectTypeVcdVapp                         ActivityObjectType = "VcdVapp"
	ObjectTypeOlvmDatacenter                  ActivityObjectType = "OLVM_DATACENTER"
	ObjectTypeWindowsHost                     ActivityObjectType = "WindowsHost"
	ObjectTypeStorm                           ActivityObjectType = "Storm"
	ObjectTypeVmwareComputeCluster            ActivityObjectType = "VmwareComputeCluster"
	ObjectTypeExchangeDatabase                ActivityObjectType = "ExchangeDatabase"
	ObjectTypeM365BackupStorageSite           ActivityObjectType = "M365_BACKUP_STORAGE_SITE"
	ObjectTypeCapacityBundle                  ActivityObjectType = "CapacityBundle"
	ObjectTypeMongoDatabase                   ActivityObjectType = "MONGO_DATABASE"
	ObjectTypeEc2Instance                     ActivityObjectType = "Ec2Instance"
	ObjectTypeKmsKeyVault                     ActivityObjectType = "KMS_KEY_VAULT"
	ObjectTypeConfluenceSpace                 ActivityObjectType = "CONFLUENCE_SPACE"
	ObjectTypeNutanixEra                      ActivityObjectType = "NUTANIX_ERA"
	ObjectTypeJiraProject                     ActivityObjectType = "JIRA_PROJECT"
	ObjectTypeGoogleWorkspaceUserDrive        ActivityObjectType = "GOOGLE_WORKSPACE_USER_DRIVE"
	ObjectTypeStorageArrayVolumeGroup         ActivityObjectType = "StorageArrayVolumeGroup"
	ObjectTypeSnapMirrorCloud                 ActivityObjectType = "SnapMirrorCloud"
	ObjectTypeGcpNativeGceInstance            ActivityObjectType = "GcpNativeGceInstance"
	ObjectTypeOlvmComputeCluster              ActivityObjectType = "OLVM_COMPUTE_CLUSTER"
	ObjectTypePrincipalOrgWide                ActivityObjectType = "PRINCIPAL_ORG_WIDE"
	ObjectTypeEnvoy                           ActivityObjectType = "Envoy"
	ObjectTypeAzureDevopsRepository           ActivityObjectType = "AZURE_DEVOPS_REPOSITORY"
	ObjectTypeCluster                         ActivityObjectType = "Cluster"
	ObjectTypeAppBlueprint                    ActivityObjectType = "AppBlueprint"
	ObjectTypeOpenstackVirtualMachine         ActivityObjectType = "OPENSTACK_VIRTUAL_MACHINE"
	ObjectTypeAzureNativeSubscription         ActivityObjectType = "AzureNativeSubscription"
	ObjectTypeCloudDirectNasBucket            ActivityObjectType = "CLOUD_DIRECT_NAS_BUCKET"
	ObjectTypeIdpSharepoint                   ActivityObjectType = "IDP_SHAREPOINT"
	ObjectTypeCertificate                     ActivityObjectType = "Certificate"
	ObjectTypePrincipalComputer               ActivityObjectType = "PRINCIPAL_COMPUTER"
	ObjectTypeIdpLocalAd                      ActivityObjectType = "IDP_LOCAL_AD"
	ObjectTypeSamlSso                         ActivityObjectType = "SamlSso"
	ObjectTypeOracleDb                        ActivityObjectType = "OracleDb"
	ObjectTypeFailoverClusterApp              ActivityObjectType = "FailoverClusterApp"
	ObjectTypeM365BackupStorageOnedrive       ActivityObjectType = "M365_BACKUP_STORAGE_ONEDRIVE"
	ObjectTypeHost                            ActivityObjectType = "Host"
	ObjectTypeGcpNativeDisk                   ActivityObjectType = "GcpNativeDisk"
	ObjectTypeJiraSettings                    ActivityObjectType = "JIRA_SETTINGS"
	ObjectTypeInformixInstance                ActivityObjectType = "INFORMIX_INSTANCE"
	ObjectTypeSupportBundle                   ActivityObjectType = "SupportBundle"
	ObjectTypeNutanixCluster                  ActivityObjectType = "NutanixCluster"
	ObjectTypeComputeInstance                 ActivityObjectType = "ComputeInstance"
	ObjectTypeCassandraKeyspace               ActivityObjectType = "CASSANDRA_KEYSPACE"
	ObjectTypeSapHanaDb                       ActivityObjectType = "SapHanaDb"
	ObjectTypeAwsAccount                      ActivityObjectType = "AwsAccount"
	ObjectTypeOracle                          ActivityObjectType = "Oracle"
	ObjectTypeOracleHost                      ActivityObjectType = "OracleHost"
	ObjectTypeUnknownObjectType               ActivityObjectType = "UnknownObjectType"
	ObjectTypeAzureSqlDatabaseServer          ActivityObjectType = "AzureSqlDatabaseServer"
	ObjectTypeStorageLocation                 ActivityObjectType = "StorageLocation"
)

// EventObjectType represents the type of an event object.
type EventObjectType string

const (
	EventObjectTypeUnknown                         EventObjectType = "UNKNOWN_EVENT_OBJECT_TYPE"
	EventObjectTypeRubrikSaasAccount               EventObjectType = "RUBRIK_SAAS_ACCOUNT"
	EventObjectTypeAppBlueprint                    EventObjectType = "APP_BLUEPRINT"
	EventObjectTypeAppFlows                        EventObjectType = "APP_FLOWS"
	EventObjectTypeAuthDomain                      EventObjectType = "OBJECT_TYPE_AUTH_DOMAIN"
	EventObjectTypeAwsAccount                      EventObjectType = "AWS_ACCOUNT"
	EventObjectTypeAwsEventType                    EventObjectType = "AWS_EVENT_TYPE"
	EventObjectTypeAzureNativeSubscription         EventObjectType = "AZURE_NATIVE_SUBSCRIPTION"
	EventObjectTypeAzureNativeVm                   EventObjectType = "AZURE_NATIVE_VM"
	EventObjectTypeAzureNativeDisk                 EventObjectType = "AZURE_NATIVE_DISK"
	EventObjectTypeAzureSqlDatabase                EventObjectType = "AZURE_SQL_DATABASE"
	EventObjectTypeAzureSqlManagedInstance         EventObjectType = "AZURE_SQL_MANAGED_INSTANCE"
	EventObjectTypeAzureSqlDatabaseServer          EventObjectType = "AZURE_SQL_DATABASE_SERVER"
	EventObjectTypeAzureSqlManagedInstanceDatabase EventObjectType = "AZURE_SQL_MANAGED_INSTANCE_DATABASE"
	EventObjectTypeCapacityBundle                  EventObjectType = "CAPACITY_BUNDLE"
	EventObjectTypeCloudNativeVirtualMachine       EventObjectType = "OBJECT_TYPE_CLOUD_NATIVE_VIRTUAL_MACHINE"
	EventObjectTypeCloudNativeVm                   EventObjectType = "OBJECT_TYPE_CLOUD_NATIVE_VM"
	EventObjectTypeCertificate                     EventObjectType = "CERTIFICATE"
	EventObjectTypeCluster                         EventObjectType = "CLUSTER"
	EventObjectTypeComputeInstance                 EventObjectType = "COMPUTE_INSTANCE"
	EventObjectTypeDataLocation                    EventObjectType = "DATA_LOCATION"
	EventObjectTypeDb2Database                     EventObjectType = "DB2_DATABASE"
	EventObjectTypeDb2Instance                     EventObjectType = "DB2_INSTANCE"
	EventObjectTypeEc2Instance                     EventObjectType = "EC2_INSTANCE"
	EventObjectTypeEnvoy                           EventObjectType = "ENVOY"
	EventObjectTypeFailoverClusterApp              EventObjectType = "FAILOVER_CLUSTER_APP"
	EventObjectTypeExocompute                      EventObjectType = "EXOCOMPUTE"
	EventObjectTypeExchangeDatabase                EventObjectType = "EXCHANGE_DATABASE"
	EventObjectTypeHdfs                            EventObjectType = "OBJECT_TYPE_HDFS"
	EventObjectTypeHost                            EventObjectType = "HOST"
	EventObjectTypeHypervScvmm                     EventObjectType = "OBJECT_TYPE_HYPERV_SCVMM"
	EventObjectTypeHypervServer                    EventObjectType = "OBJECT_TYPE_HYPERV_SERVER"
	EventObjectTypeHypervVm                        EventObjectType = "HYPERV_VM"
	EventObjectTypeJobInstance                     EventObjectType = "JOB_INSTANCE"
	EventObjectTypeLdap                            EventObjectType = "LDAP"
	EventObjectTypeLinuxFileset                    EventObjectType = "LINUX_FILESET"
	EventObjectTypeLinuxHost                       EventObjectType = "LINUX_HOST"
	EventObjectTypeManagedVolume                   EventObjectType = "MANAGED_VOLUME"
	EventObjectTypeMssql                           EventObjectType = "MSSQL"
	EventObjectTypeNasFileset                      EventObjectType = "NAS_FILESET"
	EventObjectTypeWebhook                         EventObjectType = "WEBHOOK"
	EventObjectTypeNasHost                         EventObjectType = "NAS_HOST"
	EventObjectTypeNasSystem                       EventObjectType = "NAS_SYSTEM"
	EventObjectTypeNutanixCluster                  EventObjectType = "OBJECT_TYPE_NUTANIX_CLUSTER"
	EventObjectTypeNutanixVm                       EventObjectType = "NUTANIX_VM"
	EventObjectTypeO365Calendar                    EventObjectType = "O365_CALENDAR"
	EventObjectTypeO365Mailbox                     EventObjectType = "O365_MAILBOX"
	EventObjectTypeO365Onedrive                    EventObjectType = "O365_ONEDRIVE"
	EventObjectTypeO365Site                        EventObjectType = "O365_SITE"
	EventObjectTypeO365SharePointDrive             EventObjectType = "O365_SHARE_POINT_DRIVE"
	EventObjectTypeO365SharePointList              EventObjectType = "O365_SHARE_POINT_LIST"
	EventObjectTypeO365Team                        EventObjectType = "O365_TEAM"
	EventObjectTypeO365Organization                EventObjectType = "O365_ORGANIZATION"
	EventObjectTypeO365Group                       EventObjectType = "O365_GROUP"
	EventObjectTypeObjectProtection                EventObjectType = "OBJECT_PROTECTION"
	EventObjectTypeOracle                          EventObjectType = "ORACLE"
	EventObjectTypeOracleDb                        EventObjectType = "ORACLE_DB"
	EventObjectTypeOracleHost                      EventObjectType = "ORACLE_HOST"
	EventObjectTypeOracleRac                       EventObjectType = "ORACLE_RAC"
	EventObjectTypeAwsNativeAccount                EventObjectType = "AWS_NATIVE_ACCOUNT"
	EventObjectTypeAwsNativeEbsVolume              EventObjectType = "AWS_NATIVE_EBS_VOLUME"
	EventObjectTypeRubrikSaasEbsVolume             EventObjectType = "RUBRIK_SAAS_EBS_VOLUME"
	EventObjectTypeRubrikSaasEc2Instance           EventObjectType = "RUBRIK_SAAS_EC2_INSTANCE"
	EventObjectTypePublicCloudMachineInstance      EventObjectType = "PUBLIC_CLOUD_MACHINE_INSTANCE"
	EventObjectTypeSamlSso                         EventObjectType = "SAML_SSO"
	EventObjectTypeSapHanaDb                       EventObjectType = "SAP_HANA_DB"
	EventObjectTypeSapHanaSystem                   EventObjectType = "SAP_HANA_SYSTEM"
	EventObjectTypeShareFileset                    EventObjectType = "SHARE_FILESET"
	EventObjectTypeSlaDomain                       EventObjectType = "SLA_DOMAIN"
	EventObjectTypeSmbDomain                       EventObjectType = "SMB_DOMAIN"
	EventObjectTypeSnapMirrorCloud                 EventObjectType = "SNAP_MIRROR_CLOUD"
	EventObjectTypeStorageArray                    EventObjectType = "OBJECT_TYPE_STORAGE_ARRAY"
	EventObjectTypeStorageArrayVolumeGroup         EventObjectType = "STORAGE_ARRAY_VOLUME_GROUP"
	EventObjectTypeStorageLocation                 EventObjectType = "STORAGE_LOCATION"
	EventObjectTypeStorm                           EventObjectType = "STORM"
	EventObjectTypeSupportBundle                   EventObjectType = "SUPPORT_BUNDLE"
	EventObjectTypeUser                            EventObjectType = "USER"
	EventObjectTypeUpgrade                         EventObjectType = "OBJECT_TYPE_UPGRADE"
	EventObjectTypeVcd                             EventObjectType = "OBJECT_TYPE_VCD"
	EventObjectTypeVcdVapp                         EventObjectType = "VCD_VAPP"
	EventObjectTypeVcenter                         EventObjectType = "OBJECT_TYPE_VCENTER"
	EventObjectTypeVmwareComputeCluster            EventObjectType = "VMWARE_COMPUTE_CLUSTER"
	EventObjectTypeVmwareVm                        EventObjectType = "VMWARE_VM"
	EventObjectTypeVolumeGroup                     EventObjectType = "OBJECT_TYPE_VOLUME_GROUP"
	EventObjectTypeWindowsFileset                  EventObjectType = "WINDOWS_FILESET"
	EventObjectTypeWindowsHost                     EventObjectType = "WINDOWS_HOST"
	EventObjectTypeGcpNativeProject                EventObjectType = "GCP_NATIVE_PROJECT"
	EventObjectTypeAwsNativeRdsInstance            EventObjectType = "AWS_NATIVE_RDS_INSTANCE"
	EventObjectTypeGcpNativeGceInstance            EventObjectType = "GCP_NATIVE_GCE_INSTANCE"
	EventObjectTypeGcpNativeDisk                   EventObjectType = "GCP_NATIVE_DISK"
	EventObjectTypeKuprCluster                     EventObjectType = "KUPR_CLUSTER"
	EventObjectTypeKuprNamespace                   EventObjectType = "KUPR_NAMESPACE"
	EventObjectTypeCassandraColumnFamily           EventObjectType = "CASSANDRA_COLUMN_FAMILY"
	EventObjectTypeCassandraKeyspace               EventObjectType = "CASSANDRA_KEYSPACE"
	EventObjectTypeCassandraSource                 EventObjectType = "CASSANDRA_SOURCE"
	EventObjectTypeMongodbCollection               EventObjectType = "MONGODB_COLLECTION"
	EventObjectTypeMongodbDatabase                 EventObjectType = "MONGODB_DATABASE"
	EventObjectTypeMongodbSource                   EventObjectType = "MONGODB_SOURCE"
	EventObjectTypeCloudDirectNasExport            EventObjectType = "CLOUD_DIRECT_NAS_EXPORT"
	EventObjectTypeMongoCollection                 EventObjectType = "MONGO_COLLECTION"
	EventObjectTypeMongoDatabase                   EventObjectType = "MONGO_DATABASE"
	EventObjectTypeMongoSource                     EventObjectType = "MONGO_SOURCE"
	EventObjectTypeCertificateManagement           EventObjectType = "CERTIFICATE_MANAGEMENT"
	EventObjectTypeAwsNativeS3Bucket               EventObjectType = "AWS_NATIVE_S3_BUCKET"
	EventObjectTypeAzureStorageAccount             EventObjectType = "AZURE_STORAGE_ACCOUNT"
	EventObjectTypeK8sCluster                      EventObjectType = "K8S_CLUSTER"
	EventObjectTypeK8sProtectionSet                EventObjectType = "K8S_PROTECTION_SET"
	EventObjectTypeAzureAdDirectory                EventObjectType = "AZURE_AD_DIRECTORY"
	EventObjectTypeEncryptionManagement            EventObjectType = "ENCRYPTION_MANAGEMENT"
	EventObjectTypeActiveDirectoryDomain           EventObjectType = "ACTIVE_DIRECTORY_DOMAIN"
	EventObjectTypeActiveDirectoryDomainController EventObjectType = "ACTIVE_DIRECTORY_DOMAIN_CONTROLLER"
	EventObjectTypeNutanixPrismCentral             EventObjectType = "OBJECT_TYPE_NUTANIX_PRISM_CENTRAL"
	EventObjectTypeVmwareHost                      EventObjectType = "VMWARE_HOST"
	EventObjectTypeAtlassianSite                   EventObjectType = "ATLASSIAN_SITE"
	EventObjectTypeJiraProject                     EventObjectType = "JIRA_PROJECT"
	EventObjectTypeJiraSettings                    EventObjectType = "JIRA_SETTINGS"
	EventObjectTypeReplicationPair                 EventObjectType = "REPLICATION_PAIR"
	EventObjectTypeOauthToken                      EventObjectType = "OAUTH_TOKEN"
	EventObjectTypePostgresDbCluster               EventObjectType = "POSTGRES_DB_CLUSTER"
	EventObjectTypeCrossAccountPair                EventObjectType = "CROSS_ACCOUNT_PAIR"
	EventObjectTypeSalesforceOrganization          EventObjectType = "SALESFORCE_ORGANIZATION"
	EventObjectTypeSalesforceObject                EventObjectType = "SALESFORCE_OBJECT"
	EventObjectTypeSalesforceMetadata              EventObjectType = "SALESFORCE_METADATA"
	EventObjectTypeIntelFeed                       EventObjectType = "INTEL_FEED"
	EventObjectTypeOrganization                    EventObjectType = "ORGANIZATION"
	EventObjectTypeNutanixEra                      EventObjectType = "OBJECT_TYPE_NUTANIX_ERA"
	EventObjectTypeActiveDirectoryForest           EventObjectType = "ACTIVE_DIRECTORY_FOREST"
	EventObjectTypeConfluenceSpace                 EventObjectType = "CONFLUENCE_SPACE"
	EventObjectTypeM365BackupStorageOrg            EventObjectType = "M365_BACKUP_STORAGE_ORG"
	EventObjectTypeM365BackupStorageMailbox        EventObjectType = "M365_BACKUP_STORAGE_MAILBOX"
	EventObjectTypeM365BackupStorageOnedrive       EventObjectType = "M365_BACKUP_STORAGE_ONEDRIVE"
	EventObjectTypeM365BackupStorageSite           EventObjectType = "M365_BACKUP_STORAGE_SITE"
	EventObjectTypeM365BackupStorageGroup          EventObjectType = "M365_BACKUP_STORAGE_GROUP"
	EventObjectTypeKmsKeyVault                     EventObjectType = "KMS_KEY_VAULT"
	EventObjectTypeK8sVirtualMachine               EventObjectType = "K8S_VIRTUAL_MACHINE"
	EventObjectTypeK8sNamespaceV2                  EventObjectType = "K8S_NAMESPACE_V2"
	EventObjectTypeD365Organization                EventObjectType = "D365_ORGANIZATION"
	EventObjectTypeD365DataverseTable              EventObjectType = "D365_DATAVERSE_TABLE"
	EventObjectTypeD365Metadata                    EventObjectType = "D365_METADATA"
	EventObjectTypeMysqldbInstance                 EventObjectType = "MYSQLDB_INSTANCE"
	EventObjectTypeOrionThreatHunt                 EventObjectType = "ORION_THREAT_HUNT"
	EventObjectTypeAwsNativeDynamodbTable          EventObjectType = "AWS_NATIVE_DYNAMODB_TABLE"
	EventObjectTypeOpenstackEnvironment            EventObjectType = "OPENSTACK_ENVIRONMENT"
	EventObjectTypeOpenstackVirtualMachine         EventObjectType = "OPENSTACK_VIRTUAL_MACHINE"
	EventObjectTypeInformixInstance                EventObjectType = "INFORMIX_INSTANCE"
	EventObjectTypeK8sLabel                        EventObjectType = "K8S_LABEL"
	EventObjectTypeGcpCloudSqlInstance             EventObjectType = "GCP_CLOUD_SQL_INSTANCE"
	EventObjectTypePrincipalGroup                  EventObjectType = "PRINCIPAL_GROUP"
	EventObjectTypePrincipalAssumableIdentity      EventObjectType = "PRINCIPAL_ASSUMABLE_IDENTITY"
	EventObjectTypePrincipalExternalAccount        EventObjectType = "PRINCIPAL_EXTERNAL_ACCOUNT"
	EventObjectTypePrincipalServiceAccount         EventObjectType = "PRINCIPAL_SERVICE_ACCOUNT"
	EventObjectTypePrincipalExternalPrincipal      EventObjectType = "PRINCIPAL_EXTERNAL_PRINCIPAL"
	EventObjectTypePrincipalPublic                 EventObjectType = "PRINCIPAL_PUBLIC"
	EventObjectTypePrincipalOrgWide                EventObjectType = "PRINCIPAL_ORG_WIDE"
	EventObjectTypePrincipalGpo                    EventObjectType = "PRINCIPAL_GPO"
	EventObjectTypePrincipalComputer               EventObjectType = "PRINCIPAL_COMPUTER"
	EventObjectTypeIdpOnPremAd                     EventObjectType = "IDP_ON_PREM_AD"
	EventObjectTypeIdpAws                          EventObjectType = "IDP_AWS"
	EventObjectTypeIdpEntraId                      EventObjectType = "IDP_ENTRA_ID"
	EventObjectTypeIdpLocalAd                      EventObjectType = "IDP_LOCAL_AD"
	EventObjectTypeIdpSharepoint                   EventObjectType = "IDP_SHAREPOINT"
	EventObjectTypeGoogleWorkspaceOrganization     EventObjectType = "GOOGLE_WORKSPACE_ORGANIZATION"
	EventObjectTypeGoogleWorkspaceOrgUnit          EventObjectType = "GOOGLE_WORKSPACE_ORG_UNIT"
	EventObjectTypeGoogleWorkspaceUser             EventObjectType = "GOOGLE_WORKSPACE_USER"
	EventObjectTypeGoogleWorkspaceUserDrive        EventObjectType = "GOOGLE_WORKSPACE_USER_DRIVE"
	EventObjectTypeGoogleWorkspaceSharedDrive      EventObjectType = "GOOGLE_WORKSPACE_SHARED_DRIVE"
	EventObjectTypeAzureDevopsOrganization         EventObjectType = "AZURE_DEVOPS_ORGANIZATION"
	EventObjectTypeAzureDevopsProject              EventObjectType = "AZURE_DEVOPS_PROJECT"
	EventObjectTypeAzureDevopsRepository           EventObjectType = "AZURE_DEVOPS_REPOSITORY"
	EventObjectTypeCloudDirectNasShare             EventObjectType = "CLOUD_DIRECT_NAS_SHARE"
	EventObjectTypeCloudDirectNasSystem            EventObjectType = "CLOUD_DIRECT_NAS_SYSTEM"
	EventObjectTypeCloudAccount                    EventObjectType = "CLOUD_ACCOUNT"
	EventObjectTypeCloudDirectNasNamespace         EventObjectType = "CLOUD_DIRECT_NAS_NAMESPACE"
	EventObjectTypeCloudDirectNasBucket            EventObjectType = "CLOUD_DIRECT_NAS_BUCKET"
	EventObjectTypeOlvmManager                     EventObjectType = "OLVM_MANAGER"
	EventObjectTypeOlvmDatacenter                  EventObjectType = "OLVM_DATACENTER"
	EventObjectTypeOlvmComputeCluster              EventObjectType = "OLVM_COMPUTE_CLUSTER"
	EventObjectTypeOlvmHost                        EventObjectType = "OLVM_HOST"
	EventObjectTypeOlvmVirtualMachine              EventObjectType = "OLVM_VIRTUAL_MACHINE"
)

// EventSeriesSortField is used to sort the event series.
type EventSeriesSortField string

const (
	EventSeriesSortFieldLastUpdated EventSeriesSortField = "LAST_UPDATED"
)

// ActivityStatus represents the status of an activity.
type ActivityStatus string

const (
	ActivityStatusTaskFailure    ActivityStatus = "TaskFailure"
	ActivityStatusWarning        ActivityStatus = "Warning"
	ActivityStatusQueued         ActivityStatus = "Queued"
	ActivityStatusTaskSuccess    ActivityStatus = "TaskSuccess"
	ActivityStatusFailure        ActivityStatus = "Failure"
	ActivityStatusInfo           ActivityStatus = "Info"
	ActivityStatusCanceled       ActivityStatus = "Canceled"
	ActivityStatusPartialSuccess ActivityStatus = "PARTIAL_SUCCESS"
	ActivityStatusCanceling      ActivityStatus = "Canceling"
	ActivityStatusRunning        ActivityStatus = "Running"
	ActivityStatusSuccess        ActivityStatus = "Success"
)

// ActivityType represents the type of an activity.
type ActivityType string

const (
	ActivityTypeStorage                       ActivityType = "Storage"
	ActivityTypeThreatMonitoring              ActivityType = "THREAT_MONITORING"
	ActivityTypePermissionAssessment          ActivityType = "PERMISSION_ASSESSMENT"
	ActivityTypeTpr                           ActivityType = "Tpr"
	ActivityTypeClassification                ActivityType = "Classification"
	ActivityTypeLegalHold                     ActivityType = "LegalHold"
	ActivityTypeHypervScvmm                   ActivityType = "HypervScvmm"
	ActivityTypeThreatFeed                    ActivityType = "THREAT_FEED"
	ActivityTypeHdfs                          ActivityType = "Hdfs"
	ActivityTypeScheduleRecovery              ActivityType = "SCHEDULE_RECOVERY"
	ActivityTypeRadarAnalysis                 ActivityType = "RadarAnalysis"
	ActivityTypeVolumeGroup                   ActivityType = "VolumeGroup"
	ActivityTypeLockSnapshot                  ActivityType = "LockSnapshot"
	ActivityTypeInstantiate                   ActivityType = "Instantiate"
	ActivityTypeIdentityViolation             ActivityType = "IDENTITY_VIOLATION"
	ActivityTypeBulkRecovery                  ActivityType = "BULK_RECOVERY"
	ActivityTypeLogBackup                     ActivityType = "LOG_BACKUP"
	ActivityTypeHypervServer                  ActivityType = "HypervServer"
	ActivityTypeIsolatedRecovery              ActivityType = "ISOLATED_RECOVERY"
	ActivityTypeConfiguration                 ActivityType = "Configuration"
	ActivityTypeUpgrade                       ActivityType = "Upgrade"
	ActivityTypeEncryptionManagementOperation ActivityType = "ENCRYPTION_MANAGEMENT_OPERATION"
	ActivityTypeCloudNativeVm                 ActivityType = "CloudNativeVm"
	ActivityTypeStorageArray                  ActivityType = "StorageArray"
	ActivityTypeConnection                    ActivityType = "Connection"
	ActivityTypeConversion                    ActivityType = "Conversion"
	ActivityTypeAuthDomain                    ActivityType = "AuthDomain"
	ActivityTypeUnknownEventType              ActivityType = "UnknownEventType"
	ActivityTypeQuarantine                    ActivityType = "QUARANTINE"
	ActivityTypeCloudNativeVirtualMachine     ActivityType = "CloudNativeVirtualMachine"
	ActivityTypeDiscovery                     ActivityType = "Discovery"
	ActivityTypeMaintenance                   ActivityType = "Maintenance"
	ActivityTypeSupport                       ActivityType = "Support"
	ActivityTypeReplication                   ActivityType = "Replication"
	ActivityTypeSecurityViolation             ActivityType = "SECURITY_VIOLATION"
	ActivityTypeFileset                       ActivityType = "Fileset"
	ActivityTypeLocalRecovery                 ActivityType = "LocalRecovery"
	ActivityTypeSystem                        ActivityType = "System"
	ActivityTypeFailover                      ActivityType = "Failover"
	ActivityTypeOwnership                     ActivityType = "OWNERSHIP"
	ActivityTypeStormResource                 ActivityType = "StormResource"
	ActivityTypeDiagnostic                    ActivityType = "Diagnostic"
	ActivityTypeVcd                           ActivityType = "Vcd"
	ActivityTypeAnomaly                       ActivityType = "Anomaly"
	ActivityTypeSeeding                       ActivityType = "SEEDING"
	ActivityTypeArchive                       ActivityType = "Archive"
	ActivityTypeCloudNativeSource             ActivityType = "CloudNativeSource"
	ActivityTypeHostEvent                     ActivityType = "HostEvent"
	ActivityTypeAwsEvent                      ActivityType = "AwsEvent"
	ActivityTypeResourceOperations            ActivityType = "ResourceOperations"
	ActivityTypeIdentityAlerts                ActivityType = "IDENTITY_ALERTS"
	ActivityTypeBackup                        ActivityType = "Backup"
	ActivityTypeSync                          ActivityType = "Sync"
	ActivityTypeTenantQuota                   ActivityType = "TENANT_QUOTA"
	ActivityTypeHardware                      ActivityType = "Hardware"
	ActivityTypeTestFailover                  ActivityType = "TestFailover"
	ActivityTypeRecovery                      ActivityType = "Recovery"
	ActivityTypeDownload                      ActivityType = "Download"
	ActivityTypeEmbeddedEvent                 ActivityType = "EmbeddedEvent"
	ActivityTypeProtectedObjectDeletion       ActivityType = "PROTECTED_OBJECT_DELETION"
	ActivityTypeTenantOverlap                 ActivityType = "TENANT_OVERLAP"
	ActivityTypeNutanixCluster                ActivityType = "NutanixCluster"
	ActivityTypeVCenter                       ActivityType = "VCenter"
	ActivityTypeIndex                         ActivityType = "Index"
	ActivityTypeThreatHunt                    ActivityType = "ThreatHunt"
	ActivityTypeIdentityIntelligence          ActivityType = "USER_INTELLIGENCE"
)

// ClusterType represents the type of a cluster.
type ClusterType string

const (
	ClusterTypeUnknown ClusterType = "UNKNOWN_CLUSTER_TYPE"
	ClusterTypeOnPrem  ClusterType = "ON_PREM"
	ClusterTypeRobo    ClusterType = "ROBO"
	ClusterTypeCloud   ClusterType = "CLOUD"
	ClusterTypeRubrik  ClusterType = "RUBRIK_SAAS"
	ClusterTypeExo     ClusterType = "EXO_COMPUTE"
)

// EventSeverity represents the severity of an event.
type EventSeverity string

const (
	EventSeverityInfo     EventSeverity = "SEVERITY_INFO"
	EventSeverityCritical EventSeverity = "SEVERITY_CRITICAL"
	EventSeverityWarning  EventSeverity = "SEVERITY_WARNING"
)

// EventSeriesFilter is used to filter the event series.
type EventSeriesFilter struct {
	ClusterID          []uuid.UUID       `json:"clusterId,omitempty"`
	ClusterType        []ClusterType     `json:"clusterType,omitempty"`
	LastActivityStatus []ActivityStatus  `json:"lastActivityStatus,omitempty"`
	LastActivityType   []ActivityType    `json:"lastActivityType,omitempty"`
	LastUpdatedTimeGt  string            `json:"lastUpdatedTimeGt,omitempty"`
	LastUpdatedTimeLt  string            `json:"lastUpdatedTimeLt,omitempty"`
	ObjectFID          []uuid.UUID       `json:"objectFid,omitempty"`
	ObjectName         string            `json:"objectName,omitempty"`
	ObjectType         []EventObjectType `json:"objectType,omitempty"`
	OrgIDs             []string          `json:"orgIds,omitempty"`
	Severity           []EventSeverity   `json:"severity,omitempty"`
	SearchTerm         string            `json:"searchTerm,omitempty"`
	StartTimeGt        string            `json:"startTimeGt,omitempty"`
	StartTimeLt        string            `json:"startTimeLt,omitempty"`
	UserIDs            []string          `json:"userIds,omitempty"`
}

// EventSeries represents an event series.
type EventSeries struct {
	ID                   int                `json:"id"`
	FID                  string             `json:"fid"`
	ActivitySeriesID     string             `json:"activitySeriesId"`
	LastUpdated          time.Time          `json:"lastUpdated"`
	LastActivityType     ActivityType       `json:"lastActivityType"`
	LastActivityStatus   ActivityStatus     `json:"lastActivityStatus"`
	ObjectID             string             `json:"objectId"`
	ObjectName           string             `json:"objectName"`
	ObjectType           ActivityObjectType `json:"objectType"`
	Severity             string             `json:"severity"`
	Progress             string             `json:"progress"`
	IsCancelable         bool               `json:"isCancelable"`
	IsPolarisEventSeries bool               `json:"isPolarisEventSeries"`
	Location             string             `json:"location"`
	EffectiveThroughput  int                `json:"effectiveThroughput"`
	DataTransferred      int                `json:"dataTransferred"`
	LogicalSize          int                `json:"logicalSize"`
	SlaDomainName        string             `json:"slaDomainName"`
	Organizations        []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organizations"`
	ClusterUUID string `json:"clusterUuid"`
	ClusterName string `json:"clusterName"`
	Username    string `json:"username"`
	Cluster     struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	} `json:"cluster"`
	Activities struct {
		Nodes []struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"nodes"`
	} `json:"activityConnection"`
}

// EventSeries returns the event series matching the specified filter.
func (a API) EventSeries(ctx context.Context, after string, filters EventSeriesFilter, first int, sortBy EventSeriesSortField, sortOrder core.SortOrderEnum) ([]EventSeries, error) {
	a.log.Print(log.Trace)

	query := eventSeriesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		After     string               `json:"after,omitempty"`
		Filters   EventSeriesFilter    `json:"filters,omitempty"`
		First     int                  `json:"first"`
		SortBy    EventSeriesSortField `json:"sortBy"`
		SortOrder core.SortOrderEnum   `json:"sortOrder"`
	}{After: after, Filters: filters, First: first, SortBy: sortBy, SortOrder: sortOrder})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Edges []struct {
					Node EventSeries `json:"node"`
				} `json:"edges"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var events []EventSeries
	for _, edge := range payload.Data.Result.Edges {
		events = append(events, edge.Node)
	}

	return events, nil
}

type ActivitySeries struct {
	ID                   int                `json:"id"`
	FID                  string             `json:"fid"`
	ActivitySeriesID     string             `json:"activitySeriesId"`
	LastUpdated          time.Time          `json:"lastUpdated"`
	LastActivityType     ActivityType       `json:"lastActivityType"`
	LastActivityStatus   ActivityStatus     `json:"lastActivityStatus"`
	ObjectID             string             `json:"objectId"`
	ObjectName           string             `json:"objectName"`
	ObjectType           ActivityObjectType `json:"objectType"`
	Severity             string             `json:"severity"`
	Progress             string             `json:"progress"`
	IsCancelable         bool               `json:"isCancelable"`
	IsPolarisEventSeries bool               `json:"isPolarisEventSeries"`
	Location             string             `json:"location"`
	EffectiveThroughput  int                `json:"effectiveThroughput"`
	DataTransferred      int                `json:"dataTransferred"`
	LogicalSize          int                `json:"logicalSize"`
	SlaDomainName        string             `json:"slaDomainName"`
	Organizations        []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organizations"`
	ClusterUUID string `json:"clusterUuid"`
	ClusterName string `json:"clusterName"`
	Username    string `json:"username"`
	Cluster     struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	} `json:"cluster"`
	Activities struct {
		Nodes []struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"nodes"`
	} `json:"activityConnection"`
}

func (a API) ActivitySeries(ctx context.Context, activitySeriesID string, clusterUUID string) (EventSeries, error) {
	a.log.Print(log.Trace)

	query := activitySeriesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ActivitySeriesID string `json:"activitySeriesId"`
		ClusterUUID      string `json:"clusterUuid,omitempty"`
	}{ActivitySeriesID: activitySeriesID, ClusterUUID: clusterUUID})
	if err != nil {
		return EventSeries{}, graphql.RequestError(query, err)
	}
	var payload struct {
		Data struct {
			Result EventSeries `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return EventSeries{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
