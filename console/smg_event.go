package console

import (
	"strings"

	"github.com/accurateproject/accurate/sessionmanager"
)

type AttrSmgEvent struct {
	Method string // shoul be ignored after RPC call
	sessionmanager.SMGenericEvent
}

func init() {
	c := &CmdSmgEvent{
		name: "smg_event",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSmgEvent struct {
	name      string
	rpcMethod string
	rpcParams interface{}
	*CommandExecuter
}

func (self *CmdSmgEvent) Name() string {
	return self.name
}

func (self *CmdSmgEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSmgEvent) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &AttrSmgEvent{}
	}
	return self.rpcParams
}

func (self *CmdSmgEvent) PostprocessRpcParams() error {
	param := self.rpcParams.(*AttrSmgEvent)
	self.rpcMethod = "SMGenericV1." + param.Method
	self.rpcParams = param.SMGenericEvent
	return nil
}

func (self *CmdSmgEvent) RpcResult() interface{} {
	methodElems := strings.Split(self.rpcMethod, ".")
	if len(methodElems) != 2 {
		return nil
	}
	switch methodElems[1] {
	case "SessionEnd", "ChargeEvent", "ProcessCdr":
		var s string
		return &s
	case "SessionStart", "SessionUpdate", "GetMaxUsage":
		var f float64
		return &f
	case "GetLcrSuppliers":
		ss := make([]string, 0)
		return ss
	}
	return nil
}
