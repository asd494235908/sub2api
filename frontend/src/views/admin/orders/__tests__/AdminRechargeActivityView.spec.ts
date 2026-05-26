import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminRechargeActivityView from '../AdminRechargeActivityView.vue'

const {
  getRechargeActivityConfig,
  getRechargeActivityStats,
  updateRechargeActivityConfig,
  updateRechargeActivityRecordFulfillment,
  getFirstRechargeConfig,
  getMemberLevelConfig,
  updateFirstRechargeConfig,
  updateMemberLevelConfig,
  grantFirstRechargeChance,
  bulkUpdateFirstRechargeChances,
} = vi.hoisted(() => ({
  getRechargeActivityConfig: vi.fn(),
  getRechargeActivityStats: vi.fn(),
  updateRechargeActivityConfig: vi.fn(),
  updateRechargeActivityRecordFulfillment: vi.fn(),
  getFirstRechargeConfig: vi.fn(),
  getMemberLevelConfig: vi.fn(),
  updateFirstRechargeConfig: vi.fn(),
  updateMemberLevelConfig: vi.fn(),
  grantFirstRechargeChance: vi.fn(),
  bulkUpdateFirstRechargeChances: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getRechargeActivityConfig,
    getRechargeActivityStats,
    updateRechargeActivityConfig,
    updateRechargeActivityRecordFulfillment,
    getFirstRechargeConfig,
    getMemberLevelConfig,
    updateFirstRechargeConfig,
    updateMemberLevelConfig,
    grantFirstRechargeChance,
    bulkUpdateFirstRechargeChances,
  },
  default: {
    getRechargeActivityConfig,
    getRechargeActivityStats,
    updateRechargeActivityConfig,
    updateRechargeActivityRecordFulfillment,
    getFirstRechargeConfig,
    getMemberLevelConfig,
    updateFirstRechargeConfig,
    updateMemberLevelConfig,
    grantFirstRechargeChance,
    bulkUpdateFirstRechargeChances,
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
    intro_text: '充值活动简介',
    rules_title: '充值活动规则',
    rules_items: ['规则 1'],
    prizes: [
      { id: 'third', name: '三等奖', reward_amount: 0, reward_description: '联系客服领取实体礼品', probability: 70, min_pay_amount: 20, enabled: true, sort_order: 3 },
      { id: 'second', name: '二等奖', reward_amount: 0, reward_description: '赠送站外会员 30 天', probability: 20, min_pay_amount: 50, enabled: true, sort_order: 2 },
      { id: 'first', name: '一等奖', reward_amount: 0, reward_description: '人工发放定制奖励', probability: 10, min_pay_amount: 100, enabled: true, sort_order: 1 },
    ],
  },
}

describe('AdminRechargeActivityView', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    getRechargeActivityConfig.mockReset()
    getRechargeActivityStats.mockReset()
    updateRechargeActivityConfig.mockReset()
    updateRechargeActivityRecordFulfillment.mockReset()
    getFirstRechargeConfig.mockReset()
    getMemberLevelConfig.mockReset()
    updateFirstRechargeConfig.mockReset()
    updateMemberLevelConfig.mockReset()
    grantFirstRechargeChance.mockReset()
    bulkUpdateFirstRechargeChances.mockReset()

    getRechargeActivityConfig.mockResolvedValue({ data: configPayload })
    getFirstRechargeConfig.mockResolvedValue({
      data: {
        enabled: true,
        config: {
          tiers: [
            { id: 'tier-30', pay_amount: 30, bonus_amount: 15, enabled: true, sort_order: 1 },
          ],
        },
      },
    })
    getMemberLevelConfig.mockResolvedValue({
      data: {
        enabled: true,
        config: {
          levels: [
            { id: 'default', name: '默认等级', min_recharge_amount: 0, rate_multiplier: 1, enabled: true, sort_order: 1 },
          ],
        },
      },
    })
    getRechargeActivityStats.mockResolvedValue({
      data: {
        enabled: true,
        total_chances: 5,
        pending_chances: 2,
        drawn_chances: 3,
        pending_fulfillments: 1,
        fulfilled_records: 2,
        total_reward_amount: 0,
        recent_records_total: 1,
        recent_records_page: 1,
        recent_records_page_size: 20,
        recent_records_keyword: '',
        recent_records: [
          {
            id: 8,
            chance_id: 7,
            user_id: 1,
            user_email: 'admin@sub2api.local',
            user_name: 'Admin',
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
            created_at: '2026-05-18T10:00:00Z',
          },
        ],
      },
    })
    updateRechargeActivityConfig.mockResolvedValue({ data: configPayload })
    updateRechargeActivityRecordFulfillment.mockResolvedValue({ data: { id: 8, fulfillment_status: 'fulfilled' } })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  function mountView() {
    return mount(AdminRechargeActivityView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Pagination: {
            props: ['total', 'page', 'pageSize'],
            emits: ['update:page', 'update:pageSize'],
            template: '<div data-test="records-pagination"><button data-test="go-page-2" @click="$emit(\'update:page\', 2)">page 2</button><button data-test="set-page-size-50" @click="$emit(\'update:pageSize\', 50)">size 50</button></div>',
          },
          Toggle: {
            props: ['modelValue'],
            emits: ['update:modelValue'],
            template: '<button data-test="toggle" @click="$emit(\'update:modelValue\', !modelValue)">toggle</button>',
          },
        },
      },
    })
  }

  it('renders configurable prizes and stats', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('首冲赠送额度')
    expect(wrapper.text()).toContain('赠送平台额度')
    expect(wrapper.text()).not.toContain('首冲返额')
    expect(wrapper.findAll('[data-test="recharge-prize-row"]')).toHaveLength(3)
    expect(wrapper.text()).toContain('联系客服领取实体礼品')
    expect(wrapper.text()).toContain('admin@sub2api.local')
    expect(wrapper.text()).toContain('Admin')
    expect(wrapper.get('[data-test="recharge-record-search"]').attributes('placeholder')).toBe('rechargeActivity.adminSearchUserPlaceholder')
    expect(wrapper.findAll('[data-test="recharge-record-row"]')).toHaveLength(1)
    expect(wrapper.find('[data-test="records-pagination"]').exists()).toBe(true)
    expect((wrapper.get('textarea[data-test="intro-text"]').element as HTMLTextAreaElement).value).toBe('充值活动简介')
  })

  it('searches records by winning user and resets to the first page', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await flushPromises()
    getRechargeActivityStats.mockClear()

    await wrapper.get('[data-test="go-page-2"]').trigger('click')
    await flushPromises()
    expect(getRechargeActivityStats).toHaveBeenLastCalledWith({
      page: 2,
      page_size: 20,
      user_keyword: undefined,
    })

    await wrapper.get('input[data-test="recharge-record-search"]').setValue('admin@sub2api.local')
    await vi.advanceTimersByTimeAsync(350)
    await flushPromises()

    expect(getRechargeActivityStats).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      user_keyword: 'admin@sub2api.local',
    })
  })

  it('keeps the current search when changing record pagination', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input[data-test="recharge-record-search"]').setValue('admin')
    await vi.advanceTimersByTimeAsync(350)
    await flushPromises()
    getRechargeActivityStats.mockClear()

    await wrapper.get('[data-test="go-page-2"]').trigger('click')
    await flushPromises()
    expect(getRechargeActivityStats).toHaveBeenLastCalledWith({
      page: 2,
      page_size: 20,
      user_keyword: 'admin',
    })

    await wrapper.get('[data-test="set-page-size-50"]').trigger('click')
    await flushPromises()
    expect(getRechargeActivityStats).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 50,
      user_keyword: 'admin',
    })
  })

  it('blocks save when enabled probabilities are invalid', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('input[data-test="prize-probability"]')[0].setValue('0')
    await flushPromises()

    expect(wrapper.text()).toContain('rechargeActivity.adminValidationProbability')
    expect(wrapper.get('button[data-test="save-config"]').attributes('disabled')).toBeDefined()
  })

  it('adds a prize and persists threshold and probability', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[data-test="add-prize"]').trigger('click')
    await flushPromises()
    const rows = wrapper.findAll('[data-test="recharge-prize-row"]')
    expect(rows).toHaveLength(4)

    const newRow = rows[3]
    await newRow.get('input[data-test="prize-id"]').setValue('special')
    await newRow.get('input[data-test="prize-name"]').setValue('特别奖')
    await newRow.get('textarea[data-test="prize-reward-description"]').setValue('线下人工发放')
    await newRow.get('input[data-test="prize-probability"]').setValue('5')
    await newRow.get('input[data-test="prize-min-pay"]').setValue('10')
    await rows[0].get('input[data-test="prize-probability"]').setValue('65')
    await wrapper.get('button[data-test="save-config"]').trigger('click')
    await flushPromises()

    expect(updateRechargeActivityConfig).toHaveBeenCalledWith(expect.objectContaining({
      enabled: true,
      config: expect.objectContaining({
        prizes: expect.arrayContaining([
          expect.objectContaining({
            id: 'special',
            name: '特别奖',
            reward_description: '线下人工发放',
            probability: 5,
            min_pay_amount: 10,
          }),
        ]),
      }),
    }))
  })

  it('updates manual fulfillment status for a recent record', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input[data-test="fulfillment-note"]').setValue('已联系用户')
    await wrapper.get('button[data-test="toggle-fulfillment"]').trigger('click')
    await flushPromises()

    expect(updateRechargeActivityRecordFulfillment).toHaveBeenCalledWith(8, {
      status: 'fulfilled',
      note: '已联系用户',
    })
    expect(getRechargeActivityStats).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      user_keyword: undefined,
    })
  })

  it('requires a note before marking a record fulfilled', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input[data-test="fulfillment-note"]').setValue('   ')
    await wrapper.get('button[data-test="toggle-fulfillment"]').trigger('click')
    await flushPromises()

    expect(updateRechargeActivityRecordFulfillment).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('rechargeActivity.adminFulfillmentNoteRequired')
  })

  it('allows reverting a fulfilled record to pending without a note', async () => {
    getRechargeActivityStats.mockResolvedValueOnce({
      data: {
        enabled: true,
        total_chances: 5,
        pending_chances: 2,
        drawn_chances: 3,
        pending_fulfillments: 0,
        fulfilled_records: 3,
        total_reward_amount: 0,
        recent_records_total: 1,
        recent_records_page: 1,
        recent_records_page_size: 20,
        recent_records_keyword: '',
        recent_records: [
          {
            id: 8,
            chance_id: 7,
            user_id: 1,
            user_email: 'admin@sub2api.local',
            user_name: 'Admin',
            source_order_id: 101,
            prize_id: 'third',
            prize_name: '三等奖',
            reward_amount: 0,
            reward_description: '联系客服领取实体礼品',
            probability: 70,
            min_pay_amount: 20,
            prize_snapshot: '{}',
            eligible_prize_ids: ['third'],
            fulfillment_status: 'fulfilled',
            fulfillment_note: '',
            created_at: '2026-05-18T10:00:00Z',
          },
        ],
      },
    })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('input[data-test="fulfillment-note"]').setValue('   ')
    await wrapper.get('button[data-test="toggle-fulfillment"]').trigger('click')
    await flushPromises()

    expect(updateRechargeActivityRecordFulfillment).toHaveBeenCalledWith(8, {
      status: 'pending',
      note: '',
    })
  })
})
