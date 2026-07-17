import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import HomeView from '../HomeView.vue'
import enMessages from '@/i18n/locales/en'
import zhMessages from '@/i18n/locales/zh'
import { TOKEN_DOC_URL } from '@/constants/externalLinks'

const authState = {
  isAuthenticated: false,
  isAdmin: false,
  checkAuth: vi.fn()
}

const appState = {
  siteLogo: '',
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn(),
  cachedPublicSettings: {
    home_content: '',
    site_logo: '/brand.svg',
    qq_group: '',
    wechat_contact: '',
    home_links: []
  }
}

vi.mock('@/stores', () => ({
  useAuthStore: () => authState,
  useAppStore: () => appState
}))

let currentLocale = 'zh'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale: { value: currentLocale },
    t: (key: string) => ({
      'home.landing.nav.gpshop': '格品购物',
      'home.landing.nav.gpci': '格品生图',
      'home.landing.nav.community': '技术社群',
      'home.landing.footer.columns.links.title': '链接',
      'home.landing.footer.columns.company.title': '公司',
      'home.landing.footer.columns.company.about': '关于我们',
      'home.landing.footer.columns.company.contact': '联系我们',
      'home.landing.footer.columns.company.terms': '服务条款',
      'home.landing.footer.columns.company.privacy': '隐私政策',
      'home.landing.hero.badge': '稳定 · 安全 · 高效的国际中转服务',
      'home.landing.hero.titleLead': '让 ',
      'home.landing.hero.titleHighlight': 'AI 开发',
      'home.landing.hero.titleSuffix': '更简单',
      'home.landing.hero.subtitle': '企业级安全接入',
      'home.landing.console.enter': '进入控制台'
    }[key] ?? (key.startsWith('home.landing.') ? key.split('.').at(-1) || key : key))
  })
}))

vi.mock('@/components/common/LocaleSwitcher.vue', () => ({
  default: {
    name: 'LocaleSwitcher',
    template: '<div data-test="locale-switcher" />'
  }
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: {
    name: 'Icon',
    props: ['name'],
    template: '<span :data-icon="name" />'
  }
}))

describe('HomeView', () => {
  beforeEach(() => {
    currentLocale = 'zh'
    authState.isAuthenticated = false
    authState.isAdmin = false
    authState.checkAuth.mockReset()

    appState.publicSettingsLoaded = true
    appState.fetchPublicSettings.mockReset()
    appState.siteLogo = ''
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: []
    }

    localStorage.clear()
    document.documentElement.classList.remove('dark')
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn()
      }))
    })
  })

  it('renders the redesigned homepage by default', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('稳定 · 安全 · 高效的国际中转服务')
    expect(wrapper.text()).toContain('让 AI 开发更简单')
    expect(wrapper.text()).toContain('企业级安全接入')
    expect(wrapper.text()).toContain('import International from "international"')
    expect(wrapper.text()).not.toContain('import OpenAI from "openai"')
    expect(wrapper.text()).not.toContain('home.landing.')
    expect(wrapper.find('[data-test="locale-switcher"]').exists()).toBe(true)
  })

  it('renders updated home model, nav links, and footer links', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const text = wrapper.text()
    const links = wrapper.findAll('a')

    const supportedModels = [
      'DeepSeek V4-Pro',
      'DeepSeek V4-Flash',
      'Doubao-Seed_1.5',
      'Doubao-Seed_1.6',
      'Doubao-Seed_1.8',
      'Doubao-Seed_2.0',
      'Qwen3.5-Plus',
      'Qwen',
      'Happy horse',
      'Kling Video API',
      'MiniMax-M2.7',
      'MiniMax-M2.5',
      'GLM-5.1',
      'GLM-5-Turbo',
      'GLM-5'
    ]
    supportedModels.forEach((model) => expect(text).toContain(model))
    supportedModels.reduce((previousIndex, model) => {
      const currentIndex = text.indexOf(model, previousIndex + 1)
      expect(currentIndex).toBeGreaterThan(previousIndex)
      return currentIndex
    }, -1)
    expect(text).not.toContain('GPT-5.5')
    expect(text).not.toContain('GPT-5.4')
    expect(text).not.toContain('GPT Image 2')
    expect(text).not.toContain('Qwen3.5以前')
    expect(text).not.toContain('Qwen3.6以后')

    expect(text).toContain('格品购物')
    expect(text).toContain('格品生图')
    expect(links.some((link) => link.attributes('href') === 'https://card.gepinkeji.com')).toBe(true)
    expect(links.some((link) => link.attributes('href') === 'https://chat.gepinkeji.com/')).toBe(true)
    expect(links.some((link) => link.text().includes('技术社群') && link.attributes('href') === '#footer')).toBe(true)

    expect(text).toContain('链接')
    expect(text).toContain('公司')
    expect(text).toContain('关于我们')
    expect(text).toContain('联系我们')
    expect(text).toContain('服务条款')
    expect(text).toContain('隐私政策')
    expect(text).not.toContain('成都格品科技有限公司版权所有')
    expect(text).not.toContain('蜀ICP备17044249号-1')
    expect(text).not.toContain('© 2018')
    expect(text).toContain('格品购物')
    expect(text).toContain('格品生图')
    expect(text).not.toContain('home.landing.footer.columns.support.title')
    expect(text).not.toContain('home.landing.footer.columns.support.help')
    expect(text).not.toContain('home.landing.footer.columns.support.community')
  })

  it('shows iframe when home content is a URL', () => {
    appState.cachedPublicSettings = {
      home_content: 'https://example.com/landing',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: []
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: true
        }
      }
    })

    const frame = wrapper.get('iframe')
    expect(frame.attributes('src')).toBe('https://example.com/landing')
    expect(wrapper.text()).not.toContain('home.landing.hero.badge')
  })

  it('shows raw HTML when home content is custom markup', () => {
    appState.cachedPublicSettings = {
      home_content: '<section><h1>Custom Home</h1></section>',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: []
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: true
        }
      }
    })

    expect(wrapper.html()).toContain('<h1>Custom Home</h1>')
    expect(wrapper.find('iframe').exists()).toBe(false)
  })

  it('switches CTA target for authenticated admin users', () => {
    authState.isAuthenticated = true
    authState.isAdmin = true

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.some((link) => link.attributes('href') === '/admin/dashboard')).toBe(true)
    expect(wrapper.text()).toContain('进入控制台')
  })

  it('renders homepage contact items as plain text when configured', () => {
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '123456789',
      wechat_contact: 'sub2api_support',
      contact_info: 'qq群: 123456789 微信:sub2api_support',
      home_links: []
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('qq')
    expect(wrapper.text()).toContain('123456789')
    expect(wrapper.text()).toContain('wechat')
    expect(wrapper.text()).toContain('sub2api_support')
    expect(wrapper.text()).not.toContain('common.contact')
    expect(wrapper.text()).not.toContain('qq群: 123456789 微信:sub2api_support')
    expect(wrapper.find('img[alt="home.landing.community.qrAlt.qq"]').exists()).toBe(false)
  })

  it('renders newline-separated homepage contacts as trimmed rows', () => {
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: ' 123456789 \n\n987654321 ',
      wechat_contact: ' sub2api_support\nwechat_helper ',
      home_links: []
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const contactRows = wrapper.findAll('[data-test="footer-contact-line"]').map((node) => node.text())
    expect(contactRows).toEqual(['123456789', '987654321', 'sub2api_support', 'wechat_helper'])
    expect(contactRows).not.toContain('')
  })

  it('hides homepage contact section when both contacts are empty', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).not.toContain('home.landing.community.qrLabels.qq')
    expect(wrapper.text()).not.toContain('home.landing.community.qrLabels.wechat')
  })

  it('renders configured home links in top nav and footer links', () => {
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: [
        { id: 'disabled', label: '隐藏链接', label_zh: '隐藏链接', label_en: 'Hidden Link', url: 'https://hidden.example.com', enabled: false, sort_order: 1 },
        { id: 'docs2', label: '配置链接 B', label_zh: '配置链接乙', label_en: 'Configured Link B', url: 'https://b.example.com', enabled: true, sort_order: 2 },
        { id: 'docs1', label: '配置链接 A', label_zh: '配置链接甲', label_en: 'Configured Link A', url: 'https://a.example.com', enabled: true, sort_order: 0 }
      ]
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.filter((link) => link.text().includes('配置链接甲'))).toHaveLength(2)
    expect(links.filter((link) => link.text().includes('配置链接乙'))).toHaveLength(2)
    expect(links.some((link) => link.text().includes('隐藏链接'))).toBe(false)
    expect(links.some((link) => link.text().includes('配置链接甲') && link.attributes('href') === 'https://a.example.com')).toBe(true)
    expect(links.some((link) => link.text().includes('配置链接乙') && link.attributes('href') === 'https://b.example.com')).toBe(true)
  })

  it('renders English home link labels when locale is English', () => {
    currentLocale = 'en'
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: [
        { id: 'docs1', label: '配置链接 A', label_zh: '配置链接甲', label_en: 'Configured Link A', url: 'https://a.example.com', enabled: true, sort_order: 0 }
      ]
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.filter((link) => link.text().includes('Configured Link A'))).toHaveLength(2)
    expect(wrapper.text()).not.toContain('配置链接甲')
  })

  it('falls back to legacy home link label when localized labels are missing', () => {
    currentLocale = 'en'
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      home_links: [
        { id: 'legacy', label: 'Legacy Link', url: 'https://legacy.example.com', enabled: true, sort_order: 0 }
      ]
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.filter((link) => link.text().includes('Legacy Link'))).toHaveLength(2)
  })

  it('falls back to default home links when configured home links are empty', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.some((link) => link.text().includes('格品购物') && link.attributes('href') === 'https://card.gepinkeji.com')).toBe(true)
    expect(links.some((link) => link.text().includes('格品生图') && link.attributes('href') === 'https://chat.gepinkeji.com/')).toBe(true)
  })

  it('routes homepage documentation links to the token documentation page', () => {
    expect(TOKEN_DOC_URL).toBe('https://doc.gptk.cc.cd/')

    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '',
      wechat_contact: '',
      doc_url: '',
      home_links: []
    }

    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const docLinks = wrapper.findAll('a').filter((link) => link.text().includes('docs') || link.text().includes('API 文档'))
    expect(docLinks.length).toBeGreaterThan(0)
    expect(docLinks.every((link) => link.attributes('href') === TOKEN_DOC_URL)).toBe(true)
  })

  it('routes homepage company footer links to their section targets', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
          }
        }
      }
    })

    const links = wrapper.findAll('a')
    expect(links.find((link) => link.text().includes('关于我们'))?.attributes('href')).toBe('#overview')
    expect(links.find((link) => link.text().includes('联系我们'))?.attributes('href')).toBe('#footer')
    expect(links.find((link) => link.text().includes('服务条款'))?.attributes('href')).toBe(TOKEN_DOC_URL)
    expect(links.find((link) => link.text().includes('隐私政策'))?.attributes('href')).toBe(TOKEN_DOC_URL)
  })

  it('defines every home landing key used by the homepage in both locales', async () => {
    const source = await import('../HomeView.vue?raw')
    const keys = Array.from(source.default.matchAll(/t\('([^']+)'\)/g))
      .map((match) => match[1])
      .filter((key) => key.startsWith('home.landing.'))

    function lookup(messages: Record<string, any>, key: string) {
      return key.split('.').reduce<any>((value, part) => value?.[part], messages)
    }

    for (const key of new Set(keys)) {
      expect(lookup(zhMessages, key), `missing zh translation for ${key}`).toBeTruthy()
      expect(lookup(enMessages, key), `missing en translation for ${key}`).toBeTruthy()
    }
  })
})
