// Code generated by 'yaegi extract git.golaxy.org/framework/addins/dsvc'. DO NOT EDIT.

package fwlib

import (
	"context"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/utils/concurrent"
	"reflect"
	"time"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/dsvc/dsvc"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Install":   reflect.ValueOf(&dsvc.Install).Elem(),
		"Name":      reflect.ValueOf(&dsvc.Name).Elem(),
		"Uninstall": reflect.ValueOf(&dsvc.Uninstall).Elem(),
		"Using":     reflect.ValueOf(&dsvc.Using).Elem(),
		"With":      reflect.ValueOf(&dsvc.With).Elem(),

		// type definitions
		"DistServiceOptions": reflect.ValueOf((*dsvc.DistServiceOptions)(nil)),
		"IDistService":       reflect.ValueOf((*dsvc.IDistService)(nil)),
		"IWatcher":           reflect.ValueOf((*dsvc.IWatcher)(nil)),
		"NodeDetails":        reflect.ValueOf((*dsvc.NodeDetails)(nil)),
		"RecvMsgHandler":     reflect.ValueOf((*dsvc.RecvMsgHandler)(nil)),

		// interface wrapper definitions
		"_IDistService": reflect.ValueOf((*_git_golaxy_org_framework_addins_dsvc_IDistService)(nil)),
		"_IWatcher":     reflect.ValueOf((*_git_golaxy_org_framework_addins_dsvc_IWatcher)(nil)),
	}
}

// _git_golaxy_org_framework_addins_dsvc_IDistService is an interface wrapper for IDistService type
type _git_golaxy_org_framework_addins_dsvc_IDistService struct {
	IValue          interface{}
	WGetFutures     func() *concurrent.Futures
	WGetNodeDetails func() *dsvc.NodeDetails
	WSendMsg        func(dst string, msg gap.Msg) error
	WWatchMsg       func(ctx context.Context, handler dsvc.RecvMsgHandler) dsvc.IWatcher
}

func (W _git_golaxy_org_framework_addins_dsvc_IDistService) GetFutures() *concurrent.Futures {
	return W.WGetFutures()
}
func (W _git_golaxy_org_framework_addins_dsvc_IDistService) GetNodeDetails() *dsvc.NodeDetails {
	return W.WGetNodeDetails()
}
func (W _git_golaxy_org_framework_addins_dsvc_IDistService) SendMsg(dst string, msg gap.Msg) error {
	return W.WSendMsg(dst, msg)
}
func (W _git_golaxy_org_framework_addins_dsvc_IDistService) WatchMsg(ctx context.Context, handler dsvc.RecvMsgHandler) dsvc.IWatcher {
	return W.WWatchMsg(ctx, handler)
}

// _git_golaxy_org_framework_addins_dsvc_IWatcher is an interface wrapper for IWatcher type
type _git_golaxy_org_framework_addins_dsvc_IWatcher struct {
	IValue      interface{}
	WDeadline   func() (deadline time.Time, ok bool)
	WDone       func() <-chan struct{}
	WErr        func() error
	WTerminate  func() async.AsyncRet
	WTerminated func() async.AsyncRet
	WValue      func(key any) any
}

func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Deadline() (deadline time.Time, ok bool) {
	return W.WDeadline()
}
func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Done() <-chan struct{} { return W.WDone() }
func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Err() error            { return W.WErr() }
func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Terminate() async.AsyncRet {
	return W.WTerminate()
}
func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Terminated() async.AsyncRet {
	return W.WTerminated()
}
func (W _git_golaxy_org_framework_addins_dsvc_IWatcher) Value(key any) any { return W.WValue(key) }
