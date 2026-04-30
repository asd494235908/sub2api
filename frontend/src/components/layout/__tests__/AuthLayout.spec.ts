import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AuthLayout from '../AuthLayout.vue'

const appStoreMock = vi.hoisted(() => ({
  siteName: '格品API',
  siteLogo: '',
  cachedPublicSettings: {
    site_subtitle: 'Subscription to API Conversion Platform'
  },
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn()
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreMock
}))

vi.mock('@/utils/url', () => ({
  sanitizeUrl: (url: string) => url
}))

describe('AuthLayout', () => {
  beforeEach(() => {
    appStoreMock.siteName = '格品API'
    appStoreMock.siteLogo = ''
    appStoreMock.cachedPublicSettings = {
      site_subtitle: 'Subscription to API Conversion Platform'
    }
    appStoreMock.publicSettingsLoaded = true
    appStoreMock.fetchPublicSettings.mockReset()
  })

  it('renders the ICP record as a link to MIIT', () => {
    const wrapper = mount(AuthLayout, {
      slots: {
        default: '<div>content</div>'
      }
    })

    const icpLink = wrapper.find('a[href="https://beian.miit.gov.cn/"]')

    expect(icpLink.exists()).toBe(true)
    expect(icpLink.text()).toContain('蜀ICP备17044249号-1')
    expect(icpLink.attributes('target')).toBe('_blank')
    expect(icpLink.attributes('rel')).toBe('noopener noreferrer')
  })
})
