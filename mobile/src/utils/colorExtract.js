// 品牌色提取与主题派生工具（brand-color-extract spec）。
//
// 纯前端方案：Canvas 读取 logo 像素 → 色相分桶取众数 → 派生 Element Plus 主题变量族
// → 注入 :root。不引入任何第三方依赖，也不做后端图像处理（design D1）。
//
// 桌面端与移动端各持一份完全相同的实现：spec 要求两端导出接口一致，
// 由 backend/internal/guard/color_extract_test.go 编译成会红的静态检查。

// 混合目标色（Element Plus 的 light-N / dark-N 派生就是与主色做线性混合）
const WHITE = '#ffffff'
const BLACK = '#000000'

// 回退主题色：无 logo 或提取失败时的期望呈现。
// 注意：这两个常量是「回退应该长什么样」的单一真相，回退本身并不靠注入它们实现——
// 见 clearThemeVars 的说明：回退 = 不注入，由各端 CSS 的 var() fallback 兜底。
export const DEFAULT_PRIMARY = '#1f6feb'
export const DEFAULT_BRAND_TO = '#0d3b8c'

// 采样尺寸：64×64 = 4096 像素，遍历 < 1ms，用户无感（design D2）
const SAMPLE_SIZE = 64
// 色相分桶粒度：每 10° 一桶，共 36 桶
const HUE_BUCKETS = 36
// 像素过滤阈值（spec：跳过透明背景、灰白黑文字与描边）
const ALPHA_MIN = 128
const SATURATION_MIN = 0.15
const LIGHTNESS_MIN = 0.1
const LIGHTNESS_MAX = 0.9
// 亮色主色的明度钳制（design D4）：L > 0.6 时钳到 0.5，避免 light-9 几乎不可见
const LIGHTNESS_CLAMP_THRESHOLD = 0.6
const LIGHTNESS_CLAMP_TARGET = 0.5
// 登录页品牌区遮罩透明度：沿用现有视觉的 0.72 / 0.85
const SCRIM_ALPHA_FROM = 0.72
const SCRIM_ALPHA_TO = 0.85

// Element Plus 变体权重：light-N = 与白色混合 N/10，dark-2 = 与黑色混合 0.2。
// 与 EP 的 SCSS 派生口径一致（D3），保证 hover / active 态不出现色偏。
const LIGHT_3_WEIGHT = 0.3
const LIGHT_5_WEIGHT = 0.5
const LIGHT_7_WEIGHT = 0.7
const LIGHT_8_WEIGHT = 0.8
const LIGHT_9_WEIGHT = 0.9
const DARK_2_WEIGHT = 0.2
// 登录页渐变终点：主色与黑色按 3:7 混合
const BRAND_TO_WEIGHT = 0.3

function clamp01(n) {
  return Math.min(1, Math.max(0, n))
}

function clamp255(n) {
  return Math.min(255, Math.max(0, n))
}

/**
 * hex 转 RGB。支持 `#rgb` / `#rrggbb`（带不带 # 均可），非法输入返回 null。
 */
export function hexToRgb(hex) {
  let h = String(hex == null ? '' : hex).trim().replace(/^#/, '')
  if (h.length === 3) {
    h = h
      .split('')
      .map((c) => c + c)
      .join('')
  }
  if (!/^[0-9a-fA-F]{6}$/.test(h)) return null
  return {
    r: parseInt(h.slice(0, 2), 16),
    g: parseInt(h.slice(2, 4), 16),
    b: parseInt(h.slice(4, 6), 16),
  }
}

/**
 * RGB 转 hex。各通道先钳到 0-255 再取整。
 */
export function rgbToHex(rgb) {
  const to2 = (n) => Math.round(clamp255(n)).toString(16).padStart(2, '0')
  return `#${to2(rgb.r)}${to2(rgb.g)}${to2(rgb.b)}`
}

/**
 * RGB（0-255）转 HSL（h/s/l 均归一化到 0-1）。
 */
export function rgbToHsl(rgb) {
  const r = clamp255(rgb.r) / 255
  const g = clamp255(rgb.g) / 255
  const b = clamp255(rgb.b) / 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const l = (max + min) / 2
  const d = max - min

  let h = 0
  let s = 0
  if (d !== 0) {
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    if (max === r) {
      h = (g - b) / d + (g < b ? 6 : 0)
    } else if (max === g) {
      h = (b - r) / d + 2
    } else {
      h = (r - g) / d + 4
    }
    h /= 6
  }
  return { h, s, l }
}

/**
 * HSL（h/s/l 均 0-1）转 RGB（0-255）。
 */
export function hslToRgb(h, s, l) {
  const hue = ((h % 1) + 1) % 1
  const sat = clamp01(s)
  const lig = clamp01(l)

  if (sat === 0) {
    const v = lig * 255
    return { r: v, g: v, b: v }
  }

  const q = lig < 0.5 ? lig * (1 + sat) : lig + sat - lig * sat
  const p = 2 * lig - q
  const channel = (t) => {
    let x = t
    if (x < 0) x += 1
    if (x > 1) x -= 1
    if (x < 1 / 6) return p + (q - p) * 6 * x
    if (x < 1 / 2) return q
    if (x < 2 / 3) return p + (q - p) * (2 / 3 - x) * 6
    return p
  }

  return {
    r: channel(hue + 1 / 3) * 255,
    g: channel(hue) * 255,
    b: channel(hue - 1 / 3) * 255,
  }
}

/**
 * RGB 线性空间混合：`color × (1 - weight) + target × weight`（design D3）。
 * 与 Element Plus SCSS 的 mix() 口径一致，返回 hex。非法入参原样返回 color。
 */
export function mixHex(color, target, weight) {
  const c = hexToRgb(color)
  const t = hexToRgb(target)
  if (!c || !t) return color
  const w = clamp01(weight)
  return rgbToHex({
    r: c.r * (1 - w) + t.r * w,
    g: c.g * (1 - w) + t.g * w,
    b: c.b * (1 - w) + t.b * w,
  })
}

// rgbaOf 把 hex 转成带 alpha 的 rgba() 字符串。
// rgba() 不接受 CSS 变量作为色值分量，故 scrim 色必须预派生后整体注入（task 5.2）。
function rgbaOf(hex, alpha) {
  const c = hexToRgb(hex)
  if (!c) return `rgba(31, 111, 235, ${alpha})`
  return `rgba(${c.r}, ${c.g}, ${c.b}, ${alpha})`
}

// clampPrimaryLightness 把过亮的主色明度钳到 0.5（design D4）。
function clampPrimaryLightness(hex) {
  const rgb = hexToRgb(hex)
  if (!rgb) return hex
  const hsl = rgbToHsl(rgb)
  if (hsl.l <= LIGHTNESS_CLAMP_THRESHOLD) return hex
  return rgbToHex(hslToRgb(hsl.h, hsl.s, LIGHTNESS_CLAMP_TARGET))
}

// sampleDominantColor 在 64×64 canvas 上采样图片主色，失败返回 null。
function sampleDominantColor(img) {
  try {
    const canvas = document.createElement('canvas')
    canvas.width = SAMPLE_SIZE
    canvas.height = SAMPLE_SIZE
    const ctx = canvas.getContext('2d')
    if (!ctx) return null

    ctx.drawImage(img, 0, 0, SAMPLE_SIZE, SAMPLE_SIZE)
    const { data } = ctx.getImageData(0, 0, SAMPLE_SIZE, SAMPLE_SIZE)

    // 色相分桶：桶内累计 RGB，最后取像素最多的桶做平均
    const buckets = new Map()
    for (let i = 0; i < data.length; i += 4) {
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      const a = data[i + 3]

      if (a < ALPHA_MIN) continue
      const hsl = rgbToHsl({ r, g, b })
      if (hsl.s < SATURATION_MIN) continue
      if (hsl.l < LIGHTNESS_MIN || hsl.l > LIGHTNESS_MAX) continue

      const index = Math.min(HUE_BUCKETS - 1, Math.floor(hsl.h * HUE_BUCKETS))
      const bucket = buckets.get(index) || { count: 0, r: 0, g: 0, b: 0 }
      bucket.count += 1
      bucket.r += r
      bucket.g += g
      bucket.b += b
      buckets.set(index, bucket)
    }

    let best = null
    buckets.forEach((bucket) => {
      if (!best || bucket.count > best.count) best = bucket
    })
    if (!best) return null

    return rgbToHex({
      r: best.r / best.count,
      g: best.g / best.count,
      b: best.b / best.count,
    })
  } catch (e) {
    // 跨域图片会 taint canvas，getImageData 抛 SecurityError → 回退蓝色
    return null
  }
}

/**
 * 从 logo 图片提取主色。
 *
 * @param {string} logoUrl logo 地址
 * @returns {Promise<string|null>} 主色 hex；无 logo / 加载失败 / 全像素被过滤时为 null
 */
export function extractDominantColor(logoUrl) {
  return new Promise((resolve) => {
    if (!logoUrl) {
      resolve(null)
      return
    }

    const img = new Image()
    // 跨域 logo 需服务端给 CORS 头才能读像素；没有则走 onerror 回退，不会报错。
    img.crossOrigin = 'anonymous'
    img.onload = () => resolve(sampleDominantColor(img))
    img.onerror = () => resolve(null)
    img.src = logoUrl
  })
}

/**
 * 从主色派生主题变量族：Element Plus 主色变体 + 登录页渐变 + 品牌区遮罩。
 *
 * @param {string} primaryHex 主色 hex
 * @returns {Object<string, string>} CSS 变量名 → 变量值
 */
export function deriveThemeVars(primaryHex) {
  const primary = clampPrimaryLightness(hexToRgb(primaryHex) ? primaryHex : DEFAULT_PRIMARY)
  const brandTo = mixHex(primary, BLACK, BRAND_TO_WEIGHT)

  return {
    '--el-color-primary': primary,
    '--el-color-primary-light-3': mixHex(primary, WHITE, LIGHT_3_WEIGHT),
    '--el-color-primary-light-5': mixHex(primary, WHITE, LIGHT_5_WEIGHT),
    '--el-color-primary-light-7': mixHex(primary, WHITE, LIGHT_7_WEIGHT),
    '--el-color-primary-light-8': mixHex(primary, WHITE, LIGHT_8_WEIGHT),
    '--el-color-primary-light-9': mixHex(primary, WHITE, LIGHT_9_WEIGHT),
    '--el-color-primary-dark-2': mixHex(primary, BLACK, DARK_2_WEIGHT),
    '--brand-from': primary,
    '--brand-to': brandTo,
    '--brand-scrim-from': rgbaOf(primary, SCRIM_ALPHA_FROM),
    '--brand-scrim-to': rgbaOf(brandTo, SCRIM_ALPHA_TO),
  }
}

// documentRoot 取 :root 元素，SSR / 非浏览器环境下返回 null（本模块不假设一定有 DOM）。
function documentRoot() {
  return typeof document === 'undefined' ? null : document.documentElement
}

/**
 * 把主题变量注入 :root，使 Element Plus 全站组件与登录页渐变跟随变化。
 */
export function applyThemeVars(vars) {
  const root = documentRoot()
  if (!root || !vars) return
  Object.keys(vars).forEach((key) => {
    root.style.setProperty(key, vars[key])
  })
}

/**
 * 清除已注入的主题变量，回到「未注入」状态——这是回退路径的实现方式。
 *
 * 为什么不注入一组默认值：桌面端与移动端登录页的渐变起点本来就不同色
 * （#1f6feb vs #1989fa）。若回退时统一注入 #1f6feb，移动端会在「无 logo」
 * 这一默认配置下发生视觉变化，proposal 的「无破坏性 / 零配置开箱即用」就不成立。
 * 清干净后由各端 CSS 的 var(--x, 默认值) fallback 兜底，回退结果与现状逐像素一致。
 *
 * 键集合从 deriveThemeVars 现取，避免另维护一份常量清单而与之漂移。
 */
export function clearThemeVars() {
  const root = documentRoot()
  if (!root) return
  Object.keys(deriveThemeVars(DEFAULT_PRIMARY)).forEach((key) => {
    root.style.removeProperty(key)
  })
}
