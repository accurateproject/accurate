package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdGetAccountActionPlan{
		name:      "account_actionplan_get",
		rpcMethod: "ApiV1.GetAccountActionPlan",
		rpcParams: &v1.AttrAcntAction{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccountActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrAcntAction
	*CommandExecuter
}

func (self *CmdGetAccountActionPlan) Name() string {
	return self.name
}

func (self *CmdGetAccountActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccountActionPlan) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAcntAction{}
	}
	return self.rpcParams
}

func (self *CmdGetAccountActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAccountActionPlan) RpcResult() interface{} {
	s := make([]*v1.AccountActionTiming, 0)
	return &s
}
