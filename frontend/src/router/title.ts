import { i18n } from '@/i18n'

type HeadMetaLike = {
  title?: unknown
  titleKey?: string
  titleAbsolute?: boolean
  description?: unknown
  descriptionKey?: string
}

/**
 * 统一生成页面标题，避免多处写入 document.title 产生覆盖冲突。
 * 优先使用 titleKey 通过 i18n 翻译，fallback 到静态 routeTitle。
 */
export function resolveDocumentTitle(routeTitle: unknown, siteName?: string, titleKey?: string, titleAbsolute = false): string {
  const normalizedSiteName = typeof siteName === 'string' && siteName.trim() ? siteName.trim() : 'GPTK'

  if (typeof titleKey === 'string' && titleKey.trim()) {
    const translated = i18n.global.t(titleKey, { siteName: normalizedSiteName })
    if (translated && translated !== titleKey) {
      if (titleAbsolute) {
        return String(translated).trim()
      }
      return `${translated} - ${normalizedSiteName}`
    }
  }

  if (typeof routeTitle === 'string' && routeTitle.trim()) {
    return `${routeTitle.trim()} - ${normalizedSiteName}`
  }

  return normalizedSiteName
}

/**
 * 统一解析页面描述，优先使用 i18n key，fallback 到静态 route description。
 */
export function resolveDocumentDescription(routeDescription: unknown, descriptionKey?: string): string {
  if (typeof descriptionKey === 'string' && descriptionKey.trim()) {
    const translated = i18n.global.t(descriptionKey)
    if (translated && translated !== descriptionKey) {
      return String(translated).trim()
    }
  }

  if (typeof routeDescription === 'string' && routeDescription.trim()) {
    return routeDescription.trim()
  }

  return ''
}

function ensureMetaDescriptionTag(): HTMLMetaElement | null {
  if (typeof document === 'undefined') {
    return null
  }

  let tag = document.querySelector<HTMLMetaElement>('meta[name="description"]')
  if (!tag) {
    tag = document.createElement('meta')
    tag.setAttribute('name', 'description')
    document.head.appendChild(tag)
  }

  return tag
}

/**
 * 统一同步页面 title 和 meta description，供路由切换与初始化复用。
 */
export function syncDocumentHead(meta: HeadMetaLike, siteName?: string): void {
  if (typeof document === 'undefined') {
    return
  }

  document.title = resolveDocumentTitle(meta.title, siteName, meta.titleKey, meta.titleAbsolute === true)

  const description = resolveDocumentDescription(meta.description, meta.descriptionKey)
  const metaTag = ensureMetaDescriptionTag()

  if (!metaTag) {
    return
  }

  if (description) {
    metaTag.setAttribute('content', description)
  } else {
    metaTag.removeAttribute('content')
  }
}
