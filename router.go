package noise

import (
	"sync"
)

type (
	Socket string
	Table  map[Socket]*Peer
)

func (t Table) Add(peer *Peer) {
	t[peer.Socket()] = peer
}

func (t Table) Remove(peer *Peer) {
	delete(t, peer.Socket())
}

// Router hash table to associate Socket with Peers.
// Unstructured mesh architecture
// eg. {127.0.0.1:4000: Peer}
type Router struct {
	sync.RWMutex
	table Table
}

func newRouter() *Router {
	return &Router{
		table: make(Table),
	}
}

// Table return current routing table
func (r *Router) Table() Table { return r.table }

// Return connection interface based on socket
func (r *Router) Query(socket Socket) *Peer {
	// Mutex for reading topics.
	// Do not write while topics are read.
	// Write Lock can’t be acquired until all Read Locks are released.
	// ref: https://pkg.go.dev/sync#RWMutex.Lock
	r.RWMutex.RLock()
	defer r.RWMutex.RUnlock()

	if peer, ok := r.table[socket]; ok {
		return peer
	}

	return nil
}

// Add create new socket connection association.
// It return recently added peer.
func (r *Router) Add(peer *Peer) *Peer {
	// Lock write/read table while add operation
	// A blocked Lock call excludes new readers from acquiring the lock.
	// ref: https://pkg.go.dev/sync#RWMutex.Lock
	r.RWMutex.Lock()
	r.table.Add(peer)
	r.RWMutex.Unlock()
	return peer
}

// Len return the number of connections
func (r *Router) Len() int {
	return len(r.table)
}

// Remove removes a connection from router.
// It return recently removed peer.
func (r *Router) Remove(peer *Peer) *Peer {
	// Lock write/read table while add operation
	// A blocked Lock call excludes new readers from acquiring the lock.
	// ref: https://pkg.go.dev/sync#RWMutex.Lock
	r.RWMutex.Lock()
	r.table.Remove(peer)
	r.RWMutex.Unlock()
	return peer
}
