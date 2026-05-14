import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'

const { changePasswordMock, sendPhoneCodeMock, showSuccessMock, showErrorMock, authStoreState } = vi.hoisted(() => ({
  changePasswordMock: vi.fn(),
  sendPhoneCodeMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  authStoreState: {
    user: {
      phone_number: '18380640817'
    }
  }
}))

vi.mock('@/api', () => ({
  userAPI: {
    changePassword: changePasswordMock,
    sendChangePasswordPhoneCode: sendPhoneCodeMock
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock
  })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  const locale = { value: 'en' }
  return {
    ...actual,
    useI18n: () => ({
      locale,
      t: (key: string) => {
        const translations: Record<string, string> = {
          'profile.changePassword': 'Change Password',
          'profile.currentPassword': 'Current Password',
          'profile.newPassword': 'New Password',
          'profile.confirmNewPassword': 'Confirm New Password',
          'profile.passwordHint': 'Password must be at least 8 characters long',
          'profile.changingPassword': 'Changing...',
          'profile.changePasswordButton': 'Change Password',
          'profile.passwordsNotMatch': 'New passwords do not match',
          'profile.passwordTooShort': 'Password must be at least 8 characters long',
          'profile.passwordChangeSuccess': 'Password changed successfully',
          'profile.passwordChangeFailed': 'Failed to change password',
          'profile.phoneBinding.codeLabel': 'SMS verification code',
          'profile.phoneBinding.currentPhoneValue': `Current bound phone: 18380640817`,
          'profile.phoneBinding.unbound': 'No phone number is bound to this account',
          'profile.phoneBinding.sendCodeSuccess': 'SMS verification code sent',
          'profile.phoneBinding.sendCodeFailed': 'Failed to send SMS verification code',
          'auth.sendingCode': 'Sending...',
          'auth.sendCode': 'Send code'
        }
        return translations[key] ?? key
      }
    })
  }
})

describe('ProfilePasswordForm', () => {
  beforeEach(() => {
    changePasswordMock.mockReset()
    sendPhoneCodeMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    authStoreState.user = {
      phone_number: '18380640817'
    }
  })

  it('shows validation failures as toast messages instead of inline errors', async () => {
    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('#old_password').setValue('old-password')
    await wrapper.get('#new_password').setValue('new-password')
    await wrapper.get('#confirm_password').setValue('different-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('New passwords do not match')
    expect(wrapper.find('.input-error-text').exists()).toBe(false)
  })

  it('shows API failures as toast messages', async () => {
    changePasswordMock.mockRejectedValue({
      response: { data: { detail: 'backend failure' } }
    })

    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('#old_password').setValue('old-password')
    await wrapper.get('#new_password').setValue('new-password')
    await wrapper.get('#confirm_password').setValue('new-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).toHaveBeenCalledWith('old-password', 'new-password', '')
    expect(showErrorMock).toHaveBeenCalledWith('backend failure')
    expect(wrapper.find('.input-error-text').exists()).toBe(false)
  })

  it('starts a 60 second cooldown after sending phone code successfully', async () => {
    sendPhoneCodeMock.mockResolvedValue({ message: 'ok', countdown: 60 })
    vi.useFakeTimers()

    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('button[type="button"]').trigger('click')
    await Promise.resolve()

    expect(sendPhoneCodeMock).toHaveBeenCalled()
    expect(showSuccessMock).toHaveBeenCalledWith('SMS verification code sent')
    expect(wrapper.get('button[type="button"]').text()).toContain('60s')

    vi.advanceTimersByTime(1000)
    await Promise.resolve()

    expect(wrapper.get('button[type="button"]').text()).toContain('59s')
    vi.useRealTimers()
  })

  it('uses server cooldown countdown when phone code is rate limited', async () => {
    sendPhoneCodeMock.mockRejectedValue({
      response: {
        data: {
          reason: 'VERIFY_CODE_TOO_FREQUENT',
          metadata: { countdown: '42' },
          detail: '请在 42 秒后重试'
        }
      }
    })
    vi.useFakeTimers()

    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('button[type="button"]').trigger('click')
    await Promise.resolve()

    expect(showErrorMock).toHaveBeenCalledWith('请在 42 秒后重试')
    expect(wrapper.get('button[type="button"]').text()).toContain('42s')
    vi.useRealTimers()
  })

  it('renders phone verification copy in English without Chinese hardcoded text', () => {
    const wrapper = mount(ProfilePasswordForm)

    expect(wrapper.text()).toContain('SMS verification code')
    expect(wrapper.text()).toContain('Current bound phone: 18380640817')
    expect(wrapper.get('button[type="button"]').text()).toBe('Send code')
    expect(wrapper.text()).not.toContain('短信验证码')
    expect(wrapper.text()).not.toContain('当前绑定手机号')
    expect(wrapper.text()).not.toContain('发送验证码')
  })

  it('shows localized unbound-phone error before sending a phone code', async () => {
    authStoreState.user = {
      phone_number: ''
    }
    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('button[type="button"]').trigger('click')

    expect(sendPhoneCodeMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('No phone number is bound to this account')
  })

  it('shows localized fallback when sending phone code fails without server detail', async () => {
    sendPhoneCodeMock.mockRejectedValue({
      response: { data: {} }
    })
    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('button[type="button"]').trigger('click')
    await Promise.resolve()

    expect(showErrorMock).toHaveBeenCalledWith('Failed to send SMS verification code')
  })
})
