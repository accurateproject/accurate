package console

import (
	"github.com/accurateproject/accurate/apier/v1"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdRemRatingProfile{
		name:      "ratingprofile_rem",
		rpcMethod: "ApiV1.RemoveRatingProfile",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdRemRatingProfile) Name() string {
	return self.name
}

func (self *CmdRemRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemRatingProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveRatingProfile{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdRemRatingProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemRatingProfile) RpcResult() interface{} {
	var s string
	return &s
}
