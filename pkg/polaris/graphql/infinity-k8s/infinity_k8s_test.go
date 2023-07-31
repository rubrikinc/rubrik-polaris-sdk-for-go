package infinityk8s_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
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

// TestExportK8sResourceSetSnapshot verifies that the SDK can perfrom the get
// on demand export k8s resource set snapshot job operation and then
// get the job instance details for the job on a real RSC instance
//
// To run this test against an RSC instance, the following are required
// - a CDM cluster UUID
// - a target K8s cluster fid on the CDM
// - a k8s resource set snapshot fid on the CDM
func TestExportK8sResourceSetSnapshot(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()

	// TODO: Replace with valid snapshot fid and cdm cluster UUID
	validSnapshotFid := "dc058c1e-5753-57c7-a31b-4b6c48f885dd"
	validCDMId := "d5ead8ab-1129-4e23-87db-70b2317f534a"
	validTargetK8sClusterFid := "7991fb33-2731-4356-af01-d2e3c4237f66"

	// 1. Start the on demand export k8s resource set snapshot job
	exportJobResp, err := infinityK8sClient.ExportK8sResourceSetSnapshot(
		ctx,
		validSnapshotFid,
		infinityk8s.ExportK8sResourceSetSnapshotJobConfig{
			TargetNamespaceName: "sdk-export-ns",
			TargetClusterFid:    validTargetK8sClusterFid,
			IgnoreErrors:        false,
			Filter:              "{}",
		},
	)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "export job response: %+v", exportJobResp)

	// 2. Use the job id of the new job and call the get job instance operation
	// exportJobResp.Id would have a valid job instance id
	getJobResp, err := infinityK8sClient.GetJobInstance(ctx, exportJobResp.Id, validCDMId)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get job response: %+v", getJobResp)
}

// TestObjectIdTranslation verifies that the SDK can perform the translation from
// Internal ID to FID and then FID back to Internal ID on a real RSC instance
// To run this test against an RSC instance, the following are required
// - a CDM cluster UUID
// - a valid object internal ID on the CDM
func TestObjectIdTranslation(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()

	validCDMId := uuid.MustParse("d5ead8ab-1129-4e23-87db-70b2317f534a")
	validObjectInternalId := uuid.MustParse("7991fb33-2731-4356-af01-d2e3c4237f66")

	// 1. Get the object FID from RSC
	fid, err := infinityK8sClient.GetK8sObjectFid(ctx, validObjectInternalId, validCDMId)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get fid response: %v", fid)

	// 2. Get back the object internal ID from the FID
	interalId, err := infinityK8sClient.GetK8sObjectInternalId(ctx, fid)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get internal id response: %v", interalId)
	if interalId != validObjectInternalId {
		t.Errorf("internal id %v doesn't match expectation %v", interalId, validObjectInternalId)
	}
	logger.Print(log.Info, "got valid internalId in response")
}

// TestAssignSLAForK8sResourceSet verifies that the SDK can perform the assign SLA
// operation for a K8s resource set on a real RSC instance
// To run this test against an RSC instance, the following are required
// - a K8s resource set fid
// - a SLA id with the correct object type
func TestAssignSLAForK8sResourceSet(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	slaClient := core.Wrap(client.GQL)
	logger := slaClient.GQL.Log()

	// 1. Assign SLA to resource set
	validResourceSetFid := uuid.MustParse("d825b9a0-9abc-53e8-a421-9c0f38ba2122")
	validSlaId := uuid.MustParse("089f4548-d268-501a-a1a6-3d6fdcf2a606")
	assignResp, err := slaClient.AssignSlaForSnappableHierarchies(
		ctx,
		&validSlaId,
		core.ProtectWithSLAID,
		[]uuid.UUID{validResourceSetFid},
		nil,
		true,  // shouldApplyToExistingSnapshots
		false, // shouldApplyToNonPolicySnapshots
		core.RetainSnapshots,
		"", // userNote
	)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "%v\n", assignResp)

	// 2. Get resourceset and check the sla
	infinityK8sClient := infinityk8s.Wrap(client)
	getResp, err := infinityK8sClient.GetK8sResourceSet(ctx, validResourceSetFid)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get succeeded, %+v", getResp)

	if getResp.EffectiveSlaDomain.ID != validSlaId.String() {
		t.Errorf(
			"sla domain id %v in the resource set object doesn't match expected value %v",
			getResp.ConfiguredSlaDomain.ID,
			validSlaId.String(),
		)
	}
}

// TestCreateK8sResourceSetSnapshot verifies that the SDK can perform the
// creation of on demand snapshot.
// To run this test against an RSC instance, the following are required
// - a K8s resource set fid
// - a SLA id with the correct object type
func TestCreateK8sResourceSetSnapshot(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()

	resourceSetFID := "f200529a-edd8-5fa0-bd3b-de869758bb68"
	slaFID := "f42158b6-0c1f-5240-afc6-fbca6fc6e856"
	ret, err := infinityK8sClient.CreateK8sResourceSnapshot(
		ctx,
		resourceSetFID,
		infinityk8s.BaseOnDemandSnapshotConfigInput{SLAID: slaFID},
	)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "response: %+v", ret)
}
