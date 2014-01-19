// Copyright (c) 2014 The go-exchange AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package exchange

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/tchap/go-patricia/patricia"
)

//------------------------------------------------------------------------------
// Exchange
//------------------------------------------------------------------------------

type (
	Event        interface{}
	EventHandler func(Topic, Event)
	Handle       uint
	Topic        []byte
)

const (
	stateRunning = iota
	stateTerminating
	stateTerminated
)

type Exchange struct {
	state              int
	trie               *patricia.Trie
	topicForHandle     map[Handle]Topic
	numRunningHandlers int32
	nextHandle         Handle
	mu                 *sync.RWMutex
	cond               *sync.Cond
}

func New() *Exchange {
	mu := new(sync.RWMutex)
	return &Exchange{
		trie:           patricia.NewTrie(),
		topicForHandle: make(map[Handle]Topic),
		mu:             mu,
		cond:           sync.NewCond(mu),
	}
}

type handlerRecord struct {
	handle  Handle
	handler EventHandler
}

// Public API ------------------------------------------------------------------

func (exchange *Exchange) Subscribe(topicPrefix Topic, handler EventHandler) (Handle, error) {
	exchange.cond.L.Lock()
	defer exchange.cond.L.Unlock()

	if exchange.state != stateRunning {
		return 0, ErrInvalidState
	}

	handle, err := exchange.getHandle(topicPrefix)
	if err != nil {
		return 0, err
	}

	record := &handlerRecord{
		handle:  handle,
		handler: handler,
	}

	if item := exchange.trie.Get(patricia.Prefix(topicPrefix)); item == nil {
		records := []*handlerRecord{record}
		exchange.trie.Insert(patricia.Prefix(topicPrefix), &records)
	} else {
		records := item.(*[]*handlerRecord)
		*records = append(*records, record)
	}

	return record.handle, nil
}

func (exchange *Exchange) Unsubscribe(handle Handle) error {
	exchange.cond.L.Lock()
	defer exchange.cond.L.Unlock()

	if exchange.state != stateRunning {
		return ErrInvalidState
	}

	topic, ok := exchange.topicForHandle[handle]
	if !ok {
		return ErrInvalidHandle
	}

	delete(exchange.topicForHandle, handle)

	item := exchange.trie.Get(patricia.Prefix(topic))
	if item == nil {
		return ErrInvalidHandle
	}

	records := item.(*[]*handlerRecord)
	if len(*records) == 1 {
		if exchange.trie.Delete(patricia.Prefix(topic)) {
			return nil
		}
	} else {
		for i, record := range *records {
			if record.handle == handle {
				*records = append((*records)[:i], (*records)[i+1:]...)
				return nil
			}
		}
	}

	return ErrInvalidHandle
}

func (exchange *Exchange) Publish(topic Topic, event Event) error {
	exchange.mu.RLock()

	if exchange.state != stateRunning {
		return ErrInvalidState
	}

	exchange.trie.VisitPrefixes(
		patricia.Prefix(topic),
		func(prefix patricia.Prefix, item patricia.Item) error {
			for _, record := range *(item.(*[]*handlerRecord)) {
				exchange.runHandler(record.handler, topic, event)
			}
			return nil
		})

	exchange.mu.RUnlock()
	return nil
}

func (exchange *Exchange) Wait() error {
	exchange.cond.L.Lock()

	if exchange.state != stateRunning && exchange.state != stateTerminating {
		exchange.cond.L.Unlock()
		return ErrInvalidState
	}

	for atomic.LoadInt32(&exchange.numRunningHandlers) != 0 {
		exchange.cond.Wait()
	}

	exchange.cond.L.Unlock()
	return nil
}

func (exchange *Exchange) Terminate() error {
	exchange.cond.L.Lock()

	if exchange.state != stateRunning {
		exchange.cond.L.Unlock()
		return ErrInvalidState
	}

	for atomic.LoadInt32(&exchange.numRunningHandlers) != 0 {
		exchange.cond.Wait()
	}

	exchange.state = stateTerminated
	exchange.cond.L.Unlock()
	return nil
}

// Private helper methods  -----------------------------------------------------

func (exchange *Exchange) getHandle(topic Topic) (Handle, error) {
	overflow := exchange.nextHandle - 1
	for next := exchange.nextHandle; next != overflow; next++ {
		if _, ok := exchange.topicForHandle[next]; !ok {
			exchange.nextHandle = next + 1
			exchange.topicForHandle[next] = topic
			return next, nil
		}
	}
	return 0, ErrHandlesDepleted
}

func (exchange *Exchange) runHandler(handler EventHandler, topic Topic, event Event) {
	atomic.AddInt32(&exchange.numRunningHandlers, 1)
	go func() {
		defer func() {
			recover()
			atomic.AddInt32(&exchange.numRunningHandlers, -1)
			exchange.cond.Broadcast()
		}()
		handler(topic, event)
	}()
}

// Errors ----------------------------------------------------------------------

var (
	ErrInvalidState    = errors.New("invalid exchange state")
	ErrInvalidHandle   = errors.New("invalid handle")
	ErrHandlesDepleted = errors.New("handles depleted")
)
