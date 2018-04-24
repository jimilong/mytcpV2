package bytes

import (
	"testing"
)

func TestChainBuffer(t *testing.T) {
	p := NewPool(1, 10)
	b := p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	t.Logf("1:%p\n", b.Bytes())
	b = p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	t.Logf("2:%p\n", b.Bytes())
	b = p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	t.Logf("3:%p\n", b.Bytes())
	p.Put(b)
	b = p.Get()
	if b.Bytes() == nil || len(b.Bytes()) == 0 {
		t.FailNow()
	}
	t.Logf("4:%p\n", b.Bytes())
}
