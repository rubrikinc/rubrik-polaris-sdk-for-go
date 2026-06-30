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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const schemaQuery = `
	query SchemaQuery {
		__schema {
			queryType {
				name
			}
			mutationType {
				name
			}
			subscriptionType {
				name
			}
			types {
				...FullType
			}
			directives {
				name
				description
				locations
				args {
					...InputValue
				}
			}
		}
	}

	fragment FullType on __Type {
		kind
		name
		description
		fields(includeDeprecated: true) {
			name
			description
			args {
				...InputValue
			}
			type {
				...TypeRef
			}
			isDeprecated
			deprecationReason
		}
		inputFields {
			...InputValue
		}
		interfaces {
			...TypeRef
		}
		enumValues(includeDeprecated: true) {
			name
			description
			isDeprecated
			deprecationReason
		}
		possibleTypes {
			...TypeRef
		}
	}

	fragment InputValue on __InputValue {
		name
		description
		type {
			...TypeRef
		}
		defaultValue
	}

	fragment TypeRef on __Type {
		kind
		name
		ofType {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
							}
						}
					}
				}
			}
		}
	}
`

func main() {
	format := flag.String("format", "sdl", "output format: json or sdl")
	output := flag.String("output", "", "output file (default: schema.json or schema.graphql)")
	flag.Parse()

	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}

	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}

	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := client.GQL.Request(context.Background(), schemaQuery, struct{}{})
	if err != nil {
		log.Fatal(err)
	}

	switch *format {
	case "json":
		if *output == "" {
			*output = "schema.json"
		}
	case "sdl":
		if *output == "" {
			*output = "schema.graphql"
		}
	default:
		log.Fatalf("unknown format: %s (use json or sdl)", *format)
	}

	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	switch *format {
	case "json":
		writeJSON(buf, f)
	case "sdl":
		writeSDL(buf, f)
	}
}

// writeJSON pretty-prints and writes the introspection result as JSON.
func writeJSON(buf []byte, w io.Writer) {
	var schema json.RawMessage
	if err := json.Unmarshal(buf, &schema); err != nil {
		log.Fatal(err)
	}
	pretty, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := w.Write(pretty); err != nil {
		log.Fatal(err)
	}
}

// writeSDL converts the introspection result to SDL format and writes it.
func writeSDL(buf []byte, w io.Writer) {
	var resp struct {
		Data struct {
			Schema introspectionSchema `json:"__schema"`
		} `json:"data"`
	}

	// Try nested format first (data.__schema), fall back to direct (__schema).
	if err := json.Unmarshal(buf, &resp); err != nil {
		log.Fatal(err)
	}
	if resp.Data.Schema.Types == nil {
		var direct struct {
			Schema introspectionSchema `json:"__schema"`
		}
		if err := json.Unmarshal(buf, &direct); err != nil {
			log.Fatal(err)
		}
		resp.Data.Schema = direct.Schema
	}

	sdl := renderSDL(resp.Data.Schema)
	if _, err := io.WriteString(w, sdl); err != nil {
		log.Fatal(err)
	}
}

type introspectionSchema struct {
	QueryType        *typeName   `json:"queryType"`
	MutationType     *typeName   `json:"mutationType"`
	SubscriptionType *typeName   `json:"subscriptionType"`
	Types            []fullType  `json:"types"`
	Directives       []directive `json:"directives"`
}

type typeName struct {
	Name string `json:"name"`
}

type fullType struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Fields        []field      `json:"fields"`
	InputFields   []inputValue `json:"inputFields"`
	EnumValues    []enumValue  `json:"enumValues"`
	PossibleTypes []typeRef    `json:"possibleTypes"`
	Interfaces    []typeRef    `json:"interfaces"`
}

type field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Args              []inputValue `json:"args"`
	Type              typeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

type inputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         typeRef `json:"type"`
	DefaultValue *string `json:"defaultValue"`
}

type enumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   *string  `json:"name"`
	OfType *typeRef `json:"ofType"`
}

type directive struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Locations   []string     `json:"locations"`
	Args        []inputValue `json:"args"`
}

// renderSDL converts an introspection schema to SDL text.
func renderSDL(schema introspectionSchema) string {
	var b strings.Builder

	// Collect types by kind for ordered output.
	var (
		scalars    []fullType
		enums      []fullType
		inputs     []fullType
		objects    []fullType
		interfaces []fullType
		unions     []fullType
	)

	queryTypeName := ""
	mutationTypeName := ""
	subscriptionTypeName := ""
	if schema.QueryType != nil {
		queryTypeName = schema.QueryType.Name
	}
	if schema.MutationType != nil {
		mutationTypeName = schema.MutationType.Name
	}
	if schema.SubscriptionType != nil {
		subscriptionTypeName = schema.SubscriptionType.Name
	}

	for _, t := range schema.Types {
		// Skip built-in types.
		if strings.HasPrefix(t.Name, "__") {
			continue
		}
		switch t.Kind {
		case "SCALAR":
			scalars = append(scalars, t)
		case "ENUM":
			enums = append(enums, t)
		case "INPUT_OBJECT":
			inputs = append(inputs, t)
		case "OBJECT":
			objects = append(objects, t)
		case "INTERFACE":
			interfaces = append(interfaces, t)
		case "UNION":
			unions = append(unions, t)
		}
	}

	// Sort each category by name.
	sortTypes := func(types []fullType) {
		slices.SortFunc(types, func(a, b fullType) int {
			return strings.Compare(a.Name, b.Name)
		})
	}
	sortTypes(scalars)
	sortTypes(enums)
	sortTypes(inputs)
	sortTypes(interfaces)
	sortTypes(objects)
	sortTypes(unions)

	// Schema definition.
	b.WriteString("schema {\n")
	if queryTypeName != "" {
		fmt.Fprintf(&b, "  query: %s\n", queryTypeName)
	}
	if mutationTypeName != "" {
		fmt.Fprintf(&b, "  mutation: %s\n", mutationTypeName)
	}
	if subscriptionTypeName != "" {
		fmt.Fprintf(&b, "  subscription: %s\n", subscriptionTypeName)
	}
	b.WriteString("}\n\n")

	// Scalars.
	for _, t := range scalars {
		if t.Name == "String" || t.Name == "Int" || t.Name == "Float" || t.Name == "Boolean" || t.Name == "ID" {
			continue
		}
		writeDescription(&b, t.Description, "")
		fmt.Fprintf(&b, "scalar %s\n\n", t.Name)
	}

	// Enums.
	for _, t := range enums {
		writeDescription(&b, t.Description, "")
		fmt.Fprintf(&b, "enum %s {\n", t.Name)
		for _, v := range t.EnumValues {
			writeDescription(&b, v.Description, "  ")
			if v.IsDeprecated {
				fmt.Fprintf(&b, "  %s @deprecated(reason: %q)\n", v.Name, v.DeprecationReason)
			} else {
				fmt.Fprintf(&b, "  %s\n", v.Name)
			}
		}
		b.WriteString("}\n\n")
	}

	// Input types.
	for _, t := range inputs {
		writeDescription(&b, t.Description, "")
		fmt.Fprintf(&b, "input %s {\n", t.Name)
		for _, f := range t.InputFields {
			writeDescription(&b, f.Description, "  ")
			fmt.Fprintf(&b, "  %s: %s", f.Name, renderTypeRef(f.Type))
			if f.DefaultValue != nil {
				fmt.Fprintf(&b, " = %s", *f.DefaultValue)
			}
			b.WriteString("\n")
		}
		b.WriteString("}\n\n")
	}

	// Interfaces.
	for _, t := range interfaces {
		writeDescription(&b, t.Description, "")
		fmt.Fprintf(&b, "interface %s {\n", t.Name)
		for _, f := range t.Fields {
			writeFieldSDL(&b, f)
		}
		b.WriteString("}\n\n")
	}

	// Unions.
	for _, t := range unions {
		writeDescription(&b, t.Description, "")
		var memberNames []string
		for _, pt := range t.PossibleTypes {
			if pt.Name != nil {
				memberNames = append(memberNames, *pt.Name)
			}
		}
		fmt.Fprintf(&b, "union %s = %s\n\n", t.Name, strings.Join(memberNames, " | "))
	}

	// Directives.
	slices.SortFunc(schema.Directives, func(a, b directive) int {
		return strings.Compare(a.Name, b.Name)
	})
	builtIn := map[string]bool{"skip": true, "include": true, "deprecated": true, "specifiedBy": true}
	for _, d := range schema.Directives {
		if builtIn[d.Name] {
			continue
		}
		writeDescription(&b, d.Description, "")
		fmt.Fprintf(&b, "directive @%s", d.Name)
		if len(d.Args) > 0 {
			b.WriteString("(\n")
			for _, arg := range d.Args {
				writeDescription(&b, arg.Description, "  ")
				fmt.Fprintf(&b, "  %s: %s", arg.Name, renderTypeRef(arg.Type))
				if arg.DefaultValue != nil {
					fmt.Fprintf(&b, " = %s", *arg.DefaultValue)
				}
				b.WriteString("\n")
			}
			b.WriteString(")")
		}
		fmt.Fprintf(&b, " on %s\n\n", strings.Join(d.Locations, " | "))
	}

	// Object types (Query and Mutation first, then the rest).
	for _, t := range objects {
		if t.Name != queryTypeName && t.Name != mutationTypeName {
			continue
		}
		writeObjectSDL(&b, t)
	}
	for _, t := range objects {
		if t.Name == queryTypeName || t.Name == mutationTypeName {
			continue
		}
		writeObjectSDL(&b, t)
	}

	return b.String()
}

// writeObjectSDL writes an object type definition.
func writeObjectSDL(b *strings.Builder, t fullType) {
	writeDescription(b, t.Description, "")
	fmt.Fprintf(b, "type %s", t.Name)
	if len(t.Interfaces) > 0 {
		var names []string
		for _, iface := range t.Interfaces {
			if iface.Name != nil {
				names = append(names, *iface.Name)
			}
		}
		if len(names) > 0 {
			fmt.Fprintf(b, " implements %s", strings.Join(names, " & "))
		}
	}
	b.WriteString(" {\n")
	for _, f := range t.Fields {
		writeFieldSDL(b, f)
	}
	b.WriteString("}\n\n")
}

// writeFieldSDL writes a field definition with arguments.
func writeFieldSDL(b *strings.Builder, f field) {
	writeDescription(b, f.Description, "  ")
	if len(f.Args) == 0 {
		fmt.Fprintf(b, "  %s: %s", f.Name, renderTypeRef(f.Type))
	} else {
		fmt.Fprintf(b, "  %s(\n", f.Name)
		for _, arg := range f.Args {
			writeDescription(b, arg.Description, "    ")
			fmt.Fprintf(b, "    %s: %s", arg.Name, renderTypeRef(arg.Type))
			if arg.DefaultValue != nil {
				fmt.Fprintf(b, " = %s", *arg.DefaultValue)
			}
			b.WriteString("\n")
		}
		fmt.Fprintf(b, "  ): %s", renderTypeRef(f.Type))
	}
	if f.IsDeprecated {
		fmt.Fprintf(b, " @deprecated(reason: %q)", f.DeprecationReason)
	}
	b.WriteString("\n")
}

// writeDescription writes a description using GraphQL spec syntax.
func writeDescription(b *strings.Builder, desc, indent string) {
	if desc == "" {
		return
	}
	if !strings.Contains(desc, "\n") {
		escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(desc)
		fmt.Fprintf(b, "%s\"%s\"\n", indent, escaped)
		return
	}
	fmt.Fprintf(b, "%s\"\"\"\n", indent)
	for line := range strings.SplitSeq(desc, "\n") {
		fmt.Fprintf(b, "%s%s\n", indent, line)
	}
	fmt.Fprintf(b, "%s\"\"\"\n", indent)
}

// renderTypeRef converts a type reference to SDL notation.
func renderTypeRef(ref typeRef) string {
	switch ref.Kind {
	case "NON_NULL":
		if ref.OfType != nil {
			return renderTypeRef(*ref.OfType) + "!"
		}
		return "!"
	case "LIST":
		if ref.OfType != nil {
			return "[" + renderTypeRef(*ref.OfType) + "]"
		}
		return "[]"
	default:
		if ref.Name != nil {
			return *ref.Name
		}
		return "Unknown"
	}
}
