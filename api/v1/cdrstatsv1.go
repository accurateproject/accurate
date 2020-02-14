package v1

import (
	"strings"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

// Interact with Stats server
type CDRStatsV1 struct {
	CdrStats rpcclient.RpcClientConnection
}

func (sts *CDRStatsV1) GetMetrics(attr utils.AttrStatsQueueID, reply *map[string]*dec.Dec) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	return sts.CdrStats.Call("CDRStatsV1.GetMetrics", attr, reply)
}

type AttrTenant struct {
	Tenant string
}

func (sts *CDRStatsV1) GetQueueIDs(attr AttrTenant, reply *[]string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	return sts.CdrStats.Call("CDRStatsV1.GetQueueIDs", attr.Tenant, reply)
}

func (sts *CDRStatsV1) GetQueue(attr utils.AttrStatsQueueID, sq *engine.StatsQueue) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	return sts.CdrStats.Call("CDRStatsV1.GetQueue", attr, sq)
}

func (sts *CDRStatsV1) AddQueue(cs *engine.CdrStats, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.AddQueue", cs, reply)
}

func (sts *CDRStatsV1) DisableQueue(attr utils.AttrStatsQueueDisable, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.DisableQueue", attr, reply)
}

func (sts *CDRStatsV1) RemoveQueue(attr utils.AttrStatsQueueIDs, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.RemoveQueue", attr, reply)
}

func (sts *CDRStatsV1) GetQueueTriggers(attr utils.AttrStatsQueueID, ats *engine.ActionTriggers) error {
	return sts.CdrStats.Call("CDRStatsV1.GetQueueTriggers", attr, ats)
}

func (sts *CDRStatsV1) ReloadQueues(attr utils.AttrStatsQueueIDs, reply *string) error {
	var out int
	if err := sts.CdrStats.Call("CDRStatsV1.ReloadQueues", attr, &out); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (sts *CDRStatsV1) ResetQueues(attr utils.AttrStatsQueueIDs, reply *string) error {
	var out int
	if err := sts.CdrStats.Call("CDRStatsV1.ResetQueues", attr, &out); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (sts *CDRStatsV1) AppendCDR(cdr *engine.CDR, reply *int) error {
	return sts.CdrStats.Call("CDRStatsV1.AppendCDR", cdr, reply)
}
