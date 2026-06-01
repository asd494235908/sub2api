import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminAffiliateRecordsTable from '../affiliates/AdminAffiliateRecordsTable.vue'

const { listRebateRecords } = vi.hoisted(() => ({
  listRebateRecords: vi.fn(),
}))

vi.mock('@/api/admin/affiliates', () => {
  const api = {
    listRebateRecords,
    getUserOverview: vi.fn(),
  }
  return {
    affiliatesAPI: api,
    default: api,
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'error',
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => `FMT:${value}`,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.affiliates.records.order': '订单',
    'admin.affiliates.records.inviter': '邀请人',
    'admin.affiliates.records.invitee': '被邀请人',
    'admin.affiliates.records.orderAmount': '订单金额',
    'admin.affiliates.records.payAmount': '支付金额',
    'admin.affiliates.records.rebateBaseAmount': '返利基数',
    'admin.affiliates.records.rebateAmount': '返利金额',
    'admin.affiliates.records.paymentType': '支付方式',
    'admin.affiliates.records.orderStatus': '订单状态',
    'admin.affiliates.records.rebatedAt': '返利时间',
    'common.refresh': '刷新',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, fallback?: string) => messages[key] ?? fallback ?? key,
    }),
  }
})

const DataTableStub = {
  props: ['columns', 'data'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map((column) => column.label).join('|') }}</div>
      <div v-for="row in data" :key="row.order_id">
        <slot name="cell-rebate_base_amount" :row="row" />
      </div>
    </div>
  `,
}

describe('AdminAffiliateRecordsTable', () => {
  beforeEach(() => {
    listRebateRecords.mockReset()
    listRebateRecords.mockResolvedValue({
      items: [
        {
          order_id: 1001,
          out_trade_no: 'sub2_subscription_1001',
          inviter_id: 11,
          inviter_email: 'owner@example.com',
          inviter_username: 'owner',
          invitee_id: 22,
          invitee_email: 'friend@example.com',
          invitee_username: 'friend',
          order_amount: 1560,
          pay_amount: 1560,
          rebate_base_amount: 3000,
          rebate_amount: 300,
          payment_type: 'alipay',
          order_status: 'completed',
          created_at: '2026-06-01T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
  })

  it('shows rebate base amount for rebate records', async () => {
    const wrapper = mount(AdminAffiliateRecordsTable, {
      props: { type: 'rebates' },
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: true,
          Icon: true,
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(listRebateRecords).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20 }))
    expect(wrapper.find('[data-test="columns"]').text()).toContain('返利基数')
    expect(wrapper.text()).toContain('¥3000.00')
  })
})
