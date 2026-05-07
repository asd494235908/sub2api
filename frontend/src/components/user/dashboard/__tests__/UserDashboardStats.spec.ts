import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('UserDashboardStats', () => {
  it('renders balance and cost values with the yuan symbol', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        isSimple: false,
        balance: 12.5,
        stats: {
          total_api_keys: 1,
          active_api_keys: 1,
          today_requests: 2,
          total_requests: 3,
          today_actual_cost: 1.2345,
          today_cost: 2.3456,
          total_actual_cost: 3.4567,
          total_cost: 4.5678,
          today_tokens: 10,
          today_input_tokens: 4,
          today_output_tokens: 6,
          total_tokens: 20,
          total_input_tokens: 8,
          total_output_tokens: 12,
          rpm: 0,
          tpm: 0,
          average_duration_ms: 100,
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('¥12.50')
    expect(wrapper.text()).toContain('¥1.2345')
    expect(wrapper.text()).toContain('/ ¥2.3456')
    expect(wrapper.text()).toContain('¥3.4567')
    expect(wrapper.text()).toContain('/ ¥4.5678')
  })
})
