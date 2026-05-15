import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import HomeView from '../HomeView.vue'

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
      'home.landing.footer.columns.links.gpshop': '格品购物',
      'home.landing.footer.columns.links.gpci': '格品生图',
      'home.landing.nav.community': '技术社群',
      'home.landing.footer.columns.links.title': '链接',
      'home.landing.hero.subtitle': 'AI API接口服务 | 企业级安全接入 | 国内稳定快速响应'
    }[key] ?? key)
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

    expect(wrapper.text()).toContain('home.landing.hero.badge')
    expect(wrapper.text()).toContain('home.landing.overview.title')
    expect(wrapper.text()).toContain('AI API接口服务 | 企业级安全接入 | 国内稳定快速响应')
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

    expect(text).toContain('GPT-5.4')
    expect(text).not.toContain('GPT-5.4 mini')
    expect(text).toContain('GPT Image 2')

    expect(text).toContain('格品购物')
    expect(text).toContain('格品生图')
    expect(links.some((link) => link.attributes('href') === 'https://card.gepinkeji.com')).toBe(true)
    expect(links.some((link) => link.attributes('href') === 'https://chat.gepinkeji.com/')).toBe(true)
    expect(links.some((link) => link.text().includes('技术社群') && link.attributes('href') === '#footer')).toBe(true)

    expect(text).toContain('链接')
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
    expect(wrapper.text()).toContain('home.landing.console.enter')
  })

  it('renders homepage contact items as plain text when configured', () => {
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg',
      qq_group: '123456789',
      wechat_contact: 'sub2api_support',
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

    expect(wrapper.text()).toContain('home.landing.community.qrLabels.qq')
    expect(wrapper.text()).toContain('123456789')
    expect(wrapper.text()).toContain('home.landing.community.qrLabels.wechat')
    expect(wrapper.text()).toContain('sub2api_support')
    expect(wrapper.find('img[alt="home.landing.community.qrAlt.qq"]').exists()).toBe(false)
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
})
