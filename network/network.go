package network

import (
	"io"
	"log"
	"net"

	"github.com/geolffreym/p2p-noise/errors"
	"github.com/geolffreym/p2p-noise/pubsub"
)

const PROTOCOL = "tcp"

type Network struct {
	table  Router
	closed chan bool
	Events pubsub.Channel
}

func New() *Network {
	return &Network{
		table:  make(Router),
		closed: make(chan bool, 1),
		Events: make(pubsub.Channel),
	}
}

// Create a peer from net connection
func (network *Network) peer(conn net.Conn) *Peer {
	return &Peer{
		conn:   conn,
		socket: Socket(conn.RemoteAddr().String()),
	}
}

// Initialize route in routing table
func (network *Network) routing(conn net.Conn) *Peer {
	// Keep routing for each connection
	socket := Socket(conn.RemoteAddr().String())
	route := network.peer(conn)
	network.table.Add(socket, route)
	return route
}

// Run routed stream message in goroutine
// Each incoming message processed in concurrent approach
func (network *Network) stream(peer *Peer) {
	go func(n *Network, p *Peer) {
		buf := make([]byte, 1024)

	KEEPALIVE:
		for {
			// Stop routine
			if n.IsClosed() {
				return
			}

			_, err := p.Read(buf)
			if err != nil {
				if err == io.EOF {
					break KEEPALIVE
				}
			}

			// TODO Need refactor to handle biggest messages
			// Emit new incoming
			message := pubsub.NewMessage(pubsub.MESSAGE_RECEIVED, buf)
			n.Events.Publish(message)

		}
	}(network, peer)
}

// Concurrent Bind network and set routing to start listening for streams
func (network *Network) bind(listener net.Listener) {
	go func(n *Network, l net.Listener) {
		for {
			// Block/Hold while waiting for new incoming connection
			conn, err := l.Accept()
			if err != nil || n.IsClosed() {
				log.Fatalf(errors.Binding(err).Error())
				return
			}

			// Routing for connection
			peer := n.routing(conn)
			n.stream(peer)
			// Dispatch event
			payload := []byte(peer.Socket())
			message := pubsub.NewMessage(pubsub.NEWPEER_DETECTED, payload)
			n.Events.Publish(message)
		}
	}(network, listener)
}

// Start listening on the given address and wait for new connection
func (network *Network) Listen(addr string) (*Network, error) {
	listener, err := net.Listen(PROTOCOL, addr)
	if err != nil {
		return nil, errors.Listen(err, addr)
	}

	// Concurrent processing for each incoming connection
	network.bind(listener)
	// Dispatch event on start listening
	payload := []byte(addr)
	message := pubsub.NewMessage(pubsub.SELF_LISTENING, payload)
	network.Events.Publish(message)
	return network, nil
}

func (network *Network) Table() Router {
	return network.table
}

// Non-blocking check connection state
func (network *Network) IsClosed() bool {
	select {
	case <-network.closed:
		return true
	default:
	}

	return false
}

// Close all peers connections
func (network *Network) Close() {
	for _, route := range network.table {
		go func(r *Peer) {
			if err := r.Close(); err != nil {
				log.Fatalf(errors.Close(err).Error())
			}
		}(route)
	}

	close(network.closed)
}

// Dial to a network node and add route to table
func (network *Network) Dial(addr string) (*Network, error) {
	conn, err := net.Dial(PROTOCOL, addr)
	if err != nil {
		return nil, errors.Dial(err, addr)
	}

	// Routing for connection
	peer := network.routing(conn)
	network.stream(peer)
	// Dispatch event
	payload := []byte(peer.Socket())
	message := pubsub.NewMessage(pubsub.NEWPEER_DETECTED, payload)
	network.Events.Publish(message)
	return network, nil
}
