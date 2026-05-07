import { describe, expect, it, vi } from 'vitest'

vi.mock('@/i18n', () => ({
  i18n: {
    global: {
      t: (key: string) => key,
    },
  },
  getLocale: () => 'zh-CN',
}))

describe('formatCurrency', () => {
  it('defaults to CNY formatting', async () => {
    const { formatCurrency } = await import('@/utils/format')
    expect(formatCurrency(12.5)).toContain('¥')
  })
})
