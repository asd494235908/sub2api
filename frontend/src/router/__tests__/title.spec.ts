import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/i18n', () => ({
  i18n: {
    global: {
      t: (key: string, params?: Record<string, string>) => {
        const siteName = params?.siteName ?? 'GPTK'
        return ({
          'home.landing.seo.title': `AI API接口服务平台 - 企业级安全接入 | ${siteName}`,
          'home.landing.seo.description':
            'GPTK 提供面向开发者与企业的 AI API 接口服务，支持企业级安全接入、国内稳定快速响应与便捷集成，适合 AI 应用开发、测试与生产环境落地。',
          'keys.description': '管理和查看你的 API 密钥'
        } as Record<string, string>)[key] ?? key
      }
    }
  }
}))

import * as titleModule from '@/router/title'
import { resolveRouteDocumentTitle } from '@/router/title'

const { resolveDocumentTitle } = titleModule

describe('resolveDocumentTitle', () => {
  beforeEach(() => {
    document.title = ''
    document.head.innerHTML = ''
  })

  it('路由存在标题时，使用“路由标题 - 站点名”格式', () => {
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', () => {
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，回退默认站点名', () => {
    expect(resolveDocumentTitle('Dashboard', '')).toBe('Dashboard - GPTK')
    expect(resolveDocumentTitle(undefined, '   ')).toBe('GPTK')
  })

  it('站点名变更时仅影响后续路由标题计算', () => {
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })

  it('首页 SEO 标题支持使用完整标题且不重复追加站点名', () => {
    expect((resolveDocumentTitle as any)(undefined, 'GPTK', 'home.landing.seo.title', true)).toBe(
      'AI API接口服务平台 - 企业级安全接入 | GPTK'
    )
  })

  it('支持从 i18n key 解析页面描述', () => {
    const resolveDocumentDescription = (titleModule as any).resolveDocumentDescription

    expect(resolveDocumentDescription).toBeTypeOf('function')
    expect(resolveDocumentDescription(undefined, 'home.landing.seo.description')).toBe(
      'GPTK 提供面向开发者与企业的 AI API 接口服务，支持企业级安全接入、国内稳定快速响应与便捷集成，适合 AI 应用开发、测试与生产环境落地。'
    )
  })

  it('切换路由时会同步更新 title 和 description', () => {
    const syncDocumentHead = (titleModule as any).syncDocumentHead

    expect(syncDocumentHead).toBeTypeOf('function')

    syncDocumentHead(
      {
        titleKey: 'home.landing.seo.title',
        descriptionKey: 'home.landing.seo.description',
        titleAbsolute: true
      },
      'GPTK'
    )

    expect(document.title).toBe('AI API接口服务平台 - 企业级安全接入 | GPTK')
    expect(document.querySelector('meta[name="description"]')?.getAttribute('content')).toBe(
      'GPTK 提供面向开发者与企业的 AI API 接口服务，支持企业级安全接入、国内稳定快速响应与便捷集成，适合 AI 应用开发、测试与生产环境落地。'
    )

    syncDocumentHead(
      {
        title: 'Usage Records',
        descriptionKey: 'keys.description'
      },
      'GPTK'
    )

    expect(document.title).toBe('Usage Records - GPTK')
    expect(document.querySelector('meta[name="description"]')?.getAttribute('content')).toBe('管理和查看你的 API 密钥')
  })
})

describe('resolveRouteDocumentTitle', () => {
  it('自定义页面菜单加载后，使用菜单名称作为标题', () => {
    const route = {
      name: 'CustomPage',
      params: { id: 'scheduler' },
      meta: {
        title: 'Custom Page'
      }
    }

    expect(resolveRouteDocumentTitle(route, 'EzouAPI')).toBe('Custom Page - EzouAPI')
    expect(resolveRouteDocumentTitle(route, 'EzouAPI', [
      {
        id: 'scheduler',
        label: '账号调度器',
        icon_svg: '',
        url: 'https://example.com',
        visibility: 'admin',
        sort_order: 0
      }
    ])).toBe('账号调度器 - EzouAPI')
  })
})
