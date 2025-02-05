// Code generated by 'yaegi extract git.golaxy.org/scaffold/addins/goscr'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/scaffold/addins/goscr"
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/scaffold/addins/goscr/goscr"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"BuildEntityPT":      reflect.ValueOf(goscr.BuildEntityPT),
		"ComponentScript":    reflect.ValueOf(goscr.ComponentScript),
		"EntityScript":       reflect.ValueOf(goscr.EntityScript),
		"GetComponentScript": reflect.ValueOf(goscr.GetComponentScript),
		"GetEntityScript":    reflect.ValueOf(goscr.GetEntityScript),
		"Install":            reflect.ValueOf(&goscr.Install).Elem(),
		"Name":               reflect.ValueOf(&goscr.Name).Elem(),
		"Uninstall":          reflect.ValueOf(&goscr.Uninstall).Elem(),
		"Using":              reflect.ValueOf(&goscr.Using).Elem(),
		"With":               reflect.ValueOf(&goscr.With).Elem(),

		// type definitions
		"Component":                          reflect.ValueOf((*goscr.Component)(nil)),
		"ComponentBehavior":                  reflect.ValueOf((*goscr.ComponentBehavior)(nil)),
		"ComponentEnableLateUpdate":          reflect.ValueOf((*goscr.ComponentEnableLateUpdate)(nil)),
		"ComponentEnableUpdate":              reflect.ValueOf((*goscr.ComponentEnableUpdate)(nil)),
		"ComponentEnableUpdateAndLateUpdate": reflect.ValueOf((*goscr.ComponentEnableUpdateAndLateUpdate)(nil)),
		"Entity":                             reflect.ValueOf((*goscr.Entity)(nil)),
		"EntityBehavior":                     reflect.ValueOf((*goscr.EntityBehavior)(nil)),
		"EntityEnableLateUpdate":             reflect.ValueOf((*goscr.EntityEnableLateUpdate)(nil)),
		"EntityEnableUpdate":                 reflect.ValueOf((*goscr.EntityEnableUpdate)(nil)),
		"EntityEnableUpdateAndLateUpdate":    reflect.ValueOf((*goscr.EntityEnableUpdateAndLateUpdate)(nil)),
		"EntityPTCreator":                    reflect.ValueOf((*goscr.EntityPTCreator)(nil)),
		"IScript":                            reflect.ValueOf((*goscr.IScript)(nil)),
		"LifecycleComponentOnCreate":         reflect.ValueOf((*goscr.LifecycleComponentOnCreate)(nil)),
		"LifecycleComponentOnDisposed":       reflect.ValueOf((*goscr.LifecycleComponentOnDisposed)(nil)),
		"LifecycleComponentOnStarted":        reflect.ValueOf((*goscr.LifecycleComponentOnStarted)(nil)),
		"LifecycleComponentOnStop":           reflect.ValueOf((*goscr.LifecycleComponentOnStop)(nil)),
		"LifecycleEntityOnCreate":            reflect.ValueOf((*goscr.LifecycleEntityOnCreate)(nil)),
		"LifecycleEntityOnDisposed":          reflect.ValueOf((*goscr.LifecycleEntityOnDisposed)(nil)),
		"LifecycleEntityOnStarted":           reflect.ValueOf((*goscr.LifecycleEntityOnStarted)(nil)),
		"LifecycleEntityOnStop":              reflect.ValueOf((*goscr.LifecycleEntityOnStop)(nil)),
		"LoadedCB":                           reflect.ValueOf((*goscr.LoadedCB)(nil)),
		"LoadingCB":                          reflect.ValueOf((*goscr.LoadingCB)(nil)),
		"ScriptOptions":                      reflect.ValueOf((*goscr.ScriptOptions)(nil)),

		// interface wrapper definitions
		"_IScript":                      reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_IScript)(nil)),
		"_LifecycleComponentOnCreate":   reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnCreate)(nil)),
		"_LifecycleComponentOnDisposed": reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnDisposed)(nil)),
		"_LifecycleComponentOnStarted":  reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStarted)(nil)),
		"_LifecycleComponentOnStop":     reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStop)(nil)),
		"_LifecycleEntityOnCreate":      reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnCreate)(nil)),
		"_LifecycleEntityOnDisposed":    reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnDisposed)(nil)),
		"_LifecycleEntityOnStarted":     reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStarted)(nil)),
		"_LifecycleEntityOnStop":        reflect.ValueOf((*_git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStop)(nil)),
	}
}

// _git_golaxy_org_scaffold_addins_goscr_IScript is an interface wrapper for IScript type
type _git_golaxy_org_scaffold_addins_goscr_IScript struct {
	IValue    interface{}
	WHotfix   func() error
	WSolution func() *dynamic.Solution
}

func (W _git_golaxy_org_scaffold_addins_goscr_IScript) Hotfix() error { return W.WHotfix() }
func (W _git_golaxy_org_scaffold_addins_goscr_IScript) Solution() *dynamic.Solution {
	return W.WSolution()
}

// _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnCreate is an interface wrapper for LifecycleComponentOnCreate type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnCreate struct {
	IValue    interface{}
	WOnCreate func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnCreate) OnCreate() { W.WOnCreate() }

// _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnDisposed is an interface wrapper for LifecycleComponentOnDisposed type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnDisposed struct {
	IValue      interface{}
	WOnDisposed func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnDisposed) OnDisposed() {
	W.WOnDisposed()
}

// _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStarted is an interface wrapper for LifecycleComponentOnStarted type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStarted struct {
	IValue     interface{}
	WOnStarted func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStarted) OnStarted() {
	W.WOnStarted()
}

// _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStop is an interface wrapper for LifecycleComponentOnStop type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStop struct {
	IValue  interface{}
	WOnStop func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleComponentOnStop) OnStop() { W.WOnStop() }

// _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnCreate is an interface wrapper for LifecycleEntityOnCreate type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnCreate struct {
	IValue    interface{}
	WOnCreate func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnCreate) OnCreate() { W.WOnCreate() }

// _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnDisposed is an interface wrapper for LifecycleEntityOnDisposed type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnDisposed struct {
	IValue      interface{}
	WOnDisposed func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnDisposed) OnDisposed() {
	W.WOnDisposed()
}

// _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStarted is an interface wrapper for LifecycleEntityOnStarted type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStarted struct {
	IValue     interface{}
	WOnStarted func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStarted) OnStarted() { W.WOnStarted() }

// _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStop is an interface wrapper for LifecycleEntityOnStop type
type _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStop struct {
	IValue  interface{}
	WOnStop func()
}

func (W _git_golaxy_org_scaffold_addins_goscr_LifecycleEntityOnStop) OnStop() { W.WOnStop() }
