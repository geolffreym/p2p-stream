package network

import "net"

// Peer struct has a simplistic interface to describe a node in the network.
// Each peer has a socket address to identify itself and a connection interface to communicate with it.
// Each peer is routed and handled by route table.
type Peer struct {
	socket Socket   // IP and Port address for peer. https://en.wikipedia.org/wiki/Network_socket
	conn   net.Conn // Connection interface net.Conn to reach peer.
}

// Return peer connection interface
func (p *Peer) Connection() net.Conn { return p.conn }

// Return peer socket
func (p *Peer) Socket() Socket { return p.socket }

// Write buffered message over connection
func (p *Peer) Write(data []byte) (n int, err error) { return p.conn.Write(data) }

// Read buffered message from connection
func (p *Peer) Read(buf []byte) (n int, err error) { return p.conn.Read(buf) }

// Close peer connection
func (p *Peer) Close() error { return p.conn.Close() }
