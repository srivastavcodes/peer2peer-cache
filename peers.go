package p2pcache

import (
	"context"
	pb "peer2peer-cache/p2pcachepb/v1"
)

// ProtoGetter is an interface that must be implemented by a peer.
type ProtoGetter interface {
	Get(ctx context.Context, in *pb.GetRequest, out *pb.GetResponse) error
}

// PeerPicker is the interface that must be implemented to locate
// the peer that own a specific key.
type PeerPicker interface {
	// PeerPicker returns the peer that owns a specific key and
	// true to indicate that a remote peer was nominated.
	// It returns (nil, false) if the key owner is current peer.
	PeerPicker(key string) (peer ProtoGetter, ok bool)
}

// NoPeer is an implementation of PeerPicker that never finds a peer.
type NoPeer struct {
	// empty
}

func (NoPeer) PeerPicker(_ string) (peer ProtoGetter, ok bool) {
	return
}

var (
	portPicker func(groupName string) PeerPicker
)

// RegisterPeerPicker registers the peer initialization function. It is
// called once, when the first group is created.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPeerPicker(fn func() PeerPicker) {
	if portPicker != nil {
		panic("peer picker called more than once")
	}
	portPicker = func(_ string) PeerPicker {
		return fn()
	}
}

// RegisterPerGroupPeerPicker registers the peer initialization function,
// which takes the groupName, to be used in choosing a PeerPicker. It is
// called exactly one, when the first group is called.
// Either RegisterPeerPicker or RegisterPerGroupPeerPicker should be
// called exactly once, but not both.
func RegisterPerGroupPeerPicker(fn func(groupName string) PeerPicker) {
	if portPicker != nil {
		panic("peer picker called more than once")
	}
	portPicker = fn
}
