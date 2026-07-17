import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, shallowMount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import PaymentView from '../PaymentView.vue'
import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { CheckoutInfoResponse, MethodLimit, SubscriptionPlan } from '@/types/payment'

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
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'payment.platformQuotaAmount') return `${params?.amount} 平台额度`
        if (key === 'payment.firstRecharge.tierBonus') return `送 ${params?.amount} 平台额度`
        return key
      },
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

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: null }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getCheckoutInfo,
  },
}))

vi.mock('@/utils/device', () => ({
  isMobileDevice: () => true,
}))

function checkoutInfoFixture(overrides: Partial<CheckoutInfoResponse> = {}) {
  const wxpayMethod: MethodLimit = {
    daily_limit: 0,
    daily_used: 0,
    daily_remaining: 0,
    single_min: 0,
    single_max: 0,
    fee_rate: 0,
    available: true,
  }
  const data: CheckoutInfoResponse = {
    methods: {
      wxpay: wxpayMethod,
    },
    global_min: 0,
    global_max: 0,
    plans: [],
    balance_disabled: false,
    balance_recharge_multiplier: 1,
    subscription_usd_to_cny_rate: 0,
    recharge_fee_rate: 0,
    help_text: '',
    help_image_url: '',
    stripe_publishable_key: '',
    test_recharge_enabled: false,
  }

  return {
    data: { ...data, ...overrides },
  }
}

function checkoutInfoWithPlansFixture(options: {
  checkout?: Partial<CheckoutInfoResponse>
  method?: Partial<MethodLimit>
  plan?: Partial<SubscriptionPlan>
} = {}) {
  const base = checkoutInfoFixture(options.checkout).data
  const plan: SubscriptionPlan = {
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
    ...options.plan,
  }

  return {
    data: {
      ...base,
      methods: {
        ...base.methods,
        wxpay: {
          ...base.methods.wxpay,
          ...options.method,
        },
      },
      plans: [plan],
    },
  }
}

function checkoutInfoWithFirstRechargeFixture() {
  return checkoutInfoWithPlansFixture({
    checkout: {
      balance_recharge_multiplier: 13,
      first_recharge: {
        enabled: true,
        tiers: [{ id: 'tier-30', pay_amount: 30, bonus_amount: 15, enabled: true, sort_order: 1 }],
      },
    },
  })
}

function checkoutInfoWithSalePlanFixture(plan: Partial<SubscriptionPlan>) {
  return checkoutInfoWithPlansFixture({
    plan: {
      subscription_total_limit_usd: null,
      supported_models: [],
      ...plan,
    },
  })
}

async function mountLocalPayment(options: {
  checkout?: ReturnType<typeof checkoutInfoFixture>
  amountInput?: boolean
  subscriptionPlanCard?: boolean
} = {}) {
  vi.useRealTimers()
  routeState.path = '/purchase'
  routeState.query = {}
  routerReplace.mockReset().mockResolvedValue(undefined)
  routerPush.mockReset().mockResolvedValue(undefined)
  routerResolve.mockClear()
  createOrder.mockReset()
  refreshUser.mockReset()
  fetchActiveSubscriptions.mockReset().mockResolvedValue(undefined)
  showError.mockReset()
  showInfo.mockReset()
  showWarning.mockReset()
  getCheckoutInfo.mockReset().mockResolvedValue(options.checkout ?? checkoutInfoFixture())
  bridgeInvoke.mockReset()
  window.localStorage.clear()
  ;(window as Window & { WeixinJSBridge?: { invoke: typeof bridgeInvoke } }).WeixinJSBridge = undefined

  const wrapper = mount(PaymentView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        AmountInput: options.amountInput ?? true,
        PaymentMethodSelector: true,
        PaymentStatusPanel: true,
        SubscriptionPlanCard: options.subscriptionPlanCard ?? true,
        Icon: true,
        Teleport: true,
        Transition: false,
      },
    },
  })
  await flushPromises()
  await flushPromises()
  return wrapper
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

async function mountSubscriptionConfirm(options: Parameters<typeof checkoutInfoWithPlansFixture>[0] = {}) {
  vi.useRealTimers()
  routeState.path = '/purchase'
  routeState.query = {
    tab: 'subscription',
    group: '3',
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
  getCheckoutInfo.mockReset().mockResolvedValue(checkoutInfoWithPlansFixture(options))
  bridgeInvoke.mockReset()
  window.localStorage.clear()
  ;(window as Window & { WeixinJSBridge?: { invoke: typeof bridgeInvoke } }).WeixinJSBridge = undefined

  const wrapper = shallowMount(PaymentView, {
    global: {
      stubs: {
        AppLayout: {
          template: '<div><slot /></div>',
        },
        Teleport: true,
        Transition: false,
      },
    },
  })
  await flushPromises()
  await flushPromises()
  return wrapper
}

describe('PaymentView subscription confirmation amounts', () => {
  it('shows converted CNY pay amount using the subscription rate, not the balance multiplier', async () => {
    const wrapper = await mountSubscriptionConfirm({
      checkout: {
        balance_recharge_multiplier: 0.14,
        subscription_usd_to_cny_rate: 7.15,
      },
      method: {
        currency: 'CNY',
      },
      plan: {
        price: 9.99,
        original_price: 12.99,
      },
    })

    const text = wrapper.text()
    const convertedPrice = formatPaymentAmount(71.43, 'CNY')
    const convertedOriginalPrice = formatPaymentAmount(92.88, 'CNY')

    expect(text).toContain(convertedPrice)
    expect(text).toContain(convertedOriginalPrice)
    expect(text).not.toContain(formatPaymentAmount(9.99, 'CNY'))
    // 换算必须使用订阅汇率（×7.15），而不是余额倍率（÷0.14 = 71.36）
    expect(text).not.toContain(formatPaymentAmount(71.36, 'CNY'))
    expect(wrapper.findAll('button').some(button => button.text().includes(convertedPrice))).toBe(true)
  })

  it('checks CNY gateway limits against the converted amount only', async () => {
    const wrapper = await mountSubscriptionConfirm({
      checkout: {
        subscription_usd_to_cny_rate: 7.15,
      },
      method: {
        currency: 'CNY',
        single_min: 10,
      },
      plan: {
        price: 7.99,
      },
    })

    const convertedPrice = formatPaymentAmount(57.13, 'CNY')
    const submitButton = wrapper.findAll('button').find(button => button.text().includes(convertedPrice))

    expect(submitButton).toBeDefined()
    expect(submitButton?.attributes('disabled')).toBeUndefined()
  })

  it('keeps plan price when the subscription rate is not configured or payment currency is not CNY', async () => {
    // opt-in 回归锁：即使余额倍率已配置，未配置订阅汇率时 CNY 订阅仍按 price 直付
    const cnyWrapper = await mountSubscriptionConfirm({
      checkout: {
        balance_recharge_multiplier: 0.14,
        subscription_usd_to_cny_rate: 0,
      },
      method: {
        currency: 'CNY',
      },
      plan: {
        price: 7.99,
      },
    })

    expect(cnyWrapper.text()).toContain(formatPaymentAmount(7.99, 'CNY'))
    expect(cnyWrapper.text()).not.toContain(formatPaymentAmount(57.07, 'CNY'))
    expect(cnyWrapper.text()).not.toContain(formatPaymentAmount(57.13, 'CNY'))

    const usdWrapper = await mountSubscriptionConfirm({
      checkout: {
        subscription_usd_to_cny_rate: 7.15,
      },
      method: {
        currency: 'USD',
      },
      plan: {
        price: 7.99,
        original_price: 9.99,
      },
    })

    expect(usdWrapper.text()).toContain(formatPaymentAmount(7.99, 'USD'))
    expect(usdWrapper.text()).toContain(formatPaymentAmount(9.99, 'USD'))
  })

  it('adds fee rate after CNY rate conversion to match backend pay_amount', async () => {
    const wrapper = await mountSubscriptionConfirm({
      checkout: {
        subscription_usd_to_cny_rate: 7.15,
        recharge_fee_rate: 2.5,
      },
      method: {
        currency: 'CNY',
      },
      plan: {
        price: 9.99,
      },
    })

    const text = wrapper.text()
    const convertedPrice = formatPaymentAmount(71.43, 'CNY')
    const fee = formatPaymentAmount(1.79, 'CNY')
    const total = formatPaymentAmount(73.22, 'CNY')

    expect(text).toContain(convertedPrice)
    expect(text).toContain(fee)
    expect(text).toContain(total)
    expect(wrapper.findAll('button').some(button => button.text().includes(total))).toBe(true)
  })
})

describe('PaymentView payment recovery', () => {
  beforeEach(() => {
    vi.useRealTimers()
    routeState.path = '/purchase'
    routeState.query = {}
    routerReplace.mockReset().mockResolvedValue(undefined)
    routerPush.mockReset().mockResolvedValue(undefined)
    routerResolve.mockClear()
    createOrder.mockReset()
    refreshUser.mockReset()
    fetchActiveSubscriptions.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showInfo.mockReset()
    showWarning.mockReset()
    bridgeInvoke.mockReset()
    window.localStorage.clear()
    ;(window as Window & { WeixinJSBridge?: { invoke: typeof bridgeInvoke } }).WeixinJSBridge = undefined
  })

  it('restores a custom EasyPay method as the selected payment method', async () => {
    getCheckoutInfo.mockResolvedValue(checkoutInfoFixture({
      methods: {
        wxpay: checkoutInfoFixture().data.methods.wxpay,
        ldc: {
          daily_limit: 0,
          daily_used: 0,
          daily_remaining: 0,
          single_min: 0,
          single_max: 0,
          fee_rate: 0,
          available: true,
          display_name: 'LDC Pay',
        },
      },
    }))
    window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 888,
      amount: 66,
      qrCode: 'ldc-qr',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'ldc',
      payUrl: 'https://pay.example.com/ldc',
      outTradeNo: 'sub2_ldc_888',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 66,
      orderType: 'balance',
      paymentMode: 'popup',
      resumeToken: '',
      createdAt: Date.now(),
    }))

    const wrapper = shallowMount(PaymentView, {
      global: {
        stubs: {
          AppLayout: {
            template: '<div><slot /></div>',
          },
          PaymentStatusPanel: {
            template: '<button data-test="payment-done" @click="$emit(\'done\')" />',
          },
          PaymentMethodSelector: {
            props: ['selected'],
            template: '<div data-test="method-selector">{{ selected }}</div>',
          },
          Teleport: true,
          Transition: false,
        },
      },
    })
    await flushPromises()
    await flushPromises()
    await wrapper.find('[data-test="payment-done"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="method-selector"]').text()).toBe('ldc')
  })
})

describe('PaymentView local purchase protections', () => {
  beforeEach(() => {
    activeSubscriptionsState.value = []
  })

  it('shows credited balance and first recharge bonus as platform quota instead of RMB', async () => {
    const wrapper = await mountLocalPayment({ checkout: checkoutInfoWithFirstRechargeFixture() })

    await wrapper.get('[data-test="first-recharge-tier-tier-30"]').trigger('click')

    expect(wrapper.text()).toContain('390.00 平台额度')
    expect(wrapper.text()).toContain('+15.00 平台额度')
    expect(wrapper.text()).toContain('送 15.00 平台额度')
    expect(wrapper.text()).not.toContain('¥390.00')
    expect(wrapper.text()).not.toContain('+¥15.00')
  })

  it('hides 10 from recharge quick amount buttons', async () => {
    const wrapper = await mountLocalPayment()

    await wrapper.get('[data-test="tab-recharge"]').trigger('click')

    expect(wrapper.findComponent({ name: 'AmountInput' }).props('amounts')).toEqual([
      20, 50, 100, 200, 500, 1000, 2000, 5000,
    ])
  })

  it('hides wechat test recharge button when checkout switch is disabled', async () => {
    const wrapper = await mountLocalPayment()

    await wrapper.get('[data-test="tab-recharge"]').trigger('click')

    expect(wrapper.find('[data-test="wechat-test-recharge-button"]').exists()).toBe(false)
  })

  it('creates a 0.01 wxpay balance order from wechat test recharge button', async () => {
    const wrapper = await mountLocalPayment({
      checkout: checkoutInfoFixture({ test_recharge_enabled: true }),
    })
    createOrder.mockResolvedValue({
      order_id: 902,
      amount: 0.01,
      pay_amount: 0.01,
      fee_rate: 0,
      expires_at: '2099-01-01T00:10:00.000Z',
      payment_type: 'wxpay',
      qr_code: 'weixin://wxpay/bizpayurl?pr=test-recharge',
      out_trade_no: 'sub2_test_recharge_902',
    })

    await wrapper.get('[data-test="tab-recharge"]').trigger('click')
    await wrapper.get('[data-test="wechat-test-recharge-button"]').trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      amount: 0.01,
      payment_type: 'wxpay',
      order_type: 'balance',
      test_recharge: true,
    }))
  })

  it('shows subscription tab by default on purchase page', async () => {
    const wrapper = await mountLocalPayment({ checkout: checkoutInfoWithPlansFixture() })

    expect(wrapper.findAll('[data-test^="tab-"]')[0].attributes('data-test')).toBe('tab-subscription')
    expect(wrapper.findComponent({ name: 'SubscriptionPlanCard' }).exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'AmountInput' }).exists()).toBe(false)
  })

  it('shows first recharge tiers only inside the subscription tab', async () => {
    const wrapper = await mountLocalPayment({ checkout: checkoutInfoWithFirstRechargeFixture() })

    expect(wrapper.find('[data-test="subscription-first-recharge"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="first-recharge-tier-tier-30"]').exists()).toBe(true)

    await wrapper.get('[data-test="tab-recharge"]').trigger('click')

    expect(wrapper.find('[data-test="subscription-first-recharge"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="first-recharge-tier-tier-30"]').exists()).toBe(false)
  })

  it('shows a custom first recharge amount as platform quota', async () => {
    const wrapper = await mountLocalPayment({
      checkout: checkoutInfoWithFirstRechargeFixture(),
      amountInput: false,
    })
    const firstRecharge = wrapper.get('[data-test="subscription-first-recharge"]')

    await firstRecharge.get('input[type="text"]').setValue('88.88')

    expect(firstRecharge.text()).toContain('1155.44 平台额度')
    expect(firstRecharge.text()).not.toContain('¥1155.44')
  })

  it('creates a balance order from the subscription first recharge entry', async () => {
    const wrapper = await mountLocalPayment({ checkout: checkoutInfoWithFirstRechargeFixture() })
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

    await wrapper.get('[data-test="first-recharge-tier-tier-30"]').trigger('click')
    await wrapper.get('[data-test="subscription-first-recharge-submit"]').trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      amount: 30,
      order_type: 'balance',
    }))
  })

  it('shows daily sale details and disables purchase for an unavailable plan', async () => {
    const checkout = checkoutInfoWithSalePlanFixture({
      id: 8,
      name: 'Flash Sale',
      daily_purchase_limit: 2,
      daily_purchase_remaining: 0,
      daily_sale_starts_at: '09:00',
      daily_sale_ends_at: '18:00',
      daily_sale_status: 'sold_out',
      daily_sale_countdown_seconds: 3600,
      daily_sale_available_for_payment: false,
    })
    const wrapper = await mountLocalPayment({ checkout, subscriptionPlanCard: false })

    wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkout.data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.dailySaleTime')
    expect(wrapper.text()).toContain('09:00 - 18:00')
    expect(wrapper.text()).toContain('payment.planCard.soldOutToday')
    expect(wrapper.text()).not.toContain('payment.planCard.availableTodayLabel')
    const submit = wrapper.findAll('button').find(button => button.text().includes('payment.createOrder'))
    expect(submit?.attributes('disabled')).toBeDefined()
  })

  it('hides remaining purchase count while an available daily plan is selected', async () => {
    const checkout = checkoutInfoWithSalePlanFixture({
      id: 9,
      name: 'Available Flash Sale',
      daily_purchase_limit: 5,
      daily_purchase_remaining: 3,
      daily_sale_starts_at: '09:00',
      daily_sale_ends_at: '18:00',
      daily_sale_status: 'available',
      daily_sale_countdown_seconds: 3600,
      daily_sale_available_for_payment: true,
    })
    const wrapper = await mountLocalPayment({ checkout, subscriptionPlanCard: false })

    wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkout.data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('09:00 - 18:00')
    expect(wrapper.text()).not.toContain('payment.planCard.availableTodayLabel')
    expect(wrapper.text()).not.toContain('Remaining today')
  })

  it('shows weekly off-day details and disables purchase outside sale days', async () => {
    const checkout = checkoutInfoWithSalePlanFixture({
      id: 10,
      name: 'Weekly Flash Sale',
      daily_purchase_limit: 0,
      daily_sale_starts_at: '09:00',
      daily_sale_ends_at: '10:00',
      daily_sale_status: 'available',
      daily_sale_countdown_seconds: 3600,
      daily_sale_available_for_payment: false,
      weekly_sale_days: [1, 3, 5],
      weekly_sale_status: 'off_day',
      weekly_sale_available_for_payment: false,
    })
    const wrapper = await mountLocalPayment({ checkout, subscriptionPlanCard: false })

    wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkout.data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('09:00 - 10:00')
    expect(wrapper.text()).toContain('payment.planCard.weeklySaleOffDay')
    expect(wrapper.text()).not.toContain('payment.planCard.dailySaleCountdownToEnd')
    const submit = wrapper.findAll('button').find(button => button.text().includes('payment.createOrder'))
    expect(submit?.attributes('disabled')).toBeDefined()
  })

  it('keeps hot model summary visible and expands to all supported models', async () => {
    const checkout = checkoutInfoWithPlansFixture()
    const wrapper = await mountLocalPayment({ checkout, subscriptionPlanCard: false })

    wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkout.data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('gpt-5.5')
    expect(wrapper.text()).not.toContain('gpt-5.3-codex')
    expect(wrapper.text()).toContain('+1')

    await wrapper.get('[data-test="supported-models-summary"]').trigger('click')

    expect(wrapper.text()).toContain('gpt-5.3-codex')
  })

  it('shows subscription cycle total quota in the selected plan confirmation', async () => {
    const checkout = checkoutInfoWithPlansFixture()
    const wrapper = await mountLocalPayment({ checkout, subscriptionPlanCard: false })

    wrapper.findComponent({ name: 'SubscriptionPlanCard' }).vm.$emit('select', checkout.data.plans[0])
    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.purchaseAddsCycleQuota')
    expect(wrapper.text()).toContain('¥1560')
  })

  it('uses the subscription instance total limit to determine unlimited status', async () => {
    activeSubscriptionsState.value = [{
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
    }]

    const wrapper = await mountLocalPayment({ checkout: checkoutInfoWithPlansFixture() })

    expect(wrapper.text()).toContain('Early Bird')
    expect(wrapper.text()).not.toContain('payment.planCard.quota: payment.planCard.unlimited')
  })
})

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
})
