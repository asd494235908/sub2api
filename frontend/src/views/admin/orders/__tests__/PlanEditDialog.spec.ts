import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlanEditDialog from '../PlanEditDialog.vue'

const { createPlan, updatePlan } = vi.hoisted(() => ({
  createPlan: vi.fn(),
  updatePlan: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    createPlan,
    updatePlan,
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

const groups = [
  {
    id: 7,
    name: 'Pro Group',
    platform: 'openai',
    rate_multiplier: 1,
    subscription_type: 'subscription',
  },
]

function expectedDateTimeLocal(value: string) {
  const date = new Date(value)
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function mountDialog(plan: Record<string, unknown> | null = null) {
  return mount(PlanEditDialog, {
    props: {
      show: true,
      plan: plan as never,
      groups: groups as never,
    },
    global: {
      stubs: {
        BaseDialog: {
          props: ['show', 'title'],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>',
        },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue'],
          template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', Number($event.target.value) || $event.target.value)"><option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option></select>',
        },
        Icon: true,
        GroupBadge: true,
      },
    },
  })
}

describe('PlanEditDialog', () => {
  beforeEach(() => {
    createPlan.mockReset()
    updatePlan.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('submits sale window and daily purchase limit fields', async () => {
    createPlan.mockResolvedValue({ data: { id: 1 } })

    const wrapper = mountDialog()
    await wrapper.find('input[type="text"]').setValue('Pro Plan')
    await wrapper.find('select').setValue('7')
    await wrapper.find('textarea').setValue('Description')
    const numberInputs = wrapper.findAll('input[type="number"]')
    await numberInputs[0].setValue('19.9')
    await numberInputs[2].setValue('30')
    await numberInputs[4].setValue('8')
    const dateInputs = wrapper.findAll('input[type="datetime-local"]')
    await dateInputs[0].setValue('2026-05-27T09:30')
    await dateInputs[1].setValue('2026-05-28T18:45')
    const timeInputs = wrapper.findAll('input[type="time"]')
    await timeInputs[0].setValue('09:00')
    await timeInputs[1].setValue('18:30')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createPlan).toHaveBeenCalledTimes(1)
    expect(createPlan.mock.calls[0][0]).toMatchObject({
      name: 'Pro Plan',
      group_id: 7,
      description: 'Description',
      daily_purchase_limit: 8,
      daily_sale_starts_at: '09:00',
      daily_sale_ends_at: '18:30',
    })
    expect(new Date(createPlan.mock.calls[0][0].sale_starts_at).toISOString()).toBe(new Date('2026-05-27T09:30').toISOString())
    expect(new Date(createPlan.mock.calls[0][0].sale_ends_at).toISOString()).toBe(new Date('2026-05-28T18:45').toISOString())
  })

  it('loads existing sale window values and keeps zero as unlimited', async () => {
    updatePlan.mockResolvedValue({ data: { id: 3 } })

    const wrapper = mountDialog({
      id: 3,
      name: 'Existing',
      group_id: 7,
      description: 'Existing description',
      price: 29,
      original_price: 39,
      validity_days: 30,
      validity_unit: 'days',
      features: [],
      sort_order: 1,
      for_sale: true,
      sale_starts_at: '2026-05-27T01:30:00Z',
      sale_ends_at: '2026-05-28T10:45:00Z',
      daily_purchase_limit: 0,
      daily_sale_starts_at: '22:00',
      daily_sale_ends_at: '02:00',
    })

    const dateInputs = wrapper.findAll('input[type="datetime-local"]')
    expect(dateInputs[0].element.value).toBe(expectedDateTimeLocal('2026-05-27T01:30:00Z'))
    expect(dateInputs[1].element.value).toBe(expectedDateTimeLocal('2026-05-28T10:45:00Z'))
    const timeInputs = wrapper.findAll('input[type="time"]')
    expect(timeInputs[0].element.value).toBe('22:00')
    expect(timeInputs[1].element.value).toBe('02:00')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updatePlan).toHaveBeenCalledTimes(1)
    expect(updatePlan.mock.calls[0][1]).toMatchObject({
      daily_purchase_limit: 0,
      daily_sale_starts_at: '22:00',
      daily_sale_ends_at: '02:00',
    })
  })

  it('sends null when daily sale window inputs are cleared', async () => {
    updatePlan.mockResolvedValue({ data: { id: 4 } })

    const wrapper = mountDialog({
      id: 4,
      name: 'Clear Daily',
      group_id: 7,
      description: 'Existing description',
      price: 29,
      original_price: 39,
      validity_days: 30,
      validity_unit: 'days',
      features: [],
      sort_order: 1,
      for_sale: true,
      sale_starts_at: null,
      sale_ends_at: null,
      daily_purchase_limit: 0,
      daily_sale_starts_at: '09:00',
      daily_sale_ends_at: '18:00',
    })

    const timeInputs = wrapper.findAll('input[type="time"]')
    await timeInputs[0].setValue('')
    await timeInputs[1].setValue('')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updatePlan.mock.calls[0][1]).toMatchObject({
      daily_sale_starts_at: null,
      daily_sale_ends_at: null,
    })
  })
})
