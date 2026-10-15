---
name: new-domain
description: Scaffold a new GraphQL domain package under pkg/polaris/graphql/ and its high-level API wrapper under pkg/polaris/.
user_invocable: true
arguments:
  - name: domain
    description: Domain package name in snake_case (e.g. archival, cloud_native, sla). Becomes the directory name and Go package name.
  - name: description
    description: One-line description of what this domain wraps (e.g. "SLA domain management").
argument-hint: "<domain> <description>"
---

Scaffold a new domain package. Follow every step in order.

## 0. Decide which layers to create

- **GraphQL layer** (`pkg/polaris/graphql/<domain>/`) — always required. Thin wrappers over RSC GraphQL queries.
- **High-level layer** (`pkg/polaris/<domain>/`) — create only if the user asks for user-facing SDK methods on top of the raw GQL layer. Ask if unclear.

## 1. Create the GraphQL layer

### 1a. Top-level file: `pkg/polaris/graphql/<domain>/<domain>.go`

**Critical:** the `//go:generate` directive MUST be the very first line of the file — before the copyright header.

```go
//go:generate go run ../queries_gen.go <domain>

// Copyright <YEAR> Rubrik, Inc.
// [MIT license header — copy from any adjacent domain file verbatim]

// Package <domain> provides a low-level interface to the <description>
// GraphQL queries provided by the Polaris platform.
package <domain>

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API is the low-level interface to the <description> GraphQL queries.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap returns an API instance wrapping the given GraphQL client.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}
```

### 1b. Feature files: `pkg/polaris/graphql/<domain>/<feature>.go`

One file per logical feature group. Use `snake_case` names matching the GraphQL operation they wrap.

```go
// Copyright <YEAR> Rubrik, Inc.
// [MIT license header]

package <domain>

import (
	"context"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// <Type> represents ...
type <Type> struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// <Method> returns ...
func (a API) <Method>(ctx context.Context, ...) (<Type>, error) {
	buf, err := a.GQL.Request(ctx, queries.<QueryVar>, struct {
		ID string `json:"id"`
	}{ID: id.String()})
	if err != nil {
		return <Type>{}, fmt.Errorf("failed to ...: %w", err)
	}
	var payload struct {
		Data <Type> `json:"result"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return <Type>{}, fmt.Errorf("failed to unmarshal ...: %w", err)
	}
	return payload.Data, nil
}
```

### 1c. Create `pkg/polaris/graphql/<domain>/queries/`

Add one `.graphql` file per operation. Rules:
- Operation name MUST be `RubrikPolarisSDKRequest`
- Result MUST be aliased to `result:`
- Extrapolate input fields as individual `$variables` — never pass complex input types
- File names: `snake_case.graphql`

```graphql
query RubrikPolarisSDKRequest($id: UUID!) {
  result: <gqlFieldName>(id: $id) {
    id
    name
  }
}
```

Add `.fragment` files only if a fragment is shared across 2+ queries in this domain.

### 1d. Run code generation

```
/rx-queries <domain>
```

This creates `pkg/polaris/graphql/<domain>/queries.go`. Never edit it manually.

## 2. Import safety — circular dependency rules

**Safe imports for a new domain package:**

| Import path | Notes |
|---|---|
| `pkg/polaris/graphql` | Always safe (Client; imports no domains) |
| `pkg/polaris/graphql/core` | Safe — but see warning below |
| `pkg/polaris/graphql/core/secret` | Always safe |
| `pkg/polaris/graphql/hierarchy` | Safe |
| `pkg/polaris/graphql/regions/aws` | Safe |
| `pkg/polaris/graphql/regions/azure` | Safe |
| `pkg/polaris/graphql/regions/gcp` | Safe |
| `pkg/polaris/log` | Always safe |
| Sibling domain (e.g. `graphql/sla`) | Safe if that domain does not import yours back |

**Warning — `core` circular dep risk:**

`core` imports `hierarchy` and `sla` (for deprecated type aliases). If `core` ever needs to alias a type from YOUR new domain, importing `core` from your domain creates a cycle. Specifically:

- If your domain defines types that represent **SLA assignments, hierarchy objects, or shared platform concepts** — do NOT import `core` from that domain. Put shared types in `core` instead, and have your domain import nothing from `core`.
- If your domain is narrow and domain-specific (cloud accounts, snapshots, etc.) — importing `core` is safe.

When unsure: check whether `core/` has any comment referencing your domain's concepts before adding the import.

## 3. Create the high-level layer (if requested)

### `pkg/polaris/<domain>/<domain>.go`

```go
// Copyright <YEAR> Rubrik, Inc.
// [MIT license header]

// Package <domain> provides a high-level interface to the <description>.
package <domain>

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	gql<Domain> "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/<domain>"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API is the high-level interface to the <description>.
type API struct {
	client *polaris.Client
	log    log.Logger
}

// Wrap returns an API instance wrapping the given Polaris client.
func Wrap(client *polaris.Client) API {
	return API{client: client, log: client.GQL.Log()}
}
```

Feature files in this layer call into `gql<Domain>.Wrap(a.client.GQL).<Method>(...)`.

## 4. Add tests

Create `pkg/polaris/graphql/<domain>/<feature>_test.go`. Use `internal/handler/` for mock HTTP — not `testify/mock` or `httptest` directly.

```go
package <domain>

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func Test<Feature>(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		// Read req.Body to verify the query sent, write JSON response.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"result":{"id":"...","name":"..."}}}`)
	}))
	defer srv.Close()

	result, err := Wrap(graphql.NewTestClient(srv)).<Method>(ctx, ...)
	if err != nil {
		t.Fatal(err)
	}
	// assert result fields
}
```

Complex fixture responses go in `testdata/<feature>.json` and are read with `os.ReadFile`.

## 5. Verify

Run `/rx-checks <domain-path>` where `<domain-path>` is e.g. `./pkg/polaris/graphql/<domain>/...`.

## 6. Summary of files created

| File | Purpose |
|---|---|
| `pkg/polaris/graphql/<domain>/<domain>.go` | `go:generate`, `API` struct, `Wrap` |
| `pkg/polaris/graphql/<domain>/queries/<op>.graphql` | GraphQL source (one per operation) |
| `pkg/polaris/graphql/<domain>/queries.go` | Generated — never edit |
| `pkg/polaris/graphql/<domain>/<feature>.go` | Types + methods per feature group |
| `pkg/polaris/graphql/<domain>/<feature>_test.go` | Unit tests using handler/ |
| `pkg/polaris/<domain>/<domain>.go` | High-level `API` + `Wrap` (if needed) |
| `pkg/polaris/<domain>/<feature>.go` | High-level methods (if needed) |
