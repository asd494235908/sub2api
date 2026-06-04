import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AffiliateView from '../AffiliateView.vue'
import type { UserAffiliateDetail } from '@/types'

const {
  getAffiliateDetailMock,
  transferAffiliateQuotaMock,
  listAffiliateRecordsMock,
  createAffiliateWithdrawalMock,
  copyToClipboardMock,
  showErrorMock,
  showSuccessMock,
  refreshUserMock,
} = vi.hoisted(() => ({
  getAffiliateDetailMock: vi.fn(),
  transferAffiliateQuotaMock: vi.fn(),
  listAffiliateRecordsMock: vi.fn(),
  createAffiliateWithdrawalMock: vi.fn(),
  copyToClipboardMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  refreshUserMock: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  __esModule: true,
  default: {
    getAffiliateDetail: getAffiliateDetailMock,
    transferAffiliateQuota: transferAffiliateQuotaMock,
    listAffiliateRecords: listAffiliateRecordsMock,
    listAffiliateWithdrawals: vi.fn().mockResolvedValue({ items: [], total: 0 }),
    createAffiliateWithdrawal: createAffiliateWithdrawalMock,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser: refreshUserMock,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: copyToClipboardMock,
  }),
}))

vi.mock('@/utils/format', () => ({
  formatCurrency: (value: number) => `$${value.toFixed(2)}`,
  formatDateTime: (value: string) => value,
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: () => 'load failed',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'affiliate.title': '邀请返利',
    'affiliate.description': '邀请返利说明',
    'affiliate.stats.rebateRate': '返利比例',
    'affiliate.stats.rebateRateHint': '返利比例说明',
    'affiliate.stats.invitedUsers': '邀请人数',
    'affiliate.stats.availableQuota': '可提现现金返利',
    'affiliate.stats.totalQuota': '累计现金返利',
    'affiliate.stats.frozenQuota': '待解冻现金返利',
    'affiliate.yourCode': '我的邀请码',
    'affiliate.copyCode': '复制邀请码',
    'affiliate.inviteLink': '邀请链接',
    'affiliate.copyLink': '复制链接',
    'affiliate.tips.title': '返利说明',
    'affiliate.tips.line1': 'line1',
    'affiliate.tips.line2': 'line2',
    'affiliate.tips.line3': 'line3',
    'affiliate.tips.line4': 'line4',
    'affiliate.invitees.title': '已邀请用户',
    'affiliate.invitees.empty': '暂无邀请记录',
    'affiliate.invitees.columns.email': '邮箱',
    'affiliate.invitees.columns.username': '用户名',
    'affiliate.invitees.columns.rebate': '返利',
    'affiliate.invitees.columns.joinedAt': '加入时间',
    'affiliate.transfer.title': '现金返利转平台余额',
    'affiliate.transfer.description': '按充值倍率转入平台余额，提现按现金发放',
    'affiliate.transfer.button': '转为平台余额',
    'affiliate.transfer.transferring': '转入中',
    'affiliate.transfer.success': '已转入 {amount}',
    'affiliate.transfer.empty': '暂无可转现金返利',
    'affiliate.records.title': '现金返利流水',
    'affiliate.records.description': '返利、兑换、提现记录',
    'affiliate.records.empty': '暂无现金返利流水',
    'affiliate.records.loadFailed': '加载现金返利流水失败',
    'affiliate.records.columns.action': '类型',
    'affiliate.records.columns.amount': '金额',
    'affiliate.records.columns.source': '来源',
    'affiliate.records.columns.balanceAfter': '现金余额',
    'affiliate.records.columns.createdAt': '时间',
    'affiliate.records.actions.accrue': '返利入账',
    'affiliate.records.actions.transfer': '转平台余额',
    'affiliate.withdraw.title': '提现到微信',
    'affiliate.withdraw.description': '人工审核发放',
    'affiliate.withdraw.amount': '金额',
    'affiliate.withdraw.payoutMethod': '收款方式',
    'affiliate.withdraw.methods.wechat_manual': '微信人工转账',
    'affiliate.withdraw.accountNote': '收款说明',
    'affiliate.withdraw.submit': '申请提现',
    'affiliate.withdraw.submitting': '提交中',
    'affiliate.withdraw.empty': '暂无可提现',
    'affiliate.withdraw.noRecords': '暂无提现记录',
    'affiliate.withdraw.columns.amount': '金额',
    'affiliate.withdraw.columns.status': '状态',
    'affiliate.withdraw.columns.account': '收款说明',
    'affiliate.withdraw.columns.tradeNo': '流水',
    'affiliate.withdraw.columns.createdAt': '申请时间',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        let message = messages[key] ?? key
        for (const [name, value] of Object.entries(params ?? {})) {
          message = message.replace(`{${name}}`, String(value))
        }
        return message
      },
    }),
  }
})

function createAffiliateDetail(): UserAffiliateDetail {
  return {
    user_id: 1,
    aff_code: 'AFF123',
    inviter_id: null,
    aff_count: 3,
    aff_quota: 10,
    aff_frozen_quota: 0,
    aff_history_quota: 25,
    rebate_cash_balance: 9.9,
    frozen_rebate_cash: 0,
    total_rebate_cash: 25,
    effective_rebate_rate_percent: 12.5,
    invitees: [
      {
        user_id: 2,
        email: 'friend@example.com',
        username: 'friend',
        total_rebate: 4.5,
        created_at: '2026-04-01T00:00:00Z',
      },
    ],
  }
}

describe('AffiliateView', () => {
  beforeEach(() => {
    vi.stubEnv('BASE_URL', '/portal/')
    getAffiliateDetailMock.mockReset()
    transferAffiliateQuotaMock.mockReset()
    listAffiliateRecordsMock.mockReset()
    createAffiliateWithdrawalMock.mockReset()
    copyToClipboardMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    refreshUserMock.mockReset()

    getAffiliateDetailMock.mockResolvedValue(createAffiliateDetail())
    createAffiliateWithdrawalMock.mockResolvedValue({ id: 1 })
    listAffiliateRecordsMock.mockResolvedValue({
      items: [
        {
          ledger_id: 1,
          action: 'accrue',
          amount: 9.9,
          source_user_email: 'friend@example.com',
          rebate_cash_after: 9.9,
          created_at: '2026-04-02T00:00:00Z',
        },
      ],
      total: 1,
    })
  })

  it('restores the affiliate stats, guidance, and transfer cards', async () => {
    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.findAll('.card')).toHaveLength(9)
    expect(wrapper.text()).toContain('邀请返利')
    expect(wrapper.text()).toContain('已邀请用户')
    expect(wrapper.text()).toContain('返利比例')
    expect(wrapper.text()).toContain('12.5')
    expect(wrapper.text()).toContain('邀请人数')
    expect(wrapper.text()).toContain('可提现现金返利')
    expect(wrapper.text()).toContain('$9.90')
    expect(wrapper.text()).toContain('累计现金返利')
    expect(wrapper.text()).toContain('line1')
    expect(wrapper.text()).toContain('line2')
    expect(wrapper.text()).toContain('line3')
    expect(wrapper.text()).toContain('现金返利转平台余额')
    expect(wrapper.text()).toContain('转为平台余额')
    expect(wrapper.text()).toContain('提现到微信')
    expect(wrapper.text()).toContain('申请提现')
    expect(wrapper.text()).toContain('现金返利流水')
    expect(wrapper.text()).toContain('返利入账')
    expect(wrapper.text()).toContain('friend@example.com')
  })

  it('builds invite links with the configured router base path and aff parameter', async () => {
    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('http://localhost:3000/portal/register?aff=AFF123')
  })

  it('shows the withdrawal fields and submits the payout details', async () => {
    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('金额')
    expect(wrapper.text()).toContain('收款方式')
    expect(wrapper.text()).toContain('微信人工转账')
    expect(wrapper.text()).toContain('收款说明')

    await wrapper.find('input[type="number"]').setValue('9')
    await wrapper.find('select').setValue('wechat_manual')
    await wrapper.find('input[type="text"]').setValue('微信 test')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createAffiliateWithdrawalMock).toHaveBeenCalledWith({
      amount: 9,
      payout_method: 'wechat_manual',
      payout_account_note: '微信 test',
    })
  })

  it('keeps withdrawal controls in one row and disables inputs when no cash rebate is available', async () => {
    getAffiliateDetailMock.mockResolvedValue({
      ...createAffiliateDetail(),
      aff_quota: 0,
      rebate_cash_balance: 0,
    })

    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    const form = wrapper.get('[data-test="affiliate-withdraw-form"]')
    expect(form.classes()).toContain('xl:grid-cols-[minmax(120px,0.8fr)_minmax(150px,1fr)_minmax(220px,1.4fr)_auto]')
    expect(form.get('input[type="number"]').attributes('disabled')).toBeDefined()
    expect(form.get('select').attributes('disabled')).toBeDefined()
    expect(form.get('input[type="text"]').attributes('disabled')).toBeDefined()
    expect(form.get('[data-test="affiliate-withdraw-submit"]').attributes('disabled')).toBeDefined()
  })

  it('shows the credited platform balance amount after converting cash rebate', async () => {
    transferAffiliateQuotaMock.mockResolvedValue({
      transferred_cash: 10,
      transferred_quota: 130,
      balance: 230,
      platform_balance: 230,
    })

    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(showSuccessMock).toHaveBeenCalledWith('已转入 $130.00')
  })
})
