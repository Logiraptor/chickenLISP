package main

func Add(a, b Number) Number {
	return a + b
}

func Mult(a, b Number) Number {
	return a * b
}

func Div(a, b Number) Number {
	return a / b
}

func Sub(a, b Number) Number {
	return a - b
}

func Eq(a, b Atom) Bool {
	switch x := a.(type) {
	case Number:
		return x == b.(Number)
	case Symbol:
		return x == b.(Symbol)
	}
	return false
}

// global_env.mapping["="] = FuncLit(func(l List) (Atom, error) {
// 	if l[0].(Number) == l[1].(Number) {
// 		return true, nil
// 	} else {
// 		return nil, nil
// 	}
// })
// global_env.mapping["car"] = FuncLit(func(l List) (Atom, error) {
// 	return l[0].(List)[0], nil
// })
// global_env.mapping["cdr"] = FuncLit(func(l List) (Atom, error) {
// 	return l[0].(List)[1:], nil
// })
// global_env.mapping["list"] = FuncLit(func(l List) (Atom, error) {
// 	return l, nil
// })
