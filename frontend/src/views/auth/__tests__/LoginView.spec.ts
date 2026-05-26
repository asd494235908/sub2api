import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LoginView from '@/views/auth/LoginView.vue'

const {
  getPublicSettingsMock,
  buildCasdoorLoginUrlMock,
  authStoreLoginMock,
  showErrorMock,
  showSuccessMock,
  showWarningMock,
  pushMock,
  apiClientPostMock,
} = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  buildCasdoorLoginUrlMock: vi.fn(),
  authStoreLoginMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showWarningMock: vi.fn(),
  pushMock: vi.fn(),
  apiClientPostMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
    currentRoute: { value: { query: {} } },
  }),
  RouterLink: {
    props: ['to'],
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
  }),
}))

vi.mock('@/components/layout', () => ({
  AuthLayout: {
    template: '<div><slot /><slot name="footer" /></div>',
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    login: (...args: any[]) => authStoreLoginMock(...args),
    login2FA: vi.fn(),
  }),
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
    showSuccess: (...args: any[]) => showSuccessMock(...args),
    showWarning: (...args: any[]) => showWarningMock(...args),
  }),
}))

vi.mock('@/api/auth', () => {
  return {
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    buildCasdoorLoginUrl: (...args: any[]) => buildCasdoorLoginUrlMock(...args),
    isWeChatWebOAuthEnabled: () => false,
    isTotp2FARequired: (response: any) => response?.requires_2fa === true,
  }
})

vi.mock('@/api/client', () => ({
  apiClient: {
    post: (...args: any[]) => apiClientPostMock(...args),
  },
}))

vi.mock('@/utils/oauthAffiliate', () => ({
  clearAllAffiliateReferralCodes: vi.fn(),
}))

describe('LoginView', () => {
  const defaultPublicSettings = () => ({
    registration_enabled: true,
    email_verify_enabled: false,
    phone_verify_enabled: false,
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

  const mountLoginView = () => mount(LoginView, {
    global: {
      stubs: {
        AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
        LinuxDoOAuthSection: true,
        WechatOAuthSection: true,
        OidcOAuthSection: true,
        TotpLoginModal: true,
        'router-link': { template: '<a><slot /></a>' },
        Icon: true,
        TurnstileWidget: true,
      },
    },
  })

  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
    getPublicSettingsMock.mockResolvedValue(defaultPublicSettings())
    buildCasdoorLoginUrlMock.mockReturnValue('/api/v1/auth/casdoor/login?redirect=%2Fdashboard')
    authStoreLoginMock.mockResolvedValue({
      access_token: 'token',
      token_type: 'Bearer',
      user: {
        id: 1,
        username: 'user',
        email: 'user@example.com',
        role: 'user',
        balance: 0,
        concurrency: 5,
        status: 'active',
        allowed_groups: null,
        created_at: '',
        updated_at: '',
      },
    })
  })

  it('hides phone login controls when phone verification is disabled', async () => {
    const wrapper = mountLoginView()
    await flushPromises()

    expect(wrapper.get('label[for="identifier"]').text()).toBe('auth.emailLabel')
    expect(wrapper.get('#identifier').attributes('type')).toBe('email')
    expect(wrapper.find('#sms_code').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('auth.sendCode')
  })

  it('blocks numeric phone input when phone verification is disabled', async () => {
    const wrapper = mountLoginView()
    await flushPromises()

    await wrapper.get('#identifier').setValue('13800138000')
    await wrapper.get('#password').setValue('password123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(authStoreLoginMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('auth.invalidEmail')
  })

  it('logs in with email and password when phone verification is disabled', async () => {
    const wrapper = mountLoginView()
    await flushPromises()

    await wrapper.get('#identifier').setValue('user@example.com')
    await wrapper.get('#password').setValue('password123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(authStoreLoginMock).toHaveBeenCalledWith({
      identifier: 'user@example.com',
      email: 'user@example.com',
      phone_number: undefined,
      sms_code: undefined,
      password: 'password123',
      turnstile_token: undefined,
    })
  })

  it('shows a Casdoor login button that starts the isolated Casdoor flow', async () => {
    const wrapper = mountLoginView()
    await flushPromises()

    const casdoorLink = wrapper.get('[data-testid="casdoor-login-link"]')

    expect(casdoorLink.text()).toContain('Casdoor 登录')
    expect(casdoorLink.attributes('href')).toBe('/api/v1/auth/casdoor/login?redirect=%2Fdashboard')
    expect(buildCasdoorLoginUrlMock).toHaveBeenCalledWith('/dashboard')
  })

  it('shows sms code controls for phone login when phone verification is enabled', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...defaultPublicSettings(),
      phone_verify_enabled: true,
    })

    const wrapper = mountLoginView()
    await flushPromises()

    await wrapper.get('#identifier').setValue('13800138000')
    await flushPromises()

    expect(wrapper.get('label[for="identifier"]').text()).toBe('auth.identifierLabel')
    expect(wrapper.get('#identifier').attributes('type')).toBe('text')
    expect(wrapper.find('#sms_code').exists()).toBe(true)

    await wrapper.get('#sms_code').setValue('123456')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(authStoreLoginMock).toHaveBeenCalledWith({
      identifier: '13800138000',
      email: undefined,
      phone_number: '13800138000',
      sms_code: '123456',
      password: undefined,
      turnstile_token: undefined,
    })
  })
})
