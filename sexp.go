package main

import (
	"bytes"
	"fmt"
	"log"
)

var LeftBracket byte = 40
var RightBracket byte = 41

type OctetString struct {
	Value []byte
}

type Set struct {
	Value []Node
}

type Range struct {
	valueType  byte
	boundary   [2][]byte
	numLimit   [2]int
	ipv4Limit  [2][4]int
	alphaLimit [2]string
	dateLimit  [2]string
	timeLimit  [2]string
	ipv6Limit  [2][]byte
}

type Prefix struct {
	Value []byte
}

type Suffix struct {
	Value []byte
}

type Node struct {
	SExpression bool
	next        *Node
	part        *Node
	Octet       *OctetString
	Set         *Set
	Range       *Range
	Prefix      *Prefix
	Suffix      *Suffix
}

func (nod Node) IsType(typ string) bool {
	if typ == "sexpression" && nod.SExpression == true {
		return true
	} else if typ == "octet_string" && nod.Octet != nil {
		return true
	} else if typ == "set" && nod.Set != nil {
		return true
	} else if typ == "range" && nod.Range != nil {
		return true
	} else if typ == "prefix" && nod.Prefix != nil {
		return true
	} else if typ == "suffix" && nod.Suffix != nil {
		return true
	}
	return false
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
	for i, val := range inp.bs[inp.currentPosition:] {
		if Digit(val) {
			if n != 0 {
				n *= 10
			}
			n += int(val) - 48 // '0' ascii
		} else {
			b = inp.currentPosition + i
			break
		}
	}
	if n == 0 {
		return -1, b, fmt.Errorf("no digit found")
	}
	inp.currentPosition = b + n + 1
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
	oct := OctetString{
		Value: inp.Slice(octStrStart+1, octStrStart+octStrLen+1),
	}

	node = Node{
		Octet: &oct,
		next:  nil,
		part:  nil,
	}
	inp.currentPosition = octStrStart + octStrLen + 1
	return &node, nil
}

func GetSexp(inp *Input) (*Node, error) {
	var tag, node, next *Node
	var err error
	nb := 0

	// first element MUST be a tag
	tag, err = GetOctet(inp)

	if err != nil {
		log.Fatal(err)
		return tag, err
	}
	tag.SExpression = true
	node = tag

	for inp.Remaining() > 0 {
		if inp.NextByte() == LeftBracket {
			nb++
			inp.currentPosition += 1
			fmt.Println(inp.RemainingString())
			// can be either an s-expr or a star-form
			// a star-form starts with 1:*
			if bytes.Equal(inp.Prefix(3), StarFormPrefix) {
				next, err = GetStarForm(inp)
				if err != nil {
					log.Fatal(err)
					return next, err
				}
				node.part = next
				node = next

			} else {
				next, err = GetSexp(inp)
				if err != nil {
					log.Fatal(err)
					return next, err
				}
				node.next = next
				node = next
			}
		} else if inp.NextByte() == RightBracket {
			nb--
			inp.currentPosition += 1
			if nb <= 0 {
				break
			}
		} else { // MUST be an octet-string
			next, err = GetOctet(inp)
			node.part = next
			node = next
		}
	}
	return tag, nil
}

func GetStarForm(inp *Input) (*Node, error) {
	var node *Node
	var err error
	var c []byte
	var setItem *Set
	var rangeItem *Range
	var prefixItem *Prefix
	var suffixItem *Suffix

	// first element MUST be '1:*'
	// Skip the '1:*' Prefix
	inp.currentPosition += 3
	node, err = GetOctet(inp)
	if err != nil {
		log.Fatal(err)
		return node, err
	}
	// First the star form type
	c = node.Octet.Value
	if bytes.Equal(SetStarform, c) {
		setItem, err = GetSet(inp)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		node.Set = setItem
		node.Octet = nil
	} else if bytes.Equal(RangeStarform, c) {
		rangeItem, err = GetRange(inp)
		if err != nil {
			log.Fatal(err)
			return node, err
		}
		node.Range = rangeItem
		node.Octet = nil
	} else if bytes.Equal(PrefixStarform, c) {
		prefixItem, err = GetPrefix(inp)
		if err != nil {
			log.Fatal(err)
			return node, err
		}
		node.Prefix = prefixItem
		node.Octet = nil
	} else if bytes.Equal(SuffixStarform, c) {
		suffixItem, err = GetSuffix(inp)
		if err != nil {
			log.Fatal(err)
			return node, err
		}
		node.Suffix = suffixItem
		node.Octet = nil
	} else {
		return node, fmt.Errorf("invalid star form")
	}

	return node, nil
}

func GetSet(inp *Input) (*Set, error) {
	// set = "3:set" 1*[s-expr / tag]
	var item *Node
	var prim Set
	var err error
	// n := 0

	prim = Set{}

	for {
		if inp.NextByte() == LeftBracket {
			inp.currentPosition += 1
			item, err = GetSexp(inp)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		} else if inp.NextByte() == RightBracket {
			break
		} else {
			item, err = GetOctet(inp)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}
		prim.Value = append(prim.Value, *item)
	}
	return &prim, nil
}

func PrintIndent(level int) {
	for ; level > 0; level-- {
		fmt.Printf("%s", TAB)
	}
}

func PrintOctet(node *Node, level int) {
	PrintIndent(level)
	fmt.Printf("%s", node.Octet.Value)

	if node.next != nil {
		PrintOctet(node.next, level+1)
	} else {
		fmt.Println()
	}
}

func PrintPrefix(node *Node, level int) {
	var txt string
	txt = fmt.Sprintf("Prefix %s", node.Octet.Value)
	fmt.Println(txt)
}

func PrintSuffix(node *Node, level int) {
	var txt string

	PrintIndent(level)
	txt = fmt.Sprintf("Suffix %s", node.Octet.Value)
	fmt.Println(txt)
}

func PrintSequence(node *Node, level int) {
	for node != nil {
		if node.IsType("sexpression") {
			PrintSExpression(node, level)
		} else if node.IsType("octet_string") {
			PrintOctet(node, level)
		} else if node.IsType("set") {
			PrintSet(node, level)
		} else if node.IsType("range") {
			PrintRange(node.Range, level)
		} else if node.IsType("prefix") {
			PrintPrefix(node, level)
		} else if node.IsType("suffix") {
			PrintSuffix(node, level)
		}
		node = node.next
	}
}
func PrintSExpression(node *Node, level int) {
	PrintIndent(level)
	fmt.Println(string(node.Octet.Value))

	if node.part != nil {
		PrintSequence(node.part, level+1)
	}
	if node.next != nil {
		PrintSequence(node.next, level+1)
	}
}

func PrintSet(node *Node, level int) {
	if node.next != nil {
		node = node.next
	}

	for _, nod := range node.Set.Value {
		if nod.IsType("sexpression") {
			PrintSExpression(&nod, level)
		} else if nod.IsType("octet_string") {
			PrintOctet(&nod, level)
		} else if nod.IsType("set") {
			for ; level > 0; level-- {
				fmt.Printf("%s", TAB)
			}
			fmt.Println("Set")
		}
	}
}
