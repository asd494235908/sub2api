import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CasdoorSuccessView from '@/views/auth/CasdoorSuccessView.vue'

const {
  routeState,
  routerReplaceMock,
  exchangeCasdoorTicketMock,
  persistOAuthTokenContextMock,
  setTokenMock,
} = vi.hoisted(() => ({
  routeState: {
    query: {} as Record<string, unknown>,
  },
  routerReplaceMock: vi.fn(),
  exchangeCasdoorTicketMock: vi.fn(),
  persistOAuthTokenContextMock: vi.fn(),
  setTokenMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    replace: (...args: any[]) => routerReplaceMock(...args),
  }),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    exchangeCasdoorTicket: (...args: any[]) => exchangeCasdoorTicketMock(...args),
    persistOAuthTokenContext: (...args: any[]) => persistOAuthTokenContextMock(...args),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    setToken: (...args: any[]) => setTokenMock(...args),
  }),
}))

describe('CasdoorSuccessView', () => {
  beforeEach(() => {
    routeState.query = {}
    routerReplaceMock.mockReset()
    exchangeCasdoorTicketMock.mockReset()
    persistOAuthTokenContextMock.mockReset()
    setTokenMock.mockReset()
  })

  it('exchanges ticket, persists token context, and redirects to same-site target', async () => {
    routeState.query = {
      ticket: 'ticket-1',
      redirect: '/billing',
    }
    exchangeCasdoorTicketMock.mockResolvedValue({
      access_token: 'access-1',
      refresh_token: 'refresh-1',
      expires_in: 3600,
      token_type: 'Bearer',
    })

    mount(CasdoorSuccessView)
    await vi.dynamicImportSettled()

    expect(exchangeCasdoorTicketMock).toHaveBeenCalledWith('ticket-1')
    expect(persistOAuthTokenContextMock).toHaveBeenCalledWith({
      access_token: 'access-1',
      refresh_token: 'refresh-1',
      expires_in: 3600,
      token_type: 'Bearer',
    })
    expect(setTokenMock).toHaveBeenCalledWith('access-1')
    expect(routerReplaceMock).toHaveBeenCalledWith('/billing')
  })

  it('redirects to error page when ticket is missing', async () => {
    routeState.query = {
      redirect: '/dashboard',
    }

    mount(CasdoorSuccessView)
    await vi.dynamicImportSettled()

    expect(exchangeCasdoorTicketMock).not.toHaveBeenCalled()
    expect(routerReplaceMock).toHaveBeenCalledWith({
      path: '/auth/casdoor/error',
      query: { error: 'missing_ticket' },
    })
  })

  it('rejects external redirect targets and falls back to dashboard', async () => {
    routeState.query = {
      ticket: 'ticket-2',
      redirect: 'https://evil.example/path',
    }
    exchangeCasdoorTicketMock.mockResolvedValue({
      access_token: 'access-2',
      token_type: 'Bearer',
    })

    mount(CasdoorSuccessView)
    await vi.dynamicImportSettled()

    expect(routerReplaceMock).toHaveBeenCalledWith('/dashboard')
  })

  it('redirects to error page when exchange fails', async () => {
    routeState.query = {
      ticket: 'ticket-3',
      redirect: '/dashboard',
    }
    exchangeCasdoorTicketMock.mockRejectedValue(new Error('ticket invalid'))

    mount(CasdoorSuccessView)
    await vi.dynamicImportSettled()

    expect(routerReplaceMock).toHaveBeenCalledWith({
      path: '/auth/casdoor/error',
      query: { error: 'exchange_failed' },
    })
  })
})
