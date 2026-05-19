import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import SMSBroadcastsView from '../SMSBroadcastsView.vue'

const {
  listBroadcasts,
  createBroadcast,
  cancelBroadcast,
  listUsers,
  showError
} = vi.hoisted(() => ({
  listBroadcasts: vi.fn(),
  createBroadcast: vi.fn(),
  cancelBroadcast: vi.fn(),
  listUsers: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    smsBroadcasts: {
      list: listBroadcasts,
      create: createBroadcast,
      cancel: cancelBroadcast
    },
    users: {
      list: listUsers
    }
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'admin.smsBroadcasts.form.selectedCount') return `${params?.count} selected`
        return key
      }
    })
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError
  })
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => `FMT:${value}`
}))

const user = (overrides: Partial<AdminUser> = {}): AdminUser => ({
  id: 1,
  email: 'alice@example.com',
  username: 'alice',
  phone_number: '13800138000',
  role: 'user',
  balance: 0,
  concurrency: 1,
  status: 'active',
  allowed_groups: null,
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  notes: '',
  created_at: '2026-05-19T00:00:00Z',
  updated_at: '2026-05-19T00:00:00Z',
  ...overrides
})

const mountView = () =>
  mount(SMSBroadcastsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /></div>'
        },
        DataTable: {
          props: ['columns', 'data'],
          template: `
            <div>
              <div data-test="table-columns">{{ columns.map(column => column.key).join(',') }}</div>
              <div v-for="row in data" :key="row.id">
                <slot name="cell-title" :row="row" />
                <slot name="cell-created_at" :row="row" />
                <slot
                  v-if="columns.some(column => column.key === 'actions')"
                  name="cell-actions"
                  :row="row"
                />
              </div>
            </div>
          `
        },
        BaseDialog: {
          props: ['show'],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>'
        },
        Icon: true
      }
    }
  })

describe('admin SMSBroadcastsView', () => {
  beforeEach(() => {
    listBroadcasts.mockReset()
    createBroadcast.mockReset()
    cancelBroadcast.mockReset()
    listUsers.mockReset()
    showError.mockReset()

    listBroadcasts.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
    createBroadcast.mockResolvedValue({ id: 1 })
    listUsers.mockResolvedValue({
      items: [
        user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' }),
        user({ id: 2, username: 'bob', email: 'bob@example.com', phone_number: '' }),
        user({ id: 3, username: 'carol', email: 'carol@example.com', phone_number: '13900139000' })
      ],
      total: 3,
      page: 1,
      page_size: 20,
      pages: 1
    })
  })

  it('renders loaded broadcast campaigns in the table area', async () => {
    listBroadcasts.mockResolvedValue({
      items: [{
        id: 9,
        title: 'Template launch',
        template_id: '309190',
        status: 'succeeded',
        total_recipients: 1,
        sent_count: 1,
        failed_count: 0,
        created_at: '2026-05-19T00:00:00Z'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Template launch')
    expect(wrapper.text()).toContain('FMT:2026-05-19T00:00:00Z')
  })

  it('hides the actions column when loaded broadcasts have no available actions', async () => {
    listBroadcasts.mockResolvedValue({
      items: [{
        id: 9,
        title: 'Done campaign',
        template_id: '309190',
        status: 'succeeded',
        total_recipients: 1,
        sent_count: 1,
        failed_count: 0,
        created_at: '2026-05-19T00:00:00Z'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-test="table-columns"]').text()).not.toContain('actions')
  })

  it('shows the actions column when a loaded broadcast can be canceled', async () => {
    listBroadcasts.mockResolvedValue({
      items: [{
        id: 10,
        title: 'Queued campaign',
        template_id: '309190',
        status: 'queued',
        total_recipients: 1,
        sent_count: 0,
        failed_count: 0,
        created_at: '2026-05-19T00:00:00Z'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-test="table-columns"]').text()).toContain('actions')
  })

  it('renders a visible cancel action for cancellable broadcasts', async () => {
    listBroadcasts.mockResolvedValue({
      items: [
        {
          id: 10,
          title: 'Queued campaign',
          template_id: '309190',
          status: 'queued',
          total_recipients: 1,
          sent_count: 0,
          failed_count: 0,
          created_at: '2026-05-19T00:00:00Z'
        },
        {
          id: 11,
          title: 'Done campaign',
          template_id: '309190',
          status: 'succeeded',
          total_recipients: 1,
          sent_count: 1,
          failed_count: 0,
          created_at: '2026-05-19T00:00:00Z'
        }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('common.cancel')
  })

  it('adds and removes selected users and creates broadcasts with explicit user ids', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()

    expect(listUsers).toHaveBeenLastCalledWith(1, 20, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))

    await wrapper.get('[data-test="sms-title"]').setValue('Maintenance')
    await wrapper.get('[data-test="sms-template-id"]').setValue('broadcast-template')
    await wrapper.get('[data-test="sms-var-key-0"]').setValue('window')
    await wrapper.get('[data-test="sms-var-value-0"]').setValue('tonight')
    await wrapper.get('[data-test="sms-add-var"]').trigger('click')
    await wrapper.get('[data-test="sms-var-key-1"]').setValue('contact')
    await wrapper.get('[data-test="sms-var-value-1"]').setValue('support')
    await wrapper.get('[data-test="sms-remove-var-1"]').trigger('click')

    await wrapper.get('[data-test="sms-add-user-1"]').trigger('click')
    await wrapper.get('[data-test="sms-add-user-1"]').trigger('click')
    await wrapper.get('[data-test="sms-add-user-2"]').trigger('click')
    await wrapper.get('[data-test="sms-add-user-3"]').trigger('click')
    await wrapper.get('[data-test="sms-remove-user-1"]').trigger('click')

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(createBroadcast).toHaveBeenLastCalledWith(expect.objectContaining({
      title: 'Maintenance',
      template_id: 'broadcast-template',
      vars: [
        { key: 'window', value: 'tonight' }
      ],
      audience: { user_ids: [3] }
    }))
    expect(createBroadcast.mock.calls.at(-1)?.[0]).not.toHaveProperty('mode')
    expect(createBroadcast.mock.calls.at(-1)?.[0]).not.toHaveProperty('body')
    expect(createBroadcast.mock.calls.at(-1)?.[0].audience).not.toHaveProperty('search')
    expect(createBroadcast.mock.calls.at(-1)?.[0].audience).not.toHaveProperty('status')
    expect(createBroadcast.mock.calls.at(-1)?.[0].audience).not.toHaveProperty('role')
  })

  it('loads more audience users, appends unique users, and keeps selected recipients', async () => {
    listUsers
      .mockResolvedValueOnce({
        items: [
          user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' }),
          user({ id: 2, username: 'bob', email: 'bob@example.com', phone_number: '13900139000' })
        ],
        total: 4,
        page: 1,
        page_size: 2,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [
          user({ id: 2, username: 'bob', email: 'bob@example.com', phone_number: '13900139000' }),
          user({ id: 3, username: 'carol', email: 'carol@example.com', phone_number: '13700137000' })
        ],
        total: 4,
        page: 2,
        page_size: 2,
        pages: 2
      })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-test="sms-add-user-1"]').trigger('click')
    await wrapper.get('[data-test="sms-load-more-users"]').trigger('click')
    await flushPromises()

    expect(listUsers).toHaveBeenNthCalledWith(1, 1, 20, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
    expect(listUsers).toHaveBeenNthCalledWith(2, 2, 20, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
    expect(wrapper.findAll('[data-test="sms-add-user-2"]')).toHaveLength(1)
    expect(wrapper.find('[data-test="sms-add-user-3"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('1 selected')
    expect(wrapper.text()).toContain('admin.smsBroadcasts.form.usersLoadedSummary')
  })

  it('resets audience pagination when filters change', async () => {
    listUsers
      .mockResolvedValueOnce({
        items: [user({ id: 1, email: 'alice@example.com', username: 'alice' })],
        total: 2,
        page: 1,
        page_size: 1,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [user({ id: 2, email: 'bob@example.com', username: 'bob' })],
        total: 2,
        page: 2,
        page_size: 1,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [user({ id: 4, email: 'admin@example.com', username: 'admin', role: 'admin' })],
        total: 1,
        page: 1,
        page_size: 1,
        pages: 1
      })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="sms-load-more-users"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="sms-audience-role"]').setValue('admin')
    await flushPromises()

    expect(listUsers).toHaveBeenLastCalledWith(1, 20, {
      status: 'active',
      role: 'admin',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
    expect(wrapper.find('[data-test="sms-add-user-1"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="sms-add-user-2"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="sms-add-user-4"]').exists()).toBe(true)
  })

  it('changes audience page size and reloads from the first page', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="sms-audience-page-size"]').setValue('50')
    await flushPromises()

    expect(listUsers).toHaveBeenLastCalledWith(1, 50, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
  })

  it('adds all phone-bound users across pages using the current filters', async () => {
    listUsers
      .mockResolvedValueOnce({
        items: [
          user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' }),
          user({ id: 2, username: 'bob', email: 'bob@example.com', phone_number: '' })
        ],
        total: 4,
        page: 1,
        page_size: 20,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [
          user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' }),
          user({ id: 2, username: 'bob', email: 'bob@example.com', phone_number: '' })
        ],
        total: 4,
        page: 1,
        page_size: 20,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [
          user({ id: 3, username: 'carol', email: 'carol@example.com', phone_number: '13900139000' }),
          user({ id: 4, username: 'dave', email: 'dave@example.com', phone_number: ' ' })
        ],
        total: 4,
        page: 2,
        page_size: 20,
        pages: 2
      })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="sms-add-user-1"]').trigger('click')
    await wrapper.get('[data-test="sms-add-all-phone-users"]').trigger('click')
    await flushPromises()

    expect(listUsers).toHaveBeenNthCalledWith(2, 1, 20, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
    expect(listUsers).toHaveBeenNthCalledWith(3, 2, 20, {
      status: 'active',
      role: 'user',
      search: undefined,
      include_subscriptions: false,
      has_phone: true
    }, expect.any(Object))
    expect(wrapper.text()).toContain('2 selected')
    expect(wrapper.find('[data-test="sms-remove-user-1"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="sms-remove-user-3"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="sms-remove-user-2"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="sms-remove-user-4"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="sms-add-user-3"]').exists()).toBe(true)
  })

  it('does not partially add users when adding all phone-bound users fails', async () => {
    listUsers
      .mockResolvedValueOnce({
        items: [user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' })],
        total: 2,
        page: 1,
        page_size: 20,
        pages: 2
      })
      .mockResolvedValueOnce({
        items: [user({ id: 1, username: 'alice', email: 'alice@example.com', phone_number: '13800138000' })],
        total: 2,
        page: 1,
        page_size: 20,
        pages: 2
      })
      .mockRejectedValueOnce({ response: { data: { detail: 'boom' } } })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="sms-add-all-phone-users"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('0 selected')
    expect(showError).toHaveBeenCalledWith('boom')
  })

  it('blocks empty variable values and duplicate keys before submit', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-test="sms-title"]').setValue('Maintenance')
    await wrapper.get('[data-test="sms-template-id"]').setValue('broadcast-template')
    await wrapper.get('[data-test="sms-var-key-0"]').setValue('window')

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()
    expect(createBroadcast).not.toHaveBeenCalled()

    await wrapper.get('[data-test="sms-var-value-0"]').setValue('tonight')
    await wrapper.get('[data-test="sms-add-var"]').trigger('click')
    await wrapper.get('[data-test="sms-var-key-1"]').setValue('window')
    await wrapper.get('[data-test="sms-var-value-1"]').setValue('tomorrow')

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()
    expect(createBroadcast).not.toHaveBeenCalled()
  })

  it('blocks submit until at least one user is selected', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="sms-broadcast-create"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-test="sms-title"]').setValue('Maintenance')
    await wrapper.get('[data-test="sms-template-id"]').setValue('broadcast-template')

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(createBroadcast).not.toHaveBeenCalled()
  })
})
