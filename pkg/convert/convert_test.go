package convert

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

func TestToJSON(t *testing.T) {
	doc, err := parser.ParseString(`<root><name>John</name><age>30</age></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := ToJSON(doc, DefaultJSONOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "John") {
		t.Error("expected JSON to contain 'John'")
	}
	if !strings.Contains(result, "name") {
		t.Error("expected JSON to contain 'name'")
	}
}

func TestToJSONWithAttributes(t *testing.T) {
	doc, err := parser.ParseString(`<root><person id="1"><name>John</name></person></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := ToJSON(doc, DefaultJSONOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "@id") {
		t.Error("expected JSON to contain '@id' attribute")
	}
}

func TestToJSONCompact(t *testing.T) {
	doc, err := parser.ParseString(`<root><name>John</name></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	opts := DefaultJSONOptions()
	opts.Compact = true
	result, err := ToJSON(doc, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Compact should not have indentation
	if strings.Contains(result, "\n  ") {
		t.Error("expected compact JSON without indentation")
	}
}

func TestToJSONBytes(t *testing.T) {
	doc, err := parser.ParseString(`<root><name>John</name></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := ToJSONBytes(doc, DefaultJSONOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestToCSV(t *testing.T) {
	doc, err := parser.ParseString(`<users><user name="John" age="30"/><user name="Jane" age="25"/></users>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := ToCSV(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "name") {
		t.Error("expected CSV to contain 'name' header")
	}
	if !strings.Contains(result, "John") {
		t.Error("expected CSV to contain 'John'")
	}
}

func TestToCSVEmpty(t *testing.T) {
	doc, err := parser.ParseString(`<root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ToCSV(doc)
	// Empty root should still work (just headers)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
