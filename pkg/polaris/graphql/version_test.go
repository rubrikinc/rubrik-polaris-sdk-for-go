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

func TestVersionBefore(t *testing.T) {
	// Latest - special case, always returns false with no error.
	before, err := Version("latest").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("latest should not return error: %v", err)
	}
	if before {
		t.Error("latest should not be before any version")
	}

	// Master
	before, err = Version("master-47837").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("master-47837 should not return error: %v", err)
	}
	if !before {
		t.Error("master-47837 should be older than master-47838")
	}

	before, err = Version("master-47838").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("master-47838 should not return error: %v", err)
	}
	if before {
		t.Error("master-47838 should not be older than master-47838")
	}

	before, err = Version("master-47839").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("master-47839 should not return error: %v", err)
	}
	if before {
		t.Error("master-47839 should not be older than master-47838")
	}

	// Full release version
	before, err = Version("v20210727-7").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("v20210727-7 should not return error: %v", err)
	}
	if !before {
		t.Error("v20210727-7 should be older than v20210727-8")
	}

	before, err = Version("v20210727-8").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("v20210727-8 should not return error: %v", err)
	}
	if before {
		t.Error("v20210727-8 should not be older than v20210727-8")
	}

	before, err = Version("v20210727-9").Before("master-47838", "v20210727-8")
	if err != nil {
		t.Errorf("v20210727-9 should not return error: %v", err)
	}
	if before {
		t.Error("v20210727-9 should not be older than v20210727-8")
	}

	// Partial release version.
	before, err = Version("v20210726-8").Before("master-47838", "v20210727")
	if err != nil {
		t.Errorf("v20210726-8 should not return error: %v", err)
	}
	if !before {
		t.Error("v20210726-7 should be older than v20210727")
	}

	before, err = Version("v20210727-7").Before("master-47838", "v20210727")
	if err != nil {
		t.Errorf("v20210727-7 should not return error: %v", err)
	}
	if before {
		t.Error("v20210727-8 should not be older than v20210727")
	}

	before, err = Version("v20210728-6").Before("master-47838", "v20210727")
	if err != nil {
		t.Errorf("v20210728-6 should not return error: %v", err)
	}
	if before {
		t.Error("v20210728-9 should not be older than v20210727")
	}

	// No matching format in tags - should return error.
	before, err = Version("master-47838").Before("v20210727-8")
	if err == nil {
		t.Error("master version should return error when only release tags provided")
	}
	if before {
		t.Error("master version should not be before when only release tags provided")
	}

	before, err = Version("v20210727-8").Before("master-47838")
	if err == nil {
		t.Error("release version should return error when only master tags provided")
	}
	if before {
		t.Error("release version should not be before when only master tags provided")
	}
}
