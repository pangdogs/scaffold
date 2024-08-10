package view

import (
	"git.golaxy.org/core/ec"
	"reflect"
)

// IPropSync 属性同步接口
type IPropSync interface {
	IProp
	iPropSync
	// Load 加载
	Load(service string) error
	// Save 保存
	Save(service string) error
	// Managed 托管的属性
	Managed() IProp
	// Mem Mem
	Mem() any
	// Reflected 反射值
	Reflected() reflect.Value
}

type iPropSync interface {
	init(view IPropView, entity ec.Entity, name string, reflected reflect.Value)
	setSyncTo(syncTo []string)
	setMem(m any)
	sync(revision int64, op string, args ...any)
}

// PropSync 属性同步器
type PropSync struct {
	view      IPropView
	entity    ec.Entity
	name      string
	mem       any
	reflected reflect.Value
	syncTo    []string
}

func (ps *PropSync) init(view IPropView, entity ec.Entity, name string, reflected reflect.Value) {
	ps.view = view
	ps.entity = entity
	ps.name = name
	ps.reflected = reflected
}

func (ps *PropSync) setSyncTo(syncTo []string) {
	ps.syncTo = syncTo
}

func (ps *PropSync) setMem(m any) {
	ps.mem = m
}

// Load 加载数据
func (ps *PropSync) Load(service string) ([]byte, int64, error) {
	return ps.view.load(ps, service)
}

// Save 保存数据
func (ps *PropSync) Save(service string, data []byte, revision int64) error {
	return ps.view.save(ps, service, data, revision)
}

// Mem Mem
func (ps *PropSync) Mem() any {
	return ps.mem
}

// Reflected 反射值
func (ps *PropSync) Reflected() reflect.Value {
	return ps.reflected
}

func (ps *PropSync) sync(revision int64, op string, args ...any) {
	ps.view.sync(ps, revision, op, args...)
}