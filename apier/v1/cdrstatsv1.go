package v1

import (
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/cgrates/rpcclient"
)

// Interact with Stats server
type CDRStatsV1 struct {
	CdrStats rpcclient.RpcClientConnection
}

type AttrGetMetrics struct {
	StatsQueueId string // Id of the stats instance queried
}

func (sts *CDRStatsV1) GetMetrics(attr AttrGetMetrics, reply *map[string]float64) error {
	if len(attr.StatsQueueId) == 0 {
		return fmt.Errorf("%s:StatsQueueId", utils.ErrMandatoryIeMissing.Error())
	}
	return sts.CdrStats.Call("CDRStatsV1.GetValues", attr.StatsQueueId, reply)
}

func (sts *CDRStatsV1) GetQueueIds(empty string, reply *[]string) error {
	return sts.CdrStats.Call("CDRStatsV1.GetQueueIds", 0, reply)
}

func (sts *CDRStatsV1) GetQueue(id string, sq *engine.StatsQueue) error {
	return sts.CdrStats.Call("CDRStatsV1.GetQueue", id, sq)
}

func (sts *CDRStatsV1) AddQueue(cs *engine.CdrStats, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.AddQueue", cs, reply)
}

func (sts *CDRStatsV1) RemoveQueue(qID string, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.RemoveQueue", qID, reply)
}

func (sts *CDRStatsV1) GetQueueTriggers(id string, ats *engine.ActionTriggers) error {
	return sts.CdrStats.Call("CDRStatsV1.GetQueueTriggers", id, ats)
}

func (sts *CDRStatsV1) ReloadQueues(attr utils.AttrCDRStatsReloadQueues, reply *string) error {
	var out int
	if err := sts.CdrStats.Call("CDRStatsV1.ReloadQueues", attr.StatsQueueIds, &out); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (sts *CDRStatsV1) ResetQueues(attr utils.AttrCDRStatsReloadQueues, reply *string) error {
	var out int
	if err := sts.CdrStats.Call("CDRStatsV1.ResetQueues", attr.StatsQueueIds, &out); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}
