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
	cdmID      = uuid.MustParse("8c02fa3d-e452-4d40-82f2-ff94812e994a")
	k8sFID     = uuid.MustParse("71cb046b-f588-5c22-8599-57caab3fd44b")
	goldSLAFID = uuid.MustParse("351d014c-f601-5e7d-981e-16607e69d7a4")
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

// TestIntegration tests the flow from creation of a ResourceSet to export.
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

	// 1. Add ResourceSet.
	config := infinityk8s.AddK8sResourceSetConfig{
		KubernetesClusterUUID: k8sFID.String(),
		KubernetesNamespace:   "default",
		Definition:            "{}",
		Name:                  "default-rs",
		RSType:                "namespace",
	}

	addResp, err := infinityK8sClient.AddK8sResourceSet(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	if addResp.ID == "" {
		t.Errorf("add failed, %+v", addResp)
	}

	logger.Printf(log.Info, "add succeeded, %+v", addResp)

	// 2. Get resourceset.
	// If we get to this point, addResp.ID should be a valid uuid.
	rsFID := uuid.Must(uuid.Parse(addResp.ID))
	getResp, err := infinityK8sClient.GetK8sResourceSet(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get succeeded, %+v", getResp)

	defer func() {
		// N + 1. Delete resourceset.
		delResp, err := infinityK8sClient.DeleteK8sResourceSet(
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
	assignResp, err := slaClient.AssignSlaForSnappableHierarchies(
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

	// 4. Get resourceset and check the sla.
	time.Sleep(1 * time.Minute)
	getResp, err = infinityK8sClient.GetK8sResourceSet(ctx, rsFID)
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
	snapshotRet, err := infinityK8sClient.CreateK8sResourceSnapshot(
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
				snapshots, err := infinityK8sClient.GetResourceSetSnapshots(
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

	// 8. Start the on demand export k8s resource set snapshot job
	exportJobResp, err := infinityK8sClient.ExportK8sResourceSetSnapshot(
		ctx,
		snaps[0],
		infinityk8s.ExportK8sResourceSetSnapshotJobConfig{
			TargetNamespaceName: "default-export-ns",
			TargetClusterFid:    k8sFID.String(),
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

	// 11. Translate FID to internal_id and back.
	interalID, err := infinityK8sClient.GetK8sObjectInternalID(ctx, rsFID)
	if err != nil {
		t.Error(err)
		return
	}
	logger.Printf(log.Info, "get internal id response: %v", interalID)

	// Get the object FID from RSC
	fid, err := infinityK8sClient.GetK8sObjectFid(
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
