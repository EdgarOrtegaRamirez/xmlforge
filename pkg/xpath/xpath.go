// Package xpath provides XPath-like query capabilities for XML documents.
package xpath

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

// Query represents a parsed XPath expression.
type Query struct {
	Steps []Step
}

// Step represents a single step in an XPath expression.
type Step struct {
	Name      string // Element name ("*" for any)
	Predicate string // Optional predicate (e.g., [@id="1"])
	Axis      string // Axis: child (default), descendant, parent, self
}

// Compile parses an XPath-like expression into a Query.
func Compile(expr string) (*Query, error) {
	q := &Query{}
	expr = strings.TrimSpace(expr)

	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Handle absolute path
	if strings.HasPrefix(expr, "/") {
		expr = expr[1:]
	}

	parts := strings.Split(expr, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}

		step := Step{}

		// Parse axis
		if strings.HasPrefix(part, "..") {
			step.Axis = "parent"
			step.Name = "*"
			part = part[2:]
		} else if strings.HasPrefix(part, ".") {
			step.Axis = "self"
			step.Name = "*"
			part = part[1:]
		}

		// Parse predicate
		if idx := strings.Index(part, "["); idx >= 0 {
			if end := strings.Index(part[idx:], "]"); end >= 0 {
				step.Predicate = part[idx+1 : idx+end]
				part = part[:idx]
			}
		}

		// Parse name
		if step.Name == "" {
			step.Name = part
		}

		q.Steps = append(q.Steps, step)
	}

	return q, nil
}

// MustCompile is like Compile but panics on error.
func MustCompile(expr string) *Query {
	q, err := Compile(expr)
	if err != nil {
		panic(err)
	}
	return q
}

// Execute runs the query against a document and returns matching nodes.
func Execute(doc *parser.Document, expr string) ([]*parser.Node, error) {
	q, err := Compile(expr)
	if err != nil {
		return nil, err
	}
	return q.Execute(doc)
}

// Execute runs the compiled query against a document.
func (q *Query) Execute(doc *parser.Document) ([]*parser.Node, error) {
	if doc.Root == nil {
		return nil, nil
	}

	// Start with root
	current := []*parser.Node{doc.Root}

	// Execute each step
	for _, step := range q.Steps {
		var next []*parser.Node
		for _, node := range current {
			matched := executeStep(node, step)
			next = append(next, matched...)
		}
		current = next
	}

	return current, nil
}

// Find finds nodes matching an XPath expression.
func Find(root *parser.Node, expr string) []*parser.Node {
	q, err := Compile(expr)
	if err != nil {
		return nil
	}

	current := []*parser.Node{root}
	for _, step := range q.Steps {
		var next []*parser.Node
		for _, node := range current {
			matched := executeStep(node, step)
			next = append(next, matched...)
		}
		current = next
	}
	return current
}

// FindOne finds the first node matching an XPath expression.
func FindOne(root *parser.Node, expr string) *parser.Node {
	results := Find(root, expr)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// FindText returns the text content of the first matching node.
func FindText(root *parser.Node, expr string) string {
	node := FindOne(root, expr)
	if node == nil {
		return ""
	}
	return node.GetText()
}

// FindAll returns all matching nodes.
func FindAll(root *parser.Node, expr string) []*parser.Node {
	return Find(root, expr)
}

func executeStep(node *parser.Node, step Step) []*parser.Node {
	var candidates []*parser.Node

	switch step.Axis {
	case "parent":
		if node.Parent != nil {
			candidates = []*parser.Node{node.Parent}
		}
	case "self":
		candidates = []*parser.Node{node}
	case "descendant":
		candidates = getDescendants(node)
	default: // child
		candidates = getChildren(node)
	}

	// Filter by name
	var result []*parser.Node
	for _, c := range candidates {
		if step.Name == "*" || c.Name == step.Name {
			// Apply predicate
			if step.Predicate == "" || matchPredicate(c, step.Predicate) {
				result = append(result, c)
			}
		}
	}

	return result
}

func getChildren(node *parser.Node) []*parser.Node {
	var children []*parser.Node
	for _, child := range node.Children {
		if child.Type == parser.NodeElement {
			children = append(children, child)
		}
	}
	return children
}

func getDescendants(node *parser.Node) []*parser.Node {
	var result []*parser.Node
	for _, child := range node.Children {
		if child.Type == parser.NodeElement {
			result = append(result, child)
			result = append(result, getDescendants(child)...)
		}
	}
	return result
}

func matchPredicate(node *parser.Node, predicate string) bool {
	predicate = strings.TrimSpace(predicate)

	// Attribute equality: @attr="value"
	if strings.HasPrefix(predicate, "@") {
		eqIdx := strings.Index(predicate, "=")
		if eqIdx >= 0 {
			attrName := predicate[1:eqIdx]
			expected := strings.Trim(strings.TrimSpace(predicate[eqIdx+1:]), "\"'")
			for _, attr := range node.Attributes {
				if attr.Name == attrName && attr.Value == expected {
					return true
				}
			}
			return false
		}
	}

	// Text content: .="text"
	if strings.HasPrefix(predicate, ".=") {
		expected := strings.Trim(strings.TrimSpace(predicate[2:]), "\"'")
		return node.GetText() == expected
	}

	return true
}

// String returns a summary of query results.
func String(results []*parser.Node) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d node(s):\n", len(results)))
	for i, node := range results {
		sb.WriteString(fmt.Sprintf("  %d. <%s>%s</%s>\n", i+1, node.Name, truncateText(node.GetText(), 50), node.Name))
	}
	return sb.String()
}

func truncateText(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
