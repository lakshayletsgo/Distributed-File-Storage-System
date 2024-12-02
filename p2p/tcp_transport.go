// Always keep public functions at the top and private fucntions at the bottom
// Keep Important function at the top

package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCP Peer represents the remote node over a TCP established connection
type TCPPeer struct {
	// conn is the underlying connection of peer
	net.Conn

	// This is used to tell whether the peer is accepting or sending to the database
	// If sending then bool is true
	// Else false
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// This will close the connection of the peer
// func (t *TCPPeer) Close() error {
// 	return t.Conn.Close()
// }

type TCPTransporrtOps struct {
	ListenAddr    string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransporrtOps
	listener net.Listener
	rpcch    chan RPC

	// This mutex will protect the peer
	// mu    sync.RWMutex
	// peers map[net.Addr]Peer
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// This is used to create a new transport layer
// Here we have a choice we can either return the transport interface or the TCPTransporrt
// If we return the transport layer then it would be difficult to test like we have to cast to the tcptransport so we return this
func NewTCPTransport(opts TCPTransporrtOps) *TCPTransport {
	return &TCPTransport{
		TCPTransporrtOps: opts,
		rpcch:            make(chan RPC, 1024),
	}

}

// This syntax means it can only read not write on the channel
// This will return the read only channel that are recieved from another peer
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

// A transport always listens or accepts some data
// THis is a method of struct TCPTransport
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}
	go t.startAcceptLoop()

	log.Printf("TCP Transport listening on port %s\n", t.ListenAddr)
	return nil
}

func (t *TCPPeer) Send(b []byte) error {
	_, err := t.Conn.Write(b)
	return err
}

func (t *TCPTransport) Addr() string {

	return t.ListenAddr
}

// There is another function called accept loop that will go on for accepting some files
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("Acceptance Letter Err : %v\n", err)

		}

		go t.handleConn(conn, false)

	}
}

type Temp struct{}

// func (p *TCPPeer) RemoteAddr() net.Addr {
// 	return p.conn.RemoteAddr()
// }

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection %v\n ", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)

	if err = t.HandShakeFunc(peer); err != nil {
		// conn.Close()
		// fmt.Printf("TCP handshake error : %s\n", err)
		return
	}

	// lenDecoderError := 0
	// Read Loop
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}
	// buf := make([]byte, 2000)
	for {
		rpc := RPC{}
		// n, err := conn.Read(buf)

		err = t.Decoder.Decode(conn, &rpc)
		// if err == net.ErrClosed {
		// 	return
		// }

		// Here the connection will be dropped if there is any type of error
		// But we want to have a connection dropped for a particular error only not for all the error
		if err != nil {
			// lenDecoderError++
			// if lenDecoderError == 5 {
			// 	conn.Close()
			// 	// break
			// 	return
			// }
			// fmt.Printf("TCP read error : %s\n", err)
			// continue

			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%v] incomming message, waiting for the stream....\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%v] stream closed, resuming the read loop\n", conn.RemoteAddr())
			continue

		}

		// fmt.Println("Waiting till stream is done")
		// fmt.Printf("This is the message %+v\n", rpc)
		t.rpcch <- rpc
		// fmt.Println("Streaming done continuing normal loop")

	}

	// msg:=&Message{}
	// reader := bufio.NewReader(conn)
	// for {
	// 	m, err := reader.ReadString('\n') // Reads until newline
	// 	if err != nil {
	// 		fmt.Printf("TCP error: %s\n", err)
	// 		break
	// 	}
	// 	fmt.Printf("Received: %s\n", input)
	// }

}

// Every transport should have a channel to communicate
