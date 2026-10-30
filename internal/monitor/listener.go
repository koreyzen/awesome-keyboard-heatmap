package monitor

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	hook "github.com/robotn/gohook"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/zxc7563598/key-heat/pkg/keymap"
)

// A process-wide revision also stays ordered when monitoring is restarted.
var keyStateRevision atomic.Uint64

type KeyState struct {
	Sequence  uint64   `json:"sequence"`
	Pressed   []string `json:"pressed"`
	CapsPulse uint64   `json:"capsPulse"`
}

type KeyEventListener struct {
	keyChan         chan<- string
	stopChan        chan struct{}
	mapper          keymap.KeyMapper
	once            sync.Once
	emitKey         func(KeyState)
	pressed         map[string]struct{}
	capsPulse       uint64
	isWindowVisible func() bool
}

func NewKeyEventListener(keyChan chan<- string, app *application.App, isWindowVisible func() bool) *KeyEventListener {
	return &KeyEventListener{
		keyChan:  keyChan,
		stopChan: make(chan struct{}),
		mapper:   keymap.GetGlobalMapper(),
		pressed:  make(map[string]struct{}),
		emitKey: func(state KeyState) {
			app.Event.Emit("key:state", state)
		},
		isWindowVisible: isWindowVisible,
	}
}

// 开始监听
func (l *KeyEventListener) Start() error {
	log.Println("键盘监听启动中...")
	evChan := hook.Start()
	log.Println("键盘监听已启动")
	for {
		select {
		case <-l.stopChan:
			log.Println("键盘监听收到停止信号")
			hook.End() // 主动结束 hook
			return nil
		case ev, ok := <-evChan:
			if !ok {
				log.Println("键盘事件通道关闭")
				return nil
			}
			l.handleEvent(ev)
		}
	}
}

// macOS reports Caps Lock toggles, rather than a physical down/up pair.
// Keep those toggles out of the held-key set and count each one as a tap.
func (l *KeyEventListener) handleEvent(ev hook.Event) {
	if ev.Kind != hook.KeyDown && ev.Kind != hook.KeyUp {
		return
	}
	key := l.mapper.Normalize(ev.Rawcode)
	capsTap := runtime.GOOS == "darwin" && key == keymap.Key_CapsLock
	if ev.Kind == hook.KeyDown || capsTap {
		select {
		case l.keyChan <- key:
		default:
			log.Printf("通道满，丢弃: %s", key)
		}
	}
	if !capsTap {
		if ev.Kind == hook.KeyDown {
			l.pressed[key] = struct{}{}
		} else {
			delete(l.pressed, key)
		}
	}
	if !l.isWindowVisible() {
		return
	}
	sequence := keyStateRevision.Add(1)
	if capsTap {
		l.capsPulse = sequence
	}
	pressed := make([]string, 0, len(l.pressed))
	for key := range l.pressed {
		pressed = append(pressed, key)
	}
	l.emitKey(KeyState{Sequence: sequence, Pressed: pressed, CapsPulse: l.capsPulse})
}

// 停止监听
func (l *KeyEventListener) Stop() {
	l.once.Do(func() {
		close(l.stopChan)
	})
}
