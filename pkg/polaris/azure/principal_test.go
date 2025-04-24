package azure

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
)

var (
	appID    = uuid.MustParse("4558e8e8-9493-4ece-89cf-1b9ecc026755")
	tenantID = uuid.MustParse("ac02265a-440d-42e9-abda-782d8b182d0d")
)

func TestKeyFile(t *testing.T) {
	tt := []struct {
		name         string
		keyFile      string
		tenantDomain string
	}{{
		name:         "v0",
		keyFile:      "testdata/key_file_v0.json",
		tenantDomain: "domain.onmicrosoft.com",
	}, {
		name:         "v1",
		keyFile:      "testdata/key_file_v1.json",
		tenantDomain: "domain.onmicrosoft.com",
	}, {
		name:         "v2",
		keyFile:      "testdata/key_file_v2.json",
		tenantDomain: "domain.onmicrosoft.com",
	}}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			principalFunc := KeyFile(tc.keyFile, tc.tenantDomain)
			principal, err := principalFunc(context.Background())
			if err != nil {
				t.Error(err)
			}

			if principal.appID != appID {
				t.Errorf("invalid app id: %s", principal.appID)
			}
			if principal.appName != "App" {
				t.Errorf("invalid app name: %s", principal.appName)
			}
			if principal.appSecret != "secret" {
				t.Errorf("invalid app secret: %s", principal.appSecret)
			}
			if principal.tenantID != tenantID {
				t.Errorf("invalid tenant id: %s", principal.tenantID)
			}
			if principal.tenantDomain != tc.tenantDomain {
				t.Errorf("invalid tenant domain: %s", principal.tenantDomain)
			}
		})
	}
}

func TestServicePrincipal(t *testing.T) {
	principalFunc := ServicePrincipal(appID, "App", "secret", tenantID, "domain.onmicrosoft.com")
	principal, err := principalFunc(context.Background())
	if err != nil {
		t.Error(err)
	}

	if principal.appID != appID {
		t.Errorf("invalid app id: %s", principal.appID)
	}
	if principal.appName != "App" {
		t.Errorf("invalid app name: %s", principal.appName)
	}
	if principal.appSecret != "secret" {
		t.Errorf("invalid app secret: %s", principal.appSecret)
	}
	if principal.tenantID != tenantID {
		t.Errorf("invalid tenant id: %s", principal.tenantID)
	}
	if principal.tenantDomain != "domain.onmicrosoft.com" {
		t.Errorf("invalid tenant domain: %s", principal.tenantDomain)
	}
}

func TestServicePrincipalWithoutAppName(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSub, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	principalFunc := ServicePrincipal(testSub.PrincipalID, "", testSub.PrincipalSecret, testSub.TenantID,
		testSub.TenantDomain)
	principal, err := principalFunc(context.Background())
	if err != nil {
		t.Error(err)
	}

	if principal.appID != testSub.PrincipalID {
		t.Errorf("invalid app id: %s", principal.appID)
	}
	if !strings.HasPrefix(principal.appName, "app-") {
		t.Errorf("invalid app name: %s", principal.appName)
	}
	if principal.appSecret != testSub.PrincipalSecret {
		t.Errorf("invalid app secret: %s", principal.appSecret)
	}
	if principal.tenantID != testSub.TenantID {
		t.Errorf("invalid tenant id: %s", principal.tenantID)
	}
	if principal.tenantDomain != testSub.TenantDomain {
		t.Errorf("invalid tenant domain: %s", principal.tenantDomain)
	}
}
