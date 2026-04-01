---
name: rx-queries
description: Run GraphQL code generation for a set of domain packages. Use after editing any `*.graphql` or `*.fragment` files.
user_invocable: true
arguments:
  - name: domain
    description: Optional domain package name (e.g., aws, access, archival, exocompute). Defaults to all domain packages
      with modified `*.graphql` or `*.fragment` files.
argument-hint: "[domain]"
---

Run GraphQL code generation for a set of domain packages. If the `domain` argument is provided, run code generation for
that domain package. Otherwise, determine which domain packages have modified `*.graphql` or `*.fragment` files and run
code generation for those.

## Determine Modified Domain Packages

Run `git diff --name-only HEAD` filtered to `*.graphql` and `*.fragment` to determine which domain packages need to be
regenerated. If no files are found and no domain argument was provided, report "nothing to regenerate" and stop.

## Run Code Generation for a Domain Package

1. Validate that `pkg/polaris/graphql/$domain/` exists. If not, list available domains and stop.
2. Run `go generate ./pkg/polaris/graphql/$domain/...` to update `queries.go`.
3. Run `gofmt -w pkg/polaris/graphql/$domain/queries.go` to ensure correct formatting.
4. Run `go vet ./pkg/polaris/graphql/$domain/...` to catch any errors.
5. Run `git diff pkg/polaris/graphql/$domain/queries.go` to show what changed.

## Report Results

Report the results of code generation for each domain package as table. If no domain package was provided, report the
results for all domain packages that had modified `*.graphql` or `*.fragment` files.
