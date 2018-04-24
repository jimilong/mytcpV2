package bytes

import (
	"strconv"
)

type Buffer struct {
	n   int
	buf []byte
}

func NewBufferSize(n int) *Buffer {
	return &Buffer{buf: make([]byte, n)}
}

func (b *Buffer) Size() int {
	return len(b.buf)
}

func (b *Buffer) Reset() {
	b.n = 0
}

func (b *Buffer) Buffer() []byte {
	return b.buf[:b.n]
}

func (b *Buffer) Peek(n int) []byte {
	var buf []byte
	b.grow(n)
	buf = b.buf[b.n : b.n+n]
	b.n += n
	return buf
}

func (b *Buffer) Write(p []byte) {
	b.grow(len(p))
	b.n += copy(b.buf[b.n:], p)
}

func (b *Buffer) grow(n int) {
	var buf []byte
	if b.n+n < len(b.buf) {
		return
	}
	buf = make([]byte, 2*len(b.buf)+n)
	copy(buf, b.buf[:b.n])
	b.buf = buf
	return
}

//new func longmsdu
func (b *Buffer) WriteString(s string) {
	b.grow(len(s))
	b.n += copy(b.buf[b.n:], s)
}

//new func longmsdu
func (b *Buffer) AppendInt(i int64, base int) {
	s := strconv.FormatInt(i, base)
	b.WriteString(s)
}
