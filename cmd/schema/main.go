package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
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
	format := flag.String("format", "json", "output format: json or sdl")
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

	buf, err := client.GQL.Request(context.Background(), schemaQuery, nil)
	if err != nil {
		log.Fatal(err)
	}

	switch *format {
	case "json":
		outputFile := *output
		if outputFile == "" {
			outputFile = "schema.json"
		}
		writeJSON(buf, outputFile)
	case "sdl":
		outputFile := *output
		if outputFile == "" {
			outputFile = "schema.graphql"
		}
		writeSDL(buf, outputFile)
	default:
		log.Fatalf("unknown format: %s (use json or sdl)", *format)
	}
}

// writeJSON pretty-prints and writes the introspection result as JSON.
func writeJSON(buf []byte, outputFile string) {
	var schema json.RawMessage
	if err := json.Unmarshal(buf, &schema); err != nil {
		log.Fatal(err)
	}
	pretty, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(outputFile, pretty, 0o644); err != nil {
		log.Fatal(err)
	}
}

// writeSDL converts the introspection result to SDL format and writes it.
func writeSDL(buf []byte, outputFile string) {
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
	if err := os.WriteFile(outputFile, []byte(sdl), 0o644); err != nil {
		log.Fatal(err)
	}
}

type introspectionSchema struct {
	QueryType        *typeName          `json:"queryType"`
	MutationType     *typeName          `json:"mutationType"`
	Types            []fullType         `json:"types"`
	Directives       []directive        `json:"directives"`
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
	if schema.QueryType != nil {
		queryTypeName = schema.QueryType.Name
	}
	if schema.MutationType != nil {
		mutationTypeName = schema.MutationType.Name
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
		sort.Slice(types, func(i, j int) bool {
			return types[i].Name < types[j].Name
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

// writeDescription writes a description as a block comment if non-empty.
func writeDescription(b *strings.Builder, desc, indent string) {
	if desc == "" {
		return
	}
	for line := range strings.SplitSeq(desc, "\n") {
		fmt.Fprintf(b, "%s# %s\n", indent, line)
	}
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
