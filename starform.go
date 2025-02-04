package main

import (
	"bytes"
)

type StarForm struct {
	value_type byte
	boundary1  *Node
	limit1     *Node
	boundary2  *Node
	limit2     *Node
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

func GetLimit(bs []byte, begin int) ([]byte, []byte) {
	var goge_lole, value *Node
	var err error
	var lim_value, value_slice []byte

	goge_lole, err = GetOctet(bs, begin)
	if err != nil {
		return nil, []byte("error")
	}
	lim_value = bs[goge_lole.end-begin : goge_lole.end]
	if CorrectLimit(lim_value) == false {
		return nil, []byte("Incorrect boundary type " + string(lim_value))
	}

	value, err = GetOctet(bs, goge_lole.end+1)
	if err != nil {
		return nil, []byte("error")
	}
	value_slice = bs[value.begin:value.end]

	return lim_value, value_slice
}

func VerifyLimit(bs []byte, limit_typ []byte, value []byte) bool {
	return True
}

func GetRange(bs []byte, begin int) (*StarForm, error) {
	var node *Node
	var err error
	var slice, limit, value []byte

	node, err = GetOctet(bs, begin)
	if err == nil {
		return nil, err
	}

	star_form = StarForm{
		value_type: ' ',
	}

	slice = bs[node.begin:node.end]
	if bytes.Equal(Alpha, slice) {
		star_form.value_type = 'a'
		limit, value = GetLimit(bs, node.end+1)
	} else if bytes.Equal(Numeric, slice) {
		star_form.value_type = 'n'
	} else if bytes.Equal(Date, slice) {
		star_form.value_type = 'd'
	} else if bytes.Equal(Time, slice) {
		star_form.value_type = 't'
	} else if bytes.Equal(Ipv4, slice) {
		star_form.value_type = '4'
	} else if bytes.Equal(Ipv6, slice) {
		star_form.value_type = '6'
	}

	return &start_form, nil
}
