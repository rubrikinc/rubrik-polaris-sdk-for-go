---
name: rx-queries
description: Run GraphQL code generation for a set of domain packages. Use after editing any `*.graphql` files.
user_invocable: true
arguments:
  - name: domain
    description: Optional domain package name (e.g., aws, access, archival, exocompute). Defaults to all domain packages
      with modified `*.graphql` files.
argument-hint: "[domain]"
---

Run GraphQL code generation for a set of domain packages. If the `domain` argument is provided, run code generation for
that domain package. Otherwise, determine which domain packages have modified `*.graphql` files and run code generation
for those.

## Determine Modified Domain Packages

Run `git diff --name-only --diff-filter=d HEAD` filtered to `*.graphql` to determine which domain packages need to be
regenerated.

## Run Code Generation for a Domain Package

1. Validate that `pkg/polaris/graphql/$domain/` exists. If not, list available domains and stop.
2. Run `go generate ./pkg/polaris/graphql/$domain/...` to update `queries.go`.
3. Run `gofmt -l pkg/polaris/graphql/$domain/queries.go` to verify formatting.
4. Run `git diff pkg/polaris/graphql/$domain/queries.go` to show what changed.
5. Run `go vet ./pkg/polaris/graphql/$domain/...` to catch any errors.
