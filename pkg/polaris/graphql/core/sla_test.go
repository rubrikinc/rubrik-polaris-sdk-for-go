package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestAssignSLAForSnappableHierarchies(t *testing.T) {

	objIDs := []uuid.UUID{uuid.New()}

	testClient := func(httpResp error, gResp bool) (
		*graphql.Client,
		*http.Server,
	) {
		client, lis := graphql.NewTestClient("john", "doe", log.DiscardLogger{})
		var srv *http.Server
		defer func(srv *http.Server) {
			if srv != nil {
				_ = srv.Shutdown(context.Background())
			}
		}(srv)

		srv = testnet.ServeJSONWithStaticToken(
			lis, func(w http.ResponseWriter, req *http.Request) {
				buf, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				// The request payload should be of the following format.
				// Unmarshal the request payload and check if it matches the expected
				// format.
				var payload struct {
					Query     string `json:"query"`
					Variables struct {
						GlobalSLAOptionalFID            *uuid.UUID                       `json:"globalSlaOptionalFid"`
						GlobalSLAAssignType             SLAAssignType                    `json:"globalSlaAssignType"`
						ObjectIDs                       []uuid.UUID                      `json:"objectIds"`
						ApplicableSnappableTypes        []SnappableLevelHierarchyType    `json:"applicableSnappableTypes"`
						ShouldApplyToExistingSnapshots  *bool                            `json:"shouldApplyToExistingSnapshots"`
						ShouldApplyToNonPolicySnapshots *bool                            `json:"shouldApplyToNonPolicySnapshots"`
						GlobalExistingSnapshotRetention *GlobalExistingSnapshotRetention `json:"globalExistingSnapshotRetention"`
						UserNote                        *string                          `json:"userNote"`
					}
				}
				if err := json.Unmarshal(buf, &payload); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(objIDs, payload.Variables.ObjectIDs) {
					t.Errorf(
						"AssignSLAForSnappableHierarchies() got = %v, want %v",
						payload.Variables.ObjectIDs,
						objIDs,
					)
				}

				if httpResp != nil {
					http.Error(w, httpResp.Error(), 500)
					return
				}
				type data struct {
					Success bool `json:"success"`
				}
				var response struct {
					Data struct {
						AssignSLAForSnappableHierarchies []data `json:"assignSlasForSnappableHierarchies"`
					} `json:"data"`
				}

				response.Data.AssignSLAForSnappableHierarchies = []data{{gResp}}

				ob, err := json.Marshal(response)
				if err != nil {
					t.Fatal(err)
				}

				_, err = w.Write(ob)
				if err != nil {
					t.Fatal(err)
				}
				return
			},
		)
		return client, srv
	}

	tests := []struct {
		name         string
		mockResponse bool
		mockHTTPResp error
		want         []bool
		wantErr      bool
	}{
		{
			name:         "success in assignment",
			mockResponse: true,
			mockHTTPResp: nil,
			want:         []bool{true},
			wantErr:      false,
		},
		{
			name:         "failure in assignment no http error",
			mockResponse: false,
			mockHTTPResp: nil,
			want:         []bool{false},
			wantErr:      false,
		},
		{
			name:         "failure in assignment with http error",
			mockResponse: false,
			mockHTTPResp: errors.New("http error"),
			want:         nil,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				client, srv := testClient(tt.mockHTTPResp, tt.mockResponse)
				defer func(srv *http.Server, ctx context.Context) {
					_ = srv.Shutdown(ctx)
				}(srv, context.Background())

				a := Wrap(client)
				got, err := a.AssignSLAForSnappableHierarchies(
					context.Background(),
					nil,
					ProtectWithSLAID,
					objIDs,
					nil,
					nil,
					nil,
					nil,
					nil,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf(
						"AssignSLAForSnappableHierarchies() error = %v, wantErr %v",
						err,
						tt.wantErr,
					)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf(
						"AssignSLAForSnappableHierarchies() got = %v, want %v",
						got,
						tt.want,
					)
				}
			},
		)
	}
}
