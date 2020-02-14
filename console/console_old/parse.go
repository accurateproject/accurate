package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdParse{
		name:      "parse",
		rpcParams: &AttrParse{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type AttrParse struct {
	Expression string
	Value      string
}

type CmdParse struct {
	name      string
	rpcMethod string
	rpcParams *AttrParse
	*CommandExecuter
}

func (self *CmdParse) Name() string {
	return self.name
}

func (self *CmdParse) RpcMethod() string {
	return ""
}

func (self *CmdParse) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &AttrParse{}
	}
	return self.rpcParams
}

func (self *CmdParse) RpcResult() interface{} {
	return nil
}

func (self *CmdParse) PostprocessRpcParams() error {
	return nil
}

func (self *CmdParse) LocalExecute() string {
	if self.rpcParams.Expression == "" {
		return "Empty expression error"
	}
	if self.rpcParams.Value == "" {
		return "Empty value error"
	}
	if rsrField, err := utils.NewRSRField(self.rpcParams.Expression); err == nil {
		return rsrField.ParseValue(self.rpcParams.Value)
	} else {
		return err.Error()
	}
}
