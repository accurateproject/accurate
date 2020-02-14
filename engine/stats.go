package engine

import (
	"reflect"
	"strings"
	"sync"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

type StatsInterface interface {
	GetValues(utils.AttrStatsQueueID, *map[string]float64) error
	GetQueueIDs(string, *[]string) error
	GetQueue(utils.AttrStatsQueueID, *StatsQueue) error
	GetQueueTriggers(string, *ActionTriggers) error
	AppendCDR(*CDR, *int) error
	AddQueue(*CdrStats, *int) error
	DisableQueue(utils.AttrStatsQueueDisable, *int) error
	RemoveQueue(utils.AttrStatsQueueIDs, *int) error
	ReloadQueues(utils.AttrStatsQueueIDs, *int) error
	ResetQueues(utils.AttrStatsQueueIDs, *int) error
	Stop(int, *int) error
}

type Stats struct {
	queues       map[string]*StatsQueue
	mux          sync.RWMutex
	ratingDB     RatingStorage
	accountingDB AccountingStorage
	cdrDB        CdrStorage
}

func NewStats(ratingDB RatingStorage, accountingDB AccountingStorage, cdrDB CdrStorage) *Stats {
	cdrStats := &Stats{ratingDB: ratingDB, accountingDB: accountingDB, cdrDB: cdrDB}

	if err := cdrStats.UpdateQueues(""); err != nil {
		utils.Logger.Error("cannot load cdr stats", zap.Error(err))
	}
	return cdrStats
}

func (s *Stats) GetQueueIDs(tenant string, ids *[]string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	result := make([]string, 0)
	for id, _ := range s.queues {
		parts := strings.SplitN(id, utils.CONCATENATED_KEY_SEP, 2)
		if parts[0] == tenant {
			result = append(result, id)
		}
	}
	*ids = result
	return nil
}

func (s *Stats) GetQueue(attr utils.AttrStatsQueueID, sq *StatsQueue) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	q, found := s.queues[utils.ConcatKey(attr.Tenant, attr.ID)]
	if !found {
		return utils.ErrNotFound
	}
	*sq = *q
	return nil
}

func (s *Stats) GetQueueTriggers(attr utils.AttrStatsQueueID, ats *ActionTriggers) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	q, found := s.queues[utils.ConcatKey(attr.Tenant, attr.ID)]
	if !found {
		return utils.ErrNotFound
	}
	if q.getTriggers() != nil {
		*ats = q.getTriggers()
	} else {
		*ats = ActionTriggers{}
	}
	return nil
}

func (s *Stats) GetMetrics(attr utils.AttrStatsQueueID, values *map[string]*dec.Dec) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if sq, ok := s.queues[utils.ConcatKey(attr.Tenant, attr.ID)]; ok {
		*values = sq.getMetricValues()
		return nil
	}
	return utils.ErrNotFound
}

func (s *Stats) AddQueue(cs *CdrStats, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}
	var sq *StatsQueue
	var exists bool
	id := utils.ConcatKey(cs.Tenant, cs.Name)
	if sq, exists = s.queues[id]; exists {
		sq.updateConf(cs)
	} else {
		sq = NewStatsQueue(cs, s.accountingDB, s.cdrDB)
		s.queues[id] = sq
	}
	// save the conf
	if err := s.ratingDB.SetCdrStats(cs); err != nil {
		return err
	}
	return nil
}

func (s *Stats) RemoveQueue(attr utils.AttrStatsQueueIDs, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}

	for _, id := range attr.IDs {
		qID := utils.ConcatKey(attr.Tenant, id)
		delete(s.queues, qID)
		// remove conf
		if err := s.ratingDB.RemoveCdrStats(attr.Tenant, id); err != nil {
			return err
		}
		// remove stats queue
		if err := s.accountingDB.RemoveCdrStatsQueue(attr.Tenant, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stats) DisableQueue(attr utils.AttrStatsQueueDisable, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}
	// get the conf
	cs, err := s.ratingDB.GetCdrStats(attr.Tenant, attr.ID)
	if err != nil {
		return err
	}
	cs.Disabled = attr.Disable
	var sq *StatsQueue
	var exists bool
	id := utils.ConcatKey(cs.Tenant, cs.Name)
	if sq, exists = s.queues[id]; exists {
		sq.updateConf(cs)
	} else {
		sq = NewStatsQueue(cs, s.accountingDB, s.cdrDB)
		s.queues[id] = sq
	}
	// save the conf
	if err := s.ratingDB.SetCdrStats(cs); err != nil {
		return err
	}
	return nil
}

func (s *Stats) ReloadQueues(attr utils.AttrStatsQueueIDs, out *int) error {
	if len(attr.IDs) == 0 {
		return s.UpdateQueues(attr.Tenant)

	}
	for _, id := range attr.IDs {
		if cs, err := s.ratingDB.GetCdrStats(attr.Tenant, id); err == nil {
			if err := s.AddQueue(cs, nil); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *Stats) ResetQueues(attr utils.AttrStatsQueueIDs, out *int) error {
	if len(attr.IDs) == 0 {
		for _, sq := range s.queues {
			if err := sq.resetQCDRs(); err != nil {
				utils.Logger.Error("error reseting stats queue items: ", zap.Error(err))
				return err
			}
			sq.resetMetrics()
		}
	} else {
		for _, id := range attr.IDs {
			id = utils.ConcatKey(attr.Tenant, id)
			sq, exists := s.queues[id]
			if !exists {
				utils.Logger.Warn("cannot reset queue, Not Found", zap.String("id", id))
				continue
			}
			if err := sq.resetQCDRs(); err != nil {
				utils.Logger.Error("error reseting stats queue items: ", zap.Error(err))
				return err
			}
			sq.resetMetrics()
		}
	}
	return nil
}

// change the existing ones
// add new ones
// delete the ones missing from the new list
func (s *Stats) UpdateQueues(tenant string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	oldQueues := s.queues
	s.queues = make(map[string]*StatsQueue)
	csIter := s.ratingDB.Iterator(ColCrs, "", map[string]interface{}{"tenant": tenant})
	cs := &CdrStats{}
	for csIter.Next(cs) {
		var sq *StatsQueue
		var existing bool
		if oldQueues != nil {
			id := utils.ConcatKey(cs.Tenant, cs.Name)
			if sq, existing = oldQueues[id]; existing {
				sq.updateConf(cs)
			}
		}
		if sq == nil {
			sq = NewStatsQueue(cs, s.accountingDB, s.cdrDB)
			if err := sq.load(); err != nil {
				utils.Logger.Error("error restoring stats queue items:", zap.Error(err))
				return err
			}
		}
		s.queues[utils.ConcatKey(cs.Tenant, cs.Name)] = sq
		cs = &CdrStats{}
	}
	if err := csIter.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Stats) AppendCDR(cdr *CDR, out *int) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, sq := range s.queues {
		sq.appendCDR(cdr)
	}
	return nil
}

func (s *Stats) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(s).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}

	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
