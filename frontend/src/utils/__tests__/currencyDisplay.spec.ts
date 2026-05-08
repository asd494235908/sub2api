import { describe, expect, it } from 'vitest'

import { formatCurrency } from '../format'
import { formatScaled } from '../pricing'
import { formatTokenPricePerMillion } from '../usagePricing'

describe('currency display helpers', () => {
  it('defaults to CNY formatting for generic currency amounts', () => {
    const formatted = formatCurrency(12.5)

    expect(formatted).toContain('¥')
  })

  it('uses yen symbol for scaled prices', () => {
    expect(formatScaled(0.5, 1)).toBe('¥0.5')
  })

  it('uses yen symbol for token prices per million', () => {
    expect(formatTokenPricePerMillion(1, 1_000_000)).toBe('¥1.0000')
  })
})
