package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type ShouldBeCancelled bool
type ITask func(ctx context.Context)

type Job struct {
	Id          string
	Fn          ITask
	NextRunTime time.Time
	Interval    time.Duration
	Time        string
	Ctx         context.Context
}

type Schedule struct {
	Next time.Ticker
}
type Manager struct {
	executor   IJobExecutor
	jobstore   IJobStore
	timeParser ITimeParser
	baseLogger log.Logger
}

func (m *Manager) Add(ctx context.Context, taskId string, task ITask, timeStr string) (*Job, error) {
	taskTime, err := m.timeParser.Parse(timeStr)
	if err != nil {
		return &Job{}, err
	}
	job, err := m.jobstore.Save(ctx, taskId, task, taskTime, timeStr)
	if err != nil {
		m.baseLogger.Println(fmt.Sprintf(`add task with id: %s err: %s`, taskId, err.Error()))
		return &Job{}, err
	}
	m.baseLogger.Printf("job with id %s and time %s was add next run time is %s", job.Id, timeStr, job.NextRunTime.Format("2006-01-02 15:04"))
	go m.executor.Run(job)
	return job, nil
}
func (m *Manager) AddWithDuration(ctx context.Context, taskId string, task ITask, interval time.Duration) (*Job, error) {
	job, err := m.jobstore.Save(ctx, taskId, task, interval, fmt.Sprintf(`%d:%d`, time.Now().Add(interval).Hour(), time.Now().Add(interval).Minute()))
	if err != nil {
		m.baseLogger.Println(fmt.Sprintf(`add task with id: %s err: %s`, taskId, err.Error()))
		return job, err
	}
	go m.executor.Run(job)
	return job, err
}
func (m *Manager) Get(ctx context.Context, taskId string) (*Job, error) {
	return m.jobstore.Get(ctx, taskId)
}
func (m *Manager) Reschedule(ctx context.Context, taskId string, timeStr string) (*Job, error) {
	job, err := m.jobstore.Get(ctx, taskId)
	if err != nil {
		m.baseLogger.Printf(`get job with id: %s err: %s`, job.Id, err.Error())
		return nil, err
	}
	newInterval, err := m.timeParser.Parse(timeStr)
	if err != nil {
		return nil, err
	}
	newJob, err := m.jobstore.Reschedule(ctx, taskId, newInterval, timeStr)
	if err != nil {
		return newJob, err
	}
	if err := m.executor.Remove(taskId); err != nil {
		return newJob, err
	}
	newJob.Ctx = context.Background()
	go m.executor.Run(newJob)
	return newJob, nil
}
func (m *Manager) RescheduleWithDuration(ctx context.Context, taskId string, newInterval time.Duration) (*Job, error) {
	job, err := m.jobstore.Get(ctx, taskId)
	if err != nil {
		m.baseLogger.Printf(`get job with id: %s err: %s`, job.Id, err.Error())
		return nil, err
	}
	newJob, err := m.jobstore.Reschedule(ctx, taskId, newInterval, fmt.Sprintf(`%d:%d`, time.Now().Add(newInterval).Hour(), time.Now().Add(newInterval).Minute()))
	if err != nil {
		return newJob, err
	}
	if err := m.executor.Remove(taskId); err != nil {
		return newJob, err
	}
	newJob.Ctx = context.Background()
	go m.executor.Run(newJob)
	return newJob, nil
}
func (m *Manager) Remove(ctx context.Context, taskId string) error {
	if err := m.executor.Remove(taskId); err != nil {
		return err
	}
	return m.jobstore.Remove(ctx, taskId)
}
func (m *Manager) RemoveAll(ctx context.Context) error {
	return m.jobstore.RemoveAll(ctx)
}
func (m *Manager) Modify(ctx context.Context, taskId string, task ITask) error {
	job, err := m.jobstore.Get(ctx, taskId)
	if err != nil {
		return err
	}
	job.Fn = task
	if err := m.jobstore.Remove(ctx, taskId); err != nil {
		return err
	}
	if err := m.executor.Remove(taskId); err != nil {
		return err
	}
	newJob, err := m.jobstore.Save(job.Ctx, job.Id, job.Fn, job.Interval, job.Time)
	if err != nil {
		return err
	}
	newJob.Ctx = context.Background()
	go m.executor.Run(newJob)
	return nil
}

func NewDefault(timezone string) *Manager {
	jobStore := NewMemoryJobStore()
	return &Manager{executor: NewExecutor(jobStore, NewTimeParser(timezone), log.New(os.Stdout, "log", 1)), jobstore: jobStore, baseLogger: *log.New(os.Stdout, "log", 1), timeParser: NewTimeParser(timezone)}
}
