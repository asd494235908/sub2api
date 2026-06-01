import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import GroupSelector from '@/components/common/GroupSelector.vue'
import type { AdminGroup, GroupPlatform } from '@/types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'common.selectedCount') return `(${params?.count} selected)`
        if (key === 'common.selectAll') return 'Select all'
        return key
      },
    }),
  }
})

function buildGroup(id: number, platform: GroupPlatform, name = `${platform}-${id}`): AdminGroup {
  return {
    id,
    name,
    description: null,
    platform,
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    subscription_total_limit_usd: null,
    allow_image_generation: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: false,
  }
}

function mountSelector(props: Partial<InstanceType<typeof GroupSelector>['$props']> = {}) {
  return mount(GroupSelector, {
    props: {
      modelValue: [],
      groups: [],
      ...props,
    },
    global: {
      stubs: {
        Icon: true,
        GroupBadge: {
          props: ['name'],
          template: '<span>{{ name }}</span>',
        },
      },
    },
  })
}

describe('GroupSelector', () => {
  it('selects only groups allowed by the current platform', async () => {
    const wrapper = mountSelector({
      platform: 'openai',
      groups: [
        buildGroup(1, 'openai'),
        buildGroup(2, 'anthropic'),
        buildGroup(3, 'openai'),
      ],
    })

    await wrapper.get('[data-testid="group-selector-select-all"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([[1, 3]])
  })

  it('selects all platform-allowed groups even when search hides some of them', async () => {
    const wrapper = mountSelector({
      platform: 'openai',
      searchable: true,
      groups: [
        buildGroup(1, 'openai', 'visible match'),
        buildGroup(2, 'openai', 'hidden choice'),
        buildGroup(3, 'anthropic', 'visible other platform'),
      ],
    })

    await wrapper.get('input[type="text"]').setValue('visible')
    await wrapper.get('[data-testid="group-selector-select-all"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([[1, 2]])
  })

  it('includes anthropic and gemini groups for antigravity mixed scheduling', async () => {
    const wrapper = mountSelector({
      platform: 'antigravity',
      mixedScheduling: true,
      groups: [
        buildGroup(1, 'antigravity'),
        buildGroup(2, 'anthropic'),
        buildGroup(3, 'gemini'),
        buildGroup(4, 'openai'),
      ],
    })

    await wrapper.get('[data-testid="group-selector-select-all"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([[1, 2, 3]])
  })

  it('disables select all when no available groups remain to select', () => {
    const fullySelected = mountSelector({
      platform: 'openai',
      modelValue: [1],
      groups: [buildGroup(1, 'openai'), buildGroup(2, 'anthropic')],
    })
    expect(fullySelected.get('[data-testid="group-selector-select-all"]').attributes('disabled')).toBeDefined()

    const noAvailableGroups = mountSelector({
      platform: 'openai',
      groups: [buildGroup(2, 'anthropic')],
    })
    expect(noAvailableGroups.get('[data-testid="group-selector-select-all"]').attributes('disabled')).toBeDefined()
  })
})
