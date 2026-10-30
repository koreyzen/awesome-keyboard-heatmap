export const themes = [
  { name: '红', rgb: [220, 65, 73] },
  { name: '橙', rgb: [220, 130, 46] },
  { name: '黄', rgb: [190, 154, 44] },
  { name: '绿', rgb: [62, 158, 111] },
  { name: '青', rgb: [50, 155, 170] },
  { name: '蓝', rgb: [71, 123, 205] },
  { name: '紫', rgb: [145, 96, 191] },
]

export function heatStyle(rgb, heat, dark = false) {
  // Blend white into a softer theme colour for clean, non-neon shades.
  const channels = dark
    ? rgb.map((channel, index) => {
      const base = [54, 54, 58][index]
      return Math.round(base + (channel * 0.68 - base) * heat)
    })
    : heat === 0 ? [247, 247, 247]
      : rgb.map(channel => Math.round(255 + (channel - 255) * heat))
  return {
    '--t': heat,
    '--heat-color': `rgb(${channels.join(', ')})`,
    '--key-ink': dark ? '#f2f2f7' : heat > 0.6 ? '#fff' : '#333',
  }
}
