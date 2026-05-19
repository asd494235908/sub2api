import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RechargeActivityView from '../RechargeActivityView.vue'

const { getRechargeActivitySummary, drawRechargeActivity } = vi.hoisted(() => ({
  getRechargeActivitySummary: vi.fn(),
  drawRechargeActivity: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getRechargeActivitySummary,
    drawRechargeActivity,
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
        if (key === 'rechargeActivity.toastWin' && params?.name != null) return `${key}:${params.name}`
        if (params?.amount != null) return `${key}:${params.amount}`
        if (params?.name != null) return `${key}:${params.name}`
        return key
      },
    }),
  }
})

const summaryPayload = {
  enabled: true,
  config: {
    eligible_order_types: ['balance', 'subscription'],
    intro_text: '充值活动简介',
    rules_title: '充值活动规则',
    rules_items: ['满 20 元可抽三等奖'],
    prizes: [
      { id: 'third', name: '三等奖', reward_amount: 0, reward_description: '联系客服领取实体礼品', probability: 70, min_pay_amount: 20, enabled: true, sort_order: 3 },
      { id: 'second', name: '二等奖', reward_amount: 0, reward_description: '赠送站外会员 30 天', probability: 20, min_pay_amount: 50, enabled: true, sort_order: 2 },
      { id: 'first', name: '一等奖', reward_amount: 0, reward_description: '人工发放定制奖励', probability: 10, min_pay_amount: 100, enabled: true, sort_order: 1 },
    ],
  },
  pending_chances: [
    {
      id: 9,
      user_id: 1,
      source_order_id: 101,
      source_order_type: 'balance',
      source_pay_amount: 20,
      drawn: false,
      created_at: '2026-05-18T10:00:00Z',
      updated_at: '2026-05-18T10:00:00Z',
    },
  ],
  history_records: [
    {
      id: 1,
      chance_id: 7,
      user_id: 1,
      source_order_id: 99,
      prize_id: 'third',
      prize_name: '三等奖',
      reward_amount: 0,
      reward_description: '联系客服领取实体礼品',
      probability: 70,
      min_pay_amount: 20,
      prize_snapshot: '{}',
      eligible_prize_ids: ['third'],
      fulfillment_status: 'pending',
      fulfillment_note: '',
      created_at: '2026-05-18T09:00:00Z',
    },
  ],
}

describe('RechargeActivityView', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    showError.mockReset()
    showSuccess.mockReset()
    getRechargeActivitySummary.mockReset()
    drawRechargeActivity.mockReset()
    getRechargeActivitySummary.mockResolvedValue({ data: summaryPayload })
    drawRechargeActivity.mockResolvedValue({
      data: {
        chance_id: 9,
        record: {
          id: 2,
          chance_id: 9,
          user_id: 1,
          source_order_id: 101,
          prize_id: 'third',
          prize_name: '三等奖',
          reward_amount: 0,
          reward_description: '联系客服领取实体礼品',
          probability: 70,
          min_pay_amount: 20,
          prize_snapshot: '{}',
          eligible_prize_ids: ['third'],
          fulfillment_status: 'pending',
          fulfillment_note: '',
          created_at: '2026-05-18T10:01:00Z',
        },
      },
    })
  })

  function mountView() {
    return mount(RechargeActivityView, {
      attachTo: document.body,
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
        },
      },
    })
  }

  it('renders pending chance, prizes, and history', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('充值活动简介')
    expect(wrapper.text()).toContain('充值活动规则')
    expect(wrapper.text()).toContain('满 20 元可抽三等奖')
    expect(wrapper.text()).toContain('20.00')
    expect(wrapper.text()).toContain('三等奖')
    expect(wrapper.text()).toContain('联系客服领取实体礼品')
    expect(wrapper.text()).not.toContain('rechargeActivity.minPay')
    expect(wrapper.text()).not.toContain('70%')
    expect(wrapper.text()).not.toContain('50.00 · 20%')
    expect(wrapper.text()).not.toContain('+10.00')
    expect(wrapper.find('[data-test="recharge-activity-wheel"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test="recharge-activity-prize"]')).toHaveLength(3)
    expect(wrapper.findAll('[data-test="recharge-activity-history-item"]')).toHaveLength(1)
  })

  it('draws the first pending chance and reveals the result after spinning', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[data-test="draw-recharge-activity"]').trigger('click')
    await flushPromises()

    expect(drawRechargeActivity).toHaveBeenCalledWith(9)
    expect(document.body.querySelector('[data-test="recharge-activity-result-modal"]')).toBeNull()
    await vi.advanceTimersByTimeAsync(4600)
    await flushPromises()
    expect(document.body.querySelector('[data-test="recharge-activity-result-modal"]')).not.toBeNull()
    expect(wrapper.text()).toContain('联系客服领取实体礼品')
    expect(showSuccess).toHaveBeenCalledWith('rechargeActivity.toastWin:三等奖')
    vi.useRealTimers()
  })
})
