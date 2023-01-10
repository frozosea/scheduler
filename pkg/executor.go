package pkg

import (
	"context"
	"log"
	"sync"
	"time"
)

type IJobExecutor interface {
	Run(job *Job)
	Remove(taskId string) error
}
type Executor struct {
	wg            *sync.WaitGroup
	cancellations map[string]context.CancelFunc
	jobStore      IJobStore
	logger        *log.Logger
	timeParser    ITimeParser
}

func (e *Executor) checkJobExist(id string) bool {
	return e.cancellations[id] != nil
}
func (e *Executor) Run(job *Job) {
	ctx, cancel := context.WithCancel(job.Ctx)
	job.Ctx = ctx
	e.cancellations[job.Id] = cancel
	e.wg.Add(1)
	e.process(job)
}
func (e *Executor) process(job *Job) {
	ticker := time.NewTicker(job.Interval)
	for {
		select {
		case <-ticker.C:
			e.logger.Printf(`job with id: %s now run`, job.Id)
			job.Fn(job.Ctx)
			e.logger.Printf(`task with id: %s was completed`, job.Id)
			nextInterval, err := e.timeParser.Parse(job.Time)
			if err != nil {
				continue
			}
			job.NextRunTime = time.Now().Add(nextInterval)
			continue
		case <-job.Ctx.Done():
			e.logger.Printf(`job with id: %s ctx done time: %s`, job.Id, job.Time)
			e.wg.Done()
			ticker.Stop()
			if err := e.Remove(job.Id); err != nil {
				e.logger.Printf(`remove job with id: %s err: %s`, job.Id, err.Error())
				return
			}
			if err := e.jobStore.Remove(job.Ctx, job.Id); err != nil {
				e.logger.Printf(`remove job with id: %s err: %s`, job.Id, err.Error())
				return
			}
			e.logger.Printf(`job with id %s was removed`, job.Id)
			return
		default:
			continue
		}
	}
}
func (e *Executor) Remove(taskId string) error {
	for jobId, cancel := range e.cancellations {
		if jobId == taskId {
			cancel()
			delete(e.cancellations, taskId)
			return nil
		}
	}
	return &LookupJobError{}
}
func NewExecutor(jobStore IJobStore, timeParser ITimeParser, logger *log.Logger) *Executor {
	return &Executor{
		wg:            &sync.WaitGroup{},
		cancellations: make(map[string]context.CancelFunc),
		jobStore:      jobStore,
		logger:        logger,
		timeParser:    timeParser,
	}
}
