package parser

import (
	"strings"
	"testing"
)

func TestWholeLine(t *testing.T) {

	s := NewScanner(strings.NewReader("lstat(45) = 3"))
	tok, lit := s.Scan()
	if tok != IDENT {
		t.Fatalf("Expected %d, got %d for %s", IDENT, tok, lit)
	}
	tok, lit = s.Scan()
	if tok != OPEN_PAREN {
		t.Fatalf("Expected %d, got %d for %s", OPEN_PAREN, tok, lit)
	}
	tok, lit = s.Scan()
	if tok != IDENT {
		t.Fatalf("Expected %d, got %d for %s", IDENT, tok, lit)
	}
	tok, lit = s.Scan()
	if tok != CLOSE_PAREN {
		t.Fatalf("Expected %d, got %d for %s", CLOSE_PAREN, tok, lit)
	}
}
