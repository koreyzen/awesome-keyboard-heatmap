export const themes = [
  { name: '红', rgb: [255, 40, 40] },
  { name: '橙', rgb: [255, 135, 0] },
  { name: '黄', rgb: [255, 210, 0] },
  { name: '绿', rgb: [25, 200, 85] },
  { name: '青', rgb: [0, 195, 215] },
  { name: '蓝', rgb: [0, 110, 255] },
  { name: '紫', rgb: [165, 55, 255] },
]

export function heatStyle(rgb, heat) {
  // Tint the vivid theme colour with white instead of desaturating it toward grey.
  const channels = heat === 0 ? [247, 247, 247]
    : rgb.map(channel => Math.round(255 + (channel - 255) * heat))
  const linear = channels.map(channel => {
    const value = channel / 255
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  })
  const luminance = linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722
  const whiteContrast = 1.05 / (luminance + 0.05)
  const darkContrast = (luminance + 0.05) / 0.066
  return {
    '--t': heat,
    '--heat-color': `rgb(${channels.join(', ')})`,
    '--key-ink': whiteContrast > darkContrast ? '#fff' : '#222',
  }
}
