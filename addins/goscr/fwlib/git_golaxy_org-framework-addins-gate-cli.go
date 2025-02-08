// Code generated by 'yaegi extract git.golaxy.org/framework/addins/gate/cli'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/gate/cli"
	"reflect"
	"time"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/gate/cli/cli"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Connect":            reflect.ValueOf(cli.Connect),
		"ErrInactiveTimeout": reflect.ValueOf(&cli.ErrInactiveTimeout).Elem(),
		"ErrReconnectFailed": reflect.ValueOf(&cli.ErrReconnectFailed).Elem(),
		"Reconnect":          reflect.ValueOf(cli.Reconnect),
		"TCP":                reflect.ValueOf(cli.TCP),
		"WebSocket":          reflect.ValueOf(cli.WebSocket),
		"With":               reflect.ValueOf(&cli.With).Elem(),

		// type definitions
		"Client":           reflect.ValueOf((*cli.Client)(nil)),
		"ClientOptions":    reflect.ValueOf((*cli.ClientOptions)(nil)),
		"IWatcher":         reflect.ValueOf((*cli.IWatcher)(nil)),
		"NetProtocol":      reflect.ValueOf((*cli.NetProtocol)(nil)),
		"RecvDataHandler":  reflect.ValueOf((*cli.RecvDataHandler)(nil)),
		"RecvEventHandler": reflect.ValueOf((*cli.RecvEventHandler)(nil)),
		"ResponseTime":     reflect.ValueOf((*cli.ResponseTime)(nil)),

		// interface wrapper definitions
		"_IWatcher": reflect.ValueOf((*_git_golaxy_org_framework_addins_gate_cli_IWatcher)(nil)),
	}
}

// _git_golaxy_org_framework_addins_gate_cli_IWatcher is an interface wrapper for IWatcher type
type _git_golaxy_org_framework_addins_gate_cli_IWatcher struct {
	IValue      interface{}
	WDeadline   func() (deadline time.Time, ok bool)
	WDone       func() <-chan struct{}
	WErr        func() error
	WTerminate  func() async.AsyncRet
	WTerminated func() async.AsyncRet
	WValue      func(key any) any
}

func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Deadline() (deadline time.Time, ok bool) {
	return W.WDeadline()
}
func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Done() <-chan struct{} { return W.WDone() }
func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Err() error            { return W.WErr() }
func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Terminate() async.AsyncRet {
	return W.WTerminate()
}
func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Terminated() async.AsyncRet {
	return W.WTerminated()
}
func (W _git_golaxy_org_framework_addins_gate_cli_IWatcher) Value(key any) any { return W.WValue(key) }
