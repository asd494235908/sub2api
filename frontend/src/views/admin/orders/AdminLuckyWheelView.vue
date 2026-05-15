<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-2xl font-black text-slate-900 dark:text-white">{{ t('nav.luckyWheelAdmin') }}</h1>
          <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('luckyWheel.adminDescription') }}</p>
        </div>
        <div class="flex gap-3">
          <button class="btn btn-secondary" :disabled="loading || saving" @click="loadAll">{{ t('luckyWheel.refresh') }}</button>
          <button
            class="btn btn-primary"
            data-test="save-config"
            :disabled="saving || validationIssues.length > 0"
            @click="saveConfig"
          >
            {{ saving ? t('common.processing') : t('common.save') }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-16">
        <LoadingSpinner />
      </div>

      <template v-else>
        <div class="grid gap-6 xl:grid-cols-[1.2fr,0.8fr]">
          <section class="space-y-6 rounded-[28px] border border-slate-200 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
            <div class="flex items-center justify-between">
              <div>
                <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ t('luckyWheel.adminEnabledTitle') }}</h2>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('luckyWheel.adminEnabledHint') }}</p>
              </div>
              <Toggle v-model="enabled" />
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <label class="space-y-2">
                <span class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminMultiplierStep') }}</span>
                <input
                  v-model.number="config.multiplier_step"
                  data-test="multiplier-step"
                  type="number"
                  min="0.1"
                  step="0.1"
                  class="input input-bordered w-full"
                />
              </label>
              <label class="space-y-2">
                <span class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminGlobalMaxMultiplier') }}</span>
                <input
                  v-model.number="config.global_max_multiplier"
                  type="number"
                  min="1"
                  step="0.1"
                  class="input input-bordered w-full"
                />
              </label>
            </div>

            <div class="rounded-3xl border border-slate-200 p-4">
              <div class="mb-4 flex items-center justify-between">
                <div>
                  <h3 class="text-base font-bold text-slate-900">{{ t('luckyWheel.adminOrderTypesTitle') }}</h3>
                  <p class="text-sm text-slate-500">{{ t('luckyWheel.adminOrderTypesHint') }}</p>
                </div>
              </div>
              <div class="flex flex-wrap gap-3">
                <label class="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-2 text-sm text-slate-700">
                  <input v-model="eligibleOrderTypeSet.balance" type="checkbox" class="checkbox checkbox-sm" />
                  <span>{{ t('luckyWheel.balanceOrder') }}</span>
                </label>
                <label class="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-2 text-sm text-slate-700">
                  <input v-model="eligibleOrderTypeSet.subscription" type="checkbox" class="checkbox checkbox-sm" />
                  <span>{{ t('luckyWheel.subscriptionOrder') }}</span>
                </label>
              </div>
            </div>

            <div v-if="validationIssues.length > 0" class="rounded-2xl border border-amber-200 bg-amber-50 p-4">
              <h3 class="text-sm font-bold text-amber-800">{{ t('luckyWheel.adminValidationTitle') }}</h3>
              <ul class="mt-2 space-y-1 text-sm text-amber-700">
                <li v-for="issue in validationIssues" :key="issue">• {{ issue }}</li>
              </ul>
            </div>

            <div class="rounded-3xl border border-slate-200 p-4">
              <div class="mb-4">
                <h3 class="text-base font-bold text-slate-900">{{ t('luckyWheel.adminCopyTitle') }}</h3>
                <p class="text-sm text-slate-500">{{ t('luckyWheel.adminCopyHint') }}</p>
              </div>
              <div class="space-y-4">
                <label class="space-y-2">
                  <span class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminIntroTextLabel') }}</span>
                  <textarea v-model="config.intro_text" data-test="intro-text" rows="3" class="textarea textarea-bordered w-full" />
                </label>
                <label class="space-y-2">
                  <span class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminRulesTitleLabel') }}</span>
                  <input v-model="config.rules_title" data-test="rules-title" class="input input-bordered w-full" />
                </label>
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminRulesItemsLabel') }}</span>
                    <button class="btn btn-secondary btn-sm" type="button" @click="addRuleItem">{{ t('luckyWheel.adminAddRuleItem') }}</button>
                  </div>
                  <div v-for="(_rule, index) in config.rules_items" :key="`rule-${index}`" class="flex items-center gap-3">
                    <input v-model="config.rules_items[index]" data-test="rules-item" class="input input-bordered w-full" />
                    <button class="btn btn-ghost btn-xs text-rose-500" type="button" @click="removeRuleItem(index)">{{ t('common.delete') }}</button>
                  </div>
                </div>
              </div>
            </div>

            <div class="rounded-3xl border border-slate-200 p-4">
              <div class="mb-4 flex items-center justify-between">
                <div>
                  <h3 class="text-base font-bold text-slate-900">{{ t('luckyWheel.adminTiersTitle') }}</h3>
                  <p class="text-sm text-slate-500">{{ t('luckyWheel.adminTiersHint') }}</p>
                </div>
                <button class="btn btn-secondary btn-sm" data-test="add-tier" @click="addTier">{{ t('luckyWheel.adminAddTier') }}</button>
              </div>

              <div class="space-y-4">
                <div
                  v-for="(tier, index) in config.amount_tiers"
                  :key="`${tier.id}-${index}`"
                  data-test="tier-row"
                  class="rounded-2xl border border-slate-100 bg-slate-50/70 p-4"
                >
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <div class="text-sm font-semibold text-slate-700">{{ t('luckyWheel.adminTierLabel', { value: index + 1 }) }}</div>
                    <button class="btn btn-ghost btn-xs text-rose-500" @click="removeTier(index)">{{ t('common.delete') }}</button>
                  </div>
                  <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                    <label class="space-y-2">
                      <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierIdLabel') }}</span>
                      <input v-model.trim="tier.id" placeholder="tier_id" class="input input-bordered w-full" />
                    </label>
                    <label class="space-y-2">
                      <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierNameLabel') }}</span>
                      <input v-model.trim="tier.name" placeholder="tier_name" class="input input-bordered w-full" />
                    </label>
                    <label class="space-y-2">
                      <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierDrawCountLabel') }}</span>
                      <input v-model.number="tier.draw_count" type="number" min="1" step="1" class="input input-bordered w-full" />
                    </label>
                    <label class="space-y-2">
                      <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierMinAmountLabel') }}</span>
                      <input v-model.number="tier.min_amount" type="number" min="0" step="1" class="input input-bordered w-full" />
                    </label>
                    <label class="space-y-2">
                      <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierMaxAmountLabel') }}</span>
                      <input :value="tier.max_amount ?? ''" class="input input-bordered w-full" @input="updateTierMax(index, ($event.target as HTMLInputElement).value)" />
                    </label>
                    <div class="grid grid-cols-2 gap-3">
                      <label class="space-y-2">
                        <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierMinMultiplierLabel') }}</span>
                        <input v-model.number="tier.min_multiplier" data-test="tier-min-multiplier" type="number" min="0.01" step="0.1" class="input input-bordered w-full" />
                      </label>
                      <label class="space-y-2">
                        <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminTierMaxMultiplierLabel') }}</span>
                        <input v-model.number="tier.max_multiplier" data-test="tier-max-multiplier" type="number" min="0.01" step="0.1" class="input input-bordered w-full" />
                      </label>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="grid gap-6 lg:grid-cols-2">
              <div class="rounded-3xl border border-slate-200 p-4">
                <div class="mb-4 flex items-center justify-between">
                  <div>
                    <h3 class="text-base font-bold text-slate-900">{{ t('luckyWheel.adminInviteBonusTitle') }}</h3>
                    <p class="text-sm text-slate-500">{{ t('luckyWheel.adminInviteBonusHint') }}</p>
                  </div>
                  <Toggle v-model="config.invite_bonus.enabled" />
                </div>
                <div class="grid gap-3 md:grid-cols-2">
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminInviteQualifyingAmountLabel') }}</span>
                    <input v-model.number="config.invite_bonus.qualifying_amount" type="number" min="0" step="1" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminInviteBonusPerInviteeLabel') }}</span>
                    <input
                      v-model.number="config.invite_bonus.bonus_per_invitee"
                      data-test="invite-bonus-per-invitee"
                      type="number"
                      min="0"
                      step="0.1"
                      class="input input-bordered w-full"
                    />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminInviteMaxBonusLabel') }}</span>
                    <input v-model.number="config.invite_bonus.max_bonus" type="number" min="0" step="0.1" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminInviteConsumePolicyLabel') }}</span>
                    <input v-model.trim="config.invite_bonus.consume_policy" class="input input-bordered w-full" />
                  </label>
                </div>
              </div>

              <div class="rounded-3xl border border-slate-200 p-4">
                <div class="mb-4 flex items-center justify-between">
                  <div>
                    <h3 class="text-base font-bold text-slate-900">{{ t('luckyWheel.adminGoldenWindowTitle') }}</h3>
                    <p class="text-sm text-slate-500">{{ t('luckyWheel.adminGoldenWindowHint') }}</p>
                  </div>
                  <Toggle v-model="config.golden_window.enabled" />
                </div>
                <div class="grid gap-3 md:grid-cols-2">
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenTimezoneLabel') }}</span>
                    <input v-model.trim="config.golden_window.timezone" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenStartTimeLabel') }}</span>
                    <input v-model.trim="config.golden_window.start_time" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenEndTimeLabel') }}</span>
                    <input v-model.trim="config.golden_window.end_time" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenMinAmountLabel') }}</span>
                    <input v-model.number="config.golden_window.min_amount" type="number" min="0" step="1" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenExtraDrawsLabel') }}</span>
                    <input v-model.number="config.golden_window.extra_draws" type="number" min="0" step="1" class="input input-bordered w-full" />
                  </label>
                  <label class="space-y-2">
                    <span class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">{{ t('luckyWheel.adminGoldenDailyQuotaLabel') }}</span>
                    <input v-model.number="config.golden_window.daily_quota" type="number" min="0" step="1" class="input input-bordered w-full" />
                  </label>
                </div>
              </div>
            </div>
          </section>

          <section class="space-y-6">
            <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-1">
              <div class="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
                <div class="text-sm text-slate-500">{{ t('luckyWheel.adminTotalSessions') }}</div>
                <div class="mt-2 text-3xl font-black text-amber-500">{{ stats?.total_sessions ?? 0 }}</div>
              </div>
              <div class="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
                <div class="text-sm text-slate-500">{{ t('luckyWheel.adminPendingSessions') }}</div>
                <div class="mt-2 text-3xl font-black text-blue-500">{{ stats?.pending_sessions ?? 0 }}</div>
              </div>
              <div class="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
                <div class="text-sm text-slate-500">{{ t('luckyWheel.adminSettledSessions') }}</div>
                <div class="mt-2 text-3xl font-black text-emerald-500">{{ stats?.settled_sessions ?? 0 }}</div>
              </div>
              <div class="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
                <div class="text-sm text-slate-500">{{ t('luckyWheel.adminTotalBonusAmount') }}</div>
                <div class="mt-2 text-3xl font-black text-rose-500">{{ (stats?.total_bonus_amount ?? 0).toFixed(2) }}</div>
              </div>
            </div>

            <div class="rounded-[28px] border border-slate-200 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-bold text-slate-900">{{ t('luckyWheel.adminDistributionTitle') }}</h2>
                <span class="text-xs uppercase tracking-[0.2em] text-slate-400">
                  {{ stats?.golden_window_used_today ?? 0 }}/{{ stats?.golden_window_daily_quota ?? 0 }}
                </span>
              </div>
              <div v-if="!multiplierDistribution.length" class="py-8 text-center text-sm text-slate-500">{{ t('luckyWheel.adminNoData') }}</div>
              <div v-else class="mt-4 space-y-3">
                <div v-for="item in multiplierDistribution" :key="item.multiplier" data-test="distribution-item" class="space-y-1">
                  <div class="flex items-center justify-between gap-3 text-sm">
                    <span class="font-semibold text-slate-900">{{ item.multiplier.toFixed(1) }}x</span>
                    <span class="text-slate-500">{{ item.draw_count }}</span>
                  </div>
                  <div class="h-3 overflow-hidden rounded-full bg-slate-100">
                    <div class="h-full rounded-full bg-gradient-to-r from-amber-400 to-rose-500" :style="{ width: `${item.percent}%` }" />
                  </div>
                </div>
              </div>
            </div>

            <div class="rounded-[28px] border border-slate-200 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
              <h2 class="mb-4 text-lg font-bold text-slate-900">{{ t('luckyWheel.historyTitle') }}</h2>
              <div v-if="!stats?.recent_sessions?.length" class="py-8 text-center text-sm text-slate-500">{{ t('luckyWheel.adminNoData') }}</div>
              <div v-else data-test="settlement-history-list" class="max-h-[440px] space-y-3 overflow-y-auto pr-2">
                <div
                  v-for="session in stats.recent_sessions"
                  :key="session.id"
                  data-test="settlement-history-item"
                  class="rounded-2xl border border-slate-100 bg-slate-50/70 p-4"
                >
                  <div class="flex items-center justify-between gap-3">
                    <div>
                      <div class="font-semibold text-slate-900">#{{ session.user_id }} · {{ session.matched_tier_name }}</div>
                      <div class="mt-1 text-xs text-slate-500">
                        {{ session.source_order_type }} · {{ session.source_pay_amount.toFixed(2) }} · {{ session.completed_draws }}/{{ session.total_draws }}
                      </div>
                    </div>
                    <div class="text-right">
                      <div class="font-black text-emerald-500">{{ session.best_multiplier.toFixed(1) }}x</div>
                      <div class="text-xs text-slate-400">{{ formatDateTime(session.updated_at) }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Toggle from '@/components/common/Toggle.vue'
import { adminPaymentAPI } from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import type { LuckyWheelConfig, LuckyWheelStats } from '@/types/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const enabled = ref(false)
const stats = ref<LuckyWheelStats | null>(null)
const config = reactive<LuckyWheelConfig>({
  eligible_order_types: ['balance', 'subscription'],
  multiplier_step: 0.1,
  global_max_multiplier: 3,
  intro_text: '',
  rules_title: '',
  rules_items: [],
  prizes: [],
  tiers: [],
  amount_tiers: [],
  invite_bonus: {
    enabled: true,
    qualifying_amount: 20,
    bonus_per_invitee: 0.2,
    max_bonus: 1,
    consume_policy: 'next_session_once',
  },
  golden_window: {
    enabled: true,
    timezone: 'Asia/Shanghai',
    start_time: '20:00',
    end_time: '22:00',
    min_amount: 51,
    extra_draws: 1,
    daily_quota: 5,
  },
})

const eligibleOrderTypeSet = reactive({
  balance: true,
  subscription: true,
})

watch(
  () => [eligibleOrderTypeSet.balance, eligibleOrderTypeSet.subscription],
  () => {
    const orderTypes: Array<'balance' | 'subscription'> = []
    if (eligibleOrderTypeSet.balance) orderTypes.push('balance')
    if (eligibleOrderTypeSet.subscription) orderTypes.push('subscription')
    config.eligible_order_types = orderTypes
  },
  { deep: true },
)

const validationIssues = computed(() => {
  const issues: string[] = []
  if (config.multiplier_step <= 0) issues.push(t('luckyWheel.adminValidationStep'))
  if (config.global_max_multiplier <= 0) issues.push(t('luckyWheel.adminValidationCap'))
  if (config.eligible_order_types.length === 0) issues.push(t('luckyWheel.adminValidationOrderTypes'))
  if (config.amount_tiers.length === 0) issues.push(t('luckyWheel.adminValidationTiers'))

  const sorted = [...config.amount_tiers].sort((a, b) => a.min_amount - b.min_amount)
  for (let index = 0; index < sorted.length; index += 1) {
    const tier = sorted[index]
    if (!tier.id?.trim()) issues.push(t('luckyWheel.adminValidationTierId'))
    if (tier.draw_count <= 0) issues.push(t('luckyWheel.adminValidationDrawCount'))
    if (tier.min_multiplier <= 0 || tier.max_multiplier < tier.min_multiplier) issues.push(t('luckyWheel.adminValidationMultiplierRange'))
    if (tier.max_multiplier > config.global_max_multiplier) issues.push(t('luckyWheel.adminValidationMultiplierCap'))
    if (tier.max_amount != null && tier.max_amount < tier.min_amount) issues.push(t('luckyWheel.adminValidationAmountRange'))
    if (index > 0) {
      const previous = sorted[index - 1]
      const previousMax = previous.max_amount ?? Number.POSITIVE_INFINITY
      if (tier.min_amount <= previousMax) issues.push(t('luckyWheel.adminValidationOverlap'))
    }
    if (index < sorted.length - 1 && tier.max_amount == null) issues.push(t('luckyWheel.adminValidationOpenTier'))
  }

  if (config.invite_bonus.enabled && config.invite_bonus.bonus_per_invitee <= 0) issues.push(t('luckyWheel.adminValidationInviteBonus'))
  if (config.golden_window.enabled && config.golden_window.daily_quota <= 0) issues.push(t('luckyWheel.adminValidationGoldenQuota'))
  return [...new Set(issues)]
})

const multiplierDistribution = computed(() => {
  const rows = stats.value?.multiplier_stats ?? []
  const maxDrawCount = rows.reduce((currentMax, item) => Math.max(currentMax, item.draw_count), 0)
  return rows.map((item) => ({
    ...item,
    percent: maxDrawCount > 0 ? Math.max(8, (item.draw_count / maxDrawCount) * 100) : 0,
  }))
})

function applyConfig(next: LuckyWheelConfig) {
  config.eligible_order_types = [...(next.eligible_order_types ?? [])]
  config.multiplier_step = next.multiplier_step
  config.global_max_multiplier = next.global_max_multiplier
  config.intro_text = next.intro_text ?? ''
  config.rules_title = next.rules_title ?? ''
  config.rules_items = [...(next.rules_items ?? [])]
  config.amount_tiers = (next.amount_tiers ?? []).map((tier) => ({ ...tier }))
  config.invite_bonus = { ...next.invite_bonus }
  config.golden_window = { ...next.golden_window }
  eligibleOrderTypeSet.balance = config.eligible_order_types.includes('balance')
  eligibleOrderTypeSet.subscription = config.eligible_order_types.includes('subscription')
}

async function loadAll() {
  loading.value = true
  try {
    const [configResp, statsResp] = await Promise.all([
      adminPaymentAPI.getLuckyWheelConfig(),
      adminPaymentAPI.getLuckyWheelStats(),
    ])
    enabled.value = configResp.data.enabled
    applyConfig(configResp.data.config)
    stats.value = statsResp.data
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  if (validationIssues.value.length > 0) {
    appStore.showError(validationIssues.value[0] || t('common.error'))
    return
  }
  saving.value = true
  try {
    await adminPaymentAPI.updateLuckyWheelConfig({
      enabled: enabled.value,
      config: JSON.parse(JSON.stringify(config)) as LuckyWheelConfig,
    })
    appStore.showSuccess(t('common.saveSuccess'))
    await loadAll()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}

function addTier() {
  config.amount_tiers.push({
    id: `tier_${config.amount_tiers.length + 1}`,
    name: '',
    min_amount: 0,
    max_amount: null,
    min_multiplier: 1,
    max_multiplier: 1,
    draw_count: 1,
  })
}

function removeTier(index: number) {
  config.amount_tiers.splice(index, 1)
}

function addRuleItem() {
  config.rules_items.push('')
}

function removeRuleItem(index: number) {
  config.rules_items.splice(index, 1)
}

function updateTierMax(index: number, value: string) {
  config.amount_tiers[index].max_amount = value.trim() === '' ? null : Number(value)
}

function formatDateTime(value?: string) {
  if (!value) return '--'
  return new Date(value).toLocaleString()
}

onMounted(() => {
  loadAll()
})
</script>
