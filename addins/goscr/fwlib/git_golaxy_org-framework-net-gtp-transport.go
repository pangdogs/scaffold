// Code generated by 'yaegi extract git.golaxy.org/framework/net/gtp/transport'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"io"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/net/gtp/transport/transport"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CastRstErr":                 reflect.ValueOf(transport.CastRstErr),
		"EOF":                        reflect.ValueOf(&transport.EOF).Elem(),
		"ErrClosed":                  reflect.ValueOf(&transport.ErrClosed).Elem(),
		"ErrDeadlineExceeded":        reflect.ValueOf(&transport.ErrDeadlineExceeded).Elem(),
		"ErrDiscardSeq":              reflect.ValueOf(&transport.ErrDiscardSeq).Elem(),
		"ErrEvent":                   reflect.ValueOf(&transport.ErrEvent).Elem(),
		"ErrIncorrectMsg":            reflect.ValueOf(&transport.ErrIncorrectMsg).Elem(),
		"ErrNetIO":                   reflect.ValueOf(&transport.ErrNetIO).Elem(),
		"ErrProtocol":                reflect.ValueOf(&transport.ErrProtocol).Elem(),
		"ErrRenew":                   reflect.ValueOf(&transport.ErrRenew).Elem(),
		"ErrShortBuffer":             reflect.ValueOf(&transport.ErrShortBuffer).Elem(),
		"ErrShortWrite":              reflect.ValueOf(&transport.ErrShortWrite).Elem(),
		"ErrSynchronizer":            reflect.ValueOf(&transport.ErrSynchronizer).Elem(),
		"ErrTrans":                   reflect.ValueOf(&transport.ErrTrans).Elem(),
		"ErrUnexpectedEOF":           reflect.ValueOf(&transport.ErrUnexpectedEOF).Elem(),
		"ErrUnexpectedMsg":           reflect.ValueOf(&transport.ErrUnexpectedMsg).Elem(),
		"ErrUnexpectedSeq":           reflect.ValueOf(&transport.ErrUnexpectedSeq).Elem(),
		"NewSequencedSynchronizer":   reflect.ValueOf(transport.NewSequencedSynchronizer),
		"NewUnsequencedSynchronizer": reflect.ValueOf(transport.NewUnsequencedSynchronizer),

		// type definitions
		"AuthAccept":                reflect.ValueOf((*transport.AuthAccept)(nil)),
		"ChangeCipherSpecAccept":    reflect.ValueOf((*transport.ChangeCipherSpecAccept)(nil)),
		"ChangeCipherSpecFin":       reflect.ValueOf((*transport.ChangeCipherSpecFin)(nil)),
		"ContinueAccept":            reflect.ValueOf((*transport.ContinueAccept)(nil)),
		"CtrlProtocol":              reflect.ValueOf((*transport.CtrlProtocol)(nil)),
		"ECDHESecretKeyExchangeFin": reflect.ValueOf((*transport.ECDHESecretKeyExchangeFin)(nil)),
		"ErrorHandler":              reflect.ValueOf((*transport.ErrorHandler)(nil)),
		"EventDispatcher":           reflect.ValueOf((*transport.EventDispatcher)(nil)),
		"EventHandler":              reflect.ValueOf((*transport.EventHandler)(nil)),
		"FinishedAccept":            reflect.ValueOf((*transport.FinishedAccept)(nil)),
		"HandshakeProtocol":         reflect.ValueOf((*transport.HandshakeProtocol)(nil)),
		"HeartbeatHandler":          reflect.ValueOf((*transport.HeartbeatHandler)(nil)),
		"HelloAccept":               reflect.ValueOf((*transport.HelloAccept)(nil)),
		"HelloFin":                  reflect.ValueOf((*transport.HelloFin)(nil)),
		"IEvent":                    reflect.ValueOf((*transport.IEvent)(nil)),
		"ISynchronizer":             reflect.ValueOf((*transport.ISynchronizer)(nil)),
		"PayloadHandler":            reflect.ValueOf((*transport.PayloadHandler)(nil)),
		"Retry":                     reflect.ValueOf((*transport.Retry)(nil)),
		"RstError":                  reflect.ValueOf((*transport.RstError)(nil)),
		"RstHandler":                reflect.ValueOf((*transport.RstHandler)(nil)),
		"SecretKeyExchangeAccept":   reflect.ValueOf((*transport.SecretKeyExchangeAccept)(nil)),
		"SequencedSynchronizer":     reflect.ValueOf((*transport.SequencedSynchronizer)(nil)),
		"SyncTimeHandler":           reflect.ValueOf((*transport.SyncTimeHandler)(nil)),
		"TransProtocol":             reflect.ValueOf((*transport.TransProtocol)(nil)),
		"Transceiver":               reflect.ValueOf((*transport.Transceiver)(nil)),
		"UnsequencedSynchronizer":   reflect.ValueOf((*transport.UnsequencedSynchronizer)(nil)),

		// interface wrapper definitions
		"_ISynchronizer": reflect.ValueOf((*_git_golaxy_org_framework_net_gtp_transport_ISynchronizer)(nil)),
	}
}

// _git_golaxy_org_framework_net_gtp_transport_ISynchronizer is an interface wrapper for ISynchronizer type
type _git_golaxy_org_framework_net_gtp_transport_ISynchronizer struct {
	IValue           interface{}
	WAck             func(ack uint32)
	WAckSeq          func() uint32
	WCached          func() int
	WCap             func() int
	WClean           func()
	WRecvSeq         func() uint32
	WSendSeq         func() uint32
	WSynchronization func(remoteRecvSeq uint32) error
	WValidate        func(msgHead gtp.MsgHead, msgBuf []byte) error
	WWrite           func(p []byte) (n int, err error)
	WWriteTo         func(w io.Writer) (n int64, err error)
}

func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Ack(ack uint32) { W.WAck(ack) }
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) AckSeq() uint32 {
	return W.WAckSeq()
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Cached() int { return W.WCached() }
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Cap() int    { return W.WCap() }
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Clean()      { W.WClean() }
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) RecvSeq() uint32 {
	return W.WRecvSeq()
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) SendSeq() uint32 {
	return W.WSendSeq()
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Synchronization(remoteRecvSeq uint32) error {
	return W.WSynchronization(remoteRecvSeq)
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Validate(msgHead gtp.MsgHead, msgBuf []byte) error {
	return W.WValidate(msgHead, msgBuf)
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) Write(p []byte) (n int, err error) {
	return W.WWrite(p)
}
func (W _git_golaxy_org_framework_net_gtp_transport_ISynchronizer) WriteTo(w io.Writer) (n int64, err error) {
	return W.WWriteTo(w)
}