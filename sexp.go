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

// var SemiColon byte = 58
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

func GetLen(bs []byte, begin int) (int, int, error) {
	n := 0
	b := 0

	if len(bs) == 0 {
		return -1, b, fmt.Errorf("Empty string")
	}
	for i, val := range bs[begin:] {
		if Digit(val) {
			if n != 0 {
				n *= 10
			}
			n += int(val) - 48 // '0' ascii
		} else {
			b = begin + i
			break
		}
	}
	if n == 0 {
		return -1, b, fmt.Errorf("No digit found")
	}
	return n, b, nil
}

func FindBalancing(bs []byte, lead byte, tail byte) uint16 {
	var seen int = 0

	for index, val := range bs {
		if lead == val {
			if index != 0 {
				seen++
			}
		} else if tail == val {
			if seen == 0 {
				return uint16(index)
			} else {
				seen--
			}
		}
	}
	return 0
}

func LastByte(node *Node) int {
	for node.next != nil {
		node = node.next
	}
	return node.end
}

// func GetList(bs []byte) []byte {
// 	return bs
// }

func GetOctet(bs []byte, begin int) (*Node, error) {
	octstr_start := 0
	var _node Node
	var octstr_len int
	var err error

	// Get byte array
	octstr_len, octstr_start, err = GetLen(bs, begin)
	if err != nil {
		log.Fatal(err)
	}
	_node = Node{
		typ:   Octet,
		begin: octstr_start + 1,
		end:   octstr_start + octstr_len + 1,
		next:  nil,
		part:  nil,
	}
	return &_node, nil
}

func GetSexp(bs []byte, begin int) (*Node, error) {
	var tag, node, next *Node
	var sexp Node
	var err error
	blen := len(bs)
	var last_byte int
	nb := 0

	// first element MUST be a tag
	tag, err = GetOctet(bs, begin)
	if err != nil {
		log.Fatal(err)
		return tag, err
	}
	node = tag
	last_byte = LastByte(node)
	for last_byte < blen {
		if bs[last_byte] == LeftBracket {
			nb++
			// can be either an s-expr or a star-form
			// a star-form starts with 1:*
			if bytes.Equal(bs[last_byte:last_byte+3], StarFormPrefix) {
				next, err = GetStarForm(bs, last_byte+1)
			} else {
				next, err = GetSexp(bs, last_byte+1)
			}
			if err != nil {
				log.Fatal(err)
				return next, err
			} else {
				last_byte = LastByte(next.part)
				node.next = next
				node = next
			}
		} else if bs[last_byte] == RightBracket {
			nb--
			if nb < 0 {
				break
			}
			last_byte += 1
		} else { // MUST be an octet-string
			next, err = GetOctet(bs, last_byte)
			last_byte = next.end
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

func GetStarForm(bs []byte, begin int) (*Node, error) {
	var node *Node
	var err error
	var c []byte

	// first element MUST star form
	// Skip the '1:*' prefix
	node, err = GetOctet(bs, begin+3)
	if err != nil {
		log.Fatal(err)
		return node, err
	}
	// First the star form type
	c = bs[node.begin:node.end]
	if bytes.Equal(SetStarform, c) {
		node.typ = Set
	} else if bytes.Equal(RangeStarform, c) {
		node.typ = Range
	} else if bytes.Equal(PrefixStarform, c) {
		node.typ = Prefix
	} else if bytes.Equal(SuffixStarform, c) {
		node.typ = Suffix
	} else {
		return node, fmt.Errorf("Invalid star form")
	}

	return node, nil
}

func ParseSexp(bs []byte, begin int) (*Node, error) {
	var node *Node
	var n uint16
	var star_form = []byte{'0', '*'}
	var err error

	if bs[0] == '(' { // Sexp
		n = FindBalancing(bs, '(', ')')
		if n > 0 {
			node, err = GetSexp(bs[:n], begin)
			if err != nil {
				log.Fatal(err)
				return node, err
			}
		}
	} else if bytes.Equal(bs[0:1], star_form) {
		if len(bs) == 2 {
			node = &Node{
				typ: Wildcard,
			}
		} else {
			node, err = GetStarForm(bs[2:], begin)
		}
	} else {
		node, err = GetOctet(bs, begin)
	}

	return node, nil
}
