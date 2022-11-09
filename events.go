package noise

import (
	"unsafe"
)

// [Event] aliases for int type.
type Event uint8

const (
	// Event to notify when a new peer get connected
	NewPeerDetected Event = iota
	// On new message received event
	MessageReceived
	// Closed peer connection
	PeerDisconnected
)

// events handle event exchange between [Node] and network.
type events struct {
	broker     *broker
	subscriber *subscriber
}

func newEvents() *events {
	subscriber := newSubscriber()
	broker := newBroker()
	// register default events
	broker.Register(NewPeerDetected, subscriber)
	broker.Register(MessageReceived, subscriber)
	broker.Register(PeerDisconnected, subscriber)

	return &events{
		broker,
		subscriber,
	}
}

// Flush broker topics and shutdown subscriber.
func (e *events) Flush() {
	e.broker.Flush()
	// Don't close the channel
	// https://stackoverflow.com/questions/8593645/is-it-ok-to-leave-a-channel-open
	// e.subscriber.Close()
}

// Listen forward to Listen method to internal subscriber.
func (e *events) Listen(ch chan<- Signal) {
	e.subscriber.Listen(ch)
}

// PeerConnected dispatch event when new peer is detected.
func (e *events) PeerConnected(peer *peer) {
	// Emit new notification
	body := peer.ID().String()
	header := header{peer, NewPeerDetected}
	signal := Signal{header, body}
	e.broker.Publish(signal)
}

// PeerDisconnected dispatch event when peer get disconnected.
func (e *events) PeerDisconnected(peer *peer) {
	// Emit new notification
	body := peer.ID().String()
	header := header{peer, PeerDisconnected}
	signal := Signal{header, body}
	e.broker.Publish(signal)
}

// NewMessage dispatch event when a new message is received.
func (e *events) NewMessage(peer *peer, msg []byte) {
	// Emit new notification
	// perf: no-copy convert to string
	message := *(*string)(unsafe.Pointer(&msg))
	header := header{peer, MessageReceived}
	signal := Signal{header, message}
	e.broker.Publish(signal)
}
