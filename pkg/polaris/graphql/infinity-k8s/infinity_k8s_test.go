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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	infinityk8s "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/infinity-k8s"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
var (
	cdmID      = uuid.MustParse("22950e27-9a7d-4ac1-b32d-0d858de01f9e")
	k8sFID     = uuid.MustParse("ea176753-d4ee-59b0-b770-54f77335abe3")
	goldSLAFID = uuid.MustParse("34d1f3c4-3521-5747-836e-4345b363175d")
	client     *polaris.Client
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
// - Onboarded K8s Cluster and its FID.
// - Gold SLA fid from RSC cluster. The SLA should be owned by CDM.
func TestIntegration(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 1. Add ProtectionSet.
	config := infinityk8s.AddK8sProtectionSetConfig{
		KubernetesClusterId: k8sFID.String(),
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
		t.Errorf("add failed, %+v", addResp)
	}

	logger.Printf(log.Info, "add succeeded, %+v", addResp)

	// 2. Get ProtectionSet.
	// If we get to this point, addResp.ID should be a valid uuid.
	rsFID := uuid.Must(uuid.Parse(addResp.ID))
	getResp, err := infinityK8sClient.GetK8sProtectionSet(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get succeeded, %+v", getResp)

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
			t.Errorf("delete failed, %v", delResp)
		}
		logger.Printf(log.Info, "del succeeded, %+v", delResp)
	}()

	// 3. Assign SLA to resource set
	slaClient := core.Wrap(client.GQL)
	assignResp, err := slaClient.AssignSLAForSnappableHierarchies(
		ctx,
		&goldSLAFID,
		core.ProtectWithSLAID,
		[]uuid.UUID{rsFID},
		nil,
		true,  // shouldApplyToExistingSnapshots
		false, // shouldApplyToNonPolicySnapshots
		core.RetainSnapshots,
		"", // userNote
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "Assign SLA response %v\n", assignResp)

	// 4. Get ProtectionSet and check the sla.
	time.Sleep(1 * time.Minute)
	getResp, err = infinityK8sClient.GetK8sProtectionSet(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get succeeded, %+v", getResp)

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
	getJobResp, err := infinityK8sClient.GetJobInstance(
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
				snapshots, err := infinityK8sClient.GetProtectionSetSnapshots(
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
	series, err := infinityK8sClient.GetActivitySeries(
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
	getJobResp, err = infinityK8sClient.GetJobInstance(
		ctx,
		exportJobResp.ID,
		cdmID.String(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get job response: %+v", getJobResp)

	// 9.1 check for events on the restore job. Since the job is incomplete, we
	// should not see all the events.
	resi := getJobResp.EventSeriesID
	rseries, err := infinityK8sClient.GetActivitySeries(
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

	// 11. Translate FID to internal_id and back.
	interalID, err := infinityK8sClient.GetK8sObjectInternalID(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get internal id response: %v", interalID)

	// Get the object FID from RSC
	fid, err := infinityK8sClient.GetK8sObjectFID(
		ctx,
		interalID,
		cdmID,
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

	infinityK8sClient := infinityk8s.Wrap(client)
	logger := infinityK8sClient.GQL.Log()
	logger.SetLogLevel(log.Debug)

	// 6. Get the status of the job.
	getJobResp, err := infinityK8sClient.GetJobInstance(
		ctx,
		"CREATE_MN_K8S_SNAPSHOT_d75bccbe-2a4e-4b0b-8ca1-c6a604540a03_80e516da-fd28-4091-80a1-2bb60e16d6e3:::0",
		cdmID.String(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get snapshot job response: %+v", getJobResp)

	// 6.1 check for events on the snapshot job. Since the job is complete, we
	// should see all the events.
	esi := getJobResp.EventSeriesID
	series, err := infinityK8sClient.GetActivitySeries(
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
			"snapshot activity: %s, %s, %s, %s, %+v",
			act.ActivityInfo,
			act.Message,
			act.Status,
			act.Severity,
			act.Time,
		)
	}
}
