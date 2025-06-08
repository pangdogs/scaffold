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

package propview

import (
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc"
	"git.golaxy.org/framework/addins/rpc/rpcli"
	"git.golaxy.org/framework/net/gap/variant"
	"reflect"
)

var (
	ErrEntityNotFound               = variant.Errorln(-2, "entity not found")
	ErrEntityNoPropTab              = variant.Errorln(-3, "entity hasn't prop table")
	ErrEntityNoProp                 = variant.Errorln(-4, "entity hasn't prop")
	ErrOutdatedRevision             = variant.Errorln(-5, "synchronized revision is outdated")
	ErrDiscontinuousRevision        = variant.Errorln(-6, "synchronized revision is discontinuous")
	ErrMethodNotFound               = variant.Errorln(-7, "op method not found")
	ErrMethodParameterCountMismatch = variant.Errorln(-8, "op method parameter count mismatch")
	ErrMethodParameterTypeMismatch  = variant.Errorln(-9, "op method parameter type mismatch")
	ErrLoadFromServiceItself        = variant.Errorln(-10, "can't load data from the service itself")
	ErrSaveToServiceItself          = variant.Errorln(-11, "can't save data to the service itself")
)

// IPropView 属性视图插件接口
type IPropView interface {
	// Load 加载属性数据
	Load(entityId uid.Id, prop string, service string) ([]byte, int64, error)
	// Save 保存属性数据
	Save(entityId uid.Id, prop string, service string, data []byte, revision int64) error
	// Sync 同步属性变化
	Sync(entityId uid.Id, prop string, syncTo []string, revision int64, op string, args ...any)
}

func newPropView(...any) IPropView {
	return &_PropView{}
}

type _PropView struct {
	rt framework.IRuntime
}

func (m *_PropView) Init(rtCtx runtime.Context) {
	log.Debugf(rtCtx, "init addin %q", self.Name)

	m.rt = framework.GetRuntime(rtCtx)
}

func (m *_PropView) Shut(rtCtx runtime.Context) {
	log.Debugf(rtCtx, "shut addin %q", self.Name)
}

func (m *_PropView) Load(entityId uid.Id, prop string, service string) ([]byte, int64, error) {
	if service == m.rt.GetService().GetName() {
		log.Errorf(m.rt, "load data failed, entity=%q, prop=%q, service=%q, %s", entityId, prop, service, ErrLoadFromServiceItself)
		return nil, 0, ErrLoadFromServiceItself
	}
	return rpc.Assert3[[]byte, int64, error](
		<-rpc.ProxyRuntime(m.rt, entityId).RPC(service, Name, "DoLoad", entityId, prop),
	)
}

func (m *_PropView) Save(entityId uid.Id, prop string, service string, data []byte, revision int64) error {
	if service == m.rt.GetService().GetName() {
		log.Errorf(m.rt, "save data failed, entity=%q, prop=%q, service=%q, size=%d, revision=%d, %s",
			entityId, prop, service, len(data), revision, ErrSaveToServiceItself)
		return ErrSaveToServiceItself
	}
	return rpc.Assert1[error](
		<-rpc.ProxyRuntime(m.rt, entityId).RPC(service, Name, "DoSave", entityId, prop, data, revision),
	)
}

func (m *_PropView) Sync(entityId uid.Id, prop string, syncTo []string, revision int64, op string, args ...any) {
	for _, dst := range syncTo {
		if gate.CliDetails.DomainUnicast.Equal(dst) {
			// 同步至实体客户端
			rpc.ProxyEntity(m.rt, entityId).CliOnewayRPC(rpcli.Main, "DoSync", prop, revision, op, args)

		} else if gate.CliDetails.DomainMulticast.Contains(dst) {
			// 同步至指定分组
			group, _ := gate.CliDetails.DomainMulticast.Relative(dst)
			rpc.ProxyGroup(m.rt, dst).CliOnewayRPC(group, "DoSync", entityId, prop, revision, op, args)

		} else if gate.CliDetails.DomainBroadcast.Equal(dst) {
			// 同步至包含实体的所有分组
			rpc.ProxyEntity(m.rt, entityId).BroadcastCliOnewayRPC(rpcli.Main, "DoSync", prop, revision, op, args)

		} else {
			// 同步至其他服务
			core.Await(m.rt, rpc.ProxyRuntime(m.rt, entityId).RPC(dst, Name, "DoSync", entityId, prop, revision, op, args)).
				AnyVoid(m.doSyncRet, dst, entityId, prop, revision, op)
		}
	}
}

func (m *_PropView) DoLoad(entityId uid.Id, propName string) ([]byte, int64, error) {
	caller := m.rt.GetRPCStack().CallChain().Last()

	entity, ok := m.rt.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m.rt, `do load data failed, entity=%q, prop=%q, caller="%s:%s", %s`,
			entityId, propName, caller.Svc, caller.Addr, ErrEntityNotFound)
		return nil, 0, ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m.rt, `do load data failed, entity=%q, prop=%q, caller="%s:%s", %s`,
			entityId, propName, caller.Svc, caller.Addr, ErrEntityNoPropTab)
		return nil, 0, ErrEntityNoPropTab
	}

	prop := propTab.GetProp(propName)
	if prop == nil {
		log.Errorf(m.rt, `do load data failed, entity=%q, prop=%q, caller="%s:%s", %s`,
			entityId, propName, caller.Svc, caller.Addr, ErrEntityNoProp)
		return nil, 0, ErrEntityNoProp
	}

	data, revision, err := prop.Managed().Marshal()
	if err != nil {
		log.Errorf(m.rt, `do load data failed, entity=%q, prop=%q, caller="%s:%s", marshal failed, %s`,
			entityId, propName, caller.Svc, caller.Addr, err)
		return nil, 0, err
	}

	log.Infof(m.rt, `do load data ok, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s"`,
		entityId, propName, len(data), revision, caller.Svc, caller.Addr)
	return data, revision, nil
}

func (m *_PropView) DoSave(entityId uid.Id, propName string, data []byte, revision int64) error {
	caller := m.rt.GetRPCStack().CallChain().Last()

	entity, ok := m.rt.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m.rt, `do save data failed, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s", %s`,
			entityId, propName, len(data), revision, caller.Svc, caller.Addr, ErrEntityNotFound)
		return ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m.rt, `do save data failed, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s", %s`,
			entityId, propName, len(data), revision, caller.Svc, caller.Addr, ErrEntityNoPropTab)
		return ErrEntityNoPropTab
	}

	prop := propTab.GetProp(propName)
	if prop == nil {
		log.Errorf(m.rt, `do save data failed, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s", %s`,
			entityId, propName, len(data), revision, caller.Svc, caller.Addr, ErrEntityNoProp)
		return ErrEntityNoProp
	}

	err := prop.Managed().Unmarshal(data, revision)
	if err != nil {
		log.Errorf(m.rt, `do save data failed, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s", unmarshal failed, %s`,
			entityId, propName, len(data), revision, caller.Svc, caller.Addr, err)
		return err
	}

	log.Infof(m.rt, `do save data ok, entity=%q, prop=%q, size=%d, revision=%d, caller="%s:%s`,
		entityId, propName, len(data), revision, caller.Svc, caller.Addr)
	return nil
}

func (m *_PropView) DoSync(entityId uid.Id, propName string, revision int64, op string, argsRV []reflect.Value) error {
	caller := m.rt.GetRPCStack().CallChain().Last()

	entity, ok := m.rt.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, op, caller.Svc, caller.Addr, ErrEntityNotFound)
		return ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, op, caller.Svc, caller.Addr, ErrEntityNoPropTab)
		return ErrEntityNoPropTab
	}

	prop := propTab.GetProp(propName)
	if prop == nil {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, op, caller.Svc, caller.Addr, ErrEntityNoProp)
		return ErrEntityNoProp
	}

	if revision <= prop.Managed().Revision() {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, local_revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, prop.Managed().Revision(), op, caller.Svc, caller.Addr, ErrOutdatedRevision)
		return ErrOutdatedRevision
	}

	if revision != prop.Managed().Revision()+1 {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, local_revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, prop.Managed().Revision(), op, caller.Svc, caller.Addr, ErrDiscontinuousRevision)
		return ErrDiscontinuousRevision
	}

	methodRV := prop.ReflectedManaged().MethodByName(op)
	if !methodRV.IsValid() {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, op, caller.Svc, caller.Addr, ErrMethodNotFound)
		return ErrMethodNotFound
	}
	methodRT := methodRV.Type()

	if methodRT.NumIn() != len(argsRV) {
		log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
			entityId, propName, revision, op, caller.Svc, caller.Addr, ErrMethodParameterCountMismatch)
		return ErrMethodParameterCountMismatch
	}

	for i := range argsRV {
		argRV := argsRV[i]
		argRT := argRV.Type()
		paramRT := methodRT.In(i)

	retry:
		if argRT.AssignableTo(paramRT) {
			argsRV[i] = argRV
			continue
		}

		if argRV.CanConvert(paramRT) {
			if argRT.Size() > paramRT.Size() {
				log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
					entityId, propName, revision, op, caller.Svc, caller.Addr, ErrMethodParameterTypeMismatch)
				return ErrMethodParameterTypeMismatch
			}
			argsRV[i] = argRV.Convert(paramRT)
			continue
		}

		if argRT.Kind() == reflect.Pointer {
			argRV = argRV.Elem()
			argRT = argRV.Type()
			goto retry
		}

		argRV, err := argsRV[i].Interface().(variant.Variant).Convert(paramRT)
		if err != nil {
			log.Errorf(m.rt, `do sync op failed, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s", %s`,
				entityId, propName, revision, op, caller.Svc, caller.Addr, ErrMethodParameterTypeMismatch)
			return ErrMethodParameterTypeMismatch
		}

		argsRV[i] = argRV
	}

	methodRV.Call(argsRV)
	prop.Managed().incrRevision()

	log.Infof(m.rt, `do sync op ok, entity=%q, prop=%q, revision=%d, op=%q, caller="%s:%s"`,
		entityId, propName, revision, op, caller.Svc, caller.Addr)
	return nil
}

func (m *_PropView) doSyncRet(ctx runtime.Context, ret async.Ret, args ...any) {
	err, retErr := rpc.Result1[error](ret).Extract()
	if err == nil && retErr == nil {
		return
	}

	dst := args[0].(string)
	entityId := args[1].(uid.Id)
	propName := args[2].(string)
	revision := args[3].(int64)
	op := args[4].(string)

	if retErr != nil {
		log.Errorf(m.rt, `sync op failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s`, entityId, propName, revision, op, dst, err)
		return
	}

	var syncErr *variant.Error

	if ok := errors.As(err, &syncErr); !ok {
		log.Errorf(m.rt, `sync op failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s`, entityId, propName, revision, op, dst, err)
		return
	}

	switch syncErr.Code {
	case ErrOutdatedRevision.Code, ErrDiscontinuousRevision.Code, ErrMethodNotFound.Code, ErrMethodParameterCountMismatch.Code, ErrMethodParameterTypeMismatch.Code:
		log.Warnf(m.rt, `sync op failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s, trying to save`, entityId, propName, revision, op, dst, err)

		entity, ok := m.rt.GetEntityManager().GetEntity(entityId)
		if !ok {
			log.Errorf(m.rt, `sync op trying to save failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s`, entityId, propName, revision, op, dst, ErrEntityNotFound)
			return
		}

		err := entity.(IPropTab).GetProp(propName).Save(dst)
		if err != nil {
			log.Errorf(m.rt, `sync op trying to save failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s`, entityId, propName, revision, op, dst, err)
			return
		}

		log.Infof(m.rt, `sync op trying to save ok, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q`, entityId, propName, revision, op, dst)
		return
	default:
		log.Errorf(m.rt, `sync op failed, entity=%q, prop=%q, revision=%d, op=%q, sync_to_dst=%q, %s`, entityId, propName, revision, op, dst, err)
		return
	}
}
