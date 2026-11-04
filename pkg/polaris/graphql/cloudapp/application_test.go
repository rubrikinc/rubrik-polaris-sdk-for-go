// Copyright 2026 Rubrik, Inc.
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

package cloudapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// TestCreateAWSApplicationRequest verifies the shape of the create request:
// the region as an AwsCommonRegion enum value, every service flag present
// since RSC replaces the whole service configuration, and empty lists sent as
// [] rather than null.
func TestCreateAWSApplicationRequest(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	appID := uuid.MustParse("c7ec46ba-d419-4a9d-88e2-26665e1479db")
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}

		want := map[string]string{
			"cloudNativeAccountId":       `"22bf439e-9922-486e-a32c-ae9009987924"`,
			"targetCloudNativeAccountId": `"ce09d0e4-d0a5-4201-9db2-97f0b6644514"`,
			"name":                       `"app"`,
			"awsRegion":                  `"US_EAST_1"`,
			"matchingTags":               `[{"key":"Version","value":"1.0"}]`,
			"resourceInclusionList":      `[]`,
			"resourceExclusionList":      `[]`,
			"serviceConfigured": `{"applicationCloudConfig":{"awsConfig":{"ec2Enabled":true,"ecrEnabled":false,` +
				`"isEcsEnabled":false,"eksEnabled":false,"elasticacheEnabled":false,"elbEnabled":false,` +
				`"lambdaEnabled":true,"rdsEnabled":false,"s3Enabled":false,"secretsEnabled":false}}}`,
		}
		for name, value := range want {
			if got := string(body.Variables[name]); got != value {
				t.Errorf("invalid variable %s: %s", name, got)
			}
		}

		if _, err := w.Write([]byte(`{"data":{"result":{"application":{"id":"` + appID.String() + `"}}}}`)); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	id, err := CreateAWSApplication(ctx, graphql.NewTestClient(srv), CreateAWSApplicationParams{
		CloudAccountID:       uuid.MustParse("22bf439e-9922-486e-a32c-ae9009987924"),
		TargetCloudAccountID: uuid.MustParse("ce09d0e4-d0a5-4201-9db2-97f0b6644514"),
		Name:                 "app",
		Region:               aws.RegionUsEast1.ToRegionEnum(),
		MatchingTags:         []core.Tag{{Key: "Version", Value: "1.0"}},
		ServiceConfig:        NewServiceConfig(AWSServices{EC2: true, Lambda: true}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != appID {
		t.Errorf("invalid application ID: %s", id)
	}
}

// TestApplicationByID verifies that an application is read from a captured
// RSC response, and that an unknown ID, which RSC rejects with a 403 error,
// is reported as not found.
func TestApplicationByID(t *testing.T) {
	appID := uuid.MustParse("c7ec46ba-d419-4a9d-88e2-26665e1479db")

	tests := []struct {
		name     string
		response string
		notFound bool
	}{
		{name: "Live", response: "testdata/cloud_native_application_response.json"},
		{name: "Unknown", notFound: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				if test.response == "" {
					_, err := w.Write([]byte(`{"data":{"result":null,"hierarchy":null},"errors":[{"message":` +
						`"Objects are not authorized.","path":["result"],"extensions":{"code":403}}]}`))
					if err != nil {
						cancel(err)
					}
					return
				}
				buf, err := os.ReadFile(test.response)
				if err != nil {
					cancel(err)
					return
				}
				if _, err := w.Write(buf); err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			app, err := ApplicationByID(ctx, graphql.NewTestClient(srv), appID)
			if test.notFound {
				if !errors.Is(err, graphql.ErrNotFound) {
					t.Fatalf("expected graphql.ErrNotFound, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if app.ID != appID || app.Name != "my-application" {
				t.Errorf("invalid application: %s %q", app.ID, app.Name)
			}
			if app.Region.Region.AWSRegion.Region != aws.RegionUsEast1 {
				t.Errorf("invalid region: %s", app.Region.Region.AWSRegion.Region.Name())
			}
			if !slices.Equal(app.MatchingTags.Tags, []core.Tag{{Key: "Application", Value: "web"}}) {
				t.Errorf("invalid matching tags: %v", app.MatchingTags.Tags)
			}
			if !slices.Equal(app.ResourceARNs, []string{"arn:aws:lambda:us-east-1:123456789012:function:my-function"}) {
				t.Errorf("invalid resource ARNs: %v", app.ResourceARNs)
			}
			if services := app.ServiceConfig.CloudConfig.AWS; !services.Lambda || !services.ECS || !services.Secrets {
				t.Errorf("invalid services: %+v", services)
			}
			if app.RefreshStatus != RefreshStatusSuccess || app.LastRefreshAt == nil {
				t.Errorf("invalid refresh status: %s %v", app.RefreshStatus, app.LastRefreshAt)
			}
		})
	}
}

// TestHasApplicationSnapshot verifies that the error RSC returns for an application
// without config snapshots is reported as no snapshot, and that other errors
// are returned.
func TestHasApplicationSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     bool
		wantErr  bool
	}{
		{name: "Snapshot", response: `{"data":{"result":{"configSnapshot":{"snapshotFid":` +
			`"4bd3c4ae-4c2b-4b7a-a1b4-50f6ce0b5c8d","snapshotDate":"2026-11-03T20:00:00.000Z"}}}}`, want: true},
		{name: "NoSnapshot", response: `{"data":null,"errors":[{"message":"NOT_FOUND: no config snapshot found for ` +
			`application c7ec46ba-d419-4a9d-88e2-26665e1479db in the requested time window","extensions":{"code":404}}]}`},
		{name: "OtherError", response: `{"data":null,"errors":[{"message":"INTERNAL: boom","extensions":{"code":500}}]}`,
			wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				if _, err := w.Write([]byte(test.response)); err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			got, err := HasApplicationSnapshot(ctx, graphql.NewTestClient(srv), uuid.New())
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("HasApplicationSnapshot() = %t, want %t", got, test.want)
			}
		})
	}
}
