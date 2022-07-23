// Package conf provide a "functional option" design pattern to handle node settings.
// See also: https://github.com/crazybber/awesome-patterns/blob/master/idiom/functional-options.md
package config

import "time"

// Functional options
type Config struct {
	maxPeersConnected uint8
	peerDeadline      time.Duration
}

type Setter func(*Config)

// Return default settings
func New() *Config {
	return &Config{
		// Max peer consecutively connected.
		// Each of this peers is equivalent to one routine, limit this is a performance consideration.
		maxPeersConnected: 100,
		// Max time waiting for I/O or peer interaction. After this time the connection will timeout and considered inactive.
		// Default 1800 seconds = 30 minutes.
		peerDeadline: 1800,
	}
}

// Write stores settings in `Settings` struct reference.
// All the settings are passed as an array of `setters` to then get called with `Settings`` reference as param.
// ref: https://github.com/crazybber/awesome-patterns/blob/master/idiom/functional-options.md
func (s *Config) Write(c ...Setter) {
	for _, setter := range c {
		setter(s)
	}
}

// MaxPeersConnected returns the max number of connections.
func (s *Config) MaxPeersConnected() uint8 {
	return s.maxPeersConnected
}

// PeerDeadline returns the max time waiting for I/O or peer interaction
func (s *Config) PeerDeadline() time.Duration {
	return s.peerDeadline
}

// SetMaxPeersConnected sets the maximum number of connections allowed for routing.
// If the number of connections > MaxPeersConnected then router drop new connections.
func SetMaxPeersConnected(maxPeers uint8) Setter {
	return func(conf *Config) {
		conf.maxPeersConnected = maxPeers
	}
}

// SetPeerDeadline sets how long in seconds peer connections can remain idle.
// If deadline for I/O is exceeded a timeout is raised and the connection is closed.
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to Read or Write.
// ref: https://pkg.go.dev/net#Conn
func SetPeerDeadline(timeout time.Duration) Setter {
	return func(conf *Config) {
		conf.peerDeadline = timeout
	}
}
