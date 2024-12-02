package p2p

type HandShakeFunc func(Peer) error

func NoHandShakeFunc(Peer) error { return nil }
