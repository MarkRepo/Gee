package cache

import pb "github.com/MarkRepo/Gee/GeeCache/cache/cachepb"

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}

// PeerPicker is the interface that must be implemented to locate the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
