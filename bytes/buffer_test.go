package bytes

import (
	"testing"
)

func TestBuffer(t *testing.T) {
	w := NewBufferSize(10)
	w.WriteString("my writer test")
	w.WriteString(" new+")

	t.Logf("print:%s\n", w.Buffer())
}
