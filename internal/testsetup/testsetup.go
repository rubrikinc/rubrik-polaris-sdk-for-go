package testsetup

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// testAwsAccount hold AWS account information used in the integration tests.
// Normally used to assert that the information read from Polaris is correct.
type testAwsAccount struct {
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName"`

	Exocompute struct {
		VPCID   string `json:"vpcId"`
		Subnets []struct {
			ID               string `json:"id"`
			AvailabilityZone string `json:"availabilityZone"`
		} `json:"subnets"`
	} `json:"exocompute"`
}

// Load test account information from the file pointed to by the
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

	if n := len(testAccount.Exocompute.Subnets); n != 2 {
		return testAwsAccount{}, fmt.Errorf("file contains the wrong number of subnets: %d", n)
	}

	return testAccount, nil
}

// testAzureSubscription hold Azure subscription information used in the
// integration tests. Normally used to assert that the information read from
// Polaris is correct.
type testAzureSubscription struct {
	SubscriptionID   uuid.UUID `json:"subscriptionId"`
	SubscriptionName string    `json:"subscriptionName"`
	TenantDomain     string    `json:"tenantDomain"`

	Exocompute struct {
		SubnetID string `json:"subnetId"`
	} `json:"exocompute"`
}

// Load test project information from the file pointed to by the
// TEST_AZURESUBSCRIPTION_FILE environment variable.
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
// Normally used to assert that the information read from Polaris is correct.
type testGcpProject struct {
	ProjectName      string `json:"projectName"`
	ProjectID        string `json:"projectId"`
	ProjectNumber    int64  `json:"projectNumber"`
	OrganizationName string `json:"organizationName"`
}

// Load test project information from the file pointed to by the
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
