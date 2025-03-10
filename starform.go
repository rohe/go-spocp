package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"time"
)

// type StarForm struct {
// 	valueType byte
// 	boundary  [2][]byte
// 	limit     [2]any
// }

const (
	ALPHA   = "Alpha"
	NUMERIC = "Numeric"
	DATE    = "Date"
	TIME    = "Time"
	IPV4    = "Ipv4"
	IPV6    = "Ipv6"
)

var Alpha = []byte{'a', 'l', 'p', 'h', 'a'}
var Numeric = []byte{'n', 'u', 'm', 'e', 'r', 'i', 'c'}
var Date = []byte{'d', 'a', 't', 'e'}
var Time = []byte{'t', 'i', 'm', 'e'}
var Ipv4 = []byte{'i', 'p', 'v', '4'}
var Ipv6 = []byte{'i', 'p', 'v', '6'}

var limits = []string{"le", "lt", "ge", "gt"}

func CorrectLimit(val string) bool {
	// tests that the given limit type (ge, gt, ...) is one that is expected
	for _, lim := range limits {
		if lim == val {
			return true
		}
	}
	return false
}

func GetLimit(inp *Input) (string, []byte) {
	var gogeLole, value *Node
	var err error
	var limValue string

	gogeLole, err = GetOctet(inp)
	if err != nil {
		return "", []byte("error")
	}
	limValue = string(gogeLole.Octet.Value)
	if CorrectLimit(limValue) == false {

		return "", []byte("Incorrect boundary type " + string(limValue))
	}

	value, err = GetOctet(inp)
	if err != nil {
		return "", []byte("error")
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

func VerifyIPv4(rng *Range, value []byte, n int) error {
	var err error
	var addr netip.Addr

	addr, err = netip.ParseAddr(string(value))
	if err != nil {
		return fmt.Errorf("not an IP address: %v", addr)
	}
	if !addr.Is4() {
		return errors.New("not an IPv4 address, but IPv6")
	}
	rng.ipv4Limit[n] = addr
	return nil
}

func VerifyIPv6(rng *Range, value []byte, n int) error {
	var err error
	var addr netip.Addr

	addr, err = netip.ParseAddr(string(value))
	if err != nil {
		return fmt.Errorf("not an IP address: %v", addr)
	}
	if !addr.Is6() {
		return errors.New("not an IPv6 address, but IPv4")
	}
	rng.ipv4Limit[n] = addr
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

func VerifyDate(rng *Range, value []byte, n int) error {
	var err error

	t, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		return err
	}
	rng.dateLimit[n] = t
	return nil
}

func VerifyTime(rng *Range, value []byte, n int) error {
	var err error

	t, err := time.Parse("15:04:05", string(value))
	if err != nil {
		return err
	}
	rng.timeLimit[n] = t
	return nil
}

func GetRestrictions(inp *Input, rng *Range, n int) error {
	var limit string
	var value []byte

	limit, value = GetLimit(inp)
	rng.boundary[n] = limit

	if rng.valueType == ALPHA {
		return VerifyAlpha(rng, value, n)
	} else if rng.valueType == NUMERIC {
		return VerifyNumeric(rng, value, n)
	} else if rng.valueType == IPV4 {
		return VerifyIPv4(rng, value, n)
	} else if rng.valueType == DATE {
		return VerifyDate(rng, value, n)
	} else if rng.valueType == TIME {
		return VerifyTime(rng, value, n)
	} else if rng.valueType == IPV6 {
		return VerifyIPv6(rng, value, n)
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
		starRange.valueType = ALPHA
	} else if bytes.Equal(Numeric, rangeType.Octet.Value) {
		starRange.valueType = NUMERIC
	} else if bytes.Equal(Date, rangeType.Octet.Value) {
		starRange.valueType = DATE
	} else if bytes.Equal(Time, rangeType.Octet.Value) {
		starRange.valueType = TIME
	} else if bytes.Equal(Ipv4, rangeType.Octet.Value) {
		starRange.valueType = IPV4
	} else if bytes.Equal(Ipv6, rangeType.Octet.Value) {
		starRange.valueType = IPV6
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

func FormatIPv(addr netip.Addr) string {
	return addr.StringExpanded()
}

func FormatNumeric(num int) string {
	return strconv.Itoa(num)
}

func Boundary(rng *Range, n int) string {
	var limit string

	if rng.valueType == IPV4 {
		limit = FormatIPv(rng.ipv4Limit[n])
	} else if rng.valueType == IPV6 {
		limit = FormatIPv(rng.ipv4Limit[n])
	} else if rng.valueType == NUMERIC {
		limit = FormatNumeric(rng.numLimit[n])
	} else if rng.valueType == DATE {
		limit = rng.dateLimit[n].Format(time.RFC3339)
	} else if rng.valueType == ALPHA {
		limit = fmt.Sprintf("%v", rng.alphaLimit[n])
	} else if rng.valueType == TIME {
		limit = rng.timeLimit[n].Format("15:04:00")
	}
	return fmt.Sprintf(" %s %s", rng.boundary[n], limit)
}

func PrintRange(rng *Range, indent int) {
	for ; indent > 0; indent-- {
		fmt.Printf("%s", TAB)
	}
	var text string

	text = fmt.Sprintf(" - [%v]", rng.valueType)
	text += Boundary(rng, 0)
	if rng.boundary[1] != "" {
		text += Boundary(rng, 1)
	}
	fmt.Println(text)
}
