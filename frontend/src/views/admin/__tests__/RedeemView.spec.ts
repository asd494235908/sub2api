import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RedeemView from '../RedeemView.vue'

const { listRedeemCodes, getAllGroups } = vi.hoisted(() => ({
  listRedeemCodes: vi.fn(),
  getAllGroups: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    redeem: {
      list: listRedeemCodes,
      generate: vi.fn(),
      exportCodes: vi.fn(),
      delete: vi.fn(),
      batchDelete: vi.fn(),
    },
    groups: {
      getAll: getAllGroups,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('@/composables/usePersistedPageSize', () => ({
  getPersistedPageSize: () => 20,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.redeem.types.weekly_balance': '周额度领取',
  }
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh' },
      te: (key: string) => Object.prototype.hasOwnProperty.call(messages, key),
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

describe('Admin RedeemView', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: true,
        media: query,
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    })
    listRedeemCodes.mockReset()
    getAllGroups.mockReset()
    listRedeemCodes.mockResolvedValue({
      items: [
        {
          id: -702,
          code: 'LUCKY-702',
          type: 'lucky_wheel_bonus',
          value: 90,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T10:00:00Z',
          created_at: '2026-05-14T09:00:00Z',
          user: { id: 42, email: 'lucky@example.com' },
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    getAllGroups.mockResolvedValue([])
  })

  it('renders lucky wheel bonus type with localized fallback instead of raw i18n key', async () => {
    const wrapper = mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
          },
          Icon: { template: '<span />' },
          Select: {
            props: ['options'],
            template: '<div><span v-for="option in options" :key="option.value">{{ option.label }}</span></div>',
          },
          Pagination: { template: '<div />' },
          ConfirmDialog: { template: '<div />' },
          GroupBadge: { template: '<span />' },
          GroupOptionItem: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('转盘奖励')
    expect(wrapper.text()).not.toContain('admin.redeem.types.lucky_wheel_bonus')
  })
})
