package p2pcache

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// Sink receives data from a Get call. Implementation of Getter must call
// exactly one of the Set methods on success.
type Sink interface {
	// SetProto sets the value to the encoded version of m. The caller
	// retains the ownership of m.
	SetProto(m proto.Message) error

	// SetBytes sets the value to the contents of b. The caller retains
	// ownership of v.
	SetBytes(b []byte) error

	// SetString sets the value to str.
	SetString(str string) error

	// view returns a frozen view of the bytes for caching.
	view() (ByteView, error)
}

// cloneBytes deep copies b and returns the copy.
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func setSinkView(sink Sink, v ByteView) error {
	// viewSetter is a Sink that can also receive its value from a ByteView.
	// This is a fast path to minimize copies when the item was already
	// cached locally in memory (where it's cached as ByteView)
	type viewSetter interface {
		setView(v ByteView) error
	}
	if vs, ok := sink.(viewSetter); ok {
		return vs.setView(v)
	}
	if v.b != nil {
		return sink.SetBytes(v.b)
	}
	return sink.SetString(v.s)
}

type stringSink struct {
	str *string
	v   ByteView
}

// StringSink returns a Sink that populates the provided string pointer.
func StringSink(str *string) Sink {
	return &stringSink{str: str}
}

func (sk *stringSink) SetProto(m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	sk.v.b = b
	*sk.str = string(b)
	return nil
}

func (sk *stringSink) SetString(str string) error {
	sk.v.b = nil
	sk.v.s = str
	*sk.str = str
	return nil
}

func (sk *stringSink) SetBytes(b []byte) error {
	return sk.SetString(string(b))
}

func (sk *stringSink) view() (ByteView, error) {
	return sk.v, nil
}

type byteViewSink struct {
	dst *ByteView
}

// ByteViewSink returns a Sink that populates a ByteView.
func ByteViewSink(dst *ByteView) Sink {
	if dst == nil {
		panic("nil dst")
	}
	return &byteViewSink{dst: dst}
}

func (sk *byteViewSink) SetProto(m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	*sk.dst = ByteView{b: b}
	return nil
}

func (sk *byteViewSink) setView(v ByteView) error {
	*sk.dst = v
	return nil
}

func (sk *byteViewSink) view() (ByteView, error) {
	return *sk.dst, nil
}

func (sk *byteViewSink) SetBytes(b []byte) error {
	*sk.dst = ByteView{b: cloneBytes(b)}
	return nil
}

func (sk *byteViewSink) SetString(str string) error {
	*sk.dst = ByteView{s: str}
	return nil
}

type protoSink struct {
	dst proto.Message // authoritative value
	typ string

	v ByteView // encoded
}

// ProtoSink returns a Sink that unmarshals binary proto values into m.
func ProtoSink(m proto.Message) Sink {
	return &protoSink{
		dst: m,
	}
}

func (sk *protoSink) SetProto(m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(b, sk.dst)
	if err != nil {
		return err
	}
	sk.v.b = b
	sk.v.s = ""
	return nil
}

func (sk *protoSink) view() (ByteView, error) {
	return sk.v, nil
}

func (sk *protoSink) SetBytes(b []byte) error {
	err := proto.Unmarshal(b, sk.dst)
	if err != nil {
		return err
	}
	sk.v.b = cloneBytes(b)
	sk.v.s = ""
	return nil
}

func (sk *protoSink) SetString(str string) error {
	b := []byte(str)
	err := proto.Unmarshal(b, sk.dst)
	if err != nil {
		return err
	}
	sk.v.b = b
	sk.v.s = ""
	return nil
}

type allocBytesSink struct {
	dst *[]byte
	v   ByteView
}

// AllocatingByteSliceSink returns a Sink that allocates a byte slice
// to hold the received value and assigns it to *dst. The memory is
// not retained by cache.
func AllocatingByteSliceSink(dst *[]byte) Sink {
	return &allocBytesSink{dst: dst}
}

func (sk *allocBytesSink) SetProto(m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	return sk.setBytesOwned(b)
}

func (sk *allocBytesSink) view() (ByteView, error) {
	return sk.v, nil
}

func (sk *allocBytesSink) setView(v ByteView) error {
	if v.b != nil {
		*sk.dst = cloneBytes(v.b)
	} else {
		*sk.dst = []byte(v.s)
	}
	sk.v = v
	return nil
}

func (sk *allocBytesSink) SetBytes(b []byte) error {
	return sk.setBytesOwned(cloneBytes(b))
}

func (sk *allocBytesSink) setBytesOwned(b []byte) error {
	if sk.dst == nil {
		return errors.New("nil AllocatingByteSliceSink *[]byte dst")
	}
	*sk.dst = cloneBytes(b) // another copy, protecting the read-only sk.v.b view
	sk.v.b = b
	sk.v.s = ""
	return nil
}

func (sk *allocBytesSink) SetString(str string) error {
	if sk.dst == nil {
		return errors.New("nil AllocatingByteSliceSink *[]byte dst")
	}
	*sk.dst = []byte(str)
	sk.v.b = nil
	sk.v.s = str
	return nil
}

type truncBytesSink struct {
	dst *[]byte
	v   ByteView
}

// TruncatingByteSliceSink returns a Sink that writes up to len(*dst) bytes
// to *dst. If more bytes are available, they're silently truncated. If
// fewer bytes are available than len(*dst), *dst is shrunk to fit the number
// of bytes available.
func TruncatingByteSliceSink(dst *[]byte) Sink {
	return &truncBytesSink{dst: dst}
}

func (sk *truncBytesSink) SetProto(m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	return sk.setBytesOwned(b)
}

func (sk *truncBytesSink) view() (ByteView, error) {
	return sk.v, nil
}

func (sk *truncBytesSink) SetBytes(b []byte) error {
	return sk.setBytesOwned(cloneBytes(b))
}

func (sk *truncBytesSink) setBytesOwned(b []byte) error {
	if sk.dst == nil {
		return errors.New("nil TruncatingByteSliceSink *[]byte dst")
	}
	n := copy(*sk.dst, b)
	if n < len(*sk.dst) {
		*sk.dst = (*sk.dst)[:n]
	}
	sk.v.b = b
	sk.v.s = ""
	return nil
}

func (sk *truncBytesSink) SetString(str string) error {
	if sk.dst == nil {
		return errors.New("nil TruncatingByteSliceSink *[]byte dst")
	}
	n := copy(*sk.dst, str)
	if n < len(*sk.dst) {
		*sk.dst = (*sk.dst)[:n]
	}
	sk.v.b = nil
	sk.v.s = str
	return nil
}
