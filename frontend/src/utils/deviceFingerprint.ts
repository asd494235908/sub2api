import type { DeviceFingerprint } from '@/types'

function normalizeComponent(value: unknown): string {
  if (value == null) return ''
  if (Array.isArray(value)) return value.join(',')
  return String(value)
}

async function sha256(value: string): Promise<string> {
  const input = new TextEncoder().encode(value)
  if (globalThis.crypto?.subtle) {
    const digest = await globalThis.crypto.subtle.digest('SHA-256', input)
    return Array.from(new Uint8Array(digest))
      .map((part) => part.toString(16).padStart(2, '0'))
      .join('')
  }

  let hash = 2166136261
  for (let i = 0; i < value.length; i += 1) {
    hash ^= value.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return (hash >>> 0).toString(16).padStart(8, '0')
}

async function collectCanvasHash(): Promise<string> {
  try {
    const canvas = document.createElement('canvas')
    canvas.width = 280
    canvas.height = 80
    const ctx = canvas.getContext('2d')
    if (!ctx) return ''

    ctx.textBaseline = 'top'
    ctx.font = '16px Arial'
    ctx.fillStyle = '#f60'
    ctx.fillRect(10, 8, 90, 22)
    ctx.fillStyle = '#069'
    ctx.fillText('Sub2API fingerprint 2026', 12, 12)
    ctx.fillStyle = 'rgba(102, 204, 0, 0.7)'
    ctx.fillText('邀请身份', 18, 42)
    return sha256(canvas.toDataURL())
  } catch {
    return ''
  }
}

async function collectWebGLHash(): Promise<{ hash: string; components: Record<string, string> }> {
  const components: Record<string, string> = {}
  try {
    const canvas = document.createElement('canvas')
    const gl =
      canvas.getContext('webgl') ||
      canvas.getContext('experimental-webgl')
    if (!gl) {
      return { hash: '', components }
    }

    const context = gl as WebGLRenderingContext
    const debugInfo = context.getExtension('WEBGL_debug_renderer_info')
    if (debugInfo) {
      components.webgl_vendor = normalizeComponent(
        context.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL)
      )
      components.webgl_renderer = normalizeComponent(
        context.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL)
      )
    }
    components.webgl_version = normalizeComponent(context.getParameter(context.VERSION))
    components.webgl_shading_language = normalizeComponent(
      context.getParameter(context.SHADING_LANGUAGE_VERSION)
    )
    components.webgl_extensions = normalizeComponent(context.getSupportedExtensions() || [])

    return { hash: await sha256(JSON.stringify(components)), components }
  } catch {
    return { hash: '', components }
  }
}

export async function collectDeviceFingerprint(): Promise<DeviceFingerprint> {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return { composite_hash: '', canvas_hash: '', webgl_hash: '', components: {} }
  }
  if (import.meta.env.MODE === 'test' || /jsdom/i.test(navigator.userAgent || '')) {
    return { composite_hash: '', canvas_hash: '', webgl_hash: '', components: {} }
  }

  const [canvasHash, webgl] = await Promise.all([
    collectCanvasHash(),
    collectWebGLHash(),
  ])

  const components: Record<string, string> = {
    user_agent: normalizeComponent(navigator.userAgent),
    language: normalizeComponent(navigator.language),
    languages: normalizeComponent(navigator.languages),
    platform: normalizeComponent(navigator.platform),
    hardware_concurrency: normalizeComponent(navigator.hardwareConcurrency),
    device_memory: normalizeComponent((navigator as Navigator & { deviceMemory?: number }).deviceMemory),
    timezone: normalizeComponent(Intl.DateTimeFormat().resolvedOptions().timeZone),
    screen: normalizeComponent(`${screen.width}x${screen.height}x${screen.colorDepth}`),
    pixel_ratio: normalizeComponent(window.devicePixelRatio),
    touch_points: normalizeComponent(navigator.maxTouchPoints),
    canvas_hash: canvasHash,
    webgl_hash: webgl.hash,
    ...webgl.components,
  }

  const compositeHash = await sha256(JSON.stringify(components))

  return {
    composite_hash: compositeHash,
    canvas_hash: canvasHash,
    webgl_hash: webgl.hash,
    components,
  }
}

export function deviceFingerprintPayload(
  fingerprint: DeviceFingerprint | null | undefined
): { device_fingerprint?: DeviceFingerprint } {
  return fingerprint?.composite_hash ? { device_fingerprint: fingerprint } : {}
}
