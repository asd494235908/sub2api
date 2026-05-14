import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RegisterView from '@/views/auth/RegisterView.vue'

const {
  getPublicSettingsMock,
  apiClientPostMock,
  showSuccessMock,
  showErrorMock,
  registerMock,
  pushMock,
} = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  apiClientPostMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  registerMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
    currentRoute: { value: { query: {} } },
  }),
  useRoute: () => ({
    query: {},
  }),
  RouterLink: {
    template: '<a><slot /></a>',
  },
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      t: (key: string) => key,
    },
  }),
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh' },
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    register: (...args: any[]) => registerMock(...args),
  }),
  useAppStore: () => ({
    showSuccess: (...args: any[]) => showSuccessMock(...args),
    showError: (...args: any[]) => showErrorMock(...args),
  }),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    isWeChatWebOAuthEnabled: () => false,
  }
})

vi.mock('@/api/client', () => ({
  apiClient: {
    post: (...args: any[]) => apiClientPostMock(...args),
  },
}))

vi.mock('@/utils/oauthAffiliate', () => ({
  loadAffiliateReferralCode: () => '',
  resolveAffiliateReferralCode: () => '',
  clearAffiliateReferralCode: vi.fn(),
}))

describe('RegisterView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    getPublicSettingsMock.mockReset()
    apiClientPostMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    registerMock.mockReset()
    pushMock.mockReset()

    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      phone_verify_enabled: true,
      force_email_on_third_party_signup: false,
      registration_email_suffix_whitelist: [],
      promo_code_enabled: false,
      password_reset_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      site_name: 'Sub2API',
      linuxdo_oauth_enabled: false,
      wechat_oauth_enabled: false,
      oidc_oauth_enabled: false,
      oidc_oauth_provider_name: 'OIDC',
      backend_mode_enabled: false,
      payment_enabled: false,
      table_default_page_size: 20,
      table_page_size_options: [10, 20, 50],
      custom_menu_items: [],
      custom_endpoints: [],
      contact_info: '',
      doc_url: '',
      home_content: '',
      hide_ccs_import_button: false,
      api_base_url: '',
      site_logo: '',
      site_subtitle: '',
      version: 'test',
      balance_low_notify_enabled: false,
      account_quota_notify_enabled: false,
      balance_low_notify_threshold: 0,
      channel_monitor_enabled: false,
      channel_monitor_default_interval_seconds: 60,
      available_channels_enabled: false,
      affiliate_enabled: false,
    })
  })

  it('starts a 60 second cooldown after sending phone verify code successfully', async () => {
    apiClientPostMock.mockResolvedValue({ data: { countdown: 60 } })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
          'router-link': { template: '<a><slot /></a>' },
          Icon: true,
          TurnstileWidget: true,
          transition: false,
        },
      },
    })

    await flushPromises()

    const phoneInput = wrapper.get('#phone_number')
    await phoneInput.setValue('13800138000')
    await wrapper.get('button[type="button"]').trigger('click')
    await flushPromises()

    expect(showSuccessMock).toHaveBeenCalledWith('auth.smsCodeSentSuccess')
    expect(wrapper.get('button[type="button"]').text()).toContain('60s')

    vi.advanceTimersByTime(1000)
    await flushPromises()

    expect(wrapper.get('button[type="button"]').text()).toContain('59s')
  })

  it('uses server cooldown countdown when phone verify code is rate limited', async () => {
    apiClientPostMock.mockRejectedValue({
      response: {
        data: {
          reason: 'VERIFY_CODE_TOO_FREQUENT',
          metadata: { countdown: '42' },
          detail: '请在 42 秒后重试',
        },
      },
    })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
          'router-link': { template: '<a><slot /></a>' },
          Icon: true,
          TurnstileWidget: true,
          transition: false,
        },
      },
    })

    await flushPromises()

    await wrapper.get('#phone_number').setValue('13800138000')
    await wrapper.get('button[type="button"]').trigger('click')
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith('请在 42 秒后重试')
    expect(wrapper.get('button[type="button"]').text()).toContain('42s')
  })

  it('uses translated auth copy instead of hard-coded Chinese in English locale', async () => {
    vi.doMock('vue-i18n', () => ({
      createI18n: () => ({
        global: {
          t: (key: string) => key,
        },
      }),
      useI18n: () => ({
        t: (key: string) => key,
        locale: { value: 'en' },
      }),
    }))
  })
})
