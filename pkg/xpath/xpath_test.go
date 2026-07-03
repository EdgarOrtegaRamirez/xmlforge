package xpath

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
)

const testXML = `<library>
  <book id="1">
    <title>Go Programming</title>
    <author>John Doe</author>
  </book>
  <book id="2">
    <title>XML Processing</title>
    <author>Jane Smith</author>
  </book>
</library>`

func TestCompile(t *testing.T) {
	q, err := Compile("library/book")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(q.Steps))
	}
}

func TestCompileAbsolutePath(t *testing.T) {
	q, err := Compile("/library/book")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(q.Steps))
	}
}

func TestCompileEmpty(t *testing.T) {
	_, err := Compile("")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestFind(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := Find(doc.Root, "book")
	if len(results) != 2 {
		t.Errorf("expected 2 books, got %d", len(results))
	}
}

func TestFindDeep(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := Find(doc.Root, "book/title")
	if len(results) != 2 {
		t.Errorf("expected 2 titles, got %d", len(results))
	}
}

func TestFindOne(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node := FindOne(doc.Root, "book")
	if node == nil {
		t.Fatal("expected to find a book")
	}
	if node.Name != "book" {
		t.Errorf("expected 'book', got %q", node.Name)
	}
}

func TestFindText(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := FindText(doc.Root, "book/title")
	if text != "Go Programming" {
		t.Errorf("expected 'Go Programming', got %q", text)
	}
}

func TestFindWithPredicate(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := Find(doc.Root, `book[@id="1"]`)
	if len(results) != 1 {
		t.Errorf("expected 1 book, got %d", len(results))
	}
}

func TestExecute(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := Execute(doc, "book/author")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 authors, got %d", len(results))
	}
}

func TestFindAll(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := FindAll(doc.Root, "book")
	if len(results) != 2 {
		t.Errorf("expected 2 books, got %d", len(results))
	}
}

func TestString(t *testing.T) {
	doc, err := parser.ParseString(testXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := Find(doc.Root, "book")
	output := String(results)
	if !strings.Contains(output, "Found 2 node(s)") {
		t.Error("expected output to contain node count")
	}
}

func TestMustCompile(t *testing.T) {
	q := MustCompile("book")
	if q == nil {
		t.Error("expected non-nil query")
	}
}

func TestMustCompilePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty expression")
		}
	}()
	MustCompile("")
}
