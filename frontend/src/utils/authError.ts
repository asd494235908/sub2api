interface APIErrorLike {
  message?: string
  response?: {
    data?: {
      code?: string | number
      detail?: string
      message?: string
      reason?: string
      metadata?: Record<string, string>
    }
  }
}

function extractErrorMessage(error: unknown): string {
  const err = (error || {}) as APIErrorLike
  return err.response?.data?.detail || err.response?.data?.message || err.message || ''
}

export function buildAuthErrorMessage(
  error: unknown,
  options: {
    fallback: string
    phoneExistsMessage?: string
  }
): string {
  const { fallback, phoneExistsMessage } = options
  const err = (error || {}) as APIErrorLike
  const responseData = err.response?.data
  const countdownRaw = responseData?.metadata?.countdown
  const countdown = countdownRaw ? Number.parseInt(countdownRaw, 10) : NaN

  if (
    responseData?.reason === 'VERIFY_CODE_TOO_FREQUENT' &&
    Number.isFinite(countdown) &&
    countdown > 0
  ) {
    return `请在 ${countdown} 秒后重试`
  }

  if (responseData?.reason === 'PHONE_EXISTS' && phoneExistsMessage) {
    return phoneExistsMessage
  }

  const message = extractErrorMessage(error)
  return message || fallback
}
