// Code generated by 'yaegi extract git.golaxy.org/core/ec/ectx'. DO NOT EDIT.

package fwlib

import (
	"context"
	"git.golaxy.org/core/ec/ectx"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/iface"
	"reflect"
	"sync"
	"time"
)

func init() {
	Symbols["git.golaxy.org/core/ec/ectx/ectx"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"UnsafeContext": reflect.ValueOf(ectx.UnsafeContext),

		// type definitions
		"ConcurrentContextProvider": reflect.ValueOf((*ectx.ConcurrentContextProvider)(nil)),
		"Context":                   reflect.ValueOf((*ectx.Context)(nil)),
		"ContextBehavior":           reflect.ValueOf((*ectx.ContextBehavior)(nil)),
		"CurrentContextProvider":    reflect.ValueOf((*ectx.CurrentContextProvider)(nil)),

		// interface wrapper definitions
		"_ConcurrentContextProvider": reflect.ValueOf((*_git_golaxy_org_core_ec_ectx_ConcurrentContextProvider)(nil)),
		"_Context":                   reflect.ValueOf((*_git_golaxy_org_core_ec_ectx_Context)(nil)),
		"_CurrentContextProvider":    reflect.ValueOf((*_git_golaxy_org_core_ec_ectx_CurrentContextProvider)(nil)),
	}
}

// _git_golaxy_org_core_ec_ectx_ConcurrentContextProvider is an interface wrapper for ConcurrentContextProvider type
type _git_golaxy_org_core_ec_ectx_ConcurrentContextProvider struct {
	IValue                interface{}
	WGetConcurrentContext func() iface.Cache
}

func (W _git_golaxy_org_core_ec_ectx_ConcurrentContextProvider) GetConcurrentContext() iface.Cache {
	return W.WGetConcurrentContext()
}

// _git_golaxy_org_core_ec_ectx_Context is an interface wrapper for Context type
type _git_golaxy_org_core_ec_ectx_Context struct {
	IValue            interface{}
	WDeadline         func() (deadline time.Time, ok bool)
	WDone             func() <-chan struct{}
	WErr              func() error
	WGetAutoRecover   func() bool
	WGetParentContext func() context.Context
	WGetReportError   func() chan error
	WGetWaitGroup     func() *sync.WaitGroup
	WTerminate        func() async.AsyncRet
	WTerminated       func() async.AsyncRet
	WValue            func(key any) any
}

func (W _git_golaxy_org_core_ec_ectx_Context) Deadline() (deadline time.Time, ok bool) {
	return W.WDeadline()
}
func (W _git_golaxy_org_core_ec_ectx_Context) Done() <-chan struct{} { return W.WDone() }
func (W _git_golaxy_org_core_ec_ectx_Context) Err() error            { return W.WErr() }
func (W _git_golaxy_org_core_ec_ectx_Context) GetAutoRecover() bool  { return W.WGetAutoRecover() }
func (W _git_golaxy_org_core_ec_ectx_Context) GetParentContext() context.Context {
	return W.WGetParentContext()
}
func (W _git_golaxy_org_core_ec_ectx_Context) GetReportError() chan error { return W.WGetReportError() }
func (W _git_golaxy_org_core_ec_ectx_Context) GetWaitGroup() *sync.WaitGroup {
	return W.WGetWaitGroup()
}
func (W _git_golaxy_org_core_ec_ectx_Context) Terminate() async.AsyncRet  { return W.WTerminate() }
func (W _git_golaxy_org_core_ec_ectx_Context) Terminated() async.AsyncRet { return W.WTerminated() }
func (W _git_golaxy_org_core_ec_ectx_Context) Value(key any) any          { return W.WValue(key) }

// _git_golaxy_org_core_ec_ectx_CurrentContextProvider is an interface wrapper for CurrentContextProvider type
type _git_golaxy_org_core_ec_ectx_CurrentContextProvider struct {
	IValue                interface{}
	WGetConcurrentContext func() iface.Cache
	WGetCurrentContext    func() iface.Cache
}

func (W _git_golaxy_org_core_ec_ectx_CurrentContextProvider) GetConcurrentContext() iface.Cache {
	return W.WGetConcurrentContext()
}
func (W _git_golaxy_org_core_ec_ectx_CurrentContextProvider) GetCurrentContext() iface.Cache {
	return W.WGetCurrentContext()
}
