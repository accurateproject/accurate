package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdRateCdrs{
		name:      "cdrs_rate",
		rpcMethod: "CdrsV1.RateCDRs",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRateCdrs struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrRateCdrs
	*CommandExecuter
}

func (self *CmdRateCdrs) Name() string {
	return self.name
}

func (self *CmdRateCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRateCdrs) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrRateCdrs{}
	}
	return self.rpcParams
}

func (self *CmdRateCdrs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRateCdrs) RpcResult() interface{} {
	var s string
	return &s
}
