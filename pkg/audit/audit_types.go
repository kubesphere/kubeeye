package audit

import "sync"

type TaskQueue struct {
	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue []t

	processing set

	cond *sync.Cond

	shuttingDown bool
}

type empty struct{}
type t interface{}
type set map[t]empty

func (s set) has(item t) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item t) {
	s[item] = empty{}
}

func (s set) delete(item t) {
	delete(s, item)
}

func (s set) len() int {
	return len(s)
}

// Add marks item as needing processing.
func (q *TaskQueue) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}
	// add task when task is not in processing set
	if q.processing.has(item) {
		return
	}
	q.queue = append(q.queue, item)
	q.cond.Signal()
}

// Len returns the current queue length, for informational purposes only. You
// shouldn't e.g. gate a call to Add() or Get() on Len() being a particular
// value, that can't be synchronized properly.
func (q *TaskQueue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

// Get blocks until it can return an item to be processed. If shutdown = true,
// the caller should end their goroutine. You must call Done with item when you
// have finished processing it.
func (q *TaskQueue) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}

	item = q.queue[0]
	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	q.queue[0] = nil
	q.queue = q.queue[1:]

	q.processing.insert(item)
	return item, false
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *TaskQueue) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.delete(item)

	if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

// ShutDown will cause q to ignore all new items added to it and
// immediately instruct the worker goroutines to exit.
func (q *TaskQueue) ShutDown() {
	q.shutdown()
}

func (q *TaskQueue) shutdown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}

func NewTaskQueue() *TaskQueue {

	return &TaskQueue{
		processing: set{},
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}
