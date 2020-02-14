package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdExportCdrs{
		name:      "cdrs_export",
		rpcMethod: "ApiV2.ExportCdrsToFile",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExportCdrs struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrExportCdrsToFile
	*CommandExecuter
}

func (self *CmdExportCdrs) Name() string {
	return self.name
}

func (self *CmdExportCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExportCdrs) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrExportCdrsToFile{}
	}
	return self.rpcParams
}

func (self *CmdExportCdrs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExportCdrs) RpcResult() interface{} {
	return &utils.AttrExportCdrsToFile{}
}
