package main

import (
	"log"
)

var TAB = []byte{32, 32, 32, 32}

type Input struct {
	bs              []byte
	currentPosition int
}

func (inp Input) Remaining() int {
	return len(inp.bs) - inp.currentPosition
}
func (inp Input) NextByte() byte {
	return inp.bs[inp.currentPosition]
}
func (inp Input) Slice(begin int, end int) []byte {
	return inp.bs[begin:end]
}
func (inp Input) Prefix(length int) []byte {
	return inp.bs[inp.currentPosition : inp.currentPosition+length]
}
func (inp Input) RemainingString() string {
	return string(inp.bs[inp.currentPosition:])
}

func main() {
	// s := "(gopher foo)"
	// s := "(11:certificate(6:issuer3:bob)(7:subject5:alice))"
	var SExpressions = map[string][]string{
		// "(11:certificate(6:issuer3:bob)(7:subject))": []string{"(11:certificate(6:issuer3:bob)(7:subject5:alice))"},
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range7:numeric2:le3:100)))": {"(11:certificate(" +
		//	"6:issuer3:bob)(5:level2:99))", "(11:certificate(6:issuer3:bob)(5:level3:101))"},
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range5:alpha2:ge3:abc)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv42:ge11:130.239.1.12:lt13:130.239.1.127)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge39:FEDC:BA98:7654:3210:FEDC:BA98:7654:3210)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge26:1080:0:0:0:8:800:200C:417A)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge21:1080::8:800:200C:417A)))",
		//
		// "(11:certificate(6:issuer3:bob)(5:fruit(1:*3:set5:apple6:orange5:lemon)))": {
		// 	"(11:certificate(6:issuer3:bob)(5:fruit5:apple)))",
		// 	"(11:certificate(6:issuer3:bob)(5:fruit6:orange)))",
		// 	"(11:certificate(6:issuer3:bob)(5:fruit5:lemon)))",
		// 	"(11:certificate(6:issuer3:bob)(5:fruit4:pear)))"},
		//
		// "(1:t(1:*3:set(1:a1:b)(1:c(1:d1:e))(1:f)1:g))",
		//
		// "(11:certificate(6:issuer3:bob)(4:when(1:*5:range4:date2:ge25:2023-12-22T17:25:33+01:00)))": {
		// 	"(11:certificate(6:issuer3:bob)(4:when25:2025-03-05T11:00:00+01:00))",
		// 	"(11:certificate(6:issuer3:bob)(4:when25:2020-03-05T11:00:00+01:00))",
		// },
		//
		// "(11:certificate(6:issuer3:bob)(4:when(1:*5:range4:date2:ge25:2023-12-22T17:25:33+01:002:le25:2030-12-31T23" +
		// 	":59:59+01:00)))": {
		// 	"(11:certificate(6:issuer3:bob)(4:when25:2025-03-05T11:00:00+01:00))",
		// 	"(11:certificate(6:issuer3:bob)(4:when25:2020-03-05T11:00:00+01:00))",
		// 	"(11:certificate(6:issuer3:bob)(4:when25:2035-03-05T11:00:00+01:00))",
		// },
		//
		"(11:certificate(6:issuer3:bob)(4:when(1:*5:range4:time2:ge8:10:30:00)))": {
			"(11:certificate(6:issuer3:bob)(4:when8:12:00:00)))",
			"(11:certificate(6:issuer3:bob)(4:when8:09:00:00)))",
		},
	}
	for stringRule, QueryList := range SExpressions {
		var Rule, Query *Node
		var err error
		var cmp bool

		// Skip the first '('
		var inp = Input{[]byte(stringRule), 1}

		Rule, err = GetSexp(&inp)
		if err != nil {
			log.Fatal("Parse error")
		}
		for _, query := range QueryList {
			println(query)
			// Skip the first '('
			inp = Input{[]byte(query), 1}

			Query, err = GetSexp(&inp)
			if err != nil {
				log.Fatal("Parse error")
			}
			cmp, err = Query.Compare(Rule)
			println(cmp)
		}
	}
}
