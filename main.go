package main

import (
	"fmt"
	"os"

	"github.com/mattrco/difftrace/parser"
)

func main() {
	p := parser.NewParser(os.Stdin)
	for {
		line, err := p.Parse()
		if err != nil {
			if err != parser.ErrEOF {
				fmt.Println(err.Error())
			}
			break
		} else {
			fmt.Println(line.Unparse())
		}
	}
}
