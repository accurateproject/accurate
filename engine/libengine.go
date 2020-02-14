package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

func NewRPCPool(dispatchStrategy string, connAttempts, reconnects int, connectTimeout, replyTimeout time.Duration,
	rpcConnCfgs []*config.HaPool, internalConnChan chan rpcclient.RpcClientConnection, ttl time.Duration) (*rpcclient.RpcClientPool, error) {
	var rpcClient *rpcclient.RpcClient
	var err error
	rpcPool := rpcclient.NewRpcClientPool(dispatchStrategy, replyTimeout)
	atLestOneConnected := false // If one connected we don't longer return errors
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.MetaInternal {
			var internalConn rpcclient.RpcClientConnection
			select {
			case internalConn = <-internalConnChan:
				internalConnChan <- internalConn
			case <-time.After(ttl):
				return nil, errors.New("TTL triggered")
			}
			rpcClient, err = rpcclient.NewRpcClient("", "", connAttempts, reconnects, connectTimeout, replyTimeout, rpcclient.INTERNAL_RPC, internalConn, false)
		} else if utils.IsSliceMember([]string{utils.MetaJSONrpc, utils.MetaGOBrpc, ""}, rpcConnCfg.Transport) {
			codec := utils.GOB
			if rpcConnCfg.Transport != "" {
				codec = rpcConnCfg.Transport[1:] // Transport contains always * before codec understood by rpcclient
			}
			rpcClient, err = rpcclient.NewRpcClient("tcp", rpcConnCfg.Address, connAttempts, reconnects, connectTimeout, replyTimeout, codec, nil, false)
		} else {
			return nil, fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
		}
		if err == nil {
			atLestOneConnected = true
		}
		rpcPool.AddClient(rpcClient)
	}
	if atLestOneConnected {
		err = nil
	}
	return rpcPool, err
}
