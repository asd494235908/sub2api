import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

const mountOpenAIUseKeyModal = (apiKey = 'sk-test-secret') => mount(UseKeyModal, {
  props: {
    show: true,
    apiKey,
    baseUrl: 'https://example.com/v1',
    platform: 'openai'
  },
  global: {
    stubs: {
      BaseDialog: {
        template: '<div><slot /><slot name="footer" /></div>'
      },
      Icon: {
        template: '<span />'
      }
    }
  }
})

describe('UseKeyModal', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('keeps the original Codex CLI config files as the default tab', () => {
    const wrapper = mountOpenAIUseKeyModal()

    expect(wrapper.text()).toContain('keys.useKeyModal.cliTabs.codexCli')
    expect(wrapper.text()).toContain('keys.useKeyModal.cliTabs.codexOneClick')

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    const config = codeBlock.text()

    expect(config).toContain('model_provider = "OpenAI"')
    expect(config).toContain('model = "gpt-5.5"')
    expect(config).toContain('review_model = "gpt-5.5"')
    expect(config).toContain('base_url = "https://example.com/v1"')
    expect(config).toContain('[features]\ngoals = true')
    expect(config).not.toContain('model = "gpt-5.4"')
    expect(config).not.toContain('model_context_window')
    expect(config).not.toContain('model_auto_compact_token_limit')
    expect(wrapper.text()).toContain('~/.codex/config.toml')
    expect(wrapper.text()).toContain('~/.codex/auth.json')
    expect(wrapper.text()).toContain('sk-test-secret')
    expect(config).not.toContain('read -rsp')
    expect(config).not.toContain('curl -fsS')
  })

  it('appends /v1 to Codex base_url when the configured endpoint omits it', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test-secret',
        baseUrl: 'https://example.com',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.text()).toContain('base_url = "https://example.com/v1"')
  })

  it('renders GPT-5.5 and goals feature in OpenAI Codex WebSocket config', async () => {
    const wrapper = mountOpenAIUseKeyModal('sk-test')

    const wsTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )

    expect(wsTab).toBeDefined()
    await wsTab!.trigger('click')
    await nextTick()

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('supports_websockets = true'))

    expect(configToml).toBeDefined()
    expect(configToml).toContain('base_url = "https://example.com/v1"')
    expect(configToml).toContain('model = "gpt-5.5"')
    expect(configToml).toContain('review_model = "gpt-5.5"')
    expect(configToml).not.toContain('model = "gpt-5.4"')
    expect(configToml).not.toContain('model_context_window')
    expect(configToml).not.toContain('model_auto_compact_token_limit')
    expect(configToml).toContain('supports_websockets = true\nrequires_openai_auth = true')
    expect(configToml).toContain('[features]\nresponses_websockets_v2 = true\ngoals = true')
  })

  it('shows a macOS one-click Codex CLI/App setup script without embedding the API key', async () => {
    const wrapper = mountOpenAIUseKeyModal()

    const oneClickTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexOneClick')
    )

    expect(oneClickTab).toBeDefined()
    await oneClickTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    const script = codeBlock.text()

    expect(script).toContain('https://token.gepinkeji.com/v1')
    expect(script).toContain('mkdir -p "$HOME/.codex"')
    expect(script).toContain('read -rsp')
    expect(script).toContain('.codex/config.toml')
    expect(script).toContain('.codex/auth.json')
    expect(script).toContain('.bak')
    expect(script).toContain('/v1/models')
    expect(script).toContain('curl -fsS')
    expect(script).not.toContain('sk-test-secret')
  })

  it('lets users download the one-click setup script', async () => {
    const createObjectURL = vi.fn(() => 'blob:codex-script')
    const revokeObjectURL = vi.fn()
    vi.stubGlobal('URL', { createObjectURL, revokeObjectURL })
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    const wrapper = mountOpenAIUseKeyModal()

    const oneClickTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexOneClick')
    )
    expect(oneClickTab).toBeDefined()
    await oneClickTab!.trigger('click')
    await nextTick()

    const downloadButton = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.download')
    )
    expect(downloadButton).toBeDefined()
    await downloadButton!.trigger('click')

    expect(createObjectURL).toHaveBeenCalledTimes(1)
    expect(click).toHaveBeenCalledTimes(1)
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:codex-script')
  })

  it('shows a Windows PowerShell Codex CLI/App setup script', async () => {
    const wrapper = mountOpenAIUseKeyModal()

    const oneClickTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexOneClick')
    )
    expect(oneClickTab).toBeDefined()
    await oneClickTab!.trigger('click')
    await nextTick()

    const windowsTab = wrapper.findAll('button').find((button) => button.text().includes('Windows'))
    expect(windowsTab).toBeDefined()
    await windowsTab!.trigger('click')
    await nextTick()

    const script = wrapper.find('pre code').text()
    expect(script).toContain('$env:USERPROFILE')
    expect(script).toContain('.codex\\config.toml')
    expect(script).toContain('.codex\\auth.json')
    expect(script).toContain('Read-Host -AsSecureString')
    expect(script).toContain('.bak')
    expect(script).toContain('Invoke-RestMethod')
    expect(script).toContain('/v1/models')
    expect(script).not.toContain('sk-test-secret')
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mountOpenAIUseKeyModal('sk-test')

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })
})
