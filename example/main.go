package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yowenter/conditions"
)

func main() {
	// cccccOur condition to check
	type people struct {
		Name   string
		Height int32
		Male   bool
		Goods  []string
		Birth  time.Time
	}

	s := ` $Name == "test" AND $Height > 100 AND $Male == false`

	// Parse the condition language and get expression
	p := conditions.NewParser(strings.NewReader(s))
	expr, err := p.Parse()
	if err != nil {
		fmt.Println(err)
		return
		// ...
	}

	// Evaluate expression passing data for $vars
	p1 := map[string]interface{}{"Name": "test", "Height": 180, "Male": false}
	r, err := conditions.Evaluate(expr, p1)
	if err != nil {
		fmt.Println(err)
		// ...
	}
	fmt.Printf("Condition: `%v`, Val: `%v`, Result: `%v`\n", s, p1, r)

	// use struct
	var p2 = people{Name: "test", Height: 200, Male: false, Goods: []string{"A", "B"}}
	r, err = conditions.Evaluate(expr, p2)
	if err != nil {
		fmt.Println(err)
		// ...
	}
	fmt.Printf("Condition: `%v`, Val: `%v`, Result: `%v`\n", s, p2, r)

	// test invalid args . not map or struct.
	r, err = conditions.Evaluate(expr, "")
	if err != nil {
		fmt.Println(err)
		// ...
	}
	fmt.Printf("Condition: `%v`, Val: `%v`, Result: `%v`\n", s, "invalid", r)

	contains := ` ($Goods CONTAINS "A") AND $Name == "test" `

	// Parse the condition language and get expression
	containsP := conditions.NewParser(strings.NewReader(contains))
	containsExpr, err := containsP.Parse()
	if err != nil {
		fmt.Println(err)
		return
		// ...
	}
	r, err = conditions.Evaluate(containsExpr, p2)
	if err != nil {
		fmt.Println(err)
		// ...
	}
	fmt.Printf("Condition: `%v`, Val: `%v`, Result: `%v`\n", contains, p2, r)

	s3 := ` $Birth BEFORE 1`
	beforeP := conditions.NewParser(strings.NewReader(s3))
	beforeExpr, err := beforeP.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}
	r, err = conditions.Evaluate(beforeExpr, people{Birth: time.Now().Add(-time.Hour * 48)})
	fmt.Println(r, err)
}
