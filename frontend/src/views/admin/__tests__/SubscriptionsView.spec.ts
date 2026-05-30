import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SubscriptionsView from '../SubscriptionsView.vue'

const { listSubscriptions, listGroups, searchUsers } = vi.hoisted(() => ({
  listSubscriptions: vi.fn(),
  listGroups: vi.fn(),
  searchUsers: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    subscriptions: {
      list: listSubscriptions,
    },
    groups: {
      getAll: listGroups,
    },
    usage: {
      searchUsers,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
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

vi.mock('@/composables/usePersistedPageSize', () => ({
  getPersistedPageSize: () => 20,
}))

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
}
const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-usage" :row="row" />
      </div>
    </div>
  `,
}

describe('admin SubscriptionsView', () => {
  beforeEach(() => {
    listSubscriptions.mockReset()
    listGroups.mockReset()
    searchUsers.mockReset()

    listSubscriptions.mockResolvedValue({
      items: [
        {
          id: 1,
          group_id: 10,
          user_id: 2,
          status: 'active',
          starts_at: '2026-01-01T00:00:00Z',
          expires_at: '2027-01-01T00:00:00Z',
          daily_usage_usd: 0,
          weekly_usage_usd: 0,
          monthly_usage_usd: 0,
          total_limit_usd: 3120,
          total_usage_usd: 260,
          daily_window_start: null,
          weekly_window_start: null,
          monthly_window_start: null,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          group: {
            id: 10,
            name: 'Early Bird',
            description: null,
            platform: 'openai',
            rate_multiplier: 1,
            is_exclusive: true,
            status: 'active',
            subscription_type: 'subscription',
            daily_limit_usd: null,
            weekly_limit_usd: null,
            monthly_limit_usd: null,
            subscription_total_limit_usd: 1560,
            allow_image_generation: false,
            image_rate_independent: false,
            image_rate_multiplier: 1,
            image_price_1k: null,
            image_price_2k: null,
            image_price_4k: null,
            claude_code_only: false,
            fallback_group_id: null,
            fallback_group_id_on_invalid_request: null,
            require_oauth_only: false,
            require_privacy_set: false,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listGroups.mockResolvedValue([])
    searchUsers.mockResolvedValue([])
  })

  it('shows total quota instead of unlimited for total-limit-only plans', async () => {
    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: true,
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('payment.planCard.cycleTotalShort')
    expect(wrapper.text()).toContain('¥260.00')
    expect(wrapper.text()).toContain('¥3120.00')
    expect(wrapper.text()).not.toContain('¥1560.00')
    expect(wrapper.text()).not.toContain('admin.subscriptions.unlimited')
  })
})
