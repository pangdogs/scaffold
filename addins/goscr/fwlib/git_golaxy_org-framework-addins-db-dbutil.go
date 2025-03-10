// Code generated by 'yaegi extract git.golaxy.org/framework/addins/db/dbutil'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/addins/db/dbutil"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/db/dbutil/dbutil"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"MigrateDB": reflect.ValueOf(dbutil.MigrateDB),
		"MongoDB":   reflect.ValueOf(dbutil.MongoDB),
		"RedisDB":   reflect.ValueOf(dbutil.RedisDB),
		"SQLDB":     reflect.ValueOf(dbutil.SQLDB),

		// type definitions
		"IMigrateDB": reflect.ValueOf((*dbutil.IMigrateDB)(nil)),

		// interface wrapper definitions
		"_IMigrateDB": reflect.ValueOf((*_git_golaxy_org_framework_addins_db_dbutil_IMigrateDB)(nil)),
	}
}

// _git_golaxy_org_framework_addins_db_dbutil_IMigrateDB is an interface wrapper for IMigrateDB type
type _git_golaxy_org_framework_addins_db_dbutil_IMigrateDB struct {
	IValue     interface{}
	WMigrateDB func() error
}

func (W _git_golaxy_org_framework_addins_db_dbutil_IMigrateDB) MigrateDB() error {
	return W.WMigrateDB()
}
