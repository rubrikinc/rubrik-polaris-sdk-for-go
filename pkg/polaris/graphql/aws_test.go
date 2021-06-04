package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"text/template"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestAwsCloudAccounts(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/aws_cloudaccounts.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	srv := serveJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				ColumnFilter string `json:"column_filter,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			t.Fatal(err)
		}

		err = tmpl.Execute(w, struct {
			NativeID string
		}{NativeID: payload.Variables.ColumnFilter})
		if err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	accounts, err := client.AwsCloudAccounts(context.Background(), "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(accounts); n != 1 {
		t.Fatalf("invalid number of accounts: %v", n)
	}
	if accounts[0].AwsCloudAccount.AccountName != "My-Account-Name" {
		t.Fatalf("invalid name: %v", accounts[0].AwsCloudAccount.AccountName)
	}
	if accounts[0].AwsCloudAccount.NativeID != "123456789012" {
		t.Fatalf("invalid native id: %v", accounts[0].AwsCloudAccount.NativeID)
	}
	if n := len(accounts[0].FeatureDetails); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if accounts[0].FeatureDetails[0].Feature != "CLOUD_NATIVE_PROTECTION" {
		t.Fatalf("invalid feature name: %v", accounts[0].FeatureDetails[0].Feature)
	}
	if regions := accounts[0].FeatureDetails[0].AwsRegions; !reflect.DeepEqual(regions, []string{"US_EAST_2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if accounts[0].FeatureDetails[0].Status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", accounts[0].FeatureDetails[0].Status)
	}
}

func TestAwsStartNativeAccountDisableJob(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/aws_startnativeaccountdisablejob.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	srv := serveJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				AccountID         string `json:"aws_account_rubrik_id"`
				ProtectionFeature string `json:"aws_native_protection_feature"`
				DeleteSnapshots   bool   `json:"delete_native_snapshots"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			t.Fatal(err)
		}

		var msg string
		if payload.Variables.AccountID != "123456789012" {
			msg = fmt.Sprintf("invalid account id: %s", payload.Variables.AccountID)
		}
		if payload.Variables.ProtectionFeature != "EC2" {
			msg = fmt.Sprintf("invalid protection feature: %s", payload.Variables.ProtectionFeature)
		}
		if payload.Variables.DeleteSnapshots != true {
			msg = fmt.Sprintf("invalid delete snapshots: %t", payload.Variables.DeleteSnapshots)
		}

		err = tmpl.Execute(w, struct {
			JobID string
			Error string
		}{JobID: "b7945260-e39f-40bb-9dad-c20d93129686", Error: msg})
		if err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	jobID, err := client.AwsStartNativeAccountDisableJob(context.Background(), "123456789012", AwsEC2, true)
	if err != nil {
		t.Fatal(err)
	}
	if jobID != "b7945260-e39f-40bb-9dad-c20d93129686" {
		t.Fatal("")
	}

	_, err = client.AwsStartNativeAccountDisableJob(context.Background(), "fail", AwsEC2, true)
	if err == nil {
		t.Fatal("request should fail")
	}
}
