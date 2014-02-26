package main

import (
	"errors"
	"fmt"
	"reflect"
)

// MustWrap calls Wrap, and panics on error. Convenient for static initialization.
func MustWrap(f interface{}) func(List) (Atom, error) {
	wf, err := Wrap(f)
	if err != nil {
		panic(err)
	}
	return wf
}

// Wrap takes a function f and converts it into a function which takes interface{} as input and returns []interface{}.
func Wrap(f interface{}) (func(List) (Atom, error), error) {
	// get arg types
	funcCal := reflect.TypeOf(f)
	if funcCal.Kind() != reflect.Func {
		return nil, errors.New("wrap: f is not a function.")
	}
	numArgs := funcCal.NumIn()
	inTypes := make([]reflect.Type, numArgs)
	for i := 0; i < numArgs; i++ {
		inTypes[i] = funcCal.In(i)
	}

	numRets := funcCal.NumOut()
	outTypes := make([]reflect.Type, numRets)
	for i := 0; i < numRets; i++ {
		outTypes[i] = funcCal.Out(i)
	}

	funcVal := reflect.ValueOf(f)

	return func(args List) (Atom, error) {
		if len(args) != len(inTypes) && !funcCal.IsVariadic() {
			return nil, fmt.Errorf("function takes %d arguments, got %d", len(inTypes), len(args))
		}
		argVals := make([]reflect.Value, len(args))
		for i := range args {
			argVals[i] = reflect.ValueOf(args[i])
		}

		// validates arg types
		for i := range inTypes {
			if !argVals[i].Type().AssignableTo(inTypes[i]) {
				// if this is the variadic parameter
				if funcCal.IsVariadic() && i == len(inTypes)-1 {
					elemType := inTypes[i].Elem()
					for j := range argVals[i:] {
						if !argVals[j+i].Type().AssignableTo(elemType) {
							return nil, fmt.Errorf("variadic type error: %v is not assignable to %v", argVals[i+j].Type(), elemType)
						}
					}
				} else {
					return nil, fmt.Errorf("type error: %v is not assignable to %v", argVals[i].Type(), inTypes[i])
				}
			}
		}

		// call f
		rVals := funcVal.Call(argVals)

		// transform return values back into native types
		retVals := make(List, len(rVals))
		var err error
		for i := range rVals {
			retVals[i], err = convert(rVals[i])
			if err != nil {
				return nil, err
			}
		}
		if len(retVals) == 1 {
			return retVals[0], nil
		}
		return retVals, nil
	}, nil
}

func convert(val reflect.Value) (Atom, error) {
	switch val.Kind() {
	case reflect.Bool:
		return Bool(val.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Number(float64(val.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return Number(val.Float()), nil
	// case reflect.Complex64, reflect.Complex128:
	case reflect.String:
		return Symbol(val.String()), nil
		// case reflect.Map:
		// case reflect.Struct:
		// case reflect.Interface:
		// case reflect.Array, reflect.Slice:
		// case reflect.Ptr:
		// case reflect.Chan, reflect.Func, reflect.UnsafePointer:
	}
	return nil, fmt.Errorf("invalid return type %v", val.Type())
}
