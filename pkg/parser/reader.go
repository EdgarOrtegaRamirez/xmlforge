package parser

import (
	"io"
	"unicode/utf8"
)

// RuneReader wraps an io.Reader and provides peek and unread functionality.
type RuneReader struct {
	reader io.Reader
	buf    []byte
	pos    int
}

// NewRuneReader creates a new RuneReader.
func NewRuneReader(r io.Reader) *RuneReader {
	return &RuneReader{reader: r}
}

// fillBuffer ensures there's at least one rune in the buffer.
func (rr *RuneReader) fillBuffer() error {
	// Decode the next rune from the buffer if possible
	if rr.pos < len(rr.buf) {
		return nil
	}
	// Read more bytes from the underlying reader
	b := make([]byte, 256)
	n, err := rr.reader.Read(b)
	if n > 0 {
		rr.buf = append(rr.buf, b[:n]...)
	}
	if n == 0 && err != nil {
		return err
	}
	return nil
}

// peekRune decodes one rune starting at rr.buf[rr.pos] without consuming it.
func (rr *RuneReader) peekRune() (rune, int, error) {
	if rr.pos >= len(rr.buf) {
		if err := rr.fillBuffer(); err != nil {
			return -1, 0, err
		}
	}
	if rr.pos >= len(rr.buf) {
		return -1, 0, io.EOF
	}
	ch, size := utf8.DecodeRune(rr.buf[rr.pos:])
	if ch == utf8.RuneError && size <= 1 {
		// Invalid UTF-8, treat as single byte
		return rune(rr.buf[rr.pos]), 1, nil
	}
	return ch, size, nil
}

// Peek returns the next rune without consuming it.
func (rr *RuneReader) Peek() (rune, error) {
	ch, _, err := rr.peekRune()
	return ch, err
}

// Read returns the next rune, consuming it.
func (rr *RuneReader) Read() (rune, error) {
	ch, size, err := rr.peekRune()
	if ch == -1 {
		return -1, err
	}
	rr.pos += size
	// Compact buffer periodically
	if rr.pos > 1024 && rr.pos > len(rr.buf)/2 {
		rr.buf = rr.buf[rr.pos:]
		rr.pos = 0
	}
	return ch, nil
}

// Unread puts a character back to be read again.
func (rr *RuneReader) Unread(ch rune) {
	var b [4]byte
	size := utf8.EncodeRune(b[:], ch)
	// Insert at current position
	rr.buf = append(rr.buf, 0)
	copy(rr.buf[rr.pos+size:], rr.buf[rr.pos:])
	copy(rr.buf[rr.pos:rr.pos+size], b[:size])
}
