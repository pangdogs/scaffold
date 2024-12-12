// Code generated by 'yaegi extract git.golaxy.org/framework/net/gap'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/net/gap"
	"go/constant"
	"go/token"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/net/gap/gap"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"DefaultMsgCreator": reflect.ValueOf(gap.DefaultMsgCreator),
		"ErrGAP":            reflect.ValueOf(&gap.ErrGAP).Elem(),
		"ErrNotDeclared":    reflect.ValueOf(&gap.ErrNotDeclared).Elem(),
		"MakeMsgId":         reflect.ValueOf(gap.MakeMsgId),
		"MsgId_Customize":   reflect.ValueOf(constant.MakeFromLiteral("32", token.INT, 0)),
		"MsgId_Forward":     reflect.ValueOf(gap.MsgId_Forward),
		"MsgId_None":        reflect.ValueOf(gap.MsgId_None),
		"MsgId_OnewayRPC":   reflect.ValueOf(gap.MsgId_OnewayRPC),
		"MsgId_RPC_Reply":   reflect.ValueOf(gap.MsgId_RPC_Reply),
		"MsgId_RPC_Request": reflect.ValueOf(gap.MsgId_RPC_Request),
		"NewMsgCreator":     reflect.ValueOf(gap.NewMsgCreator),

		// type definitions
		"IMsgCreator":   reflect.ValueOf((*gap.IMsgCreator)(nil)),
		"Msg":           reflect.ValueOf((*gap.Msg)(nil)),
		"MsgForward":    reflect.ValueOf((*gap.MsgForward)(nil)),
		"MsgHead":       reflect.ValueOf((*gap.MsgHead)(nil)),
		"MsgId":         reflect.ValueOf((*gap.MsgId)(nil)),
		"MsgOnewayRPC":  reflect.ValueOf((*gap.MsgOnewayRPC)(nil)),
		"MsgPacket":     reflect.ValueOf((*gap.MsgPacket)(nil)),
		"MsgRPCReply":   reflect.ValueOf((*gap.MsgRPCReply)(nil)),
		"MsgRPCRequest": reflect.ValueOf((*gap.MsgRPCRequest)(nil)),
		"MsgReader":     reflect.ValueOf((*gap.MsgReader)(nil)),
		"MsgWriter":     reflect.ValueOf((*gap.MsgWriter)(nil)),
		"SerializedMsg": reflect.ValueOf((*gap.SerializedMsg)(nil)),

		// interface wrapper definitions
		"_IMsgCreator": reflect.ValueOf((*_git_golaxy_org_framework_net_gap_IMsgCreator)(nil)),
		"_Msg":         reflect.ValueOf((*_git_golaxy_org_framework_net_gap_Msg)(nil)),
		"_MsgReader":   reflect.ValueOf((*_git_golaxy_org_framework_net_gap_MsgReader)(nil)),
		"_MsgWriter":   reflect.ValueOf((*_git_golaxy_org_framework_net_gap_MsgWriter)(nil)),
	}
}

// _git_golaxy_org_framework_net_gap_IMsgCreator is an interface wrapper for IMsgCreator type
type _git_golaxy_org_framework_net_gap_IMsgCreator struct {
	IValue     interface{}
	WDeclare   func(msg gap.Msg)
	WNew       func(msgId uint32) (gap.Msg, error)
	WUndeclare func(msgId uint32)
}

func (W _git_golaxy_org_framework_net_gap_IMsgCreator) Declare(msg gap.Msg) {
	W.WDeclare(msg)
}
func (W _git_golaxy_org_framework_net_gap_IMsgCreator) New(msgId uint32) (gap.Msg, error) {
	return W.WNew(msgId)
}
func (W _git_golaxy_org_framework_net_gap_IMsgCreator) Undeclare(msgId uint32) {
	W.WUndeclare(msgId)
}

// _git_golaxy_org_framework_net_gap_Msg is an interface wrapper for Msg type
type _git_golaxy_org_framework_net_gap_Msg struct {
	IValue interface{}
	WMsgId func() uint32
	WRead  func(p []byte) (n int, err error)
	WSize  func() int
	WWrite func(p []byte) (n int, err error)
}

func (W _git_golaxy_org_framework_net_gap_Msg) MsgId() uint32 {
	return W.WMsgId()
}
func (W _git_golaxy_org_framework_net_gap_Msg) Read(p []byte) (n int, err error) {
	return W.WRead(p)
}
func (W _git_golaxy_org_framework_net_gap_Msg) Size() int {
	return W.WSize()
}
func (W _git_golaxy_org_framework_net_gap_Msg) Write(p []byte) (n int, err error) {
	return W.WWrite(p)
}

// _git_golaxy_org_framework_net_gap_MsgReader is an interface wrapper for MsgReader type
type _git_golaxy_org_framework_net_gap_MsgReader struct {
	IValue interface{}
	WMsgId func() uint32
	WRead  func(p []byte) (n int, err error)
	WSize  func() int
}

func (W _git_golaxy_org_framework_net_gap_MsgReader) MsgId() uint32 {
	return W.WMsgId()
}
func (W _git_golaxy_org_framework_net_gap_MsgReader) Read(p []byte) (n int, err error) {
	return W.WRead(p)
}
func (W _git_golaxy_org_framework_net_gap_MsgReader) Size() int {
	return W.WSize()
}

// _git_golaxy_org_framework_net_gap_MsgWriter is an interface wrapper for MsgWriter type
type _git_golaxy_org_framework_net_gap_MsgWriter struct {
	IValue interface{}
	WMsgId func() uint32
	WSize  func() int
	WWrite func(p []byte) (n int, err error)
}

func (W _git_golaxy_org_framework_net_gap_MsgWriter) MsgId() uint32 {
	return W.WMsgId()
}
func (W _git_golaxy_org_framework_net_gap_MsgWriter) Size() int {
	return W.WSize()
}
func (W _git_golaxy_org_framework_net_gap_MsgWriter) Write(p []byte) (n int, err error) {
	return W.WWrite(p)
}