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
    site_logo: '/brand.svg'
  }
}

vi.mock('@/stores', () => ({
  useAuthStore: () => authState,
  useAppStore: () => appState
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
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
    authState.isAuthenticated = false
    authState.isAdmin = false
    authState.checkAuth.mockReset()

    appState.publicSettingsLoaded = true
    appState.fetchPublicSettings.mockReset()
    appState.siteLogo = ''
    appState.cachedPublicSettings = {
      home_content: '',
      site_logo: '/brand.svg'
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
    expect(wrapper.find('[data-test="locale-switcher"]').exists()).toBe(true)
  })

  it('shows iframe when home content is a URL', () => {
    appState.cachedPublicSettings = {
      home_content: 'https://example.com/landing',
      site_logo: '/brand.svg'
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
      site_logo: '/brand.svg'
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
})
