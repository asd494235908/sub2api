import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import LuckyWheelPrizeEditor from '../LuckyWheelPrizeEditor.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('LuckyWheelPrizeEditor', () => {
  it('emits reorder when draggable order changes', async () => {
    const wrapper = mount(LuckyWheelPrizeEditor, {
      props: {
        prizes: [
          { id: 'p2', name: '二等奖', reward_amount: 8.88, enabled: true },
          { id: 'p1', name: '一等奖', reward_amount: 18.88, enabled: true },
        ],
      },
      global: {
        stubs: {
          VueDraggable: {
            props: ['modelValue'],
            emits: ['update:modelValue', 'end'],
            template: '<div><slot /></div>',
          },
        },
      },
    })

    wrapper.vm.$emit('reorder', [
      { id: 'p1', name: '一等奖', reward_amount: 18.88, enabled: true },
      { id: 'p2', name: '二等奖', reward_amount: 8.88, enabled: true },
    ])

    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('reorder')).toBeTruthy()
  })
})
