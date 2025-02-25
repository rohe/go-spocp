package main

import (
	"fmt"
	"log"
)

var TAB = []byte{32, 32, 32, 32}

type Input struct {
	bs              []byte
	currentPosition int
}

func (inp Input) Remaining() int {
	return len(inp.bs) - inp.currentPosition
}
func (inp Input) NextByte() byte {
	return inp.bs[inp.currentPosition]
}
func (inp Input) Slice(begin int, end int) []byte {
	return inp.bs[begin:end]
}
func (inp Input) Prefix(length int) []byte {
	return inp.bs[inp.currentPosition : inp.currentPosition+length]
}
func (inp Input) RemainingString() string {
	return string(inp.bs[inp.currentPosition:])
}

func main() {
	// s := "(gopher foo)"
	// s := "(11:certificate(6:issuer3:bob)(7:subject5:alice))"
	var s_expressions = []string{
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range7:numeric2:ge3:100)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range5:alpha2:ge3:abc)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv42:ge11:130.239.1.12:lt13:130.239.1.127)))",
		// "(11:certificate(6:issuer3:bob)(5:fruit(1:*3:set5:apple6:orange5:lemon)))",
		"(1:t(1:*3:set(1:a1:b)(1:c(1:d1:e))(1:f)1:g))",
		// "(t (* set (a (x y)) (b c) (a d)))", // invalid
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
		PrintSExpression(SExpression, 0)
	}
}
