import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, shallowMount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import PaymentView from '../PaymentView.vue'
import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'

const AppLayoutStub = defineComponent({
  setup(_, { slots }) {
    return () => h('div', slots.default?.())
  },
})

const routeState = vi.hoisted(() => ({
  path: '/purchase',
  query: {} as Record<string, unknown>,
}))

const routerReplace = vi.hoisted(() => vi.fn())
const routerPush = vi.hoisted(() => vi.fn())
const routerResolve = vi.hoisted(() => vi.fn(() => ({ href: '/payment/stripe?mock=1' })))
const createOrder = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const fetchActiveSubscriptions = vi.hoisted(() => vi.fn().mockResolvedValue(undefined))
const activeSubscriptionsState = vi.hoisted(() => ({
  value: [] as any[],
}))
const showError = vi.hoisted(() => vi.fn())
const showInfo = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())
const getCheckoutInfo = vi.hoisted(() => vi.fn())
const bridgeInvoke = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({
      replace: routerReplace,
      push: routerPush,
      resolve: routerResolve,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      username: 'demo-user',
      balance: 0,
    },
    refreshUser,
  }),
}))

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    createOrder,
  }),
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    activeSubscriptions: activeSubscriptionsState.value,
    fetchActiveSubscriptions,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showInfo,
    showWarning,
  }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getCheckoutInfo,
  },
}))

vi.mock('@/utils/device', () => ({
  isMobileDevice: () => true,
}))

function checkoutInfoFixture() {
  return {
    data: {
      methods: {
        wxpay: {
          daily_limit: 0,
          daily_used: 0,
          daily_remaining: 0,
          single_min: 0,
          single_max: 0,
          fee_rate: 0,
          available: true,
        },
      },
      global_min: 0,
      global_max: 0,
      plans: [],
      balance_disabled: false,
      balance_recharge_multiplier: 1,
      recharge_fee_rate: 0,
      help_text: '',
      help_image_url: '',
      stripe_publishable_key: '',
    },
  }
}

function checkoutInfoWithPlansFixture() {
  return {
    data: {
      ...checkoutInfoFixture().data,
      plans: [
        {
          id: 7,
          group_id: 3,
          name: 'Starter',
          description: '',
          price: 128,
          original_price: 0,
          validity_days: 30,
          validity_unit: 'day',
          rate_multiplier: 1,
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null,
          subscription_total_limit_usd: 1560,
          features: [],
          group_platform: 'openai',
          sort_order: 1,
          for_sale: true,
          daily_purchase_limit: 0,
          group_name: 'OpenAI',
          supported_models: ['gpt-5.3-codex', 'gpt-image-2', 'gpt-5.4-mini', 'gpt-5.4', 'gpt-5.5'],
        },
      ],
    },
  }
}

function checkoutInfoWithFirstRechargeFixture() {
  return {
    data: {
      ...checkoutInfoWithPlansFixture().data,
      balance_recharge_multiplier: 13,
      first_recharge: {
        enabled: true,
        tiers: [
          {
            id: 'tier-30',
            pay_amount: 30,
            bonus_amount: 15,
            enabled: true,
            sort_order: 1,
          },
        ],
      },
    },
  }
}

function checkoutInfoWithUnavailableDailyPlanFixture() {
  return {
    data: {
      ...checkoutInfoFixture().data,
      plans: [
        {
          id: 8,
          group_id: 3,
          name: 'Flash Sale',
          description: '',
          price: 128,
          original_price: 0,
          validity_days: 30,
          validity_unit: 'day',
          rate_multiplier: 1,
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null,
          features: [],
          group_platform: 'openai',
          sort_order: 1,
          for_sale: true,
          group_name: 'OpenAI',
          daily_purchase_limit: 2,
          daily_purchase_remaining: 0,
          daily_sale_starts_at: '09:00',
          daily_sale_ends_at: '18:00',
          daily_sale_status: 'sold_out',
          daily_sale_countdown_seconds: 3600,
          daily_sale_available_for_payment: false,
        },
      ],
    },
  }
}

function checkoutInfoWithAvailableDailyPlanFixture() {
  return {
    data: {
      ...checkoutInfoFixture().data,
      plans: [
        {
          id: 9,
          group_id: 3,
          name: 'Available Flash Sale',
          description: '',
          price: 128,
          original_price: 0,
          validity_days: 30,
          validity_unit: 'day',
          rate_multiplier: 1,
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null,
          features: [],
          group_platform: 'openai',
          sort_order: 1,
          for_sale: true,
          group_name: 'OpenAI',
          daily_purchase_limit: 5,
          daily_purchase_remaining: 3,
          daily_sale_starts_at: '09:00',
          daily_sale_ends_at: '18:00',
          daily_sale_status: 'available',
          daily_sale_countdown_seconds: 3600,
          daily_sale_available_for_payment: true,
        },
      ],
    },
  }
}

function jsapiOrderFixture(resumeToken: string) {
  return {
    order_id: 123,
    amount: 88,
    pay_amount: 88,
    fee_rate: 0,
    expires_at: '2099-01-01T00:10:00.000Z',
    payment_type: 'wxpay',
    out_trade_no: 'sub2_jsapi_123',
    result_type: 'jsapi_ready' as const,
    resume_token: resumeToken,
    jsapi: {
      appId: 'wx123',
      timeStamp: '1712345678',
      nonceStr: 'nonce',
      package: 'prepay_id=wx123',
      signType: 'RSA',
      paySign: 'signed',
    },
  }
}

function oauthOrderFixture() {
  return {
    order_id: 456,
    amount: 128,
    pay_amount: 128,
    fee_rate: 0,
    expires_at: '2099-01-01T00:10:00.000Z',
    payment_type: 'wxpay',
    result_type: 'oauth_required' as const,
    oauth: {
      authorize_url: '/api/v1/auth/oauth/wechat/payment/start?payment_type=wxpay&redirect=%2Fpurchase%3Ffrom%3Dwechat',
      appid: 'wx123',
      scope: 'snsapi_base',
      redirect_url: '/auth/wechat/payment/callback',
    },
  }
}

describe('PaymentView WeChat JSAPI flow', () => {
  beforeEach(() => {
    routeState.path = '/purchase'
    routeState.query = {
      wechat_resume: '1',
      wechat_resume_token: 'resume-token-123',
    }
    routerReplace.mockReset().mockResolvedValue(undefined)
    routerPush.mockReset().mockResolvedValue(undefined)
    routerResolve.mockClear()
    createOrder.mockReset()
    refreshUser.mockReset()
    fetchActiveSubscriptions.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showInfo.mockReset()
    showWarning.mockReset()
    getCheckoutInfo.mockReset().mockResolvedValue(checkoutInfoFixture())
    bridgeInvoke.mockReset()
    activeSubscriptionsState.value = []
    window.localStorage.clear()
    ;(window as Window & { WeixinJSBridge?: { invoke: typeof bridgeInvoke } }).WeixinJSBridge = {
      invoke: bridgeInvoke,
    }
  })

  it('resets payment state and redirects to /payment/result after JSAPI reports success', async () => {
    createOrder.mockResolvedValue(jsapiOrderFixture('resume-token-123'))
    bridgeInvoke.mockImplementation((_action, _payload, callback) => {
      callback({ err_msg: 'get_brand_wcpay_request:ok' })
    })

    shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(routerReplace).toHaveBeenCalledWith({ path: '/purchase', query: {} })
    expect(routerPush).toHaveBeenCalledWith({
      path: '/payment/result',
      query: {
        order_id: '123',
        out_trade_no: 'sub2_jsapi_123',
        resume_token: 'resume-token-123',
      },
    })
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('resets payment state when JSAPI reports cancellation', async () => {
    createOrder.mockResolvedValue(jsapiOrderFixture('resume-token-cancel'))
    bridgeInvoke.mockImplementation((_action, _payload, callback) => {
      callback({ err_msg: 'get_brand_wcpay_request:cancel' })
    })

    shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(showInfo).toHaveBeenCalledWith('payment.qr.cancelled')
    expect(routerPush).not.toHaveBeenCalled()
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('clears stale recovery state when JSAPI never becomes available', async () => {
    vi.useFakeTimers()
    createOrder.mockResolvedValue(jsapiOrderFixture('resume-token-missing-bridge'))
    ;(window as Window & { WeixinJSBridge?: { invoke: typeof bridgeInvoke } }).WeixinJSBridge = undefined

    const wrapper = shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(4000)
    await flushPromises()
    await flushPromises()

    expect(showError).toHaveBeenCalledWith(
      'payment.errors.wechatJsapiUnavailable payment.errors.wechatOpenInWeChatHint',
    )
    expect(routerPush).not.toHaveBeenCalled()
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
    expect(wrapper.html()).not.toContain('payment-status-panel-stub')
  })

  it('clears a stale recovery snapshot before handling wechat resume callback params', async () => {
    createOrder.mockRejectedValueOnce(new Error('resume failed'))
    window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 999,
      amount: 66,
      qrCode: 'stale-qr',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'alipay',
      payUrl: 'https://pay.example.com/stale',
      outTradeNo: 'stale-out-trade-no',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 66,
      orderType: 'balance',
      paymentMode: 'popup',
      resumeToken: '',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))

    shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      wechat_resume_token: 'resume-token-123',
    }))
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('keeps subscription resume context for token-only WeChat callbacks', async () => {
    routeState.query = {
      wechat_resume: '1',
      wechat_resume_token: 'resume-subscription-7',
      payment_type: 'wxpay_direct',
      order_type: 'subscription',
      plan_id: '7',
    }
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithPlansFixture())
    createOrder.mockResolvedValue(oauthOrderFixture())

    const originalLocation = window.location
    const locationState = {
      href: 'http://localhost/purchase',
      origin: 'http://localhost',
    }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState,
    })

    shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(routerReplace).toHaveBeenCalledWith({ path: '/purchase', query: {} })
    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      payment_type: 'wxpay',
      order_type: 'subscription',
      plan_id: 7,
      wechat_resume_token: 'resume-subscription-7',
    }))
    expect(locationState.href).toContain('/api/v1/auth/oauth/wechat/payment/start?')
    expect(new URL(locationState.href, 'http://localhost').searchParams.get('redirect')).toBe(
      '/purchase?from=wechat&payment_type=wxpay&order_type=subscription&plan_id=7',
    )

    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    })
  })

  it('falls back to QR flow when mobile WeChat payment is unavailable', async () => {
    routeState.query = {
      wechat_resume: '1',
      wechat_resume_token: 'resume-token-h5',
      payment_type: 'wxpay_direct',
    }
    createOrder
      .mockRejectedValueOnce({ reason: 'WECHAT_H5_NOT_AUTHORIZED' })
      .mockResolvedValueOnce({
        order_id: 778,
        amount: 88,
        pay_amount: 88,
        fee_rate: 0,
        expires_at: '2099-01-01T00:10:00.000Z',
        payment_type: 'wxpay',
        qr_code: 'weixin://wxpay/bizpayurl?pr=fallback-native',
        out_trade_no: 'sub2_qr_778',
      })

    shallowMount(PaymentView, {
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(createOrder).toHaveBeenNthCalledWith(1, expect.objectContaining({
      payment_type: 'wxpay',
      is_mobile: true,
      wechat_resume_token: 'resume-token-h5',
    }))
    expect(createOrder).toHaveBeenNthCalledWith(2, expect.objectContaining({
      payment_type: 'wxpay',
      is_mobile: false,
      payment_source: 'hosted_redirect',
    }))
    expect(showWarning).toHaveBeenCalledWith('payment.errors.mobilePaymentFallbackToQr')
    expect(showError).not.toHaveBeenCalled()
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toContain('weixin://wxpay/bizpayurl?pr=fallback-native')
  })

  it('shows credited balance and first recharge bonus as platform quota instead of RMB', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithFirstRechargeFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.find('[data-test="first-recharge-tier-tier-30"]').trigger('click')

    expect(wrapper.text()).toContain('390.00 平台额度')
    expect(wrapper.text()).toContain('+15.00 平台额度')
    expect(wrapper.text()).toContain('送 15.00 平台额度')
    expect(wrapper.text()).not.toContain('¥390.00')
    expect(wrapper.text()).not.toContain('+¥15.00')
    expect(wrapper.text()).not.toContain('返 ¥15.00')
  })

  it('hides 10 from recharge quick amount buttons', async () => {
    routeState.path = '/purchase'
    routeState.query = {}

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.find('[data-test="tab-recharge"]').trigger('click')
    await flushPromises()

    const amountInput = wrapper.findComponent({ name: 'AmountInput' })

    expect(amountInput.props('amounts')).toEqual([20, 50, 100, 200, 500, 1000, 2000, 5000])
  })

  it('shows subscription tab by default on purchase page', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithPlansFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    const tabButtons = wrapper.findAll('[data-test^="tab-"]')
    expect(tabButtons[0].attributes('data-test')).toBe('tab-subscription')
    expect(wrapper.findComponent({ name: 'SubscriptionPlanCard' }).exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'AmountInput' }).exists()).toBe(false)
  })

  it('shows first recharge tiers inside subscription tab instead of recharge tab', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithFirstRechargeFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.find('[data-test="subscription-first-recharge"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="first-recharge-tier-tier-30"]').exists()).toBe(true)

    await wrapper.find('[data-test="tab-recharge"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="subscription-first-recharge"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="first-recharge-tier-tier-30"]').exists()).toBe(false)
  })

  it('creates a balance order when paying from subscription first recharge entry', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithFirstRechargeFixture())
    createOrder.mockResolvedValue({
      order_id: 901,
      amount: 30,
      pay_amount: 30,
      fee_rate: 0,
      expires_at: '2099-01-01T00:10:00.000Z',
      payment_type: 'wxpay',
      qr_code: 'weixin://wxpay/bizpayurl?pr=first-recharge',
      out_trade_no: 'sub2_first_recharge_901',
    })

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.find('[data-test="first-recharge-tier-tier-30"]').trigger('click')
    await wrapper.find('[data-test="subscription-first-recharge-submit"]').trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      amount: 30,
      order_type: 'balance',
    }))
  })

  it('shows daily sale details and disables subscription submit for unavailable selected plan', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithUnavailableDailyPlanFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkoutInfoWithUnavailableDailyPlanFixture().data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.dailySaleTime')
    expect(wrapper.text()).toContain('09:00 - 18:00')
    expect(wrapper.text()).toContain('payment.planCard.soldOutToday')
    expect(wrapper.text()).not.toContain('payment.planCard.availableTodayLabel')
    const buttons = wrapper.findAll('button')
    const submit = buttons.find(button => button.text().includes('payment.createOrder'))
    expect(submit?.attributes('disabled')).toBeDefined()
  })

  it('hides daily remaining purchase count for available selected plan', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithAvailableDailyPlanFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).not.toContain('payment.planCard.availableTodayLabel')
    expect(wrapper.text()).not.toContain('Remaining today')

    await wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkoutInfoWithAvailableDailyPlanFixture().data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.dailySaleTime')
    expect(wrapper.text()).toContain('09:00 - 18:00')
    expect(wrapper.text()).not.toContain('payment.planCard.availableTodayLabel')
    expect(wrapper.text()).not.toContain('Remaining today')
  })

  it('keeps hot supported model summary visible and reveals all models after selecting a subscription plan', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithPlansFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkoutInfoWithPlansFixture().data.plans[0])
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('payment.planCard.models')
    expect(text).toContain('gpt-5.5')
    expect(text).toContain('gpt-5.4')
    expect(text).toContain('gpt-5.4-mini')
    expect(text).toContain('gpt-image-2')
    expect(text).not.toContain('gpt-5.3-codex')
    expect(text).toContain('+1')

    await wrapper.get('[data-test="supported-models-summary"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.supportedModelsTitle')
    expect(wrapper.text()).toContain('gpt-5.3-codex')
  })

  it('shows subscription cycle total quota in selected plan confirmation', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithPlansFixture())

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    await wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkoutInfoWithPlansFixture().data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.purchaseAddsCycleQuota')
    expect(wrapper.text()).toContain('¥1560')
  })

  it('uses subscription instance total limit when deciding whether active subscription is unlimited', async () => {
    routeState.path = '/purchase'
    routeState.query = {}
    getCheckoutInfo.mockResolvedValue(checkoutInfoWithPlansFixture())
    activeSubscriptionsState.value = [
      {
        id: 47,
        group_id: 34,
        user_id: 2,
        status: 'active',
        starts_at: '2026-06-26T15:57:02Z',
        expires_at: '2028-06-26T15:57:02Z',
        daily_usage_usd: 0,
        weekly_usage_usd: 0,
        monthly_usage_usd: 0,
        total_limit_usd: 3120,
        total_usage_usd: 0,
        daily_window_start: null,
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: '2026-06-26T15:57:02Z',
        updated_at: '2026-06-26T15:57:02Z',
        group: {
          id: 34,
          name: 'Early Bird',
          platform: 'openai',
          rate_multiplier: 1,
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null,
          subscription_total_limit_usd: 1560,
        },
      },
    ]

    const wrapper = mount(PaymentView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          AmountInput: true,
          PaymentMethodSelector: true,
          PaymentStatusPanel: true,
          SubscriptionPlanCard: true,
          Icon: true,
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('Early Bird')
    expect(wrapper.text()).not.toContain('payment.planCard.quota: payment.planCard.unlimited')
  })
})
