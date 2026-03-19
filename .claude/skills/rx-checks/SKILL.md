---
name: rx-checks
description: Run the local CI check suite for the Go SDK. Use before pushing changes.
user_invocable: true
arguments:
  - name: package
    description: Optional Go package path to check (e.g., `./pkg/polaris/aws/...`). Defaults to `./...`.
argument-hint: "[package]"
---

Run the local CI check suite for the Go SDK. If the `package` argument is not provided, it defaults to `./...`.

## Ensure Dependencies are Installed

Ensure that the `staticcheck` tool is installed. If the tool is not installed it can be installed with
`go install honnef.co/go/tools/cmd/staticcheck@latest`.

## Run Checks

1. Run `gofmt -l` on changed `.go` files (use `git diff --name-only --diff-filter=d HEAD` filtered to `*.go`) and report
any unformatted files.
2. Run `go vet $package` and report pass/fail.
3. Run `staticcheck $package` and report pass/fail.
4. Run `go test $package` and report pass/fail.

## Report Results

Report the results of each check as a table. If any check failed, report the command output.
