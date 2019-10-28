package engine

import (
	"sync"
	"time"

	"github.com/phf/go-queue/queue"
)

type OperationHandler func(interface{})

type Operation struct {
	Handler OperationHandler
	Context interface{}
}

type OperationQueue struct {
	lock    sync.Mutex
	wg      sync.WaitGroup
	opQueue *queue.Queue
	notify  chan bool
	timer   *time.Timer
}

func NewOperationQueue() *OperationQueue {
	queue := &OperationQueue{
		lock:    sync.Mutex{},
		wg:      sync.WaitGroup{},
		opQueue: queue.New(),
		notify:  make(chan bool, 1000),
	}

	queue.wg.Add(1)
	go queue.waitForNotification()
	return queue
}

func (ctx *OperationQueue) Enqueue(operation *Operation) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ctx.opQueue.PushBack(operation)
	ctx.notify <- true
}

func (ctx *OperationQueue) Stop() {
	close(ctx.notify)
	ctx.wg.Wait()
}

func (ctx *OperationQueue) waitForNotification() {
loop:
	for {
		select {
		case <-ctx.notify:

		default:
			ctx.wg.Done()
			break loop
		}
	}
}
