package stats

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

func TestAnalyze(t *testing.T) {
	doc, err := parser.ParseString(`<root xmlns:ns="http://example.com"><child ns:attr="val">text</child><!-- comment --></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := Analyze(doc)
	if s.TotalElements != 2 {
		t.Errorf("expected 2 elements, got %d", s.TotalElements)
	}
	if s.MaxDepth != 2 {
		t.Errorf("expected max depth 2, got %d", s.MaxDepth)
	}
	if s.TotalAttributes != 2 {
		t.Errorf("expected 2 attributes (xmlns:ns + ns:attr), got %d", s.TotalAttributes)
	}
	if s.TotalTextNodes != 1 {
		t.Errorf("expected 1 text node, got %d", s.TotalTextNodes)
	}
	if s.TotalComments != 1 {
		t.Errorf("expected 1 comment, got %d", s.TotalComments)
	}
	if s.TotalNamespaces != 1 {
		t.Errorf("expected 1 namespace, got %d", s.TotalNamespaces)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	doc, err := parser.ParseString("<root/>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := Analyze(doc)
	if s.TotalElements != 1 {
		t.Errorf("expected 1 element, got %d", s.TotalElements)
	}
	if s.MaxDepth != 1 {
		t.Errorf("expected max depth 1, got %d", s.MaxDepth)
	}
}

func TestFormat(t *testing.T) {
	doc, err := parser.ParseString(`<root><child a="b">text</child></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := Analyze(doc)
	result := s.Format()
	if !strings.Contains(result, "Total Elements:") {
		t.Error("expected formatted output to contain 'Total Elements:'")
	}
	if !strings.Contains(result, "child") {
		t.Error("expected formatted output to contain element names")
	}
}

func TestSummary(t *testing.T) {
	doc, err := parser.ParseString(`<root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := Analyze(doc)
	result := s.Summary()
	if !strings.HasPrefix(result, "elements=") {
		t.Errorf("expected summary to start with 'elements=', got %q", result)
	}
}
