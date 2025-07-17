package infinityk8s_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/infinityk8s"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
var (
	cdmID      = uuid.MustParse("e257be55-ebbc-4870-8d84-b57aa61445f3")
	goldSLAFID = uuid.MustParse("ff5a4d4d-1da0-58c3-a2a9-6aecbcc5c60d")
	client     *polaris.Client
	trueValue  = true
	falseValue = false

	k8sClusterAddWithKubeconfig = infinityk8s.K8sClusterAddInput{
		Name:         "test-k8s-cluster",
		Distribution: "vanilla",
		Kubeconfig:   "<kubeconfig>",
		Transport:    "loadbalancer",
	}

	k8sUpdateTransportConfig = infinityk8s.K8sClusterUpdateConfigInput{
		Transport: "nodeport",
	}
)

func TestMain(m *testing.M) {

	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// When enabling integration tests, uncomment the below section.
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(false)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
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

// TestIntegration tests the flow from creation of a ProtectionSet to export.
// Needs
// - RSC service account setup in ~/.rubrik/polaris-service-account.json
// - CDM onboarded onto above account and its ID.
// - A valid kubeconfig set in k8sClusterAddWithKubeconfig.Kubeconfig.
// - Gold SLA fid from RSC cluster. The SLA should be owned by CDM.
func TestIntegration(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 0. Add K8s Cluster.
	addK8sResponse, err := infinityK8sClient.AddK8sCluster(
		ctx,
		cdmID,
		k8sClusterAddWithKubeconfig,
	)
	if err != nil {
		t.Error(err)
		return
	}
	k8sFID := uuid.MustParse(addK8sResponse.ID)
	logger.Printf(log.Info, "add k8s cluster succeeded, %+v", addK8sResponse)

	// 0.1 Update K8s Cluster.
	updateK8sResponse, err := infinityK8sClient.UpdateK8sCluster(
		ctx,
		k8sFID,
		k8sUpdateTransportConfig,
	)
	if err != nil {
		t.Error(err)
		return
	}
	if !updateK8sResponse {
		t.Error("update k8s cluster failed")
		return
	}
	logger.Printf(
		log.Info,
		"update k8s cluster succeeded, %+v",
		updateK8sResponse,
	)

	// 1. Add ProtectionSet.
	config := infinityk8s.AddK8sProtectionSetConfig{
		KubernetesClusterID: k8sFID.String(),
		KubernetesNamespace: "default",
		Definition:          "{}",
		Name:                "default-rs",
		RSType:              "namespace",
	}

	addResp, err := infinityK8sClient.AddK8sProtectionSet(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	if addResp.ID == "" {
		t.Errorf("add protection set failed, %+v", addResp)
	}

	logger.Printf(log.Info, "add protection set succeeded, %+v", addResp)

	// 2. Get ProtectionSet.
	// If we get to this point, addResp.ID should be a valid uuid.
	rsFID := uuid.Must(uuid.Parse(addResp.ID))
	getResp, err := infinityK8sClient.K8sProtectionSet(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get protection set succeeded, %+v", getResp)

	defer func() {
		// N + 1. Delete ProtectionSet.
		delResp, err := infinityK8sClient.DeleteK8sProtectionSet(
			ctx,
			rsFID.String(),
			false,
		)
		if err != nil {
			t.Error(err)
			return
		}
		if delResp != true {
			t.Errorf("delete protection set failed, %v", delResp)
		}
		logger.Printf(log.Info, "del protection set succeeded, %+v", delResp)
	}()

	// 2.1 update ProtectionSet.
	updateResp, err := infinityK8sClient.UpdateK8sProtectionSet(
		ctx,
		rsFID.String(),
		infinityk8s.UpdateK8sProtectionSetConfig{
			Definition: "{\"includes\": [{\"resource\": \"Deployment\"}]}",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	if updateResp != true {
		t.Errorf("update protection set failed, %v", updateResp)
	}
	logger.Printf(log.Info, "update protection set succeeded, %+v", updateResp)

	// 3. Assign SLA to resource set
	slaClient := core.Wrap(client.GQL)
	assignResp, err := slaClient.AssignSLAForSnappableHierarchies(
		ctx,
		&goldSLAFID,
		core.ProtectWithSLAID,
		[]uuid.UUID{rsFID},
		nil,
		&trueValue,  // shouldApplyToExistingSnapshots
		&falseValue, // shouldApplyToNonPolicySnapshots
		nil,
		nil, // userNote
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "Assign SLA response %v\n", assignResp)

	// 4. Get ProtectionSet and check the sla.
	time.Sleep(1 * time.Minute)
	getResp, err = infinityK8sClient.K8sProtectionSet(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get protection set succeeded, %+v", getResp)

	if getResp.EffectiveSLADomain.ID != goldSLAFID.String() {
		logger.Printf(
			log.Warn,
			"sla domain id %v in the resource set object doesn't match expected value %v",
			getResp.ConfiguredSLADomain.ID,
			goldSLAFID.String(),
		)
	}

	// 5. Create an on-demand snapshot.
	snapshotRet, err := infinityK8sClient.CreateK8sProtectionSetSnapshot(
		ctx,
		rsFID.String(),
		infinityk8s.BaseOnDemandSnapshotConfigInput{SLAID: goldSLAFID.String()},
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "response: %+v", snapshotRet)

	// 6. Get the status of the job.
	getJobResp, err := infinityK8sClient.JobInstance(
		ctx,
		snapshotRet.ID,
		cdmID.String(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get snapshot job response: %+v", getJobResp)

	// 7. Wait for a snapshot.
	snaps, err := func() ([]string, error) {
		timeoutCTX, cancel := context.WithDeadline(
			ctx,
			time.Now().Add(5*time.Minute),
		)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return nil, errors.New("timeout")
			default:
				snapshots, err := infinityK8sClient.ProtectionSetSnapshots(
					timeoutCTX,
					rsFID.String(),
				)
				if err != nil {
					return nil, err
				}
				if len(snapshots) >= 1 {
					return snapshots, nil
				}
				time.Sleep(30 * time.Second)
			}
		}
	}()

	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "Get snapshot response: %+v", snaps)

	// 6.1 check for events on the snapshot job. Since the job is complete, we
	// should see all the events.
	esi := getJobResp.EventSeriesID
	series, err := infinityK8sClient.ActivitySeries(
		ctx,
		uuid.MustParse(esi),
		cdmID,
	)
	if err != nil {
		t.Error(err)
		return
	}
	for _, act := range series {
		logger.Printf(
			log.Info,
			"snapshot activity: %s, %s, %s",
			act.ActivityInfo,
			act.Message,
			act.Status,
		)
	}

	// 8. Start the on demand export k8s resource set snapshot job
	exportJobResp, err := infinityK8sClient.ExportK8sProtectionSetSnapshot(
		ctx,
		snaps[0],
		infinityk8s.ExportK8sProtectionSetSnapshotJobConfig{
			TargetNamespaceName: "default-export-ns",
			TargetClusterFID:    k8sFID.String(),
			IgnoreErrors:        false,
			Filter:              "{}",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "export job response: %+v", exportJobResp)

	// 9. Use the job id of the new job and call the get job instance operation
	// exportJobResp.ID would have a valid job instance id
	getJobResp, err = infinityK8sClient.JobInstance(
		ctx,
		exportJobResp.ID,
		cdmID.String(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get job response: %+v", getJobResp)

	// 10 check for events on the restore job. Since the job is incomplete, we
	// should not see all the events.
	resi := getJobResp.EventSeriesID
	rseries, err := infinityK8sClient.ActivitySeries(
		ctx,
		uuid.MustParse(resi),
		cdmID,
	)
	if err != nil {
		t.Error(err)
		return
	}
	for _, act := range rseries {
		logger.Printf(
			log.Info,
			"restore activity: %s, %s, %s",
			act.ActivityInfo,
			act.Message,
			act.Status,
		)
	}

	// 11. Start the on demand restore k8s resource set snapshot job
	restoreJobResp, err := infinityK8sClient.RestoreK8sProtectionSetSnapshot(
		ctx,
		snaps[0],
		infinityk8s.RestoreK8sProtectionSetSnapshotJobConfig{
			IgnoreErrors: false,
			Filter:       "{}",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "restore job response: %+v", restoreJobResp)

	// 12. Use the job id of the new job and call the get job instance operation
	// restoreJobResp.ID would have a valid job instance id
	getJobResp, err = infinityK8sClient.JobInstance(
		ctx,
		restoreJobResp.ID,
		cdmID.String(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	// 13. Translate FID to internal_id and back.
	internalID, err := infinityK8sClient.K8sObjectInternalIDByType(
		ctx,
		rsFID,
		cdmID,
		infinityk8s.K8sObjectTypeInventory,
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get internal id response: %v", internalID)

	// Get the object FID from RSC
	fid, err := infinityK8sClient.K8sObjectFIDByType(
		ctx,
		internalID,
		cdmID,
		infinityk8s.K8sObjectTypeInventory,
	)
	if err != nil {
		t.Error(err)
	}
	logger.Printf(log.Info, "get fid response: %v", fid)

	if fid != rsFID {
		t.Errorf(
			"internal id %s doesn't match expectation %s",
			fid.String(),
			rsFID.String(),
		)
	}
}

func TestIntegrationTemp(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}
	slaFID := uuid.MustParse("95475fd9-070f-49ae-800d-6e8631ebf8a3")
	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)
	// 1. Assign SLA to resource set
	slaClient := core.Wrap(client.GQL)
	retention := core.RetainSnapshots
	slaResp, err := slaClient.AssignSLAForSnappableHierarchies(
		ctx,
		&slaFID,
		core.ProtectWithSLAID,
		[]uuid.UUID{uuid.MustParse("288e8033-9d6b-5a58-850b-62c53579dd3d")},
		nil,
		nil,         // shouldApplyToExistingSnapshots
		&falseValue, // shouldApplyToNonPolicySnapshots
		&retention,
		nil, // userNote
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "Assign SLA response %v\n", slaResp)
}

func TestIntegrationK8sJobInstance(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}
	cdmUUID := uuid.MustParse("8f7ce6e6-5d20-4496-b45d-ea2c50efa025")
	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)
	// 1. Get the Job status using k8sJobInstance endpoint.
	jobResp, err := infinityK8sClient.K8sJobInstance(
		ctx,
		"CREATE_MN_K8S_SNAPSHOT_f7101244-c3ca-495d-946c-d4d0c2cd676c_cc859058-f5fd-48ea-8aa2-ccb80fc089c6:::0",
		cdmUUID,
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "Get k8sJobInstance response %v\n", jobResp)
}

func TestIntegrationTranslation(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 1. Translate Global SLA FID to internal_id and back.
	silverFID := uuid.MustParse("313297f1-0b9d-4a1f-826c-ccc607dae06a")
	clusterUUID := uuid.MustParse("10ae4970-e22f-4d92-a04f-cbf241103190")
	internalID, err := infinityK8sClient.K8sObjectInternalIDByType(
		ctx,
		silverFID,
		clusterUUID,
		infinityk8s.K8sObjectTypeSLA,
	)
	if err != nil {
		t.Error(err)
		return
	}

	logger.Printf(log.Info, "get internal id response: %v", internalID)

	// Get the object FID from RSC
	fid, err := infinityK8sClient.K8sObjectFIDByType(
		ctx,
		internalID,
		clusterUUID,
		infinityk8s.K8sObjectTypeSLA,
	)
	if err != nil {
		t.Error(err)
		return
	}

	logger.Printf(log.Info, "get fid response: %v", fid)
	if silverFID != fid {
		t.Errorf(
			"internal id %s doesn't match expectation %s",
			fid.String(),
			silverFID.String(),
		)
	}

	// 2. Translate internal_id to Global SLA FID and back.
	slaInternalID := uuid.MustParse("ff8be367-25c4-43bf-bd56-e16d66984aef")
	fid, err = infinityK8sClient.K8sObjectFIDByType(
		ctx,
		slaInternalID,
		clusterUUID,
		infinityk8s.K8sObjectTypeSLA,
	)
	if err != nil {
		t.Error(err)
		return
	}

	logger.Printf(log.Info, "get fid response: %v", fid)

	// Get the object internal_id from RSC
	internalID, err = infinityK8sClient.K8sObjectInternalIDByType(
		ctx,
		fid,
		clusterUUID,
		infinityk8s.K8sObjectTypeSLA,
	)
	if err != nil {
		t.Error(err)
		return
	}

	logger.Printf(log.Info, "get internal id response: %v", internalID)
	if internalID != slaInternalID {
		t.Errorf(
			"internal id %s doesn't match expectation %s",
			internalID.String(),
			slaInternalID.String(),
		)
	}
}

func TestMissingObjectK8sObjectFIDByType(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 1. Translate Global SLA FID to internal_id and back.
	missingInternalID := uuid.MustParse("313297f1-0b9d-4a1f-826c-ccc607dae06b")
	clusterUUID := uuid.MustParse("10ae4970-e22f-4d92-a04f-cbf241103190")
	_, err := infinityK8sClient.K8sObjectFIDByType(
		ctx,
		missingInternalID,
		clusterUUID,
		infinityk8s.K8sObjectTypeInventory,
	)
	t.Logf("error: %v", err)
	if err == nil {
		t.Error("expected error, got nil")
		return
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected not found error, got %v", err)
		return
	}
}

func TestMissingK8sProtectionSet(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 1. Get ProtectionSet.
	_, err := infinityK8sClient.K8sProtectionSet(
		ctx,
		uuid.MustParse("313297f1-0b9d-4a1f-826c-ccc607dae06b"),
	)
	t.Logf("error: %v", err)
	if err == nil {
		t.Error("expected error, got nil")
		return
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected not found error, got %v", err)
		return
	}
}
