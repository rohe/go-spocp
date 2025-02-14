package main

import (
	"fmt"
	"log"
)

var TAB = []byte{32, 32, 32, 32}

type Input struct {
	bs        []byte
	startByte int
}

func (inp Input) Remaining() int {
	return len(inp.bs) - inp.startByte
}
func (inp Input) NextByte() byte {
	return inp.bs[inp.startByte]
}
func (inp Input) Slice(begin int, end int) []byte {
	return inp.bs[begin:end]
}
func (inp Input) Prefix(length int) []byte {
	return inp.bs[inp.startByte : inp.startByte+length]
}
func (inp Input) Left() string {
	return string(inp.bs[inp.startByte:])
}

func main() {
	// s := "(gopher foo)"
	// s := "(11:certificate(6:issuer3:bob)(7:subject5:alice))"
	var s_expressions = []string{
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range7:numeric2:ge3:100)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range5:alpha2:ge3:abc)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv42:ge11:130.239.1.1)))",
		"(11:certificate(6:issuer3:bob)(5:fruit(1:*3:set5:apple6:orange5:lemon)))",
	}
	for _, expression := range s_expressions {
		bs := []byte(expression)
		var SExpression *Node
		var err error
		// Skip the first '('
		var inp = Input{bs, 1}

		SExpression, err = GetSexp(&inp)
		if err != nil {
			log.Fatal("Parse error")
		}
		fmt.Println("Done")
		PrintSExpression(inp, SExpression, 0)
	}
}
