// To run test in this we run this command
// go test ./... -v

// To run this i have used the default run
// go run .
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/lakshayletsgo/Distributed_File_Storage_System/p2p"
)

func makeserver(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransporrtOps{
		ListenAddr:    listenAddr,
		HandShakeFunc: p2p.NoHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		EncKey:            EncryptionKey(),
		StorageRoot:       listenAddr[1:] + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeserver(":3000", "")
	s2 := makeserver(":4000", ":3000")
	s3 := makeserver(":5000", ":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(1 * time.Second)
	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(2 * time.Second)
	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {

		// key := "systemHaiLadle.jpg"
		key := fmt.Sprintf("system_%d.png", i)
		data := bytes.NewReader([]byte("Finally Ho gya!"))
		s2.Store(key, data)
		// time.Sleep(time.Millisecond * 5)

		if err := s2.store.Delete(s2.ID, key); err != nil {
			log.Fatal(err)
		}
		n, err := s2.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(n)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))

	}
	// select {}
}
