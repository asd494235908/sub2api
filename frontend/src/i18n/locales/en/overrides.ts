export default {
  setup: {
    title: 'GPTK Setup',
    description: 'Configure your GPTK instance',
  },
  keys: {
    useKeyModal: {
      openai: {
        description: 'Copy and run the script for your system, then enter the API Key manually to configure Codex CLI and Codex App.',
        note: 'After the script finishes, restart Codex App if it is already open so it reloads ~/.codex.',
        noteWindows: 'The script writes to %userprofile%\\.codex. After it finishes, restart Codex App if it is already open.',
      },
    },
    quotaAmount: 'Quota Amount (CNY)',
    quotaAmountPlaceholder: 'Enter quota limit in CNY',
    rateLimit5h: '5-Hour Limit (CNY)',
    rateLimit1d: 'Daily Limit (CNY)',
    rateLimit7d: '7-Day Limit (CNY)',
  },
  affiliate: {
    description: 'Invite new users and earn cash rebates you can withdraw or convert to balance',
    transferFailed: 'Failed to convert cash rebate',
    stats: {
      availableQuota: 'Withdrawable Cash Rebate',
      frozenQuota: 'Pending Cash Rebate',
      frozenQuotaHint: 'Recently earned cash rebates pending release',
      totalQuota: 'Total Cash Rebates',
    },
    transfer: {
      title: 'Convert Cash Rebate',
      description: 'Convert cash rebate into platform balance using the current recharge multiplier. Withdrawals still pay the cash amount.',
      button: 'Convert to Balance',
      empty: 'No available cash rebate',
    },
  },
  redeem: {
    balanceAddedAffiliate: 'Platform Balance Added (Affiliate Transfer)',
  },
  admin: {
    users: {
      typeAffiliateBalance: 'Platform Balance (Affiliate Transfer)',
    },
    groups: {
      subscription: {
        dailyLimit: 'Daily Limit (CNY)',
        weeklyLimit: 'Weekly Limit (CNY)',
        monthlyLimit: 'Monthly Limit (CNY)',
      },
    },
    accounts: {
      dataImportHint: 'Upload one or more exported JSON files to import accounts and proxies.',
      dataImportCompletedWithErrors: 'Import completed with errors: account failed {account_failed}, proxy failed {proxy_failed}, file failed {file_failed} ({failed_files})',
      quotaLimitHint: "Set daily/weekly/total spending limits (CNY). Anthropic API key accounts can also configure client affinity. Changing limits won't reset usage.",
    },
    redeem: {
      amount: 'Amount (¥)',
    },
    promo: {
      bonusAmount: 'Bonus Amount (¥)',
    },
    settings: {
      linuxdo: {
        description: 'Configure LinuxDo Connect OAuth for GPTK end-user login',
      },
      site: {
        siteNamePlaceholder: 'GPTK',
        siteSubtitlePlaceholder: 'GPTK AI Service Platform',
      },
      payment: {
        balanceRechargeMultiplierHint: 'How many CNY balance the user receives for each 1 CNY paid',
        balanceRechargePreview: 'Preview: 1 CNY = {usd} CNY',
      },
      smtp: {
        fromNamePlaceholder: 'GPTK',
      },
    },
  },
  onboarding: {
    admin: {
      welcome: {
        title: '👋 Welcome to GPTK',
        description: `<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">GPTK is a powerful AI service gateway platform that helps you easily manage and distribute AI services.</p><p style="margin-bottom: 12px;"><b>🎯 Core Features:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>📦 <b>Group Management</b> - Create service tiers (VIP, Free Trial, etc.)</li><li>🔗 <b>Account Pool</b> - Connect multiple upstream AI service accounts</li><li>🔑 <b>Key Distribution</b> - Generate independent API Keys for users</li><li>💰 <b>Billing Control</b> - Flexible rate and quota management</li></ul><p style="color: #10b981; font-weight: 600;">Let's complete the initial setup in 3 minutes →</p></div>`,
      },
      groupManage: {
        description: `<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>What is a Group?</b></p><p style="margin-bottom: 12px;">Groups are the core concept of GPTK, like a "service package":</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>🎯 Each group can contain multiple upstream accounts</li><li>💰 Each group has independent billing multiplier</li><li>👥 Can be set as public or exclusive</li></ul><p style="margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Example:</b> You can create "VIP Premium" (high rate) and "Free Trial" (low rate) groups</p><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "Group Management" on the left sidebar</p></div>`,
      },
      groupMultiplier: {
        description: `<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set the billing multiplier to control user charges.</p><div style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚙️ Billing Rules:</b><ul style="margin: 8px 0 0 16px;"><li><b>1.0</b> - Original price (cost price)</li><li><b>1.5</b> - User consumes ¥1, charged ¥1.5</li><li><b>2.0</b> - User consumes ¥1, charged ¥2</li><li><b>0.8</b> - Subsidy mode (loss-making)</li></ul></div><p style="font-size: 13px; color: #6b7280;">Recommend setting test group to 1.0</p></div>`,
      },
    },
    user: {
      welcome: {
        title: '👋 Welcome to GPTK',
        description: `<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Hello! Welcome to the GPTK AI service platform.</p><p style="margin-bottom: 12px;"><b>🎯 Quick Start:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🔑 Create API Key</li><li>📋 Copy key to your application</li><li>🚀 Start using AI services</li></ul><p style="color: #10b981; font-weight: 600;">Just 1 minute, let's get started →</p></div>`,
      },
    },
  },
  payment: {
    rechargeRatePreview: 'Current rate: 1 CNY = {usd} CNY',
    admin: {
      insufficientBalance: 'Insufficient balance — will deduct to ¥0',
    },
  },
}
