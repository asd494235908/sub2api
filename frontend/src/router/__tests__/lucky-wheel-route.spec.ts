import { describe, expect, it } from 'vitest'

interface RouteMeta {
  requiresAuth?: boolean
  requiresPayment?: boolean
}

function luckyWheelVisible(settings: { payment_enabled?: boolean; lucky_wheel_enabled?: boolean }, meta: RouteMeta): boolean {
  if (meta.requiresPayment && settings.payment_enabled !== true) {
    return false
  }
  return settings.lucky_wheel_enabled === true
}

describe('lucky wheel route visibility', () => {
  it('hides the user lucky wheel route when the feature flag is disabled', () => {
    expect(luckyWheelVisible({ payment_enabled: true, lucky_wheel_enabled: false }, { requiresAuth: true, requiresPayment: true })).toBe(false)
  })

  it('hides the user lucky wheel route when payment is disabled', () => {
    expect(luckyWheelVisible({ payment_enabled: false, lucky_wheel_enabled: true }, { requiresAuth: true, requiresPayment: true })).toBe(false)
  })

  it('shows the user lucky wheel route only when both payment and lucky wheel are enabled', () => {
    expect(luckyWheelVisible({ payment_enabled: true, lucky_wheel_enabled: true }, { requiresAuth: true, requiresPayment: true })).toBe(true)
  })
})
