import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import GuideView from '../GuideView.vue'

function createWrapper() {
  return mount(GuideView, {
    global: {
      stubs: {
        Icon: {
          template: '<span data-test="icon" />'
        },
        LocaleSwitcher: {
          template: '<div data-test="locale-switcher">LocaleSwitcher</div>'
        },
        'router-link': {
          props: ['to'],
          template: '<a :href="typeof to === \'string\' ? to : to?.path"><slot /></a>'
        }
      }
    }
  })
}

describe('GuideView', () => {
  beforeEach(() => {
    document.documentElement.classList.remove('dark')
    localStorage.removeItem('theme')
  })

  it('renders the guide landing page with aggregated document navigation and content', () => {
    const wrapper = createWrapper()

    expect(wrapper.text()).toContain('GPAPI 接入总入口')
    expect(wrapper.text()).toContain('文档目录')
    expect(wrapper.text()).toContain('快速接入模板')
    expect(wrapper.text()).toContain('Codex 接入中转站 API 教程')
    expect(wrapper.text()).toContain('Chatbox 接入中转站 API 教程')
    expect(wrapper.text()).toContain('OpenCode 接入 OpenAI 兼容 API 教程')
    expect(wrapper.find('[data-test="locale-switcher"]').exists()).toBe(true)
    expect(wrapper.find('a[href="#doc-codex"]').exists()).toBe(true)
    expect(wrapper.find('a[href="#doc-chatbox"]').exists()).toBe(true)
    expect(wrapper.find('a[href="#doc-guide"]').exists()).toBe(true)
  })

  it('syncs with external theme changes on the document element', async () => {
    const wrapper = createWrapper()

    const button = wrapper.get('button[title="切换深色模式"]')
    expect(button.exists()).toBe(true)
    expect(wrapper.get('.markdown-body').classes()).toContain('markdown-body-light')

    document.documentElement.classList.add('dark')
    await nextTick()
    await nextTick()

    expect(wrapper.get('button[title="切换浅色模式"]').exists()).toBe(true)
    expect(wrapper.get('.markdown-body').classes()).toContain('markdown-body-dark')
  })

  it('renders markdown code blocks without syntax highlight classes', () => {
    const wrapper = createWrapper()

    const codeBlock = wrapper.find('.markdown-body pre code')

    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.attributes('class')).toBeUndefined()
  })
})
