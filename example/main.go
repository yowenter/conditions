package main

import (
	"fmt"
	"strings"

	"github.com/yowenter/conditions"
)

func main() {
	// Our condition to check
	type people struct {
		Name   string
		Height int32
		Male   bool
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
	var p2 = people{Name: "test", Height: 200, Male: false}
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
}
