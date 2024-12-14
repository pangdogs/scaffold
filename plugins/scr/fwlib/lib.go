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

package fwlib

import (
	_ "github.com/glebarez/sqlite"
	_ "github.com/go-redsync/redsync/v4"
	_ "github.com/redis/go-redis/v9"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "gorm.io/gorm"
	"reflect"
)

var Symbols = map[string]map[string]reflect.Value{}

// export git.golaxy.org/core
//go:generate yaegi extract git.golaxy.org/core
//go:generate yaegi extract git.golaxy.org/core/define
//go:generate yaegi extract git.golaxy.org/core/ec
//go:generate yaegi extract git.golaxy.org/core/ec/pt
//go:generate yaegi extract git.golaxy.org/core/event
//go:generate yaegi extract git.golaxy.org/core/extension
//go:generate yaegi extract git.golaxy.org/core/runtime
//go:generate yaegi extract git.golaxy.org/core/service
//go:generate yaegi extract git.golaxy.org/core/utils
//go:generate yaegi extract git.golaxy.org/core/utils/async
//go:generate yaegi extract git.golaxy.org/core/utils/exception
//go:generate yaegi extract git.golaxy.org/core/utils/generic
//go:generate yaegi extract git.golaxy.org/core/utils/iface
//go:generate yaegi extract git.golaxy.org/core/utils/meta
//go:generate yaegi extract git.golaxy.org/core/utils/option
//go:generate yaegi extract git.golaxy.org/core/utils/reinterpret
//go:generate yaegi extract git.golaxy.org/core/utils/types
//go:generate yaegi extract git.golaxy.org/core/utils/uid

// export git.golaxy.org/framework
//go:generate yaegi extract git.golaxy.org/framework
//go:generate yaegi extract git.golaxy.org/framework/net/gap
//go:generate yaegi extract git.golaxy.org/framework/net/gtp
//go:generate yaegi extract git.golaxy.org/framework/net/netpath
//go:generate yaegi extract git.golaxy.org/framework/net/gap/codec
//go:generate yaegi extract git.golaxy.org/framework/net/gap/variant
//go:generate yaegi extract git.golaxy.org/framework/net/gtp/codec
//go:generate yaegi extract git.golaxy.org/framework/net/gtp/method
//go:generate yaegi extract git.golaxy.org/framework/net/gtp/transport
//go:generate yaegi extract git.golaxy.org/framework/plugins/broker
//go:generate yaegi extract git.golaxy.org/framework/plugins/conf
//go:generate yaegi extract git.golaxy.org/framework/plugins/db
//go:generate yaegi extract git.golaxy.org/framework/plugins/dentq
//go:generate yaegi extract git.golaxy.org/framework/plugins/dentr
//go:generate yaegi extract git.golaxy.org/framework/plugins/discovery
//go:generate yaegi extract git.golaxy.org/framework/plugins/dsvc
//go:generate yaegi extract git.golaxy.org/framework/plugins/dsync
//go:generate yaegi extract git.golaxy.org/framework/plugins/gate
//go:generate yaegi extract git.golaxy.org/framework/plugins/log
//go:generate yaegi extract git.golaxy.org/framework/plugins/router
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpc
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpcstack
//go:generate yaegi extract git.golaxy.org/framework/plugins/broker/nats_broker
//go:generate yaegi extract git.golaxy.org/framework/plugins/db/dbutil
//go:generate yaegi extract git.golaxy.org/framework/plugins/db/mongodb
//go:generate yaegi extract git.golaxy.org/framework/plugins/db/redisdb
//go:generate yaegi extract git.golaxy.org/framework/plugins/db/sqldb
//go:generate yaegi extract git.golaxy.org/framework/plugins/discovery/cache_discovery
//go:generate yaegi extract git.golaxy.org/framework/plugins/discovery/etcd_discovery
//go:generate yaegi extract git.golaxy.org/framework/plugins/discovery/redis_discovery
//go:generate yaegi extract git.golaxy.org/framework/plugins/dsync/etcd_dsync
//go:generate yaegi extract git.golaxy.org/framework/plugins/dsync/redis_dsync
//go:generate yaegi extract git.golaxy.org/framework/plugins/gate/cli
//go:generate yaegi extract git.golaxy.org/framework/plugins/log/console_log
//go:generate yaegi extract git.golaxy.org/framework/plugins/log/zap_log
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpc/callpath
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpc/rpcli
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpc/rpcpcsr
//go:generate yaegi extract git.golaxy.org/framework/plugins/rpc/rpcutil
//go:generate yaegi extract git.golaxy.org/framework/utils/binaryutil
//go:generate yaegi extract git.golaxy.org/framework/utils/concurrent

// export git.golaxy.org/scaffold
//go:generate yaegi extract git.golaxy.org/scaffold/plugins/acl
//go:generate yaegi extract git.golaxy.org/scaffold/plugins/view
//go:generate yaegi extract git.golaxy.org/scaffold/tools/excelc/excelutils
