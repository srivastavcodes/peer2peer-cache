package p2pcache

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

// ByteView holds an immutable view of bytes. Internally it wraps
// either a []byte or a string.
//
// ByteView is meant to be used as a value not a pointer.
type ByteView struct {
	b []byte // b is used if non-nil, else s is used.
	s string
}

// Copy copies v data into dst and returns the number of bytes copied.
func (v ByteView) Copy(dst []byte) int {
	if v.b != nil {
		return copy(dst, v.b)
	}
	return copy(dst, v.s)
}

// ByteSlice returns a copy of data as a byte slice.
func (v ByteView) ByteSlice() []byte {
	if v.b != nil {
		return cloneBytes(v.b)
	}
	return []byte(v.s)
}

// At returns the byte at index i.
func (v ByteView) At(i int) byte {
	if v.b != nil {
		return v.b[i]
	}
	return v.s[i]
}

// Equal returns whether the bytes in v.b are the same as the bytes in bv2.
func (v ByteView) Equal(bv2 ByteView) bool {
	if bv2.b == nil {
		return v.EqualString(bv2.s)
	}
	return v.EqualBytes(bv2.b)
}

// Slice slices the view between the provided form and to indices.
func (v ByteView) Slice(from, to int) ByteView {
	if v.b != nil {
		return ByteView{b: v.b[from:to]}
	}
	return ByteView{s: v.s[from:to]}
}

// SliceFrom slices the view from the provided index until the end.
func (v ByteView) SliceFrom(from int) ByteView {
	if v.b != nil {
		return ByteView{b: v.b[from:]}
	}
	return ByteView{s: v.s[from:]}
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	if v.b != nil {
		return string(v.b)
	}
	return v.s
}

// Len returns the view's length.
func (v ByteView) Len() int {
	if v.b != nil {
		return len(v.b)
	} else {
		return len(v.s)
	}
}

// EqualString returns whether the bytes in v are the same as the bytes in s.
func (v ByteView) EqualString(s string) bool {
	if v.b == nil {
		return v.s == s
	}
	if l := v.Len(); l != len(s) {
		return false
	}
	for i, p := range v.b {
		if p != s[i] {
			return false
		}
	}
	return true
}

// EqualBytes returns whether the bytes in v are the same as the bytes in s.
func (v ByteView) EqualBytes(p []byte) bool {
	if v.b != nil {
		return bytes.Equal(v.b, p)
	}
	if l := v.Len(); l != len(p) {
		return false
	}
	for i, b := range p {
		if v.s[i] != b {
			return false
		}
	}
	return true
}

// Reader returns an io.ReadSeeker for the bytes in v.
func (v ByteView) Reader() io.ReadSeeker {
	if v.b != nil {
		return bytes.NewReader(v.b)
	}
	return strings.NewReader(v.s)
}

// ReadAt implements io.ReaderAt on the bytes in v.
func (v ByteView) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("byteview: invalid offset")
	}
	if off >= int64(len(p)) {
		return 0, io.EOF
	}
	n = v.SliceFrom(int(off)).Copy(p)
	if n < len(p) {
		err = io.EOF
	}
	return
}

// WriteTo implements io.WriterTo on the byte in v.
func (v ByteView) WriteTo(w io.Writer) (n int64, err error) {
	var c int
	if v.b != nil {
		c, err = w.Write(v.b)
	} else {
		c, err = io.WriteString(w, v.s)
	}
	if err == nil && c < v.Len() {
		err = io.ErrShortWrite
	}
	n = int64(c)
	return
}
