# XmlForge

A comprehensive XML processing toolkit written in Go. Parse, format, query, diff, convert, and validate XML documents with a powerful CLI tool and library API.

## Features

- **Parse** — Robust XML parser handling elements, attributes, text, comments, CDATA, processing instructions, namespaces, and BOM
- **Format** — Pretty-print and minify XML
- **Query** — XPath-like queries to find and extract data
- **Diff** — Compare two XML documents and highlight differences
- **Convert** — Transform XML to JSON, CSV, and other formats
- **Validate** — Check XML against configurable rules
- **Stats** — Analyze document structure and statistics

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/xmlforge/cmd/xmlforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/xmlforge.git
cd xmlforge
go build -o xmlforge ./cmd/xmlforge/
```

## CLI Usage

```bash
# Parse and display XML structure
xmlforge parse input.xml

# Pretty-print XML
xmlforge format input.xml

# Minify XML
xmlforge compress input.xml

# Show document statistics
xmlforge stats input.xml

# Query with XPath
xmlforge xpath "root/child" input.xml
xmlforge xpath 'book[@id="1"]/title' input.xml

# Compare two XML files
xmlforge diff file1.xml file2.xml

# Convert to JSON
xmlforge convert --to json input.xml

# Convert to CSV
xmlforge convert --to csv input.xml

# Validate XML
xmlforge validate input.xml
xmlforge validate --max-depth 10 --no-empty input.xml
```

## Library API

### Parsing

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"

doc, err := parser.ParseString(`<root><child>text</child></root>`)
if err != nil {
    log.Fatal(err)
}

fmt.Println(doc.Root.Name) // "root"
fmt.Println(doc.Root.GetChild(0).GetText()) // "text"
```

### XPath Queries

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/xpath"

results, _ := xpath.Execute(doc, "root/child")
fmt.Printf("Found %d nodes\n", len(results))

// Or use convenience functions
text := xpath.FindText(doc.Root, "root/child") // "text"
node := xpath.FindOne(doc.Root, "book[@id='1']")
```

### Diff

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/diff"

result, _ := diff.CompareStrings(xml1, xml2)
if !result.Identical {
    fmt.Print(result.Format())
}
```

### Convert

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/convert"

jsonStr, _ := convert.ToJSON(doc, convert.DefaultJSONOptions())
csvStr, _ := convert.ToCSV(doc)
```

### Validate

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/validator"

rules := validator.Rules{
    MaxDepth:      10,
    MaxAttributes: 5,
    RequiredAttrs: map[string][]string{
        "item": {"id", "name"},
    },
}

result, _ := validator.ValidateString(xml, rules)
if !result.Valid {
    fmt.Print(result.Format())
}
```

### Stats

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/stats"

s := stats.Analyze(doc)
fmt.Print(s.Format())
```

### Format

```go
import "github.com/EdgarOrtegaRamirez/xmlforge/pkg/format"

formatted, _ := format.Format(xml, format.DefaultOptions())
minified := format.Compress(xml)
```

## Architecture

```
xmlforge/
├── cmd/xmlforge/     # CLI entry point
├── pkg/
│   ├── parser/       # XML parser (AST, tokenizer)
│   ├── xpath/        # XPath-like queries
│   ├── diff/         # XML comparison
│   ├── convert/      # XML to JSON/CSV conversion
│   ├── stats/        # Document statistics
│   ├── format/       # XML formatting/pretty-printing
│   └── validator/    # XML validation rules
├── testdata/         # Test XML files
└── tests/            # Integration tests
```

## Features in Detail

### Parser
- Handles all XML 1.0 constructs: elements, attributes, text, comments, CDATA, processing instructions
- Namespace support with `xmlns` declarations
- BOM (Byte Order Mark) handling
- Lenient mode for malformed XML
- Detailed error reporting with line/column information

### XPath
- Slash-separated path expressions
- Wildcard matching with `*`
- Attribute predicates: `[@attr="value"]`
- Axes: child (default), descendant, parent (`..`), self (`.`)

### Diff
- Element-by-element comparison
- Attribute change detection
- Added/removed node detection
- Text content change detection
- XPath-style paths for differences

## License

MIT
