package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Logiraptor/chicken/peg"
)

var QUOTE Symbol = "quote"
var IF Symbol = "if"
var SET Symbol = "set"
var DEFINE Symbol = "define"
var LAMBDA Symbol = "lambda"
var BEGIN Symbol = "begin"

type Env struct {
	mapping map[Symbol]Atom
	outer   *Env
}

func NewEnv() *Env {
	return &Env{
		mapping: make(map[Symbol]Atom),
	}
}

func NewEnvFrom(vars, args List, outer *Env) *Env {
	e := NewEnv()
	e.outer = outer
	for i := range vars {
		e.mapping[vars[i].(Symbol)] = args[i]
	}
	return e
}

func (e *Env) find(s Symbol) map[Symbol]Atom {
	if _, ok := e.mapping[s]; ok {
		return e.mapping
	} else if e.outer != nil {
		return e.outer.find(s)
	} else {
		return nil
	}
}

type Atom interface{}

type List []Atom
type Number float64
type Symbol string
type Bool bool

type FuncLit func(List) (Atom, error)

func (f FuncLit) Call(l List) (Atom, error) {
	return f(l)
}

type Procedure struct {
	params List
	body   Atom
	env    *Env
}

func (p *Procedure) Call(l List) (Atom, error) {
	return nil, fmt.Errorf("procedure call: tail recursion is required")
}

type Func interface {
	Call(List) (Atom, error)
}

var global_env = NewEnv()

func init() {
	global_env.mapping["+"] = FuncLit(MustWrap(Add))
	global_env.mapping["*"] = FuncLit(MustWrap(Mult))
	global_env.mapping["/"] = FuncLit(MustWrap(Div))
	global_env.mapping["-"] = FuncLit(MustWrap(Sub))
	global_env.mapping["="] = FuncLit(MustWrap(Eq))
	global_env.mapping["car"] = FuncLit(func(l List) (Atom, error) {
		return l[0].(List)[0], nil
	})
	global_env.mapping["cdr"] = FuncLit(func(l List) (Atom, error) {
		return l[0].(List)[1:], nil
	})
	global_env.mapping["list"] = FuncLit(func(l List) (Atom, error) {
		return l, nil
	})

	var err error
	Parser, err = peg.NewParser(strings.NewReader(listGrammar))
	if err != nil {
		fmt.Println(err)
		return
	}
}

var Parser *peg.Language

func eval(a Atom, e *Env) (Atom, error) {
	var err error
	for {
		switch a.(type) {
		case Number:
			return a, nil
		case Symbol:
			s := a.(Symbol)
			return e.find(s)[s], nil
		case List:
			l := a.(List)
			switch l[0] {
			case QUOTE:
				return l[1], nil
			case IF:
				test, conseq, alt := l[1], l[2], l[3]
				t, err := eval(test, e)
				if err != nil {
					return nil, err
				}
				if t.(Bool) {
					a = conseq
				} else {
					a = alt
				}
			case SET:
				name, exp := l[1].(Symbol), l[2]
				e.find(name)[name], err = eval(exp, e)
				return nil, err
			case DEFINE:
				name, exp := l[1].(Symbol), l[2]
				e.mapping[name], err = eval(exp, e)
				return nil, err
			case LAMBDA:
				vars, exp := l[1].(List), l[2]
				return &Procedure{
					vars, exp, e,
				}, nil
			case BEGIN:
				for _, exp := range l[1 : len(l)-1] {
					_, err = eval(exp, e)
					if err != nil {
						return nil, err
					}
				}
				a = l[len(l)-1]
			default:
				exps := make([]Atom, len(l))
				for i, exp := range l {
					exps[i], err = eval(exp, e)
					if err != nil {
						return nil, err
					}
				}
				proc := exps[0].(Func)
				if p, ok := proc.(*Procedure); ok {
					a = p.body
					e = NewEnvFrom(p.params, exps[1:], p.env)
				} else {
					x, err := proc.Call(exps[1:])
					if err != nil {
						return nil, fmt.Errorf("%s: %s", l[0], err)
					}
					return x, nil
				}
			}
		default:
			// panic(fmt.Sprintf("Cannot eval %v", a))
		}
	}
	return nil, nil
}

func transformPRGM(p *peg.ParseTree, out chan List) {
	for _, child := range p.Children {
		out <- transform(child).(List)
	}
	close(out)
}

func transform(p *peg.ParseTree) Atom {
	switch p.Type {
	case "number":
		val, err := strconv.ParseFloat(string(p.Data), 64)
		if err != nil {
			panic(err)
		}
		return Number(val)
	case "name", "op":
		return Symbol(string(p.Data))
	case "list":
		var resp = make(List, len(p.Children[1].Children))
		for i, child := range p.Children[1].Children {
			resp[i] = transform(child)
		}
		return resp
	default:
		fmt.Println(p)
		panic(fmt.Sprintf("Cannot transform %s", p.Type))
	}
}
