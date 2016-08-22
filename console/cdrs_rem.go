package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdRemCdrs{
		name:      "cdrs_rem",
		rpcMethod: "ApierV1.RemCdrs",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemCdrs struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrRemCdrs
	*CommandExecuter
}

func (self *CmdRemCdrs) Name() string {
	return self.name
}

func (self *CmdRemCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemCdrs) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrRemCdrs{}
	}
	return self.rpcParams
}

func (self *CmdRemCdrs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemCdrs) RpcResult() interface{} {
	var s string
	return &s
}
