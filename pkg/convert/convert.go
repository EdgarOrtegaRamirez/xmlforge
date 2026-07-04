// Package convert provides XML conversion utilities (XML to JSON, CSV).
package convert

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

// JSONOptions configures JSON output.
type JSONOptions struct {
	Indent          string // JSON indentation (default: "  ")
	AttributePrefix string // Prefix for attributes (default: "@")
	TextKey         string // Key for text content (default: "#text")
	Compact         bool   // Compact output (no whitespace)
}

// DefaultJSONOptions returns sensible default options.
func DefaultJSONOptions() JSONOptions {
	return JSONOptions{
		Indent:          "  ",
		AttributePrefix: "@",
		TextKey:         "#text",
	}
}

// ToJSON converts an XML document to JSON.
func ToJSON(doc *parser.Document, opts JSONOptions) (string, error) {
	result := nodeToMap(doc.Root, opts)
	data, err := marshalJSON(result, opts)
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(data), nil
}

// ToJSONBytes converts an XML document to JSON bytes.
func ToJSONBytes(doc *parser.Document, opts JSONOptions) ([]byte, error) {
	result := nodeToMap(doc.Root, opts)
	return marshalJSON(result, opts)
}

func nodeToMap(node *parser.Node, opts JSONOptions) map[string]interface{} {
	result := make(map[string]interface{})

	// Add attributes
	for _, attr := range node.Attributes {
		key := opts.AttributePrefix + attr.Name
		result[key] = attr.Value
	}

	// Process children
	textParts := []string{}
	childElements := make(map[string][]interface{})

	for _, child := range node.Children {
		switch child.Type {
		case parser.NodeText:
			textParts = append(textParts, strings.TrimSpace(child.Value))
		case parser.NodeComment:
			// Skip comments in JSON output
		case parser.NodeInstruction:
			// Skip PIs in JSON output
		case parser.NodeElement:
			childMap := nodeToMap(child, opts)
			name := child.Name
			childElements[name] = append(childElements[name], childMap)
		}
	}

	// Handle text content
	text := strings.Join(textParts, " ")
	if text != "" {
		if len(childElements) > 0 {
			result[opts.TextKey] = text
		} else if len(result) == 0 {
			result[opts.TextKey] = text
		} else {
			result[opts.TextKey] = text
		}
	}

	// Add child elements
	for name, children := range childElements {
		if len(children) == 1 {
			result[name] = children[0]
		} else {
			result[name] = children
		}
	}

	return result
}

func marshalJSON(data interface{}, opts JSONOptions) ([]byte, error) {
	if opts.Compact {
		return json.Marshal(data)
	}
	return json.MarshalIndent(data, "", opts.Indent)
}

// ToCSV converts XML to CSV format (for flat structures).
func ToCSV(doc *parser.Document) (string, error) {
	if doc.Root == nil {
		return "", fmt.Errorf("empty document")
	}

	var sb strings.Builder

	// Collect all unique attributes and text from children
	headers := make(map[string]bool)
	var rows []map[string]string

	for _, child := range doc.Root.Children {
		if child.Type != parser.NodeElement {
			continue
		}
		row := make(map[string]string)
		for _, attr := range child.Attributes {
			headers[attr.Name] = true
			row[attr.Name] = attr.Value
		}
		if text := child.GetText(); text != "" {
			headers["#text"] = true
			row["#text"] = text
		}
		rows = append(rows, row)
	}

	// Write headers
	headerList := make([]string, 0, len(headers))
	for h := range headers {
		headerList = append(headerList, h)
	}
	sb.WriteString(strings.Join(headerList, ","))
	sb.WriteString("\n")

	// Write rows
	for _, row := range rows {
		values := make([]string, 0, len(headerList))
		for _, h := range headerList {
			values = append(values, row[h])
		}
		sb.WriteString(strings.Join(values, ","))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
