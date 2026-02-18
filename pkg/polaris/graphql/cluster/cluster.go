//go:generate go run ../queries_gen.go cluster

// Copyright 2025 Rubrik, Inc.
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

package cluster

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC Cluster API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Cluster API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// RemovalPrechecks represents the precheck information for cluster removal.
type RemovalPrechecks struct {
	Disconnected       bool   `json:"isDisconnected"`
	IgnorePrecheckTime string `json:"ignorePrecheckTime"`
	LastConnectionTime string `json:"lastConnectionTime"`
	IgnorePrecheck     bool   `json:"canIgnorePrecheck"`
	AirGapped          bool   `json:"isAirGapped"`
}

// CanIgnoreClusterRemovalPrechecks returns whether the cluster removal prechecks can be ignored.
func CanIgnoreClusterRemovalPrechecks(ctx context.Context, gql *graphql.Client, clusterUUID uuid.UUID) (RemovalPrechecks, error) {
	gql.Log().Print(log.Trace)

	query := canIgnoreClusterRemovalPrechecksQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID `json:"clusterUuid"`
	}{ClusterUUID: clusterUUID})
	if err != nil {
		return RemovalPrechecks{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result RemovalPrechecks `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return RemovalPrechecks{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// RCVLocation represents a Recovery Cloud Vault location.
type RCVLocation struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ClusterRCVLocations returns all RCV locations for the specified cluster.
func ClusterRCVLocations(ctx context.Context, gql *graphql.Client, clusterUUID uuid.UUID, pagination core.Pagination) ([]RCVLocation, error) {
	gql.Log().Print(log.Trace)

	query := clusterRcvLocationsQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID       `json:"clusterUuid"`
		First       *int            `json:"first,omitempty"`
		After       *string         `json:"after,omitempty"`
		Last        *int            `json:"last,omitempty"`
		Before      *string         `json:"before,omitempty"`
		SortOrder   *core.SortOrder `json:"sortOrder,omitempty"`
	}{
		ClusterUUID: clusterUUID,
		First:       pagination.First,
		After:       pagination.After,
		Last:        pagination.Last,
		Before:      pagination.Before,
		SortOrder:   pagination.SortOrder,
	})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Edges []struct {
					Cursor string      `json:"cursor"`
					Node   RCVLocation `json:"node"`
				} `json:"edges"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var locations []RCVLocation
	for _, edge := range payload.Data.Result.Edges {
		locations = append(locations, edge.Node)
	}

	return locations, nil
}

// GlobalSLA represents a global SLA configuration.
type GlobalSLA struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AllClusterGlobalSLAs returns all global SLAs for the specified cluster.
func AllClusterGlobalSLAs(ctx context.Context, gql *graphql.Client, clusterUUID uuid.UUID) ([]GlobalSLA, error) {
	gql.Log().Print(log.Trace)

	query := allClusterGlobalSlasQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID `json:"clusterUuid"`
	}{ClusterUUID: clusterUUID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []GlobalSLA `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// SLAReplicationInfo represents SLA replication information for a cluster.
type SLAReplicationInfo struct {
	IsActiveSLA bool `json:"isActiveSla"`
}

// VerifySLAWithReplicationToCluster verifies if there are active SLAs with replication to the specified cluster.
func VerifySLAWithReplicationToCluster(ctx context.Context, gql *graphql.Client, clusterUUID uuid.UUID, includeArchived bool) (SLAReplicationInfo, error) {
	gql.Log().Print(log.Trace)

	query := verifySlaWithReplicationToClusterQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID     uuid.UUID `json:"clusterUuid"`
		IncludeArchived bool      `json:"includeArchived"`
	}{ClusterUUID: clusterUUID, IncludeArchived: includeArchived})
	if err != nil {
		return SLAReplicationInfo{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result SLAReplicationInfo `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return SLAReplicationInfo{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// RemoveCDMCluster removes the specified CDM cluster.
func RemoveCDMCluster(ctx context.Context, gql *graphql.Client, clusterUUID uuid.UUID, expireInDays int, isForce bool) (bool, error) {
	gql.Log().Print(log.Trace)

	query := removeCdmClusterQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID  uuid.UUID `json:"clusterUuid"`
		ExpireInDays int       `json:"expireInDays"`
		IsForce      bool      `json:"isForce"`
	}{ClusterUUID: clusterUUID, ExpireInDays: expireInDays, IsForce: isForce})
	if err != nil {
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// SLADataLocationCluster represents a cluster in the SLA data location.
type SLADataLocationCluster struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	ClusterInfo struct {
		IsConnected bool `json:"isConnected"`
	} `json:"clusterInfo"`
}

// ClusterFilter represents a filter for SLA source clusters.
type ClusterFilter struct {
	Field  string   `json:"field"`
	Values []string `json:"texts"`
}

// SLASourceClusters returns all SLA source clusters matching the specified filters.
func SLASourceClusters(ctx context.Context, gql *graphql.Client, filters []ClusterFilter) ([]SLADataLocationCluster, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var clusters []SLADataLocationCluster
	for {
		query := slaSourceClustersQuery
		buf, err := gql.Request(ctx, query, struct {
			First               int             `json:"first,omitempty"`
			After               string          `json:"after,omitempty"`
			SortBy              string          `json:"sortBy,omitempty"`
			SortOrder           core.SortOrder  `json:"sortOrder,omitempty"`
			Filter              []ClusterFilter `json:"filter"`
			IsArchivalSelected  bool            `json:"isArchivalSelected"`
			SelectedReplication string          `json:"selectedReplication,omitempty"`
			SLAID               *string         `json:"slaId,omitempty"`
		}{First: 50, After: cursor, SortBy: "CLUSTER_NAME", SortOrder: core.SortOrderAsc, Filter: filters, IsArchivalSelected: false, SelectedReplication: "ONE_TO_ONE", SLAID: nil})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Cursor string `json:"cursor"`
						Node   struct {
							DisableReasons      []string               `json:"disableReasons"`
							HasProtectedObjects bool                   `json:"hasProtectedObjects"`
							Cluster             SLADataLocationCluster `json:"cluster"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}

		for _, edge := range payload.Data.Result.Edges {
			clusters = append(clusters, edge.Node.Cluster)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return clusters, nil
}

// DNSServersAndDomains represents the cluster DNS servers and search domains.
type DNSServersAndDomains struct {
	Servers []string `json:"servers"`
	Domains []string `json:"domains"`
}

// DNSServers returns the cluster DNS servers.
func (a API) DNSServers(ctx context.Context, clusterID uuid.UUID) (DNSServersAndDomains, error) {
	a.log.Print(log.Trace)

	query := clusterDnsServersQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID uuid.UUID `json:"clusterUuid"`
	}{ClusterID: clusterID})

	if err != nil {
		return DNSServersAndDomains{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result DNSServersAndDomains `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return DNSServersAndDomains{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// NTPSymmetricKey represents the NTP server symmetric key.
type NTPSymmetricKey struct {
	Key     string `json:"key"`
	KeyId   string `json:"keyId"`
	KeyType string `json:"keyType"`
}

// NTPServersAndKeys represents the cluster NTP servers and symmetric keys.
type NTPServersAndKeys struct {
	Server       string          `json:"server"`
	SymmetricKey NTPSymmetricKey `json:"symmetricKey,omitempty"`
}

// NTPServers returns the cluster NTP servers.
func (a API) NTPServers(ctx context.Context, clusterID uuid.UUID) ([]NTPServersAndKeys, error) {
	a.log.Print(log.Trace)

	query := clusterNtpServersQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID string `json:"id"`
	}{ClusterID: clusterID.String()})

	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Data []NTPServersAndKeys `json:"data"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Data, nil
}

// IPMIInfo represents the cluster IPMI information.
type IPMIInfo struct {
	IsAvailable bool `json:"isAvailable"`
	UsesHTTPS   bool `json:"usesHttps"`
	UsesIKVM    bool `json:"usesIkvm"`
}

// Settings represents the cluster settings.
type Settings struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Status      Status    `json:"status"`
	GeoLocation string    `json:"geoLocation"`
	Timezone    string    `json:"timezone"`
	IPMIInfo    IPMIInfo  `json:"ipmiInfo,omitempty"`
}

// ClusterSettings returns the cloud cluster settings.
func (a API) ClusterSettings(ctx context.Context, clusterID uuid.UUID) (Settings, error) {
	a.log.Print(log.Trace)

	query := clusterSettingsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID uuid.UUID `json:"id"`
	}{ClusterID: clusterID})

	if err != nil {
		return Settings{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result Settings `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Settings{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// UpdateClusterNTPServersInput represents the input for the UpdateNTPServers mutation.
type UpdateClusterNTPServersInput struct {
	ClusterID uuid.UUID `json:"id"`
	Servers   []struct {
		Server       string          `json:"server"`
		SymmetricKey NTPSymmetricKey `json:"symmetricKey,omitempty"`
	} `json:"ntpServerConfigs"`
}

// UpdateNTPServers updates the cloud cluster NTP servers.
func (a API) UpdateNTPServers(ctx context.Context, input UpdateClusterNTPServersInput) error {
	a.log.Print(log.Trace)

	query := updateClusterNtpServersQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input UpdateClusterNTPServersInput `json:"input"`
	}{
		Input: input,
	})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, errors.New("failed to update NTP servers"))
	}

	return nil
}

// UpdateDNSServersAndSearchDomainsInput represents the input for the updateClusterDnsServersAndSearchDomains mutation.
type UpdateDNSServersAndSearchDomainsInput struct {
	ClusterID     uuid.UUID `json:"id"`
	DNSServers    []string  `json:"servers"`
	SearchDomains []string  `json:"domains"`
}

// UpdateDNSServersAndSearchDomains updates the cluster DNS servers and search domains.
func (a API) UpdateDNSServersAndSearchDomains(ctx context.Context, input UpdateDNSServersAndSearchDomainsInput) error {
	a.log.Print(log.Trace)

	query := updateCusterDnsAndSearchDomainsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID     uuid.UUID `json:"id"`
		DNSServers    []string  `json:"servers"`
		SearchDomains []string  `json:"domains"`
	}{
		ClusterID:     input.ClusterID,
		DNSServers:    input.DNSServers,
		SearchDomains: input.SearchDomains,
	})

	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, errors.New("failed to update DNS servers and search domains"))
	}

	return nil
}
