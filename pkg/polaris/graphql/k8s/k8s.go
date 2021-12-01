//go:generate go run ../queries_gen.go gcp

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

// Package k8s provides a low level interface to the K8s GraphQL queries
// provided by the Polaris platform.
package k8s

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the Polaris K8s API.
type API struct {
	Version string
	GQL     *graphql.Client
}

// Wrap the GraphQL client in the GCP API.
func Wrap(gql *graphql.Client) API {
	return API{Version: gql.Version, GQL: gql}
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client) API {
	return API{Version: gql.Version, GQL: gql}
}

type Filter struct {
	Field string   `json:"field"`
	Texts []string `json:"texts"`
}

type K8sNamespace struct {
	ID                  uuid.UUID      `json:"id"`
	K8sClusterID        uuid.UUID      `json:"k8sClusterID"`
	NamespaceName       string         `json:"namespaceName"`
	IsRelic             bool           `json:"isRelic"`
	ConfiguredSLADomain core.SLADomain `json:"configuredSlaDomain"`
	EffectiveSLADomain  core.SLADomain `json:"effectiveSlaDomain"`
}

func (a API) ListSLA(ctx context.Context) ([]core.GlobalSLA, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/k8s.ListSLA")

	slaDomains := make([]core.GlobalSLA, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			listSlaQuery,
			struct {
				After  string                      `json:"after,omitempty"`
				Filter []core.GlobalSLAFilterInput `json:"filter"`
			}{
				After: cursor,
				Filter: []core.GlobalSLAFilterInput{
					{
						Field: core.ObjectType,
						Text:  "",
						ObjectTypeList: []core.SLAObjectType{
							core.KuprObjectType,
						},
					},
				},
			},
		)
		if err != nil {
			return nil, err
		}

		a.GQL.Log().Printf(log.Debug, "globalSlaConnection(): %s", string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node core.GlobalSLA `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"globalSlaConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, edge := range payload.Data.Query.Edges {
			slaDomains = append(slaDomains, edge.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}
	return slaDomains, nil
}

func (a API) ListK8sNamespace(ctx context.Context, clusterID uuid.UUID) ([]K8sNamespace, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/k8s.GetKuprNamespaces")

	namespaces := make([]K8sNamespace, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			getNamespacesQuery,
			struct {
				After     string    `json:"after,omitempty"`
				Filter    []Filter  `json:"filter,omitempty"`
				ClusterID uuid.UUID `json:"clusterID"`
			}{
				After:     cursor,
				Filter:    []Filter{},
				ClusterID: clusterID,
			},
		)
		if err != nil {
			return nil, err
		}

		a.GQL.Log().Printf(log.Debug, "GetNamespaces(%s): %s", clusterID.String(), string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node K8sNamespace `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"k8sNamespaces"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, edge := range payload.Data.Query.Edges {
			namespaces = append(namespaces, edge.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}
	return namespaces, nil
}
