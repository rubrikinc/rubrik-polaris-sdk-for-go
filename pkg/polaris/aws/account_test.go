package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
)

func TestProfile(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	config, err := Profile(testAccount.Profile)(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id := config.NativeID; id != testAccount.AccountID {
		t.Errorf("invalid account id: %s", id)
	}
	if name := config.name; name != fmt.Sprintf("%s : %s", testAccount.AccountID, testAccount.Profile) {
		t.Errorf("invalid account name: %s", name)
	}
}

func TestProfileWithAccountID(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	config, err := ProfileWithAccountID(testAccount.Profile, "123456789012")(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id := config.NativeID; id != "123456789012" {
		t.Errorf("invalid account id: %s", id)
	}
	if name := config.name; name != fmt.Sprintf("123456789012 : %s", testAccount.Profile) {
		t.Errorf("invalid account name: %s", name)
	}
}
