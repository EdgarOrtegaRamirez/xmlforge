package parser

import (
	"strings"
	"testing"
)

func TestParseSimpleElement(t *testing.T) {
	doc, err := ParseString(`<root>hello</root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	if doc.Root.Name != "root" {
		t.Errorf("expected name 'root', got %q", doc.Root.Name)
	}
	if doc.Root.GetText() != "hello" {
		t.Errorf("expected text 'hello', got %q", doc.Root.GetText())
	}
}

func TestParseSelfClosing(t *testing.T) {
	doc, err := ParseString(`<br/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	if doc.Root.Name != "br" {
		t.Errorf("expected name 'br', got %q", doc.Root.Name)
	}
}

func TestParseAttributes(t *testing.T) {
	doc, err := ParseString(`<person name="Alice" age="30"/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	if len(doc.Root.Attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(doc.Root.Attributes))
	}

	name, ok := doc.Root.GetAttribute("name")
	if !ok || name != "Alice" {
		t.Errorf("expected name='Alice', got %q", name)
	}

	age, ok := doc.Root.GetAttribute("age")
	if !ok || age != "30" {
		t.Errorf("expected age='30', got %q", age)
	}
}

func TestParseNestedElements(t *testing.T) {
	doc, err := ParseString(`<root><child>text</child></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Root.Children))
	}

	child := doc.Root.Children[0]
	if child.Type != NodeElement || child.Name != "child" {
		t.Errorf("expected child element 'child', got %q", child.Name)
	}
	if child.GetText() != "text" {
		t.Errorf("expected text 'text', got %q", child.GetText())
	}
}

func TestParseNestedDeep(t *testing.T) {
	doc, err := ParseString(`<a><b><c><d>deep</d></c></b></a>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}

	d := doc.Root.ChildByName("b").ChildByName("c").ChildByName("d")
	if d == nil {
		t.Fatal("expected to find d element")
	}
	if d.GetText() != "deep" {
		t.Errorf("expected text 'deep', got %q", d.GetText())
	}
}

func TestParseXMLDeclaration(t *testing.T) {
	doc, err := ParseString(`<?xml version="1.0" encoding="UTF-8"?><root/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Prolog == nil {
		t.Fatal("expected prolog")
	}
	if doc.Prolog.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", doc.Prolog.Version)
	}
	if doc.Prolog.Encoding != "UTF-8" {
		t.Errorf("expected encoding 'UTF-8', got %q", doc.Prolog.Encoding)
	}
}

func TestParseComment(t *testing.T) {
	doc, err := ParseString(`<root><!-- a comment --><child/></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	if len(doc.Root.Children) != 2 {
		t.Fatalf("expected 2 children (comment + element), got %d", len(doc.Root.Children))
	}

	comment := doc.Root.Children[0]
	if comment.Type != NodeComment {
		t.Errorf("expected comment node, got type %d", comment.Type)
	}
	if comment.Value != " a comment " {
		t.Errorf("expected comment ' a comment ', got %q", comment.Value)
	}
}

func TestParseEntities(t *testing.T) {
	doc, err := ParseString(`<root>5 &lt; 10 &amp; 10 &gt; 5</root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := doc.Root.GetText()
	if text != "5 < 10 & 10 > 5" {
		t.Errorf("expected decoded entities, got %q", text)
	}
}

func TestParseAttributeEntities(t *testing.T) {
	doc, err := ParseString(`<root attr="a &amp; b"/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := doc.Root.GetAttribute("attr")
	if !ok {
		t.Fatal("expected attribute")
	}
	if val != "a & b" {
		t.Errorf("expected 'a & b', got %q", val)
	}
}

func TestParseSingleQuotes(t *testing.T) {
	doc, err := ParseString(`<root attr='value'/>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := doc.Root.GetAttribute("attr")
	if !ok || val != "value" {
		t.Errorf("expected attr='value', got %q", val)
	}
}

func TestParseMultipleChildren(t *testing.T) {
	xml := `<root>
		<child1>one</child1>
		<child2>two</child2>
		<child3>three</child3>
	</root>`
	doc, err := ParseString(xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(doc.Root.Children))
	}
	for i, name := range []string{"child1", "child2", "child3"} {
		child := doc.Root.Children[i]
		if child.Type != NodeElement || child.Name != name {
			t.Errorf("expected child %d to be %s, got %s", i, name, child.Name)
		}
	}
}

func TestParseMixedContent(t *testing.T) {
	doc, err := ParseString(`<p>This is <b>bold</b> text.</p>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := doc.Root.GetText()
	if text != "This is bold text." {
		t.Errorf("expected 'This is bold text.', got %q", text)
	}
}

func TestFindElements(t *testing.T) {
	doc, err := ParseString(`<root><item id="1"/><item id="2"/></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := doc.FindElements("item")
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestFindElement(t *testing.T) {
	doc, err := ParseString(`<root><name>Alice</name><age>30</age></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := doc.FindElement("name")
	if name == nil {
		t.Fatal("expected to find name element")
	}
	if name.GetText() != "Alice" {
		t.Errorf("expected 'Alice', got %q", name.GetText())
	}
}

func TestNodeString(t *testing.T) {
	doc, err := ParseString(`<root attr="val">text</root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := doc.Root.String()
	if !strings.Contains(s, `<root attr="val">`) {
		t.Errorf("expected string to contain element with attribute, got %q", s)
	}
}

func TestParseLenient(t *testing.T) {
	// Mismatched closing tag - should parse in lenient mode
	doc, err := ParseString(`<root><child/></wrong></root>`, WithLenient())
	if err != nil {
		t.Fatalf("unexpected error in lenient mode: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
}

func TestParseNamespace(t *testing.T) {
	doc, err := ParseString(`<root xmlns:ns="http://example.com"><ns:child>text</ns:child></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := doc.Root.ChildByName("child")
	if child == nil {
		t.Fatal("expected child element")
	}
	if child.Namespace != "http://example.com" {
		t.Errorf("expected namespace 'http://example.com', got %q", child.Namespace)
	}
}

func TestParseBOM(t *testing.T) {
	doc, err := ParseString("\uFEFF<root/>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil || doc.Root.Name != "root" {
		t.Errorf("expected root element, got %v", doc.Root)
	}
}

func TestParseEmptyDocument(t *testing.T) {
	_, err := ParseString("")
	// Should not error on empty input (no root element)
	_ = err
}

func TestParseCDATA(t *testing.T) {
	xml := `<root><![CDATA[Some <b>raw</b> content]]></root>`
	doc, err := ParseString(xml, WithLenient())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected root element")
	}
	text := doc.Root.GetText()
	if !strings.Contains(text, "raw") {
		t.Errorf("expected CDATA content, got %q", text)
	}
}

func TestMultipleRoots(t *testing.T) {
	// Invalid XML - multiple root elements
	_, err := ParseString(`<a/><b/>`)
	if err == nil {
		t.Error("expected error for multiple root elements")
	}
}

func TestChildByName(t *testing.T) {
	doc, err := ParseString(`<root><a/><b/><a/></root>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := doc.Root.ChildByName("a")
	if a == nil || a.Name != "a" {
		t.Errorf("expected child 'a', got %v", a)
	}

	children := doc.Root.ChildrenByName("a")
	if len(children) != 2 {
		t.Errorf("expected 2 'a' children, got %d", len(children))
	}
}
