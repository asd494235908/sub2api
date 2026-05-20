import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import type { SubscriptionPlan } from '@/types/payment'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const planFactory = (): SubscriptionPlan => ({
  id: 1,
  group_id: 7,
  group_platform: 'openai',
  group_name: 'OpenAI',
  rate_multiplier: 1.5,
  daily_limit_usd: 95,
  weekly_limit_usd: null,
  monthly_limit_usd: null,
  supported_model_scopes: ['claude', 'gemini_text', 'gemini_image'],
  name: '标准版',
  description: '套餐每天95刀，时长30天',
  price: 135,
  validity_days: 30,
  validity_unit: 'day',
  features: [],
  for_sale: true,
  sort_order: 1,
})

describe('SubscriptionPlanCard', () => {
  it('hides model scope labels while keeping core plan details visible', () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: planFactory(),
        activeSubscriptions: [],
      },
    })

    const text = wrapper.text()

    expect(text).toContain('标准版')
    expect(text).toContain('OpenAI')
    expect(text).toContain('135')
    expect(text).toContain('×1.5')
    expect(text).toContain('¥95')
    expect(text).toContain('payment.subscribeNow')

    expect(text).not.toContain('payment.planCard.models')
    expect(text).not.toContain('Claude')
    expect(text).not.toContain('Gemini')
    expect(text).not.toContain('Imagen')
  })
})
