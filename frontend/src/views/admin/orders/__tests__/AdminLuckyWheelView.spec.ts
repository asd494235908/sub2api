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

  it('adds a new amount tier row', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[data-test="add-tier"]').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-test="tier-row"]')).toHaveLength(3)
  })
})
