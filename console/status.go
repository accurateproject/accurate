package console

import "github.com/accurateproject/accurate/engine"

func init() {
	c := &CmdStatus{
		name:      "status",
		rpcMethod: "Responder.Status",
		rpcParams: &engine.AttrStatus{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdStatus struct {
	name      string
	rpcMethod string
	rpcParams *engine.AttrStatus
	*CommandExecuter
}

func (self *CmdStatus) Name() string {
	return self.name
}

func (self *CmdStatus) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatus) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.AttrStatus{}
	}
	return self.rpcParams
}

func (self *CmdStatus) PostprocessRpcParams() error {
	return nil
}

func (self *CmdStatus) RpcResult() interface{} {
	var s map[string]interface{}
	return &s
}

func (self *CmdStatus) ClientArgs() (args []string) {
	return
}
