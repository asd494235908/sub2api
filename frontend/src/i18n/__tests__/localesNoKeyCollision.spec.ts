import { describe, expect, it } from 'vitest'

import enAdminAccounts from '../locales/en/admin/accounts'
import enAdminChannels from '../locales/en/admin/channels'
import enAdminOps from '../locales/en/admin/ops'
import enAdminOverview from '../locales/en/admin/overview'
import enAdminResources from '../locales/en/admin/resources'
import enAdminSettings from '../locales/en/admin/settings'
import enCommon from '../locales/en/common'
import enDashboard from '../locales/en/dashboard'
import enLanding from '../locales/en/landing'
import enMisc from '../locales/en/misc'
import enExtensions from '../locales/en/extensions'
import en from '../locales/en'
import zhAdminAccounts from '../locales/zh/admin/accounts'
import zhAdminChannels from '../locales/zh/admin/channels'
import zhAdminOps from '../locales/zh/admin/ops'
import zhAdminOverview from '../locales/zh/admin/overview'
import zhAdminResources from '../locales/zh/admin/resources'
import zhAdminSettings from '../locales/zh/admin/settings'
import zhCommon from '../locales/zh/common'
import zhDashboard from '../locales/zh/dashboard'
import zhLanding from '../locales/zh/landing'
import zhMisc from '../locales/zh/misc'
import zhExtensions from '../locales/zh/extensions'
import zh from '../locales/zh'

// locales/{zh,en}/index.ts 与 admin/index.ts 使用对象展开聚合各域模块，
// 展开模块之间若出现同名顶层键会静默覆盖。本测试将该风险固化为显式失败。
type Modules = Record<string, Record<string, unknown>>

function collisions(modules: Modules): string[] {
  const seen = new Map<string, string>()
  const out: string[] = []
  for (const [name, mod] of Object.entries(modules)) {
    for (const key of Object.keys(mod)) {
      const prev = seen.get(key)
      if (prev) {
        out.push(`"${key}" in both ${prev} and ${name}`)
      } else {
        seen.set(key, name)
      }
    }
  }
  return out
}

function leafPaths(node: unknown, path = '', out: string[] = []): string[] {
  if (node && typeof node === 'object' && !Array.isArray(node)) {
    for (const [key, value] of Object.entries(node as Record<string, unknown>)) {
      leafPaths(value, path ? `${path}.${key}` : key, out)
    }
  } else {
    out.push(path)
  }
  return out
}

const roots: Record<string, Modules> = {
  zh: { landing: zhLanding, common: zhCommon, dashboard: zhDashboard, misc: zhMisc },
  en: { landing: enLanding, common: enCommon, dashboard: enDashboard, misc: enMisc }
}

const admins: Record<string, Modules> = {
  zh: {
    overview: zhAdminOverview,
    channels: zhAdminChannels,
    accounts: zhAdminAccounts,
    resources: zhAdminResources,
    ops: zhAdminOps,
    settings: zhAdminSettings
  },
  en: {
    overview: enAdminOverview,
    channels: enAdminChannels,
    accounts: enAdminAccounts,
    resources: enAdminResources,
    ops: enAdminOps,
    settings: enAdminSettings
  }
}

describe.each(Object.keys(roots))('locale %s spread assembly', (locale) => {
  it('root modules have no overlapping top-level keys', () => {
    expect(collisions(roots[locale])).toEqual([])
  })

  it('root modules do not shadow the explicit "admin" namespace', () => {
    for (const [name, mod] of Object.entries(roots[locale])) {
      expect(Object.keys(mod), `module ${name} must not define "admin"`).not.toContain('admin')
    }
  })

  it('admin modules have no overlapping top-level keys', () => {
    expect(collisions(admins[locale])).toEqual([])
  })
})

describe('local locale extensions', () => {
  it('keep migrated zh/en keys aligned', () => {
    const zhKeys = leafPaths(zhExtensions).sort()
    const enKeys = leafPaths(enExtensions).sort()

    expect(zhKeys).toHaveLength(681)
    expect(enKeys).toEqual(zhKeys)
  })

  it.each([
    ['zh', zhExtensions],
    ['en', enExtensions]
  ] as const)('%s additions do not replace upstream message leaves', (locale, extensions) => {
    const admin = Object.assign({}, ...Object.values(admins[locale]))
    const base = Object.assign({}, ...Object.values(roots[locale]), { admin })
    const baseKeys = new Set(leafPaths(base))

    expect(leafPaths(extensions).filter((key) => baseKeys.has(key))).toEqual([])
  })
})

describe('assembled locale compatibility', () => {
  it('keeps final zh/en message leaves aligned', () => {
    expect(leafPaths(en).sort()).toEqual(leafPaths(zh).sort())
  })

  it('preserves local branding, currency, and affiliate wording', () => {
    expect(en.setup.title).toBe('GPTK Setup')
    expect(zh.setup.title).toBe('GPTK 安装向导')
    expect(en.keys.quotaAmount).toBe('Quota Amount (CNY)')
    expect(zh.keys.quotaAmount).toBe('额度金额 (CNY)')
    expect(en.affiliate.transfer.title).toBe('Convert Cash Rebate')
    expect(zh.affiliate.transfer.title).toBe('现金返利转平台余额')
  })
})
