package testsetup

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// testAwsAccount hold AWS account information used in the integration tests.
// Normally used to assert that the information read from RSC is correct.
type testAwsAccount struct {
	Profile          string `json:"profile"`
	AccountID        string `json:"accountId"`
	AccountName      string `json:"accountName"`
	CrossAccountID   string `json:"crossAccountId"`
	CrossAccountName string `json:"crossAccountName"`
	CrossAccountRole string `json:"crossAccountRole"`

	Exocompute struct {
		VPCID   string `json:"vpcId"`
		Subnets []struct {
			ID               string `json:"id"`
			AvailabilityZone string `json:"availabilityZone"`
		} `json:"subnets"`
	} `json:"exocompute"`
}

// AWSAccount loads test account information from the file pointed to by the
// TEST_AWSACCOUNT_FILE environment variable.
func AWSAccount() (testAwsAccount, error) {
	buf, err := os.ReadFile(os.Getenv("TEST_AWSACCOUNT_FILE"))
	if err != nil {
		return testAwsAccount{}, fmt.Errorf("failed to read file pointed to by TEST_AWSACCOUNT_FILE: %v", err)
	}

	testAccount := testAwsAccount{}
	if err := json.Unmarshal(buf, &testAccount); err != nil {
		return testAwsAccount{}, err
	}

	if n := len(testAccount.Exocompute.Subnets); n != 3 {
		return testAwsAccount{}, fmt.Errorf("file contains the wrong number of subnets: %d", n)
	}

	return testAccount, nil
}

// testAzureSubscription hold Azure subscription information used in the
// integration tests. Normally used to assert that the information read from
// RSC is correct.
type testAzureSubscription struct {
	SubscriptionID   uuid.UUID `json:"subscriptionId"`
	SubscriptionName string    `json:"subscriptionName"`
	TenantID         uuid.UUID `json:"tenantId"`
	TenantDomain     string    `json:"tenantDomain"`
	PrincipalID      uuid.UUID `json:"principalId"`
	PrincipalName    string    `json:"principalName"`
	PrincipalSecret  string    `json:"principalSecret"`

	// Should be in EastUS2 region for integration test as the region is
	// hardcoded there.
	Archival struct {
		Regions             []string `json:"regions"`
		ManagedIdentityName string   `json:"managedIdentityName"`
		PrincipalID         string   `json:"managedIdentityPrincipalId"`
		ResourceGroupName   string   `json:"resourceGroupName"`
		ResourceGroupRegion string   `json:"resourceGroupRegion"`
	} `json:"archival"`

	CloudNativeProtection struct {
		Regions             []string `json:"regions"`
		ResourceGroupName   string   `json:"resourceGroupName"`
		ResourceGroupRegion string   `json:"resourceGroupRegion"`
	} `json:"cloudNativeProtection"`

	Exocompute struct {
		Regions             []string `json:"regions"`
		ResourceGroupName   string   `json:"resourceGroupName"`
		ResourceGroupRegion string   `json:"resourceGroupRegion"`
		SubnetID            string   `json:"subnetId"`
	} `json:"exocompute"`
}

// AzureSubscription loads test project information from the file pointed to by
// the TEST_AZURESUBSCRIPTION_FILE environment variable.
func AzureSubscription() (testAzureSubscription, error) {
	buf, err := os.ReadFile(os.Getenv("TEST_AZURESUBSCRIPTION_FILE"))
	if err != nil {
		return testAzureSubscription{}, fmt.Errorf("failed to read file pointed to by TEST_AZURESUBSCRIPTION_FILE: %v", err)
	}

	testSubscription := testAzureSubscription{}
	if err := json.Unmarshal(buf, &testSubscription); err != nil {
		return testAzureSubscription{}, err
	}

	return testSubscription, nil
}

// testGcpProject hold GCP project information used in the integration tests.
// Normally used to assert that the information read from RSC is correct.
type testGcpProject struct {
	ProjectName      string `json:"projectName"`
	ProjectID        string `json:"projectId"`
	ProjectNumber    int64  `json:"projectNumber"`
	OrganizationName string `json:"organizationName"`

	Exocompute struct {
		Region         string `json:"region"`
		SubnetName     string `json:"subnetName"`
		VPCNetworkName string `json:"vpcNetworkName"`
	} `json:"exocompute"`
}

// GCPProject loads test project information from the file pointed to by the
// TEST_GCPPROJECT_FILE environment variable.
func GCPProject() (testGcpProject, error) {
	buf, err := os.ReadFile(os.Getenv("TEST_GCPPROJECT_FILE"))
	if err != nil {
		return testGcpProject{}, fmt.Errorf("failed to read file pointed to by TEST_GCPPROJECT_FILE: %v", err)
	}
	testProject := testGcpProject{}
	if err := json.Unmarshal(buf, &testProject); err != nil {
		return testGcpProject{}, err
	}
	return testProject, nil
}

// testRSCConfig hold configuration information used in the integration tests.
// Normally used to assert that the information read from RSC is correct.
type testRSCConfig struct {
	ExistingUserEmail string `json:"existingUserEmail"`
	NewUserEmail      string `json:"newUserEmail"`
}

// RSCConfig loads test configuration information from the file pointed to by
// the TEST_RSCCONFIG_FILE environment variable.
func RSCConfig() (testRSCConfig, error) {
	buf, err := os.ReadFile(os.Getenv("TEST_RSCCONFIG_FILE"))
	if err != nil {
		return testRSCConfig{}, fmt.Errorf("failed to read file pointed to by TEST_RSCCONFIG_FILE: %v", err)
	}
	testConfig := testRSCConfig{}
	if err := json.Unmarshal(buf, &testConfig); err != nil {
		return testRSCConfig{}, err
	}
	return testConfig, nil
}
