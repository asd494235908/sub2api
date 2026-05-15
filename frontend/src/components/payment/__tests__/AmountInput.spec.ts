import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AmountInput from '@/components/payment/AmountInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

describe('AmountInput', () => {
  it('shows quick amounts without rendering the custom amount text input', async () => {
    const wrapper = mount(AmountInput, {
      props: {
        modelValue: null,
        amounts: [10, 20],
      },
    })

    expect(wrapper.text()).toContain('payment.quickAmounts')
    expect(wrapper.text()).not.toContain('payment.customAmount')
    expect(wrapper.find('input[type="text"]').exists()).toBe(false)

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[10]])
  })
})
