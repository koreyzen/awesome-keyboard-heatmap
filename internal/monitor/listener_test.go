package monitor

import (
	hook "github.com/robotn/gohook"
	"testing"
)

type testKeyMapper struct{}

func (testKeyMapper) Normalize(any) string { return "A" }

func TestHiddenWindowKeepsCountsWithoutAnimation(t *testing.T) {
	keys := make(chan string, 110)
	visible := false
	var delivered []string
	listener := &KeyEventListener{
		keyChan: keys, mapper: testKeyMapper{},
		isWindowVisible: func() bool { return visible },
		emitKey:         func(key, kind string) { delivered = append(delivered, key+":"+kind) },
	}
	for i := 0; i < 100; i++ {
		listener.handleEvent(hook.Event{Kind: hook.KeyDown})
		listener.handleEvent(hook.Event{Kind: hook.KeyUp})
	}
	if len(keys) != 100 || len(delivered) != 0 {
		t.Fatalf("hidden counts=%d animations=%v", len(keys), delivered)
	}
	visible = true
	listener.handleEvent(hook.Event{Kind: hook.KeyDown})
	listener.handleEvent(hook.Event{Kind: hook.KeyUp})
	if len(keys) != 101 || len(delivered) != 2 || delivered[0] != "A:down" || delivered[1] != "A:up" {
		t.Fatalf("reopened counts=%d animations=%v", len(keys), delivered)
	}
	visible = false
	listener.handleEvent(hook.Event{Kind: hook.KeyDown})
	if len(keys) != 102 || len(delivered) != 2 {
		t.Fatalf("second hide counts=%d animations=%v", len(keys), delivered)
	}
}
