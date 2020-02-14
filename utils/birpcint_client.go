package utils

import (
	"github.com/accurateproject/rpcclient"
)

// Interface which the server needs to work as BiRPCServer
type BiRPCServer interface {
	Call(string, interface{}, interface{}) error // So we can use it also as rpcclient.RpcClientConnection
	CallBiRPC(rpcclient.RpcClientConnection, string, interface{}, interface{}) error
}

func NewBiRPCInternalClient(serverConn BiRPCServer) *BiRPCInternalClient {
	return &BiRPCInternalClient{serverConn: serverConn}
}

// Need separate client from the original RpcClientConnection since diretly passing the server is not enough without passing the client's reference
type BiRPCInternalClient struct {
	serverConn BiRPCServer
	clntConn   rpcclient.RpcClientConnection // conn to reach client and do calls over it
}

// Used in case when clientConn is not available at init time (eg: SMGAsterisk who needs the biRPCConn at initialization)
func (clnt *BiRPCInternalClient) SetClientConn(clntConn rpcclient.RpcClientConnection) {
	clnt.clntConn = clntConn
}

// Part of rpcclient.RpcClientConnection interface
func (clnt *BiRPCInternalClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return clnt.serverConn.CallBiRPC(clnt.clntConn, serviceMethod, args, reply)
}
