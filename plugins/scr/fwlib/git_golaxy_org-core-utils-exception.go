// Code generated by 'yaegi extract git.golaxy.org/core/utils/exception'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/exception"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/utils/exception/exception"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ErrArgs":     reflect.ValueOf(&exception.ErrArgs).Elem(),
		"ErrCore":     reflect.ValueOf(&exception.ErrCore).Elem(),
		"ErrPanicked": reflect.ValueOf(&exception.ErrPanicked).Elem(),
		"Error":       reflect.ValueOf(exception.Error),
		"ErrorSkip":   reflect.ValueOf(exception.ErrorSkip),
		"Errorf":      reflect.ValueOf(exception.Errorf),
		"ErrorfSkip":  reflect.ValueOf(exception.ErrorfSkip),
		"Panic":       reflect.ValueOf(exception.Panic),
		"PanicSkip":   reflect.ValueOf(exception.PanicSkip),
		"Panicf":      reflect.ValueOf(exception.Panicf),
		"PanicfSkip":  reflect.ValueOf(exception.PanicfSkip),
		"TraceStack":  reflect.ValueOf(exception.TraceStack),
	}
}
