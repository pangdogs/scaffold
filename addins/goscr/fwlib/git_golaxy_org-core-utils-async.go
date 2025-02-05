// Code generated by 'yaegi extract git.golaxy.org/core/utils/async'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/utils/async/async"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"End":               reflect.ValueOf(async.End),
		"ErrAsyncRetClosed": reflect.ValueOf(&async.ErrAsyncRetClosed).Elem(),
		"MakeAsyncRet":      reflect.ValueOf(async.MakeAsyncRet),
		"MakeRet":           reflect.ValueOf(async.MakeRet),
		"Return":            reflect.ValueOf(async.Return),
		"VoidRet":           reflect.ValueOf(&async.VoidRet).Elem(),
		"Yield":             reflect.ValueOf(async.Yield),

		// type definitions
		"AsyncRet": reflect.ValueOf((*async.AsyncRet)(nil)),
		"Callee":   reflect.ValueOf((*async.Callee)(nil)),
		"Caller":   reflect.ValueOf((*async.Caller)(nil)),
		"Ret":      reflect.ValueOf((*async.Ret)(nil)),

		// interface wrapper definitions
		"_Callee": reflect.ValueOf((*_git_golaxy_org_core_utils_async_Callee)(nil)),
		"_Caller": reflect.ValueOf((*_git_golaxy_org_core_utils_async_Caller)(nil)),
	}
}

// _git_golaxy_org_core_utils_async_Callee is an interface wrapper for Callee type
type _git_golaxy_org_core_utils_async_Callee struct {
	IValue                     interface{}
	WPushCallAsync             func(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet
	WPushCallDelegateAsync     func(fun generic.DelegateVar0[any, async.Ret], args ...any) async.AsyncRet
	WPushCallDelegateVoidAsync func(fun generic.DelegateVoidVar0[any], args ...any) async.AsyncRet
	WPushCallVoidAsync         func(fun generic.ActionVar0[any], args ...any) async.AsyncRet
}

func (W _git_golaxy_org_core_utils_async_Callee) PushCallAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return W.WPushCallAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallDelegateAsync(fun generic.DelegateVar0[any, async.Ret], args ...any) async.AsyncRet {
	return W.WPushCallDelegateAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallDelegateVoidAsync(fun generic.DelegateVoidVar0[any], args ...any) async.AsyncRet {
	return W.WPushCallDelegateVoidAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return W.WPushCallVoidAsync(fun, args...)
}

// _git_golaxy_org_core_utils_async_Caller is an interface wrapper for Caller type
type _git_golaxy_org_core_utils_async_Caller struct {
	IValue                 interface{}
	WCallAsync             func(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet
	WCallDelegateAsync     func(fun generic.DelegateVar0[any, async.Ret], args ...any) async.AsyncRet
	WCallDelegateVoidAsync func(fun generic.DelegateVoidVar0[any], args ...any) async.AsyncRet
	WCallVoidAsync         func(fun generic.ActionVar0[any], args ...any) async.AsyncRet
}

func (W _git_golaxy_org_core_utils_async_Caller) CallAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return W.WCallAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallDelegateAsync(fun generic.DelegateVar0[any, async.Ret], args ...any) async.AsyncRet {
	return W.WCallDelegateAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallDelegateVoidAsync(fun generic.DelegateVoidVar0[any], args ...any) async.AsyncRet {
	return W.WCallDelegateVoidAsync(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return W.WCallVoidAsync(fun, args...)
}
