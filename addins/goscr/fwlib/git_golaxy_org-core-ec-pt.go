// Code generated by 'yaegi extract git.golaxy.org/core/ec/pt'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/utils/generic"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/ec/pt/pt"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"DefaultComponentLib":   reflect.ValueOf(pt.DefaultComponentLib),
		"ErrPt":                 reflect.ValueOf(&pt.ErrPt).Elem(),
		"For":                   reflect.ValueOf(pt.For),
		"Inject":                reflect.ValueOf(pt.Inject),
		"InjectRV":              reflect.ValueOf(pt.InjectRV),
		"NewComponentAttribute": reflect.ValueOf(pt.NewComponentAttribute),
		"NewComponentLib":       reflect.ValueOf(pt.NewComponentLib),
		"NewEntityAttribute":    reflect.ValueOf(pt.NewEntityAttribute),
		"NewEntityLib":          reflect.ValueOf(pt.NewEntityLib),
		"UnsafeEntityLib":       reflect.ValueOf(pt.UnsafeEntityLib),

		// type definitions
		"ComponentAttribute": reflect.ValueOf((*pt.ComponentAttribute)(nil)),
		"ComponentLib":       reflect.ValueOf((*pt.ComponentLib)(nil)),
		"EntityAttribute":    reflect.ValueOf((*pt.EntityAttribute)(nil)),
		"EntityLib":          reflect.ValueOf((*pt.EntityLib)(nil)),
		"EntityPTProvider":   reflect.ValueOf((*pt.EntityPTProvider)(nil)),

		// interface wrapper definitions
		"_ComponentLib":     reflect.ValueOf((*_git_golaxy_org_core_ec_pt_ComponentLib)(nil)),
		"_EntityLib":        reflect.ValueOf((*_git_golaxy_org_core_ec_pt_EntityLib)(nil)),
		"_EntityPTProvider": reflect.ValueOf((*_git_golaxy_org_core_ec_pt_EntityPTProvider)(nil)),
	}
}

// _git_golaxy_org_core_ec_pt_ComponentLib is an interface wrapper for ComponentLib type
type _git_golaxy_org_core_ec_pt_ComponentLib struct {
	IValue         interface{}
	WDeclare       func(comp any) ec.ComponentPT
	WGet           func(prototype string) (ec.ComponentPT, bool)
	WRange         func(fun generic.Func1[ec.ComponentPT, bool])
	WReversedRange func(fun generic.Func1[ec.ComponentPT, bool])
	WUndeclare     func(prototype string)
}

func (W _git_golaxy_org_core_ec_pt_ComponentLib) Declare(comp any) ec.ComponentPT {
	return W.WDeclare(comp)
}
func (W _git_golaxy_org_core_ec_pt_ComponentLib) Get(prototype string) (ec.ComponentPT, bool) {
	return W.WGet(prototype)
}
func (W _git_golaxy_org_core_ec_pt_ComponentLib) Range(fun generic.Func1[ec.ComponentPT, bool]) {
	W.WRange(fun)
}
func (W _git_golaxy_org_core_ec_pt_ComponentLib) ReversedRange(fun generic.Func1[ec.ComponentPT, bool]) {
	W.WReversedRange(fun)
}
func (W _git_golaxy_org_core_ec_pt_ComponentLib) Undeclare(prototype string) { W.WUndeclare(prototype) }

// _git_golaxy_org_core_ec_pt_EntityLib is an interface wrapper for EntityLib type
type _git_golaxy_org_core_ec_pt_EntityLib struct {
	IValue           interface{}
	WDeclare         func(prototype any, comps ...any) ec.EntityPT
	WGet             func(prototype string) (ec.EntityPT, bool)
	WGetComponentLib func() pt.ComponentLib
	WGetEntityLib    func() pt.EntityLib
	WRange           func(fun generic.Func1[ec.EntityPT, bool])
	WRedeclare       func(prototype any, comps ...any) ec.EntityPT
	WReversedRange   func(fun generic.Func1[ec.EntityPT, bool])
	WUndeclare       func(prototype string)
}

func (W _git_golaxy_org_core_ec_pt_EntityLib) Declare(prototype any, comps ...any) ec.EntityPT {
	return W.WDeclare(prototype, comps...)
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) Get(prototype string) (ec.EntityPT, bool) {
	return W.WGet(prototype)
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) GetComponentLib() pt.ComponentLib {
	return W.WGetComponentLib()
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) GetEntityLib() pt.EntityLib { return W.WGetEntityLib() }
func (W _git_golaxy_org_core_ec_pt_EntityLib) Range(fun generic.Func1[ec.EntityPT, bool]) {
	W.WRange(fun)
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) Redeclare(prototype any, comps ...any) ec.EntityPT {
	return W.WRedeclare(prototype, comps...)
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) ReversedRange(fun generic.Func1[ec.EntityPT, bool]) {
	W.WReversedRange(fun)
}
func (W _git_golaxy_org_core_ec_pt_EntityLib) Undeclare(prototype string) { W.WUndeclare(prototype) }

// _git_golaxy_org_core_ec_pt_EntityPTProvider is an interface wrapper for EntityPTProvider type
type _git_golaxy_org_core_ec_pt_EntityPTProvider struct {
	IValue        interface{}
	WGetEntityLib func() pt.EntityLib
}

func (W _git_golaxy_org_core_ec_pt_EntityPTProvider) GetEntityLib() pt.EntityLib {
	return W.WGetEntityLib()
}
