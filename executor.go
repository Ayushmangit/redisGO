package main

import (
	"errors"
	"sync"

	"github.com/Ayushmangit/redisGO/store"
)

var ErrExecutorStopped = errors.New("executor stopped")

type commandRequest struct {
	command Command
	result  chan commandResult
}

type commandResult struct {
	value string
	err   error
}

type Executor struct {
	store    store.Storer
	commands chan commandRequest
	done     chan struct{}

	mu     sync.Mutex
	closed bool
}

func NewExecutor(s store.Storer, queueSize int) *Executor {
	if queueSize < 0 {
		panic("queue size cannot be negative")
	}
	e := &Executor{
		store:    s,
		commands: make(chan commandRequest, queueSize),
		done:     make(chan struct{}),
	}
	go e.run()
	return e
}

func (e *Executor) run() {
	defer close(e.done)

	for req := range e.commands {
		value, err := dispatch(e.store, req.command)
		req.result <- commandResult{value: value, err: err}
	}
}

func (e *Executor) Execute(c Command) (string, error) {
	result := make(chan commandResult, 1)
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return "", ErrExecutorStopped
	}
	e.commands <- commandRequest{
		command: c,
		result:  result,
	}
	e.mu.Unlock()
	res := <-result
	return res.value, res.err
}

func (e *Executor) Stop() {
	e.mu.Lock()

	if !e.closed {
		e.closed = true
		close(e.commands)
	}
	e.mu.Unlock()
	<-e.done
}
