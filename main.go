package main

import (
	"fmt"
	"log"
)

var TAB = []byte{32, 32, 32, 32}

func PrintOctet(bs []byte, node *Node, indent int) {
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	fmt.Printf("%s", bs[node.begin:node.end])
	if node.next != nil {
		PrintOctet(bs, node.next, indent+1)
	} else {
		fmt.Println()
	}
}

func PrintSExpression(bs []byte, root *Node, indent int) {

	var node *Node

	node = root.part
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	fmt.Println(string(bs[node.begin:node.end]))

	if node.next.typ == SExpression {
		PrintSExpression(bs, node.next, indent+1)
	} else if node.next.typ == Octet {
		PrintOctet(bs, node.next, indent+2)
	}
	if root.next != nil {
		if root.next.typ == SExpression {
			PrintSExpression(bs, root.next, indent+1)
		} else if root.next.typ == Octet {
			PrintOctet(bs, root.next, indent+2)
		}
	}
}

func main() {
	// TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>
	// s := "(gopher foo)"
	s := "(11:certificate(6:issuer3:bob)(7:subject5:alice))"
	bs := []byte(s)
	var s_expression *Node
	var err error

	s_expression, err = GetSexp(bs, 1)
	if err != nil {
		log.Fatal("Parse error")
	}
	fmt.Println("Done ", s_expression.typ)
	PrintSExpression(bs, s_expression, 0)
	// var n uint16 = FindBalancing(bs, '(', ')')
	// if n == 0 {
	// 	fmt.Println("Balancing failed!")
	// } else {
	// 	fmt.Printf("[%d] <%s> ", FindBalancing(bs, '(', ')'), s[1:n])
	// 	fmt.Println("")
	// }
}
