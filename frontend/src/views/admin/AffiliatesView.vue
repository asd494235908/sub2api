<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="mb-4">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.affiliates.title') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.affiliates.description') }}
          </p>
        </div>

        <div class="mb-4">
          <input
            v-model="inviterState.search"
            type="text"
            class="input"
            :placeholder="t('admin.affiliates.searchPlaceholder')"
            @input="onInviterSearchInput"
          />
        </div>

        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.email') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.username') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.code') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.invitedCount') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.totalRebate') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.col.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="inviterState.loading">
                <td colspan="6" class="px-3 py-6 text-center text-sm text-gray-500">
                  {{ t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="inviterState.entries.length === 0">
                <td colspan="6" class="px-3 py-6 text-center text-sm text-gray-500">
                  {{ t('admin.affiliates.empty') }}
                </td>
              </tr>
              <tr v-for="entry in inviterState.entries" :key="entry.user_id">
                <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{{ entry.email }}</td>
                <td class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300">{{ entry.username }}</td>
                <td class="px-3 py-2 text-sm font-mono">{{ entry.aff_code }}</td>
                <td class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300">{{ entry.aff_count }}</td>
                <td class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300">{{ formatCurrency(entry.total_rebate, 'CNY') }}</td>
                <td class="px-3 py-2 text-sm">
                  <button
                    type="button"
                    class="text-primary-600 hover:underline"
                    @click="loadInvitees(entry)"
                  >
                    {{ t('admin.affiliates.viewInvitees') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="inviterState.total > inviterState.pageSize" class="mt-4 flex items-center justify-between text-sm">
          <span class="text-gray-500">
            {{ t('admin.affiliates.totalLabel', { total: inviterState.total }) }}
          </span>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="inviterState.page <= 1"
              @click="changeInviterPage(inviterState.page - 1)"
            >
              {{ t('pagination.previous') }}
            </button>
            <span class="text-gray-500">{{ inviterState.page }} / {{ Math.max(1, Math.ceil(inviterState.total / inviterState.pageSize)) }}</span>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="inviterState.page >= Math.ceil(inviterState.total / inviterState.pageSize)"
              @click="changeInviterPage(inviterState.page + 1)"
            >
              {{ t('pagination.next') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="inviteesState.inviter" class="card p-6">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.affiliates.inviteesTitle', { email: inviteesState.inviter.email || inviteesState.inviter.username }) }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.affiliates.inviteesDescription') }}
            </p>
          </div>
          <button type="button" class="btn btn-secondary btn-sm" @click="clearInvitees">
            {{ t('common.close') }}
          </button>
        </div>

        <div v-if="inviteesState.loading" class="py-4 text-sm text-gray-500">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="inviteesState.items.length === 0" class="py-4 text-sm text-gray-500">
          {{ t('admin.affiliates.inviteesEmpty') }}
        </div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.inviteesCol.email') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.inviteesCol.username') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.inviteesCol.joinedAt') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.affiliates.inviteesCol.totalRebate') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-for="item in inviteesState.items" :key="item.user_id">
                <td class="px-3 py-2 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                <td class="px-3 py-2 text-gray-600 dark:text-gray-300">{{ item.username || '-' }}</td>
                <td class="px-3 py-2 text-gray-600 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                <td class="px-3 py-2 text-gray-600 dark:text-gray-300">{{ formatCurrency(item.total_rebate, 'CNY') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { affiliatesAPI, type AffiliateInvitee, type AffiliateInviterEntry } from '@/api/admin/affiliates'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatCurrency, formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const inviterState = reactive<{
  loading: boolean
  entries: AffiliateInviterEntry[]
  total: number
  page: number
  pageSize: number
  search: string
  searchTimer: number | null
}>({
  loading: false,
  entries: [],
  total: 0,
  page: 1,
  pageSize: 20,
  search: '',
  searchTimer: null,
})

const inviteesState = reactive<{
  loading: boolean
  inviter: AffiliateInviterEntry | null
  items: AffiliateInvitee[]
}>({
  loading: false,
  inviter: null,
  items: [],
})

function debounceTimer(slot: { searchTimer: number | null }, delayMs: number, run: () => void) {
  if (slot.searchTimer != null) window.clearTimeout(slot.searchTimer)
  slot.searchTimer = window.setTimeout(run, delayMs)
}

async function loadInviters() {
  inviterState.loading = true
  try {
    const res = await affiliatesAPI.listInviters({
      page: inviterState.page,
      page_size: inviterState.pageSize,
      search: inviterState.search,
    })
    inviterState.entries = res.items ?? []
    inviterState.total = res.total ?? 0
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    inviterState.loading = false
  }
}

function onInviterSearchInput() {
  debounceTimer(inviterState, 300, () => {
    inviterState.page = 1
    void loadInviters()
  })
}

function changeInviterPage(page: number) {
  if (page < 1) return
  inviterState.page = page
  void loadInviters()
}

async function loadInvitees(entry: AffiliateInviterEntry) {
  inviteesState.inviter = entry
  inviteesState.loading = true
  inviteesState.items = []
  try {
    inviteesState.items = await affiliatesAPI.listInviterInvitees(entry.user_id)
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    inviteesState.loading = false
  }
}

function clearInvitees() {
  inviteesState.loading = false
  inviteesState.inviter = null
  inviteesState.items = []
}

onMounted(() => {
  void loadInviters()
})
</script>
