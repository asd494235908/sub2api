import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import SubscriptionProgressMini from '../SubscriptionProgressMini.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/stores', () => ({
  useSubscriptionStore: () => ({
    hasActiveSubscriptions: true,
    fetchActiveSubscriptions: vi.fn().mockResolvedValue(undefined),
    activeSubscriptions: [
      {
        id: 1,
        group_id: 10,
        status: 'active',
        starts_at: '2026-01-01T00:00:00Z',
        expires_at: '2027-01-01T00:00:00Z',
        daily_usage_usd: 0,
        weekly_usage_usd: 0,
        monthly_usage_usd: 0,
        total_limit_usd: 3120,
        total_usage_usd: 260,
        daily_window_start: null,
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        user_id: 1,
        group: {
          id: 10,
          name: 'Early Bird',
          description: null,
          platform: 'openai',
          rate_multiplier: 1,
          is_exclusive: true,
          status: 'active',
          subscription_type: 'subscription',
          daily_limit_usd: null,
          weekly_limit_usd: null,
          monthly_limit_usd: null,
          subscription_total_limit_usd: 1560,
          allow_image_generation: false,
          image_rate_independent: false,
          image_rate_multiplier: 1,
          image_price_1k: null,
          image_price_2k: null,
          image_price_4k: null,
          claude_code_only: false,
          fallback_group_id: null,
          fallback_group_id_on_invalid_request: null,
          require_oauth_only: false,
          require_privacy_set: false,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
      },
    ],
  }),
}))

describe('SubscriptionProgressMini', () => {
  it('shows subscription cycle total usage', async () => {
    const wrapper = mount(SubscriptionProgressMini, {
      global: {
        stubs: {
          Icon: defineComponent({ template: '<span />' }),
          RouterLink: defineComponent({ template: '<a><slot /></a>' }),
          transition: false,
        },
      },
    })

    await wrapper.get('button').trigger('click')

    expect(wrapper.text()).toContain('subscriptionProgress.total')
    expect(wrapper.text()).toContain('$260.00/$3120.00')
  })
})
