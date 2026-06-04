import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import LeaderboardView from '../LeaderboardView.vue'

const { getLeaderboardMock } = vi.hoisted(() => ({
  getLeaderboardMock: vi.fn(),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: {
    getLeaderboard: getLeaderboardMock,
  },
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: () => 'load failed',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'leaderboard.title': '消耗榜单',
    'leaderboard.description': '榜单说明',
    'leaderboard.tableTitle': 'Token 消耗排名',
    'leaderboard.failedToLoad': '加载失败',
    'leaderboard.empty': '暂无榜单数据',
    'leaderboard.periods.today': '今日',
    'leaderboard.periods.yesterday': '昨日',
    'leaderboard.summary.period': '统计日期',
    'leaderboard.summary.tokens': '总 Token',
    'leaderboard.summary.requests': '总请求',
    'leaderboard.columns.rank': '排名',
    'leaderboard.columns.user': '用户',
    'leaderboard.columns.tokens': 'Token',
    'leaderboard.columns.requests': '请求',
    'leaderboard.columns.actualCost': '消耗',
    'common.refresh': '刷新',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

function mountView() {
  return mount(LeaderboardView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
      },
    },
  })
}

describe('LeaderboardView', () => {
  beforeEach(() => {
    getLeaderboardMock.mockReset()
    getLeaderboardMock.mockResolvedValue({
      period: 'today',
      start_date: '2026-06-02',
      end_date: '2026-06-02',
      total_tokens: 1024,
      total_requests: 8,
      ranking: [
        {
          rank: 1,
          user_id: 123,
          display_name: 'u***@example.com',
          tokens: 1024,
          requests: 8,
          actual_cost: 1.25,
        },
      ],
    })
  })

  it('renders ranking rows from the today leaderboard', async () => {
    const wrapper = mountView()

    await flushPromises()

    expect(getLeaderboardMock).toHaveBeenCalledWith({ period: 'today', limit: 20 })
    expect(wrapper.text()).toContain('消耗榜单')
    expect(wrapper.text()).toContain('u***@example.com')
    expect(wrapper.text()).toContain('1,024')
    expect(wrapper.text()).toContain('¥1.2500')
    expect(wrapper.text()).not.toContain('user@example.com')
  })

  it('reloads the ranking when switching to yesterday', async () => {
    const wrapper = mountView()
    await flushPromises()

    getLeaderboardMock.mockResolvedValueOnce({
      period: 'yesterday',
      start_date: '2026-06-01',
      end_date: '2026-06-01',
      total_tokens: 2048,
      total_requests: 9,
      ranking: [],
    })

    const yesterdayButton = wrapper.findAll('button').find((button) => button.text() === '昨日')
    expect(yesterdayButton).toBeDefined()
    await yesterdayButton?.trigger('click')
    await flushPromises()

    expect(getLeaderboardMock).toHaveBeenLastCalledWith({ period: 'yesterday', limit: 20 })
    expect(wrapper.text()).toContain('暂无榜单数据')
    expect(wrapper.text()).toContain('2,048')
    expect(wrapper.text()).toContain('9')
  })

  it('refreshes the current period when clicking refresh', async () => {
    const wrapper = mountView()
    await flushPromises()

    getLeaderboardMock.mockResolvedValueOnce({
      period: 'today',
      start_date: '2026-06-02',
      end_date: '2026-06-02',
      total_tokens: 4096,
      total_requests: 12,
      ranking: [
        {
          rank: 1,
          user_id: 456,
          display_name: 'n***',
          tokens: 4096,
          requests: 12,
          actual_cost: 2.5,
        },
      ],
    })

    const refreshButton = wrapper.findAll('button').find((button) => button.text().includes('刷新'))
    expect(refreshButton).toBeDefined()
    await refreshButton?.trigger('click')
    await flushPromises()

    expect(getLeaderboardMock).toHaveBeenCalledTimes(2)
    expect(getLeaderboardMock).toHaveBeenLastCalledWith({ period: 'today', limit: 20 })
    expect(wrapper.text()).toContain('4,096')
    expect(wrapper.text()).toContain('n***')
  })

  it('shows empty state and API errors', async () => {
    getLeaderboardMock.mockResolvedValueOnce({
      period: 'today',
      start_date: '2026-06-02',
      end_date: '2026-06-02',
      total_tokens: 0,
      total_requests: 0,
      ranking: [],
    })
    const emptyWrapper = mountView()
    await flushPromises()
    expect(emptyWrapper.text()).toContain('暂无榜单数据')

    getLeaderboardMock.mockRejectedValueOnce(new Error('boom'))
    const errorWrapper = mountView()
    await flushPromises()
    expect(errorWrapper.text()).toContain('load failed')
  })

  it('clears stale ranking rows when a reload fails', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('u***@example.com')

    getLeaderboardMock.mockRejectedValueOnce(new Error('boom'))
    const refreshButton = wrapper.findAll('button').find((button) => button.text().includes('刷新'))
    await refreshButton?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('load failed')
    expect(wrapper.text()).not.toContain('u***@example.com')
  })
})
