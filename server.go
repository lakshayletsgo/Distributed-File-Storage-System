package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/lakshayletsgo/Distributed_File_Storage_System/p2p"
)

// func init() {
// 	gob.Register(Message{})
// }

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type MessageStoreFile struct {
	Id   string
	Key  string
	Size int64
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.ID) == 0 {
		opts.ID = generateId()
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected to remote %v", p.RemoteAddr())

	return nil
}

// func (s *FileServer) broadcast(msg *Message) error {
// 	// buf := new(bytes.Buffer)
// 	// for _, peer := range s.peers {
// 	// 	if err := gob.NewEncoder(buf).Encode(p); err != nil {
// 	// 		return err
// 	// 	}
// 	// 	peer.Send(buf.Bytes())
// 	// }
// 	// return nil

// 	// for _, peer := range s.peers {

// 	// 	if err := gob.NewEncoder(peer).Encode(p); err != nil {
// 	// 		return err
// 	// 	}
// 	// }
// 	// return nil

// 	peers := []io.Writer{}
// 	for _, peer := range s.peers {
// 		peers = append(peers, peer)
// 	}
// 	mw := io.MultiWriter(peers...)
// 	return gob.NewEncoder(mw).Encode(msg)
// }

// func (s *FileServer) broadcast(p *Payload) error {
// 	log.Printf("Broadcasting payload: %+v", p)
// 	peers := []io.Writer{}
// 	for _, peer := range s.peers {
// 		peers = append(peers, peer)
// 	}
// 	mw := io.MultiWriter(peers...)
// 	err := gob.NewEncoder(mw).Encode(p)
// 	if err != nil {
// 		log.Printf("Error during broadcasting: %v", err)
// 	}
// 	return err
// }

func (s *FileServer) stream(msg *Message) error {
	// log.Printf("Broadcasting payload: %+v", p)
	// Encode the payload to a byte buffer
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	// Send the encoded data to all peers
	for _, peer := range s.peers {
		err := peer.Send(buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		fmt.Printf("Error in sending the payload")
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type Message struct {
	// From    string
	Payload any
}
type MessageGet struct {
	ID  string
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(s.ID, key) {
		fmt.Printf("%v Serving file %v from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}
	fmt.Printf("[%s] Don't have the file with the key (%v). Searching on the network....\n", s.Transport.Addr(), key)
	msg := Message{
		Payload: MessageGet{
			ID:  s.ID,
			Key: hashKey(key),
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)
	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		// fileBuff := new(bytes.Buffer)
		// n, err := io.CopyN(fileBuff, peer, 15)
		// // fmt.Printf("There is an error %v\n", key)
		// if err != nil {
		// 	return nil, err

		// }
		fmt.Printf("[%v] Recieved %d bytes on the network from %v\n ", s.Transport.Addr(), n, peer.RemoteAddr())
		// fmt.Println(fileBuff.String())

		peer.CloseStream()
	}
	// select {}
	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)
	// if err := s.store.Write(key, tee); err != nil {
	// 	fmt.Printf("Inside the if statement : %v\n", err)
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }
	// // fmt.Println(buf.Bytes())

	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Id:   s.ID,
			Key:  hashKey(key),
			Size: size + 16,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})

	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Recieved and Written %d bytes to disk.\n", s.Transport.Addr(), n)

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File Server Stopped")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				// fmt.Println("Phir se chud gye")
				log.Println("Decoding Error : ", err)
				// return
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Handle Message Error : ", err)
				// log.Println(err)
				// return
			}

			// fmt.Printf("Recieved :%v\n", string(msg.Payload.([]byte)))
			// fmt.Printf("Payload: %v\n", msg.Payload)

			// peer, ok := s.peers[rpc.From]
			// if !ok {
			// 	panic("peer not found in the peer map")
			// }
			// b := make([]byte, 1000)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }

			// fmt.Printf("%v\n", string(b))

			// peer.(*p2p.TCPPeer).Wg.Done()
			// panic("ddd")

			// if err := s.handleMessage(&m); err != nil {
			// 	log.Println(err)
			// }
		case <-s.quitch:
			return

		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGet:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}
func (s *FileServer) handleMessageGetFile(from string, msg MessageGet) error {
	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve a file (%v) but it doesn't exist on the disk", s.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%v] serving file (%s) on the network.\n", s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("Closing the read closer")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not in the map", from)
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] Written %d bytes over the network to %v\n", s.Transport.Addr(), n, from)

	return nil

}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) couldn't be found.\n", from)
	}
	n, err := s.store.Write(msg.Id, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s]Writtne %d bytes to diksk\n", s.Transport.Addr(), n)
	// peer.(*p2p.TCPPeer).Wg.Done()

	peer.CloseStream()
	return nil
}

// This will help to store the file
// func (s *FileServer) Store(key string, r io.Reader) error {
// 	return s.store.Write(key, r)

// }

func (s *FileServer) bootstrapnetwork() error {

	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("Attempting to connect to : %v\n", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("There is error while dialing %v\n", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapnetwork()
	s.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGet{})
}
