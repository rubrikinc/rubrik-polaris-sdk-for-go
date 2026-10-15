---
name: rx-schema
description: Search the GraphQL schema for types, fields, or enums matching a keyword. Use when exploring the API or
  looking up type definitions.
user_invocable: true
arguments:
  - name: keyword
    description: The keyword to search for in the schema (e.g., a type name, field name, or enum value).
argument-hint: "<keyword>"
---

Search the GraphQL schema for types, fields, or enums matching a keyword. Requires a Polaris service account to be
configured.

## Download Schema

Check if `$TMPDIR/polaris-schema.graphql` exists and is less than 4 hours old. If so, skip the download and use the
cached file. Otherwise, run `go run ./cmd/schema -output $TMPDIR/polaris-schema.graphql` to download a fresh copy.
If the command fails, report the error and stop.

## Search Schema

If no keyword was provided, ask the user what to search for before proceeding.

Search `$TMPDIR/polaris-schema.graphql` for the keyword (case-insensitive). Present the matching type definitions,
fields, or enum values with enough surrounding context to understand the full type they belong to.
