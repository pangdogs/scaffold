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
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"github.com/elliotchance/pie/v2"
	"strings"
)

// BuildEntityPT 创建实体原型
func BuildEntityPT(svcCtx service.Context, prototype string) EntityPTCreator {
	if svcCtx == nil {
		exception.Panicf("%w: svcCtx is nil", core.ErrArgs)
	}
	c := EntityPTCreator{
		svcCtx: svcCtx,
	}
	c.atti.Prototype = prototype
	return c
}

// EntityPTCreator 实体原型构建器
type EntityPTCreator struct {
	svcCtx service.Context
	atti   pt.EntityAttribute
	comps  []any
}

// SetInstance 设置实例，用于扩展实体能力
func (c EntityPTCreator) SetInstance(instance any) EntityPTCreator {
	c.atti.Instance = instance
	return c
}

// SetScope 设置实体的可访问作用域
func (c EntityPTCreator) SetScope(scope ec.Scope) EntityPTCreator {
	c.atti.Scope = &scope
	return c
}

// SetComponentNameIndexing 设置是否开启组件名称索引
func (c EntityPTCreator) SetComponentNameIndexing(b bool) EntityPTCreator {
	c.atti.ComponentNameIndexing = &b
	return c
}

// SetComponentAwakeOnFirstTouch 设置当实体组件首次被访问时，生命周期是否进入唤醒（Awake）
func (c EntityPTCreator) SetComponentAwakeOnFirstTouch(b bool) EntityPTCreator {
	c.atti.ComponentAwakeOnFirstTouch = &b
	return c
}

// SetComponentUniqueID 设置是否为实体组件分配唯一Id
func (c EntityPTCreator) SetComponentUniqueID(b bool) EntityPTCreator {
	c.atti.ComponentUniqueID = &b
	return c
}

// SetExtra 设置自定义属性
func (c EntityPTCreator) SetExtra(extra map[string]any) EntityPTCreator {
	for k, v := range extra {
		c.atti.Extra.Add(k, v)
	}
	return c
}

// SetScript 设置脚本
func (c EntityPTCreator) SetScript(script string) EntityPTCreator {
	if script == "" {
		exception.Panicf("%w: script is empty", exception.ErrArgs)
	}

	idx := strings.LastIndexByte(script, '.')
	if idx < 0 {
		panic(fmt.Errorf("incorrect script %q format", script))
	}

	scriptPkg := script[:idx]
	scriptIdent := script[idx+1:]

	c.atti.Extra.Add("script_pkg", scriptPkg)
	c.atti.Extra.Add("script_ident", scriptIdent)

	return c
}

// AddComponent 添加组件
func (c EntityPTCreator) AddComponent(comp any, name ...string) EntityPTCreator {
	switch v := comp.(type) {
	case pt.ComponentAttribute, *pt.ComponentAttribute:
		c.comps = append(c.comps, v)
	default:
		c.comps = append(c.comps, pt.BuildComponentAttribute(comp).SetName(pie.First(name)))
	}
	return c
}

// Declare 声明实体原型
func (c EntityPTCreator) Declare() {
	if c.svcCtx == nil {
		exception.Panic("svcCtx is nil")
	}
	c.svcCtx.GetEntityLib().Declare(c.atti, c.comps...)
}

// Redeclare 重新声明实体原型
func (c EntityPTCreator) Redeclare() {
	if c.svcCtx == nil {
		exception.Panic("svcCtx is nil")
	}
	c.svcCtx.GetEntityLib().Redeclare(c.atti, c.comps...)
}
