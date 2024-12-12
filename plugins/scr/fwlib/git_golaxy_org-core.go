// Code generated by 'yaegi extract git.golaxy.org/core'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/core"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Async":                 reflect.ValueOf(core.Async),
		"AsyncVoid":             reflect.ValueOf(core.AsyncVoid),
		"Await":                 reflect.ValueOf(core.Await),
		"CreateEntity":          reflect.ValueOf(core.CreateEntity),
		"CreateEntityPT":        reflect.ValueOf(core.CreateEntityPT),
		"ErrAllFailures":        reflect.ValueOf(&core.ErrAllFailures).Elem(),
		"ErrArgs":               reflect.ValueOf(&core.ErrArgs).Elem(),
		"ErrCore":               reflect.ValueOf(&core.ErrCore).Elem(),
		"ErrPanicked":           reflect.ValueOf(&core.ErrPanicked).Elem(),
		"ErrProcessQueueClosed": reflect.ValueOf(&core.ErrProcessQueueClosed).Elem(),
		"ErrProcessQueueFull":   reflect.ValueOf(&core.ErrProcessQueueFull).Elem(),
		"ErrRuntime":            reflect.ValueOf(&core.ErrRuntime).Elem(),
		"ErrService":            reflect.ValueOf(&core.ErrService).Elem(),
		"Go":                    reflect.ValueOf(core.Go),
		"GoVoid":                reflect.ValueOf(core.GoVoid),
		"NewRuntime":            reflect.ValueOf(core.NewRuntime),
		"NewService":            reflect.ValueOf(core.NewService),
		"TimeAfter":             reflect.ValueOf(core.TimeAfter),
		"TimeAt":                reflect.ValueOf(core.TimeAt),
		"TimeTick":              reflect.ValueOf(core.TimeTick),
		"UnsafeNewRuntime":      reflect.ValueOf(core.UnsafeNewRuntime),
		"UnsafeNewService":      reflect.ValueOf(core.UnsafeNewService),
		"UnsafeRuntime":         reflect.ValueOf(core.UnsafeRuntime),
		"UnsafeService":         reflect.ValueOf(core.UnsafeService),
		"With":                  reflect.ValueOf(&core.With).Elem(),

		// type definitions
		"AwaitDirector":                               reflect.ValueOf((*core.AwaitDirector)(nil)),
		"EntityCreator":                               reflect.ValueOf((*core.EntityCreator)(nil)),
		"EntityPTCreator":                             reflect.ValueOf((*core.EntityPTCreator)(nil)),
		"LifecycleComponentAwake":                     reflect.ValueOf((*core.LifecycleComponentAwake)(nil)),
		"LifecycleComponentDispose":                   reflect.ValueOf((*core.LifecycleComponentDispose)(nil)),
		"LifecycleComponentLateUpdate":                reflect.ValueOf((*core.LifecycleComponentLateUpdate)(nil)),
		"LifecycleComponentShut":                      reflect.ValueOf((*core.LifecycleComponentShut)(nil)),
		"LifecycleComponentStart":                     reflect.ValueOf((*core.LifecycleComponentStart)(nil)),
		"LifecycleComponentUpdate":                    reflect.ValueOf((*core.LifecycleComponentUpdate)(nil)),
		"LifecycleEntityAwake":                        reflect.ValueOf((*core.LifecycleEntityAwake)(nil)),
		"LifecycleEntityDispose":                      reflect.ValueOf((*core.LifecycleEntityDispose)(nil)),
		"LifecycleEntityLateUpdate":                   reflect.ValueOf((*core.LifecycleEntityLateUpdate)(nil)),
		"LifecycleEntityShut":                         reflect.ValueOf((*core.LifecycleEntityShut)(nil)),
		"LifecycleEntityStart":                        reflect.ValueOf((*core.LifecycleEntityStart)(nil)),
		"LifecycleEntityUpdate":                       reflect.ValueOf((*core.LifecycleEntityUpdate)(nil)),
		"LifecyclePluginInit":                         reflect.ValueOf((*core.LifecyclePluginInit)(nil)),
		"LifecyclePluginOnRuntimeRunningStateChanged": reflect.ValueOf((*core.LifecyclePluginOnRuntimeRunningStateChanged)(nil)),
		"LifecyclePluginShut":                         reflect.ValueOf((*core.LifecyclePluginShut)(nil)),
		"Runtime":                                     reflect.ValueOf((*core.Runtime)(nil)),
		"RuntimeBehavior":                             reflect.ValueOf((*core.RuntimeBehavior)(nil)),
		"RuntimeOptions":                              reflect.ValueOf((*core.RuntimeOptions)(nil)),
		"Service":                                     reflect.ValueOf((*core.Service)(nil)),
		"ServiceBehavior":                             reflect.ValueOf((*core.ServiceBehavior)(nil)),
		"ServiceOptions":                              reflect.ValueOf((*core.ServiceOptions)(nil)),

		// interface wrapper definitions
		"_LifecycleComponentAwake":                     reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentAwake)(nil)),
		"_LifecycleComponentDispose":                   reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentDispose)(nil)),
		"_LifecycleComponentLateUpdate":                reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentLateUpdate)(nil)),
		"_LifecycleComponentShut":                      reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentShut)(nil)),
		"_LifecycleComponentStart":                     reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentStart)(nil)),
		"_LifecycleComponentUpdate":                    reflect.ValueOf((*_git_golaxy_org_core_LifecycleComponentUpdate)(nil)),
		"_LifecycleEntityAwake":                        reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityAwake)(nil)),
		"_LifecycleEntityDispose":                      reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityDispose)(nil)),
		"_LifecycleEntityLateUpdate":                   reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityLateUpdate)(nil)),
		"_LifecycleEntityShut":                         reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityShut)(nil)),
		"_LifecycleEntityStart":                        reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityStart)(nil)),
		"_LifecycleEntityUpdate":                       reflect.ValueOf((*_git_golaxy_org_core_LifecycleEntityUpdate)(nil)),
		"_LifecyclePluginInit":                         reflect.ValueOf((*_git_golaxy_org_core_LifecyclePluginInit)(nil)),
		"_LifecyclePluginOnRuntimeRunningStateChanged": reflect.ValueOf((*_git_golaxy_org_core_LifecyclePluginOnRuntimeRunningStateChanged)(nil)),
		"_LifecyclePluginShut":                         reflect.ValueOf((*_git_golaxy_org_core_LifecyclePluginShut)(nil)),
		"_Runtime":                                     reflect.ValueOf((*_git_golaxy_org_core_Runtime)(nil)),
		"_Service":                                     reflect.ValueOf((*_git_golaxy_org_core_Service)(nil)),
	}
}

// _git_golaxy_org_core_LifecycleComponentAwake is an interface wrapper for LifecycleComponentAwake type
type _git_golaxy_org_core_LifecycleComponentAwake struct {
	IValue interface{}
	WAwake func()
}

func (W _git_golaxy_org_core_LifecycleComponentAwake) Awake() {
	W.WAwake()
}

// _git_golaxy_org_core_LifecycleComponentDispose is an interface wrapper for LifecycleComponentDispose type
type _git_golaxy_org_core_LifecycleComponentDispose struct {
	IValue   interface{}
	WDispose func()
}

func (W _git_golaxy_org_core_LifecycleComponentDispose) Dispose() {
	W.WDispose()
}

// _git_golaxy_org_core_LifecycleComponentLateUpdate is an interface wrapper for LifecycleComponentLateUpdate type
type _git_golaxy_org_core_LifecycleComponentLateUpdate struct {
	IValue      interface{}
	WLateUpdate func()
}

func (W _git_golaxy_org_core_LifecycleComponentLateUpdate) LateUpdate() {
	W.WLateUpdate()
}

// _git_golaxy_org_core_LifecycleComponentShut is an interface wrapper for LifecycleComponentShut type
type _git_golaxy_org_core_LifecycleComponentShut struct {
	IValue interface{}
	WShut  func()
}

func (W _git_golaxy_org_core_LifecycleComponentShut) Shut() {
	W.WShut()
}

// _git_golaxy_org_core_LifecycleComponentStart is an interface wrapper for LifecycleComponentStart type
type _git_golaxy_org_core_LifecycleComponentStart struct {
	IValue interface{}
	WStart func()
}

func (W _git_golaxy_org_core_LifecycleComponentStart) Start() {
	W.WStart()
}

// _git_golaxy_org_core_LifecycleComponentUpdate is an interface wrapper for LifecycleComponentUpdate type
type _git_golaxy_org_core_LifecycleComponentUpdate struct {
	IValue  interface{}
	WUpdate func()
}

func (W _git_golaxy_org_core_LifecycleComponentUpdate) Update() {
	W.WUpdate()
}

// _git_golaxy_org_core_LifecycleEntityAwake is an interface wrapper for LifecycleEntityAwake type
type _git_golaxy_org_core_LifecycleEntityAwake struct {
	IValue interface{}
	WAwake func()
}

func (W _git_golaxy_org_core_LifecycleEntityAwake) Awake() {
	W.WAwake()
}

// _git_golaxy_org_core_LifecycleEntityDispose is an interface wrapper for LifecycleEntityDispose type
type _git_golaxy_org_core_LifecycleEntityDispose struct {
	IValue   interface{}
	WDispose func()
}

func (W _git_golaxy_org_core_LifecycleEntityDispose) Dispose() {
	W.WDispose()
}

// _git_golaxy_org_core_LifecycleEntityLateUpdate is an interface wrapper for LifecycleEntityLateUpdate type
type _git_golaxy_org_core_LifecycleEntityLateUpdate struct {
	IValue      interface{}
	WLateUpdate func()
}

func (W _git_golaxy_org_core_LifecycleEntityLateUpdate) LateUpdate() {
	W.WLateUpdate()
}

// _git_golaxy_org_core_LifecycleEntityShut is an interface wrapper for LifecycleEntityShut type
type _git_golaxy_org_core_LifecycleEntityShut struct {
	IValue interface{}
	WShut  func()
}

func (W _git_golaxy_org_core_LifecycleEntityShut) Shut() {
	W.WShut()
}

// _git_golaxy_org_core_LifecycleEntityStart is an interface wrapper for LifecycleEntityStart type
type _git_golaxy_org_core_LifecycleEntityStart struct {
	IValue interface{}
	WStart func()
}

func (W _git_golaxy_org_core_LifecycleEntityStart) Start() {
	W.WStart()
}

// _git_golaxy_org_core_LifecycleEntityUpdate is an interface wrapper for LifecycleEntityUpdate type
type _git_golaxy_org_core_LifecycleEntityUpdate struct {
	IValue  interface{}
	WUpdate func()
}

func (W _git_golaxy_org_core_LifecycleEntityUpdate) Update() {
	W.WUpdate()
}

// _git_golaxy_org_core_LifecyclePluginInit is an interface wrapper for LifecyclePluginInit type
type _git_golaxy_org_core_LifecyclePluginInit struct {
	IValue interface{}
	WInit  func(svcCtx service.Context, rtCtx runtime.Context)
}

func (W _git_golaxy_org_core_LifecyclePluginInit) Init(svcCtx service.Context, rtCtx runtime.Context) {
	W.WInit(svcCtx, rtCtx)
}

// _git_golaxy_org_core_LifecyclePluginOnRuntimeRunningStateChanged is an interface wrapper for LifecyclePluginOnRuntimeRunningStateChanged type
type _git_golaxy_org_core_LifecyclePluginOnRuntimeRunningStateChanged struct {
	IValue                        interface{}
	WOnRuntimeRunningStateChanged func(rtCtx runtime.Context, state runtime.RunningState, args ...any)
}

func (W _git_golaxy_org_core_LifecyclePluginOnRuntimeRunningStateChanged) OnRuntimeRunningStateChanged(rtCtx runtime.Context, state runtime.RunningState, args ...any) {
	W.WOnRuntimeRunningStateChanged(rtCtx, state, args...)
}

// _git_golaxy_org_core_LifecyclePluginShut is an interface wrapper for LifecyclePluginShut type
type _git_golaxy_org_core_LifecyclePluginShut struct {
	IValue interface{}
	WShut  func(svcCtx service.Context, rtCtx runtime.Context)
}

func (W _git_golaxy_org_core_LifecyclePluginShut) Shut(svcCtx service.Context, rtCtx runtime.Context) {
	W.WShut(svcCtx, rtCtx)
}

// _git_golaxy_org_core_Runtime is an interface wrapper for Runtime type
type _git_golaxy_org_core_Runtime struct {
	IValue                interface{}
	WGetConcurrentContext func() iface.Cache
	WGetCurrentContext    func() iface.Cache
	WGetInstanceFaceCache func() iface.Cache
	WPushCall             func(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WPushCallDelegate     func(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WPushCallVoid         func(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any]
	WPushCallVoidDelegate func(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any]
	WRun                  func() <-chan struct{}
	WTerminate            func() <-chan struct{}
	WTerminated           func() <-chan struct{}
}

func (W _git_golaxy_org_core_Runtime) GetConcurrentContext() iface.Cache {
	return W.WGetConcurrentContext()
}
func (W _git_golaxy_org_core_Runtime) GetCurrentContext() iface.Cache {
	return W.WGetCurrentContext()
}
func (W _git_golaxy_org_core_Runtime) GetInstanceFaceCache() iface.Cache {
	return W.WGetInstanceFaceCache()
}
func (W _git_golaxy_org_core_Runtime) PushCall(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WPushCall(fun, args...)
}
func (W _git_golaxy_org_core_Runtime) PushCallDelegate(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WPushCallDelegate(fun, args...)
}
func (W _git_golaxy_org_core_Runtime) PushCallVoid(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WPushCallVoid(fun, args...)
}
func (W _git_golaxy_org_core_Runtime) PushCallVoidDelegate(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WPushCallVoidDelegate(fun, args...)
}
func (W _git_golaxy_org_core_Runtime) Run() <-chan struct{} {
	return W.WRun()
}
func (W _git_golaxy_org_core_Runtime) Terminate() <-chan struct{} {
	return W.WTerminate()
}
func (W _git_golaxy_org_core_Runtime) Terminated() <-chan struct{} {
	return W.WTerminated()
}

// _git_golaxy_org_core_Service is an interface wrapper for Service type
type _git_golaxy_org_core_Service struct {
	IValue                interface{}
	WGetContext           func() service.Context
	WGetInstanceFaceCache func() iface.Cache
	WRun                  func() <-chan struct{}
	WTerminate            func() <-chan struct{}
	WTerminated           func() <-chan struct{}
}

func (W _git_golaxy_org_core_Service) GetContext() service.Context {
	return W.WGetContext()
}
func (W _git_golaxy_org_core_Service) GetInstanceFaceCache() iface.Cache {
	return W.WGetInstanceFaceCache()
}
func (W _git_golaxy_org_core_Service) Run() <-chan struct{} {
	return W.WRun()
}
func (W _git_golaxy_org_core_Service) Terminate() <-chan struct{} {
	return W.WTerminate()
}
func (W _git_golaxy_org_core_Service) Terminated() <-chan struct{} {
	return W.WTerminated()
}