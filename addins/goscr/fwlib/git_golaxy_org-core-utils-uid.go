// Code generated by 'yaegi extract git.golaxy.org/core/utils/uid'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/utils/uid/uid"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"From": reflect.ValueOf(&uid.From).Elem(),
		"New":  reflect.ValueOf(&uid.New).Elem(),
		"Nil":  reflect.ValueOf(&uid.Nil).Elem(),

		// type definitions
		"Id": reflect.ValueOf((*uid.Id)(nil)),
	}
}
