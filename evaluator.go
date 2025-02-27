package conditions

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	falseExpr = &BooleanLiteral{Val: false}
)

// Evaluate takes an expr and evaluates it using given args
func Evaluate(expr Expr, args interface{}) (bool, error) {
	if expr == nil {
		return false, fmt.Errorf("Provided expression is nil")
	}

	result, err := evaluateSubtree(expr, args)
	if err != nil {
		return false, err
	}
	switch n := result.(type) {
	case *BooleanLiteral:
		return n.Val, nil
	}
	return false, fmt.Errorf("Unexpected result of the root expression: %#v", result)
}

// evaluateSubtree performs given expr evaluation recursively
func evaluateSubtree(expr Expr, args interface{}) (Expr, error) {
	if expr == nil {
		return falseExpr, fmt.Errorf("Provided expression is nil")
	}

	var (
		err    error
		lv, rv Expr
	)

	switch n := expr.(type) {
	case *ParenExpr:
		return evaluateSubtree(n.Expr, args)
	case *BinaryExpr:
		lv, err = evaluateSubtree(n.LHS, args)
		if err != nil {
			return falseExpr, err
		}
		rv, err = evaluateSubtree(n.RHS, args)
		if err != nil {
			return falseExpr, err
		}
		return applyOperator(n.Op, lv, rv)
	case *VarRef:
		//index, err := strconv.Atoi(strings.Replace(n.Val, "$", "", -1))
		index := n.Val
		if err != nil {
			return falseExpr, fmt.Errorf("Failed to resolve argument index %s: %s", n.Val, err.Error())
		}
		argsKind := reflect.TypeOf(args).Kind()
		var val interface{}

		switch argsKind {
		case reflect.Map:
			argsMap, ok := args.(map[string]interface{})
			if !ok {
				return falseExpr, fmt.Errorf("Args: `%v` convert to map not ok", args)
			}
			if _, ok := argsMap[index]; !ok {
				return falseExpr, fmt.Errorf("Argument: `%v` not found", index)
			}
			val, _ = argsMap[index]
		case reflect.Struct:
			ps := reflect.ValueOf(args)
			fval := ps.FieldByName(index)
			if !fval.IsValid() {
				return falseExpr, fmt.Errorf("Argument: `%v` not found in args `%v`", index, args)
			}
			val = fval.Interface()
		default:
			return falseExpr, fmt.Errorf("Args: `%v` is not map or struct", args)
		}
		if t, ok := val.(time.Time); ok {
			return &TimeLiteral{Val: t}, nil
		}
		kind := reflect.TypeOf(val).Kind()
		switch kind {
		case reflect.Int:
			return &NumberLiteral{Val: float64(val.(int))}, nil
		case reflect.Int32:
			return &NumberLiteral{Val: float64(val.(int32))}, nil
		case reflect.Int64:
			return &NumberLiteral{Val: float64(val.(int64))}, nil
		case reflect.Float32:
			return &NumberLiteral{Val: float64(val.(float32))}, nil
		case reflect.Float64:
			return &NumberLiteral{Val: float64(val.(float64))}, nil
		case reflect.String:
			return &StringLiteral{Val: val.(string)}, nil
		case reflect.Bool:
			return &BooleanLiteral{Val: val.(bool)}, nil
		case reflect.Slice:
			return &SliceStringLiteral{Val: val.([]string)}, nil
		}
		return falseExpr, fmt.Errorf("Unsupported argument %s type: %s", n.Val, kind)
	}

	return expr, nil
}

// applyOperator is a dispatcher of the evaluation according to operator
func applyOperator(op Token, l, r Expr) (*BooleanLiteral, error) {
	switch op {
	case AND:
		return applyAND(l, r)
	case OR:
		return applyOR(l, r)
	case EQ:
		return applyEQ(l, r)
	case NEQ:
		return applyNQ(l, r)
	case GT:
		return applyGT(l, r)
	case GTE:
		return applyGTE(l, r)
	case LT:
		return applyLT(l, r)
	case LTE:
		return applyLTE(l, r)
	case XOR:
		return applyXOR(l, r)
	case NAND:
		return applyNAND(l, r)
	case IN:
		return applyIN(l, r)
	case CONTAINS:
		return applyContains(l, r)
	case BEFORE:
		return applyBefore(l, r)
	case NOTIN:
		return applyNOTIN(l, r)
	case EREG:
		return applyEREG(l, r)
	case NEREG:
		return applyNEREG(l, r)
	}
	return &BooleanLiteral{Val: false}, fmt.Errorf("Unsupported operator: %s", op)
}

// applyEREG applies EREG operation to l/r operands
func applyNEREG(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyEREG(l, r)
	result.Val = !result.Val
	return result, err
}

// applyEREG applies EREG operation to l/r operands
func applyEREG(l, r Expr) (*BooleanLiteral, error) {
	var (
		a     string
		b     string
		err   error
		match bool
	)
	a, err = getString(l)
	if err != nil {
		return nil, err
	}

	b, err = getString(r)
	if err != nil {
		return nil, err
	}
	match = false
	match, err = regexp.MatchString(b, a)

	// pp.Print(a, b, match)
	return &BooleanLiteral{Val: match}, err
}

// applyNOTIN applies NOT IN operation to l/r operands
func applyNOTIN(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyIN(l, r)
	result.Val = !result.Val
	return result, err
}

func applyBefore(l, r Expr) (*BooleanLiteral, error) {
	switch t := l.(type) {
	case *TimeLiteral:
		dt, err := getTime(l)
		if err != nil {
			return nil, err
		}

		switch r.(type) {
		case *NumberLiteral:
			days, err := getNumber(r)
			if err != nil {
				return nil, err
			}
			dur := time.Duration(days) * time.Second * 86400
			if time.Since(dt) > dur {
				return &BooleanLiteral{
					Val: true,
				}, nil
			} else {
				return &BooleanLiteral{
					Val: false,
				}, nil
			}

		}
	default:
		return nil, fmt.Errorf("Can not evaluate Literal of unknow type %s %T", t, t)
	}

	return &BooleanLiteral{Val: false}, nil

}

// applyContains applies CONTAINS to l/r operations
func applyContains(l, r Expr) (*BooleanLiteral, error) {
	var (
		err error
		in  bool
	)
	switch t := r.(type) {
	case *StringLiteral:
		var a string
		var b []string
		var bb string
		a, err = getString(r)
		if err != nil {
			return nil, err
		}

		var ltIsString bool
		switch l.(type) {
		case *StringLiteral:
			ltIsString = true
		}
		if ltIsString {
			bb, err = getString(l)
			if err != nil {
				return nil, err
			}
			return &BooleanLiteral{
				Val: strings.Contains(bb, a),
			}, nil
		}

		b, err = getSliceString(l)

		if err != nil {
			return nil, err
		}

		in = false
		for _, e := range b {
			if a == e {
				in = true
			}
		}
	case *NumberLiteral:
		var a float64
		var b []float64
		a, err = getNumber(r)
		if err != nil {
			return nil, err
		}

		b, err = getSliceNumber(l)

		if err != nil {
			return nil, err
		}

		in = false
		for _, e := range b {
			if a == e {
				in = true
			}
		}
	default:
		return nil, fmt.Errorf("Can not evaluate Literal of unknow type %s %T", t, t)
	}

	return &BooleanLiteral{Val: in}, nil
}

// applyIN applies IN operation to l/r operands
func applyIN(l, r Expr) (*BooleanLiteral, error) {
	var (
		err   error
		found bool
	)
	// pp.Print(l)
	switch t := l.(type) {
	case *StringLiteral:
		var a string
		var b []string
		a, err = getString(l)
		if err != nil {
			return nil, err
		}

		b, err = getSliceString(r)

		if err != nil {
			return nil, err
		}

		found = false
		for _, e := range b {
			if a == e {
				found = true
			}
		}
	case *NumberLiteral:
		var a float64
		var b []float64
		a, err = getNumber(l)
		if err != nil {
			return nil, err
		}

		b, err = getSliceNumber(r)

		if err != nil {
			return nil, err
		}

		found = false
		for _, e := range b {
			if a == e {
				found = true
			}
		}
	default:
		return nil, fmt.Errorf("Can not evaluate Literal of unknow type %s %T", t, t)
	}

	return &BooleanLiteral{Val: found}, nil
}

// applyXOR applies || operation to l/r operands
func applyXOR(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a != b)}, nil
}

// applyNAND applies NAND operation to l/r operands
func applyNAND(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (!(a && b))}, nil
}

// applyAND applies && operation to l/r operands
func applyAND(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a && b)}, nil
}

// applyOR applies || operation to l/r operands
func applyOR(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a || b)}, nil
}

// applyEQ applies == operation to l/r operands
func applyEQ(l, r Expr) (*BooleanLiteral, error) {
	var (
		as, bs string
		an, bn float64
		ab, bb bool
		err    error
	)
	as, err = getString(l)
	if err == nil {
		bs, err = getString(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare string with non-string")
		}
		return &BooleanLiteral{Val: (as == bs)}, nil
	}
	an, err = getNumber(l)
	if err == nil {
		bn, err = getNumber(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare number with non-number")
		}
		return &BooleanLiteral{Val: (an == bn)}, nil
	}
	ab, err = getBoolean(l)
	if err == nil {
		bb, err = getBoolean(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare boolean with non-boolean")
		}
		return &BooleanLiteral{Val: (ab == bb)}, nil
	}
	return falseExpr, nil
}

// applyNQ applies != operation to l/r operands
func applyNQ(l, r Expr) (*BooleanLiteral, error) {
	var (
		as, bs string
		an, bn float64
		ab, bb bool
		err    error
	)
	as, err = getString(l)
	if err == nil {
		bs, err = getString(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare string with non-string")
		}
		return &BooleanLiteral{Val: (as != bs)}, nil
	}
	an, err = getNumber(l)
	if err == nil {
		bn, err = getNumber(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare number with non-number")
		}
		return &BooleanLiteral{Val: (an != bn)}, nil
	}
	ab, err = getBoolean(l)
	if err == nil {
		bb, err = getBoolean(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare boolean with non-boolean")
		}
		return &BooleanLiteral{Val: (ab != bb)}, nil
	}
	return falseExpr, nil
}

// applyGT applies > operation to l/r operands
func applyGT(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a > b)}, nil
}

// applyGTE applies >= operation to l/r operands
func applyGTE(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a >= b)}, nil
}

// applyLT applies < operation to l/r operands
func applyLT(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a < b)}, nil
}

// applyLTE applies <= operation to l/r operands
func applyLTE(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return falseExpr, err
	}
	b, err = getNumber(r)
	if err != nil {
		return falseExpr, err
	}
	return &BooleanLiteral{Val: (a <= b)}, nil
}

// getBoolean performs type assertion and returns boolean value or error
func getBoolean(e Expr) (bool, error) {
	switch n := e.(type) {
	case *BooleanLiteral:
		return n.Val, nil
	default:
		return false, fmt.Errorf("Literal is not a boolean: %v", n)
	}
}

// getTime performs type assertion and returns string value or error
func getTime(e Expr) (time.Time, error) {
	switch t := e.(type) {
	case *TimeLiteral:
		return t.Val, nil
	default:
		return time.Time{}, fmt.Errorf("Literal is not a time: %v", t)
	}
}

// getTime performs type assertion and returns string value or error
func getTimeDuration(e Expr) (time.Duration, error) {
	switch t := e.(type) {
	case *DurationLiteral:
		return t.Val, nil
	default:
		return time.Second * 0, fmt.Errorf("Literal is not a time: %v", t)
	}
}

// getString performs type assertion and returns string value or error
func getString(e Expr) (string, error) {
	switch n := e.(type) {
	case *StringLiteral:
		return n.Val, nil
	default:
		return "", fmt.Errorf("Literal is not a string: %v", n)
	}
}

// getSliceNumber performs type assertion and returns []float64 value or error
func getSliceNumber(e Expr) ([]float64, error) {
	switch n := e.(type) {
	case *SliceNumberLiteral:
		return n.Val, nil
	default:
		return []float64{}, fmt.Errorf("Literal is not a slice of float64: %v", n)
	}
}

// getSliceString performs type assertion and returns []string value or error
func getSliceString(e Expr) ([]string, error) {
	switch n := e.(type) {
	case *SliceStringLiteral:
		return n.Val, nil
	default:
		return []string{}, fmt.Errorf("Literal is not a slice of string: %v", n)
	}
}

// getNumber performs type assertion and returns float64 value or error
func getNumber(e Expr) (float64, error) {
	switch n := e.(type) {
	case *NumberLiteral:
		return n.Val, nil
	default:
		return 0, fmt.Errorf("Literal is not a number: %v", n)
	}
}
