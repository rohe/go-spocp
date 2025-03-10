package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Compare functions return -1 if rule is less than query
// 0 if equal and 1 if rule is greater then query

func SExpressionCompare(query, rule *Node) (bool, error) {
	var err error
	var cmp bool

	// compare tag
	cmp, err = OctetCompare(rule.Octet.Value, query.Octet.Value)
	if err != nil {
		return false, err
	}
	if cmp == false {
		return false, nil
	}

	if query.part != nil && rule.part != nil {
		cmp, err = CompareSequence(query.part, rule.part)
		if err != nil {
			return false, err
		}
		if cmp == false {
			return false, nil
		}
	} else if rule.part != nil {
		return false, fmt.Errorf("rule more specific than query")
	}

	if query.next != nil && rule.next != nil {
		cmp, err = CompareSequence(query.next, rule.next)
		if err != nil {
			return false, err
		}
		if cmp == false {
			return false, nil
		}
	} else if rule.next != nil {
		return false, fmt.Errorf("rule more specific than query")
	}
	return true, nil
}

func OctetCompare(query, rule []byte) (bool, error) {
	if bytes.Equal(query, rule) {
		return true, nil
	} else {
		return false, nil
	}
}

func OctetToSetCompare(query []byte, rule []Node) (bool, error) {
	var err error
	var cmp bool
	var matched int

	// at least one must match == be less or equal to
	for _, nod := range rule {
		if nod.IsType("octet_string") {
			cmp, err = OctetCompare(query, nod.Octet.Value)
		}
		if err != nil {
			return false, err
		}
		if cmp == true {
			matched++
			break
		}
	}
	if matched > 0 {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid node comparison")
	}
}

func SetToSetCompare(query []Node, rule []Node) (bool, error) {
	var cmp bool
	var err error

	for _, nod := range query {
		if nod.IsType("octet_string") {
			cmp, err = OctetToSetCompare(nod.Octet.Value, rule)
		}
		if err != nil {
			return false, err
		}
		if cmp == false {
			return false, fmt.Errorf("item didn't match any in a set")
		}
	}
	return true, nil
}

func RangeCompare(query, rule *Range) (bool, error) {
	// place holder
	return false, fmt.Errorf("invalid node comparison")
}

func NumericRangeCompare(query *OctetString, rule *Range, num int) (bool, error) {
	Val, err := strconv.Atoi(string(query.Value))
	if err != nil {
		return false, err
	}
	if "le" == rule.boundary[num] {
		if uint8(Val) <= uint8(rule.numLimit[0]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "lt" == rule.boundary[num] {
		if uint8(Val) < uint8(rule.numLimit[0]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "ge" == rule.boundary[num] {
		if uint8(Val) >= uint8(rule.numLimit[0]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "gt" == rule.boundary[num] {
		if uint8(Val) > uint8(rule.numLimit[0]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, fmt.Errorf("invalid range comparison")
}

func DateRangeCompare(query *OctetString, rule *Range, num int) (bool, error) {
	tid, err := time.Parse(time.RFC3339, string(query.Value))
	if err != nil {
		return false, err
	}
	// Could use Time Before(), Equal() and After() but I'm just learning what tools there are
	queryUnixTime := tid.Unix()
	ruleUnixTime := rule.dateLimit[num].Unix()
	if err != nil {
		return false, err
	}
	if "le" == rule.boundary[num] {
		if queryUnixTime <= ruleUnixTime {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "lt" == rule.boundary[num] {
		if queryUnixTime < ruleUnixTime {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "ge" == rule.boundary[num] {
		if queryUnixTime >= ruleUnixTime {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "gt" == rule.boundary[num] {
		if queryUnixTime > ruleUnixTime {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, fmt.Errorf("invalid range comparison")
}

func TimeRangeCompare(query *OctetString, rule *Range, num int) (bool, error) {
	queryTime, err := time.Parse("15:04:05", string(query.Value))
	if err != nil {
		return false, err
	}
	if "le" == rule.boundary[num] {
		if queryTime.Before(rule.timeLimit[num]) || queryTime.Equal(rule.timeLimit[num]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "lt" == rule.boundary[num] {
		if queryTime.Before(rule.timeLimit[num]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "ge" == rule.boundary[num] {
		if queryTime.After(rule.timeLimit[num]) || queryTime.Equal(rule.timeLimit[num]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	if "gt" == rule.boundary[num] {
		if queryTime.After(rule.timeLimit[num]) {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, fmt.Errorf("invalid range comparison")
}

func OctetToRangeCompare(query *OctetString, rule *Range) (bool, error) {
	var cmp bool
	var err error

	if rule.valueType == NUMERIC {
		cmp, err = NumericRangeCompare(query, rule, 0)
		if err != nil {
			return false, err
		}
		if cmp == true {
			if rule.boundary[1] != "" {
				return NumericRangeCompare(query, rule, 1)
			} else {
				return true, nil
			}
		}
		return cmp, nil
	} else if rule.valueType == DATE {
		cmp, err = DateRangeCompare(query, rule, 0)
		if err != nil {
			return false, err
		}
		if cmp == true {
			if rule.boundary[1] != "" {
				return DateRangeCompare(query, rule, 1)
			} else {
				return true, nil
			}
		}
		return cmp, nil
	} else if rule.valueType == TIME {
		cmp, err = TimeRangeCompare(query, rule, 0)
		if err != nil {
			return false, err
		}
		if cmp == true {
			if rule.boundary[1] != "" {
				return TimeRangeCompare(query, rule, 1)
			} else {
				return true, nil
			}
		}
		return cmp, nil
	}
	return false, fmt.Errorf("invalid range comparison")
}

func PrefixCompare(query, rule []byte) (bool, error) {
	return OctetCompare(query, rule)
}

func SuffixCompare(query, rule []byte) (bool, error) {
	return OctetCompare(query, rule)
}

func LessOrEqualTo(query, rule *Node) (bool, error) {
	switch {
	case rule.IsType("sexpression") && query.IsType("sexpression"):
		return SExpressionCompare(query, rule)
	case rule.IsType("octet_string") && query.IsType("octet_string"):
		return OctetCompare(query.Octet.Value, rule.Octet.Value)
	case rule.IsType("set") && query.IsType("set"):
		return SetToSetCompare(query.Set.Value, rule.Set.Value)
	case rule.IsType("set") && query.IsType("octet_string"):
		return OctetToSetCompare(query.Octet.Value, rule.Set.Value)
	case rule.IsType("range") && query.IsType("range"):
		return RangeCompare(query.Range, rule.Range)
	case rule.IsType("range") && query.IsType("octet_string"):
		return OctetToRangeCompare(query.Octet, rule.Range)
	case rule.IsType("prefix") && query.IsType("prefix"):
		return PrefixCompare(query.Prefix.Value, rule.Prefix.Value)
	case rule.IsType("suffix") && query.IsType("suffix"):
		return SuffixCompare(query.Suffix.Value, rule.Suffix.Value)
	default:
		return false, fmt.Errorf("unknown value type or not matching value types")
	}
}

func CompareSequence(query, rule *Node) (bool, error) {
	var cmp bool
	var err error

	for rule != nil {
		cmp, err = LessOrEqualTo(query, rule)

		if err != nil {
			return false, err
		}
		if cmp == false {
			return cmp, nil
		}

		rule = rule.next
		if rule != nil {
			query = query.next
			if query == nil {
				return false, fmt.Errorf("query list shorter than rule")
			}
		}
	}
	return true, nil
}
