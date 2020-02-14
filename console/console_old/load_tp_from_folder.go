package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &LoadTpFromFolder{
		name:      "load_tp_from_folder",
		rpcMethod: "ApiV1.LoadTariffPlanFromFolder",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type LoadTpFromFolder struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrLoadTpFromFolder
	rpcResult string
	*CommandExecuter
}

func (self *LoadTpFromFolder) Name() string {
	return self.name
}

func (self *LoadTpFromFolder) RpcMethod() string {
	return self.rpcMethod
}

func (self *LoadTpFromFolder) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrLoadTpFromFolder{}
	}
	return self.rpcParams
}

func (self *LoadTpFromFolder) PostprocessRpcParams() error {
	return nil
}

func (self *LoadTpFromFolder) RpcResult() interface{} {
	var s string
	return &s
}
