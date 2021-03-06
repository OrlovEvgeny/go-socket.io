package socketio

import "sync"

// BroadcastAdaptor is the adaptor to handle broadcasts.
type BroadcastAdaptor interface {

	// Join causes the socket to join a room.
	Join(room string, socket Socket) error

	// JoinBy causes the socket to join a room by uniq index.
	JoinBy(room string, index string, socket Socket) error

	//LeaveBy causes the socket to leave a room.
	LeaveBy(room string, index string, socket Socket) error

	// Leave causes the socket to leave a room.
	Leave(room string, socket Socket) error

	// Send will send an event with args to the room. If "ignore" is not nil, the event will be excluded from being sent to "ignore".
	Send(ignore Socket, room, event string, args ...interface{}) error

	//Len socket in room
	Len(room string) int
}

var newBroadcast = newBroadcastDefault

type broadcast struct {
	m map[string]map[string]Socket
	sync.RWMutex
}

func newBroadcastDefault() BroadcastAdaptor {
	return &broadcast{
		m: make(map[string]map[string]Socket),
	}
}

func (b *broadcast) Join(room string, socket Socket) error {
	b.Lock()
	sockets, ok := b.m[room]
	if !ok {
		sockets = make(map[string]Socket)
	}
	sockets[socket.Id()] = socket
	b.m[room] = sockets
	b.Unlock()
	return nil
}

func (b *broadcast) JoinBy(room string, index string, socket Socket) error  {
	b.Lock()
	sockets, ok := b.m[room]
	if !ok {
		sockets = make(map[string]Socket)
	}
	sockets[index] = socket
	b.m[room] = sockets
	b.Unlock()
	return nil
}

func (b *broadcast) LeaveBy(room string, index string, socket Socket) error {
	b.Lock()
	defer b.Unlock()
	sockets, ok := b.m[room]
	if !ok {
		return nil
	}
	delete(sockets, index)
	if len(sockets) == 0 {
		delete(b.m, room)
		return nil
	}
	b.m[room] = sockets
	return nil
}

func (b *broadcast) Leave(room string, socket Socket) error {
	b.Lock()
	defer b.Unlock()
	sockets, ok := b.m[room]
	if !ok {
		return nil
	}
	delete(sockets, socket.Id())
	if len(sockets) == 0 {
		delete(b.m, room)
		return nil
	}
	b.m[room] = sockets
	return nil
}

func (b *broadcast) Send(ignore Socket, room, event string, args ...interface{}) error {
	b.RLock()
	sockets := b.m[room]
	for id, s := range sockets {
		if ignore != nil && ignore.Id() == id {
			continue
		}
		s.Emit(event, args...)
	}
	b.RUnlock()
	return nil
}

func (b *broadcast) Len(room string) int {
	return len(b.m[":"+room])
}
