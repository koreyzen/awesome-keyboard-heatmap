# Keyboard Heatmap

把每天的键盘使用，变成一张看得见的热力图。

基于 [zxc7563598/key-heat](https://github.com/zxc7563598/key-heat) 的 macOS 体验改进版，由 [koreyzen](https://github.com/koreyzen) 维护。保留原工程的全局按键统计、本地 SQLite 存储和开机自启动能力，重点重新打磨键盘布局与交互。

![Keyboard Heatmap macOS 实机界面](assets/keyboard-heatmap-macos.png)

## 相比原工程，改了什么？

| 部分 | 原工程 | Keyboard Heatmap |
| --- | --- | --- |
| macOS 键盘布局 | 带数字区、导航区的完整键盘 | MacBook Pro 风格的紧凑 ANSI 配列，按统一键宽比例计算位置 |
| 热力配色 | 单一红色，低、中热度容易发灰 | 红橙黄绿青蓝紫七种柔和主题，以白色调浅主题色；高热度保留白字 |
| 显示方式 | 基础、热力图两种模式 | 热力图与次数图独立开关，可叠加显示 |
| 时间范围 | 手动选择起止日期 | 今日、昨日、本周、上周、历史全部，一键切换 |
| 总计 | 查看各键次数 | 增加当前范围总计，与画面中各键次数之和一致 |
| 悬浮提示 | 多行信息，随鼠标移动 | 按键名与次数一行展示，固定在对应键上方，0.3 秒淡入淡出 |
| 窗口与字体 | 较宽的可调整窗口 | 固定 780 × 429 窗口，苹方字体，布局避开 macOS 窗口控制区 |
| 按键细节 | 不同字形的粗细、位置不一致 | 统一回车、Shift、电源图标；边缘宽键文字按标准键中心线对齐 |
| 方向键 | 独立方向键 | 左右键正常高度，上下键共用方形外观，两半独立统计、着色与提示 |
| 顶部操作 | 常驻启动、停止类操作 | 已启动的操作自动隐藏，保留三项状态；色块仅悬停单个时高亮 |

macOS 使用 Dock 作为应用入口，不显示菜单栏托盘图标。关闭窗口后继续后台统计，点击 Dock 图标可重新打开。

主题选择保存在本机，切换颜色不会额外查询数据库。文字不会被误选中，时间筛选和总计始终显示。

## 使用

开启「热力图」查看按键使用分布；开启「次数图」把键帽文字替换成次数；两者可以同时开启。顶部色块切换主题，底部选择时间范围。「生成报告」生成可复制的分析提示文本。

统计按日期和按键名称聚合，保存于本机 SQLite，macOS 路径为：

```text
~/Library/Application Support/KeyHeat/keyheat.db
```

沿用原工程每 10 秒或累计 500 次按键批量写入的方式。记录的是每个键的次数，不保存输入文本或按键顺序。当前总计对应屏幕所展示的配列；外接键盘独有的按键不计入这张配列的总计。

## macOS 本地构建

准备 Go 1.25+、Node.js 22 和 Xcode Command Line Tools，然后执行：

```bash
git clone https://github.com/koreyzen/keyboard-heatmap.git
cd keyboard-heatmap
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.74
export PATH="$(go env GOPATH)/bin:$PATH"
npm --prefix frontend ci
wails3 task darwin:package
```

产物为 `bin/Keyboard Heatmap.app`。放入「应用程序」后启动，在「系统设置 → 隐私与安全性 → 辅助功能」中开启 Keyboard Heatmap 的权限。

也可在 [Actions](https://github.com/koreyzen/keyboard-heatmap/actions/workflows/build-macos.yml) 的成功构建中下载 `macos-app` 产物。自动构建使用临时签名；面向分发的 Developer ID 签名和公证由发布者配置。

持续本地更新时，可以使用钥匙串中固定的代码签名证书：

```bash
KEY_HEAT_SIGN_IDENTITY="你的代码签名证书名称" wails3 task darwin:package
```

保持签名身份、应用标识和安装路径一致，可减少更新后重复授予辅助功能权限的情况。本项目的界面改进在 macOS 上验证；仓库同时保留原工程的 Windows 构建支持。

## 来源与许可证

感谢何俊杰创建的 [Key Heat](https://github.com/zxc7563598/key-heat)。本仓库保留原工程历史和版权声明，在其基础上进行界面与体验改进，继续使用 [MIT License](LICENSE)。

GitHub 搜索：`keyboard-heatmap user:koreyzen`。
