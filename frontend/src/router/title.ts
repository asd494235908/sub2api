import { i18n } from '@/i18n'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import type { CustomMenuItem } from '@/types'

/**
 * 统一生成页面标题，避免多处写入 document.title 产生覆盖冲突。
 * 优先使用 titleKey 通过 i18n 翻译，fallback 到静态 routeTitle。
 */
export function resolveDocumentTitle(
  routeTitle: unknown,
  siteName?: string,
  titleKey?: string,
  titleAbsolute = false
): string {
  const normalizedSiteName = typeof siteName === 'string' && siteName.trim() ? siteName.trim() : 'GPTK'

  if (typeof titleKey === 'string' && titleKey.trim()) {
    const translated = i18n.global.t(titleKey, { siteName: normalizedSiteName })
    if (translated && translated !== titleKey) {
      if (titleAbsolute) {
        return translated
      }
      return `${translated} - ${normalizedSiteName}`
    }
  }

  if (typeof routeTitle === 'string' && routeTitle.trim()) {
    return `${routeTitle.trim()} - ${normalizedSiteName}`
  }

  return normalizedSiteName
}

export function resolveDocumentDescription(routeDescription: unknown, descriptionKey?: string): string {
  if (typeof descriptionKey === 'string' && descriptionKey.trim()) {
    const translated = i18n.global.t(descriptionKey)
    if (translated && translated !== descriptionKey) {
      return translated
    }
  }

  return typeof routeDescription === 'string' ? routeDescription.trim() : ''
}

export function syncDocumentHead(
  meta: {
    title?: unknown
    titleKey?: string
    titleAbsolute?: boolean
    description?: unknown
    descriptionKey?: string
  },
  siteName?: string
): void {
  document.title = resolveDocumentTitle(meta.title, siteName, meta.titleKey, meta.titleAbsolute)

  const description = resolveDocumentDescription(meta.description, meta.descriptionKey)
  if (!description) {
    return
  }

  let tag = document.querySelector<HTMLMetaElement>('meta[name="description"]')
  if (!tag) {
    tag = document.createElement('meta')
    tag.name = 'description'
    document.head.appendChild(tag)
  }
  tag.content = description
}

export function resolveRouteDocumentTitle(
  route: Pick<RouteLocationNormalizedLoaded, 'name' | 'params' | 'meta'>,
  siteName: string | undefined,
  customMenuItems: CustomMenuItem[] = [],
): string {
  const id = typeof route.params.id === 'string' ? route.params.id : ''
  const menuItem = route.name === 'CustomPage' && id
    ? customMenuItems.find((item) => item.id === id)
    : undefined
  const menuTitle = menuItem?.label.trim()

  return resolveDocumentTitle(menuTitle || route.meta.title, siteName, menuTitle ? undefined : route.meta.titleKey as string)
}
