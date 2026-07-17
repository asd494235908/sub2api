import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AmountInput from '@/components/payment/AmountInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

describe('AmountInput', () => {
  it('shows quick amounts and allows entering a custom amount', async () => {
    const wrapper = mount(AmountInput, {
      props: {
        modelValue: null,
        amounts: [10, 20],
      },
    })

    expect(wrapper.text()).toContain('payment.quickAmounts')
    expect(wrapper.text()).toContain('payment.customAmount')
    expect(wrapper.find('input[type="text"]').exists()).toBe(true)

    await wrapper.get('button').trigger('click')
    await wrapper.get('input[type="text"]').setValue('88.88')

    expect(wrapper.emitted('update:modelValue')).toEqual([[10], [88.88]])
  })
})
