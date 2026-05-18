import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RedeemView from '../RedeemView.vue'

const { getHistory, getWeeklyQuota, getPublicSettings, claimWeeklyQuota, authUser, showError } = vi.hoisted(() => ({
  getHistory: vi.fn(),
  getWeeklyQuota: vi.fn(),
  getPublicSettings: vi.fn(),
  claimWeeklyQuota: vi.fn(),
  authUser: {
    balance: 12.34,
    concurrency: 5,
    phone_number: undefined as string | undefined,
  },
  showError: vi.fn(),
}))

vi.mock('@/api', () => ({
  redeemAPI: {
    getHistory,
    getWeeklyQuota,
    redeem: vi.fn(),
    claimWeeklyQuota,
  },
  authAPI: {
    getPublicSettings,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: authUser,
    refreshUser: vi.fn(),
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
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
    claimWeeklyQuota.mockReset()
    showError.mockReset()
    authUser.balance = 12.34
    authUser.concurrency = 5
    authUser.phone_number = undefined

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

  function mountRedeemView() {
    return mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<span />' },
        },
      },
    })
  }

  it('shows lucky wheel bonus entries inside recent activity', async () => {
    const wrapper = mountRedeemView()

    await flushPromises()

    expect(wrapper.text()).toContain('Lucky Wheel Bonus')
    expect(wrapper.text()).toContain('+¥12.34')
  })

  it('disables weekly quota claim and shows phone binding prompt when phone verification is enabled', async () => {
    getWeeklyQuota.mockResolvedValue({
      enabled: true,
      amount: 12.5,
      status: 'claimable',
      window_started_at: '2026-05-13T00:00:00Z',
      window_ends_at: '2026-05-20T00:00:00Z',
      total_claim_count: 0,
      total_claim_amount: 0,
    })
    getPublicSettings.mockResolvedValue({ contact_info: '', phone_verify_enabled: true })

    const wrapper = mountRedeemView()
    await flushPromises()

    expect(wrapper.text()).toContain('redeem.weeklyQuotaPhoneRequired')
    const buttons = wrapper.findAll('button')
    const claimButton = buttons.find((button) => button.text() === 'redeem.weeklyQuotaClaimButton')
    expect(claimButton?.attributes('disabled')).toBeDefined()
  })

  it('does not call weekly quota claim API when phone binding is required', async () => {
    getWeeklyQuota.mockResolvedValue({
      enabled: true,
      amount: 12.5,
      status: 'claimable',
      window_started_at: '2026-05-13T00:00:00Z',
      window_ends_at: '2026-05-20T00:00:00Z',
      total_claim_count: 0,
      total_claim_amount: 0,
    })
    getPublicSettings.mockResolvedValue({ contact_info: '', phone_verify_enabled: true })

    const wrapper = mountRedeemView()
    await flushPromises()

    ;(wrapper.vm as any).handleClaimWeeklyQuota()
    await flushPromises()

    expect(claimWeeklyQuota).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('redeem.weeklyQuotaBindPhoneFirst')
  })

  it('keeps weekly quota claim available when phone verification is disabled', async () => {
    getWeeklyQuota.mockResolvedValue({
      enabled: true,
      amount: 12.5,
      status: 'claimable',
      window_started_at: '2026-05-13T00:00:00Z',
      window_ends_at: '2026-05-20T00:00:00Z',
      total_claim_count: 0,
      total_claim_amount: 0,
    })
    getPublicSettings.mockResolvedValue({ contact_info: '', phone_verify_enabled: false })

    const wrapper = mountRedeemView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('redeem.weeklyQuotaPhoneRequired')
    const buttons = wrapper.findAll('button')
    const claimButton = buttons.find((button) => button.text() === 'redeem.weeklyQuotaClaimButton')
    expect(claimButton?.attributes('disabled')).toBeUndefined()
  })

  it('keeps weekly quota claim available when phone is bound and phone verification is enabled', async () => {
    authUser.phone_number = '+8613800138000'
    getWeeklyQuota.mockResolvedValue({
      enabled: true,
      amount: 12.5,
      status: 'claimable',
      window_started_at: '2026-05-13T00:00:00Z',
      window_ends_at: '2026-05-20T00:00:00Z',
      total_claim_count: 0,
      total_claim_amount: 0,
    })
    getPublicSettings.mockResolvedValue({ contact_info: '', phone_verify_enabled: true })

    const wrapper = mountRedeemView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('redeem.weeklyQuotaPhoneRequired')
    const buttons = wrapper.findAll('button')
    const claimButton = buttons.find((button) => button.text() === 'redeem.weeklyQuotaClaimButton')
    expect(claimButton?.attributes('disabled')).toBeUndefined()
  })
})
