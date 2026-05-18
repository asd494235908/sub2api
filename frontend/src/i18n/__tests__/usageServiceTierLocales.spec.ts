import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('usage service tier locale keys', () => {
  it('contains zh labels for service tier tooltip', () => {
    expect(zh.usage.serviceTier).toBe('服务档位')
    expect(zh.usage.serviceTierPriority).toBe('Fast')
    expect(zh.usage.serviceTierFlex).toBe('Flex')
    expect(zh.usage.serviceTierStandard).toBe('Standard')
  })

  it('contains en labels for service tier tooltip', () => {
    expect(en.usage.serviceTier).toBe('Service tier')
    expect(en.usage.serviceTierPriority).toBe('Fast')
    expect(en.usage.serviceTierFlex).toBe('Flex')
    expect(en.usage.serviceTierStandard).toBe('Standard')
  })
})

describe('admin redeem locale keys', () => {
  it('contains zh label for weekly quota redeem type', () => {
    expect(zh.admin.redeem.types.weekly_balance).toBe('周额度领取')
  })

  it('contains en label for weekly quota redeem type', () => {
    expect(en.admin.redeem.types.weekly_balance).toBe('Weekly Quota Claim')
  })

  it('contains lucky wheel redeem type labels', () => {
    expect(zh.admin.redeem.types.lucky_wheel_bonus).toBe('转盘奖励')
    expect(en.admin.redeem.types.lucky_wheel_bonus).toBe('Lucky Wheel Bonus')
  })

  it('contains lucky wheel admin copy labels', () => {
    expect(zh.luckyWheel.adminCopyTitle).toBe('页面文案')
    expect(en.luckyWheel.adminCopyTitle).toBe('Page Copy')
    expect(zh.luckyWheel.adminIntroTextLabel).toBe('活动介绍')
    expect(en.luckyWheel.adminIntroTextLabel).toBe('Activity Introduction')
  })
})
