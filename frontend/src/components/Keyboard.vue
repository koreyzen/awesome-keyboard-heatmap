<script setup>
import { usePreferredDark } from "@vueuse/core";
import { Events } from "@wailsio/runtime";
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { GetHeatmap, GetKeyLayout, GetPermission, SetMonitorStart, GetMonitorStatus, GetBootStartup, OpenBootStartup, GetPrompt } from "../../bindings/github.com/zxc7563598/key-heat/greetservice";
import KeyTooltip from "./KeyTooltip.vue";
import { createKeyboardGeometry } from "../keyboardGeometry.js";
import { themes, heatStyle } from "../heatColors.js";
import PromptDialog from "./PromptDialog.vue"

const isDark = usePreferredDark()
const theme = ref(themes.find(item => item.name === localStorage.getItem('heat-theme')) || themes[0])
const selectTheme = (item) => {
  theme.value = item
  localStorage.setItem('heat-theme', item.name)
}

// 当前监听状态
const monitor = ref(false)
const permission = ref(false)
const bootStartup = ref(false)

// 启动监听
const setMonitorStart = () => {
  console.log("启动监听", "启动")
  SetMonitorStart().then(() => {
    GetMonitorStatus().then(async (result) => {
      monitor.value = result
    }).catch((err) => {
      console.log(err)
    })
  }).catch((err) => {
    console.log(err)
  })
}

// 开启自启动
const openBootStartup = () => {
  OpenBootStartup().then(() => {
    GetBootStartup().then(async (result) => {
      bootStartup.value = result
    }).catch((err) => {
      console.log(err)
    })
  }).catch((err) => {
    console.log(err)
  })
}

// 按键按压状态
const activeKeys = ref(new Set())
const handleBlur = () => {
  activeKeys.value = new Set()
}

// All positions and sizes share the same standard-key coordinate system.
const geometry = ref({ columns: 14.5, rows: 6 })
const getKeyContentStyle = (key) => ({
  left: `${key.X / geometry.value.columns * 100}%`,
  top: `${key.Y / geometry.value.rows * 100}%`,
  width: `${key.W / geometry.value.columns * 100}%`,
  height: `${key.H / geometry.value.rows * 100}%`,
  '--label-offset': key.X === 0 ? (1 - key.W) / 2
    : key.X + key.W === geometry.value.columns ? (key.W - 1) / 2 : 0,
})

// 热力图作为键盘上的可切换叠层
const heatmapEnabled = ref(false)
const countsEnabled = ref(false)
const statsEnabled = computed(() => heatmapEnabled.value || countsEnabled.value)
const totalCount = computed(() => layout.value.reduce((total, key) => total + key.Count, 0))
const toggleHeatmap = () => {
  heatmapEnabled.value = !heatmapEnabled.value
  hoverKey.value = null
  if (heatmapEnabled.value && !countsEnabled.value) applyRange()
}
const toggleCounts = () => {
  countsEnabled.value = !countsEnabled.value
  hoverKey.value = null
  if (countsEnabled.value && !heatmapEnabled.value) applyRange()
}
// 快捷日期范围，周一为每周第一天
const ranges = ['今日', '昨日', '本周', '上周', '历史全部']
const selectedRange = ref('今日')
const dateRange = (range) => {
  if (range === '历史全部') return ['', '']
  const start = new Date()
  const end = new Date(start)
  if (range === '昨日') {
    start.setDate(start.getDate() - 1)
    end.setDate(end.getDate() - 1)
  } else if (range === '本周' || range === '上周') {
    start.setDate(start.getDate() - (start.getDay() + 6) % 7 - (range === '上周' ? 7 : 0))
    end.setTime(start.getTime())
    end.setDate(start.getDate() + 6)
  }
  const format = (date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  return [format(start), format(end)]
}
const selectRange = (range) => {
  selectedRange.value = range
  applyRange()
}
const applyRange = () => {
  const [start, end] = dateRange(selectedRange.value)
  GetHeatmap(start, end).then(async (result) => {
    // 找最大值
    const values = Object.values(result)
    const max = Math.max(...values, 0)
    // 映射 + 写回
    layout.value.forEach(key => {
      const raw = result[key.Code] || 0
      const t = ref(0)
      if (raw != 0) {
        const logVal = Math.log(raw + 1) / Math.log(max + 1) // log 压缩
        const normalized = Math.pow(logVal, 1) // 再轻微调整分布
        t.value = Math.pow(normalized, 1.4) // 视觉映射
      }
      Object.assign(key, {
        Hot: t.value,
        Count: raw
      })
    })
  }).catch((err) => {
    console.log("热力图获取失败", err)
  })
};

// 悬停提示固定在当前按键上方
const hoverKey = ref(null)
const onKeyEnter = (key) => {
  if (statsEnabled.value) hoverKey.value = key
}
const onKeyLeave = () => {
  hoverKey.value = null
}

// 导出数据
const showDialog = ref(false)
const promptText = ref("")

const getPrompt = () => {
  const [start, end] = dateRange(selectedRange.value)
  GetPrompt(start, end).then((result) => {
    promptText.value = result
    showDialog.value = true
  }).catch((err) => {
    console.log("提示词获取失败", err)
  })
}

const layout = ref([])
let stopStatsRefresh
const refreshVisibleStats = () => {
  if (document.visibilityState === 'visible') {
    applyRange()
  }
}
const handleFocus = () => {
  checkPermission()
  refreshVisibleStats()
}
onMounted(async () => {
  stopStatsRefresh = Events.On('stats:updated', refreshVisibleStats)
  // 获取键盘布局
  GetKeyLayout("ANSI").then(async (result) => {
    const positioned = createKeyboardGeometry(result)
    geometry.value = { columns: positioned.columns, rows: positioned.rows }
    layout.value = positioned.keys.map(key => ({ ...key, Hot: 0, Count: 0 }))
    applyRange()
    window.addEventListener("blur", handleBlur)
  }).catch((err) => {
    console.log(err)
  })
  // 键盘事件监听
  Events.On("key:pressed", (event) => {
    const { key, type } = event.data
    const next = new Set(activeKeys.value)
    if (type === "down") {
      next.add(key)
    } else if (type === "up") {
      next.delete(key)
    }
    activeKeys.value = next
  })
  Events.On("active:switch", (event) => {
    monitor.value = event.data
  })
  // 检查权限
  await checkPermission()
  document.addEventListener("visibilitychange", handleFocus)
  window.addEventListener("focus", handleFocus)
  if (!permission.value) {
    startPermissionWatch()
  }
})

onBeforeUnmount(() => {
  stopStatsRefresh?.()
  document.removeEventListener("visibilitychange", handleFocus)
  window.removeEventListener("focus", handleFocus)
  window.removeEventListener("blur", handleBlur)
})

const checkPermission = () => {
  // 确认监听权限
  GetPermission().then(async (result) => {
    permission.value = result
  }).catch((err) => {
    console.log(err)
  })
  // 确认监听状态
  GetMonitorStatus().then(async (result) => {
    monitor.value = result
  }).catch((err) => {
    console.log(err)
  })
  // 确认开机自启动
  GetBootStartup().then(async (result) => {
    bootStartup.value = result
  }).catch((err) => {
    console.log(err)
  })
}

let timer = null
const startPermissionWatch = () => {
  if (timer) return
  timer = setInterval(async () => {
    checkPermission()
    if (permission.value) {
      clearInterval(timer)
      timer = null
    }
  }, 1500)
}

</script>

<template>
  <div class="status-bar">
    <div class="status-left">
      <div class="action-btn primary" v-if="!monitor" @click="setMonitorStart">
        启动监听
      </div>
      <div class="action-btn primary" v-if="!bootStartup" @click="openBootStartup">
        开机自启动
      </div>
    </div>
    <div class="status-right">
      <div class="theme-palette" role="group" aria-label="热力主题色">
        <button v-for="item in themes" :key="item.name" class="theme-swatch"
          :style="{ '--swatch': `rgb(${item.rgb.join(', ')})` }"
          :aria-label="`${item.name}色主题`" :aria-pressed="theme.name === item.name"
          @click="selectTheme(item)"></button>
      </div>
      <div class="status-item">
        <div class="dot-wrapper">
          <span class="dot" :class="{ ok: bootStartup, bad: !bootStartup }"></span>
          <div class="tooltip-status">
            {{ bootStartup ? "已开启开机自启动" : "未开启开机自启动" }}
          </div>
        </div>
        <span class="label">自启动</span>
      </div>
      <div class="status-item">
        <div class="dot-wrapper">
          <span class="dot" :class="{ ok: permission, bad: !permission }"></span>
          <div class="tooltip-status">
            {{ permission ? "已授予辅助功能权限" : "未授予辅助功能权限" }}
          </div>
        </div>
        <span class="label">权限</span>
      </div>
      <div class="status-item">
        <div class="dot-wrapper">
          <span class="dot monitor" :class="{ ok: monitor, bad: !monitor }"></span>
          <div class="tooltip-status">
            {{ monitor ? "正在监听键盘输入" : "监听已停止" }}
          </div>
        </div>
        <span class="label">监听</span>
      </div>
    </div>
  </div>
  <div class="keyboard" :style="{ '--keyboard-columns': geometry.columns, '--keyboard-rows': geometry.rows }">
    <div class="key-grid">
      <div v-for="key in layout" :key="key.Code" class="key-content"
        :class="{ 'split-top': key.Code === '↑', 'split-bottom': key.Code === '↓' }"
        :style="getKeyContentStyle(key)">
        <div class="key" :class="{ active: activeKeys.has(key.Code) }"
          :style="heatStyle(theme.rgb, heatmapEnabled ? key.Hot : 0, isDark)"
          @mouseenter="onKeyEnter(key)" @mouseleave="onKeyLeave">
          <span v-if="countsEnabled" class="key-count">{{ key.Count }}</span>
          <span v-else class="key-label">
            <svg v-if="key.Code === 'TouchID'" class="key-icon"
              viewBox="0 0 24 24" role="img" :aria-label="key.Code">
              <circle cx="12" cy="12" r="8" stroke-width="2.4" />
            </svg>
            <svg v-else-if="key.Code === 'Enter' || key.Code === 'NumPad Enter'" class="key-icon"
              viewBox="0 0 24 24" role="img" :aria-label="key.Code">
              <path d="M19 5v8H5m5-5-5 5 5 5" />
            </svg>
            <svg v-else-if="key.Code === 'Shift' || key.Code === 'RightShift'" class="key-icon"
              viewBox="0 0 24 24" role="img" :aria-label="key.Code">
              <path d="M12 4l8 8h-5v8H9v-8H4z" />
            </svg>
            <span v-else>{{ key.Label }}</span>
          </span>
          <Transition name="key-tooltip">
            <KeyTooltip v-if="hoverKey === key" :data="key" />
          </Transition>
        </div>
      </div>
    </div>
  </div>
  <div class="functional">
    <div class="action-btn" :class="{ selected: heatmapEnabled }" @click="toggleHeatmap">
      热力图
    </div>
    <div class="action-btn count-toggle" :class="{ selected: countsEnabled }" @click="toggleCounts">
      次数图
    </div>
    <div class="date-range">
      <span class="total-count">{{ selectedRange }}共 {{ totalCount.toLocaleString() }} 次</span>
      <div v-for="range in ranges" :key="range" class="action-btn"
        :class="{ selected: selectedRange === range }" @click="selectRange(range)">
        {{ range }}
      </div>
      <div class="action-btn" @click="getPrompt">
        生成报告
      </div>
    </div>
  </div>
  <PromptDialog v-model="showDialog" :prompt="promptText" />
</template>

<style scoped>
.keyboard {
  align-self: center;
  box-sizing: border-box;
  width: min(100%, calc((100vh - 2 * var(--toolbar-height) - 30px) * var(--keyboard-columns) / var(--keyboard-rows) + 30px));
  padding: 15px;
  border-radius: 24px;
  background: var(--panel);
  backdrop-filter: blur(20px);
  box-shadow: var(--panel-shadow);
  margin: 0 auto;
}

.key-grid {
  position: relative;
  container-type: inline-size;
  aspect-ratio: var(--keyboard-columns) / var(--keyboard-rows);
}

.key-content {
  --key-gap: 5px;
  position: absolute;
  box-sizing: border-box;
  display: flex;
  font-size: 12px;
}

.key {
  position: relative;
  margin: var(--key-gap);
  width: calc(100% - 2 * var(--key-gap));
  height: calc(100% - 2 * var(--key-gap));
  --t: 0;
  text-align: center;
  white-space: pre-wrap;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  cursor: pointer;
  border-radius: 10px;
  justify-content: center;
  font-weight: 500;
  background: var(--heat-color);
  color: var(--key-ink);
  box-shadow: var(--key-shadow);
  transition: transform 0.08s ease, box-shadow 0.12s ease, background 0.2s ease, color 0.2s ease;
}

.key-label {
  display: flex;
  align-items: center;
  justify-content: center;
  transform: translateX(calc(var(--label-offset) * 100cqw / var(--keyboard-columns)));
}

.key-icon {
  width: 1.25em;
  height: 1.25em;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.key:active,
.key.active {
  transform: translateY(2px) scale(0.96) !important;
  box-shadow: var(--key-pressed-shadow) !important;
}

/* One keycap outline, with two independently coloured and interactive halves. */
.split-top::before {
  content: '';
  position: absolute;
  top: var(--key-gap);
  left: var(--key-gap);
  width: calc(100% - 2 * var(--key-gap));
  height: calc(200% - 2 * var(--key-gap));
  border-radius: 10px;
  box-shadow: var(--key-shadow);
  pointer-events: none;
}

.split-top .key,
.split-bottom .key {
  height: calc(100% - var(--key-gap));
  box-shadow: none;
}

.split-top .key {
  margin-bottom: 0;
  border-radius: 10px 10px 0 0;
}

.split-bottom .key {
  margin-top: 0;
  border-radius: 0 0 10px 10px;
}

:is(.split-top, .split-bottom) .key:is(:active, .active) {
  transform: none !important;
}

.functional {
  display: flex;
  align-items: center;
  width: 100%;
}

.count-toggle {
  margin-left: 12px;
}

.total-count {
  margin-right: 8px;
  font-size: 13px;
  color: var(--secondary-text);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.key-count {
  font-variant-numeric: tabular-nums;
}

.date-range {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.action-btn.selected {
  color: var(--control-hover);
  background: var(--selected);
  box-shadow: var(--selected-shadow);
}

.status-bar {
  --wails-draggable: drag;
  box-sizing: border-box;
  padding-left: calc(112px - var(--page-inset));
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}


.status-left {
  display: flex;
  align-items: center;
  gap: 18px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 6px;
  position: relative;
}

.label {
  font-size: 13px;
  color: var(--secondary-text);
  letter-spacing: 0.2px;
}

.dot-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #c7c7cc;
  transition: all 0.25s ease;
}

.dot.ok {
  background: #34c759;
  box-shadow: 0 0 6px rgba(52, 199, 89, 0.6);
}

.dot.bad {
  background: #ff3b30;
  box-shadow: 0 0 6px rgba(255, 59, 48, 0.5);
}

.dot.monitor.ok {
  animation: pulse 1.8s ease-in-out infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 4px rgba(52, 199, 89, 0.5);
  }

  50% {
    box-shadow: 0 0 10px rgba(52, 199, 89, 0.9);
  }

  100% {
    box-shadow: 0 0 4px rgba(52, 199, 89, 0.5);
  }
}

.tooltip-status {
  position: absolute;
  top: 140%;
  left: 50%;
  transform: translateX(-50%) scale(0.96);
  transform-origin: top center;
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 12px;
  color: #fff;
  white-space: nowrap;
  background: var(--popover);
  backdrop-filter: blur(8px);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.18s ease, transform 0.18s cubic-bezier(0.4, 0, 0.2, 1);
}

.tooltip-status::after {
  content: "";
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-bottom-color: var(--popover);
}

.dot-wrapper:hover .tooltip-status {
  opacity: 1;
  transform: translateX(-50%) scale(1);
}

.status-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.theme-palette {
  --wails-draggable: no-drag;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-right: 8px;
}

.theme-swatch {
  width: 12px;
  height: 12px;
  padding: 0;
  border: 0;
  border-radius: 3px;
  background: var(--swatch);
  opacity: var(--swatch-opacity);
  cursor: pointer;
  transition: opacity 0.3s ease;
}

.theme-swatch[aria-pressed="true"] {
  outline: 1px solid var(--swatch);
  outline-offset: 2px;
}

.theme-swatch:hover,
.theme-swatch:focus-visible {
  opacity: 1;
  filter: saturate(1.45);
}

.action-btn {
  --wails-draggable: no-drag;
  white-space: nowrap;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
  background: var(--control);
  color: var(--control-text);
  box-shadow: var(--control-shadow);
  transition: all 0.2s ease;
}

.action-btn:hover {
  transform: translateY(-0.5px);
  color: var(--control-hover);
}

.action-btn:active {
  transform: scale(0.96);
}

.action-btn.primary {
  color: var(--positive);
}

</style>
