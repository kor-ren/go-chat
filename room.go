package main

import (
	"sync"
	"time"
)

type Room struct {
	clients []chan interface{}
	lock    sync.RWMutex
}

func (r *Room) Register(channel chan interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.clients = append(r.clients, channel)
}

func (r *Room) Unregister(channel chan interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i, c := range r.clients {
		if c == channel {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			return
		}
	}
}

func (r *Room) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, c := range r.clients {
		close(c)
	}
}

func (r *Room) Send(msg interface{}) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, c := range r.clients {
		select {
		case c <- msg:
		case <-time.After(1 * time.Second):
		}
	}
}
