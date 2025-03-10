// Code generated by 'yaegi extract git.golaxy.org/core/extension'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/extension/extension"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"AddInState_Active":   reflect.ValueOf(extension.AddInState_Active),
		"AddInState_Inactive": reflect.ValueOf(extension.AddInState_Inactive),
		"AddInState_Loaded":   reflect.ValueOf(extension.AddInState_Loaded),
		"ErrExtension":        reflect.ValueOf(&extension.ErrExtension).Elem(),
		"NewAddInManager":     reflect.ValueOf(extension.NewAddInManager),
		"Uninstall":           reflect.ValueOf(extension.Uninstall),
		"UnsafeAddInManager":  reflect.ValueOf(extension.UnsafeAddInManager),
		"UnsafeAddInStatus":   reflect.ValueOf(extension.UnsafeAddInStatus),

		// type definitions
		"AddInManager":  reflect.ValueOf((*extension.AddInManager)(nil)),
		"AddInProvider": reflect.ValueOf((*extension.AddInProvider)(nil)),
		"AddInState":    reflect.ValueOf((*extension.AddInState)(nil)),
		"AddInStatus":   reflect.ValueOf((*extension.AddInStatus)(nil)),

		// interface wrapper definitions
		"_AddInManager":  reflect.ValueOf((*_git_golaxy_org_core_extension_AddInManager)(nil)),
		"_AddInProvider": reflect.ValueOf((*_git_golaxy_org_core_extension_AddInProvider)(nil)),
		"_AddInStatus":   reflect.ValueOf((*_git_golaxy_org_core_extension_AddInStatus)(nil)),
	}
}

// _git_golaxy_org_core_extension_AddInManager is an interface wrapper for AddInManager type
type _git_golaxy_org_core_extension_AddInManager struct {
	IValue           interface{}
	WGet             func(name string) (extension.AddInStatus, bool)
	WGetAddInManager func() extension.AddInManager
	WInstall         func(addInFace iface.FaceAny, name ...string)
	WRange           func(fun generic.Func1[extension.AddInStatus, bool])
	WReversedRange   func(fun generic.Func1[extension.AddInStatus, bool])
	WUninstall       func(name string)
}

func (W _git_golaxy_org_core_extension_AddInManager) Get(name string) (extension.AddInStatus, bool) {
	return W.WGet(name)
}
func (W _git_golaxy_org_core_extension_AddInManager) GetAddInManager() extension.AddInManager {
	return W.WGetAddInManager()
}
func (W _git_golaxy_org_core_extension_AddInManager) Install(addInFace iface.FaceAny, name ...string) {
	W.WInstall(addInFace, name...)
}
func (W _git_golaxy_org_core_extension_AddInManager) Range(fun generic.Func1[extension.AddInStatus, bool]) {
	W.WRange(fun)
}
func (W _git_golaxy_org_core_extension_AddInManager) ReversedRange(fun generic.Func1[extension.AddInStatus, bool]) {
	W.WReversedRange(fun)
}
func (W _git_golaxy_org_core_extension_AddInManager) Uninstall(name string) { W.WUninstall(name) }

// _git_golaxy_org_core_extension_AddInProvider is an interface wrapper for AddInProvider type
type _git_golaxy_org_core_extension_AddInProvider struct {
	IValue           interface{}
	WGetAddInManager func() extension.AddInManager
}

func (W _git_golaxy_org_core_extension_AddInProvider) GetAddInManager() extension.AddInManager {
	return W.WGetAddInManager()
}

// _git_golaxy_org_core_extension_AddInStatus is an interface wrapper for AddInStatus type
type _git_golaxy_org_core_extension_AddInStatus struct {
	IValue        interface{}
	WInstanceFace func() iface.FaceAny
	WName         func() string
	WReflected    func() reflect.Value
	WState        func() extension.AddInState
}

func (W _git_golaxy_org_core_extension_AddInStatus) InstanceFace() iface.FaceAny {
	return W.WInstanceFace()
}
func (W _git_golaxy_org_core_extension_AddInStatus) Name() string                { return W.WName() }
func (W _git_golaxy_org_core_extension_AddInStatus) Reflected() reflect.Value    { return W.WReflected() }
func (W _git_golaxy_org_core_extension_AddInStatus) State() extension.AddInState { return W.WState() }
