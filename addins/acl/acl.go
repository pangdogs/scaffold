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

package acl

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework"
	"git.golaxy.org/framework/addins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

// IACL 访问控制表插件接口
type IACL interface {
	// Enabled 是否开启
	Enabled(tag string) bool
	// Allowed 是否允许
	Allowed(tag, item string) bool
}

func newACL(...any) IACL {
	return &_ACL{
		config: viper.New(),
	}
}

type _ACL struct {
	svc    framework.IService
	config *viper.Viper
}

func (acl *_ACL) Init(svcCtx service.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	acl.svc = reinterpret.Cast[framework.IService](svcCtx)

	localFilePath := acl.svc.GetServiceConf().GetString("acl.local_path")
	remoteFilePath := acl.svc.GetServiceConf().GetString("acl.remote_path")
	if localFilePath == "" && remoteFilePath == "" {
		log.Infof(acl.svc, "no acl file")
		return
	}

	acl.config.SetConfigType(acl.svc.GetServiceConf().GetString("acl.format"))

	if localFilePath != "" {
		acl.config.SetConfigFile(localFilePath)
		if err := acl.config.ReadInConfig(); err != nil {
			log.Panicf(acl.svc, "read acl local config %q failed, %s", localFilePath, err)
		}
		log.Infof(acl.svc, "load acl local config %q config ok", localFilePath)
	}

	remoteProvider := acl.svc.GetServiceConf().GetString("acl.remote_provider")
	remoteEndpoint := acl.svc.GetServiceConf().GetString("acl.remote_endpoint")

	if remoteFilePath != "" {
		if err := acl.config.AddRemoteProvider(remoteProvider, remoteEndpoint, remoteFilePath); err != nil {
			log.Panicf(acl.svc, `read acl remote config "%s - %s - %s" failed, %s`, remoteProvider, remoteEndpoint, remoteFilePath, err)
		}
		if err := acl.config.ReadRemoteConfig(); err != nil {
			log.Panicf(acl.svc, `read acl remote config "%s - %s - %s" failed, %s`, remoteProvider, remoteEndpoint, remoteFilePath, err)
		}

		log.Infof(acl.svc, `load acl remote config "%s - %s - %s" ok`, remoteProvider, remoteEndpoint, remoteFilePath)
	}

	if localFilePath != "" {
		acl.config.OnConfigChange(func(e fsnotify.Event) {
			log.Infof(acl.svc, "reload acl local config %q ok", localFilePath)
		})
		acl.config.WatchConfig()
	}

	if remoteFilePath != "" {
		go func() {
			for {
				time.Sleep(time.Second * 3)

				err := acl.config.WatchRemoteConfig()
				if err != nil {
					log.Errorf(acl.svc, `watch acl remote config "%s - %s - %s" changes failed, %s`, remoteProvider, remoteEndpoint, remoteFilePath, err)
					continue
				}

				log.Infof(acl.svc, `reload acl remote config "%s - %s - %s" ok`, remoteProvider, remoteEndpoint, remoteFilePath)
			}
		}()
	}
}

func (acl *_ACL) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)
}

// Enabled 是否开启
func (acl *_ACL) Enabled(tag string) bool {
	return acl.config.GetBool(tag+".allow.enable") || acl.config.GetBool(tag+".deny.enable")
}

// Allowed 是否允许
func (acl *_ACL) Allowed(tag, item string) bool {
	if acl.config.GetBool(tag + ".allow.enable") {
		return acl.config.GetBool(tag + ".allow.items." + item)
	}

	if acl.config.GetBool(tag + ".deny.enable") {
		if acl.config.GetBool(tag + ".deny.items." + item) {
			return false
		}
	}

	return true
}
