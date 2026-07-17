export default {
  setup: {
    title: 'GPTK 安装向导',
    description: '配置您的 GPTK 实例',
  },
  keys: {
    useKeyModal: {
      openai: {
        description: '复制并运行对应系统脚本，按提示手动输入 API Key，即可为 Codex CLI 与 Codex App 写入配置。',
        note: '脚本配置完成后，如 Codex App 已打开，请重启 Codex App 让它重新读取 ~/.codex 配置。',
        noteWindows: '脚本会写入 %userprofile%\\.codex。配置完成后，如 Codex App 已打开，请重启 Codex App。',
      },
    },
    quotaAmount: '额度金额 (CNY)',
    quotaAmountPlaceholder: '输入 CNY 额度限制',
    resetQuotaConfirmMessage: '确定要将密钥 "{name}" 的已用额度（¥{used}）重置为 0 吗？此操作不可撤销。',
    rateLimit5h: '5小时限额 (CNY)',
    rateLimit1d: '日限额 (CNY)',
    rateLimit7d: '7天限额 (CNY)',
  },
  affiliate: {
    description: '邀请新用户注册，获得可提现或可转余额的现金返利',
    transferFailed: '现金返利转余额失败',
    stats: {
      availableQuota: '可提现现金返利',
      frozenQuota: '待解冻现金返利',
      frozenQuotaHint: '新产生的现金返利正在冻结期中',
      totalQuota: '累计现金返利',
    },
    transfer: {
      title: '现金返利转平台余额',
      description: '按当前充值倍率转入平台余额；微信提现仍按现金金额发放。',
      button: '转为平台余额',
      empty: '当前没有可转入现金返利',
      success: '已转入平台余额：{amount}',
    },
  },
  redeem: {
    balanceAddedAffiliate: '平台余额充值（返利转入）',
  },
  admin: {
    users: {
      typeAffiliateBalance: '平台余额（返利转入）',
    },
    groups: {
      subscription: {
        dailyLimit: '每日限额（CNY）',
        weeklyLimit: '每周限额（CNY）',
        monthlyLimit: '每月限额（CNY）',
      },
    },
    accounts: {
      dataImportHint: '上传一个或多个导出的 JSON 文件以批量导入账号与代理。',
      dataImportCompletedWithErrors: '导入完成但有错误：账号失败 {account_failed}，代理失败 {proxy_failed}，文件失败 {file_failed}（{failed_files}）',
    },
    redeem: {
      amount: '金额 (¥)',
    },
    promo: {
      bonusAmount: '赠送金额 (¥)',
    },
    settings: {
      linuxdo: {
        description: '配置 LinuxDo Connect OAuth，用于 GPTK 用户登录',
      },
      site: {
        siteNamePlaceholder: 'GPTK',
      },
      payment: {
        balanceRechargeMultiplierHint: '用户每支付 1 CNY 可获得多少 CNY 余额',
        balanceRechargePreview: '预览：1 CNY = {usd} CNY',
      },
      smtp: {
        fromNamePlaceholder: 'GPTK',
      },
    },
  },
  onboarding: {
    admin: {
      welcome: {
        title: '👋 欢迎使用 GPTK',
        description: `<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">GPTK 是一个强大的 AI 服务中转平台，让您轻松管理和分发 AI 服务。</p><p style="margin-bottom: 12px;"><b>🎯 核心功能：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>📦 <b>分组管理</b> - 创建不同的服务套餐（VIP、免费试用等）</li><li>🔗 <b>账号池</b> - 连接多个上游 AI 服务商账号</li><li>🔑 <b>密钥分发</b> - 为用户生成独立的 API Key</li><li>💰 <b>计费管理</b> - 灵活的费率和配额控制</li></ul><p style="color: #10b981; font-weight: 600;">接下来，我们将用 3 分钟带您完成首次配置 →</p></div>`,
      },
      groupManage: {
        description: `<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>什么是分组？</b></p><p style="margin-bottom: 12px;">分组是 GPTK 的核心概念，它就像一个"服务套餐"：</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>🎯 每个分组可以包含多个上游账号</li><li>💰 每个分组有独立的计费倍率</li><li>👥 可以设置为公开或专属分组</li></ul><p style="margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 示例：</b>您可以创建"VIP专线"（高倍率）和"免费试用"（低倍率）两个分组</p><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 点击左侧的"分组管理"开始</p></div>`,
      },
      groupMultiplier: {
        description: `<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">设置该分组的计费倍率，控制用户的实际扣费。</p><div style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚙️ 计费规则：</b><ul style="margin: 8px 0 0 16px;"><li><b>1.0</b> - 原价计费（成本价）</li><li><b>1.5</b> - 用户消耗 ¥1，扣除 ¥1.5</li><li><b>2.0</b> - 用户消耗 ¥1，扣除 ¥2</li><li><b>0.8</b> - 补贴模式（亏本运营）</li></ul></div><p style="font-size: 13px; color: #6b7280;">建议测试分组设置为 1.0</p></div>`,
      },
    },
    user: {
      welcome: {
        title: '👋 欢迎使用 GPTK',
        description: `<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">您好！欢迎来到 GPTK AI 服务平台。</p><p style="margin-bottom: 12px;"><b>🎯 快速开始：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🔑 创建 API 密钥</li><li>📋 复制密钥到您的应用</li><li>🚀 开始使用 AI 服务</li></ul><p style="color: #10b981; font-weight: 600;">只需 1 分钟，让我们开始吧 →</p></div>`,
      },
    },
  },
  payment: {
    rechargeRatePreview: '当前倍率：1 CNY = {usd} CNY',
    admin: {
      insufficientBalance: '余额不足，将扣至 ¥0',
    },
  },
}
