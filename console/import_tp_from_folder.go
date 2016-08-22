
package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &ImportTpFromFolder{
		name:      "import_tp_from_folder",
		rpcMethod: "ApierV1.ImportTariffPlanFromFolder",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type ImportTpFromFolder struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrImportTPFromFolder
	rpcResult string
	*CommandExecuter
}

func (self *ImportTpFromFolder) Name() string {
	return self.name
}

func (self *ImportTpFromFolder) RpcMethod() string {
	return self.rpcMethod
}

func (self *ImportTpFromFolder) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrImportTPFromFolder{}
	}
	return self.rpcParams
}

func (self *ImportTpFromFolder) PostprocessRpcParams() error {
	return nil
}

func (self *ImportTpFromFolder) RpcResult() interface{} {
	var s string
	return &s
}
