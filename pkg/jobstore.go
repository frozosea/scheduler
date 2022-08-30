package scheduler

import (
	"context"
	"time"
)

type AddJobError struct{}

func (a *AddJobError) Error() string {
	return "job with this id already exists."
}

type JobAlreadyExistsError struct{}

func (g *JobAlreadyExistsError) Error() string {
	return "job with this id already exists."
}

type LookupJobError struct{}

func (l *LookupJobError) Error() string {
	return "cannot find job with this id"
}

type IJobStore interface {
	Save(ctx context.Context, taskId string, task ITask, interval time.Duration, args []interface{}, time string) (*Job, error)
	Get(ctx context.Context, taskId string) (*Job, error)
	GetAll(ctx context.Context) ([]*Job, error)
	Reschedule(ctx context.Context, taskId string, interval time.Duration, newStrTime string) (*Job, error)
	Remove(ctx context.Context, taskId string) error
	RemoveAll(ctx context.Context) error
}
type MemoryJobStore struct {
	jobs map[string]*Job
}

func (m *MemoryJobStore) Save(ctx context.Context, taskId string, task ITask, interval time.Duration, args []interface{}, strTime string) (*Job, error) {
	getJob := m.jobs[taskId]
	if getJob != nil {
		return getJob, &AddJobError{}
	}
	job := &Job{
		Id:          taskId,
		Fn:          task,
		NextRunTime: time.Now().Add(interval),
		Args:        args,
		Interval:    interval,
		Ctx:         ctx,
		Time:        strTime,
	}
	m.jobs[taskId] = job
	return job, nil
}
func (m *MemoryJobStore) Get(_ context.Context, taskId string) (*Job, error) {
	job := m.jobs[taskId]
	if job == nil {
		return new(Job), &LookupJobError{}
	}
	return job, nil
}
func (m *MemoryJobStore) GetAll(_ context.Context) ([]*Job, error) {
	var jobs []*Job
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}
func (m *MemoryJobStore) checkTask(taskId string) (*Job, error) {
	job := m.jobs[taskId]
	if job == nil {
		return nil, &JobAlreadyExistsError{}
	}
	return job, nil
}
func (m *MemoryJobStore) Reschedule(ctx context.Context, taskId string, interval time.Duration, newStrTime string) (*Job, error) {
	job := m.jobs[taskId]
	if job == nil {
		return new(Job), &LookupJobError{}
	}
	modifiedJob := &Job{
		Id:          taskId,
		Fn:          job.Fn,
		NextRunTime: time.Now().Add(interval),
		Args:        job.Args,
		Interval:    interval,
		Ctx:         ctx,
		Time:        newStrTime,
	}
	m.jobs[taskId] = modifiedJob
	return modifiedJob, nil
}
func (m *MemoryJobStore) Remove(_ context.Context, taskId string) error {
	job := m.jobs[taskId]
	if job == nil {
		return &LookupJobError{}
	}
	delete(m.jobs, taskId)
	return nil
}
func (m *MemoryJobStore) RemoveAll(_ context.Context) error {
	for k := range m.jobs {
		delete(m.jobs, k)
	}
	return nil
}
func NewMemoryJobStore() *MemoryJobStore {
	return &MemoryJobStore{jobs: make(map[string]*Job)}
}
