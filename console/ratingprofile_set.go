
package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdSetRatingProfile{
		name:      "ratingprofile_set",
		rpcMethod: "ApierV1.SetRatingProfile",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TPRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdSetRatingProfile) Name() string {
	return self.name
}

func (self *CmdSetRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetRatingProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TPRatingProfile{}
	}
	return self.rpcParams
}

func (self *CmdSetRatingProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetRatingProfile) RpcResult() interface{} {
	var s string
	return &s
}
