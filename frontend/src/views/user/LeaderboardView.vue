<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('leaderboard.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('leaderboard.description') }}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <div class="inline-flex rounded-md border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-800">
            <button
              v-for="option in periodOptions"
              :key="option"
              type="button"
              class="rounded px-3 py-1.5 text-sm font-medium transition-colors"
              :class="period === option ? 'bg-primary-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
              @click="setPeriod(option)"
            >
              {{ t(`leaderboard.periods.${option}`) }}
            </button>
          </div>
          <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadLeaderboard">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            <span class="ml-2">{{ t('common.refresh') }}</span>
          </button>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.summary.period') }}</p>
          <p class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
            {{ data?.start_date || '-' }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.summary.tokens') }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatTokens(data?.total_tokens || 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.summary.requests') }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ (data?.total_requests || 0).toLocaleString() }}
          </p>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('leaderboard.tableTitle') }}
          </h2>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-16">
          <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
        </div>
        <div v-else-if="error" class="px-6 py-12 text-center">
          <Icon name="exclamationCircle" size="lg" class="mx-auto text-red-500" />
          <p class="mt-3 text-sm text-gray-600 dark:text-gray-300">{{ error }}</p>
        </div>
        <div v-else-if="!data?.ranking.length" class="px-6 py-12 text-center">
          <Icon name="chart" size="lg" class="mx-auto text-gray-400" />
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
            {{ t('leaderboard.empty') }}
          </p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('leaderboard.columns.rank') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('leaderboard.columns.user') }}</th>
                <th class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('leaderboard.columns.tokens') }}</th>
                <th class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('leaderboard.columns.requests') }}</th>
                <th class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('leaderboard.columns.actualCost') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-for="item in data.ranking" :key="item.user_id">
                <td class="whitespace-nowrap px-6 py-4">
                  <span class="inline-flex h-7 min-w-7 items-center justify-center rounded bg-gray-100 px-2 text-sm font-semibold text-gray-700 dark:bg-dark-700 dark:text-gray-200">
                    {{ item.rank }}
                  </span>
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900 dark:text-white">
                  {{ item.display_name }}
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-right text-sm font-semibold text-gray-900 dark:text-white">
                  {{ formatTokens(item.tokens) }}
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-600 dark:text-gray-300">
                  {{ item.requests.toLocaleString() }}
                </td>
                <td class="whitespace-nowrap px-6 py-4 text-right text-sm text-gray-600 dark:text-gray-300">
                  ¥{{ item.actual_cost.toFixed(4) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { usageAPI } from '@/api/usage'
import type { LeaderboardPeriod, UsageLeaderboardResponse } from '@/api/usage'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()

const periodOptions: LeaderboardPeriod[] = ['today', 'yesterday']
const period = ref<LeaderboardPeriod>('today')
const data = ref<UsageLeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref('')

function formatTokens(value: number): string {
  return Math.round(value).toLocaleString()
}

async function loadLeaderboard(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    data.value = await usageAPI.getLeaderboard({ period: period.value, limit: 20 })
  } catch (err: unknown) {
    error.value = extractApiErrorMessage(err, t('leaderboard.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function setPeriod(nextPeriod: LeaderboardPeriod): void {
  if (period.value === nextPeriod) return
  period.value = nextPeriod
  void loadLeaderboard()
}

onMounted(() => {
  void loadLeaderboard()
})
</script>
