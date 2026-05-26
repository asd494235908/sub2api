import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AffiliatesView from '../AffiliatesView.vue'

const { createInviteRelation, listInviters, listInviterInvitees, lookupUsers, showError, showSuccess } = vi.hoisted(() => ({
  createInviteRelation: vi.fn(),
  listInviters: vi.fn(),
  listInviterInvitees: vi.fn(),
  lookupUsers: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/affiliates', () => {
  const api = {
    createInviteRelation,
    listInviters,
    listInviterInvitees,
    lookupUsers,
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
    listInviters.mockReset()
    listInviterInvitees.mockReset()
    lookupUsers.mockReset()
    createInviteRelation.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

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
