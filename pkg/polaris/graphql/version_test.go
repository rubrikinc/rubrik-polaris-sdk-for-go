package graphql

import "testing"

func TestQueryName(t *testing.T) {
	if name := queryName(azureCloudAccountTenantsQuery); name != "azureCloudAccountTenants" {
		t.Fatalf("invalid query name: %s", name)
	}

	if name := queryName(azureCloudAccountTenantsV0Query); name != "azureCloudAccountTenantsV0" {
		t.Fatalf("invalid query name: %s", name)
	}

	if name := queryName("invalidquery"); name != "<invalid-query>" {
		t.Fatalf("invalid query name: %s", name)
	}
}

func TestVersionOlderThan(t *testing.T) {
	// Latest
	if VersionOlderThan("latest", "master-47838", "v20210727-8") {
		t.Error("latest should not be older than master-47838 or v20210727-8")
	}

	// Master
	if !VersionOlderThan("master-47837", "master-47838", "v20210727-8") {
		t.Error("master-47837 should be older than master-47838")
	}

	if VersionOlderThan("master-47838", "master-47838", "v20210727-8") {
		t.Error("master-47838 should not be older than master-47838")
	}

	if VersionOlderThan("master-47839", "master-47838", "v20210727-8") {
		t.Error("master-47839 should not be older than master-47838")
	}

	// Full release version
	if !VersionOlderThan("v20210727-7", "master-47838", "v20210727-8") {
		t.Error("v20210727-7 should be older than v20210727-8")
	}

	if VersionOlderThan("v20210727-8", "master-47838", "v20210727-8") {
		t.Error("v20210727-8 should not be older than v20210727-8")
	}

	if VersionOlderThan("v20210727-9", "master-47838", "v20210727-8") {
		t.Error("v20210727-9 should not be older than v20210727-8")
	}

	// Partial release version.
	if !VersionOlderThan("v20210726-8", "master-47838", "v20210727") {
		t.Error("v20210726-7 should be older than v20210727")
	}

	if VersionOlderThan("v20210727-7", "master-47838", "v20210727") {
		t.Error("v20210727-8 should not be older than v20210727")
	}

	if VersionOlderThan("v20210728-6", "master-47838", "v20210727") {
		t.Error("v20210728-9 should not be older than v20210727")
	}
}
