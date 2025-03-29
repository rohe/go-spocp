package main

import (
	"fmt"
	"log"
	"net/netip"
	"time"
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
	valueType  string
	boundary   [2]string
	numLimit   [2]int
	ipv4Limit  [2]netip.Addr
	alphaLimit [2]string
	dateLimit  [2]time.Time
	timeLimit  [2]time.Time
	ipv6Limit  [2]netip.Addr
}

type Prefix struct {
	Value []byte
}

type Suffix struct {
	Value []byte
}

type Node struct {
	SExpression bool
	// sExp        *Node
	sPart  []Node
	Octet  *OctetString
	Set    *Set
	Range  *Range
	Prefix *Prefix
	Suffix *Suffix
}

var ValueType = []string{"sexpression", "octet_string", "set", "range", "prefix", "suffix"}

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

func (nod Node) Compare(nod2 Node) (bool, error) {
	if nod.IsType("sexpression") && nod2.IsType("sexpression") {
		return SExpressionCompare(nod, nod2)
	} else if nod.IsType("octet_string") && nod2.IsType("octet_string") {
		return OctetCompare(nod.Octet.Value, nod2.Octet.Value)
	} else {
		return false, fmt.Errorf("invalid comparison operation")
	}
}

func (nod Node) SameType(nod2 Node) string {
	for _, typ := range ValueType {
		if nod.IsType(typ) && nod2.IsType(typ) {
			return typ
		}
	}
	return ""
}

const (
	SetStarform    = "set"
	RangeStarform  = "range"
	PrefixStarform = "prefix"
	SuffixStarform = "suffix"
)

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

func FindBalancing(bs []byte, lead byte, tail byte) int {
	seen := 0

	for index, val := range bs {
		if lead == val {
			if index != 0 {
				seen++
			}
		} else if tail == val {
			if seen == 0 {
				return index
			} else {
				seen--
			}
		}
	}
	return 0
}

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
		sPart: nil,
	}
	inp.currentPosition = octStrStart + octStrLen + 1
	return &node, nil
}

func GetParts(inp *Input, brackets *int) ([]Node, error) {
	var element *Node
	var members []Node
	var arrayLen int
	var localInput Input
	var err error

	for inp.Remaining() > 0 {
		if inp.NextByte() == LeftBracket {
			arrayLen = FindBalancing(inp.RemainingBytes(), '(', ')')
			if arrayLen == 0 {
				return nil, fmt.Errorf("no balancing '%c' found", ')')
			}
			// nb++
			localInput = Input{
				inp.Slice(inp.currentPosition+1, inp.currentPosition+arrayLen),
				0,
			}
			element, err = GetSexp(&localInput, brackets)
			if err != nil {
				return nil, err
			}
			inp.currentPosition += arrayLen + 1
			*brackets++
			members = append(members, *element)
		} else if inp.NextByte() == RightBracket {
			if *brackets > 0 {
				*brackets--
				inp.currentPosition++
			} else {
				return nil, fmt.Errorf("balancing brackets found, where none should be")
			}
		} else { // MUST be an octet-string
			element, err = GetOctet(inp)
			if err != nil {
				return nil, err
			}
			members = append(members, *element)
		}
	}
	return members, nil
}

func GetSexp(inp *Input, brackets *int) (*Node, error) {
	var tag *Node
	var parts []Node
	var err error

	// first element MUST be a tag
	tag, err = GetOctet(inp)

	if err != nil {
		log.Fatal(err)
		return tag, err
	}
	tag.SExpression = true

	if string(tag.Octet.Value) == "*" {
		parts, err = GetStarForm(inp, brackets)
		if err != nil {
			return nil, err
		}
		tag = &parts[0]
	} else {
		parts, err = GetParts(inp, brackets)
		if err != nil {
			return nil, err
		}
		tag.sPart = parts
	}
	return tag, nil
}

func GetStarForm(inp *Input, brackets *int) ([]Node, error) {
	var node *Node
	var result []Node

	var err error
	var setItem *Set
	var rangeItem *Range
	var prefixItem *Prefix
	var suffixItem *Suffix

	node, err = GetOctet(inp)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// First the star form type
	switch c := string(node.Octet.Value); c {
	case SetStarform:
		setItem, err = GetSet(inp, brackets)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		node.Set = setItem
		node.Octet = nil
	case RangeStarform:
		rangeItem, err = GetRange(inp)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		node.Range = rangeItem
		node.Octet = nil
	case PrefixStarform:
		prefixItem, err = GetPrefix(inp)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		node.Prefix = prefixItem
		node.Octet = nil
	case SuffixStarform:
		suffixItem, err = GetSuffix(inp)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		node.Suffix = suffixItem
		node.Octet = nil
	default:
		return nil, fmt.Errorf("invalid star form")
	}
	result = append(result, *node)
	return result, nil
}

func GetSet(inp *Input, brackets *int) (*Set, error) {
	// set = "3:set" 1*[s-expr / tag]
	var item *Node
	var prim Set
	var err error
	// n := 0

	prim = Set{}

	for {
		if inp.NextByte() == LeftBracket {
			inp.currentPosition += 1
			item, err = GetSexp(inp, brackets)
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

	// Verify that there are no two s-expression with the same tag, the same for octet strings
	seenSexp := make(map[string]bool)
	seenOctet := make(map[string]bool)

	for _, nod := range prim.Value {
		if nod.IsType("sexpression") {
			if seenSexp[string(nod.Octet.Value)] {
				err = fmt.Errorf("duplicate s-expression tags")
				return nil, err
			}
			seenSexp[string(nod.Octet.Value)] = true
		} else if nod.IsType("octet_string") {
			if seenOctet[string(nod.Octet.Value)] {
				err = fmt.Errorf("duplicate octet string")
				return nil, err
			}
			seenOctet[string(nod.Octet.Value)] = true
		}
	}

	return &prim, nil
}

func PrintIndent(level int) {
	for ; level > 0; level-- {
		fmt.Printf("%s", TAB)
	}
}

func PrintOctet(node Node, level int) {
	PrintIndent(level)
	fmt.Printf("%s", node.Octet.Value)

	// for _, v := range node.sPart {
	// 	PrintOctet(node.sExp, level+1)
	// }
	// } else {
	// 	fmt.Println()
	// }
}

func PrintPrefix(node Node, level int) {
	var txt string

	PrintIndent(level)
	txt = fmt.Sprintf("Prefix %s", node.Octet.Value)
	fmt.Println(txt)
}

func PrintSuffix(node Node, level int) {
	var txt string

	PrintIndent(level)
	txt = fmt.Sprintf("Suffix %s", node.Octet.Value)
	fmt.Println(txt)
}

func PrintSequence(member []Node, level int) {
	for _, node := range member {
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
	}
}
func PrintSExpression(node Node, level int) {
	PrintOctet(node, level)

	if node.sPart != nil {
		PrintSequence(node.sPart, level+1)
	}
}

func PrintSet(node Node, level int) {
	for _, nod := range node.Set.Value {
		if nod.IsType("sexpression") {
			PrintSExpression(nod, level)
		} else if nod.IsType("octet_string") {
			PrintOctet(nod, level)
		} else if nod.IsType("set") {
			for ; level > 0; level-- {
				fmt.Printf("%s", TAB)
			}
			fmt.Println("Set")
		}
	}
}
