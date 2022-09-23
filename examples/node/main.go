package main

import (
	"context"

	"github.com/geolffreym/p2p-noise/config"
	noise "github.com/geolffreym/p2p-noise/node"
)

func main() {

	// Create configuration from params and write in configuration reference
	configuration := config.New()
	configuration.Write(
		config.SetMaxPeersConnected(10),
		config.SetPeerDeadline(1800),
	)

	// Node factory
	node := noise.New(configuration)
	// Network events channel
	ctx, cancel := context.WithCancel(context.Background())
	var signals <-chan noise.Signal = node.Signals(ctx)

	go func() {
		for signal := range signals {
			// Here could be handled events
			if signal.Type() == noise.NewPeerDetected {
				cancel() // stop listening for events
			}
		}
	}()

	// ... some code here
	// node.Dial("192.168.1.1:4008")
	// node.Close()

	// ... more code here
	node.Listen()

}
