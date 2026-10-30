package monitor

import (
	"log"
	"sync"

	hook "github.com/robotn/gohook"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/zxc7563598/key-heat/pkg/keymap"
)

type KeyEventListener struct {
	keyChan         chan<- string
	stopChan        chan struct{}
	mapper          keymap.KeyMapper
	once            sync.Once
	emitKey         func(string, string)
	isWindowVisible func() bool
}

func NewKeyEventListener(keyChan chan<- string, app *application.App, isWindowVisible func() bool) *KeyEventListener {
	return &KeyEventListener{
		keyChan:  keyChan,
		stopChan: make(chan struct{}),
		mapper:   keymap.GetGlobalMapper(),
		emitKey: func(key, kind string) {
			app.Event.Emit("key:pressed", map[string]any{"key": key, "type": kind})
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

// Count every keydown; animation delivery is only needed for a visible window.
func (l *KeyEventListener) handleEvent(ev hook.Event) {
	if ev.Kind != hook.KeyDown && ev.Kind != hook.KeyUp {
		return
	}
	key := l.mapper.Normalize(ev.Rawcode)
	kind := "up"
	if ev.Kind == hook.KeyDown {
		kind = "down"
		select {
		case l.keyChan <- key:
		default:
			log.Printf("通道满，丢弃: %s", key)
		}
	}
	if l.isWindowVisible() {
		l.emitKey(key, kind)
	}
}

// 停止监听
func (l *KeyEventListener) Stop() {
	l.once.Do(func() {
		close(l.stopChan)
	})
}
