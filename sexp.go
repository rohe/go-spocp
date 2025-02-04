package main

import (
	"bytes"
	"fmt"
	"log"
)

type SexpPart int

const (
	Octet SexpPart = iota
	SExpression
	Set
	Range
	Prefix
	Suffix
	Wildcard
)

var LeftBracket byte = 40
var RightBracket byte = 41

type Node struct {
	typ   SexpPart
	begin int
	end   int
	next  *Node
	part  *Node
}

var SetStarform = []byte{'s', 'e', 't'}
var RangeStarform = []byte{'r', 'a', 'n', 'g', 'e'}
var PrefixStarform = []byte{'p', 'r', 'e', 'f', 'i', 'x'}
var SuffixStarform = []byte{'s', 'u', 'f', 'f', 'i', 'x'}

var StarFormPrefix = []byte{'1', ':', '*'}

func Digit(c byte) bool {
	if c >= 48 && c <= 57 {
		return true
	} else {
		return false
	}
}

func GetLen(inp *Input) (int, int, error) {
	n := 0
	b := 0

	remainder := inp.Remaining()
	if remainder == 0 {
		return -1, b, fmt.Errorf("empty string to work with")
	}
	for i, val := range inp.bs[inp.startByte:] {
		if Digit(val) {
			if n != 0 {
				n *= 10
			}
			n += int(val) - 48 // '0' ascii
		} else {
			b = inp.startByte + i
			break
		}
	}
	if n == 0 {
		return -1, b, fmt.Errorf("no digit found")
	}
	inp.startByte = b + n + 1
	return n, b, nil
}

// func FindBalancing(bs []byte, lead byte, tail byte) uint16 {
//	var seen int = 0
//
//	for index, val := range bs {
//		if lead == val {
//			if index != 0 {
//				seen++
//			}
//		} else if tail == val {
//			if seen == 0 {
//				return uint16(index)
//			} else {
//				seen--
//			}
//		}
//	}
//	return 0
// }

func GetOctet(inp *Input) (*Node, error) {
	octStrStart := 0
	var node Node
	var octStrLen int
	var err error

	// Get byte array
	octStrLen, octStrStart, err = GetLen(inp)
	if err != nil {
		log.Fatal(err)
	}
	node = Node{
		typ:   Octet,
		begin: octStrStart + 1,
		end:   octStrStart + octStrLen + 1,
		next:  nil,
		part:  nil,
	}
	inp.startByte = octStrStart + octStrLen + 1
	return &node, nil
}

func GetSexp(inp *Input) (*Node, error) {
	var tag, node, next *Node
	var sexp Node
	var err error
	nb := 0

	// first element MUST be a tag
	tag, err = GetOctet(inp)
	if err != nil {
		log.Fatal(err)
		return tag, err
	}
	node = tag

	for inp.Remaining() > 0 {
		if inp.NextByte() == LeftBracket {
			nb++
			inp.startByte += 1
			// can be either an s-expr or a star-form
			// a star-form starts with 1:*
			if bytes.Equal(inp.Prefix(3), StarFormPrefix) {
				next, err = GetStarForm(inp)
			} else {
				next, err = GetSexp(inp)
			}
			if err != nil {
				log.Fatal(err)
				return next, err
			} else {
				node.next = next
				node = next
			}
		} else if inp.NextByte() == RightBracket {
			nb--
			inp.startByte += 1
			if nb < 0 {
				break
			}
		} else { // MUST be an octet-string
			next, err = GetOctet(inp)
			node.next = next
			node = next
		}
	}
	sexp = Node{
		typ: SExpression,
	}
	sexp.part = tag
	return &sexp, nil
}

func GetStarForm(inp *Input) (*Node, error) {
	var node *Node
	var err error
	var c []byte

	// first element MUST be '1:*'
	// Skip the '1:*' Prefix
	inp.startByte += 3
	node, err = GetOctet(inp)
	if err != nil {
		log.Fatal(err)
		return node, err
	}
	// First the star form type
	c = inp.bs[node.begin:node.end]
	if bytes.Equal(SetStarform, c) {
		node.typ = Set
	} else if bytes.Equal(RangeStarform, c) {
		node.typ = Range
	} else if bytes.Equal(PrefixStarform, c) {
		node.typ = Prefix
	} else if bytes.Equal(SuffixStarform, c) {
		node.typ = Suffix
	} else {
		return node, fmt.Errorf("invalid star form")
	}

	return node, nil
}

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
