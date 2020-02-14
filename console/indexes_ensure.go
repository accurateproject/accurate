package console

import (
	"github.com/accurateproject/accurate/api/v1"
)

func init() {
	c := &CmdEnsureIndexes{
		name:      "indexes_ensure",
		rpcMethod: "ApiV1.EnsureIndexes",
		rpcParams: &v1.AttrEmpty{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdEnsureIndexes struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrEmpty
	*CommandExecuter
}

func (self *CmdEnsureIndexes) Name() string {
	return self.name
}

func (self *CmdEnsureIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdEnsureIndexes) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrEmpty{}
	}
	return self.rpcParams
}

func (self *CmdEnsureIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdEnsureIndexes) RpcResult() interface{} {
	var s string
	return &s
}

func (self *CmdEnsureIndexes) ClientArgs() (args []string) {
	return
}
