# AGENTS.md

## Project Overview

XmlForge is a comprehensive XML processing toolkit written in Go. It provides both a CLI tool and a library API for parsing, formatting, querying, diffing, converting, and validating XML documents.

## Project Structure

- `cmd/xmlforge/` — CLI entry point, command handlers
- `pkg/parser/` — Core XML parser with AST, tokenizer, and reader
- `pkg/xpath/` — XPath-like query engine
- `pkg/diff/` — XML document comparison
- `pkg/convert/` — XML to JSON/CSV conversion
- `pkg/stats/` — Document statistics and analysis
- `pkg/format/` — XML formatting and pretty-printing
- `pkg/validator/` — XML validation with configurable rules
- `testdata/` — Test XML files
- `tests/` — Integration tests

## Development Guidelines

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./pkg/parser/ -v -count=1

# Run benchmarks (if any)
go test -bench=. ./...
```

### Building

```bash
# Build CLI binary
go build -o xmlforge ./cmd/xmlforge/

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o xmlforge-linux ./cmd/xmlforge/
GOOS=darwin GOARCH=arm64 go build -o xmlforge-mac ./cmd/xmlforge/
GOOS=windows GOARCH=amd64 go build -o xmlforge.exe ./cmd/xmlforge/
```

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Add comments for exported functions and types
- Handle errors explicitly (no `_` for errors)
- Use structured logging where appropriate

### Testing Conventions

- Test files should be in the same package as the code they test
- Use table-driven tests for multiple test cases
- Test both happy paths and error cases
- Use meaningful test names that describe the scenario

### Adding New Features

1. Create the package in `pkg/`
2. Add tests in the same package
3. Update the CLI in `cmd/xmlforge/main.go`
4. Update this AGENTS.md and README.md

## Key API Patterns

### Parser Usage

```go
doc, err := parser.ParseString(xml)
// or
doc, err := parser.Parse(reader, parser.WithLenient(true))
```

### XPath Queries

```go
results, err := xpath.Execute(doc, "root/child")
text := xpath.FindText(doc.Root, "root/child")
```

### Diff Comparison

```go
result, err := diff.CompareStrings(xml1, xml2)
fmt.Print(result.Format())
```

## Dependencies

- Go 1.21+ standard library only (no external dependencies)
- All packages use only `encoding/json`, `fmt`, `strings`, `regexp`, etc.

## Performance Notes

- Parser uses streaming approach (reads from io.Reader)
- XPath uses tree traversal (not compiled to bytecode)
- Diff compares in-memory ASTs
- Format uses string manipulation (not DOM-based)
