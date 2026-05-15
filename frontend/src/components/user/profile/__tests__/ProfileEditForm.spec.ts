import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { User } from '@/types'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'

const {
  updateProfileMock,
  sendPhoneBindingCodeMock,
  bindPhoneNumberMock,
  showSuccessMock,
  showErrorMock,
  authStoreState,
} = vi.hoisted(() => ({
  updateProfileMock: vi.fn(),
  sendPhoneBindingCodeMock: vi.fn(),
  bindPhoneNumberMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  authStoreState: {
    user: null as User | null,
  },
}))

vi.mock('@/api', () => ({
  userAPI: {
    updateProfile: updateProfileMock,
    sendPhoneBindingCode: sendPhoneBindingCodeMock,
    bindPhoneNumber: bindPhoneNumberMock,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => {
        const translations: Record<string, string> = {
          'profile.editProfile': 'Edit Profile',
          'profile.username': 'Username',
          'profile.enterUsername': 'Enter username',
          'profile.phoneBinding.title': 'Bind Phone',
          'profile.phoneBinding.currentPhone': 'Current phone',
          'profile.phoneBinding.unbound': 'No phone bound',
          'profile.phoneBinding.phoneLabel': 'Phone',
          'profile.phoneBinding.phonePlaceholder': 'Enter phone number',
          'profile.phoneBinding.codeLabel': 'SMS Code',
          'profile.phoneBinding.codePlaceholder': 'Enter verification code',
          'profile.phoneBinding.sendCodeSuccess': 'SMS verification code sent',
          'profile.phoneBinding.bindSuccess': 'Phone bound successfully',
          'profile.phoneBinding.rebindSuccess': 'Phone updated successfully',
          'profile.phoneBinding.phoneRequired': 'Phone number is required',
          'profile.phoneBinding.codeRequired': 'SMS verification code is required',
          'profile.updateProfile': 'Update Profile',
          'profile.updating': 'Updating...',
          'profile.updateSuccess': 'Profile updated',
          'profile.usernameRequired': 'Username is required',
          'profile.updateFailed': 'Failed to update profile',
          'auth.sendCode': 'Send code',
          'auth.sendingCode': 'Sending...',
        }
        return translations[key] ?? key
      },
    }),
  }
})

function createUser(overrides: Partial<User> = {}): User {
  return {
    id: 5,
    username: 'alice',
    email: 'alice@example.com',
    phone_number: '',
    role: 'user',
    balance: 10,
    concurrency: 2,
    status: 'active',
    allowed_groups: null,
    balance_notify_enabled: true,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-04-20T00:00:00Z',
    updated_at: '2026-04-20T00:00:00Z',
    ...overrides,
  }
}

describe('ProfileEditForm', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    updateProfileMock.mockReset()
    sendPhoneBindingCodeMock.mockReset()
    bindPhoneNumberMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    authStoreState.user = createUser()
    updateProfileMock.mockResolvedValue(authStoreState.user)
    sendPhoneBindingCodeMock.mockResolvedValue({ message: 'ok', countdown: 60 })
    bindPhoneNumberMock.mockResolvedValue(createUser({ phone_number: '+8613800138000' }))
  })

  it('shows bind phone fields for unbound users', async () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: true,
      },
    })

    await flushPromises()

    expect(wrapper.find('#phone_number').exists()).toBe(true)
    expect(wrapper.find('#phone_verify_code').exists()).toBe(true)
  })

  it('renders phone binding copy in English', async () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: true,
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('SMS Code')
    expect(wrapper.get('[data-testid="profile-phone-send-code"]').text()).toBe('Send code')
    expect(wrapper.text()).not.toContain('发送验证码')
    expect(wrapper.text()).not.toContain('短信验证码')
  })

  it('prefills existing phone number for rebind', async () => {
    authStoreState.user = createUser({ phone_number: '+8613800138000' })
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: true,
      },
    })

    await flushPromises()

    expect((wrapper.get('#phone_number').element as HTMLInputElement).value).toBe('+8613800138000')
  })

  it('starts a 60 second cooldown after sending phone binding code successfully', async () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: true,
      },
    })

    await flushPromises()
    await wrapper.get('#phone_number').setValue('13800138000')
    await wrapper.get('[data-testid="profile-phone-send-code"]').trigger('click')
    await flushPromises()

    expect(sendPhoneBindingCodeMock).toHaveBeenCalledWith('13800138000')
    expect(showSuccessMock).toHaveBeenCalledWith('SMS verification code sent')
    expect(wrapper.get('[data-testid="profile-phone-send-code"]').text()).toContain('60s')
  })

  it('submits phone_number and phone_verify_code when binding phone', async () => {
    const updated = createUser({ phone_number: '+8613800138000' })
    bindPhoneNumberMock.mockResolvedValue(updated)

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: true,
      },
    })

    await flushPromises()
    await wrapper.get('#phone_number').setValue('13800138000')
    await wrapper.get('#phone_verify_code').setValue('123456')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(bindPhoneNumberMock).toHaveBeenCalledWith({
      phone_number: '13800138000',
      phone_verify_code: '123456',
    })
    expect(authStoreState.user?.phone_number).toBe('+8613800138000')
  })

  it('hides phone binding controls and only updates username when phone verification is disabled', async () => {
    const updated = createUser({ username: 'alice-new' })
    updateProfileMock.mockResolvedValue(updated)

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        embedded: true,
        phoneVerifyEnabled: false,
      },
    })

    await flushPromises()

    expect(wrapper.find('#phone_number').exists()).toBe(false)
    expect(wrapper.find('#phone_verify_code').exists()).toBe(false)
    expect(wrapper.find('[data-testid="profile-phone-send-code"]').exists()).toBe(false)

    await wrapper.get('#username').setValue('alice-new')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateProfileMock).toHaveBeenCalledWith({ username: 'alice-new' })
    expect(sendPhoneBindingCodeMock).not.toHaveBeenCalled()
    expect(bindPhoneNumberMock).not.toHaveBeenCalled()
    expect(authStoreState.user?.username).toBe('alice-new')
  })
})
