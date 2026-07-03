package validator

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

func TestValidateValid(t *testing.T) {
	doc, err := parser.ParseString(`<root><child>text</child></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := New(DefaultRules())
	result := v.Validate(doc)
	if !result.Valid {
		t.Errorf("expected valid document, got errors: %v", result.Errors)
	}
}

func TestValidateNoRoot(t *testing.T) {
	doc, err := parser.ParseString(`<root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	doc.Root = nil

	rules := Rules{RequireRoot: true}
	v := New(rules)
	result := v.Validate(doc)
	if result.Valid {
		t.Error("expected invalid document")
	}
}

func TestValidateMaxDepth(t *testing.T) {
	doc, err := parser.ParseString(`<a><b><c><d>text</d></c></b></a>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{MaxDepth: 2}
	v := New(rules)
	result := v.Validate(doc)
	if result.Valid {
		t.Error("expected invalid document due to depth")
	}
}

func TestValidateMaxAttributes(t *testing.T) {
	doc, err := parser.ParseString(`<root a="1" b="2" c="3"><child>text</child></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{MaxAttributes: 2}
	v := New(rules)
	result := v.Validate(doc)
	if result.Valid {
		t.Error("expected invalid document due to too many attributes")
	}
}

func TestValidateAllowedElements(t *testing.T) {
	doc, err := parser.ParseString(`<root><allowed>text</allowed></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{AllowedElements: []string{"root", "allowed"}}
	v := New(rules)
	result := v.Validate(doc)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateForbiddenElements(t *testing.T) {
	doc, err := parser.ParseString(`<root><forbidden>text</forbidden></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{ForbiddenElements: []string{"forbidden"}}
	v := New(rules)
	result := v.Validate(doc)
	if result.Valid {
		t.Error("expected invalid document due to forbidden element")
	}
}

func TestValidateRequiredAttributes(t *testing.T) {
	doc, err := parser.ParseString(`<root><item>text</item></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{
		RequiredAttrs: map[string][]string{
			"item": {"id"},
		},
	}
	v := New(rules)
	result := v.Validate(doc)
	if result.Valid {
		t.Error("expected invalid document due to missing required attribute")
	}
}

func TestValidateString(t *testing.T) {
	result, err := ValidateString(`<root><child>text</child></root>`, DefaultRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestCheckRequiredElements(t *testing.T) {
	doc, err := parser.ParseString(`<root><name>John</name></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errors := CheckRequiredElements(doc, []string{"name"})
	if len(errors) > 0 {
		t.Errorf("unexpected errors: %v", errors)
	}

	errors = CheckRequiredElements(doc, []string{"name", "age"})
	if len(errors) != 1 {
		t.Errorf("expected 1 error for missing 'age', got %d", len(errors))
	}
}

func TestFormat(t *testing.T) {
	doc, err := parser.ParseString(`<root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := New(DefaultRules())
	result := v.Validate(doc)
	output := result.Format()
	if !strings.Contains(output, "valid") {
		t.Error("expected output to contain 'valid'")
	}
}

func TestSummary(t *testing.T) {
	doc, err := parser.ParseString(`<root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := New(DefaultRules())
	result := v.Validate(doc)
	summary := result.Summary()
	if !strings.HasPrefix(summary, "valid") {
		t.Errorf("expected summary to start with 'valid', got %q", summary)
	}
}

func TestNoEmptyElements(t *testing.T) {
	doc, err := parser.ParseString(`<root><empty/></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{NoEmptyElements: true}
	v := New(rules)
	result := v.Validate(doc)
	// Should have warnings, not errors
	if !result.Valid {
		t.Error("expected document to be valid (with warnings)")
	}
	if len(result.Errors) == 0 {
		t.Error("expected warnings for empty elements")
	}
}

func TestValidateNames(t *testing.T) {
	doc, err := parser.ParseString(`<root><valid-name>text</valid-name></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules := Rules{ValidateNames: true}
	v := New(rules)
	result := v.Validate(doc)
	// Should pass validation
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}
