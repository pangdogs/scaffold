// Code generated by 'yaegi extract git.golaxy.org/scaffold/addins/goscr/dynamic'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/scaffold/addins/goscr/dynamic/dynamic"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Func":         reflect.ValueOf(dynamic.Func),
		"NewCodeFs":    reflect.ValueOf(dynamic.NewCodeFs),
		"NewScriptLib": reflect.ValueOf(dynamic.NewScriptLib),
		"NewSolution":  reflect.ValueOf(dynamic.NewSolution),
		"None":         reflect.ValueOf(dynamic.None),
		"Struct":       reflect.ValueOf(dynamic.Struct),

		// type definitions
		"BindMode":     reflect.ValueOf((*dynamic.BindMode)(nil)),
		"CodeFs":       reflect.ValueOf((*dynamic.CodeFs)(nil)),
		"Method":       reflect.ValueOf((*dynamic.Method)(nil)),
		"MethodBinder": reflect.ValueOf((*dynamic.MethodBinder)(nil)),
		"Project":      reflect.ValueOf((*dynamic.Project)(nil)),
		"Script":       reflect.ValueOf((*dynamic.Script)(nil)),
		"ScriptBundle": reflect.ValueOf((*dynamic.ScriptBundle)(nil)),
		"ScriptLib":    reflect.ValueOf((*dynamic.ScriptLib)(nil)),
		"Solution":     reflect.ValueOf((*dynamic.Solution)(nil)),
		"This":         reflect.ValueOf((*dynamic.This)(nil)),
	}
}
