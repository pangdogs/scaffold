// Code generated by 'yaegi extract git.golaxy.org/scaffold/addins/view'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/scaffold/addins/view"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/scaffold/addins/view/view"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"DeclareProp":                     reflect.ValueOf(view.DeclareProp),
		"ErrDiscontinuousRevision":        reflect.ValueOf(&view.ErrDiscontinuousRevision).Elem(),
		"ErrEntityNoProp":                 reflect.ValueOf(&view.ErrEntityNoProp).Elem(),
		"ErrEntityNoPropTab":              reflect.ValueOf(&view.ErrEntityNoPropTab).Elem(),
		"ErrEntityNotFound":               reflect.ValueOf(&view.ErrEntityNotFound).Elem(),
		"ErrMethodNotFound":               reflect.ValueOf(&view.ErrMethodNotFound).Elem(),
		"ErrMethodParameterCountMismatch": reflect.ValueOf(&view.ErrMethodParameterCountMismatch).Elem(),
		"ErrMethodParameterTypeMismatch":  reflect.ValueOf(&view.ErrMethodParameterTypeMismatch).Elem(),
		"ErrOutdatedRevision":             reflect.ValueOf(&view.ErrOutdatedRevision).Elem(),
		"Install":                         reflect.ValueOf(&view.Install).Elem(),
		"Name":                            reflect.ValueOf(&view.Name).Elem(),
		"ReferenceProp":                   reflect.ValueOf(view.ReferenceProp),
		"Uninstall":                       reflect.ValueOf(&view.Uninstall).Elem(),
		"UnsafeProp":                      reflect.ValueOf(view.UnsafeProp),
		"UnsafePropSync":                  reflect.ValueOf(view.UnsafePropSync),
		"Using":                           reflect.ValueOf(&view.Using).Elem(),

		// type definitions
		"IProp":       reflect.ValueOf((*view.IProp)(nil)),
		"IPropSync":   reflect.ValueOf((*view.IPropSync)(nil)),
		"IPropTab":    reflect.ValueOf((*view.IPropTab)(nil)),
		"IPropView":   reflect.ValueOf((*view.IPropView)(nil)),
		"PropCreator": reflect.ValueOf((*view.PropCreator)(nil)),
		"PropSync":    reflect.ValueOf((*view.PropSync)(nil)),
		"PropTab":     reflect.ValueOf((*view.PropTab)(nil)),

		// interface wrapper definitions
		"_IProp":     reflect.ValueOf((*_git_golaxy_org_scaffold_addins_view_IProp)(nil)),
		"_IPropSync": reflect.ValueOf((*_git_golaxy_org_scaffold_addins_view_IPropSync)(nil)),
		"_IPropTab":  reflect.ValueOf((*_git_golaxy_org_scaffold_addins_view_IPropTab)(nil)),
		"_IPropView": reflect.ValueOf((*_git_golaxy_org_scaffold_addins_view_IPropView)(nil)),
	}
}

// _git_golaxy_org_scaffold_addins_view_IProp is an interface wrapper for IProp type
type _git_golaxy_org_scaffold_addins_view_IProp struct {
	IValue     interface{}
	WMarshal   func() ([]byte, int64, error)
	WReset     func()
	WRevision  func() int64
	WUnmarshal func(data []byte, revision int64) error
}

func (W _git_golaxy_org_scaffold_addins_view_IProp) Marshal() ([]byte, int64, error) {
	return W.WMarshal()
}
func (W _git_golaxy_org_scaffold_addins_view_IProp) Reset()          { W.WReset() }
func (W _git_golaxy_org_scaffold_addins_view_IProp) Revision() int64 { return W.WRevision() }
func (W _git_golaxy_org_scaffold_addins_view_IProp) Unmarshal(data []byte, revision int64) error {
	return W.WUnmarshal(data, revision)
}

// _git_golaxy_org_scaffold_addins_view_IPropSync is an interface wrapper for IPropSync type
type _git_golaxy_org_scaffold_addins_view_IPropSync struct {
	IValue     interface{}
	WAtti      func() any
	WLoad      func(service string) error
	WManaged   func() view.IProp
	WMarshal   func() ([]byte, int64, error)
	WReflected func() reflect.Value
	WReset     func()
	WRevision  func() int64
	WSave      func(service string) error
	WUnmarshal func(data []byte, revision int64) error
}

func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Atti() any { return W.WAtti() }
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Load(service string) error {
	return W.WLoad(service)
}
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Managed() view.IProp { return W.WManaged() }
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Marshal() ([]byte, int64, error) {
	return W.WMarshal()
}
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Reflected() reflect.Value {
	return W.WReflected()
}
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Reset()          { W.WReset() }
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Revision() int64 { return W.WRevision() }
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Save(service string) error {
	return W.WSave(service)
}
func (W _git_golaxy_org_scaffold_addins_view_IPropSync) Unmarshal(data []byte, revision int64) error {
	return W.WUnmarshal(data, revision)
}

// _git_golaxy_org_scaffold_addins_view_IPropTab is an interface wrapper for IPropTab type
type _git_golaxy_org_scaffold_addins_view_IPropTab struct {
	IValue      interface{}
	WAddProp    func(name string, ps view.IPropSync)
	WEachProps  func(fun generic.Action2[string, view.IPropSync])
	WGetProp    func(name string) view.IPropSync
	WRangeProps func(fun generic.Func2[string, view.IPropSync, bool])
}

func (W _git_golaxy_org_scaffold_addins_view_IPropTab) AddProp(name string, ps view.IPropSync) {
	W.WAddProp(name, ps)
}
func (W _git_golaxy_org_scaffold_addins_view_IPropTab) EachProps(fun generic.Action2[string, view.IPropSync]) {
	W.WEachProps(fun)
}
func (W _git_golaxy_org_scaffold_addins_view_IPropTab) GetProp(name string) view.IPropSync {
	return W.WGetProp(name)
}
func (W _git_golaxy_org_scaffold_addins_view_IPropTab) RangeProps(fun generic.Func2[string, view.IPropSync, bool]) {
	W.WRangeProps(fun)
}

// _git_golaxy_org_scaffold_addins_view_IPropView is an interface wrapper for IPropView type
type _git_golaxy_org_scaffold_addins_view_IPropView struct {
	IValue interface{}
}
