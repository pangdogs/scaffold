// Code generated by 'yaegi extract git.golaxy.org/framework/addins/rpc/callpath'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/addins/rpc/callpath"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/rpc/callpath/callpath"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Cache":   reflect.ValueOf(callpath.Cache),
		"Client":  reflect.ValueOf(callpath.Client),
		"Entity":  reflect.ValueOf(callpath.Entity),
		"Parse":   reflect.ValueOf(callpath.Parse),
		"Runtime": reflect.ValueOf(callpath.Runtime),
		"Service": reflect.ValueOf(callpath.Service),

		// type definitions
		"CallPath": reflect.ValueOf((*callpath.CallPath)(nil)),
		"Category": reflect.ValueOf((*callpath.Category)(nil)),
	}
}
