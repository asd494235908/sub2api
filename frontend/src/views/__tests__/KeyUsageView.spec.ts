import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import KeyUsageView from '../KeyUsageView.vue'

const appStoreMock = vi.hoisted(() => ({
  cachedPublicSettings: null as null | Record<string, string>,
  siteName: '格品API',
  siteLogo: '',
  docUrl: 'https://docs.example.com',
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

const messages: Record<string, string> = {
  'home.viewDocs': 'View Documentation',
  'home.switchToLight': 'Switch to Light Mode',
  'home.switchToDark': 'Switch to Dark Mode',
  'home.docs': 'Documentation',
  'keyUsage.title': 'Key Usage Query',
  'keyUsage.subtitle': 'Inspect usage and remaining quota for your key.',
  'keyUsage.placeholder': 'Enter API key',
  'keyUsage.query': 'Query',
  'keyUsage.querying': 'Querying',
  'keyUsage.privacyNote': 'Your key is only used for this query.',
  'keyUsage.dateRange': 'Date Range',
  'keyUsage.dateRangeToday': 'Today',
  'keyUsage.dateRange7d': 'Last 7 Days',
  'keyUsage.dateRange30d': 'Last 30 Days',
  'keyUsage.dateRangeCustom': 'Custom',
  'keyUsage.apply': 'Apply'
}

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreMock
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
      locale: { value: 'en' }
    })
  }
})

describe('KeyUsageView', () => {
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

    appStoreMock.cachedPublicSettings = null
    appStoreMock.siteName = '格品API'
    appStoreMock.siteLogo = ''
    appStoreMock.docUrl = 'https://docs.example.com'
    appStoreMock.publicSettingsLoaded = true
    appStoreMock.fetchPublicSettings.mockReset()
    appStoreMock.showSuccess.mockReset()
    appStoreMock.showError.mockReset()
  })

  it('renders the ICP record as a link to MIIT', () => {
    const wrapper = mount(KeyUsageView, {
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

    const icpLink = wrapper.find('a[href="https://beian.miit.gov.cn/"]')

    expect(icpLink.exists()).toBe(true)
    expect(icpLink.text()).toContain('蜀ICP备17044249号-1')
    expect(icpLink.attributes('target')).toBe('_blank')
    expect(icpLink.attributes('rel')).toBe('noopener noreferrer')
  })

  it('keeps the footer content centered on desktop layouts', () => {
    const wrapper = mount(KeyUsageView, {
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

    const footerContainer = wrapper.get('footer > div')

    expect(footerContainer.classes()).toContain('justify-center')
    expect(footerContainer.classes()).not.toContain('sm:text-left')
  })

  it('uses the external token documentation link in a new window', () => {
    const wrapper = mount(KeyUsageView, {
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

    const docLinks = wrapper.findAll('a[href="https://token.gepinkeji.com/tokenDoc"]')

    expect(docLinks.length).toBeGreaterThan(0)
    expect(docLinks.every((link) => link.attributes('target') === '_blank')).toBe(true)
    expect(docLinks.every((link) => link.attributes('rel') === 'noopener noreferrer')).toBe(true)
  })
})
