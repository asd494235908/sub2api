import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import HomeShowcase from '../HomeShowcase.vue'

const baseProps = {
  kicker: 'Kicker',
  title: 'Title',
  subtitle: 'Subtitle',
  sectionOffsetStyle: {},
  stats: [
    { icon: 'users' as const, value: '50K+', label: 'Developers' }
  ],
  quotes: [
    { quote: '示例评价', author: 'Alice', role: 'Engineer' }
  ]
}

const global = {
  stubs: {
    Icon: {
      props: ['name'],
      template: '<span :data-icon="name" />'
    }
  }
}

describe('HomeShowcase', () => {
  it('renders Chinese quote marks when provided by the locale', () => {
    const wrapper = mount(HomeShowcase, {
      props: {
        ...baseProps,
        quoteOpen: '“',
        quoteClose: '”'
      },
      global
    })

    expect(wrapper.text()).toContain('“示例评价”')
  })

  it('renders English quote marks when provided by the locale', () => {
    const wrapper = mount(HomeShowcase, {
      props: {
        ...baseProps,
        quotes: [
          { quote: 'Example quote', author: 'Alice', role: 'Engineer' }
        ],
        quoteOpen: '"',
        quoteClose: '"'
      },
      global
    })

    expect(wrapper.text()).toContain('"Example quote"')
  })
})
