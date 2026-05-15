import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LuckyWheelView from '../LuckyWheelView.vue'

const { getLuckyWheelSummary, drawLuckyWheel } = vi.hoisted(() => ({
  getLuckyWheelSummary: vi.fn(),
  drawLuckyWheel: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getLuckyWheelSummary,
    drawLuckyWheel,
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
      t: (key: string, params?: Record<string, unknown>) => {
        if (params?.value != null) return `${key}:${params.value}`
        if (params?.amount != null) return `${key}:${params.amount}`
        return key
      },
    }),
  }
})

const baseSummary = {
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
  active_session: {
    id: 7,
    user_id: 1,
    source_order_id: 100,
    source_order_type: 'balance',
    source_pay_amount: 88,
    matched_tier_id: 'tier_51_plus',
    matched_tier_name: '51+',
    min_multiplier: 1.2,
    max_multiplier: 3,
    total_draws: 4,
    completed_draws: 1,
    remaining_draws: 3,
    best_multiplier: 1.8,
    invite_bonus_multiplier: 0.2,
    golden_window_extra_draws: 1,
    settled: false,
    draw_records: [
      {
        id: 1,
        session_id: 7,
        user_id: 1,
        draw_index: 1,
        base_multiplier: 1.6,
        invite_bonus_multiplier: 0.2,
        final_multiplier: 1.8,
        is_best: true,
        created_at: '2026-05-13T10:00:00Z',
      },
    ],
    created_at: '2026-05-13T10:00:00Z',
    updated_at: '2026-05-13T10:00:00Z',
  },
  pending_sessions: [],
  history_sessions: [
    {
      id: 6,
      user_id: 1,
      source_order_id: 99,
      source_order_type: 'subscription',
      source_pay_amount: 50,
      matched_tier_id: 'tier_20_50',
      matched_tier_name: '20-50',
      min_multiplier: 1.1,
      max_multiplier: 2,
      total_draws: 2,
      completed_draws: 2,
      remaining_draws: 0,
      best_multiplier: 2,
      invite_bonus_multiplier: 0,
      golden_window_extra_draws: 0,
      settled: true,
      settled_bonus_amount: 50,
      settled_at: '2026-05-13T09:00:00Z',
      draw_records: [],
      created_at: '2026-05-13T08:00:00Z',
      updated_at: '2026-05-13T09:00:00Z',
    },
  ],
}

describe('LuckyWheelView', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    getLuckyWheelSummary.mockReset()
    drawLuckyWheel.mockReset()

    getLuckyWheelSummary.mockResolvedValue({
      data: {
        ...baseSummary,
        pending_sessions: [baseSummary.active_session],
      },
    })
    drawLuckyWheel.mockResolvedValue({
      data: {
        session_id: 7,
        best_multiplier: 2.4,
        remaining_draws: 0,
        settled: true,
        settled_bonus_amount: 123.2,
        draw_record: {
          id: 2,
          session_id: 7,
          user_id: 1,
          draw_index: 4,
          base_multiplier: 2.2,
          invite_bonus_multiplier: 0.2,
          final_multiplier: 2.4,
          is_best: true,
          created_at: '2026-05-13T10:01:00Z',
        },
        session: {
          ...baseSummary.active_session,
          completed_draws: 4,
          remaining_draws: 0,
          best_multiplier: 2.4,
          settled: true,
          settled_bonus_amount: 123.2,
          settled_at: '2026-05-13T10:01:00Z',
          draw_records: [
            ...(baseSummary.active_session.draw_records || []),
            {
              id: 2,
              session_id: 7,
              user_id: 1,
              draw_index: 4,
              base_multiplier: 2.2,
              invite_bonus_multiplier: 0.2,
              final_multiplier: 2.4,
              is_best: true,
              created_at: '2026-05-13T10:01:00Z',
            },
          ],
        },
      },
    })
  })

  function mountView() {
    return mount(LuckyWheelView, {
      attachTo: document.body,
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Teleport: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
        },
      },
    })
  }

  it('renders active session and history', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('后台活动简介')
    expect(wrapper.text()).toContain('后台活动规则')
    expect(wrapper.text()).toContain('后台规则 1')
    expect(wrapper.text()).toContain('luckyWheel.rechargeAmount')
    expect(wrapper.text()).toContain('88.00')
    expect(wrapper.text()).toContain('1.8x')
    expect(wrapper.text()).toContain('50.00')
    expect(wrapper.find('[data-test="wheel-board"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="history-scroll-container"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="wheel-rotor"]').attributes('style')).toContain('rotate(')
    expect(wrapper.find('[data-test="wheel-pointer-dot"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="wheel-pointer-arrow"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="wheel-pointer-tip"]').attributes('class')).toContain('top-[2px]')
    expect(wrapper.find('[data-test="wheel-pointer-arrow"]').attributes('class')).toContain('border-t-[28px]')
  })

  it('aligns the winning segment to the visual pointer angle instead of straight top', async () => {
    const wrapper = mountView()
    await flushPromises()

    const vm = wrapper.vm as unknown as {
      getWheelTargetRotation: (multiplier: number) => number
    }

    const normalized = ((vm.getWheelTargetRotation(1.9) % 360) + 360) % 360
    expect(normalized).toBeCloseTo(227.3684210526, 5)
  })

  it('animates the wheel before showing the settlement result', async () => {
    vi.useFakeTimers()
    try {
      const wrapper = mountView()
      await flushPromises()

      await wrapper.get('[data-test="draw-button"]').trigger('click')
      await flushPromises()

      expect(drawLuckyWheel).toHaveBeenCalledTimes(1)
      expect(drawLuckyWheel).toHaveBeenCalledWith(7)
      expect(wrapper.find('[data-test="wheel-rotor"]').classes()).toContain('is-spinning')
      expect(wrapper.get('[data-test="wheel-pointer-tip"]').classes()).toContain('is-ticking')
      expect(wrapper.find('[data-test="result-modal"]').exists()).toBe(false)

      await vi.advanceTimersByTimeAsync(4600)
      await flushPromises()

      expect(wrapper.get('[data-test="wheel-pointer-tip"]').classes()).not.toContain('is-ticking')
      expect(showSuccess).toHaveBeenCalled()
      expect(wrapper.find('[data-test="result-modal"]').exists()).toBe(true)
      expect(wrapper.text()).toContain('luckyWheel.resultHint')
      expect(wrapper.text()).toContain('123.20')
      expect(wrapper.find('[data-test="wheel-rotor"]').attributes('style')).toContain('rotate(')
    } finally {
      vi.useRealTimers()
    }
  })

  it('shows disabled state when the feature is closed', async () => {
    getLuckyWheelSummary.mockResolvedValueOnce({
      data: {
        ...baseSummary,
        enabled: false,
        active_session: null,
        pending_sessions: [],
        history_sessions: [],
      },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('luckyWheel.disabled')
    expect(wrapper.find('[data-test="draw-button"]').attributes('disabled')).toBeDefined()
  })

  it('shows empty state when no pending session exists', async () => {
    getLuckyWheelSummary.mockResolvedValueOnce({
      data: {
        ...baseSummary,
        active_session: null,
        pending_sessions: [],
      },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('luckyWheel.noActiveSession')
    expect(wrapper.find('[data-test="draw-button"]').attributes('disabled')).toBeDefined()
  })

  it('does not render the history scroll container when history is empty', async () => {
    getLuckyWheelSummary.mockResolvedValueOnce({
      data: {
        ...baseSummary,
        history_sessions: [],
      },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-test="history-scroll-container"]').exists()).toBe(false)
  })

  it('falls back to i18n copy when backend intro fields are empty', async () => {
    getLuckyWheelSummary.mockResolvedValueOnce({
      data: {
        ...baseSummary,
        config: {
          ...baseSummary.config,
          intro_text: '',
          rules_title: '',
          rules_items: [],
        },
      },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('luckyWheel.heroDescription')
    expect(wrapper.text()).toContain('luckyWheel.rulesTitle')
    expect(wrapper.text()).toContain('luckyWheel.ruleTier20To50')
  })
})
