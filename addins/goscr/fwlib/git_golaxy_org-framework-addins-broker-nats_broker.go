// Code generated by 'yaegi extract git.golaxy.org/framework/addins/broker/nats_broker'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/addins/broker/nats_broker"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/addins/broker/nats_broker/nats_broker"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Install":   reflect.ValueOf(&nats_broker.Install).Elem(),
		"Uninstall": reflect.ValueOf(&nats_broker.Uninstall).Elem(),
		"With":      reflect.ValueOf(&nats_broker.With).Elem(),

		// type definitions
		"BrokerOptions": reflect.ValueOf((*nats_broker.BrokerOptions)(nil)),
	}
}
