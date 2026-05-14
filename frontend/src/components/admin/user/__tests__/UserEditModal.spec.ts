import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import UserEditModal from '../UserEditModal.vue'

const {
  updateUserMock,
  updateUserAttributeValuesMock,
  showSuccessMock,
  showErrorMock,
} = vi.hoisted(() => ({
  updateUserMock: vi.fn(),
  updateUserAttributeValuesMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      update: (...args: any[]) => updateUserMock(...args),
    },
    userAttributes: {
      updateUserAttributeValues: (...args: any[]) => updateUserAttributeValuesMock(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: (...args: any[]) => showSuccessMock(...args),
    showError: (...args: any[]) => showErrorMock(...args),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const adminUser = (): AdminUser => ({
  id: 7,
  email: 'member@example.com',
  phone_number: '+8613800138000',
  username: 'member',
  role: 'user',
  balance: 10,
  concurrency: 3,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-05-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
  notes: 'hello',
  rpm_limit: 0,
})

describe('UserEditModal', () => {
  beforeEach(() => {
    updateUserMock.mockReset()
    updateUserAttributeValuesMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    updateUserMock.mockResolvedValue({})
    updateUserAttributeValuesMock.mockResolvedValue({})
  })

  it('prefills phone number when editing a user', async () => {
    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: adminUser(),
      },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          UserAttributeForm: { template: '<div />' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    const phoneInput = wrapper.find('input[type="tel"]')
    expect(phoneInput.exists()).toBe(true)
    expect((phoneInput.element as HTMLInputElement).value).toBe('+8613800138000')
  })

  it('submits phone_number when updating a user', async () => {
    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: adminUser(),
      },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          UserAttributeForm: { template: '<div />' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    const phoneInput = wrapper.find('input[type="tel"]')
    await phoneInput.setValue('13800138001')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateUserMock).toHaveBeenCalledWith(7, expect.objectContaining({
      phone_number: '13800138001',
    }))
  })
})
