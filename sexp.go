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

	remainder := inp.remaining()
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

//func FindBalancing(bs []byte, lead byte, tail byte) uint16 {
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
//}

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
	var lastByte int
	nb := 0

	// first element MUST be a tag
	tag, err = GetOctet(inp)
	if err != nil {
		log.Fatal(err)
		return tag, err
	}
	node = tag

	for inp.remaining() > 0 {
		if inp.nextByte() == LeftBracket {
			nb++
			// can be either an s-expr or a star-form
			// a star-form starts with 1:*
			if bytes.Equal(inp.prefix(3), StarFormPrefix) {
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
		} else if inp.nextByte() == RightBracket {
			nb--
			if nb < 0 {
				break
			}
			lastByte += 1
		} else { // MUST be an octet-string
			next, err = GetOctet(inp)
			lastByte = next.end
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
	// Skip the '1:*' prefix
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

//func ParseSexp(inp *Input) (*Node, error) {
//	var node *Node
//	var n uint16
//	var star_form = []byte{'0', '*'}
//	var err error
//
//	if inp.next_byte() == '(' { // Sexp
//		n = FindBalancing(bs, '(', ')')
//		if n > 0 {
//			node, err = GetSexp(inp)
//			if err != nil {
//				log.Fatal(err)
//				return node, err
//			}
//		}
//	} else if bytes.Equal(bs[0:1], star_form) {
//		if len(bs) == 2 {
//			node = &Node{
//				typ: Wildcard,
//			}
//		} else {
//			node, err = GetStarForm(inp)
//		}
//	} else {
//		node, err = GetOctet(inp)
//	}
//
//	return node, nil
//}
