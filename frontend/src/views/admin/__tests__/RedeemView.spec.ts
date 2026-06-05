import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'

import RedeemView from '../RedeemView.vue'

const DataTableStub = defineComponent({
  props: ['columns', 'data', 'loading'],
  setup(props, { slots }) {
    return () => h('div', [
      ...(props.columns || []).map((column: any) => h('span', { key: `heading-${column.key}` }, column.label)),
      ...(props.data || []).flatMap((row: any) =>
        (props.columns || []).map((column: any) => {
          const slot = slots[`cell-${column.key}`]
          return h('div', { key: `${row.id}-${column.key}` }, slot
            ? slot({ value: row[column.key], row })
            : String(row[column.key] ?? ''))
        }),
      ),
    ])
  },
})

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
          DataTable: DataTableStub,
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

  it('renders first recharge bonus value without a Chinese unit', async () => {
    listRedeemCodes.mockResolvedValue({
      items: [
        {
          id: -703,
          code: 'FIRST-RECHARGE-30',
          type: 'first_recharge_bonus',
          value: 15,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T10:00:00Z',
          created_at: '2026-05-14T09:00:00Z',
          user: { id: 42, email: 'first@example.com' },
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })

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
          DataTable: DataTableStub,
          Pagination: { template: '<div />' },
          ConfirmDialog: { template: '<div />' },
          GroupBadge: { template: '<span />' },
          GroupOptionItem: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('首冲赠送额度')
    expect(wrapper.text()).toContain('15.00')
    expect(wrapper.text()).not.toContain('平台额度')
    expect(wrapper.text()).not.toContain('¥15.00')
  })

  it('renders affiliate balance transfers as readonly redeem activity', async () => {
    listRedeemCodes.mockResolvedValue({
      items: [
        {
          id: -664,
          code: 'AFF-664',
          type: 'affiliate_balance',
          value: 130,
          status: 'used',
          used_by: 2,
          used_at: '2026-06-03T10:00:00Z',
          created_at: '2026-06-03T10:00:00Z',
          user: { id: 2, email: 'asd494235908@qq.com' },
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })

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
          DataTable: DataTableStub,
          Pagination: { template: '<div />' },
          ConfirmDialog: { template: '<div />' },
          GroupBadge: { template: '<span />' },
          GroupOptionItem: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('返利转余额')
    expect(wrapper.text()).toContain('¥130.00')
    expect(wrapper.text()).toContain('asd494235908@qq.com')
    expect(wrapper.text()).not.toContain('affiliate_balance')
  })

  it('prefers platform amount when rendering affiliate balance transfers', async () => {
    listRedeemCodes.mockResolvedValue({
      items: [
        {
          id: -665,
          code: 'AFF-665',
          type: 'affiliate_balance',
          value: 10,
          platform_amount: 130,
          status: 'used',
          used_by: 2,
          used_at: '2026-06-03T10:00:00Z',
          created_at: '2026-06-03T10:00:00Z',
          user: { id: 2, email: 'affiliate@example.com' },
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })

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
          DataTable: DataTableStub,
          Pagination: { template: '<div />' },
          ConfirmDialog: { template: '<div />' },
          GroupBadge: { template: '<span />' },
          GroupOptionItem: { template: '<span />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('¥130.00')
    expect(wrapper.text()).not.toContain('¥10.00')
  })
})
