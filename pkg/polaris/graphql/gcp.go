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

package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

type GcpProject struct {
	Project struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ProjectID     string `json:"projectID"`
		ProjectNumber int64  `json:"projectNumber"`
	} `json:"project"`
	FeatureDetail struct {
		Feature string `json:"feature"`
		Status  string `json:"status"`
	} `json:"featureDetail"`
}

// GcpNativeProject -
type GcpNativeProject struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	NativeName       string `json:"nativeName"`
	NativeID         string `json:"nativeId"`
	ProjectNumber    string `json:"projectNumber"`
	OrganizationName string `json:"organizationName"`
	SlaAssignment    string `json:"slaAssignment"`

	ConfiguredSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"configuredSlaDomain"`

	EffectiveSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSlaDomain"`
}

// GcpCloudAccountAddManualAuthProject adds the native GCP project to Polaris.
func (c *Client) GcpCloudAccountAddManualAuthProject(ctx context.Context, projectName, projectID string, projectNumber int64, orgName, jwtConfig string) error {
	c.log.Print(log.Trace, "graphql.Client.GcpCloudAccountAddManualAuthProject")

	_, err := c.Request(ctx, gcpCloudAccountAddManualAuthProjectQuery, struct {
		Name      string `json:"gcp_native_project_name"`
		ID        string `json:"gcp_native_project_id"`
		Number    int64  `json:"gcp_native_project_number"`
		OrgName   string `json:"organization_name,omitempty"`
		JwtConfig string `json:"service_account_auth_key,omitempty"`
	}{Name: projectName, ID: projectID, Number: projectNumber, OrgName: orgName, JwtConfig: jwtConfig})

	return err
}

// GcpCloudAccountDeleteProjects delete cloud account for the given GCP Project
// UUIDs and feature.
func (c *Client) GcpCloudAccountDeleteProjects(ctx context.Context, protectionIDs, hostProjectIDs, accountProjectIDs []string) error {
	c.log.Print(log.Trace, "graphql.Client.GcpCloudAccountDeleteProjects")

	if protectionIDs == nil {
		protectionIDs = []string{}
	}
	if hostProjectIDs == nil {
		hostProjectIDs = []string{}
	}
	if accountProjectIDs == nil {
		accountProjectIDs = []string{}
	}
	buf, err := c.Request(ctx, gcpCloudAccountDeleteProjectsQuery, struct {
		ProtectionIDs    []string `json:"native_protection_ids"`
		HostProjectIDs   []string `json:"shared_vpc_host_project_ids"`
		AccoutProjectIDs []string `json:"cloud_account_project_ids"`
	}{ProtectionIDs: protectionIDs, HostProjectIDs: hostProjectIDs, AccoutProjectIDs: accountProjectIDs})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "GcpCloudAccountDeleteProjects(%q, %q, %q): %s", protectionIDs, hostProjectIDs, accountProjectIDs, string(buf))

	var payload struct {
		Data struct {
			Projects []struct {
				ProjectUUID string `json:"projectUuid"`
				Success     bool   `json:"success"`
				Error       string `json:"error"`
			} `json:"gcpCloudAccountDeleteProjects"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	for _, project := range payload.Data.Projects {
		if !project.Success {
			return fmt.Errorf("polaris: %s", project.Error)
		}
	}

	return nil
}

// GcpCloudAccountListPermissions list the permissions needed to enable cloud
// native protection.
func (c *Client) GcpCloudAccountListPermissions(ctx context.Context) ([]string, error) {
	c.log.Print(log.Trace, "graphql.Client.GcpCloudAccountListPermissions")

	buf, err := c.Request(ctx, gcpCloudAccountListPermissionsQuery, nil)
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "GcpCloudAccountListPermissions(): %s", string(buf))

	var payload struct {
		Data struct {
			Permissions []struct {
				Permission string `json:"permission"`
			} `json:"gcpCloudAccountListPermissions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	perms := make([]string, 0, len(payload.Data.Permissions))
	for _, perm := range payload.Data.Permissions {
		perms = append(perms, perm.Permission)
	}

	return perms, nil
}

// GcpCloudAccountListProjects returns all GCP projects with cloud native
// protection, searchText can be used to search for project name, native id
// and project number.
func (c *Client) GcpCloudAccountListProjects(ctx context.Context, searchText string, statusFilters []string) ([]GcpProject, error) {
	c.log.Print(log.Trace, "graphql.Client.GcpCloudAccountListProjects")

	if statusFilters == nil {
		statusFilters = []string{}
	}
	buf, err := c.Request(ctx, gcpCloudAccountListProjectsQuery, struct {
		Text    string   `json:"search_text"`
		Filters []string `json:"status_filters"`
	}{Text: searchText, Filters: statusFilters})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "GcpCloudAccountListProjects(%q, %q): %s", searchText, statusFilters, string(buf))

	var payload struct {
		Data struct {
			GcpProject []GcpProject `json:"gcpCloudAccountListProjects"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.GcpProject, nil
}

// GcpNativeDisableProject disables the GCP project. If deleteSnapshots are
// true the snapshots are deleted otherwise they are kept.
func (c *Client) GcpNativeDisableProject(ctx context.Context, polProjectID string, deleteSnapshots bool) (TaskChainUUID, error) {
	c.log.Print(log.Trace, "graphql.Client.GcpNativeDisableProject")

	buf, err := c.Request(ctx, gcpNativeDisableProjectQuery, struct {
		PolProjectID    string `json:"rubrik_project_id"`
		DeleteSnapshots bool   `json:"delete_snapshots"`
	}{PolProjectID: polProjectID, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return "", err
	}

	c.log.Printf(log.Debug, "GcpNativeDisableProject(%q, %t): %s", polProjectID, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				TaskChainID TaskChainUUID `json:"taskchainUuid"`
			} `json:"gcpNativeDisableProject"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}
	if payload.Data.Query.TaskChainID == "" {
		return "", err
	}

	return payload.Data.Query.TaskChainID, nil
}

// GcpNativeProjectConnection returns the native accounts matching the
// specified filter.
func (c *Client) GcpNativeProjectConnection(ctx context.Context, filter string) ([]GcpNativeProject, error) {
	c.log.Print(log.Trace, "graphql.Client.GcpNativeProjectConnection")

	accounts := make([]GcpNativeProject, 0)
	var endCursor string
	for {
		buf, err := c.Request(ctx, gcpNativeProjectConnectionQuery, struct {
			After  string `json:"after,omitempty"`
			Filter string `json:"filter,omitempty"`
		}{After: endCursor, Filter: filter})
		if err != nil {
			return nil, err
		}

		c.log.Printf(log.Debug, "GcpNativeProjectConnection(%q): %s", filter, string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node GcpNativeProject `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"gcpNativeProjectConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}

		for _, account := range payload.Data.Query.Edges {
			accounts = append(accounts, account.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		endCursor = payload.Data.Query.PageInfo.EndCursor
	}

	return accounts, nil
}

// GcpSetDefaultServiceAccount sets the default GCP service account. The set
// service account will be used for GCP projects added without a service
// account key file.
func (c *Client) GcpSetDefaultServiceAccount(ctx context.Context, jwtConfig, accountName string) error {
	c.log.Print(log.Trace, "graphql.Client.GcpSetDefaultServiceAccount")

	buf, err := c.Request(ctx, gcpSetDefaultServiceAccountQuery, struct {
		JwtConfig   string `json:"jwt_config"`
		AccountName string `json:"account_name"`
	}{JwtConfig: jwtConfig, AccountName: accountName})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "GcpSetDefaultServiceAccount(%q, %q): %s", jwtConfig, accountName, string(buf))

	var payload struct {
		Data struct {
			Success bool `json:"gcpSetDefaultServiceAccountJwtConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if !payload.Data.Success {
		return errors.New("polaris: failed to set default service account")
	}

	return nil
}
