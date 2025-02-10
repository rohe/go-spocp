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
	// TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>
	// s := "(gopher foo)"
	// s := "(11:certificate(6:issuer3:bob)(7:subject5:alice))"
	s := "(11:certificate(6:issuer3:bob)(5:level(1:*5:range7:numeric2:ge3:100)))"
	bs := []byte(s)
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
