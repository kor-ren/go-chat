package main

import "sync"

type Message struct {
	Message string
	RoomId  string
	UserId  string
}

type Listener struct {
	RoomId string
	Chan   chan interface{}
}

type RoomManager struct {
	roomChannels map[string]*Room
	open         chan *Listener
	close        chan *Listener
	delete       chan string
	messages     chan *Message
}

func NewRoomManager() *RoomManager {
	manager := &RoomManager{
		roomChannels: make(map[string]*Room),
		open:         make(chan *Listener, 100),
		close:        make(chan *Listener, 100),
		delete:       make(chan string, 100),
		messages:     make(chan *Message, 100),
	}

	go manager.run()
	return manager
}

func (m *RoomManager) run() {
	for {
		select {
		case l := <-m.open:
			m.register(l)
		case l := <-m.close:
			m.derigister(l)
		case roomId := <-m.delete:
			m.deleteBroadcaster(roomId)
		case message := <-m.messages:
			m.room(message.RoomId).Send(message.UserId + ": " + message.Message)
		}
	}
}

func (m *RoomManager) room(roomId string) *Room {
	r, ok := m.roomChannels[roomId]

	if !ok {
		r = &Room{
			clients: make([]chan interface{}, 0),
			lock:    sync.RWMutex{},
		}
		m.roomChannels[roomId] = r
	}

	return r
}

func (m *RoomManager) register(l *Listener) {
	m.room(l.RoomId).Register(l.Chan)
}

func (m *RoomManager) derigister(l *Listener) {
	m.room(l.RoomId).Unregister(l.Chan)
}

func (m *RoomManager) deleteBroadcaster(roomId string) {
	b, ok := m.roomChannels[roomId]
	if ok {
		b.Close()
		delete(m.roomChannels, roomId)
	}
}

func (m *RoomManager) OpenListener(roomId string) chan interface{} {
	listener := make(chan interface{})
	m.open <- &Listener{
		RoomId: roomId,
		Chan:   listener,
	}

	return listener
}

func (m *RoomManager) CloseListener(roomId string, channel chan interface{}) {
	m.close <- &Listener{
		RoomId: roomId,
		Chan:   channel,
	}
}

func (m *RoomManager) DeleteBroadcast(roomId string) {
	m.delete <- roomId
}

func (m *RoomManager) Submit(userId, roomId, text string) {
	msg := &Message{
		UserId:  userId,
		RoomId:  roomId,
		Message: text,
	}
	m.messages <- msg
}
