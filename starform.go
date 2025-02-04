package main

import (
	"bytes"
)

type StarForm struct {
	value_type byte
	boundary1  []byte
	limit1     []byte
	boundary2  []byte
	limit2     []byte
}

var Alpha = []byte{'a', 'l', 'p', 'h', 'a'}
var Numeric = []byte{'n', 'u', 'm', 'e', 'r', 'i', 'c'}
var Date = []byte{'d', 'a', 't', 'e'}
var Time = []byte{'t', 'i', 'm', 'e'}
var Ipv4 = []byte{'i', 'p', 'v', '4'}
var Ipv6 = []byte{'i', 'p', 'v', '6'}

var LE = []byte("le")
var LT = []byte("lt")
var GE = []byte("ge")
var GT = []byte("gt")

var limits = [][]byte{LE, LT, GE, GT}

func CorrectLimit(val []byte) bool {
	for _, lim := range limits {
		if bytes.Equal(lim, val) {
			return true
		}
	}
	return false
}

func GetLimit(inp *Input) ([]byte, []byte) {
	var goge_lole, value *Node
	var err error
	var lim_value, value_slice []byte

	goge_lole, err = GetOctet(inp)
	if err != nil {
		return nil, []byte("error")
	}
	lim_value = inp.Slice(goge_lole.begin, goge_lole.end)
	if CorrectLimit(lim_value) == false {
		return nil, []byte("Incorrect boundary type " + string(lim_value))
	}

	value, err = GetOctet(inp)
	if err != nil {
		return nil, []byte("error")
	}
	value_slice = inp.Slice(value.begin, value.end)

	inp.startByte = value.end + 1
	return lim_value, value_slice
}

func VerifyLimit(bs []byte, limit_typ []byte, value []byte) bool {
	return true
}

func GetRange(inp *Input) (*StarForm, error) {
	var node *Node
	var err error
	var slice, limit, value []byte
	var endchr int

	node, err = GetOctet(inp)
	if err == nil {
		return nil, err
	}

	var starForm StarForm

	slice = inp.Slice(node.begin, node.end)
	if bytes.Equal(Alpha, slice) {
		starForm.value_type = 'a'
		limit, value = GetLimit(inp)
		starForm.boundary1 = limit
		starForm.limit1 = value
		if endchr > 0 {
			limit, value = GetLimit(inp)
			starForm.boundary1 = limit
			starForm.limit1 = value
		}
	} else if bytes.Equal(Numeric, slice) {
		starForm.value_type = 'n'
	} else if bytes.Equal(Date, slice) {
		starForm.value_type = 'd'
	} else if bytes.Equal(Time, slice) {
		starForm.value_type = 't'
	} else if bytes.Equal(Ipv4, slice) {
		starForm.value_type = '4'
	} else if bytes.Equal(Ipv6, slice) {
		starForm.value_type = '6'
	}

	return &starForm, nil
}
