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

package view

import (
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc"
	"git.golaxy.org/framework/addins/rpc/rpcli"
	"git.golaxy.org/framework/addins/rpc/rpcutil"
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
)

// IPropView 属性视图插件接口
type IPropView interface {
	iPropView
}

type iPropView interface {
	load(ps *PropSync, service string) ([]byte, int64, error)
	save(ps *PropSync, service string, data []byte, revision int64) error
	sync(ps *PropSync, revision int64, op string, args ...any)
}

func newPropView(...any) IPropView {
	return &_PropView{}
}

type _PropView struct {
	framework.IRuntimeInstance
}

func (m *_PropView) Init(_ service.Context, rtCtx runtime.Context) {
	log.Debugf(rtCtx, "init addin %q", self.Name)

	m.IRuntimeInstance = framework.GetRuntimeInstance(rtCtx)
}

func (m *_PropView) Shut(_ service.Context, rtCtx runtime.Context) {
	log.Debugf(rtCtx, "shut addin %q", self.Name)
}

func (m *_PropView) load(ps *PropSync, service string) ([]byte, int64, error) {
	if service == m.GetService().GetName() {
		return nil, 0, errors.New("can't load data from the service itself")
	}
	return rpc.Assert3[[]byte, int64, error](
		<-rpcutil.ProxyRuntime(m, ps.entity.GetId()).RPC(service, Name, "DoLoad", ps.entity.GetId(), ps.name),
	)
}

func (m *_PropView) save(ps *PropSync, service string, data []byte, revision int64) error {
	if service == m.GetService().GetName() {
		return errors.New("can't save data to the service itself")
	}
	return rpc.Assert1[error](
		<-rpcutil.ProxyRuntime(m, ps.entity.GetId()).RPC(service, Name, "DoSave", ps.entity.GetId(), ps.name, data, revision),
	)
}

func (m *_PropView) sync(ps *PropSync, revision int64, op string, args ...any) {
	for _, dst := range ps.syncTo {
		if gate.CliDetails.DomainUnicast.Equal(dst) {
			// 同步至实体客户端
			rpcutil.ProxyEntity(m, ps.entity.GetId()).CliOnewayRPC(rpcli.Main, "DoSync", ps.name, revision, op, args)

		} else if gate.CliDetails.DomainMulticast.Contains(dst) {
			// 同步至指定分组
			group, _ := gate.CliDetails.DomainMulticast.Relative(dst)
			rpcutil.ProxyGroup(m, dst).CliOnewayRPC(group, "DoSync", ps.entity.GetId(), ps.name, revision, op, args)

		} else if gate.CliDetails.DomainBroadcast.Equal(dst) {
			// 同步至包含实体的所有分组
			rpcutil.ProxyEntity(m, ps.entity.GetId()).BroadcastCliOnewayRPC(rpcli.Main, "DoSync", ps.name, revision, op, args)

		} else {
			// 同步至其他服务
			core.Await(m, rpcutil.ProxyRuntime(m, ps.entity.GetId()).RPC(dst, Name, "DoSync", ps.entity.GetId(), ps.name, revision, op, args)).
				Any(m.syncRet, dst, ps.entity, ps.name, revision, op)
		}
	}
}

func (m *_PropView) syncRet(ctx runtime.Context, ret async.Ret, args ...any) {
	err, retErr := rpc.Result1[error](ret).Extract()
	if err == nil && retErr == nil {
		return
	}

	dst := args[0].(string)
	entity := args[1].(ec.Entity)
	name := args[2].(string)
	revision := args[3].(int64)
	op := args[4].(string)

	if retErr != nil {
		log.Errorf(m, "sync entity %q prop %q revision %d op %q to %q failed, %s", entity.GetId(), name, revision, op, dst, retErr)
		return
	}

	var syncErr *variant.Error

	if ok := errors.As(err, &syncErr); !ok {
		log.Errorf(m, "sync entity %q prop %q revision %d op %q to %q failed, %s", entity.GetId(), name, revision, op, dst, err)
		return
	}

	switch syncErr.Code {
	case ErrOutdatedRevision.Code, ErrDiscontinuousRevision.Code, ErrMethodNotFound.Code,
		ErrMethodParameterCountMismatch.Code, ErrMethodParameterTypeMismatch.Code:
		log.Warnf(m, "sync entity %q prop %q revision %d op %q to %q failed, %s, retry save", entity.GetId(), name, revision, op, dst, err)

		err := entity.(IPropTab).GetProp(name).Save(dst)
		if err != nil {
			log.Errorf(m, "save entity %q prop %q revision %d to %q failed, %s", entity.GetId(), name, revision, dst, err)
		} else {
			log.Infof(m, "save entity %q prop %q revision %d to %q ok", entity.GetId(), name, revision, dst)
		}
		return
	default:
		log.Errorf(m, "sync entity %q prop %q revision %d op %q to %q failed, %s", entity.GetId(), name, revision, op, dst, syncErr)
		return
	}
}

func (m *_PropView) DoLoad(entityId uid.Id, name string) ([]byte, int64, error) {
	entity, ok := m.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m, "do load entity %q prop %q failed, entity not found", entityId, name)
		return nil, 0, ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m, "do load entity %q prop %q failed, the entity hasn't prop table", entityId, name)
		return nil, 0, ErrEntityNoPropTab
	}

	prop := propTab.GetProp(name)
	if prop == nil {
		log.Errorf(m, "do load entity %q prop %q failed, the entity hasn't prop", entityId, name)
		return nil, 0, ErrEntityNoProp
	}

	data, revision, err := prop.Marshal()
	if err != nil {
		log.Errorf(m, "do load entity %q prop %q failed, marshal failed, %s", entityId, name, err)
		return nil, 0, err
	}

	log.Infof(m, "do load entity %q prop %q ok, revision:%d, caller:%q", entityId, name, revision, m.GetRPCStack().CallChain().Last().Addr)
	return data, revision, nil
}

func (m *_PropView) DoSave(entityId uid.Id, name string, data []byte, revision int64) error {
	entity, ok := m.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m, "do save entity %q prop %q revision %d failed, entity not found", entityId, name, revision)
		return ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m, "do save entity %q prop %q revision %d failed, the entity hasn't prop table", entityId, name, revision)
		return ErrEntityNoPropTab
	}

	prop := propTab.GetProp(name)
	if prop == nil {
		log.Errorf(m, "do save entity %q prop %q revision %d failed, the entity hasn't prop", entityId, name, revision)
		return ErrEntityNoProp
	}

	err := prop.Unmarshal(data, revision)
	if err != nil {
		log.Errorf(m, "do save entity %q prop %q revision %d failed, unmarshal failed, %s", entityId, name, revision, err)
		return err
	}

	log.Infof(m, "do save entity %q prop %q revision %d ok, caller:%q", entityId, name, revision, m.GetRPCStack().CallChain().Last().Addr)
	return nil
}

func (m *_PropView) DoSync(entityId uid.Id, name string, revision int64, op string, argsRV []reflect.Value) error {
	entity, ok := m.GetEntityManager().GetEntity(entityId)
	if !ok {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, entity not found", entityId, name, revision, op)
		return ErrEntityNotFound
	}

	propTab, ok := entity.(IPropTab)
	if !ok {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, the entity hasn't prop table", entityId, name, revision, op)
		return ErrEntityNoPropTab
	}

	prop := propTab.GetProp(name)
	if prop == nil {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, the entity hasn't prop", entityId, name, revision, op)
		return ErrEntityNoProp
	}

	if revision <= prop.Revision() {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s, local revision %d",
			entityId, name, revision, op, ErrOutdatedRevision, prop.Revision())
		return ErrOutdatedRevision
	}

	if revision != prop.Revision()+1 {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s, local revision %d",
			entityId, name, revision, op, ErrDiscontinuousRevision, prop.Revision())
		return ErrDiscontinuousRevision
	}

	methodRV := prop.Reflected().MethodByName(op)
	if !methodRV.IsValid() {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s", entityId, name, revision, op, ErrMethodNotFound)
		return ErrMethodNotFound
	}
	methodRT := methodRV.Type()

	if methodRT.NumIn() != len(argsRV) {
		log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s", entityId, name, revision, op, ErrMethodParameterCountMismatch)
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
				log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s", entityId, name, revision, op, ErrMethodParameterTypeMismatch)
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
			log.Errorf(m, "do sync entity %q prop %q revision %d op %q failed, %s", entityId, name, revision, op, ErrMethodParameterTypeMismatch)
			return ErrMethodParameterTypeMismatch
		}

		argsRV[i] = argRV
	}

	methodRV.Call(argsRV)
	prop.incrRevision()

	log.Debugf(m, "do sync entity %q prop %q revision %d op %q ok", entityId, name, revision, op)
	return nil
}
