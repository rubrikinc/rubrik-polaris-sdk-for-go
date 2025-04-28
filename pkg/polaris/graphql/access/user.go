// Copyright 2023 Rubrik, Inc.
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

package access

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UserDomain represents the domain in RSC in which a user exists.
type UserDomain string

const (
	DomainLocal UserDomain = "LOCAL"
	DomainSSO   UserDomain = "SSO"
)

// UserRef is a reference to a User in RSC. A UserRef holds the ID and Email
// of a user.
type UserRef struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// User represents a user in RSC.
type User struct {
	UserRef
	Domain         UserDomain `json:"domain"`
	Status         string     `json:"status"`
	IsAccountOwner bool       `json:"isAccountOwner"`
	Groups         []string   `json:"groups"`
	Roles          []RoleRef  `json:"roles"`
}

// HasRole returns true if the user has the specified role, false otherwise.
func (user User) HasRole(roleID uuid.UUID) bool {
	for _, role := range user.Roles {
		if role.ID == roleID {
			return true
		}
	}

	return false
}

// UserFilter holds the filter parameters for a user list operation.
type UserFilter struct {
	AuthDomainIDs []string     `json:"authDomainIdsFilter,omitempty"`
	UserDomains   []UserDomain `json:"domainFilter,omitempty"`
	EmailFilter   string       `json:"emailFilter,omitempty"`
}

// ListUsers returns all users matching the specified user filter.
func ListUsers(ctx context.Context, gql *graphql.Client, filter UserFilter) ([]User, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []User
	for {
		query := usersInCurrentAndDescendantOrganizationQuery
		buf, err := gql.Request(ctx, query, struct {
			After  string     `json:"after,omitempty"`
			Filter UserFilter `json:"filter"`
		}{After: cursor, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}
		graphql.LogResponse(gql.Log(), query, buf)

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []User `json:"nodes"`
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
		nodes = append(nodes, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return nodes, nil
}

// CreateUserParams holds the parameters for a user create operation.
type CreateUserParams struct {
	Email   string      `json:"email"`
	RoleIDs []uuid.UUID `json:"roleIds"`
}

// CreateUser creates a new user. Returns the ID of the new user.
func CreateUser(ctx context.Context, gql *graphql.Client, params CreateUserParams) (string, error) {
	gql.Log().Print(log.Trace)

	query := createUserQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return "", graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result == "" {
		return "", graphql.ResponseError(query, fmt.Errorf("failed to create user %q", params.Email))
	}

	return payload.Data.Result, nil
}

// DeleteUser deletes the user with the specified ID.
func DeleteUser(ctx context.Context, gql *graphql.Client, userID string) error {
	gql.Log().Print(log.Trace)

	query := deleteUserFromAccountQuery
	buf, err := gql.Request(ctx, query, struct {
		IDs []string `json:"ids"`
	}{IDs: []string{userID}})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, fmt.Errorf("failed to delete user %q", userID))
	}

	return nil
}
