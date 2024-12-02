package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransporrtOps{
		ListenAddr:    ":3000",
		HandShakeFunc: NoHandShakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, ":3000")
	assert.Nil(t, tr.ListenAndAccept())
	// tr.ListenAndAccept()
}
