package monitor

import (
	hook "github.com/robotn/gohook"
	"runtime"
	"testing"
)

type testKeyMapper struct{}

func (testKeyMapper) Normalize(code any) string {
	if code == uint16(0x39) {
		return "CapsLock"
	}
	if code == uint16(1) {
		return "B"
	}
	return "A"
}

func TestHiddenWindowKeepsCountsWithoutAnimation(t *testing.T) {
	keys := make(chan string, 110)
	visible := false
	var delivered []KeyState
	listener := &KeyEventListener{
		keyChan: keys, mapper: testKeyMapper{}, pressed: make(map[string]struct{}),
		isWindowVisible: func() bool { return visible },
		emitKey:         func(state KeyState) { delivered = append(delivered, state) },
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
	if len(keys) != 101 || len(delivered) != 2 || len(delivered[0].Pressed) != 1 || delivered[0].Pressed[0] != "A" || len(delivered[1].Pressed) != 0 {
		t.Fatalf("reopened counts=%d animations=%v", len(keys), delivered)
	}
	visible = false
	listener.handleEvent(hook.Event{Kind: hook.KeyDown})
	if len(keys) != 102 || len(delivered) != 2 {
		t.Fatalf("second hide counts=%d animations=%v", len(keys), delivered)
	}
}

func TestKeySnapshotsRemainOrderedAndIndependent(t *testing.T) {
	var states []KeyState
	listener := &KeyEventListener{
		keyChan: make(chan string, 10), mapper: testKeyMapper{}, pressed: make(map[string]struct{}),
		isWindowVisible: func() bool { return true }, emitKey: func(state KeyState) { states = append(states, state) },
	}
	listener.handleEvent(hook.Event{Kind: hook.KeyDown})
	listener.handleEvent(hook.Event{Kind: hook.KeyDown, Rawcode: 1})
	listener.handleEvent(hook.Event{Kind: hook.KeyUp})
	if len(states[0].Pressed) != 1 || states[0].Pressed[0] != "A" {
		t.Fatalf("old snapshot changed: %v", states[0])
	}
	if len(states[2].Pressed) != 1 || states[2].Pressed[0] != "B" {
		t.Fatalf("held key lost: %v", states[2])
	}
	for i := 1; i < len(states); i++ {
		if states[i].Sequence <= states[i-1].Sequence {
			t.Fatal("state revision did not advance")
		}
	}
}

func TestMacCapsLockTogglesCountAsTapsNotHeldKeys(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS Caps Lock semantics")
	}
	keys := make(chan string, 10)
	var states []KeyState
	listener := &KeyEventListener{
		keyChan: keys, mapper: testKeyMapper{}, pressed: make(map[string]struct{}),
		isWindowVisible: func() bool { return true }, emitKey: func(state KeyState) { states = append(states, state) },
	}
	// The hook sends down when Caps Lock turns on, up when it turns off.
	listener.handleEvent(hook.Event{Kind: hook.KeyDown, Rawcode: 0x39})
	listener.handleEvent(hook.Event{Kind: hook.KeyUp, Rawcode: 0x39})
	if len(keys) != 2 {
		t.Fatalf("two toggles counted %d times", len(keys))
	}
	for _, state := range states {
		if len(state.Pressed) != 0 || state.CapsPulse == 0 {
			t.Fatalf("Caps Lock left held: %v", state)
		}
	}
	if states[1].CapsPulse <= states[0].CapsPulse {
		t.Fatal("second tap did not advance pulse")
	}
	listener.handleEvent(hook.Event{Kind: hook.KeyDown})
	if states[2].CapsPulse != states[1].CapsPulse {
		t.Fatal("ordinary key restarted Caps Lock animation")
	}
}
