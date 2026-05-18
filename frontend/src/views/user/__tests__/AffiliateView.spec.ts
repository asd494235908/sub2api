import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AffiliateView from '../AffiliateView.vue'
import type { UserAffiliateDetail } from '@/types'

const {
  getAffiliateDetailMock,
  transferAffiliateQuotaMock,
  copyToClipboardMock,
  showErrorMock,
  showSuccessMock,
  refreshUserMock,
} = vi.hoisted(() => ({
  getAffiliateDetailMock: vi.fn(),
  transferAffiliateQuotaMock: vi.fn(),
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
    'affiliate.stats.availableQuota': '可转余额',
    'affiliate.stats.totalQuota': '累计返利',
    'affiliate.stats.frozenQuota': '冻结中',
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
    'affiliate.transfer.title': '返利转余额',
    'affiliate.transfer.description': '将可转返利转入余额',
    'affiliate.transfer.button': '转入余额',
    'affiliate.transfer.transferring': '转入中',
    'affiliate.transfer.success': '已转入 {amount}',
    'affiliate.transfer.empty': '暂无可转返利',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
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
    copyToClipboardMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    refreshUserMock.mockReset()

    getAffiliateDetailMock.mockResolvedValue(createAffiliateDetail())
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

    expect(wrapper.findAll('.card')).toHaveLength(7)
    expect(wrapper.text()).toContain('邀请返利')
    expect(wrapper.text()).toContain('已邀请用户')
    expect(wrapper.text()).toContain('返利比例')
    expect(wrapper.text()).toContain('12.5')
    expect(wrapper.text()).toContain('邀请人数')
    expect(wrapper.text()).toContain('可转余额')
    expect(wrapper.text()).toContain('累计返利')
    expect(wrapper.text()).toContain('line1')
    expect(wrapper.text()).toContain('line2')
    expect(wrapper.text()).toContain('line3')
    expect(wrapper.text()).toContain('返利转余额')
    expect(wrapper.text()).toContain('转入余额')
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
})
