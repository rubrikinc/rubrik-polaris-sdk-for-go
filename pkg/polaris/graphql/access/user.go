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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// User represents a user in RSC.
type User struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	IsAccountOwner bool   `json:"isAccountOwner"`
	Roles          []Role `json:"roles"`
}

// UsersInCurrentAndDescendantOrganization returns the users matching the
// specified email address filter.
func (a API) UsersInCurrentAndDescendantOrganization(ctx context.Context, emailFilter string) ([]User, error) {
	a.GQL.Log().Print(log.Trace)

	var users []User
	var cursor string
	for {
		buf, err := a.GQL.Request(ctx, usersInCurrentAndDescendantOrganizationQuery, struct {
			After       string `json:"after,omitempty"`
			EmailFilter string `json:"emailFilter,omitempty"`
		}{After: cursor, EmailFilter: emailFilter})
		if err != nil {
			return nil, fmt.Errorf("failed to request UsersInCurrentAndDescendantOrganization: %w", err)
		}
		a.GQL.Log().Printf(log.Debug, "usersInCurrentAndDescendantOrganizationQuery(%q): %s", emailFilter, string(buf))

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Node User `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UsersInCurrentAndDescendantOrganization response: %w", err)
		}
		for _, account := range payload.Data.Result.Edges {
			users = append(users, account.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return users, nil
}

// CreateUser creates a new user with the specified email address and roles.
func (a API) CreateUser(ctx context.Context, userEmail string, roleIDs []uuid.UUID) (string, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, createUserQuery, struct {
		Email   string      `json:"email"`
		RoleIDs []uuid.UUID `json:"roleIds"`
	}{Email: userEmail, RoleIDs: roleIDs})
	if err != nil {
		return "", fmt.Errorf("failed to request CreateUser: %w", err)
	}
	a.GQL.Log().Printf(log.Debug, "createUser(%q, %v): %s", userEmail, roleIDs, string(buf))

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal CreateUser response: %w", err)
	}

	return payload.Data.Result, nil
}

// DeleteUserFromAccount deletes the users with the specified email addresses.
func (a API) DeleteUserFromAccount(ctx context.Context, ids []string) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteUserFromAccountQuery, struct {
		IDs []string `json:"ids"`
	}{IDs: ids})
	if err != nil {
		return fmt.Errorf("failed to request DeleteUserFromAccount: %w", err)
	}
	a.GQL.Log().Printf(log.Debug, "deleteUserFromAccount(%v): %s", ids, string(buf))

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal DeleteUserFromAccount response: %w", err)
	}

	return nil
}
