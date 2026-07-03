package diff

import (
	"strings"
	"testing"
)

func TestCompareIdentical(t *testing.T) {
	xml := `<root><child>text</child></root>`
	result, err := CompareStrings(xml, xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Identical {
		t.Error("expected documents to be identical")
	}
}

func TestCompareDifferentText(t *testing.T) {
	xml1 := `<root><child>text1</child></root>`
	xml2 := `<root><child>text2</child></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identical {
		t.Error("expected documents to be different")
	}
	if len(result.Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(result.Differences))
	}
}

func TestCompareAddedNode(t *testing.T) {
	xml1 := `<root><child>text</child></root>`
	xml2 := `<root><child>text</child><new>new</new></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identical {
		t.Error("expected documents to be different")
	}
	found := false
	for _, d := range result.Differences {
		if d.Type == ChangeAdded && d.Message == "child <new> added" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find added <new> node")
	}
}

func TestCompareRemovedNode(t *testing.T) {
	xml1 := `<root><child>text</child><old>old</old></root>`
	xml2 := `<root><child>text</child></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identical {
		t.Error("expected documents to be different")
	}
	found := false
	for _, d := range result.Differences {
		if d.Type == ChangeRemoved && d.Message == "child <old> removed" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find removed <old> node")
	}
}

func TestCompareDifferentAttributes(t *testing.T) {
	xml1 := `<root><child id="1">text</child></root>`
	xml2 := `<root><child id="2">text</child></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identical {
		t.Error("expected documents to be different")
	}
	found := false
	for _, d := range result.Differences {
		if d.Type == ChangeModified && strings.Contains(d.Message, "@id") {
			found = true
		}
	}
	if !found {
		t.Error("expected to find modified @id attribute")
	}
}

func TestFormat(t *testing.T) {
	xml1 := `<root><child>text1</child></root>`
	xml2 := `<root><child>text2</child></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := result.Format()
	if !strings.Contains(output, "1 difference") {
		t.Error("expected formatted output to contain difference count")
	}
}

func TestSummary(t *testing.T) {
	xml1 := `<root><child>text1</child></root>`
	xml2 := `<root><child>text2</child></root>`
	result, err := CompareStrings(xml1, xml2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summary := result.Summary()
	if !strings.Contains(summary, "modified=1") {
		t.Errorf("expected summary to contain 'modified=1', got %q", summary)
	}
}

func TestIdenticalSummary(t *testing.T) {
	xml := `<root><child>text</child></root>`
	result, err := CompareStrings(xml, xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Summary() != "identical" {
		t.Errorf("expected 'identical' summary, got %q", result.Summary())
	}
}
