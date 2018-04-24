package bytes

import (
	"sync"
)

type ChainBuffer struct {
	buf  []byte
	next *ChainBuffer // next free ChainBuffer
}

func (b *ChainBuffer) Bytes() []byte {
	return b.buf
}

// Pool is a ChainBuffer pool.
type Pool struct {
	lock sync.Mutex
	free *ChainBuffer
	max  int
	num  int
	size int
}

// NewPool new a memory ChainBuffer pool struct.
func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return
}

// Init init the memory ChainBuffer.
func (p *Pool) Init(num, size int) {
	p.init(num, size)
	return
}

// init init the memory ChainBuffer.
func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	p.max = num * size
	p.grow()
}

// grow grow the memory ChainBuffer size, and update free pointer.
func (p *Pool) grow() {
	var (
		i   int
		b   *ChainBuffer
		bs  []ChainBuffer
		buf []byte
	)
	buf = make([]byte, p.max)
	bs = make([]ChainBuffer, p.num)
	p.free = &bs[0]
	b = p.free
	for i = 1; i < p.num; i++ {
		b.buf = buf[(i-1)*p.size : i*p.size]
		b.next = &bs[i]
		b = b.next
	}
	b.buf = buf[(i-1)*p.size : i*p.size]
	b.next = nil
	return
}

// Get get a free memory ChainBuffer.
func (p *Pool) Get() (b *ChainBuffer) {
	p.lock.Lock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}
	p.free = b.next
	p.lock.Unlock()
	return
}

// Put put back a memory ChainBuffer to free.
func (p *Pool) Put(b *ChainBuffer) {
	p.lock.Lock()
	b.next = p.free
	p.free = b
	p.lock.Unlock()
	return
}
