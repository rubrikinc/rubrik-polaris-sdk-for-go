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
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// For local testing, get the service account credentials from RSC and uncommnet
		// the below section.
		//
		// polAccount = &polaris.ServiceAccount{
		// 	Name:           "<service_account_name>",
		// 	ClientID:       "<client_id>",
		// 	ClientSecret:   "<client_secret>",
		// 	AccessTokenURI: "<access_token_uri>",
		// }
		polAccount = &polaris.ServiceAccount{
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

// TestDeleteK8sResourceSet verifies that the SDK can perform the delete K8s
// resource set operation on a real RSC instance.
//
// To run this test against an RSC instance, a valid resource set fid should be
// used in the test.
// TODO: after adding other graphql endpoints, modify this test do the following
// - create a new resource set
// - get the resource set info and verify the response
// - delete the resource set and verify success
func TestDeleteK8sResourceSet(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)

	// TODO: replace validFid with real fid value.
	validFid := uuid.New().String()
	resp, err := infinityK8sClient.DeleteK8sResourceSet(ctx, validFid, false)
	if err != nil {
		t.Error(err)
	}
	if resp != true {
		t.Errorf("delete failed, %v", resp)
	}
}
