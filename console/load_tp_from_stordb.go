package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &LoadTpFromStorDb{
		name:      "load_tp_from_stordb",
		rpcMethod: "ApierV1.LoadTariffPlanFromStorDb",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type LoadTpFromStorDb struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrLoadTpFromStorDb
	rpcResult string
	*CommandExecuter
}

func (self *LoadTpFromStorDb) Name() string {
	return self.name
}

func (self *LoadTpFromStorDb) RpcMethod() string {
	return self.rpcMethod
}

func (self *LoadTpFromStorDb) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrLoadTpFromStorDb{}
	}
	return self.rpcParams
}

func (self *LoadTpFromStorDb) PostprocessRpcParams() error {
	return nil
}

func (self *LoadTpFromStorDb) RpcResult() interface{} {
	var s string
	return &s
}
