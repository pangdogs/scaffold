/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package goscr

import (
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/meta"
	"git.golaxy.org/framework"
	"github.com/elliotchance/pie/v2"
)

// BuildEntityPT 创建实体原型
func BuildEntityPT(svcInst framework.IService, prototype string) *EntityPTCreator {
	if svcInst == nil {
		exception.Panicf("goscr: %w: svcInst is nil", core.ErrArgs)
	}
	return &EntityPTCreator{
		svcInst: svcInst,
		descr:   pt.NewEntityDescriptor(prototype),
	}
}

// EntityPTCreator 实体原型构建器
type EntityPTCreator struct {
	svcInst service.Context
	descr   *pt.EntityDescriptor
	comps   []any
}

// SetInstance 设置实例，用于扩展实体能力
func (c *EntityPTCreator) SetInstance(instance any) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.SetInstance(instance)
	return c
}

// SetScope 设置实体的可访问作用域
func (c *EntityPTCreator) SetScope(scope ec.Scope) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.SetScope(scope)
	return c
}

// SetComponentAwakeOnFirstTouch 设置当实体组件首次被访问时，生命周期是否进入唤醒（Awake）
func (c *EntityPTCreator) SetComponentAwakeOnFirstTouch(b bool) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.SetComponentAwakeOnFirstTouch(b)
	return c
}

// SetComponentUniqueID 设置是否为实体组件分配唯一Id
func (c *EntityPTCreator) SetComponentUniqueID(b bool) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.SetComponentUniqueID(b)
	return c
}

// SetMeta 设置原型Meta信息
func (c *EntityPTCreator) SetMeta(dict map[string]any) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.SetMeta(dict)
	return c
}

// MergeMeta 合并原型Meta信息，如果存在则覆盖
func (c *EntityPTCreator) MergeMeta(dict map[string]any) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.MergeMeta(dict)
	return c
}

// MergeMetaIfAbsent 合并原型Meta信息，如果存在则跳过
func (c *EntityPTCreator) MergeMetaIfAbsent(dict map[string]any) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.MergeIfAbsent(dict)
	return c
}

// AssignMeta 赋值原型Meta信息
func (c *EntityPTCreator) AssignMeta(m meta.Meta) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.descr.AssignMeta(m)
	return c
}

// SetScript 设置脚本
func (c *EntityPTCreator) SetScript(script string) *EntityPTCreator {
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}

	if script == "" {
		exception.Panicf("goscr: %w: script is empty", exception.ErrArgs)
	}

	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		exception.Panicf("goscr: incorrect script %q format", script)
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	c.descr.Meta.Add("script_pkg", scriptPkg)
	c.descr.Meta.Add("script_ident", scriptIdent)

	return c
}

// AddComponent 添加组件
func (c *EntityPTCreator) AddComponent(comp any, name ...string) *EntityPTCreator {
	switch v := comp.(type) {
	case pt.ComponentDescriptor, *pt.ComponentDescriptor:
		c.comps = append(c.comps, v)
	default:
		c.comps = append(c.comps, pt.NewComponentDescriptor(comp).SetName(pie.First(name)))
	}
	return c
}

// AddComponentScript 添加脚本组件
func (c *EntityPTCreator) AddComponentScript(script string, name ...string) *EntityPTCreator {
	c.AddComponent(ComponentScript(script), name...)
	return c
}

// Declare 声明实体原型
func (c *EntityPTCreator) Declare() {
	if c.svcInst == nil {
		exception.Panic("goscr: svcInst is nil")
	}
	if c.descr == nil {
		exception.Panic("goscr: descr is nil")
	}
	c.svcInst.EntityLib().Declare(c.descr, c.comps...)
}
