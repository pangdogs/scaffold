// Code generated by 'yaegi extract git.golaxy.org/framework/addins/db/redisdb'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/addins/db/redisdb"
	"github.com/redis/go-redis/v9"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/db/redisdb/redisdb"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Install":   reflect.ValueOf(&redisdb.Install).Elem(),
		"Name":      reflect.ValueOf(&redisdb.Name).Elem(),
		"Uninstall": reflect.ValueOf(&redisdb.Uninstall).Elem(),
		"Using":     reflect.ValueOf(&redisdb.Using).Elem(),
		"With":      reflect.ValueOf(&redisdb.With).Elem(),

		// type definitions
		"IRedisDB":       reflect.ValueOf((*redisdb.IRedisDB)(nil)),
		"RedisDBOptions": reflect.ValueOf((*redisdb.RedisDBOptions)(nil)),

		// interface wrapper definitions
		"_IRedisDB": reflect.ValueOf((*_git_golaxy_org_framework_addins_db_redisdb_IRedisDB)(nil)),
	}
}

// _git_golaxy_org_framework_addins_db_redisdb_IRedisDB is an interface wrapper for IRedisDB type
type _git_golaxy_org_framework_addins_db_redisdb_IRedisDB struct {
	IValue            interface{}
	WRedisDB          func(tag string) *redis.Client
	WReflectedRedisDB func(tag string) reflect.Value
}

func (W _git_golaxy_org_framework_addins_db_redisdb_IRedisDB) RedisDB(tag string) *redis.Client {
	return W.WRedisDB(tag)
}
func (W _git_golaxy_org_framework_addins_db_redisdb_IRedisDB) ReflectedRedisDB(tag string) reflect.Value {
	return W.WReflectedRedisDB(tag)
}
