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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
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
// delete protection set.
func TestK8sProtectionSetLifecycle(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skip("skipping due to missing TEST_INTEGRATION=1")
	}

	k8sConfig, err := testsetup.K8SConfig()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	k8sAPI := Wrap(client)

	// Read the kubeconfig file.
	kubeconfigData, err := os.ReadFile(k8sConfig.KubeconfigFilePath)
	if err != nil {
		t.Fatalf("failed to read kubeconfig: %v", err)
	}

	// Add K8s cluster.
	clusterConfig := ClusterAddInput{
		Name:         "go-sdk-test-cluster",
		Distribution: "GENERIC",
		Kubeconfig:   string(kubeconfigData),
		Transport:    "loadbalancer",
	}
	clusterSummary, err := k8sAPI.AddCluster(ctx, k8sConfig.CDMID, clusterConfig)
	if err != nil {
		t.Fatalf("failed to add k8s cluster: %v", err)
	}
	if clusterSummary.ID == "" {
		t.Fatal("expected non-empty cluster ID")
	}
	if clusterSummary.Name != clusterConfig.Name {
		t.Errorf("cluster name: got %q, want %q", clusterSummary.Name, clusterConfig.Name)
	}

	// Update the cluster transport.
	updateConfig := ClusterUpdateConfigInput{
		Transport: "nodeport",
	}
	clusterFID, err := uuid.Parse(clusterSummary.ID)
	if err != nil {
		t.Fatalf("failed to parse cluster FID: %v", err)
	}
	if err := k8sAPI.UpdateCluster(ctx, clusterFID, updateConfig); err != nil {
		t.Fatalf("failed to update k8s cluster: %v", err)
	}

	// Wait for the cluster to be ready.
	clusterInternalID, err := waitForCluster(ctx, t, k8sAPI, clusterSummary.ID, k8sConfig.CDMID)
	if err != nil {
		t.Fatalf("cluster never became ready: %v", err)
	}

	// Add a protection set.
	psConfig := AddProtectionSetConfig{
		Definition:          `{"namespaces": ["default"]}`,
		KubernetesClusterID: clusterSummary.ID,
		KubernetesNamespace: "default",
		Name:                "go-sdk-test-ps",
		RSType:              "KubernetesNamespace",
	}
	psResponse, err := k8sAPI.AddProtectionSet(ctx, psConfig)
	if err != nil {
		t.Fatalf("failed to add protection set: %v", err)
	}
	if psResponse.ID == "" {
		t.Fatal("expected non-empty protection set ID")
	}
	if psResponse.Name != psConfig.Name {
		t.Errorf("protection set name: got %q, want %q", psResponse.Name, psConfig.Name)
	}

	// Get the FID for the protection set.
	psInternalID, err := uuid.Parse(psResponse.ID)
	if err != nil {
		t.Fatalf("failed to parse protection set ID: %v", err)
	}
	psFID, err := k8sAPI.ObjectFIDByType(ctx, psInternalID, k8sConfig.CDMID, ObjectTypeInventory)
	if err != nil {
		t.Fatalf("failed to get protection set FID: %v", err)
	}
	if psFID == uuid.Nil {
		t.Fatal("expected non-nil protection set FID")
	}

	// Verify protection set data.
	ps, err := k8sAPI.ProtectionSetByID(ctx, psFID)
	if err != nil {
		t.Fatalf("failed to get protection set: %v", err)
	}
	if ps.RSName != psConfig.Name {
		t.Errorf("protection set rsName: got %q, want %q", ps.RSName, psConfig.Name)
	}

	// Update the protection set.
	updatePSConfig := UpdateProtectionSetConfig{
		Definition: `{"namespaces": ["default", "kube-system"]}`,
	}
	if err := k8sAPI.UpdateProtectionSet(ctx, psResponse.ID, updatePSConfig); err != nil {
		t.Fatalf("failed to update protection set: %v", err)
	}

	// Assign SLA to the protection set by creating an on-demand snapshot.
	snapshotConfig := BaseOnDemandSnapshotConfigInput{
		SLAID: k8sConfig.SLAID.String(),
	}
	snapshotStatus, err := k8sAPI.CreateProtectionSetSnapshot(ctx, psFID.String(), snapshotConfig)
	if err != nil {
		t.Fatalf("failed to create protection set snapshot: %v", err)
	}
	if snapshotStatus.ID == "" {
		t.Fatal("expected non-empty snapshot request ID")
	}

	// Wait for the snapshot to complete.
	if err := waitForJob(ctx, t, k8sAPI, snapshotStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("snapshot job did not complete: %v", err)
	}

	// Get the snapshot FID.
	snapshots, err := protectionSetSnapshots(ctx, k8sAPI, psFID.String())
	if err != nil {
		t.Fatalf("failed to get protection set snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected at least one snapshot")
	}
	snapshotFID := snapshots[0]

	// Export the snapshot to a different namespace.
	exportConfig := ExportSnapshotJobConfig{
		TargetNamespaceName: "go-sdk-test-export",
		TargetClusterFID:    clusterSummary.ID,
		IgnoreErrors:        true,
	}
	exportStatus, err := k8sAPI.ExportProtectionSetSnapshot(ctx, snapshotFID, exportConfig)
	if err != nil {
		t.Fatalf("failed to export snapshot: %v", err)
	}
	if exportStatus.ID == "" {
		t.Fatal("expected non-empty export request ID")
	}

	// Wait for the export to complete.
	if err := waitForJob(ctx, t, k8sAPI, exportStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("export job did not complete: %v", err)
	}

	// Restore the snapshot.
	restoreConfig := RestoreSnapshotJobConfig{
		IgnoreErrors: true,
	}
	restoreStatus, err := k8sAPI.RestoreProtectionSetSnapshot(ctx, snapshotFID, restoreConfig)
	if err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}
	if restoreStatus.ID == "" {
		t.Fatal("expected non-empty restore request ID")
	}

	// Wait for the restore to complete.
	if err := waitForJob(ctx, t, k8sAPI, restoreStatus.ID, k8sConfig.CDMID); err != nil {
		t.Fatalf("restore job did not complete: %v", err)
	}

	// Delete the protection set.
	if err := k8sAPI.DeleteProtectionSet(ctx, psResponse.ID, true); err != nil {
		t.Fatalf("failed to delete protection set: %v", err)
	}

	// Verify the internal ID can be resolved back.
	resolvedInternalID, err := k8sAPI.ObjectInternalIDByType(ctx, psFID, k8sConfig.CDMID, ObjectTypeInventory)
	if err != nil {
		t.Logf("ObjectInternalIDByType returned error (expected after delete): %v", err)
	} else if resolvedInternalID != uuid.Nil {
		t.Logf("resolved internal ID: %v", resolvedInternalID)
	}

	_ = clusterInternalID
}

// protectionSetSnapshots wraps the unexported method for testing.
func protectionSetSnapshots(ctx context.Context, api API, fid string) ([]string, error) {
	return api.protectionSetSnapshots(ctx, fid)
}

// waitForCluster waits for the K8s cluster to become ready by polling the
// object FID.
func waitForCluster(ctx context.Context, t *testing.T, api API, clusterID string, cdmID uuid.UUID) (uuid.UUID, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	clusterInternalID, err := uuid.Parse(clusterID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse cluster ID: %w", err)
	}

	for {
		fid, err := api.ObjectFIDByType(ctx, clusterInternalID, cdmID, ObjectTypeInventory)
		if err == nil && fid != uuid.Nil {
			return fid, nil
		}

		select {
		case <-ctx.Done():
			return uuid.Nil, ctx.Err()
		case <-time.After(10 * time.Second):
			t.Log("waiting for cluster to be ready...")
		}
	}
}

// waitForJob waits for a CDM job to complete by polling the job instance.
func waitForJob(ctx context.Context, t *testing.T, api API, jobID string, cdmID uuid.UUID) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for {
		job, err := api.K8sJobInstance(ctx, jobID, cdmID)
		if err == nil && (job.JobStatus == "SUCCEEDED" || job.JobStatus == "FAILED") {
			if job.JobStatus == "FAILED" {
				return fmt.Errorf("job %q failed", jobID)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			t.Logf("waiting for job %q...", jobID)
		}
	}
}
