package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

type Scheduler struct {
	queue       engine.ActionTimingPriorityList
	timer       *time.Timer
	restartLoop chan bool
	sync.Mutex
	storage          engine.RatingStorage
	schedulerStarted bool
}

func NewScheduler(storage engine.RatingStorage) *Scheduler {
	return &Scheduler{
		restartLoop: make(chan bool),
		storage:     storage,
	}
}

func (s *Scheduler) Loop() {
	s.schedulerStarted = true
	for {
		for len(s.queue) == 0 { //hang here if empty
			<-s.restartLoop
		}
		utils.Logger.Info(fmt.Sprintf("<Scheduler> Scheduler queue length: %v", len(s.queue)))
		s.Lock()
		a0 := s.queue[0]
		utils.Logger.Info(fmt.Sprintf("<Scheduler> Action: %s", a0.ActionsID))
		now := time.Now()
		start := a0.GetNextStartTime(now)
		if start.Equal(now) || start.Before(now) {
			go a0.Execute()
			// if after execute the next start time is in the past then
			// do not add it to the queue
			a0.ResetStartTimeCache()
			now = time.Now().Add(time.Second)
			start = a0.GetNextStartTime(now)
			if start.Before(now) {
				s.queue = s.queue[1:]
			} else {
				s.queue = append(s.queue, a0)
				s.queue = s.queue[1:]
				sort.Sort(s.queue)
			}
			s.Unlock()
		} else {
			s.Unlock()
			d := a0.GetNextStartTime(now).Sub(now)
			utils.Logger.Info(fmt.Sprintf("<Scheduler> Time to next action (%s): %v", a0.ActionsID, d))
			s.timer = time.NewTimer(d)
			select {
			case <-s.timer.C:
				// timer has expired
				utils.Logger.Info(fmt.Sprintf("<Scheduler> Time for action on %s", a0.ActionsID))
			case <-s.restartLoop:
				// nothing to do, just continue the loop
			}
		}
	}
}

func (s *Scheduler) Reload(protect bool) {
	s.loadActionPlans()
	s.restart()
}

func (s *Scheduler) loadActionPlans() {
	s.Lock()
	defer s.Unlock()
	// limit the number of concurrent tasks
	limit := make(chan bool, 10)
	// execute existing tasks
	for {
		task, err := s.storage.PopTask()
		if err != nil || task == nil {
			break
		}
		limit <- true
		go func() {
			utils.Logger.Info(fmt.Sprintf("<Scheduler> executing task %s on account %s", task.ActionsID, task.AccountID))
			task.Execute()
			<-limit
		}()
	}

	aplIter := s.storage.Iterator(engine.ColApl, "", nil)

	// recreate the queue
	s.queue = engine.ActionTimingPriorityList{}
	var actionPlan engine.ActionPlan
	for aplIter.Next(&actionPlan) {
		actionPlan.SetParentActionPlan()
		for _, at := range actionPlan.ActionTimings {
			if at.Timing == nil {
				utils.Logger.Warn("<Scheduler> Nil timing on action plan, discarding!", zap.Any("AT", at))
				continue
			}
			if at.IsASAP() {
				continue
			}
			now := time.Now()
			if at.GetNextStartTime(now).Before(now) {
				// the task is obsolete, do not add it to the queue
				continue
			}
			s.queue = append(s.queue, at)

		}
	}
	if err := aplIter.Close(); err != nil {
		utils.Logger.Warn("<Scheduler> Cannot get action plans", zap.Error(err))
	}
	sort.Sort(s.queue)
	utils.Logger.Info(fmt.Sprintf("<Scheduler> queued %d action plans", len(s.queue)))
}

func (s *Scheduler) restart() {
	if s.schedulerStarted {
		s.restartLoop <- true
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

func (s *Scheduler) GetQueue() engine.ActionTimingPriorityList {
	return s.queue
}
