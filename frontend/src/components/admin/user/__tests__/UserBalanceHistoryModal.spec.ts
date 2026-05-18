import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const { getUserBalanceHistory } = vi.hoisted(() => ({
  getUserBalanceHistory: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserBalanceHistory,
    },
  },
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

const user = {
  id: 42,
  email: 'weekly@example.com',
  username: 'weekly-user',
  notes: '',
  balance: 18.5,
  concurrency: 3,
  role: 'user' as const,
  status: 'active',
  created_at: '2026-05-13T10:00:00Z',
  updated_at: '2026-05-13T10:00:00Z',
}

function mountModal() {
  return mount(UserBalanceHistoryModal, {
    props: {
      show: false,
      user,
    },
    global: {
      stubs: {
        BaseDialog: {
          props: ['show'],
          template: '<div v-if="show"><slot /></div>',
        },
        Icon: {
          props: ['name'],
          template: '<span data-test="icon">{{ name }}</span>',
        },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue', 'change'],
          template: `
            <div>
              <button
                v-for="option in options"
                :key="option.value"
                type="button"
                @click="$emit('update:modelValue', option.value); $emit('change', option.value, option)"
              >
                {{ option.label }}
              </button>
            </div>
          `,
        },
      },
    },
  })
}

describe('UserBalanceHistoryModal', () => {
  beforeEach(() => {
    getUserBalanceHistory.mockReset()
    getUserBalanceHistory.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 15,
      pages: 1,
      total_recharged: 0,
    })
  })

  it('renders weekly quota claims as balance grants', async () => {
    getUserBalanceHistory.mockResolvedValue({
      items: [
        {
          id: 1001,
          code: 'WQ-42-1770000000',
          type: 'weekly_balance',
          value: 12.5,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T10:00:00Z',
          created_at: '2026-05-14T10:00:00Z',
          group_id: null,
          validity_days: 7,
          notes: '',
        },
      ],
      total: 1,
      page: 1,
      page_size: 15,
      pages: 1,
      total_recharged: 0,
    })

    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('redeem.weeklyQuotaHistoryTitle')
    expect(wrapper.text()).toContain('+¥12.50')
    expect(wrapper.text()).toContain('redeem.systemGrant')
    expect(wrapper.text()).not.toContain('common.unknown')
  })

  it('allows filtering balance history by weekly quota type', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const weeklyFilter = wrapper
      .findAll('button')
      .find((button) => button.text() === 'redeem.weeklyQuotaHistoryTitle')
    expect(weeklyFilter).toBeTruthy()

    await weeklyFilter!.trigger('click')
    await flushPromises()

    expect(getUserBalanceHistory).toHaveBeenLastCalledWith(42, 1, 15, 'weekly_balance')
  })

  it('renders lucky wheel bonuses as system balance grants', async () => {
    getUserBalanceHistory.mockResolvedValue({
      items: [
        {
          id: -3001,
          code: 'LUCKY-702',
          type: 'lucky_wheel_bonus',
          value: 90,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T10:00:00Z',
          created_at: '2026-05-14T09:00:00Z',
          group_id: null,
          validity_days: 0,
          notes: '20-50',
        },
      ],
      total: 1,
      page: 1,
      page_size: 15,
      pages: 1,
      total_recharged: 0,
    })

    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('redeem.luckyWheelBonusTitle')
    expect(wrapper.text()).toContain('+¥90.00')
    expect(wrapper.text()).toContain('redeem.systemGrant')
  })

  it('allows filtering balance history by lucky wheel type', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const luckyWheelFilter = wrapper
      .findAll('button')
      .find((button) => button.text() === 'redeem.luckyWheelBonusTitle')
    expect(luckyWheelFilter).toBeTruthy()

    await luckyWheelFilter!.trigger('click')
    await flushPromises()

    expect(getUserBalanceHistory).toHaveBeenLastCalledWith(42, 1, 15, 'lucky_wheel_bonus')
  })

  it('keeps existing balance, concurrency, and admin adjustment display behavior', async () => {
    getUserBalanceHistory.mockResolvedValue({
      items: [
        {
          id: 2001,
          code: 'BALANCE-REDEEM',
          type: 'balance',
          value: 5,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T10:00:00Z',
          created_at: '2026-05-14T10:00:00Z',
          group_id: null,
          validity_days: 0,
          notes: '',
        },
        {
          id: 2002,
          code: 'CONCURRENCY-REDEEM',
          type: 'concurrency',
          value: 2,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T11:00:00Z',
          created_at: '2026-05-14T11:00:00Z',
          group_id: null,
          validity_days: 0,
          notes: '',
        },
        {
          id: 2003,
          code: 'ADMIN-BALANCE',
          type: 'admin_balance',
          value: -3,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T12:00:00Z',
          created_at: '2026-05-14T12:00:00Z',
          group_id: null,
          validity_days: 0,
          notes: 'manual correction',
        },
        {
          id: 2004,
          code: 'ADMIN-CONCURRENCY',
          type: 'admin_concurrency',
          value: -1,
          status: 'used',
          used_by: 42,
          used_at: '2026-05-14T13:00:00Z',
          created_at: '2026-05-14T13:00:00Z',
          group_id: null,
          validity_days: 0,
          notes: 'lower plan',
        },
      ],
      total: 4,
      page: 1,
      page_size: 15,
      pages: 1,
      total_recharged: 5,
    })

    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('redeem.balanceAddedRedeem')
    expect(wrapper.text()).toContain('+¥5.00')
    expect(wrapper.text()).toContain('redeem.concurrencyAddedRedeem')
    expect(wrapper.text()).toContain('+2')
    expect(wrapper.text()).toContain('redeem.balanceDeductedAdmin')
    expect(wrapper.text()).toContain('¥-3.00')
    expect(wrapper.text()).toContain('redeem.concurrencyReducedAdmin')
    expect(wrapper.text()).toContain('-1')
    expect(wrapper.text()).toContain('manual correction')
    expect(wrapper.text()).toContain('lower plan')
  })
})
