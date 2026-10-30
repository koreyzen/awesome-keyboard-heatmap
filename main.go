package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/zxc7563598/key-heat/internal/monitor"
	"github.com/zxc7563598/key-heat/internal/storage"
	"github.com/zxc7563598/key-heat/pkg/permissions"
)

//go:embed all:frontend/dist
var assets embed.FS

type KeyboardApp struct {
	app           *application.App
	window        *application.WebviewWindow
	repo          storage.Repository
	mon           *monitor.Monitor
	isActive      bool
	windowVisible atomic.Bool
	appHidden     atomic.Bool
}

func main() {
	if alreadyRunning() {
		os.Exit(0)
	}
	defer cleanupLock()
	// 初始化数据库
	db, err := storage.NewDB(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("数据库初始化成功:", db.GetPath())
	// 自动迁移表结构
	if err := storage.Run(db.DB); err != nil {
		log.Fatal(err)
	}
	fmt.Println("数据库自动迁移成功")
	// 注入 repository
	repo := storage.NewRepository(db.DB)
	// 构建APP
	greet := &GreetService{
		repo: repo,
	}
	app := application.New(application.Options{
		Name:        "Keyboard Heatmap",
		Description: "键盘按键热力图监控工具 - 实时记录并可视化您的键盘使用习惯",
		Services: []application.Service{
			application.NewService(greet),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})
	// 启动应用程序
	keyboardApp := &KeyboardApp{
		app:  app,
		repo: repo,
	}
	greet.t = keyboardApp
	keyboardApp.setup()
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// 系统构建核心逻辑
func (t *KeyboardApp) setup() {
	t.openWindow()
	t.isActive = false
	if permissions.HasAccessibilityPermission() {
		t.isActive = true
		t.mon = monitor.NewMonitor(t.repo, nil, t.app, t.windowVisible.Load)
		go t.mon.Start()
	}
	t.app.Event.Emit("active:switch", t.isActive)
}

// 打开窗口
func (t *KeyboardApp) openWindow() {
	t.window = t.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:         "Keyboard Heatmap",
		Width:         780,
		Height:        429,
		DisableResize: true,
		Mac: application.MacWindow{
			Backdrop:   application.MacBackdropTranslucent,
			Appearance: application.DefaultAppearance,
			TitleBar:   application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
	t.windowVisible.Store(true)
	visibilityChanged := func(*application.WindowEvent) { t.refreshWindowVisibility() }
	windowEvents := []events.WindowEventType{events.Common.WindowShow, events.Common.WindowHide, events.Common.WindowMinimise, events.Common.WindowRestore}
	if runtime.GOOS == "darwin" {
		windowEvents = []events.WindowEventType{events.Mac.WindowShow, events.Mac.WindowHide, events.Mac.WindowDidMiniaturize, events.Mac.WindowDidDeminiaturize}
		t.app.Event.OnApplicationEvent(events.Mac.ApplicationDidHide, func(*application.ApplicationEvent) { t.appHidden.Store(true); t.windowVisible.Store(false) })
		t.app.Event.OnApplicationEvent(events.Mac.ApplicationDidUnhide, func(*application.ApplicationEvent) { t.appHidden.Store(false); t.refreshWindowVisibility() })
	}
	for _, event := range windowEvents {
		t.window.OnWindowEvent(event, visibilityChanged)
	}
	if runtime.GOOS == "darwin" {
		// Keep the window available for reopening from the Dock while counting continues.
		t.window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
			t.window.Hide()
			e.Cancel()
		})
	}
}

func (t *KeyboardApp) refreshWindowVisibility() {
	t.windowVisible.Store(!t.appHidden.Load() && t.window.IsVisible() && !t.window.IsMinimised())
}

// 开启监听
func (t *KeyboardApp) startListening() {
	log.Println("startListening开始调用")
	if permissions.HasAccessibilityPermission() {
		t.mon = monitor.NewMonitor(t.repo, nil, t.app, t.windowVisible.Load)
		go t.mon.Start()
		t.isActive = true
		t.app.Event.Emit("active:switch", t.isActive)
	} else {
		t.showPermissionDialog()
	}
}

// 关闭监听
func (t *KeyboardApp) stopListening() {
	log.Println("stopListening开始调用")
	t.mon.Stop()
	t.isActive = false
	t.app.Event.Emit("active:switch", t.isActive)
}

// 设置权限弹窗
func (t *KeyboardApp) showPermissionDialog() {
	dialog := t.app.Dialog.Question().
		SetTitle("需要权限").
		SetMessage("需要开启辅助功能权限才能监听键盘，是否前往设置？")
	goSetting := dialog.AddButton("前往设置")
	goSetting.OnClick(func() {
		exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
	})
	goCancel := dialog.AddButton("取消")
	goCancel.OnClick(func() {
		t.app.Quit()
	})
	dialog.Show()
}

// 单实例限制
func alreadyRunning() bool {
	lockFile := filepath.Join(os.TempDir(), "keyheat.lock")
	data, err := os.ReadFile(lockFile)
	if err == nil && len(data) > 0 {
		oldPID, err := strconv.Atoi(string(data))
		if err == nil {
			// 判断进程是否还活着
			if processExists(oldPID) {
				return true
			}
		}
		// 删除旧锁
		_ = os.Remove(lockFile)
	}
	// 写入当前 PID
	pid := os.Getpid()
	_ = os.WriteFile(lockFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	return false
}

// 检查进程是否存在
func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = p.Signal(os.Signal(syscall.Signal(0)))
	return err == nil
}

// 释放单实例限制
func cleanupLock() {
	lockFilePath := filepath.Join(os.TempDir(), "keyheat.lock")
	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		return
	}
	currentPID := os.Getpid()
	expected := fmt.Sprintf("%d", currentPID)
	if string(data) == expected {
		_ = os.Remove(lockFilePath)
	}
}
