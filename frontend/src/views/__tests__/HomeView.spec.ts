import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const authStoreMock = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  user: null as null | { email?: string },
  checkAuth: vi.fn()
}))

const appStoreMock = vi.hoisted(() => ({
  cachedPublicSettings: null as null | Record<string, string>,
  siteName: 'PrecisionAPI',
  siteLogo: '',
  siteSubtitle: 'One Key, All AI Models',
  docUrl: 'https://docs.example.com',
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn()
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreMock,
  useAppStore: () => appStoreMock
}))

const messages: Record<string, string> = {}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key
    })
  }
})

function createWrapper() {
  return mount(HomeView, {
    global: {
      stubs: {
        LocaleSwitcher: {
          template: '<div data-test="locale-switcher">LocaleSwitcher</div>'
        },
        Icon: {
          template: '<span data-test="icon" />'
        },
        'router-link': {
          props: ['to'],
          template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
        }
      }
    }
  })
}

describe('HomeView', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(() => ({
        matches: false,
        media: '',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn()
      }))
    })

    authStoreMock.isAuthenticated = false
    authStoreMock.isAdmin = false
    authStoreMock.user = null
    authStoreMock.checkAuth.mockReset()

    appStoreMock.cachedPublicSettings = null
    appStoreMock.siteName = 'PrecisionAPI'
    appStoreMock.siteLogo = ''
    appStoreMock.siteSubtitle = 'One Key, All AI Models'
    appStoreMock.docUrl = 'https://docs.example.com'
    appStoreMock.publicSettingsLoaded = true
    appStoreMock.fetchPublicSettings.mockReset()
  })

  it('renders iframe mode when custom home content is an external url', () => {
    appStoreMock.cachedPublicSettings = {
      home_content: 'https://landing.example.com'
    }

    const wrapper = createWrapper()

    const iframe = wrapper.get('iframe')
    expect(iframe.attributes('src')).toBe('https://landing.example.com')
  })

  it('renders the new Chinese hero and navigation content on default home page', () => {
    const wrapper = createWrapper()

    expect(wrapper.text()).toContain('首页')
    expect(wrapper.text()).toContain('定价')
    expect(wrapper.text()).toContain('API文档')
    expect(wrapper.text()).toContain('技术社群')
    expect(wrapper.text()).toContain('登录 / 注册')
    expect(wrapper.text()).toContain('注册即送 ￥5.00 体验额度')
    expect(wrapper.text()).toContain('GPAPI - 让 AI 开发告别网络烦恼')
    expect(wrapper.text()).toContain('企业级 OpenAI 私有代理中转 | 国内极速响应 | 一行代码接入')
    expect(wrapper.text()).toContain('免费注册领 ￥5')
    expect(wrapper.text()).toContain('查看接入文档')
    expect(wrapper.find('[data-test="locale-switcher"]').exists()).toBe(true)
  })

  it('renders the requested landing page sections and commercial copy', () => {
    const wrapper = createWrapper()
    const pricingSection = wrapper.get('#pricing')
    const pricingGrid = pricingSection.get('.grid')
    const pricingCards = wrapper.findAll('[data-test="pricing-card"]')
    const benefitTags = wrapper.findAll('[data-test="benefit-tag"]')
    const overviewSection = wrapper.get('#overview')

    expect(wrapper.text()).toContain('选 GPAPI 中转站的六大好处')
    expect(wrapper.text()).not.toContain('面向对象')
    expect(wrapper.text()).toContain('Windows')
    expect(wrapper.text()).toContain('GPT-4o mini')
    expect(wrapper.text()).not.toContain('后续规划接入')
    expect(wrapper.text()).not.toContain('Claude 3.5 系列（Anthropic）')
    expect(wrapper.text()).not.toContain('Trae')
    expect(wrapper.text()).toContain('个人体验版')
    expect(wrapper.text()).toContain('基础订阅')
    expect(wrapper.text()).toContain('标准订阅')
    expect(wrapper.text()).toContain('高级订阅')
    expect(wrapper.text()).toContain('商务定制版')
    expect(wrapper.text()).toContain('￥19.9/月')
    expect(wrapper.text()).toContain('每日 $5.00 额度')
    expect(wrapper.text()).toContain('￥99/月')
    expect(wrapper.text()).toContain('每日 $22.00 额度')
    expect(wrapper.text()).toContain('￥299/月')
    expect(wrapper.text()).toContain('每日 $68.00 额度')
    expect(wrapper.text()).toContain('当前页面即可横向看完全部套餐')
    expect(wrapper.text()).toContain('最适合大多数开发者')
    expect(pricingCards).toHaveLength(5)
    expect(pricingCards.every((card) => card.attributes('href') === '/purchase')).toBe(true)
    expect(pricingSection.classes()).toContain('max-w-7xl')
    expect(pricingGrid.classes()).toContain('xl:grid-cols-5')
    expect(pricingSection.find('table').exists()).toBe(false)
    expect(pricingSection.get('[data-plan-highlight="true"]').text()).toContain('最受欢迎')
    expect(pricingSection.get('[data-plan-highlight="true"]').text()).toContain('标准订阅')
    expect(pricingSection.get('[data-plan-highlight="true"]').text()).toContain('￥99/月')
    expect(benefitTags).toHaveLength(6)
    expect(overviewSection.find('table').exists()).toBe(false)
    expect(wrapper.text()).toContain('三步完成模型接入')
    expect(wrapper.text()).toContain('https://api.gpapi.com/v1')
    expect(wrapper.text()).toContain('sk-gpapi-你的密钥')
  })

  it('routes homepage documentation links to /guide', () => {
    const wrapper = createWrapper()

    expect(wrapper.get('a[href="/guide"]').text()).toContain('API文档')
    expect(wrapper.findAll('a[href="/guide"]').some((link) => link.text().includes('查看接入文档'))).toBe(true)
    expect(wrapper.findAll('a[href="/guide"]').some((link) => link.text().includes('查看完整 API 文档'))).toBe(true)
  })

  it('renders the ICP record number as a MIIT link', () => {
    const wrapper = createWrapper()

    const icpLink = wrapper.get('a[href="https://beian.miit.gov.cn/"]')
    expect(icpLink.text()).toContain('蜀ICP备17044249号-1')
    expect(icpLink.attributes('target')).toBe('_blank')
    expect(icpLink.attributes('rel')).toBe('noopener noreferrer')
  })

  it('keeps the footer content centered on desktop layouts', () => {
    const wrapper = createWrapper()

    const footerContainer = wrapper.get('footer > div')

    expect(footerContainer.classes()).toContain('justify-center')
    expect(footerContainer.classes()).not.toContain('sm:text-left')
  })

  it('computes main and anchor offset from the fixed header height', () => {
    const originalGetBoundingClientRect = HTMLElement.prototype.getBoundingClientRect
    HTMLElement.prototype.getBoundingClientRect = vi.fn(function (this: HTMLElement) {
      if (this.tagName === 'HEADER') {
        return {
          width: 1200,
          height: 96,
          top: 0,
          left: 0,
          right: 1200,
          bottom: 96,
          x: 0,
          y: 0,
          toJSON: () => ({})
        } as DOMRect
      }

      return originalGetBoundingClientRect.call(this)
    })

    const wrapper = createWrapper()
    const main = wrapper.get('main')
    const overviewSection = wrapper.get('#overview')

    expect(main.attributes('style')).toContain('padding-top: 120px;')
    expect(overviewSection.attributes('style')).toContain('scroll-margin-top: 120px;')

    HTMLElement.prototype.getBoundingClientRect = originalGetBoundingClientRect
  })
})
