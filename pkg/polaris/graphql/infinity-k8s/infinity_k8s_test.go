package infinityk8s_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	infinityk8s "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/infinity-k8s"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// When enabling integration tests, uncomment the below section.
		// // Load configuration and create client. Usually resolved using the
		// // environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		// polAccount, err := polaris.DefaultServiceAccount(true)
		// if err != nil {
		// 	fmt.Printf("failed to get default service account: %v\n", err)
		// 	os.Exit(1)
		// }

		// For local testing, get the service account credentials from RSC.
		polAccount := &polaris.ServiceAccount{
			Name:           "sdk-test",
			ClientID:       "client|cd908574-7eb6-44eb-a390-302b3604d1be",
			ClientSecret:   "ArqMlH7mN50BVs5FWsVWXx7YbJtO7r1cVEZItC-lItsl5tPk30ygy-5g65IElGVE",
			AccessTokenURI: "https://gqi.dev-142.my.rubrik-lab.com/api/client_token",
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}
		var err error

		client, err = polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}

		version, err := client.GQL.DeploymentVersion(context.Background())
		if err != nil {
			fmt.Printf("failed to get deployment version: %v\n", err)
			os.Exit(1)
		}
		logger.Printf(log.Info, "Polaris version: %s", version)
	}

	os.Exit(m.Run())
}

// TestAddGetDelK8sResourceSet verifies that the SDK can perform the add, get,
// and delete K8s resource set operation on a real RSC instance.
//
// To run this test against an RSC instance, a valid k8s cluster fid should be
// used.
func TestAddGetDelK8sResourceSet(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)

	// TODO: replace k8s cluster fid with real fid.
	k8sClusterFID := uuid.New().String()
	config := infinityk8s.AddK8sResourceSetConfig{
		KubernetesClusterUuid: k8sClusterFID,
		// KubernetesClusterUuid: "bc21d877-df10-5a58-928f-c4a8cae46e62",
		KubernetesNamespace: "fake-ns",
		Definition:          "{}",
		Name:                "fake-ns",
		RSType:              "namespace",
	}

	// 1. Add resourceset.
	addResp, err := infinityK8sClient.AddK8sResourceSet(ctx, config)
	if err != nil {
		t.Error(err)
	}

	if addResp.Id == "" {
		t.Errorf("add failed, %+v", addResp)
	}

	logger := infinityK8sClient.GQL.Log()
	logger.Printf(log.Info, "add succeeded, %+v", addResp)

	// 2. Get resourceset.
	// If we get to this point, addResp.Id should be a valid uuid.
	fid := uuid.Must(uuid.Parse(addResp.Id))
	getResp, err := infinityK8sClient.GetK8sResourceSet(ctx, fid)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get succeeded, %+v", getResp)

	// 3. Delete resourceset.
	delResp, err := infinityK8sClient.DeleteK8sResourceSet(ctx, addResp.Id, false)
	if err != nil {
		t.Error(err)
	}
	if delResp != true {
		t.Errorf("delete failed, %v", delResp)
	}
	logger.Printf(log.Info, "del succeeded, %+v", delResp)
}

// TestGetJobInstance verifies that the SDK can perfrom the get job instance operation
// on a real RSC instance
//
// To run this test against an RSC instance, a valid CDM cluster UUID and a CDM job ID
// TODO: after adding the other graphql endpoints, modify this test to do the following
// - start an ondemand job
// - get the job instance details for the newly created job
func TestGetJobInstance(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()

	//TODO: replace with valid jobId and CDM id.
	validJobId := "CREATE_K8S_SNAPSHOT_e3325e10-bdaf-473d-abfb-70984fbe6d01_c014fc7c-ca27-47c4-a4f0-35ad563dc466:::0"
	validCDMId := "d5ead8ab-1129-4e23-87db-70b2317f534a"
	resp, err := infinityK8sClient.GetJobInstance(ctx, validJobId, validCDMId)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "response: %+v", resp)
}
