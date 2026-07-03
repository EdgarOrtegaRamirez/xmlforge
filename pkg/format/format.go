// Package format provides XML formatting and pretty-printing utilities.
package format

import (
	"bytes"
	"fmt"
	"strings"
)

// Options configures XML formatting behavior.
type Options struct {
	Indent      string // Indentation string (default: "  ")
	MaxLineLen  int    // Max line length before wrapping (0 = no wrapping)
	SelfClosing bool   // Use self-closing tags for empty elements
	SpacesAttr  bool   // Space around attribute equals sign
	Prolog      bool   // Include XML declaration
	AttributeWrap bool // Wrap attributes on separate lines when long
}

// DefaultOptions returns sensible default formatting options.
func DefaultOptions() Options {
	return Options{
		Indent:      "  ",
		MaxLineLen:  0,
		SelfClosing: true,
		SpacesAttr:  true,
		Prolog:      false,
	}
}

// Format formats XML with the given options.
func Format(xml string, opts Options) (string, error) {
	var buf bytes.Buffer
	indent := 0
	indentStr := opts.Indent
	if indentStr == "" {
		indentStr = "  "
	}

	lines := strings.Split(xml, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Closing tag
		if strings.HasPrefix(trimmed, "</") {
			indent--
			if indent < 0 {
				indent = 0
			}
			buf.WriteString(strings.Repeat(indentStr, indent))
			buf.WriteString(trimmed)
			buf.WriteString("\n")
			continue
		}

		// Self-closing tag
		if strings.HasSuffix(trimmed, "/>") || (strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") && !strings.Contains(trimmed, "</")) {
			// Check if it's a self-closing tag or an opening tag
			if strings.HasSuffix(trimmed, "/>") {
				buf.WriteString(strings.Repeat(indentStr, indent))
				buf.WriteString(trimmed)
				buf.WriteString("\n")
				continue
			}
			// Single-line element with text
			buf.WriteString(strings.Repeat(indentStr, indent))
			buf.WriteString(trimmed)
			buf.WriteString("\n")
			indent++
			// Check for closing tag on same line
			if strings.Contains(trimmed, "</") {
				indent--
			}
			continue
		}

		// Opening tag
		if strings.HasPrefix(trimmed, "<") {
			buf.WriteString(strings.Repeat(indentStr, indent))
			buf.WriteString(trimmed)
			buf.WriteString("\n")
			indent++
			continue
		}

		// Text content
		buf.WriteString(strings.Repeat(indentStr, indent))
		buf.WriteString(trimmed)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// FormatBytes formats XML bytes with the given options.
func FormatBytes(xml []byte, opts Options) ([]byte, error) {
	result, err := Format(string(xml), opts)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

// Compress removes unnecessary whitespace from XML.
func Compress(xml string) string {
	var buf bytes.Buffer
	inTag := false
	prevWasSpace := false

	for _, ch := range xml {
		if ch == '<' {
			inTag = true
			prevWasSpace = false
		} else if ch == '>' {
			inTag = false
			prevWasSpace = false
		}

		if inTag {
			buf.WriteRune(ch)
			continue
		}

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if !prevWasSpace {
				buf.WriteRune(' ')
				prevWasSpace = true
			}
		} else {
			buf.WriteRune(ch)
			prevWasSpace = false
		}
	}

	return strings.TrimSpace(buf.String())
}

// Minify is an alias for Compress.
func Minify(xml string) string {
	return Compress(xml)
}

// Indent increases indentation level of XML.
func Indent(xml string, indent string) string {
	lines := strings.Split(xml, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		result = append(result, indent+trimmed)
	}

	return strings.Join(result, "\n")
}

// ValidateStructure performs basic structural validation of XML.
func ValidateStructure(xml string) []string {
	var errors []string
	depth := 0
	inTag := false
	var tagName strings.Builder

	for i, ch := range xml {
		switch {
		case ch == '<':
			if inTag {
				errors = append(errors, fmt.Sprintf("unexpected '<' inside tag at position %d", i))
			}
			inTag = true
			tagName.Reset()
		case ch == '>':
			if inTag {
				name := strings.TrimSpace(tagName.String())
				if strings.HasPrefix(name, "/") {
					// Closing tag
					depth--
					if depth < 0 {
						errors = append(errors, fmt.Sprintf("unexpected closing tag at position %d", i))
						depth = 0
					}
				} else if strings.HasSuffix(name, "/") {
					// Self-closing tag - no depth change
				} else if !strings.HasPrefix(name, "?") && !strings.HasPrefix(name, "!") {
					// Opening tag
					depth++
				}
				inTag = false
			}
		default:
			if inTag {
				tagName.WriteRune(ch)
			}
		}
	}

	if depth != 0 {
		errors = append(errors, fmt.Sprintf("unbalanced tags: depth %d", depth))
	}

	return errors
}
