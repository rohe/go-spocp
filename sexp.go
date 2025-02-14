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
	typ      SexpPart
	begin    int
	end      int
	next     *Node
	part     *Node
	StarForm *StarForm
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
			fmt.Println(inp.Left())
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
	var node, item *Node
	var err error
	var c []byte
	var spec *StarForm

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
		item, err = GetSet(inp)
		if err != nil {
			log.Fatal(err)
			return node, err
		} else {
			node.next = item
		}
	} else if bytes.Equal(RangeStarform, c) {
		node.typ = Range
		spec, err = GetRange(inp)
		node.StarForm = spec
		if err != nil {
			log.Fatal(err)
			return node, err
		}
	} else if bytes.Equal(PrefixStarform, c) {
		node.typ = Prefix
	} else if bytes.Equal(SuffixStarform, c) {
		node.typ = Suffix
	} else {
		return node, fmt.Errorf("invalid star form")
	}

	return node, nil
}

func GetSet(inp *Input) (*Node, error) {
	var node, item *Node
	var prim Node
	var err error

	prim = Node{
		typ: Set,
	}

	node = &prim
	for {
		if inp.NextByte() == LeftBracket {
			item, err = GetSexp(inp)
			if err != nil {
				log.Fatal(err)
				return node, err
			}
		} else if inp.NextByte() == RightBracket {
			break
		} else {
			item, err = GetOctet(inp)
		}
		node.part = item
		node = item
	}
	return &prim, nil
}

func PrintOctet(inp Input, node *Node, indent int) {
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	fmt.Printf("%s", inp.Slice(node.begin, node.end))

	if node.next != nil {
		PrintOctet(inp, node.next, indent+1)
	} else {
		fmt.Println()
	}
}

func PrintSExpression(inp Input, root *Node, indent int) {

	var node *Node

	node = root.part
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	fmt.Println(string(inp.Slice(node.begin, node.end)))

	if node.next.typ == SExpression {
		PrintSExpression(inp, node.next, indent+1)
	} else if node.next.typ == Octet {
		PrintOctet(inp, node.next, indent+2)
	} else if node.next.typ == Set {
		PrintSet(inp, node.next, indent+2)
	} else if node.next.StarForm != nil {
		PrintStartForm(inp, node.next, indent+2)
	}
	if root.next != nil {
		if root.next.typ == SExpression {
			PrintSExpression(inp, root.next, indent+1)
		} else if root.next.typ == Octet {
			PrintOctet(inp, root.next, indent+2)
		}
	}
}

func PrintSet(inp Input, node *Node, indent int) {
	if node.next != nil {
		node = node.next
	}
	for ; node != nil; node = node.part {
		if node.typ == SExpression {
			PrintSExpression(inp, node, indent+3)
		} else if node.typ == Octet {
			PrintOctet(inp, node, indent+3)
		} else if node.typ == Set {
			for ; indent > 0; indent-- {
				fmt.Printf("%s", TAB)
			}
			fmt.Println("Set")
		}
	}
}
