package noise

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// packet set needed properties to handle incoming message for peer.
type packet struct {
	// Ascending order for struct size
	Sig    []byte // N byte Signature
	Digest []byte // N byte Digest
}

// TODO marshall using embed encoded to reduce overhead?
// marshall encode packet to stream.
func marshall(p packet) bytes.Buffer {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	// encode packet to bytes
	encoder.Encode(p)
	return buffer
}

// unmarshall decode incoming message to packet.
func unmarshal(b []byte) packet {
	var p packet
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	// decode bytes to packet
	decoder.Decode(&p)
	return p
}

// peer its the trusty remote peer.
// Provide needed methods to interact with the secured session.
type peer struct {
	// Optimizing space with ordered types.
	// the attributes orders matters.
	// ref: https://stackoverflow.com/questions/2113751/sizeof-struct-in-go
	id   ID
	s    *session
	m    *metrics
	pool BytePool
}

// Create a new peer based on secure session
func newPeer(s *session) *peer {
	// Blake2 hashed remote public key.
	id := newBlake2ID(s.RemotePublicKey())
	return &peer{id, s, nil, nil}
}

// BindPool set a global memory pool for peer.
// Using pools remove latency from buffer allocation.
func (p *peer) BindPool(pool BytePool) {
	p.pool = pool
}

// Return peer id.
// Peer id its a blake2 hashed remote public key.
func (p *peer) ID() ID {
	return p.id
}

// Close its a forward method for internal `Close` method in session.
func (p *peer) Close() error {
	return p.s.Close()
}

// Close its a forward method for internal `SetDeadline` method in session.
func (p *peer) SetDeadline(t time.Time) error {
	return p.s.SetDeadline(t)
}

// Send send a message to Peer with size bundled in header for dynamic allocation of buffer.
// Each message is encrypted using session keys.
func (p *peer) Send(msg []byte) (uint32, error) {
	// Get a pool buffer chunk
	buffer := p.pool.Get()
	defer p.pool.Put(buffer)

	// Encrypt packet
	// we need to re-slice the buffer to avoid overflow slice because internal append.
	digest, err := p.s.Encrypt(buffer[:0], msg)
	if err != nil {
		return 0, err
	}

	// encrypted signed message
	sig := p.s.Sign(digest)
	// encode packet with signature + digest
	packed := marshall(packet{sig, digest})
	// stream encoded packet
	bytes, err := p.s.Write(packed.Bytes())
	if err != nil {
		return 0, err
	}

	return uint32(bytes), nil
}

// Listen wait for incoming messages from Peer.
// Use the needed pool buffer based on incoming header.
func (p *peer) Listen() ([]byte, error) {
	// Get a pool buffer chunk
	buffer := p.pool.Get()
	defer p.pool.Put(buffer)

	bytes, err := p.s.Read(buffer)
	log.Printf("got %d bytes from peer", bytes)

	if err == nil {
		// decode incoming package
		packet := unmarshal(buffer)
		// validate message signature
		if !p.s.Verify(packet.Digest, packet.Sig) {
			err := fmt.Errorf("invalid signature for incoming message: %s", packet.Sig)
			return nil, errVerifyingSignature(err)
		}

		// Receive secure message from peer.
		// Reuse the buffer[:0] = reset slice from byte pool.
		return p.s.Decrypt(buffer[:0], packet.Digest)
	}

	// net: don't return io.EOF from zero byte reads
	// if err == io.EOF then peer connection is closed
	_, isNetError := err.(*net.OpError)
	if err != io.EOF && !isNetError {
		// end of message, but peer is still connected
		return nil, nil
	}

	// Close disconnected peer
	if err := p.s.Close(); err != nil {
		return nil, err
	}

	// Peer disconnected
	return nil, err

}
