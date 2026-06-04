<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="dollar" size="sm" class="text-primary-500" />
              {{ t('affiliate.stats.rebateRate') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-primary-600 dark:text-primary-400">
              {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.stats.rebateRateHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCount(detail.aff_count) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(rebateCashBalance) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCurrency(totalRebateCash) }}
            </p>
            <p v-if="frozenRebateCash > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(frozenRebateCash) }}
            </p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.description') }}</p>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ t('affiliate.tips.title') }}</p>
            <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
              <li>1. {{ t('affiliate.tips.line1') }}</li>
              <li>2. {{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
              <li v-if="frozenRebateCash > 0">4. {{ t('affiliate.tips.line4') }}</li>
            </ul>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.transfer.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button
              class="btn btn-primary"
              :disabled="transferring || rebateCashBalance <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="rebateCashBalance <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="max-w-2xl">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.withdraw.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.withdraw.description') }}</p>
            </div>
            <form
              data-test="affiliate-withdraw-form"
              class="grid w-full gap-3 sm:max-w-2xl sm:grid-cols-[minmax(120px,0.8fr)_minmax(150px,1fr)_minmax(220px,1.4fr)] lg:max-w-3xl xl:max-w-5xl xl:grid-cols-[minmax(120px,0.8fr)_minmax(150px,1fr)_minmax(220px,1.4fr)_auto] xl:items-end"
              @submit.prevent="submitWithdrawal"
            >
              <label class="space-y-1.5 text-sm font-medium text-gray-700 dark:text-gray-300">
                <span>{{ t('affiliate.withdraw.amount') }}</span>
                <input
                  v-model.number="withdrawForm.amount"
                  type="number"
                  min="0"
                  step="0.01"
                  class="input"
                  :disabled="rebateCashBalance <= 0"
                  :placeholder="t('affiliate.withdraw.amount')"
                />
              </label>
              <label class="space-y-1.5 text-sm font-medium text-gray-700 dark:text-gray-300">
                <span>{{ t('affiliate.withdraw.payoutMethod') }}</span>
                <select v-model="withdrawForm.payout_method" class="input" :disabled="rebateCashBalance <= 0">
                  <option value="wechat_manual">{{ t('affiliate.withdraw.methods.wechat_manual') }}</option>
                </select>
              </label>
              <label class="space-y-1.5 text-sm font-medium text-gray-700 dark:text-gray-300">
                <span>{{ t('affiliate.withdraw.accountNote') }}</span>
                <input
                  v-model="withdrawForm.payout_account_note"
                  type="text"
                  class="input"
                  :disabled="rebateCashBalance <= 0"
                  :placeholder="t('affiliate.withdraw.accountNote')"
                />
              </label>
              <button
                data-test="affiliate-withdraw-submit"
                class="btn btn-primary h-12 justify-center whitespace-nowrap sm:col-span-3 lg:col-span-1 lg:justify-self-end xl:col-span-1"
                type="submit"
                :disabled="withdrawing || rebateCashBalance <= 0"
              >
                <Icon v-if="withdrawing" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="dollar" size="sm" />
                <span>{{ withdrawing ? t('affiliate.withdraw.submitting') : t('affiliate.withdraw.submit') }}</span>
              </button>
            </form>
          </div>
          <p v-if="rebateCashBalance <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.withdraw.empty') }}
          </p>
          <div class="mt-5 overflow-x-auto">
            <table class="w-full min-w-[720px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdraw.columns.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdraw.columns.status') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdraw.columns.account') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdraw.columns.tradeNo') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdraw.columns.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="withdrawalsLoading">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="withdrawals.length === 0">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500">{{ t('affiliate.withdraw.noRecords') }}</td>
                </tr>
                <tr
                  v-for="item in withdrawals"
                  v-else
                  :key="item.id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 font-medium text-gray-900 dark:text-white">{{ formatCurrency(item.amount) }}</td>
                  <td class="px-3 py-3"><span :class="withdrawStatusClass(item.status)">{{ withdrawStatusLabel(item.status) }}</span></td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.payout_account_note || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.payout_trade_no || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.records.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.records.description') }}</p>
          <div class="mt-5 overflow-x-auto">
            <table class="w-full min-w-[760px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.action') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.source') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.balanceAfter') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="recordsLoading">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="records.length === 0">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500">{{ t('affiliate.records.empty') }}</td>
                </tr>
                <tr
                  v-for="item in records"
                  v-else
                  :key="item.ledger_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ recordActionLabel(item.action) }}</td>
                  <td class="px-3 py-3 font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.amount) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ recordSourceLabel(item) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatOptionalCurrency(item.rebate_cash_after) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { AffiliateCashRecord, AffiliateWithdrawal, UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const withdrawing = ref(false)
const withdrawalsLoading = ref(false)
const recordsLoading = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)
const withdrawals = ref<AffiliateWithdrawal[]>([])
const records = ref<AffiliateCashRecord[]>([])
const withdrawForm = ref({
  amount: null as number | null,
  payout_method: 'wechat_manual',
  payout_account_note: '',
})

const inviteLink = computed(() => {
  if (!detail.value) return ''
  const base = (import.meta.env.BASE_URL || '/').replace(/\/?$/, '/')
  const path = `${base}register?aff=${encodeURIComponent(detail.value.aff_code)}`
  if (typeof window === 'undefined') return path
  return `${window.location.origin}${path}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

const rebateCashBalance = computed(() => detail.value?.rebate_cash_balance ?? detail.value?.aff_quota ?? 0)
const frozenRebateCash = computed(() => detail.value?.frozen_rebate_cash ?? detail.value?.aff_frozen_quota ?? 0)
const totalRebateCash = computed(() => detail.value?.total_rebate_cash ?? detail.value?.aff_history_quota ?? 0)

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function loadWithdrawals(): Promise<void> {
  withdrawalsLoading.value = true
  try {
    const resp = await userAPI.listAffiliateWithdrawals({ page: 1, page_size: 20 })
    withdrawals.value = resp.items || []
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.withdraw.loadFailed')))
  } finally {
    withdrawalsLoading.value = false
  }
}

async function loadRecords(): Promise<void> {
  recordsLoading.value = true
  try {
    const resp = await userAPI.listAffiliateRecords({ page: 1, page_size: 20 })
    records.value = resp.items || []
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.records.loadFailed')))
  } finally {
    recordsLoading.value = false
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || rebateCashBalance.value <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      loadRecords(),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

async function submitWithdrawal(): Promise<void> {
  if (!detail.value || withdrawing.value) return
  const amount = Number(withdrawForm.value.amount || 0)
  if (amount <= 0 || amount > rebateCashBalance.value) {
    appStore.showError(t('affiliate.withdraw.invalidAmount'))
    return
  }
  if (!withdrawForm.value.payout_account_note.trim()) {
    appStore.showError(t('affiliate.withdraw.accountRequired'))
    return
  }
  withdrawing.value = true
  try {
    await userAPI.createAffiliateWithdrawal({
      amount,
      payout_method: withdrawForm.value.payout_method,
      payout_account_note: withdrawForm.value.payout_account_note.trim(),
    })
    appStore.showSuccess(t('affiliate.withdraw.success'))
    withdrawForm.value = { amount: null, payout_method: 'wechat_manual', payout_account_note: '' }
    await Promise.all([
      loadAffiliateDetail(true),
      loadRecords(),
      loadWithdrawals(),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.withdraw.failed')))
  } finally {
    withdrawing.value = false
  }
}

function withdrawStatusLabel(status: string): string {
  return t(`affiliate.withdraw.status.${status}`, status)
}

function withdrawStatusClass(status: string): string {
  if (status === 'paid') {
    return 'inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
  if (status === 'rejected' || status === 'failed') {
    return 'inline-flex rounded-full bg-rose-100 px-2 py-0.5 text-xs font-medium text-rose-700 dark:bg-rose-900/30 dark:text-rose-300'
  }
  if (status === 'approved') {
    return 'inline-flex rounded-full bg-sky-100 px-2 py-0.5 text-xs font-medium text-sky-700 dark:bg-sky-900/30 dark:text-sky-300'
  }
  return 'inline-flex rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
}

function recordActionLabel(action: string): string {
  return t(`affiliate.records.actions.${action}`, action)
}

function recordSourceLabel(item: AffiliateCashRecord): string {
  if (item.source_user_email || item.source_username) {
    return item.source_user_email || item.source_username || '-'
  }
  if (item.source_order_id) {
    return `#${item.source_order_id}`
  }
  return '-'
}

function formatOptionalCurrency(value?: number | null): string {
  return typeof value === 'number' ? formatCurrency(value) : '-'
}

onMounted(() => {
  void loadAffiliateDetail()
  void loadRecords()
  void loadWithdrawals()
})
</script>
