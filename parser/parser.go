package parser

import (
	"bytes"
	"errors"
	"io"
)

type OutputLine struct {
	Signal   string
	FuncName string
	Args     []string
	Result   string
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) Parse() (*OutputLine, error) {
	line := &OutputLine{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return line, ErrEOF
	}

	// Handle signals.
	if tok == SIGNAL {
		// Parse the line unchanged.
		var buf bytes.Buffer
		buf.WriteString(lit)
		for {
			tok, lit = p.scan()
			if tok == NEWLINE {
				break
			}
			buf.WriteString(lit)
		}
		line.Signal = buf.String()
		return line, nil
	} else {
		line.FuncName = lit
	}

	tok, lit = p.scanIgnoreWhitespace()
	if tok == OPEN_PAREN {
		// Read all the args up to CLOSE_PAREN
		for {
			tok, lit = p.scanIgnoreWhitespace()
			if tok == CLOSE_PAREN {
				break
			} else if tok == MEMADDR {
				// Replace with 0x0.
				line.Args = append(line.Args, "0x0")
				// Parse any struct arguments as a single arg.
			} else if tok == OPEN_BRACE {
				var buf bytes.Buffer
				buf.WriteString(lit)
				for {
					tok, lit = p.scan()
					if tok == CLOSE_BRACE {
						buf.WriteString(lit)
						line.Args = append(line.Args, buf.String())
						break
					} else if tok == MEMADDR {
						buf.WriteString("0x0")
					} else {
						buf.WriteString(lit)
					}
				}
			} else if tok == STRING {
				line.Args = append(line.Args, lit)
			} else if tok != SEP {
				line.Args = append(line.Args, lit)
			}
		}
	} else {
		return nil, errors.New("Expected OPEN_PAREN")
	}

	tok, lit = p.scanIgnoreWhitespace()
	if tok != EQUALS {
		return nil, errors.New("Expected EQUALS")
	}

	// Read everything after '=' until newline as the result.
	var result bytes.Buffer
	for {
		tok, lit = p.scan()
		if tok == MEMADDR {
			result.WriteString("0x0")
		} else if tok == NEWLINE {
			break
		} else {
			result.WriteString(lit)
		}
	}
	line.Result = result.String()
	return line, nil
}

func (o *OutputLine) Unparse() string {
	var buf bytes.Buffer
	if o.Signal != "" {
		buf.WriteString(o.Signal)
	} else {
		buf.WriteString(o.FuncName)
		buf.WriteString("(")
		for idx, arg := range o.Args {
			buf.WriteString(arg)
			if idx < len(o.Args)-1 {
				buf.WriteString(", ")
			} else {
				buf.WriteString(")")
			}
		}
		buf.WriteString(" = ")
		buf.WriteString(o.Result)
	}
	return buf.String()
}
