import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import LuckyWheelTierEditor from '../LuckyWheelTierEditor.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('LuckyWheelTierEditor', () => {
  it('emits weight updates for a prize input', async () => {
    const wrapper = mount(LuckyWheelTierEditor, {
      props: {
        prizes: [
          { id: 'p1', name: '一等奖', reward_amount: 18.88, enabled: true },
        ],
        tiers: [
          { id: 'tier_0', name: '基础档', min_amount: 0, max_amount: 99, prize_weights: { p1: 1 } },
        ],
        warnings: [[]],
      },
    })

    const inputs = wrapper.findAll('input[type="number"][step="0.01"]')
    const weightInput = inputs[2]
    await weightInput.setValue('3.5')

    const emitted = wrapper.emitted('update-weight')
    expect(emitted).toBeTruthy()
    expect(emitted?.[0]).toEqual([0, 'p1', '3.5'])
  })

  it('renders warning messages for a tier', () => {
    const wrapper = mount(LuckyWheelTierEditor, {
      props: {
        prizes: [
          { id: 'p1', name: '一等奖', reward_amount: 18.88, enabled: true },
        ],
        tiers: [
          { id: 'tier_0', name: '基础档', min_amount: 0, max_amount: 99, prize_weights: { p1: 0 } },
        ],
        warnings: [['至少配置一个大于 0 的奖项权重', '最大金额不能小于最小金额']],
      },
    })

    expect(wrapper.text()).toContain('至少配置一个大于 0 的奖项权重')
    expect(wrapper.text()).toContain('最大金额不能小于最小金额')
  })

  it('emits empty string when max_amount is cleared', async () => {
    const wrapper = mount(LuckyWheelTierEditor, {
      props: {
        prizes: [
          { id: 'p1', name: '一等奖', reward_amount: 18.88, enabled: true },
        ],
        tiers: [
          { id: 'tier_0', name: '基础档', min_amount: 0, max_amount: 99, prize_weights: { p1: 1 } },
        ],
        warnings: [[]],
      },
    })

    const maxAmountInput = wrapper.find('input[placeholder="最大金额，留空表示无上限"]')
    await maxAmountInput.setValue('')

    const emitted = wrapper.emitted('update-tier')
    expect(emitted).toBeTruthy()
    expect(emitted?.at(-1)).toEqual([0, 'max_amount', ''])
  })
})
