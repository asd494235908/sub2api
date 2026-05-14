import { describe, expect, it } from 'vitest'
import { buildAuthErrorMessage } from '@/utils/authError'

describe('buildAuthErrorMessage', () => {
  it('prefers response detail message when available', () => {
    const message = buildAuthErrorMessage(
      {
        response: {
          data: {
            detail: 'detailed message',
            message: 'plain message'
          }
        },
      },
      { fallback: 'fallback' }
    )
    expect(message).toBe('detailed message')
  })

  it('falls back to response message when detail is unavailable', () => {
    const message = buildAuthErrorMessage(
      {
        response: {
          data: {
            message: 'plain message'
          }
        },
      },
      { fallback: 'fallback' }
    )
    expect(message).toBe('plain message')
  })

  it('falls back to error.message when response payload is unavailable', () => {
    const message = buildAuthErrorMessage(
      {
        message: 'error message'
      },
      { fallback: 'fallback' }
    )
    expect(message).toBe('error message')
  })

  it('uses fallback when no message can be extracted', () => {
    expect(buildAuthErrorMessage({}, { fallback: 'fallback' })).toBe('fallback')
  })

  it('formats cooldown message from reason metadata', () => {
    const message = buildAuthErrorMessage(
      {
        response: {
          data: {
            reason: 'VERIFY_CODE_TOO_FREQUENT',
            metadata: {
              countdown: '42',
            },
          },
        },
      },
      { fallback: 'fallback' }
    )

    expect(message).toBe('请在 42 秒后重试')
  })

  it('maps phone exists to register send-code message', () => {
    const message = buildAuthErrorMessage(
      {
        response: {
          data: {
            reason: 'PHONE_EXISTS',
            detail: 'phone number already exists',
          },
        },
      },
      { fallback: '短信验证码发送失败', phoneExistsMessage: '发送验证码失败，该手机号已注册' }
    )

    expect(message).toBe('发送验证码失败，该手机号已注册')
  })
})
