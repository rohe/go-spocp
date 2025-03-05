package main

import (
	"fmt"
	"log"
	"testing"
)

func TestSexp(t *testing.T) {
	var s_expressions = []string{
		"(11:certificate(6:issuer3:bob)(5:level(1:*5:range7:numeric2:ge3:100)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range5:alpha2:ge3:abc)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv42:ge11:130.239.1.12:lt13:130.239.1.127)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge39:FEDC:BA98:7654:3210:FEDC:BA98:7654:3210)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge26:1080:0:0:0:8:800:200C:417A)))",
		// "(11:certificate(6:issuer3:bob)(5:level(1:*5:range4:ipv62:ge21:1080::8:800:200C:417A)))",
		// "(11:certificate(6:issuer3:bob)(5:fruit(1:*3:set5:apple6:orange5:lemon)))",
		// "(1:t(1:*3:set(1:a1:b)(1:c(1:d1:e))(1:f)1:g))",
		// "(1:t(1:*3:set(1:a(1:x1:y))(1:b1:c)(1:a1:d)))", // invalid
		// "(11:certificate(6:issuer3:bob)(4:when(1:*5:range4:date2:ge25:2023-12-22T17:25:33+01:00)))",
		// "(11:certificate(6:issuer3:bob)(4:when(1:*5:range4:time2:ge8:10:30:00)))",
	}
	for _, expression := range s_expressions {
		bs := []byte(expression)
		var SExpression *Node
		var err error
		// Skip the first '('
		var inp = Input{bs, 1}

		SExpression, err = GetSexp(&inp)
		if err != nil {
			log.Fatal("Parse error")
		}
		fmt.Println("Done")
		PrintSExpression(SExpression, 0)
	}
}

func TestSexpCmp(t *testing.T) {
	var Rule = []string{
		"(11:certificate(6:issuer3:bob)(7:subject))",
	}
	var Query = []string{
		"(11:certificate(6:issuer3:bob)(7:subject5:alice))",
	}

	for n := range len(Rule) {
		var rule, query *Node
		var err error
		var cmp bool

		var inp = Input{[]byte(Rule[n]), 1}
		rule, err = GetSexp(&inp)
		if err != nil {
			log.Fatal("Parse error")
		}

		inp = Input{[]byte(Query[n]), 1}
		query, err = GetSexp(&inp)
		if err != nil {
			log.Fatal("Parse error")
		}
		cmp, err = query.Compare(rule)
		if err != nil {
			log.Fatal("compare failed")
		}
		println(cmp)
	}
}
