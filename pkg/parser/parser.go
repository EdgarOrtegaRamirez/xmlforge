package parser

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// NodeType represents the type of an XML node.
type NodeType int

const (
	NodeElement NodeType = iota
	NodeText
	NodeComment
	NodeInstruction
	NodeDocument
)

// Node represents a node in the XML DOM tree.
type Node struct {
	Type       NodeType
	Name       string
	Value      string
	Attributes []*Attribute
	Children   []*Node
	Parent     *Node
	Namespace  string
	Prefix     string
	Line       int
	Column     int
}

// Attribute represents an XML attribute.
type Attribute struct {
	Name      string
	Value     string
	Namespace string
	Prefix    string
}

// Document represents a complete XML document.
type Document struct {
	Root       *Node
	Prolog     *Prolog
	Errors     []error
	Namespaces map[string]string
}

// Prolog represents the XML declaration and processing instructions.
type Prolog struct {
	Version    string
	Encoding   string
	Standalone string
	Comments   []string
}

// Parser parses XML content into a DOM tree.
type Parser struct {
	reader   *RuneReader
	doc      *Document
	current  *Node
	stack    []*Node
	entities map[string]string
	lenient  bool
	line     int
	col      int
}

// Option configures the parser.
type Option func(*Parser)

// WithLenient enables lenient parsing.
func WithLenient() Option {
	return func(p *Parser) {
		p.lenient = true
	}
}

// New creates a new Parser.
func New(opts ...Option) *Parser {
	p := &Parser{
		line: 1,
		col:  1,
		entities: map[string]string{
			"amp":  "&",
			"lt":   "<",
			"gt":   ">",
			"apos": "'",
			"quot": "\"",
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Parse parses XML from a reader.
func (p *Parser) Parse(r io.Reader) (*Document, error) {
	p.reader = NewRuneReader(r)
	p.doc = &Document{
		Prolog:     &Prolog{},
		Namespaces: make(map[string]string),
	}
	p.stack = make([]*Node, 0)

	p.skipBOM()
	p.skipWhitespace()

	if err := p.parseProlog(); err != nil && !p.lenient {
		return nil, err
	}

	p.skipWhitespace()

	if err := p.parseElement(); err != nil {
		if len(p.doc.Errors) > 0 && p.lenient {
			return p.doc, nil
		}
		return nil, err
	}

	p.skipWhitespace()

	ch, err := p.peek()
	if err == nil && ch != -1 && ch != -2 {
		p.doc.Errors = append(p.doc.Errors, fmt.Errorf("unexpected content after root element"))
		if !p.lenient {
			return nil, p.doc.Errors[0]
		}
	}

	return p.doc, nil
}

// ParseString parses XML from a string.
func ParseString(xml string, opts ...Option) (*Document, error) {
	p := New(opts...)
	return p.Parse(strings.NewReader(xml))
}

func (p *Parser) skipBOM() {
	ch, err := p.peek()
	if err == nil && ch == '\uFEFF' {
		p.read()
	}
}

func (p *Parser) parseProlog() error {
	ch, err := p.peek()
	if err != nil || ch != '<' {
		return nil
	}

	p.read() // consume '<'
	nextCh, err := p.peek()
	if err != nil {
		return nil
	}

	switch {
	case nextCh == '?':
		p.read() // consume '?'
		content := p.readUntil("?>")
		p.parseXMLDeclaration(content)

	case nextCh == '!':
		p.read() // consume '!'
		ch2, err := p.peek()
		if err != nil {
			return nil
		}
		if ch2 == '-' {
			p.read() // consume '-'
			ch3, _ := p.peek()
			if ch3 == '-' {
				p.read() // consume '-'
				content := p.readUntil("-->")
				p.doc.Prolog.Comments = append(p.doc.Prolog.Comments, strings.TrimSpace(content))
			}
		} else if ch2 == 'D' || ch2 == 'd' {
			p.skipUntil(">")
			p.read() // '>'
		}

	case isNameStartChar(nextCh):
		// Root element — put '<' back
		p.unread('<')
		return nil

	default:
		if p.lenient {
			p.unread('<')
			return nil
		}
		return fmt.Errorf("unexpected character '%c' after '<' at line %d, col %d", nextCh, p.line, p.col)
	}

	return nil
}

func (p *Parser) parseXMLDeclaration(content string) {
	if idx := strings.Index(content, "version="); idx >= 0 {
		start := idx + 8
		if start < len(content) {
			quote := content[start]
			if quote == '"' || quote == '\'' {
				end := strings.IndexByte(content[start+1:], quote)
				if end >= 0 {
					p.doc.Prolog.Version = content[start+1 : start+1+end]
				}
			}
		}
	}
	if idx := strings.Index(content, "encoding="); idx >= 0 {
		start := idx + 9
		if start < len(content) {
			quote := content[start]
			if quote == '"' || quote == '\'' {
				end := strings.IndexByte(content[start+1:], quote)
				if end >= 0 {
					p.doc.Prolog.Encoding = content[start+1 : start+1+end]
				}
			}
		}
	}
	if idx := strings.Index(content, "standalone="); idx >= 0 {
		start := idx + 11
		if start < len(content) {
			quote := content[start]
			if quote == '"' || quote == '\'' {
				end := strings.IndexByte(content[start+1:], quote)
				if end >= 0 {
					p.doc.Prolog.Standalone = content[start+1 : start+1+end]
				}
			}
		}
	}
}

func (p *Parser) unread(ch rune) {
	p.reader.Unread(ch)
	p.col--
	if p.col < 1 {
		p.line--
		p.col = 1
	}
}

func (p *Parser) parseElement() error {
	p.skipWhitespace()

	ch, err := p.peek()
	if err != nil {
		return fmt.Errorf("unexpected end of input")
	}
	if ch == -1 || ch == -2 {
		return fmt.Errorf("unexpected end of input, expected '<'")
	}
	if ch != '<' {
		return fmt.Errorf("expected '<' at line %d, col %d, got '%c'", p.line, p.col, ch)
	}

	p.read() // consume '<'

	name, err := p.readName()
	if err != nil {
		return fmt.Errorf("reading element name: %w", err)
	}

	node := &Node{
		Type:   NodeElement,
		Name:   name,
		Line:   p.line,
		Column: p.col,
	}

	if idx := strings.IndexByte(name, ':'); idx >= 0 {
		node.Prefix = name[:idx]
		node.Name = name[idx+1:]
		node.Namespace = p.resolveNamespace(node.Prefix)
	}

	// Parse attributes
	selfClosing := false
	for {
		p.skipWhitespace()
		ch, err = p.peek()
		if err != nil {
			return fmt.Errorf("unexpected end of input while parsing attributes")
		}

		if ch == '/' {
			p.read() // consume '/'
			selfClosing = true
			ch, err = p.peek()
			if err == nil && ch == '>' {
				p.read() // consume '>'
			} else if !p.lenient {
				return fmt.Errorf("expected '>' after '/' at line %d, col %d", p.line, p.col)
			} else {
				p.skipUntil(">")
				p.read()
			}
			break
		}

		if ch == '>' {
			p.read() // consume '>'
			break
		}

		attr, err := p.parseAttribute()
		if err != nil {
			if p.lenient {
				p.skipUntil("=>")
				continue
			}
			return fmt.Errorf("parsing attribute: %w", err)
		}

		// Handle xmlns declarations
		if attr.Prefix == "xmlns" {
			p.doc.Namespaces[attr.Name] = attr.Value
		} else if attr.Name == "xmlns" {
			p.doc.Namespaces[""] = attr.Value
		}

		node.Attributes = append(node.Attributes, attr)
	}

	// Set parent and add to document
	node.Parent = p.current
	if p.doc.Root == nil {
		p.doc.Root = node
	}

	if len(p.stack) > 0 {
		p.stack[len(p.stack)-1].Children = append(p.stack[len(p.stack)-1].Children, node)
	}

	if !selfClosing {
		p.stack = append(p.stack, node)
		p.current = node

		if err := p.parseChildren(node); err != nil {
			return err
		}

		if len(p.stack) > 0 {
			p.stack = p.stack[:len(p.stack)-1]
		}
		if len(p.stack) > 0 {
			p.current = p.stack[len(p.stack)-1]
		} else {
			p.current = nil
		}
	}

	return nil
}

func (p *Parser) parseChildren(parent *Node) error {
	for {
		ch, err := p.peek()
		if err != nil || ch == -1 || ch == -2 {
			if p.lenient {
				return nil
			}
			return fmt.Errorf("unexpected end of input in element <%s>", parent.Name)
		}

		if ch == '<' {
			p.read() // consume '<'

			ch, err = p.peek()
			if err != nil {
				return fmt.Errorf("unexpected end of input")
			}

			if ch == '/' {
				// Closing tag
				p.read() // consume '/'

				name, err := p.readName()
				if err != nil {
					return fmt.Errorf("reading closing tag name: %w", err)
				}

				expectedName := parent.Name
				if parent.Prefix != "" {
					expectedName = parent.Prefix + ":" + parent.Name
				}
				if name != expectedName && !p.lenient {
					return fmt.Errorf("mismatched closing tag: expected </%s>, got </%s> at line %d, col %d",
						expectedName, name, p.line, p.col)
				}

				p.skipWhitespace()
				ch, _ = p.peek()
				if ch == '>' {
					p.read()
				} else if !p.lenient {
					return fmt.Errorf("expected '>' in closing tag at line %d, col %d", p.line, p.col)
				} else {
					p.skipUntil(">")
					p.read()
				}
				return nil
			}

			if ch == '!' {
				p.read() // consume '!'
				ch2, _ := p.peek()
				if ch2 == '-' {
					p.read() // '-'
					ch3, _ := p.peek()
					if ch3 == '-' {
						p.read() // '-'
						content := p.readUntil("-->")
						commentNode := &Node{
							Type:   NodeComment,
							Value:  content,
							Parent: parent,
							Line:   p.line,
							Column: p.col,
						}
						parent.Children = append(parent.Children, commentNode)
					}
				} else if ch2 == '[' {
					p.read() // '['
					p.readUntil("[")
					content := p.readUntil("]]>")
					textNode := &Node{
						Type:   NodeText,
						Value:  content,
						Parent: parent,
						Line:   p.line,
						Column: p.col,
					}
					parent.Children = append(parent.Children, textNode)
				}
			} else if ch == '?' {
				p.read() // '?'
				content := p.readUntil("?>")
				piNode := &Node{
					Type:   NodeInstruction,
					Value:  content,
					Parent: parent,
					Line:   p.line,
					Column: p.col,
				}
				parent.Children = append(parent.Children, piNode)
			} else {
				// Child element — put '<' back and parse
				p.unread('<')
				if err := p.parseElement(); err != nil {
					return err
				}
			}
		} else {
			// Text content — read until '<' WITHOUT consuming it
			text := p.readTextUntil("<")
			if text != "" {
				text = p.decodeEntities(text)
				// Skip whitespace-only text nodes between elements
				if strings.TrimSpace(text) == "" {
					// Check if next content is an element (not the closing tag)
					peeked, _ := p.peek()
					if peeked == '<' {
						// Skip whitespace-only text between elements
						continue
					}
				}
				textNode := &Node{
					Type:   NodeText,
					Value:  text,
					Parent: parent,
					Line:   p.line,
					Column: p.col,
				}
				parent.Children = append(parent.Children, textNode)
			}
		}
	}
}

// readTextUntil reads text until the delimiter is found, WITHOUT consuming the delimiter.
func (p *Parser) readTextUntil(delimiter string) string {
	var result strings.Builder

	for {
		ch, err := p.peek()
		if err != nil || ch == -1 || ch == -2 {
			break
		}
		if ch == rune(delimiter[0]) {
			break // Don't consume the delimiter
		}
		result.WriteRune(ch)
		p.read()
	}

	return result.String()
}

func (p *Parser) decodeEntities(text string) string {
	result := text
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&apos;", "'")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	return result
}

func (p *Parser) parseAttribute() (*Attribute, error) {
	name, err := p.readName()
	if err != nil {
		return nil, fmt.Errorf("reading attribute name: %w", err)
	}

	p.skipWhitespace()

	ch, err := p.peek()
	if err != nil || ch != '=' {
		if p.lenient {
			return &Attribute{Name: name, Value: ""}, nil
		}
		return nil, fmt.Errorf("expected '=' after attribute name at line %d, col %d", p.line, p.col)
	}
	p.read() // consume '='

	p.skipWhitespace()

	quote, err := p.peek()
	if err != nil || (quote != '"' && quote != '\'') {
		if p.lenient {
			return &Attribute{Name: name, Value: ""}, nil
		}
		return nil, fmt.Errorf("expected quote at line %d, col %d", p.line, p.col)
	}
	p.read() // consume opening quote

	var value strings.Builder
	for {
		ch, err = p.peek()
		if err != nil || ch == -1 || ch == -2 {
			break
		}
		if ch == rune(quote) {
			break
		}
		if ch == '&' {
			p.read() // consume '&'
			entityName := ""
			for {
				ch, err = p.peek()
				if err != nil || ch == ';' {
					break
				}
				entityName += string(ch)
				p.read()
			}
			if ch == ';' {
				p.read() // consume ';'
			}
			if replacement, ok := p.entities[entityName]; ok {
				value.WriteString(replacement)
			} else {
				value.WriteString("&" + entityName + ";")
			}
		} else {
			value.WriteRune(ch)
			p.read()
		}
	}

	if ch == rune(quote) {
		p.read() // consume closing quote
	}

	attr := &Attribute{
		Name:  name,
		Value: value.String(),
	}

	if idx := strings.IndexByte(name, ':'); idx >= 0 {
		attr.Prefix = name[:idx]
		attr.Name = name[idx+1:]
		attr.Namespace = p.resolveNamespace(attr.Prefix)
	}

	return attr, nil
}

func (p *Parser) resolveNamespace(prefix string) string {
	if ns, ok := p.doc.Namespaces[prefix]; ok {
		return ns
	}
	return ""
}

func (p *Parser) readName() (string, error) {
	var name strings.Builder

	ch, err := p.peek()
	if err != nil {
		return "", fmt.Errorf("unexpected end of input")
	}

	if !isNameStartChar(ch) {
		return "", fmt.Errorf("invalid name start character '%c' at line %d, col %d", ch, p.line, p.col)
	}

	for {
		ch, err = p.peek()
		if err != nil || !isNameChar(ch) {
			break
		}
		name.WriteRune(ch)
		p.read()
	}

	if name.Len() == 0 {
		return "", fmt.Errorf("empty name at line %d, col %d", p.line, p.col)
	}

	return name.String(), nil
}

func isNameStartChar(ch rune) bool {
	return ch == ':' || ch == '_' ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= 0xC0 && ch <= 0xD6) ||
		(ch >= 0xD8 && ch <= 0xF6) ||
		(ch >= 0xF8 && ch <= 0x2FF) ||
		(ch >= 0x370 && ch <= 0x37D) ||
		(ch >= 0x37F && ch <= 0x1FFF) ||
		(ch >= 0x200C && ch <= 0x200D) ||
		(ch >= 0x2070 && ch <= 0x218F) ||
		(ch >= 0x2C00 && ch <= 0x2FEF) ||
		(ch >= 0x3001 && ch <= 0xD7FF) ||
		(ch >= 0xF900 && ch <= 0xFDCF) ||
		(ch >= 0xFDF0 && ch <= 0xFFFD) ||
		(ch >= 0x10000 && ch <= 0xEFFFF)
}

func isNameChar(ch rune) bool {
	return isNameStartChar(ch) ||
		ch == '-' || ch == '.' ||
		(ch >= '0' && ch <= '9') ||
		ch == 0xB7 ||
		(ch >= 0x0300 && ch <= 0x036F) ||
		(ch >= 0x203F && ch <= 0x2040)
}

func (p *Parser) readUntil(delimiter string) string {
	var result strings.Builder
	matchIdx := 0

	for {
		ch, err := p.peek()
		if err != nil || ch == -1 || ch == -2 {
			break
		}

		result.WriteRune(ch)
		p.read()

		if ch == rune(delimiter[matchIdx]) {
			matchIdx++
			if matchIdx == len(delimiter) {
				s := result.String()
				return s[:len(s)-len(delimiter)]
			}
		} else {
			matchIdx = 0
		}
	}

	return result.String()
}

func (p *Parser) skipWhitespace() {
	for {
		ch, err := p.peek()
		if err != nil || ch == -1 || ch == -2 {
			return
		}
		if !unicode.IsSpace(ch) {
			return
		}
		p.read()
	}
}

func (p *Parser) skipUntil(s string) {
	matchIdx := 0
	for {
		ch, err := p.peek()
		if err != nil || ch == -1 || ch == -2 {
			return
		}
		p.read()
		if matchIdx < len(s) && ch == rune(s[matchIdx]) {
			matchIdx++
			if matchIdx == len(s) {
				return
			}
		} else {
			matchIdx = 0
		}
	}
}

func (p *Parser) peek() (rune, error) {
	return p.reader.Peek()
}

func (p *Parser) read() rune {
	ch, err := p.reader.Read()
	if err != nil {
		return -1
	}
	if ch == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return ch
}

// FindElements finds all elements matching a name.
func (doc *Document) FindElements(name string) []*Node {
	var results []*Node
	findByName(doc.Root, name, &results)
	return results
}

// FindElement finds the first element matching a name.
func (doc *Document) FindElement(name string) *Node {
	results := doc.FindElements(name)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// GetText returns the text content of an element.
func (n *Node) GetText() string {
	var result strings.Builder
	for _, child := range n.Children {
		if child.Type == NodeText {
			result.WriteString(child.Value)
		} else if child.Type == NodeElement {
			result.WriteString(child.GetText())
		}
	}
	return result.String()
}

// GetAttribute returns the value of an attribute by name.
func (n *Node) GetAttribute(name string) (string, bool) {
	for _, attr := range n.Attributes {
		if attr.Name == name {
			return attr.Value, true
		}
	}
	return "", false
}

// ChildrenByName returns children with a specific element name.
func (n *Node) ChildrenByName(name string) []*Node {
	var results []*Node
	for _, child := range n.Children {
		if child.Type == NodeElement && child.Name == name {
			results = append(results, child)
		}
	}
	return results
}

// ChildByName returns the first child with a specific element name.
func (n *Node) ChildByName(name string) *Node {
	for _, child := range n.Children {
		if child.Type == NodeElement && child.Name == name {
			return child
		}
	}
	return nil
}

func findByName(node *Node, name string, results *[]*Node) {
	if node == nil {
		return
	}
	if node.Type == NodeElement && node.Name == name {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		findByName(child, name, results)
	}
}

// String returns a string representation of the node tree.
func (n *Node) String() string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	n.writeString(&b, 0)
	return b.String()
}

func (n *Node) writeString(b *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n.Type {
	case NodeElement:
		b.WriteString(indent)
		b.WriteString("<")
		b.WriteString(n.Name)
		for _, attr := range n.Attributes {
			b.WriteString(" ")
			b.WriteString(attr.Name)
			b.WriteString("=\"")
			b.WriteString(attr.Value)
			b.WriteString("\"")
		}

		if len(n.Children) == 0 {
			b.WriteString("/>")
		} else {
			b.WriteString(">")
			hasChildElement := false
			for _, child := range n.Children {
				if child.Type == NodeElement {
					hasChildElement = true
					break
				}
			}
			if hasChildElement {
				b.WriteString("\n")
				for _, child := range n.Children {
					child.writeString(b, depth+1)
				}
				b.WriteString(indent)
			} else {
				for _, child := range n.Children {
					child.writeString(b, 0)
				}
			}
			b.WriteString("</")
			b.WriteString(n.Name)
			b.WriteString(">")
			if hasChildElement {
				b.WriteString("\n")
			}
		}
	case NodeText:
		b.WriteString(n.Value)
	case NodeComment:
		b.WriteString(indent)
		b.WriteString("<!--")
		b.WriteString(n.Value)
		b.WriteString("-->")
		b.WriteString("\n")
	}
}
