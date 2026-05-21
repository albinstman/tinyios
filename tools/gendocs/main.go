// Command gendocs generates the API Reference section of the README directly
// from the godoc/swagger-style annotations on the handler functions in main.go.
//
// It parses the source with go/ast (it never compiles or runs the program, so
// it has no dependency on the device libraries) and rewrites the block between
// the API:START and API:END markers in the README. Run it with `make docs`.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	startMarker = "<!-- API:START -->"
	endMarker   = "<!-- API:END -->"
)

// preferredTagOrder lays the tags out as a logical device setup flow rather
// than alphabetically. Unknown tags are appended afterwards, sorted.
var preferredTagOrder = []string{
	"device", "pairing", "activation", "supervision",
	"developer", "profiles", "apps", "wda",
}

type param struct {
	name, in, typ, desc string
	required            bool
}

type endpoint struct {
	method, path, summary, desc, tag string
	params                           []param
	consumesJSON                     bool
}

type structField struct{ json, typ string }

func main() {
	src := flag.String("src", "main.go", "Go source file to parse for annotations")
	readme := flag.String("readme", "README.md", "README file to update")
	flag.Parse()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, *src, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("parse %s: %v", *src, err)
	}

	endpoints, structs := collect(file)
	if len(endpoints) == 0 {
		log.Fatalf("no annotated endpoints found in %s", *src)
	}

	section := render(endpoints, structs)
	if err := writeBlock(*readme, section); err != nil {
		log.Fatalf("update %s: %v", *readme, err)
	}
	fmt.Printf("Wrote %d endpoints to %s\n", len(endpoints), *readme)
}

// collect walks the file for annotated handler functions and struct types.
func collect(file *ast.File) ([]endpoint, map[string][]structField) {
	var endpoints []endpoint
	structs := map[string][]structField{}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc == nil {
				continue
			}
			if ep, ok := parseEndpoint(d.Doc.Text()); ok {
				endpoints = append(endpoints, ep)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				structs[ts.Name.Name] = parseStruct(st)
			}
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		ti, tj := tagRank(endpoints[i].tag), tagRank(endpoints[j].tag)
		if ti != tj {
			return ti < tj
		}
		if endpoints[i].path != endpoints[j].path {
			return endpoints[i].path < endpoints[j].path
		}
		return endpoints[i].method < endpoints[j].method
	})
	return endpoints, structs
}

func parseEndpoint(doc string) (endpoint, bool) {
	var ep endpoint
	hasRouter := false
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "@") {
			continue
		}
		tag, rest := splitFirst(line)
		switch tag {
		case "@Summary":
			ep.summary = rest
		case "@Description":
			ep.desc = rest
		case "@Tags":
			ep.tag = strings.TrimSpace(rest)
		case "@Accept":
			if strings.Contains(rest, "json") {
				ep.consumesJSON = true
			}
		case "@Param":
			if p, ok := parseParam(rest); ok {
				ep.params = append(ep.params, p)
			}
		case "@Router":
			ep.method, ep.path = parseRouter(rest)
			hasRouter = true
		}
	}
	return ep, hasRouter
}

// parseRouter reads `/{udid}/reboot [post]` into ("POST", "/{udid}/reboot").
func parseRouter(rest string) (method, path string) {
	rest = strings.TrimSpace(rest)
	open := strings.LastIndex(rest, "[")
	close := strings.LastIndex(rest, "]")
	if open >= 0 && close > open {
		method = strings.ToUpper(strings.TrimSpace(rest[open+1 : close]))
		path = strings.TrimSpace(rest[:open])
	} else {
		path = rest
	}
	return method, path
}

// parseParam reads `udid path string true "Device UDID"`.
func parseParam(rest string) (param, bool) {
	fields := strings.Fields(rest)
	if len(fields) < 4 {
		return param{}, false
	}
	p := param{
		name:     fields[0],
		in:       normalizeIn(fields[1]),
		typ:      fields[2],
		required: fields[3] == "true",
	}
	if i := strings.Index(rest, `"`); i >= 0 {
		if j := strings.LastIndex(rest, `"`); j > i {
			p.desc = rest[i+1 : j]
		}
	}
	return p, true
}

// splitFirst splits a line into its first whitespace-delimited token and the
// trimmed remainder, e.g. "@Summary  List devices" -> ("@Summary", "List devices").
func splitFirst(line string) (head, rest string) {
	fields := strings.SplitN(strings.TrimSpace(line), " ", 2)
	head = fields[0]
	if len(fields) > 1 {
		rest = strings.TrimSpace(fields[1])
	}
	return head, rest
}

func normalizeIn(in string) string {
	switch in {
	case "formData":
		return "form"
	default:
		return in
	}
}

func parseStruct(st *ast.StructType) []structField {
	var fields []structField
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 || f.Tag == nil {
			continue
		}
		name := jsonTagName(f.Tag.Value)
		if name == "" || name == "-" {
			name = f.Names[0].Name
		}
		fields = append(fields, structField{json: name, typ: exprString(f.Type)})
	}
	return fields
}

func jsonTagName(tag string) string {
	tag = strings.Trim(tag, "`")
	idx := strings.Index(tag, `json:"`)
	if idx < 0 {
		return ""
	}
	rest := tag[idx+len(`json:"`):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	val := rest[:end]
	if comma := strings.Index(val, ","); comma >= 0 {
		val = val[:comma]
	}
	return val
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.StarExpr:
		return exprString(t.X)
	default:
		return "object"
	}
}

func tagRank(tag string) int {
	for i, t := range preferredTagOrder {
		if t == tag {
			return i
		}
	}
	return len(preferredTagOrder) + 1
}

func render(endpoints []endpoint, structs map[string][]structField) string {
	var b strings.Builder
	b.WriteString("## API Reference\n\n")
	b.WriteString("Base URL `http://localhost:8080` — every response is JSON.\n\n")
	b.WriteString("Device-scoped routes take the target device's `udid` as the first path segment. ")
	b.WriteString("List connected devices with `GET /devices` to find it.\n\n")

	// Overview table, grouped by tag.
	b.WriteString("### Overview\n\n")
	b.WriteString("| | Method | Endpoint | Description |\n")
	b.WriteString("|---|---|---|---|\n")
	lastTag := ""
	for _, ep := range endpoints {
		if ep.tag != lastTag {
			fmt.Fprintf(&b, "| **%s** | | | |\n", displayTag(ep.tag))
			lastTag = ep.tag
		}
		fmt.Fprintf(&b, "| | `%s` | `%s` | %s |\n", ep.method, ep.path, ep.summary)
	}
	b.WriteString("\n")

	// Detailed sections, grouped by tag.
	lastTag = ""
	for _, ep := range endpoints {
		if ep.tag != lastTag {
			fmt.Fprintf(&b, "### %s\n\n", displayTag(ep.tag))
			lastTag = ep.tag
		}
		renderEndpoint(&b, ep, structs)
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderEndpoint(b *strings.Builder, ep endpoint, structs map[string][]structField) {
	fmt.Fprintf(b, "#### `%s %s`\n\n", ep.method, ep.path)
	if ep.desc != "" {
		fmt.Fprintf(b, "%s\n\n", ep.desc)
	}

	if len(ep.params) > 0 {
		b.WriteString("| Name | In | Type | Required | Description |\n")
		b.WriteString("|---|---|---|:--:|---|\n")
		for _, p := range ep.params {
			req := "no"
			if p.required {
				req = "yes"
			}
			fmt.Fprintf(b, "| `%s` | %s | %s | %s | %s |\n", p.name, p.in, p.typ, req, p.desc)
		}
		b.WriteString("\n")
	}

	// Inline the JSON body shape for any body parameter we have a struct for.
	for _, p := range ep.params {
		if p.in != "body" {
			continue
		}
		if fields, ok := structs[p.typ]; ok && len(fields) > 0 {
			b.WriteString("Body:\n\n```json\n")
			b.WriteString(jsonSample(fields))
			b.WriteString("\n```\n\n")
		}
	}

	b.WriteString("```sh\n")
	b.WriteString(curlExample(ep, structs))
	b.WriteString("\n```\n\n")
}

func jsonSample(fields []structField) string {
	var lines []string
	for _, f := range fields {
		lines = append(lines, fmt.Sprintf("  %q: %s", f.json, sampleValue(f.typ)))
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n}"
}

func sampleValue(typ string) string {
	switch {
	case strings.HasPrefix(typ, "[]"):
		return "[]"
	case typ == "bool":
		return "false"
	case strings.HasPrefix(typ, "int") || strings.HasPrefix(typ, "uint") || strings.HasPrefix(typ, "float"):
		return "0"
	default:
		return `"<string>"`
	}
}

func curlExample(ep endpoint, structs map[string][]structField) string {
	url := "http://localhost:8080" + strings.ReplaceAll(ep.path, "{udid}", "<udid>")

	// The passthrough proxy accepts any method; show a representative GET.
	if ep.method == "ANY" {
		return "curl " + strings.TrimSuffix(url, "/") + "/status"
	}

	if ep.method == "GET" {
		return "curl " + url
	}

	cmd := fmt.Sprintf("curl -X %s %s", ep.method, url)

	// JSON body.
	for _, p := range ep.params {
		if p.in == "body" {
			if fields, ok := structs[p.typ]; ok && len(fields) > 0 {
				return cmd + " \\\n  -H 'Content-Type: application/json' \\\n  -d '" +
					compactJSON(fields) + "'"
			}
		}
	}

	// Form fields.
	var forms []string
	for _, p := range ep.params {
		if p.in == "form" {
			forms = append(forms, fmt.Sprintf("-d %q", p.name+"=<"+p.name+">"))
		}
	}
	if len(forms) > 0 {
		return cmd + " " + strings.Join(forms, " ")
	}

	return cmd
}

func compactJSON(fields []structField) string {
	var parts []string
	for _, f := range fields {
		parts = append(parts, fmt.Sprintf("%q: %s", f.json, sampleValue(f.typ)))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func displayTag(tag string) string {
	switch tag {
	case "wda":
		return "WDA"
	case "":
		return "Other"
	default:
		return strings.ToUpper(tag[:1]) + tag[1:]
	}
}

// writeBlock replaces the content between the markers in the README, inserting
// the markers at the end of the file if they are not already present.
func writeBlock(path, section string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(raw)

	block := startMarker + "\n" +
		"<!-- Generated by `make docs` from annotations in main.go. Do not edit by hand. -->\n\n" +
		section + "\n" + endMarker

	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if start >= 0 && end > start {
		content = content[:start] + block + content[end+len(endMarker):]
	} else {
		content = strings.TrimRight(content, "\n") + "\n\n" + block + "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
