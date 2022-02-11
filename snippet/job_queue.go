package snippet

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"
)

// ------- Job -------
type JobQueue struct {
	mu         sync.Mutex
	jobList    *list.List
	noticeChan chan struct{}
}

func (queue *JobQueue) PushJob(job JobInterface) {
	queue.jobList.PushBack(job)
	queue.noticeChan <- struct{}{}
}

func (queue *JobQueue) PopJob() JobInterface {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	if queue.jobList.Len() == 0 {
		return nil
	}

	elements := queue.jobList.Front()
	return queue.jobList.Remove(elements).(JobInterface)
}

func (queue *JobQueue) WaitJob() <-chan struct{} {
	return queue.noticeChan
}

// ---- Job -----
type JobInterface interface {
	Execute()
	Done()
	WaitDone()
}

type BaseJob struct {
	Err        error
	done       chan struct{}
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

func (j *BaseJob) Execute() {
	// implement me please
	fmt.Println("base job is executing...")

}

func (j *BaseJob) Done() {
	close(j.done)
	if j.CancelFunc != nil {
		j.CancelFunc()
	}
}

func (j *BaseJob) WaitDone() {
	select {
	case <-j.Ctx.Done():
		return
	case <-j.done:
		return
	}
}

type SquareJob struct {
	*BaseJob
	x int8
}

func (j *SquareJob) Execute() {
	ret := j.x * j.x

	fmt.Printf("square output %+v \n", ret)
	j.Err = fmt.Errorf("fake error test")
	return
}

type WorkerManager struct {
	queue     *JobQueue
	closeChan chan struct{}
}

func (m *WorkerManager) StartWork() {
	fmt.Println("Start to Work")
	for {
		select {
		case <-m.closeChan:
			return

		case <-m.queue.noticeChan:
			job := m.queue.PopJob()
			m.ConsumeJob(job)
		}
	}
}

func (m *WorkerManager) Stop() {
	close(m.closeChan)
}

func (m *WorkerManager) ConsumeJob(job JobInterface) {
	defer func() {
		job.Done()
	}()

	job.Execute()
}

func NewWorkerManager(jq *JobQueue) *WorkerManager {
	return &WorkerManager{
		queue: jq,
	}
}

func TestJobFunc() {
	queue := &JobQueue{
		jobList:    list.New(),
		noticeChan: make(chan struct{}, 10),
	}

	workerManger := NewWorkerManager(queue)

	go workerManger.StartWork()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	job := &SquareJob{
		BaseJob: &BaseJob{
			done:       make(chan struct{}, 1),
			Ctx:        ctx,
			CancelFunc: cancel,
		},
		x: 5,
	}

	jj, _ := interface{}(job).(JobInterface)
	queue.PushJob(jj)

	job.WaitDone()
	fmt.Printf("job err %+v", job.BaseJob.Err)
}
