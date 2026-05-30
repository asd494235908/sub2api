import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'
import { adminAPI } from '@/api/admin'

const mocks = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
  importData: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData: mocks.importData
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (!params) return key
      return `${key}:${JSON.stringify(params)}`
    }
  })
}))

describe('ImportDataModal', () => {
  beforeEach(() => {
    mocks.showError.mockReset()
    mocks.showSuccess.mockReset()
    mocks.importData.mockReset()
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    await wrapper.find('form').trigger('submit')
    expect(mocks.showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(mocks.showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('多个文件中一个 JSON 解析失败时继续导入其他文件并提示失败文件名', async () => {
    mocks.importData.mockResolvedValueOnce({
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      account_created: 1,
      account_failed: 0
    })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const badFile = new File(['invalid json'], 'bad.json', { type: 'application/json' })
    Object.defineProperty(badFile, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    const goodFile = new File(['{"exported_at":"2026-01-01T00:00:00Z","proxies":[],"accounts":[]}'], 'good.json', {
      type: 'application/json'
    })
    Object.defineProperty(goodFile, 'text', {
      value: () => Promise.resolve('{"exported_at":"2026-01-01T00:00:00Z","proxies":[],"accounts":[]}')
    })
    Object.defineProperty(input.element, 'files', {
      value: [badFile, goodFile]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(adminAPI.accounts.importData).toHaveBeenCalledTimes(1)
    expect(adminAPI.accounts.importData).toHaveBeenCalledWith({
      data: {
        exported_at: '2026-01-01T00:00:00Z',
        proxies: [],
        accounts: []
      },
      skip_default_group_bind: true
    })
    expect(mocks.showError).toHaveBeenCalled()
    expect(mocks.showError.mock.calls.at(-1)?.[0]).toContain('bad.json')
    expect(wrapper.emitted('imported')).toHaveLength(1)
  })

  it('某个文件接口导入失败时继续导入后续文件并提示失败文件名', async () => {
    mocks.importData
      .mockRejectedValueOnce(new Error('server rejected'))
      .mockResolvedValueOnce({
        proxy_created: 0,
        proxy_reused: 0,
        proxy_failed: 0,
        account_created: 2,
        account_failed: 0
      })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const failedFile = new File(['{"exported_at":"2026-01-01T00:00:00Z","proxies":[],"accounts":[]}'], 'failed.json', {
      type: 'application/json'
    })
    Object.defineProperty(failedFile, 'text', {
      value: () => Promise.resolve('{"exported_at":"2026-01-01T00:00:00Z","proxies":[],"accounts":[]}')
    })
    const successFile = new File(['{"exported_at":"2026-01-02T00:00:00Z","proxies":[],"accounts":[]}'], 'success.json', {
      type: 'application/json'
    })
    Object.defineProperty(successFile, 'text', {
      value: () => Promise.resolve('{"exported_at":"2026-01-02T00:00:00Z","proxies":[],"accounts":[]}')
    })
    Object.defineProperty(input.element, 'files', {
      value: [failedFile, successFile]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(adminAPI.accounts.importData).toHaveBeenCalledTimes(2)
    expect(mocks.showError).toHaveBeenCalled()
    expect(mocks.showError.mock.calls.at(-1)?.[0]).toContain('failed.json')
    expect(mocks.showError.mock.calls.at(-1)?.[0]).toContain('server rejected')
    expect(wrapper.emitted('imported')).toHaveLength(1)
  })
})
