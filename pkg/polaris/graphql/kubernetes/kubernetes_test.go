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

package kubernetes

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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
var client *graphql.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests default the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		polClient, err := polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
		client = polClient.GQL
	}

	os.Exit(m.Run())
}

// TestK8sProtectionSetLifecycle tests the full lifecycle of a K8s protection
// set: add cluster, create protection set, take snapshot, export, restore,
// delete protection set, delete cluster.
func TestK8sProtectionSetLifecycle(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skip("skipping due to missing TEST_INTEGRATION=1")
	}

	k8sConfig, err := testsetup.K8SConfig()
	if errors.Is(err, testsetup.ErrK8SConfigNotSet) {
		t.Skip("skipping due to missing TEST_K8SCONFIG_FILE")
	}
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	k8sAPI := Wrap(client)

	kubeconfigData, err := os.ReadFile(k8sConfig.KubeconfigFilePath)
	if err != nil {
		t.Fatalf("failed to read kubeconfig: %v", err)
	}

	// Add K8s cluster.
	clusterConfig := ClusterAddInput{
		Name:         "go-sdk-test-cluster",
		Distribution: "VANILLA",
		Kubeconfig:   secret.String(kubeconfigData),
		Transport:    "loadbalancer",
	}
	clusterSummary, err := k8sAPI.AddCluster(ctx, k8sConfig.CDMID, clusterConfig)
	if err != nil {
		t.Fatalf("failed to add k8s cluster: %v", err)
	}
	if clusterSummary.Name != clusterConfig.Name {
		t.Errorf("cluster name: got %q, want %q", clusterSummary.Name, clusterConfig.Name)
	}
	clusterFID, err := uuid.Parse(clusterSummary.ID)
	if err != nil {
		t.Fatalf("failed to parse cluster FID: %v", err)
	}
	t.Cleanup(func() {
		if _, err := k8sAPI.DeleteCluster(ctx, clusterFID, true, false); err != nil {
			t.Errorf("failed to delete k8s cluster: %v", err)
		}
	})

	// Update the cluster transport.
	if err := k8sAPI.UpdateCluster(ctx, clusterFID, ClusterUpdateConfigInput{Transport: "nodeport"}); err != nil {
		t.Fatalf("failed to update k8s cluster: %v", err)
	}

	if err := waitForCluster(ctx, t, k8sAPI, clusterFID, k8sConfig.CDMID); err != nil {
		t.Fatalf("cluster never became ready: %v", err)
	}

	// Add a protection set.
	psConfig := AddProtectionSetConfig{
		Definition:          `{"exclude":[{"resource":"Secret"}]}`,
		KubernetesClusterID: clusterSummary.ID,
		KubernetesNamespace: "default",
		Name:                "go-sdk-test-ps",
		RSType:              "namespace",
	}
	psResponse, err := k8sAPI.AddProtectionSet(ctx, psConfig)
	if err != nil {
		t.Fatalf("failed to add protection set: %v", err)
	}
	if psResponse.Name != psConfig.Name {
		t.Errorf("protection set name: got %q, want %q", psResponse.Name, psConfig.Name)
	}
	psFID, err := uuid.Parse(psResponse.ID)
	if err != nil {
		t.Fatalf("failed to parse protection set FID: %v", err)
	}
	psDeleted := false
	t.Cleanup(func() {
		if psDeleted {
			return
		}
		if err := k8sAPI.DeleteProtectionSet(ctx, psResponse.ID, false); err != nil {
			t.Errorf("failed to delete protection set: %v", err)
		}
	})

	ps, err := k8sAPI.ProtectionSetByID(ctx, psFID)
	if err != nil {
		t.Fatalf("failed to get protection set: %v", err)
	}
	if ps.RSName != psConfig.Name {
		t.Errorf("protection set rsName: got %q, want %q", ps.RSName, psConfig.Name)
	}

	// Round-trip the protection set FID through its CDM internal ID.
	psInternalID, err := k8sAPI.ObjectInternalIDByType(ctx, psFID, k8sConfig.CDMID, ObjectTypeInventory)
	if err != nil {
		t.Fatalf("failed to get protection set internal ID: %v", err)
	}
	resolvedFID, err := k8sAPI.ObjectFIDByType(ctx, psInternalID, k8sConfig.CDMID, ObjectTypeInventory)
	if err != nil {
		t.Fatalf("failed to get protection set FID: %v", err)
	}
	if resolvedFID != psFID {
		t.Errorf("protection set FID: got %v, want %v", resolvedFID, psFID)
	}

	// Take an on-demand snapshot.
	snapshotStatus, err := k8sAPI.CreateProtectionSetSnapshot(ctx, psResponse.ID, BaseOnDemandSnapshotConfigInput{
		SLAID: k8sConfig.SLAID.String(),
	})
	if err != nil {
		t.Fatalf("failed to create protection set snapshot: %v", err)
	}
	if err := waitForJob(ctx, t, k8sAPI, snapshotStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("snapshot job did not complete: %v", err)
	}

	snapshots, err := k8sAPI.protectionSetSnapshots(ctx, psResponse.ID)
	if err != nil {
		t.Fatalf("failed to get protection set snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected at least one snapshot")
	}
	snapshotFID := snapshots[0]

	// Export the snapshot to a different namespace.
	exportStatus, err := k8sAPI.ExportProtectionSetSnapshot(ctx, snapshotFID, ExportSnapshotJobConfig{
		TargetNamespaceName: "go-sdk-test-export",
		TargetClusterFID:    clusterSummary.ID,
		IgnoreErrors:        true,
	})
	if err != nil {
		t.Fatalf("failed to export snapshot: %v", err)
	}
	if err := waitForJob(ctx, t, k8sAPI, exportStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("export job did not complete: %v", err)
	}

	// Restore the snapshot.
	restoreStatus, err := k8sAPI.RestoreProtectionSetSnapshot(ctx, snapshotFID, RestoreSnapshotJobConfig{
		IgnoreErrors: true,
	})
	if err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}
	if err := waitForJob(ctx, t, k8sAPI, restoreStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("restore job did not complete: %v", err)
	}

	if err := k8sAPI.DeleteProtectionSet(ctx, psResponse.ID, false); err != nil {
		t.Fatalf("failed to delete protection set: %v", err)
	}
	psDeleted = true
}

// waitForCluster waits for the K8s cluster to show up in the RSC hierarchy by
// polling for its CDM internal ID.
func waitForCluster(ctx context.Context, t *testing.T, api API, clusterFID, cdmID uuid.UUID) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	for {
		_, err := api.ObjectInternalIDByType(ctx, clusterFID, cdmID, ObjectTypeInventory)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w (last error: %v)", ctx.Err(), err)
		case <-time.After(10 * time.Second):
			t.Log("waiting for cluster to be ready...")
		}
	}
}

// waitForJob waits for a CDM job to reach a terminal state by polling the job
// instance.
func waitForJob(ctx context.Context, t *testing.T, api API, jobID string, cdmID uuid.UUID) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	for {
		job, err := api.K8sJobInstance(ctx, jobID, cdmID)
		if err == nil {
			switch job.JobStatus {
			case "SUCCEEDED":
				return nil
			case "FAILED", "CANCELED":
				return fmt.Errorf("job %q: %s", jobID, job.JobStatus)
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w (last status: %q, last error: %v)", ctx.Err(), job.JobStatus, err)
		case <-time.After(10 * time.Second):
			t.Logf("waiting for job %q...", jobID)
		}
	}
}
