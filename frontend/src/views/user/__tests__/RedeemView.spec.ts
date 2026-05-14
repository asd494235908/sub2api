import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RedeemView from '../RedeemView.vue'

const { getHistory, getWeeklyQuota, getPublicSettings } = vi.hoisted(() => ({
  getHistory: vi.fn(),
  getWeeklyQuota: vi.fn(),
  getPublicSettings: vi.fn(),
}))

vi.mock('@/api', () => ({
  redeemAPI: {
    getHistory,
    getWeeklyQuota,
    redeem: vi.fn(),
    claimWeeklyQuota: vi.fn(),
  },
  authAPI: {
    getPublicSettings,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      balance: 12.34,
      concurrency: 5,
    },
    refreshUser: vi.fn(),
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
  }),
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    fetchActiveSubscriptions: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('RedeemView', () => {
  beforeEach(() => {
    getHistory.mockReset()
    getWeeklyQuota.mockReset()
    getPublicSettings.mockReset()

    getHistory.mockResolvedValue([
      {
        id: 901,
        code: 'LUCKY-SESSION-1',
        type: 'lucky_wheel_bonus',
        value: 12.34,
        status: 'used',
        used_at: '2026-05-13T10:00:00Z',
        created_at: '2026-05-13T10:00:00Z',
        source: 'lucky_wheel',
        title: 'Lucky Wheel Bonus',
      },
      {
        id: 900,
        code: 'CODE-123',
        type: 'balance',
        value: 1.25,
        status: 'used',
        used_at: '2026-05-13T09:00:00Z',
        created_at: '2026-05-13T09:00:00Z',
        source: 'redeem_code',
        title: '',
      },
    ])
    getWeeklyQuota.mockResolvedValue({
      enabled: false,
      amount: 0,
      status: 'disabled',
      window_started_at: '',
      window_ends_at: '',
      total_claim_count: 0,
      total_claim_amount: 0,
    })
    getPublicSettings.mockResolvedValue({ contact_info: '' })
  })

  it('shows lucky wheel bonus entries inside recent activity', async () => {
    const wrapper = mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Lucky Wheel Bonus')
    expect(wrapper.text()).toContain('+¥12.34')
  })
})
