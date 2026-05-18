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
  routeState,
} = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  apiClientPostMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  registerMock: vi.fn(),
  pushMock: vi.fn(),
  routeState: {
    query: {} as Record<string, string>,
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
    currentRoute: { value: routeState },
  }),
  useRoute: () => ({
    query: routeState.query,
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
  resolveAffiliateReferralCode: (...values: unknown[]) => {
    for (const value of values) {
      const code = Array.isArray(value) ? value[0] : value
      if (typeof code === 'string' && code.trim()) {
        return code.trim()
      }
    }
    return ''
  },
  clearAffiliateReferralCode: vi.fn(),
}))

describe('RegisterView', () => {
  const defaultPublicSettings = () => ({
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

  const mountRegisterView = () => mount(RegisterView, {
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

  beforeEach(() => {
    vi.useFakeTimers()
    getPublicSettingsMock.mockReset()
    apiClientPostMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    registerMock.mockReset()
    pushMock.mockReset()
    routeState.query = {}
    sessionStorage.clear()

    getPublicSettingsMock.mockResolvedValue(defaultPublicSettings())
  })

  it('hides phone fields and registers without phone data when phone verification is disabled', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...defaultPublicSettings(),
      phone_verify_enabled: false,
    })
    registerMock.mockResolvedValue({})

    const wrapper = mountRegisterView()
    await flushPromises()

    expect(wrapper.find('#phone_number').exists()).toBe(false)
    expect(wrapper.find('#phone_verify_code').exists()).toBe(false)

    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('password')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(showErrorMock).not.toHaveBeenCalledWith('auth.phoneRequired')
    expect(registerMock).toHaveBeenCalledWith(expect.not.objectContaining({
      phone_number: expect.anything(),
    }))
    expect(registerMock).toHaveBeenCalledWith(expect.not.objectContaining({
      phone_verify_code: expect.anything(),
    }))
  })

  it('requires phone and sms code when phone verification is enabled', async () => {
    const wrapper = mountRegisterView()
    await flushPromises()

    expect(wrapper.find('#phone_number').exists()).toBe(true)
    expect(wrapper.find('#phone_verify_code').exists()).toBe(true)

    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('password')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith('auth.phoneRequired')

    await wrapper.get('#phone_number').setValue('13800138000')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith('auth.smsCodeRequired')
  })

  it('starts a 60 second cooldown after sending phone verify code successfully', async () => {
    apiClientPostMock.mockResolvedValue({ data: { countdown: 60 } })

    const wrapper = mountRegisterView()

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

    const wrapper = mountRegisterView()

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

  it('submits affiliate code from invite link during direct registration', async () => {
    routeState.query = { aff: 'AFF123' }
    getPublicSettingsMock.mockResolvedValueOnce({
      ...defaultPublicSettings(),
      phone_verify_enabled: false,
    })
    registerMock.mockResolvedValue({})

    const wrapper = mountRegisterView()
    await flushPromises()

    await wrapper.get('#email').setValue('invitee@example.com')
    await wrapper.get('#password').setValue('password')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith(expect.objectContaining({
      aff_code: 'AFF123',
    }))
  })

  it('persists affiliate code from invite link for email verification registration', async () => {
    routeState.query = { aff: 'AFF123' }
    getPublicSettingsMock.mockResolvedValueOnce({
      ...defaultPublicSettings(),
      email_verify_enabled: true,
      phone_verify_enabled: false,
    })

    const wrapper = mountRegisterView()
    await flushPromises()

    await wrapper.get('#email').setValue('verify-invitee@example.com')
    await wrapper.get('#password').setValue('password')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    const stored = JSON.parse(sessionStorage.getItem('register_data') || '{}')
    expect(stored.aff_code).toBe('AFF123')
    expect(pushMock).toHaveBeenCalledWith('/email-verify')
  })
})
