import { describe, expect, it } from 'vitest'

import { buildGuideSections } from '../../../../scripts/build-guide-content.mjs'

describe('buildGuideSections', () => {
  it('normalizes markdown docs into ordered guide sections', () => {
    const sections = buildGuideSections()

    expect(sections[0]?.filename).toBe('guide')
    expect(sections.some((section) => section.filename === 'Codex')).toBe(true)
    expect(sections[0]?.markdown).not.toMatch(/^#\s+/)
    expect(sections[0]?.markdown).toContain('](#doc-codex)')
    expect(sections[0]?.headings.some((heading) => heading.text.includes('接入前准备'))).toBe(true)
    expect(sections.find((section) => section.filename === 'OpenCode')?.headings.length).toBeGreaterThan(0)
  })
})
