// Code generated by 'yaegi extract git.golaxy.org/framework/addins/discovery/etcd_discovery'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/addins/discovery/etcd_discovery"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/discovery/etcd_discovery/etcd_discovery"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Install":     reflect.ValueOf(&etcd_discovery.Install).Elem(),
		"NewRegistry": reflect.ValueOf(etcd_discovery.NewRegistry),
		"Uninstall":   reflect.ValueOf(&etcd_discovery.Uninstall).Elem(),
		"With":        reflect.ValueOf(&etcd_discovery.With).Elem(),

		// type definitions
		"RegistryOptions": reflect.ValueOf((*etcd_discovery.RegistryOptions)(nil)),
	}
}
