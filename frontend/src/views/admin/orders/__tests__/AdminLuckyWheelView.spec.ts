import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminLuckyWheelView from '../AdminLuckyWheelView.vue'

const { getLuckyWheelConfig, getLuckyWheelStats, updateLuckyWheelConfig } = vi.hoisted(() => ({
  getLuckyWheelConfig: vi.fn(),
  getLuckyWheelStats: vi.fn(),
  updateLuckyWheelConfig: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getLuckyWheelConfig,
    getLuckyWheelStats,
    updateLuckyWheelConfig,
  },
  default: {
    getLuckyWheelConfig,
    getLuckyWheelStats,
    updateLuckyWheelConfig,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
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

const configPayload = {
  enabled: true,
  config: {
    eligible_order_types: ['balance', 'subscription'],
    multiplier_step: 0.1,
    global_max_multiplier: 3,
    intro_text: '后台活动简介',
    rules_title: '后台活动规则',
    rules_items: ['后台规则 1', '后台规则 2'],
    amount_tiers: [
      { id: 'tier_20_50', name: '20-50', min_amount: 20, max_amount: 50, min_multiplier: 1.1, max_multiplier: 2, draw_count: 2 },
      { id: 'tier_51_plus', name: '51+', min_amount: 51, max_amount: null, min_multiplier: 1.2, max_multiplier: 3, draw_count: 3 },
    ],
    invite_bonus: {
      enabled: true,
      qualifying_amount: 20,
      bonus_per_invitee: 0.2,
      max_bonus: 1,
      consume_policy: 'next_session_once',
    },
    golden_window: {
      enabled: true,
      timezone: 'Asia/Shanghai',
      start_time: '20:00',
      end_time: '22:00',
      min_amount: 51,
      extra_draws: 1,
      daily_quota: 5,
    },
  },
}

describe('AdminLuckyWheelView', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    getLuckyWheelConfig.mockReset()
    getLuckyWheelStats.mockReset()
    updateLuckyWheelConfig.mockReset()

    getLuckyWheelConfig.mockResolvedValue({ data: configPayload })
    getLuckyWheelStats.mockResolvedValue({
      data: {
        enabled: true,
        total_sessions: 20,
        pending_sessions: 8,
        settled_sessions: 12,
        total_bonus_amount: 188.8,
        recent_sessions: [],
        multiplier_stats: [
          { multiplier: 1.2, draw_count: 6 },
          { multiplier: 2.4, draw_count: 3 },
        ],
        golden_window_used_today: 3,
        golden_window_daily_quota: 5,
      },
    })
    updateLuckyWheelConfig.mockResolvedValue({ data: configPayload })
  })

  function mountView() {
    return mount(AdminLuckyWheelView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Toggle: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<button data-test="toggle" @click="$emit(\'update:modelValue\', !modelValue)">toggle</button>',
          },
        },
      },
    })
  }

  it('renders amount tiers and multiplier stats', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.findAll('[data-test="tier-row"]')).toHaveLength(2)
    expect((wrapper.get('textarea[data-test="intro-text"]').element as HTMLTextAreaElement).value).toBe('后台活动简介')
    expect((wrapper.get('input[data-test="rules-title"]').element as HTMLInputElement).value).toBe('后台活动规则')
    expect(wrapper.text()).toContain('luckyWheel.adminTiersHint')
    expect(wrapper.text()).toContain('20')
    expect(wrapper.text()).toContain('188.80')
  })

  it('shows validation issues and blocks save when multiplier step is invalid', async () => {
    const wrapper = mountView()
    await flushPromises()

    const stepInput = wrapper.get('input[data-test="multiplier-step"]')
    await stepInput.setValue('0')
    await flushPromises()

    expect(wrapper.text()).toContain('luckyWheel.adminValidationStep')
    const saveButton = wrapper.get('button[data-test="save-config"]')
    expect(saveButton.attributes('disabled')).toBeDefined()
  })

  it('persists updated config payload', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input[data-test="invite-bonus-per-invitee"]').setValue('0.3')
    await wrapper.get('button[data-test="save-config"]').trigger('click')
    await flushPromises()

    expect(updateLuckyWheelConfig).toHaveBeenCalledTimes(1)
    expect(updateLuckyWheelConfig).toHaveBeenCalledWith(expect.objectContaining({
      enabled: true,
      config: expect.objectContaining({
        invite_bonus: expect.objectContaining({
          bonus_per_invitee: 0.3,
        }),
      }),
    }))
  })

  it('allows multipliers below one in amount tier config', async () => {
    const wrapper = mountView()
    await flushPromises()

    const firstTier = wrapper.findAll('[data-test="tier-row"]')[0]
    const minMultiplierInput = firstTier.get('input[data-test="tier-min-multiplier"]')
    const maxMultiplierInput = firstTier.get('input[data-test="tier-max-multiplier"]')

    expect(minMultiplierInput.attributes('min')).toBe('0.01')
    expect(maxMultiplierInput.attributes('min')).toBe('0.01')

    await minMultiplierInput.setValue('0.5')
    await maxMultiplierInput.setValue('0.8')
    await wrapper.get('button[data-test="save-config"]').trigger('click')
    await flushPromises()

    expect(updateLuckyWheelConfig).toHaveBeenCalledWith(expect.objectContaining({
      config: expect.objectContaining({
        amount_tiers: expect.arrayContaining([
          expect.objectContaining({
            min_multiplier: 0.5,
            max_multiplier: 0.8,
          }),
        ]),
      }),
    }))
  })

  it('adds a new amount tier row', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[data-test="add-tier"]').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-test="tier-row"]')).toHaveLength(3)
  })

  it('keeps settled history inside an internal scroll area', async () => {
    getLuckyWheelStats.mockResolvedValueOnce({
      data: {
        enabled: true,
        total_sessions: 20,
        pending_sessions: 8,
        settled_sessions: 12,
        total_bonus_amount: 188.8,
        recent_sessions: Array.from({ length: 12 }, (_, index) => ({
          id: index + 1,
          user_id: 1000 + index,
          source_order_id: 2000 + index,
          source_order_type: 'balance',
          source_pay_amount: 20 + index,
          matched_tier_id: 'tier_20_50',
          matched_tier_name: '20-50',
          min_multiplier: 1.1,
          max_multiplier: 2,
          total_draws: 2,
          completed_draws: 2,
          remaining_draws: 0,
          best_multiplier: 1.5,
          invite_bonus_multiplier: 0,
          golden_window_extra_draws: 0,
          settled: true,
          settled_bonus_amount: 30,
          settled_at: '2026-05-15T10:00:00Z',
          created_at: '2026-05-15T10:00:00Z',
          updated_at: '2026-05-15T10:00:00Z',
        })),
        multiplier_stats: [],
        golden_window_used_today: 0,
        golden_window_daily_quota: 5,
      },
    })

    const wrapper = mountView()
    await flushPromises()

    const historyList = wrapper.get('[data-test="settlement-history-list"]')

    expect(historyList.classes()).toContain('max-h-[440px]')
    expect(historyList.classes()).toContain('overflow-y-auto')
    expect(wrapper.findAll('[data-test="settlement-history-item"]')).toHaveLength(12)
  })

  it('persists intro and rules fields in save payload', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('textarea[data-test="intro-text"]').setValue('新的活动简介')
    await wrapper.get('input[data-test="rules-title"]').setValue('新的规则标题')
    await wrapper.findAll('input[data-test="rules-item"]')[0].setValue('新的规则 1')
    await wrapper.get('button[data-test="save-config"]').trigger('click')
    await flushPromises()

    expect(updateLuckyWheelConfig).toHaveBeenCalledWith(expect.objectContaining({
      config: expect.objectContaining({
        intro_text: '新的活动简介',
        rules_title: '新的规则标题',
        rules_items: expect.arrayContaining(['新的规则 1']),
      }),
    }))
  })
})
