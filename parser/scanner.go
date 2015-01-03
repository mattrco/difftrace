package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	NEWLINE
	WS          // Whitespace
	IDENT       // Identifier, e.g. function name.
	MEMADDR     // Hex address, e.g. 0xcafef00d
	OPEN_PAREN  // (
	CLOSE_PAREN // )
	OPEN_BRACE  // {
	CLOSE_BRACE // }
	STRING      // Delimited by "
	SEP         // ,
	EQUALS      // =
	SIGNAL      // ---
)

var eof = rune(0)
var ErrEOF = errors.New("EOF")

// As it is useful to process newline differently to other
// whitespace don't include it here.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// Scanner is a lexical scanner implemented with a buffered reader.
type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the buffered reader.
// Returns rune(0) if an error occurs.
func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

// unreadRune places the previously read rune back on the reader.
func (s *Scanner) unreadRune() error { return s.r.UnreadRune() }

// unreadRunes allows a number of runes to be unread.
func (s *Scanner) unreadRunes(runes int) error {
	for runes > 0 {
		if err := s.r.UnreadRune(); err != nil {
			return err
		}
		runes--
	}
	return nil
}

// Scan returns the next token and literal string it represents.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	r := s.read()

	if r == '0' {
		// 0 could be the start of a memory address. Read the next
		// rune and if it is 'x', pass to scanAddress.
		if r := s.read(); r == 'x' {
			return s.scanAddress()
		} else {
			s.unreadRunes(2)
		}
	} else if r == '-' {
		// - could be a signal (lines start with ---).
		if r := s.read(); r == '-' {
			if r := s.read(); r == '-' {
				return SIGNAL, "---"
			} else {
				s.unreadRunes(2)
			}
		} else {
			s.unreadRune()
		}
	} else if r == '"' {
		return s.scanString()
	} else if isWhitespace(r) {
		// Consume all contiguous whitespace.
		s.unreadRune()
		return s.scanWhitespace()
	} else if isLetter(r) || isDigit(r) {
		// If we see a letter then consume as an identifier.
		s.unreadRune()
		return s.scanIdent()
	}

	// Otherwise read the individual character.
	switch r {
	case eof:
		return EOF, ""
	case '(':
		return OPEN_PAREN, string(r)
	case ',':
		return SEP, string(r)
	case ')':
		return CLOSE_PAREN, string(r)
	case '{':
		return OPEN_BRACE, string(r)
	case '}':
		return CLOSE_BRACE, string(r)
	case '=':
		return EQUALS, string(r)
	case '\n':
		return NEWLINE, string(r)
	}

	return ILLEGAL, string(r)
}

// scanAddress consumes a memory address from the scanner.
func (s *Scanner) scanAddress() (tok Token, lit string) {
	var buf bytes.Buffer
	// Prepend 0x as memory addresses are always hex.
	buf.WriteString("0x")

	// Read up to the next whitespace rune.
	for {
		r := s.read()
		if r == eof {
			break
		} else if isWhitespace(r) || r == '\n' {
			s.unreadRune()
			break
		}
		buf.WriteRune(r)
	}

	return MEMADDR, buf.String()
}

// scanString consumes a string from the scanner.
func (s *Scanner) scanString() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune('"')

	// Read up to the next inverted comma.
	for {
		r := s.read()
		if r == eof {
			break
		}
		if r == '"' {
			buf.WriteRune(r)
			// Reached end of literal. Consume ellipsis if present.
			if r = s.read(); r == '.' {
				s.read()
				s.read()
				buf.WriteString("...")
			} else {
				s.unreadRune()
			}
			break
		}
		buf.WriteRune(r)
	}

	return STRING, buf.String()
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	for {
		if r := s.read(); r == eof {
			break
		} else if !isWhitespace(r) {
			s.unreadRune()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !isLetter(r) && !isDigit(r) && r != '_' {
			s.unreadRune()
			break
		} else {
			_, _ = buf.WriteRune(r)
		}
	}

	// Otherwise return as a regular identifier.
	return IDENT, buf.String()
}
