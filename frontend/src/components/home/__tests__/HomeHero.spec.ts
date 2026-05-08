import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import HomeHero from '../HomeHero.vue'

const props = {
  badge: 'Hero badge',
  titlePrefix: 'Brand',
  titleLead: 'Build ',
  titleHighlight: 'AI',
  titleSuffix: ' Faster',
  subtitle: 'Subtitle',
  checks: ['A', 'B', 'C'],
  consolePath: '/dashboard',
  primaryLabel: 'Start',
  secondaryLabel: 'Docs',
  docUrl: 'https://example.com/docs',
  stats: [
    { value: '99.9%', label: 'SLA' },
    { value: '50K+', label: 'Developers' },
    { value: '24/7', label: 'Support' },
    { value: '<200ms', label: 'Latency' }
  ],
  headerOffset: 120
}

describe('HomeHero', () => {
  it('uses viewport-aware hero height', () => {
    const wrapper = mount(HomeHero, {
      props,
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="to"><slot /></a>'
          }
        }
      }
    })

    const hero = wrapper.get('[data-test="home-hero"]')
    expect(hero.attributes('style')).toContain('min-height: calc(100vh - 120px);')
  })

  it('renders the animated background layers used by the hero scene', () => {
    const wrapper = mount(HomeHero, {
      props,
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="to"><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.find('[data-test="hero-scan-light"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test="hero-stream"]')).toHaveLength(3)
    expect(wrapper.findAll('[data-test="hero-particle"]')).toHaveLength(8)
    expect(wrapper.find('[data-test="hero-visual-stack"]').exists()).toBe(true)
  })
})
