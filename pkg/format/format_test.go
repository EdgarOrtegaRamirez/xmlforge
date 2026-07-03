package format

import (
	"testing"
)

func TestFormat(t *testing.T) {
	xml := "<root><child>text</child></root>"
	result, err := Format(xml, DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty formatted output")
	}
}

func TestCompress(t *testing.T) {
	xml := `<root>
  <child>text</child>
</root>`
	result := Compress(xml)
	expected := "<root> <child>text</child> </root>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestMinify(t *testing.T) {
	xml := "<root>  <child>  text  </child>  </root>"
	result := Minify(xml)
	if result == "" {
		t.Error("expected non-empty minified output")
	}
}

func TestIndent(t *testing.T) {
	xml := "<root><child>text</child></root>"
	result := Indent(xml, "    ")
	if result == "" {
		t.Error("expected non-empty indented output")
	}
}

func TestValidateStructure(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{"valid", "<root><child/></root>", false},
		{"unclosed", "<root><child></root>", true},
		{"extra close", "<root></child></root>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateStructure(tt.xml)
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected errors but got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
		})
	}
}
