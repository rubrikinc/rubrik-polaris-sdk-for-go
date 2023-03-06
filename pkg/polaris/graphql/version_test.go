package graphql

import "testing"

func TestQueryName(t *testing.T) {
	if name := QueryName("query SdkGolangMyQuery {}"); name != "myQuery" {
		t.Fatalf("invalid query name: %s", name)
	}

	if name := QueryName("query SdkGolangMyQuery($param: int) {}"); name != "myQuery" {
		t.Fatalf("invalid query name: %s", name)
	}

	if name := QueryName("invalidquery"); name != "<invalid-query>" {
		t.Fatalf("invalid query name: %s", name)
	}
}

func TestVersionOlderThan(t *testing.T) {
	// Latest
	if Version("latest").Before("master-47838", "v20210727-8") {
		t.Error("latest should not be older than master-47838 or v20210727-8")
	}

	// Master
	if !Version("master-47837").Before("master-47838", "v20210727-8") {
		t.Error("master-47837 should be older than master-47838")
	}

	if Version("master-47838").Before("master-47838", "v20210727-8") {
		t.Error("master-47838 should not be older than master-47838")
	}

	if Version("master-47839").Before("master-47838", "v20210727-8") {
		t.Error("master-47839 should not be older than master-47838")
	}

	// Full release version
	if !Version("v20210727-7").Before("master-47838", "v20210727-8") {
		t.Error("v20210727-7 should be older than v20210727-8")
	}

	if Version("v20210727-8").Before("master-47838", "v20210727-8") {
		t.Error("v20210727-8 should not be older than v20210727-8")
	}

	if Version("v20210727-9").Before("master-47838", "v20210727-8") {
		t.Error("v20210727-9 should not be older than v20210727-8")
	}

	// Partial release version.
	if !Version("v20210726-8").Before("master-47838", "v20210727") {
		t.Error("v20210726-7 should be older than v20210727")
	}

	if Version("v20210727-7").Before("master-47838", "v20210727") {
		t.Error("v20210727-8 should not be older than v20210727")
	}

	if Version("v20210728-6").Before("master-47838", "v20210727") {
		t.Error("v20210728-9 should not be older than v20210727")
	}
}
