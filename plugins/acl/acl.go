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
	"git.golaxy.org/framework/plugins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

// IACL 访问控制表
type IACL interface {
	// Enabled 是否开启
	Enabled() bool
	// AllowMobile 检测手机号是否允许
	AllowMobile(mobile string) bool
	// AllowGuest 检测游客是否允许
	AllowGuest() bool
}

func newACL(...any) IACL {
	return &_ACL{
		setting: viper.New(),
	}
}

type _ACL struct {
	framework.IServiceInstance
	setting *viper.Viper
	tab     *map[string]struct{}
}

func (acl *_ACL) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	acl.IServiceInstance = reinterpret.Cast[framework.IServiceInstance](ctx)

	localFilePath := acl.GetConf().GetStringOrDefault("acl.local_path", "")
	remoteFilePath := acl.GetConf().GetStringOrDefault("acl.remote_path", "")
	if localFilePath == "" && remoteFilePath == "" {
		log.Infof(acl, "no acl file")
		return
	}

	acl.setting.SetConfigType(acl.GetConf().GetStringOrDefault("acl.format", "json"))

	if localFilePath != "" {
		acl.setting.SetConfigFile(localFilePath)
		if err := acl.setting.ReadInConfig(); err != nil {
			log.Panicf(acl, "read acl local config %q failed, %s", localFilePath, err)
		}
		log.Infof(acl, "load acl local config %q config ok", localFilePath)
	}

	remoteProvider := acl.GetConf().GetStringOrDefault("acl.remote_provider", "")
	remoteEndpoint := acl.GetConf().GetStringOrDefault("acl.remote_endpoint", "")

	if remoteFilePath != "" {
		if err := acl.setting.AddRemoteProvider(remoteProvider, remoteEndpoint, remoteFilePath); err != nil {
			log.Panicf(acl, "read acl remote config [%q, %q, %q] failed, %s", remoteProvider, remoteEndpoint, remoteFilePath, err)
		}
		if err := acl.setting.ReadRemoteConfig(); err != nil {
			log.Panicf(acl, "read acl remote config [%q, %q, %q] failed, %s", remoteProvider, remoteEndpoint, remoteFilePath, err)
		}

		log.Infof(acl, "load acl remote config [%q, %q, %q] ok", remoteProvider, remoteEndpoint, remoteFilePath)
	}

	hotFix := func() {
		aclTab := map[string]struct{}{}

		for _, v := range acl.setting.GetStringSlice("allow.mobile_list") {
			aclTab["mobile:allow:"+v] = struct{}{}
		}

		for _, v := range acl.setting.GetStringSlice("deny.mobile_list") {
			aclTab["mobile:deny:"+v] = struct{}{}
		}

		acl.tab = &aclTab
	}
	hotFix()

	if localFilePath != "" {
		acl.setting.OnConfigChange(func(e fsnotify.Event) {
			hotFix()
			log.Infof(acl, "reload acl local config %q ok", localFilePath)
		})
		acl.setting.WatchConfig()
	}

	if remoteFilePath != "" {
		go func() {
			for {
				time.Sleep(time.Second * 3)

				err := acl.setting.WatchRemoteConfig()
				if err != nil {
					log.Errorf(acl, "watch acl remote config [%q, %q, %q] changes failed, %s", remoteProvider, remoteEndpoint, remoteFilePath, err)
					continue
				}
				hotFix()

				log.Infof(acl, "reload acl remote config [%q, %q, %q] ok", remoteProvider, remoteEndpoint, remoteFilePath)
			}
		}()
	}
}

func (acl *_ACL) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)
}

// Enabled 是否开启
func (acl *_ACL) Enabled() bool {
	return acl.setting.GetBool("allow.enable") || acl.setting.GetBool("deny.enable")
}

// AllowMobile 检测手机号是否允许
func (acl *_ACL) AllowMobile(mobile string) bool {
	aclTab := *acl.tab
	if aclTab == nil {
		return false
	}

	if acl.setting.GetBool("allow.enable") {
		_, ok := aclTab["mobile:allow:"+mobile]
		if !ok {
			return false
		}
	}

	if acl.setting.GetBool("deny.enable") {
		_, ok := aclTab["mobile:deny:"+mobile]
		if ok {
			return false
		}
	}

	return true
}

// AllowGuest 检测游客是否允许
func (acl *_ACL) AllowGuest() bool {
	if !acl.setting.GetBool("deny.enable") {
		return true
	}
	return acl.setting.GetBool("deny.guest")
}
