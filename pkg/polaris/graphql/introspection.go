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

package graphql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// EnumValue represents a single value in a GraphQL enum type, as returned by
// schema introspection.
type EnumValue struct {
	Deprecated        bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
	Description       string `json:"description"`
	Name              string `json:"name"`
}

// EnumValues returns the values of the named GraphQL enum type as a slice.
func EnumValues(ctx context.Context, client *Client, enumName string) ([]EnumValue, error) {
	client.Log().Print(log.Trace)

	enumValuesQuery := `query SdkGolangOperationEnum($enumName: String!) {
		result: __type(name: $enumName) {
			name
			kind
			enumValues {
				name
				description
				isDeprecated
				deprecationReason
			}
		}
	}`
	buf, err := client.Request(ctx, enumValuesQuery, struct {
		EnumName string `json:"enumName"`
	}{EnumName: enumName})
	if err != nil {
		return nil, RequestError(enumValuesQuery, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				EnumValues []EnumValue `json:"enumValues"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, UnmarshalError(enumValuesQuery, err)
	}
	if payload.Data.Result.EnumValues == nil {
		return nil, fmt.Errorf("%s enum %w", enumName, ErrNotFound)
	}

	return payload.Data.Result.EnumValues, nil
}

// EnumValuesAsSet returns the values of the named GraphQL enum type as a map
// keyed by name.
func EnumValuesAsSet(ctx context.Context, client *Client, enumName string) (map[string]EnumValue, error) {
	client.Log().Print(log.Trace)

	values, err := EnumValues(ctx, client, enumName)
	if err != nil {
		return nil, err
	}

	set := make(map[string]EnumValue, len(values))
	for _, v := range values {
		set[v.Name] = v
	}

	return set, nil
}
