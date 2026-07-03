// Package diff provides XML document comparison and diffing utilities.
package diff

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

// ChangeType represents the type of change.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
	ChangeModified
	ChangeEqual
)

// Diff represents a single difference between XML nodes.
type Diff struct {
	Type     ChangeType
	Path     string
	Expected interface{}
	Actual   interface{}
	Message  string
}

// Result holds all differences between two XML documents.
type Result struct {
	Differences []Diff
	Identical   bool
}

// Compare compares two parsed XML documents.
func Compare(doc1, doc2 *parser.Document) *Result {
	result := &Result{
		Differences: make([]Diff, 0),
		Identical:   true,
	}

	compareNodes(doc1.Root, doc2.Root, "", result)
	return result
}

// CompareStrings parses and compares two XML strings.
func CompareStrings(xml1, xml2 string) (*Result, error) {
	doc1, err := parser.ParseString(xml1)
	if err != nil {
		return nil, fmt.Errorf("parsing first XML: %w", err)
	}

	doc2, err := parser.ParseString(xml2)
	if err != nil {
		return nil, fmt.Errorf("parsing second XML: %w", err)
	}

	return Compare(doc1, doc2), nil
}

func compareNodes(node1, node2 *parser.Node, path string, result *Result) {
	if node1 == nil && node2 == nil {
		return
	}

	if node1 == nil {
		result.Identical = false
		result.Differences = append(result.Differences, Diff{
			Type:    ChangeAdded,
			Path:    path,
			Actual:  nodeToString(node2),
			Message: fmt.Sprintf("node added: <%s>", node2.Name),
		})
		return
	}

	if node2 == nil {
		result.Identical = false
		result.Differences = append(result.Differences, Diff{
			Type:     ChangeRemoved,
			Path:     path,
			Expected: nodeToString(node1),
			Message:  fmt.Sprintf("node removed: <%s>", node1.Name),
		})
		return
	}

	// Compare names
	if node1.Name != node2.Name {
		result.Identical = false
		result.Differences = append(result.Differences, Diff{
			Type:     ChangeModified,
			Path:     path,
			Expected: node1.Name,
			Actual:   node2.Name,
			Message:  "element name changed",
		})
		return
	}

	nodePath := path + "/" + node1.Name

	// Compare attributes
	compareAttributes(node1, node2, nodePath, result)

	// Compare text content
	text1 := getDirectText(node1)
	text2 := getDirectText(node2)
	if text1 != text2 {
		result.Identical = false
		result.Differences = append(result.Differences, Diff{
			Type:     ChangeModified,
			Path:     nodePath,
			Expected: text1,
			Actual:   text2,
			Message:  "text content changed",
		})
	}

	// Compare children
	childMap1 := make(map[string][]*parser.Node)
	childMap2 := make(map[string][]*parser.Node)

	for _, child := range node1.Children {
		if child.Type == parser.NodeElement {
			childMap1[child.Name] = append(childMap1[child.Name], child)
		}
	}
	for _, child := range node2.Children {
		if child.Type == parser.NodeElement {
			childMap2[child.Name] = append(childMap2[child.Name], child)
		}
	}

	// Check for removed children
	for name, children := range childMap1 {
		if _, ok := childMap2[name]; !ok {
			for _, child := range children {
				result.Identical = false
				result.Differences = append(result.Differences, Diff{
					Type:     ChangeRemoved,
					Path:     nodePath + "/" + name,
					Expected: nodeToString(child),
					Message:  fmt.Sprintf("child <%s> removed", name),
				})
			}
		}
	}

	// Check for added children
	for name, children := range childMap2 {
		if _, ok := childMap1[name]; !ok {
			for _, child := range children {
				result.Identical = false
				result.Differences = append(result.Differences, Diff{
					Type:    ChangeAdded,
					Path:    nodePath + "/" + name,
					Actual:  nodeToString(child),
					Message: fmt.Sprintf("child <%s> added", name),
				})
			}
		}
	}

	// Compare matching children
	for name := range childMap1 {
		if children2, ok := childMap2[name]; ok {
			children1 := childMap1[name]
			maxLen := len(children1)
			if len(children2) > maxLen {
				maxLen = len(children2)
			}
			for i := 0; i < maxLen; i++ {
				var c1, c2 *parser.Node
				if i < len(children1) {
					c1 = children1[i]
				}
				if i < len(children2) {
					c2 = children2[i]
				}
				compareNodes(c1, c2, nodePath, result)
			}
		}
	}
}

func compareAttributes(node1, node2 *parser.Node, path string, result *Result) {
	attrMap1 := make(map[string]string)
	attrMap2 := make(map[string]string)

	for _, attr := range node1.Attributes {
		key := attr.Name
		if attr.Prefix != "" {
			key = attr.Prefix + ":" + attr.Name
		}
		attrMap1[key] = attr.Value
	}
	for _, attr := range node2.Attributes {
		key := attr.Name
		if attr.Prefix != "" {
			key = attr.Prefix + ":" + attr.Name
		}
		attrMap2[key] = attr.Value
	}

	// Check for removed attributes
	for key := range attrMap1 {
		if _, ok := attrMap2[key]; !ok {
			result.Identical = false
			result.Differences = append(result.Differences, Diff{
				Type:     ChangeRemoved,
				Path:     path + "/@" + key,
				Expected: attrMap1[key],
				Message:  fmt.Sprintf("attribute @%s removed", key),
			})
		}
	}

	// Check for added attributes
	for key := range attrMap2 {
		if _, ok := attrMap1[key]; !ok {
			result.Identical = false
			result.Differences = append(result.Differences, Diff{
				Type:    ChangeAdded,
				Path:    path + "/@" + key,
				Actual:  attrMap2[key],
				Message: fmt.Sprintf("attribute @%s added", key),
			})
		}
	}

	// Check for modified attributes
	for key := range attrMap1 {
		if val2, ok := attrMap2[key]; ok {
			if attrMap1[key] != val2 {
				result.Identical = false
				result.Differences = append(result.Differences, Diff{
					Type:     ChangeModified,
					Path:     path + "/@" + key,
					Expected: attrMap1[key],
					Actual:   val2,
					Message:  fmt.Sprintf("attribute @%s changed", key),
				})
			}
		}
	}
}

func getDirectText(node *parser.Node) string {
	var parts []string
	for _, child := range node.Children {
		if child.Type == parser.NodeText {
			parts = append(parts, child.Value)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func nodeToString(node *parser.Node) string {
	if node == nil {
		return ""
	}
	return fmt.Sprintf("<%s>%s</%s>", node.Name, getDirectText(node), node.Name)
}

// Format returns a human-readable diff output.
func (r *Result) Format() string {
	if r.Identical {
		return "Documents are identical."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d difference(s):\n\n", len(r.Differences)))

	for i, d := range r.Differences {
		var prefix string
		switch d.Type {
		case ChangeAdded:
			prefix = "+"
		case ChangeRemoved:
			prefix = "-"
		case ChangeModified:
			prefix = "~"
		default:
			prefix = " "
		}

		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, prefix, d.Message))
		sb.WriteString(fmt.Sprintf("   Path: %s\n", d.Path))
		if d.Expected != nil {
			sb.WriteString(fmt.Sprintf("   Expected: %v\n", d.Expected))
		}
		if d.Actual != nil {
			sb.WriteString(fmt.Sprintf("   Actual:   %v\n", d.Actual))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Summary returns a compact summary of the diff.
func (r *Result) Summary() string {
	if r.Identical {
		return "identical"
	}
	added, removed, modified := 0, 0, 0
	for _, d := range r.Differences {
		switch d.Type {
		case ChangeAdded:
			added++
		case ChangeRemoved:
			removed++
		case ChangeModified:
			modified++
		}
	}
	return fmt.Sprintf("%d differences (added=%d, removed=%d, modified=%d)",
		len(r.Differences), added, removed, modified)
}
