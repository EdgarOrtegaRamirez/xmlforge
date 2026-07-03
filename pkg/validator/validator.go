// Package validator provides XML validation utilities.
package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

// Error represents a validation error.
type Error struct {
 Line    int
 Column  int
	Message string
	Level   string // "error" or "warning"
}

func (e Error) Error() string {
	return fmt.Sprintf("[line %d, col %d] %s", e.Line, e.Column, e.Message)
}

// Result holds validation results.
type Result struct {
	Valid  bool
	Errors []Error
}

// Rules defines validation rules.
type Rules struct {
	RequireRoot       bool     // Require a root element
	MaxDepth          int      // Maximum nesting depth (0 = unlimited)
	MaxAttributes     int      // Max attributes per element (0 = unlimited)
	AllowedElements   []string // If set, only these elements are allowed
	RequiredElements  []string // Elements that must be present
	ForbiddenElements []string // Elements that must not be present
	MaxTextLength     int      // Maximum text content length (0 = unlimited)
	NoEmptyElements   bool     // Disallow empty elements
	ValidateNames     bool     // Validate XML name rules
	RequiredAttrs     map[string][]string // element -> required attribute names
}

// DefaultRules returns sensible default validation rules.
func DefaultRules() Rules {
	return Rules{
		RequireRoot: true,
		MaxDepth:    100,
	}
}

// Validator validates XML documents.
type Validator struct {
	rules Rules
}

// New creates a new Validator with the given rules.
func New(rules Rules) *Validator {
	return &Validator{rules: rules}
}

// Validate validates a parsed XML document.
func (v *Validator) Validate(doc *parser.Document) *Result {
	result := &Result{Valid: true}

	if v.rules.RequireRoot && doc.Root == nil {
		result.Valid = false
		result.Errors = append(result.Errors, Error{
			Message: "document has no root element",
			Level:   "error",
		})
		return result
	}

	if doc.Root != nil {
		v.validateNode(doc.Root, 0, result)
	}

	return result
}

// ValidateString parses and validates an XML string.
func ValidateString(xml string, rules Rules) (*Result, error) {
	doc, err := parser.ParseString(xml)
	if err != nil {
		return nil, fmt.Errorf("parsing XML: %w", err)
	}

	v := New(rules)
	return v.Validate(doc), nil
}

func (v *Validator) validateNode(node *parser.Node, depth int, result *Result) {
	// Check max depth
	if v.rules.MaxDepth > 0 && depth > v.rules.MaxDepth {
		result.Valid = false
		result.Errors = append(result.Errors, Error{
			Line:   node.Line,
			Column: node.Column,
			Message: fmt.Sprintf("maximum depth %d exceeded at <%s>", v.rules.MaxDepth, node.Name),
			Level:  "error",
		})
		return
	}

	// Check max attributes
	if v.rules.MaxAttributes > 0 && len(node.Attributes) > v.rules.MaxAttributes {
		result.Valid = false
		result.Errors = append(result.Errors, Error{
			Line:   node.Line,
			Column: node.Column,
			Message: fmt.Sprintf("element <%s> has %d attributes (max %d)", node.Name, len(node.Attributes), v.rules.MaxAttributes),
			Level:  "error",
		})
	}

	// Check allowed elements
	if len(v.rules.AllowedElements) > 0 {
		allowed := false
		for _, name := range v.rules.AllowedElements {
			if name == node.Name {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				Line:   node.Line,
				Column: node.Column,
				Message: fmt.Sprintf("element <%s> is not allowed", node.Name),
				Level:  "error",
			})
		}
	}

	// Check forbidden elements
	for _, name := range v.rules.ForbiddenElements {
		if name == node.Name {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				Line:   node.Line,
				Column: node.Column,
				Message: fmt.Sprintf("element <%s> is forbidden", node.Name),
				Level:  "error",
			})
		}
	}

	// Check empty elements
	if v.rules.NoEmptyElements && len(node.Children) == 0 && node.GetText() == "" {
		result.Errors = append(result.Errors, Error{
			Line:   node.Line,
			Column: node.Column,
			Message: fmt.Sprintf("element <%s> is empty", node.Name),
			Level:  "warning",
		})
	}

	// Check max text length
	if v.rules.MaxTextLength > 0 {
		text := node.GetText()
		if len(text) > v.rules.MaxTextLength {
			result.Errors = append(result.Errors, Error{
				Line:   node.Line,
				Column: node.Column,
				Message: fmt.Sprintf("text content in <%s> exceeds max length %d (got %d)", node.Name, v.rules.MaxTextLength, len(text)),
				Level:  "warning",
			})
		}
	}

	// Check required attributes
	if v.rules.RequiredAttrs != nil {
		if required, ok := v.rules.RequiredAttrs[node.Name]; ok {
			for _, attrName := range required {
				found := false
				for _, attr := range node.Attributes {
					if attr.Name == attrName {
						found = true
						break
					}
				}
				if !found {
					result.Valid = false
					result.Errors = append(result.Errors, Error{
						Line:   node.Line,
						Column: node.Column,
						Message: fmt.Sprintf("element <%s> missing required attribute @%s", node.Name, attrName),
						Level:  "error",
					})
				}
			}
		}
	}

	// Validate XML names
	if v.rules.ValidateNames {
		if !isValidXMLName(node.Name) {
			result.Errors = append(result.Errors, Error{
				Line:   node.Line,
				Column: node.Column,
				Message: fmt.Sprintf("invalid XML name: %s", node.Name),
				Level:  "warning",
			})
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		if child.Type == parser.NodeElement {
			v.validateNode(child, depth+1, result)
		}
	}
}

// CheckRequiredElements checks that all required elements are present.
func CheckRequiredElements(doc *parser.Document, required []string) []Error {
	var errors []Error
	if doc.Root == nil {
		for _, name := range required {
			errors = append(errors, Error{
				Message: fmt.Sprintf("required element <%s> not found", name),
				Level:   "error",
			})
		}
		return errors
	}

	for _, name := range required {
		if !elementExists(doc.Root, name) {
			errors = append(errors, Error{
				Message: fmt.Sprintf("required element <%s> not found", name),
				Level:   "error",
			})
		}
	}

	return errors
}

func elementExists(node *parser.Node, name string) bool {
	if node.Name == name {
		return true
	}
	for _, child := range node.Children {
		if child.Type == parser.NodeElement {
			if elementExists(child, name) {
				return true
			}
		}
	}
	return false
}

// isValidXMLName checks if a name follows basic XML naming rules.
var xmlNamePattern = regexp.MustCompile(`^[a-zA-Z_:][\w\.\-:]*$`)

func isValidXMLName(name string) bool {
	return xmlNamePattern.MatchString(name)
}

// Format returns a human-readable string of validation results.
func (r *Result) Format() string {
	if r.Valid && len(r.Errors) == 0 {
		return "Document is valid."
	}

	var sb strings.Builder
	if r.Valid {
		sb.WriteString("Document is valid with warnings:\n\n")
	} else {
		sb.WriteString("Document is invalid:\n\n")
	}

	for _, e := range r.Errors {
		sb.WriteString(fmt.Sprintf("  [%s] %s\n", strings.ToUpper(e.Level), e))
	}

	return sb.String()
}

// Summary returns a compact summary.
func (r *Result) Summary() string {
	errors, warnings := 0, 0
	for _, e := range r.Errors {
		if e.Level == "error" {
			errors++
		} else {
			warnings++
		}
	}
	if r.Valid {
		return fmt.Sprintf("valid (warnings=%d)", warnings)
	}
	return fmt.Sprintf("invalid (errors=%d, warnings=%d)", errors, warnings)
}
