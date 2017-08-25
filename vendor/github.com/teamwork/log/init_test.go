package log

import "testing"

type testFlusher struct {
	flushed bool
}

func (tf *testFlusher) Flush() {
	tf.flushed = true
}

func TestInit(t *testing.T) {
	l := &Logger{}
	if len(l.flushers) != 0 {
		t.Error("Flushers not initialized to empty list")
	}
	flusher1 := &testFlusher{}
	flusher2 := &testFlusher{}
	l.addFlusher(flusher1)
	l.addFlusher(flusher2)
	if len(l.flushers) != 2 {
		t.Error("Flushers not added properly")
	}
	l.Flush()
	if flusher1.flushed != true {
		t.Errorf("flusher1 didn't flush")
	}
	if flusher2.flushed != true {
		t.Errorf("flusher2 didn't flush")
	}
}
