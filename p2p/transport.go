package p2p

import "net"

// Peer is the interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	// RemoteAddr() net.Addr
	// Close() error

	CloseStream()
}

// Transport is anything that handles the communication in between node in networks
type Transport interface {
	Addr() string
	Dial(string) error

	// No matter what type of transport it is, It need to have a listen and accept fuction
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	// ListenAddr() string
}
