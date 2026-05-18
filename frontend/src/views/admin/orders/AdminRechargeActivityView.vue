<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6 p-4 sm:p-6">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-950 dark:text-white">{{ t('nav.rechargeActivityAdmin') }}</h1>
          <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">{{ t('rechargeActivity.adminDescription') }}</p>
        </div>
        <button
          class="btn btn-primary"
          data-test="save-config"
          type="button"
          :disabled="saving || validationIssues.length > 0"
          @click="saveConfig"
        >
          {{ saving ? t('common.loading') : t('common.save') }}
        </button>
      </div>

      <div v-if="loading" class="flex min-h-[320px] items-center justify-center">
        <LoadingSpinner />
      </div>

      <template v-else>
        <section class="grid gap-4 md:grid-cols-4">
          <div class="rounded-lg border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminTotalChances') }}</div>
            <div class="mt-2 text-2xl font-black text-slate-950 dark:text-white">{{ stats?.total_chances ?? 0 }}</div>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminPendingChances') }}</div>
            <div class="mt-2 text-2xl font-black text-slate-950 dark:text-white">{{ stats?.pending_chances ?? 0 }}</div>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminDrawnChances') }}</div>
            <div class="mt-2 text-2xl font-black text-slate-950 dark:text-white">{{ stats?.drawn_chances ?? 0 }}</div>
          </div>
          <div class="rounded-lg border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminPendingFulfillments') }}</div>
            <div class="mt-2 text-2xl font-black text-amber-600 dark:text-amber-400">{{ stats?.pending_fulfillments ?? 0 }}</div>
          </div>
        </section>

        <section v-if="validationIssues.length" class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200">
          <div class="font-bold">{{ t('rechargeActivity.adminValidationTitle') }}</div>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li v-for="issue in validationIssues" :key="issue">{{ issue }}</li>
          </ul>
        </section>

        <section class="grid gap-6 lg:grid-cols-[360px,minmax(0,1fr)]">
          <div class="space-y-6">
            <div class="rounded-lg border border-slate-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.adminEnabledTitle') }}</h2>
                  <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminEnabledHint') }}</p>
                </div>
                <Toggle v-model="enabled" />
              </div>
            </div>

            <div class="rounded-lg border border-slate-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.adminOrderTypesTitle') }}</h2>
              <div class="mt-4 space-y-3">
                <label class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-200">
                  <input v-model="eligibleOrderTypeSet.balance" type="checkbox" />
                  {{ t('rechargeActivity.balanceOrder') }}
                </label>
                <label class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-200">
                  <input v-model="eligibleOrderTypeSet.subscription" type="checkbox" />
                  {{ t('rechargeActivity.subscriptionOrder') }}
                </label>
              </div>
            </div>

            <div class="rounded-lg border border-slate-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.adminCopyTitle') }}</h2>
              <div class="mt-4 space-y-4">
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
                  {{ t('rechargeActivity.adminIntroTextLabel') }}
                  <textarea v-model="config.intro_text" class="input mt-2 min-h-[96px]" data-test="intro-text" />
                </label>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
                  {{ t('rechargeActivity.adminRulesTitleLabel') }}
                  <input v-model="config.rules_title" class="input mt-2" data-test="rules-title" />
                </label>
                <div class="space-y-2">
                  <label class="text-sm font-medium text-slate-700 dark:text-slate-200">{{ t('rechargeActivity.adminRulesItemsLabel') }}</label>
                  <div v-for="(_, index) in config.rules_items" :key="index" class="flex gap-2">
                    <input v-model="config.rules_items[index]" class="input" data-test="rules-item" />
                    <button class="btn btn-secondary btn-sm" type="button" @click="removeRuleItem(index)">{{ t('common.delete') }}</button>
                  </div>
                  <button class="btn btn-secondary btn-sm" type="button" @click="addRuleItem">{{ t('rechargeActivity.adminAddRuleItem') }}</button>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-lg border border-slate-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.adminPrizesTitle') }}</h2>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminPrizesHint') }}</p>
              </div>
              <button class="btn btn-secondary btn-sm" data-test="add-prize" type="button" @click="addPrize">{{ t('rechargeActivity.adminAddPrize') }}</button>
            </div>

            <div class="mt-5 space-y-4">
              <div
                v-for="(prize, index) in config.prizes"
                :key="`${prize.id}-${index}`"
                class="rounded-lg border border-slate-100 p-4 dark:border-dark-700"
                data-test="recharge-prize-row"
              >
                <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
                  <label class="block text-xs font-medium text-slate-500 dark:text-slate-400">
                    {{ t('rechargeActivity.adminPrizeIdLabel') }}
                    <input v-model="prize.id" class="input mt-1" data-test="prize-id" />
                  </label>
                  <label class="block text-xs font-medium text-slate-500 dark:text-slate-400">
                    {{ t('rechargeActivity.adminPrizeNameLabel') }}
                    <input v-model="prize.name" class="input mt-1" data-test="prize-name" />
                  </label>
                  <label class="block text-xs font-medium text-slate-500 dark:text-slate-400">
                    {{ t('rechargeActivity.adminProbabilityLabel') }}
                    <input v-model.number="prize.probability" class="input mt-1" data-test="prize-probability" min="0.01" step="0.01" type="number" />
                  </label>
                  <label class="block text-xs font-medium text-slate-500 dark:text-slate-400">
                    {{ t('rechargeActivity.adminMinPayLabel') }}
                    <input v-model.number="prize.min_pay_amount" class="input mt-1" data-test="prize-min-pay" min="0" step="0.01" type="number" />
                  </label>
                  <div class="flex items-end gap-3">
                    <label class="flex items-center gap-2 pb-2 text-sm text-slate-700 dark:text-slate-200">
                      <input v-model="prize.enabled" type="checkbox" />
                      {{ t('common.enabled') }}
                    </label>
                    <button class="btn btn-secondary btn-sm mb-1" type="button" @click="removePrize(index)">{{ t('common.delete') }}</button>
                  </div>
                </div>
                <label class="mt-3 block text-xs font-medium text-slate-500 dark:text-slate-400">
                  {{ t('rechargeActivity.adminRewardDescriptionLabel') }}
                  <textarea v-model="prize.reward_description" class="input mt-1 min-h-[72px]" data-test="prize-reward-description" />
                </label>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-lg border border-slate-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.adminRecentRecordsTitle') }}</h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.adminRecentRecordsHint') }}</p>
            </div>
            <div class="flex flex-wrap items-center gap-3">
              <input
                v-model="recordSearch"
                class="input w-64 max-w-full"
                data-test="recharge-record-search"
                :placeholder="t('rechargeActivity.adminSearchUserPlaceholder')"
                @input="scheduleRecordSearch"
              />
              <div class="text-sm text-slate-500 dark:text-slate-400">
                {{ t('rechargeActivity.adminFulfilledRecords') }} {{ stats?.fulfilled_records ?? 0 }}
              </div>
            </div>
          </div>

          <div v-if="stats?.recent_records?.length" class="mt-4 overflow-x-auto">
            <table class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-400">
                <tr>
                  <th class="px-3 py-2">{{ t('rechargeActivity.adminWinningUser') }}</th>
                  <th class="px-3 py-2">{{ t('rechargeActivity.adminPrizeNameLabel') }}</th>
                  <th class="px-3 py-2">{{ t('rechargeActivity.adminRewardDescriptionLabel') }}</th>
                  <th class="px-3 py-2">{{ t('rechargeActivity.sourceOrder') }}</th>
                  <th class="px-3 py-2">{{ t('rechargeActivity.adminFulfillmentStatus') }}</th>
                  <th class="px-3 py-2">{{ t('rechargeActivity.adminFulfillmentNote') }}</th>
                  <th class="px-3 py-2 text-right">{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="record in stats.recent_records" :key="record.id" class="border-t border-slate-100 dark:border-dark-700" data-test="recharge-record-row">
                  <td class="px-3 py-3">
                    <div class="font-medium text-slate-900 dark:text-white">{{ record.user_email || `#${record.user_id}` }}</div>
                    <div v-if="record.user_name" class="mt-0.5 text-xs text-slate-500 dark:text-slate-400">{{ record.user_name }}</div>
                  </td>
                  <td class="px-3 py-3 font-semibold text-slate-900 dark:text-white">{{ record.prize_name }}</td>
                  <td class="max-w-[18rem] px-3 py-3 text-slate-600 dark:text-slate-300">{{ record.reward_description || '-' }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">#{{ record.source_order_id }}</td>
                  <td class="px-3 py-3">
                    <span class="rounded-full px-2 py-1 text-xs font-semibold" :class="fulfillmentBadgeClass(record.fulfillment_status)">
                      {{ fulfillmentStatusLabel(record.fulfillment_status) }}
                    </span>
                  </td>
                  <td class="px-3 py-3">
                    <input v-model="fulfillmentNotes[record.id]" class="input min-w-[180px]" data-test="fulfillment-note" />
                  </td>
                  <td class="px-3 py-3 text-right">
                    <button
                      class="btn btn-secondary btn-sm"
                      data-test="toggle-fulfillment"
                      type="button"
                      :disabled="updatingRecordId === record.id"
                      @click="toggleFulfillment(record)"
                    >
                      {{ record.fulfillment_status === 'fulfilled' ? t('rechargeActivity.markPending') : t('rechargeActivity.markFulfilled') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="mt-4 text-sm text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.noRecords') }}</div>
          <Pagination
            v-if="(stats?.recent_records_total ?? 0) > 0"
            class="mt-4"
            :total="stats?.recent_records_total ?? 0"
            :page="recordPagination.page"
            :page-size="recordPagination.page_size"
            @update:page="handleRecordPageChange"
            @update:page-size="handleRecordPageSizeChange"
          />
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import Toggle from '@/components/common/Toggle.vue'
import { adminPaymentAPI } from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { RechargeActivityConfig, RechargeActivityDrawRecord, RechargeActivityStats } from '@/types/payment'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const updatingRecordId = ref<number | null>(null)
const enabled = ref(false)
const stats = ref<RechargeActivityStats | null>(null)
const fulfillmentNotes = reactive<Record<number, string>>({})
const recordSearch = ref('')
const recordPagination = reactive({
  page: 1,
  page_size: 20,
})
let recordSearchTimer: ReturnType<typeof setTimeout> | null = null
const config = reactive<RechargeActivityConfig>({
  eligible_order_types: ['balance', 'subscription'],
  intro_text: '',
  rules_title: '',
  rules_items: [],
  prizes: [],
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
  if (config.eligible_order_types.length === 0) issues.push(t('rechargeActivity.adminValidationOrderTypes'))
  if (config.prizes.length === 0) issues.push(t('rechargeActivity.adminValidationPrizes'))
  const seen = new Set<string>()
  for (const prize of config.prizes) {
    if (!prize.id?.trim()) issues.push(t('rechargeActivity.adminValidationPrizeId'))
    if (seen.has(prize.id)) issues.push(t('rechargeActivity.adminValidationPrizeDuplicate'))
    seen.add(prize.id)
    if (prize.probability <= 0) issues.push(t('rechargeActivity.adminValidationProbability'))
    if (prize.min_pay_amount < 0) issues.push(t('rechargeActivity.adminValidationMinPay'))
  }
  if (!config.prizes.some((prize) => prize.enabled)) issues.push(t('rechargeActivity.adminValidationEnabledPrize'))
  const enabledProbabilityTotal = config.prizes
    .filter((prize) => prize.enabled)
    .reduce((total, prize) => total + (Number(prize.probability) || 0), 0)
  if (Math.abs(enabledProbabilityTotal - 100) > 0.000001) issues.push(t('rechargeActivity.adminValidationProbabilityTotal'))
  return [...new Set(issues)]
})

function applyConfig(next: RechargeActivityConfig) {
  config.eligible_order_types = [...(next.eligible_order_types ?? [])]
  config.intro_text = next.intro_text ?? ''
  config.rules_title = next.rules_title ?? ''
  config.rules_items = [...(next.rules_items ?? [])]
  config.prizes = (next.prizes ?? []).map((prize) => ({ ...prize, reward_amount: 0, reward_description: prize.reward_description ?? '' }))
  eligibleOrderTypeSet.balance = config.eligible_order_types.includes('balance')
  eligibleOrderTypeSet.subscription = config.eligible_order_types.includes('subscription')
}

async function loadAll() {
  loading.value = true
  try {
    const configResp = await adminPaymentAPI.getRechargeActivityConfig()
    enabled.value = configResp.data.enabled
    applyConfig(configResp.data.config)
    await loadStats()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  const keyword = recordSearch.value.trim()
  const statsResp = await adminPaymentAPI.getRechargeActivityStats({
    page: recordPagination.page,
    page_size: recordPagination.page_size,
    user_keyword: keyword || undefined,
  })
  stats.value = statsResp.data
  recordPagination.page = statsResp.data.recent_records_page || recordPagination.page
  recordPagination.page_size = statsResp.data.recent_records_page_size || recordPagination.page_size
  for (const record of statsResp.data.recent_records ?? []) {
    fulfillmentNotes[record.id] = record.fulfillment_note ?? ''
  }
}

function scheduleRecordSearch() {
  if (recordSearchTimer) {
    clearTimeout(recordSearchTimer)
  }
  recordSearchTimer = setTimeout(() => {
    recordPagination.page = 1
    loadStats().catch((err) => {
      appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
    })
  }, 300)
}

async function handleRecordPageChange(page: number) {
  recordPagination.page = page
  try {
    await loadStats()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function handleRecordPageSizeChange(pageSize: number) {
  recordPagination.page = 1
  recordPagination.page_size = pageSize
  try {
    await loadStats()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function saveConfig() {
  if (validationIssues.value.length > 0) {
    appStore.showError(validationIssues.value[0] || t('common.error'))
    return
  }
  saving.value = true
  try {
    await adminPaymentAPI.updateRechargeActivityConfig({
      enabled: enabled.value,
      config: JSON.parse(JSON.stringify(config)) as RechargeActivityConfig,
    })
    appStore.showSuccess(t('common.saveSuccess'))
    await loadAll()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}

function addPrize() {
  config.prizes.push({
    id: `prize_${config.prizes.length + 1}`,
    name: '',
    reward_amount: 0,
    reward_description: '',
    probability: 1,
    min_pay_amount: 0,
    enabled: true,
    sort_order: config.prizes.length + 1,
  })
}

function removePrize(index: number) {
  config.prizes.splice(index, 1)
}

function addRuleItem() {
  config.rules_items.push('')
}

function removeRuleItem(index: number) {
  config.rules_items.splice(index, 1)
}

function fulfillmentStatusLabel(value?: string) {
  return value === 'fulfilled' ? t('rechargeActivity.fulfillmentFulfilled') : t('rechargeActivity.fulfillmentPending')
}

function fulfillmentBadgeClass(value?: string) {
  return value === 'fulfilled'
    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
    : 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
}

async function toggleFulfillment(record: RechargeActivityDrawRecord) {
  updatingRecordId.value = record.id
  try {
    const nextStatus = record.fulfillment_status === 'fulfilled' ? 'pending' : 'fulfilled'
    const note = (fulfillmentNotes[record.id] ?? '').trim()
    if (nextStatus === 'fulfilled' && !note) {
      appStore.showError(t('rechargeActivity.adminFulfillmentNoteRequired'))
      return
    }
    await adminPaymentAPI.updateRechargeActivityRecordFulfillment(record.id, {
      status: nextStatus,
      note,
    })
    appStore.showSuccess(t('common.saveSuccess'))
    await loadStats()
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    updatingRecordId.value = null
  }
}

onMounted(() => {
  loadAll()
})

onBeforeUnmount(() => {
  if (recordSearchTimer) {
    clearTimeout(recordSearchTimer)
  }
})
</script>
