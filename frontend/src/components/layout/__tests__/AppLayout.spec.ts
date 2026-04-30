import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AppLayout from '../AppLayout.vue'

const appStoreMock = vi.hoisted(() => ({
  sidebarCollapsed: false
}))

const authStoreMock = vi.hoisted(() => ({
  user: { role: 'user' as const }
}))

const onboardingStoreMock = vi.hoisted(() => ({
  setReplayCallback: vi.fn()
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreMock
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => onboardingStoreMock
}))

vi.mock('@/composables/useOnboardingTour', () => ({
  useOnboardingTour: () => ({
    replayTour: vi.fn()
  })
}))

describe('AppLayout', () => {
  beforeEach(() => {
    onboardingStoreMock.setReplayCallback.mockReset()
    document.documentElement.classList.remove('dark')
  })

  it('disables the full-page mesh overlay in dark mode', () => {
    document.documentElement.classList.add('dark')

    const wrapper = mount(AppLayout, {
      global: {
        stubs: {
          AppSidebar: { template: '<aside />' },
          AppHeader: { template: '<header />' }
        }
      },
      slots: {
        default: '<div>content</div>'
      }
    })

    const overlay = wrapper.get('[data-test="app-layout-mesh"]')

    expect(overlay.classes()).toContain('bg-mesh-gradient')
    expect(overlay.classes()).toContain('dark:bg-none')
  })
})
