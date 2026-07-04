// Package stats provides XML document statistics and analysis.
package stats

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

// Stats holds statistics about an XML document.
type Stats struct {
	TotalElements      int
	MaxDepth           int
	TotalAttributes    int
	TotalTextNodes     int
	TotalComments      int
	TotalCDATA         int
	TotalPIs           int
	TotalNamespaces    int
	ElementCounts      map[string]int
	AttributeCounts    map[string]int
	NamespacePrefixes  map[string]string
	AvgChildrenPerNode float64
	AvgTextLength      float64
}

// Analyze computes statistics for a parsed XML document.
func Analyze(doc *parser.Document) *Stats {
	s := &Stats{
		ElementCounts:     make(map[string]int),
		AttributeCounts:   make(map[string]int),
		NamespacePrefixes: make(map[string]string),
	}

	// Copy namespaces
	for k, v := range doc.Namespaces {
		s.NamespacePrefixes[k] = v
		s.TotalNamespaces++
	}

	// Copy prolog info
	if doc.Prolog != nil {
		s.TotalComments = len(doc.Prolog.Comments)
	}

	// Analyze root element
	if doc.Root != nil {
		analyzeNode(doc.Root, s, 0)
	}

	// Calculate averages
	if s.TotalElements > 0 {
		s.AvgChildrenPerNode = float64(countAllChildren(doc.Root)) / float64(s.TotalElements)
	}
	if s.TotalTextNodes > 0 {
		s.AvgTextLength = float64(countTotalTextLength(doc.Root)) / float64(s.TotalTextNodes)
	}

	return s
}

func analyzeNode(node *parser.Node, s *Stats, depth int) {
	s.TotalElements++
	s.ElementCounts[node.Name]++

	if depth+1 > s.MaxDepth {
		s.MaxDepth = depth + 1
	}

	s.TotalAttributes += len(node.Attributes)
	for _, attr := range node.Attributes {
		key := attr.Name
		if attr.Prefix != "" {
			key = attr.Prefix + ":" + attr.Name
		}
		s.AttributeCounts[key]++
	}

	for _, child := range node.Children {
		switch child.Type {
		case parser.NodeText:
			s.TotalTextNodes++
		case parser.NodeComment:
			s.TotalComments++
		case parser.NodeInstruction:
			s.TotalPIs++
		case parser.NodeElement:
			analyzeNode(child, s, depth+1)
		}
	}
}

func countAllChildren(node *parser.Node) int {
	count := 0
	for _, child := range node.Children {
		if child.Type == parser.NodeElement {
			count++
			count += countAllChildren(child)
		}
	}
	return count
}

func countTotalTextLength(node *parser.Node) int {
	length := 0
	for _, child := range node.Children {
		if child.Type == parser.NodeText {
			length += len(child.Value)
		} else if child.Type == parser.NodeElement {
			length += countTotalTextLength(child)
		}
	}
	return length
}

// Format returns a human-readable string of the statistics.
func (s *Stats) Format() string {
	var sb strings.Builder

	sb.WriteString("=== XML Document Statistics ===\n\n")
	sb.WriteString(fmt.Sprintf("Total Elements:       %d\n", s.TotalElements))
	sb.WriteString(fmt.Sprintf("Max Depth:            %d\n", s.MaxDepth))
	sb.WriteString(fmt.Sprintf("Total Attributes:     %d\n", s.TotalAttributes))
	sb.WriteString(fmt.Sprintf("Total Text Nodes:     %d\n", s.TotalTextNodes))
	sb.WriteString(fmt.Sprintf("Total Comments:       %d\n", s.TotalComments))
	sb.WriteString(fmt.Sprintf("Total CDATA:          %d\n", s.TotalCDATA))
	sb.WriteString(fmt.Sprintf("Total PIs:            %d\n", s.TotalPIs))
	sb.WriteString(fmt.Sprintf("Total Namespaces:     %d\n", s.TotalNamespaces))
	sb.WriteString(fmt.Sprintf("Avg Children/Node:    %.2f\n", s.AvgChildrenPerNode))
	sb.WriteString(fmt.Sprintf("Avg Text Length:      %.2f\n\n", s.AvgTextLength))

	if len(s.ElementCounts) > 0 {
		sb.WriteString("Element Counts:\n")
		for name, count := range s.ElementCounts {
			sb.WriteString(fmt.Sprintf("  %-30s %d\n", name, count))
		}
		sb.WriteString("\n")
	}

	if len(s.AttributeCounts) > 0 {
		sb.WriteString("Attribute Counts:\n")
		for name, count := range s.AttributeCounts {
			sb.WriteString(fmt.Sprintf("  %-30s %d\n", name, count))
		}
		sb.WriteString("\n")
	}

	if len(s.NamespacePrefixes) > 0 {
		sb.WriteString("Namespaces:\n")
		for prefix, uri := range s.NamespacePrefixes {
			if prefix == "" {
				prefix = "(default)"
			}
			sb.WriteString(fmt.Sprintf("  %-20s %s\n", prefix, uri))
		}
	}

	return sb.String()
}

// Summary returns a compact summary string.
func (s *Stats) Summary() string {
	return fmt.Sprintf(
		"elements=%d depth=%d attrs=%d text=%d comments=%d namespaces=%d",
		s.TotalElements, s.MaxDepth, s.TotalAttributes,
		s.TotalTextNodes, s.TotalComments, s.TotalNamespaces,
	)
}
