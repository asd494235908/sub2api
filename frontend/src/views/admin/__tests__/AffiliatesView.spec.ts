import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AffiliatesView from '../AffiliatesView.vue'

const {
  approveWithdrawal,
  createInviteRelation,
  getIdentityConfig,
  getWithdrawSettings,
  listInviters,
  listInviterInvitees,
  listWithdrawals,
  lookupUsers,
  markWithdrawalFailed,
  markWithdrawalPaid,
  rejectWithdrawal,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  approveWithdrawal: vi.fn(),
  createInviteRelation: vi.fn(),
  getIdentityConfig: vi.fn(),
  getWithdrawSettings: vi.fn(),
  listInviters: vi.fn(),
  listInviterInvitees: vi.fn(),
  listWithdrawals: vi.fn(),
  lookupUsers: vi.fn(),
  markWithdrawalFailed: vi.fn(),
  markWithdrawalPaid: vi.fn(),
  rejectWithdrawal: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/affiliates', () => {
  const api = {
    approveWithdrawal,
    createInviteRelation,
    getIdentityConfig,
    getWithdrawSettings,
    listInviters,
    listInviterInvitees,
    listWithdrawals,
    lookupUsers,
    markWithdrawalFailed,
    markWithdrawalPaid,
    rejectWithdrawal,
  }
  return {
    affiliatesAPI: api,
    default: api,
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: () => 'error',
}))

vi.mock('@/utils/format', () => ({
  formatCurrency: (value: number, currency?: string) => `${currency ?? 'USD'}:${value.toFixed(2)}`,
  formatDateTime: (value: string) => `FMT:${value}`,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.affiliates.title': '邀请关系',
    'admin.affiliates.description': '查看所有邀请人的邀请关系与返利明细',
    'admin.affiliates.searchPlaceholder': '搜索邀请人邮箱或用户名',
    'admin.affiliates.empty': '暂无邀请关系数据',
    'admin.affiliates.viewInvitees': '查看邀请用户',
    'admin.affiliates.manualAdd': '手动添加关系',
    'admin.affiliates.manualDialogTitle': '手动添加邀请关系',
    'admin.affiliates.manualHint': '允许覆盖已有上级，不补发历史充值返利。',
    'admin.affiliates.manualInviter': '邀请人',
    'admin.affiliates.manualInvitee': '被邀请人',
    'admin.affiliates.manualSearchPlaceholder': '搜索用户邮箱、用户名或 ID',
    'admin.affiliates.manualNoOptions': '暂无匹配用户',
    'admin.affiliates.manualSelectedUser': '已选择',
    'admin.affiliates.manualSubmit': '保存关系',
    'admin.affiliates.manualSubmitting': '保存中',
    'admin.affiliates.manualSameUserError': '邀请人和被邀请人不能是同一用户',
    'admin.affiliates.manualMissingUserError': '请选择邀请人和被邀请人',
    'admin.affiliates.manualSuccess': '邀请关系已保存',
    'admin.affiliates.totalLabel': '共 {total} 条',
    'admin.affiliates.inviteesTitle': '{email} 邀请的用户',
    'admin.affiliates.inviteesDescription': '展示该邀请人当前关联的被邀请用户明细。',
    'admin.affiliates.inviteesEmpty': '该邀请人暂无可展示的邀请用户',
    'admin.affiliates.col.email': '邀请人邮箱',
    'admin.affiliates.col.username': '邀请人用户名',
    'admin.affiliates.col.code': '邀请码',
    'admin.affiliates.col.invitedCount': '邀请人数',
    'admin.affiliates.col.totalRebate': '累计返利',
    'admin.affiliates.col.actions': '操作',
    'admin.affiliates.inviteesCol.email': '被邀请用户邮箱',
    'admin.affiliates.inviteesCol.username': '被邀请用户名',
    'admin.affiliates.inviteesCol.joinedAt': '加入时间',
    'admin.affiliates.inviteesCol.totalRebate': '累计返利',
    'admin.affiliates.identityConfig.title': '邀请身份倍率',
    'admin.affiliates.identityConfig.description': '配置邀请身份',
    'admin.affiliates.identityConfig.enabled': '启用邀请身份',
    'admin.affiliates.identityConfig.inviterRate': '邀请人倍率',
    'admin.affiliates.identityConfig.inviteeRate': '新用户倍率',
    'admin.affiliates.identityConfig.durationHours': '有效期',
    'admin.affiliates.identityConfig.qualifiedPayAmount': '实付门槛',
    'admin.affiliates.identityConfig.qualifiedInviteeCount': '人数门槛',
    'admin.affiliates.identityConfig.maxAccounts': '账号上限',
    'admin.affiliates.identityConfig.orderTypes': '计入订单',
    'admin.affiliates.identityConfig.balanceOrder': '余额充值',
    'admin.affiliates.identityConfig.subscriptionOrder': '订阅购买',
    'admin.affiliates.identityConfig.fingerprintEnabled': '启用指纹风控',
    'admin.affiliates.identityConfig.save': '保存配置',
    'admin.affiliates.identityConfig.identityNone': '未获得',
    'admin.affiliates.withdraw.title': '返利提现管理',
    'admin.affiliates.withdraw.description': '审核用户提现申请',
    'admin.affiliates.withdraw.enabled': '启用返利提现',
    'admin.affiliates.withdraw.minAmount': '最低金额',
    'admin.affiliates.withdraw.maxAmount': '单笔上限',
    'admin.affiliates.withdraw.dailyLimit': '每日次数',
    'admin.affiliates.withdraw.helpText': '提现说明',
    'admin.affiliates.withdraw.saveSettings': '保存提现配置',
    'admin.affiliates.withdraw.searchPlaceholder': '搜索邮箱、用户名、收款说明或流水号',
    'admin.affiliates.withdraw.allStatuses': '全部状态',
    'admin.affiliates.withdraw.empty': '暂无提现申请',
    'admin.affiliates.withdraw.col.user': '用户',
    'admin.affiliates.withdraw.col.amount': '金额',
    'admin.affiliates.withdraw.col.status': '状态',
    'admin.affiliates.withdraw.col.account': '收款说明',
    'admin.affiliates.withdraw.col.tradeNo': '发放流水',
    'admin.affiliates.withdraw.col.createdAt': '申请时间',
    'admin.affiliates.withdraw.col.actions': '操作',
    'admin.affiliates.withdraw.status.pending_review': '待审核',
    'admin.affiliates.withdraw.status.approved': '待发放',
    'admin.affiliates.withdraw.status.paid': '已发放',
    'admin.affiliates.withdraw.status.rejected': '已驳回',
    'admin.affiliates.withdraw.status.failed': '发放失败',
    'admin.affiliates.withdraw.actions.approve': '通过',
    'admin.affiliates.withdraw.actions.reject': '驳回',
    'admin.affiliates.withdraw.actions.paid': '已发放',
    'admin.affiliates.withdraw.actions.fail': '失败',
    'pagination.previous': '上一页',
    'pagination.next': '下一页',
    'common.loading': '加载中',
    'common.close': '关闭',
    'common.error': '错误',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) =>
        (messages[key] ?? key).replace(/\{(\w+)\}/g, (_, token) => params?.[token] ?? `{${token}}`),
    }),
  }
})

describe('AffiliatesView', () => {
  beforeEach(() => {
    approveWithdrawal.mockReset()
    listInviters.mockReset()
    listInviterInvitees.mockReset()
    listWithdrawals.mockReset()
    lookupUsers.mockReset()
    createInviteRelation.mockReset()
    getIdentityConfig.mockReset()
    getWithdrawSettings.mockReset()
    markWithdrawalFailed.mockReset()
    markWithdrawalPaid.mockReset()
    rejectWithdrawal.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    getIdentityConfig.mockResolvedValue({
      enabled: false,
      config: {
        inviter_rate_multiplier: 1.5,
        invitee_rate_multiplier: 1.4,
        duration_hours: 720,
        qualified_invitee_count: 0,
        qualified_pay_amount: 50,
        eligible_order_types: ['balance', 'subscription'],
        fingerprint_enforcement_enabled: true,
        max_accounts_per_fingerprint_hash: 3,
      },
    })
    getWithdrawSettings.mockResolvedValue({
      enabled: true,
      min_amount: 10,
      max_amount: 50,
      daily_request_limit: 10,
      help_text: '本地测试人工微信提现',
    })
    listWithdrawals.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0,
    })
    listInviters.mockResolvedValue({
      items: [
        {
          user_id: 11,
          email: 'owner@example.com',
          username: 'owner',
          aff_code: 'AFFOWNER',
          aff_count: 2,
          total_rebate: 18.8,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listInviterInvitees.mockResolvedValue([
      {
        user_id: 101,
        email: 'friend@example.com',
        username: 'friend',
        created_at: '2026-04-01T00:00:00Z',
        total_rebate: 6.6,
      },
    ])
    lookupUsers.mockResolvedValue([])
    createInviteRelation.mockResolvedValue({
      inviter_user_id: 11,
      invitee_user_id: 22,
      overwritten: false,
      previous_inviter_user_id: null,
    })
  })

  it('keeps withdrawal actions compact and shows placeholders for terminal statuses', async () => {
    listWithdrawals.mockResolvedValueOnce({
      items: [
        {
          id: 1,
          user_id: 1051,
          user_email: 'withdraw-user@example.test',
          username: 'withdraw-user-local',
          amount: 10,
          status: 'failed',
          payout_account_note: '微信号 fail_case',
          payout_trade_no: '',
          created_at: '2026-06-01T00:41:17Z',
        },
        {
          id: 2,
          user_id: 1051,
          user_email: 'withdraw-user@example.test',
          username: 'withdraw-user-local',
          amount: 20,
          status: 'approved',
          payout_account_note: '微信号 approved_case',
          payout_trade_no: '',
          created_at: '2026-06-01T00:29:19Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1,
    })

    const wrapper = mount(AffiliatesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[data-test="withdraw-table"]').classes()).toEqual(expect.arrayContaining(['w-full', 'table-fixed']))
    expect(wrapper.find('[data-test="withdraw-actions-header"]').classes()).toContain('w-40')
    expect(wrapper.find('[data-test="withdraw-actions-placeholder"]').text()).toBe('-')
    expect(wrapper.text()).toContain('已发放')
    expect(wrapper.text()).toContain('失败')
  })

  it('loads inviters on mount and shows invitee details on demand', async () => {
    const wrapper = mount(AffiliatesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(listInviters).toHaveBeenCalled()
    expect(wrapper.text()).toContain('邀请关系')
    expect(wrapper.text()).toContain('owner@example.com')
    expect(wrapper.text()).toContain('AFFOWNER')
    expect(wrapper.text()).toContain('CNY:18.80')

    const button = wrapper.findAll('button').find((node) => node.text().includes('查看邀请用户'))
    expect(button).toBeDefined()
    await button?.trigger('click')
    await flushPromises()

    expect(listInviterInvitees).toHaveBeenCalledWith(11)
    expect(wrapper.text()).toContain('owner@example.com 邀请的用户')
    expect(wrapper.text()).toContain('friend@example.com')
    expect(wrapper.text()).toContain('FMT:2026-04-01T00:00:00Z')
    expect(wrapper.text()).toContain('CNY:6.60')
  })

  it('reloads inviters with date range filters and timezone', async () => {
    try {
      const wrapper = mount(AffiliatesView, {
        global: {
          stubs: {
            AppLayout: { template: '<div><slot /></div>' },
          },
        },
      })

      await flushPromises()
      listInviters.mockClear()

      const inputs = wrapper.findAll('input[type="date"]')
      expect(inputs).toHaveLength(2)

      await inputs[0].setValue('2026-05-01')
      await inputs[0].trigger('change')
      await flushPromises()

      expect(listInviters).toHaveBeenCalledWith({
        page: 1,
        page_size: 20,
        search: '',
        start_at: '2026-05-01',
        end_at: undefined,
        timezone: expect.any(String),
      })

      await inputs[1].setValue('2026-05-31')
      await inputs[1].trigger('change')
      await flushPromises()

      expect(listInviters).toHaveBeenCalledWith({
        page: 1,
        page_size: 20,
        search: '',
        start_at: '2026-05-01',
        end_at: '2026-05-31',
        timezone: expect.any(String),
      })
    } finally {}
  })

  it('opens manual relation dialog and submits overwrite payload', async () => {
    vi.useFakeTimers()
    try {
      lookupUsers
        .mockResolvedValueOnce([{ id: 11, email: 'owner@example.com', username: 'owner' }])
        .mockResolvedValueOnce([{ id: 22, email: 'friend@example.com', username: 'friend' }])

      const wrapper = mount(AffiliatesView, {
        global: {
          stubs: {
            AppLayout: { template: '<div><slot /></div>' },
            BaseDialog: { template: '<div v-if="show"><slot /><slot name="footer" /></div>', props: ['show', 'title'] },
          },
        },
      })
      await flushPromises()

      await wrapper.findAll('button').find((node) => node.text().includes('手动添加关系'))?.trigger('click')
      await flushPromises()
      expect(wrapper.text()).toContain('允许覆盖已有上级')

      const vm = wrapper.vm as unknown as {
        manualState: {
          inviterQuery: string
          inviteeQuery: string
        }
        searchManualUsers: (role: 'inviter' | 'invitee') => Promise<void>
        selectManualUser: (role: 'inviter' | 'invitee', user: { id: number; email: string; username: string }) => void
      }
      vm.manualState.inviterQuery = 'owner'
      await vm.searchManualUsers('inviter')
      await flushPromises()
      vm.selectManualUser('inviter', { id: 11, email: 'owner@example.com', username: 'owner' })

      vm.manualState.inviteeQuery = 'friend'
      await vm.searchManualUsers('invitee')
      await flushPromises()
      vm.selectManualUser('invitee', { id: 22, email: 'friend@example.com', username: 'friend' })

      await wrapper.findAll('button').find((node) => node.text().includes('保存关系'))?.trigger('click')
      await flushPromises()

      expect(createInviteRelation).toHaveBeenCalledWith({
        inviter_user_id: 11,
        invitee_user_id: 22,
        overwrite: true,
      })
      expect(showSuccess).toHaveBeenCalledWith('邀请关系已保存')
      expect(listInviters).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('blocks selecting the same user in manual relation dialog', async () => {
    const wrapper = mount(AffiliatesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          BaseDialog: { template: '<div v-if="show"><slot /><slot name="footer" /></div>', props: ['show', 'title'] },
        },
      },
    })
    await flushPromises()

    const vm = wrapper.vm as unknown as {
      manualState: {
        inviter: { id: number; email: string; username: string } | null
        invitee: { id: number; email: string; username: string } | null
      }
      submitManualRelation: () => Promise<void>
    }
    vm.manualState.inviter = { id: 11, email: 'same@example.com', username: 'same' }
    vm.manualState.invitee = { id: 11, email: 'same@example.com', username: 'same' }
    await vm.submitManualRelation()

    expect(createInviteRelation).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('邀请人和被邀请人不能是同一用户')
  })
})
