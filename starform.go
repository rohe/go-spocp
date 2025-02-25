package main

import (
	"bytes"
	"errors"
	"fmt"
)

// type StarForm struct {
// 	valueType byte
// 	boundary  [2][]byte
// 	limit     [2]any
// }

var Alpha = []byte{'a', 'l', 'p', 'h', 'a'}
var Numeric = []byte{'n', 'u', 'm', 'e', 'r', 'i', 'c'}
var Date = []byte{'d', 'a', 't', 'e'}
var Time = []byte{'t', 'i', 'm', 'e'}
var Ipv4 = []byte{'i', 'p', 'v', '4'}
var Ipv6 = []byte{'i', 'p', 'v', '6'}

var RangeTypes = map[byte]string{
	'n': "Numeric",
	'a': "Alpha",
	'd': "Date",
	't': "Time",
	'4': "Ipv4",
	'6': "Ipv6",
}

var LE = []byte("le")
var LT = []byte("lt")
var GE = []byte("ge")
var GT = []byte("gt")

var limits = [][]byte{LE, LT, GE, GT}

func CorrectLimit(val []byte) bool {
	// tests that the given limit type (ge, gt, ...) is one that is expected
	for _, lim := range limits {
		if bytes.Equal(lim, val) {
			return true
		}
	}
	return false
}

func GetLimit(inp *Input) ([]byte, []byte) {
	var gogeLole, value *Node
	var err error
	var limValue []byte

	gogeLole, err = GetOctet(inp)
	if err != nil {
		return nil, []byte("error")
	}
	limValue = gogeLole.Octet.Value
	if CorrectLimit(limValue) == false {

		return nil, []byte("Incorrect boundary type " + string(limValue))
	}

	value, err = GetOctet(inp)
	if err != nil {
		return nil, []byte("error")
	}

	return limValue, value.Octet.Value
}

func VerifyAlpha(rng *Range, value []byte, n int) error {
	// If it can be converted to a string everything is OK
	rng.alphaLimit[n] = string(value)
	return nil
}

func StringToInt(inValue []byte) (int, error) {
	var outValue int
	for _, b := range inValue {
		outValue = outValue*10 + int(b-48)
	}
	return outValue, nil
}

func IPv4Num(value []byte) ([4]int, error) {
	var ip [4]int
	var tmp int
	n := 0
	var none [4]int

	for _, b := range value {
		if b <= '9' && b >= '0' {
			tmp = tmp*10 + int(b-48)
		} else if b == '.' {
			ip[n] = tmp
			tmp = 0
			n++
		} else {
			return none, errors.New("Format error in " + string(value))
		}
	}
	if n == 3 {
		ip[n] = tmp
	} else {
		return none, errors.New("Format error in " + string(value))
	}
	return ip, nil
}

func VerifyIPv4(rng *Range, value []byte, n int) error {
	var ip [4]int
	var err error

	ip, err = IPv4Num(value)
	if err != nil {
		return err
	}
	// make sure the numbers are between 0 and 255
	for _, b := range ip {
		if b >= 0 && b <= 255 {
			continue
		} else {
			return errors.New("Format error in " + string(value))
		}
	}
	rng.ipv4Limit[n] = ip
	return nil
}

func VerifyNumeric(rng *Range, value []byte, n int) error {
	var err error
	var result int

	result, err = StringToInt(value)
	if err != nil {
		return err
	}
	rng.numLimit[n] = result
	return nil
}

// func VerifyLimit(valueType byte, value []byte) (any, error) {
//
// 	if valueType == 'a' {
// 		return VerifyAlpha(value)
// 	} else if valueType == 'n' {
// 		return VerifyNumeric(value)
// 	} else if valueType == '4' {
// 		return VerifyIPv4(value)
// 	}
// 	return nil, errors.New("Unknown value type " + string(valueType))
// }

func GetRestrictions(inp *Input, rng *Range, n int) error {
	var limit []byte
	var value []byte

	limit, value = GetLimit(inp)
	rng.boundary[n] = limit

	if rng.valueType == 'a' {
		return VerifyAlpha(rng, value, n)
	} else if rng.valueType == 'n' {
		return VerifyNumeric(rng, value, n)
	} else if rng.valueType == '4' {
		return VerifyIPv4(rng, value, n)
	}

	return nil
}

func GetRange(inp *Input) (*Range, error) {
	var rangeType *Node
	var err error
	var starRange Range

	// range type
	rangeType, err = GetOctet(inp)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(Alpha, rangeType.Octet.Value) {
		starRange.valueType = 'a'
	} else if bytes.Equal(Numeric, rangeType.Octet.Value) {
		starRange.valueType = 'n'
	} else if bytes.Equal(Date, rangeType.Octet.Value) {
		starRange.valueType = 'd'
	} else if bytes.Equal(Time, rangeType.Octet.Value) {
		starRange.valueType = 't'
	} else if bytes.Equal(Ipv4, rangeType.Octet.Value) {
		starRange.valueType = '4'
	} else if bytes.Equal(Ipv6, rangeType.Octet.Value) {
		starRange.valueType = '6'
	}

	err = GetRestrictions(inp, &starRange, 0)
	if err != nil {
		return nil, err
	}
	if inp.NextByte() != ')' {
		err = GetRestrictions(inp, &starRange, 1)
		if err != nil {
			return nil, err
		}
	}

	return &starRange, nil
}

func GetPrefix(inp *Input) (*Prefix, error) {
	var prefix Prefix
	var err error
	var node *Node

	node, err = GetOctet(inp)
	if err != nil {
		return nil, err
	}
	prefix = Prefix{
		Value: node.Octet.Value,
	}

	return &prefix, err
}

func GetSuffix(inp *Input) (*Suffix, error) {
	var suffix Suffix
	var err error

	suffix = Suffix{}
	err = errors.New("Incorrect suffix value _value_")

	return &suffix, err
}

func FormatIPv4(part [4]int) string {
	return fmt.Sprintf("%d.%d.%d.%d", part[0], part[1], part[2], part[3])
}

func Boundary(rng *Range, n int) string {
	var limit string

	if rng.valueType == '4' {
		limit = FormatIPv4(rng.ipv4Limit[n])
	}
	return fmt.Sprintf(" %s %s", rng.boundary[n], limit)
}

func PrintRange(rng *Range, indent int) {
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	var text string

	text = fmt.Sprintf(" - [%v]", RangeTypes[rng.valueType])
	text += Boundary(rng, 0)
	if rng.boundary[1] != nil {
		text += Boundary(rng, 1)
	}
	fmt.Println(text)
}
