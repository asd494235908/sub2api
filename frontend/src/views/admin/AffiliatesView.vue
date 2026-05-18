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

        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <input
            v-model="inviterState.search"
            type="text"
            class="input sm:max-w-md"
            :placeholder="t('admin.affiliates.searchPlaceholder')"
            @input="onInviterSearchInput"
          />
          <button
            type="button"
            class="btn btn-primary btn-sm inline-flex items-center justify-center gap-2"
            @click="openManualDialog"
          >
            <Icon name="userPlus" size="sm" :stroke-width="2" />
            {{ t('admin.affiliates.manualAdd') }}
          </button>
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

      <BaseDialog
        :show="manualState.show"
        :title="t('admin.affiliates.manualDialogTitle')"
        width="normal"
        @close="closeManualDialog"
      >
        <div class="min-h-[520px] space-y-5 sm:min-h-[560px]">
          <p class="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-200">
            {{ t('admin.affiliates.manualHint') }}
          </p>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('admin.affiliates.manualInviter') }}</label>
              <div class="relative">
                <input
                  v-model="manualState.inviterQuery"
                  type="text"
                  class="input"
                  :placeholder="t('admin.affiliates.manualSearchPlaceholder')"
                  @input="onManualUserSearch('inviter')"
                  @focus="manualState.inviterDropdown = true"
                />
                <div
                  v-if="manualState.inviterDropdown && (manualState.inviterLoading || manualState.inviterResults.length > 0 || manualState.inviterQuery)"
                  class="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                >
                  <div v-if="manualState.inviterLoading" class="px-3 py-2 text-sm text-gray-500">
                    {{ t('common.loading') }}
                  </div>
                  <div v-else-if="manualState.inviterResults.length === 0" class="px-3 py-2 text-sm text-gray-500">
                    {{ t('admin.affiliates.manualNoOptions') }}
                  </div>
                  <button
                    v-for="user in manualState.inviterResults"
                    :key="user.id"
                    type="button"
                    class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                    @click="selectManualUser('inviter', user)"
                  >
                    <span class="block font-medium text-gray-900 dark:text-white">{{ user.email || '-' }}</span>
                    <span class="text-xs text-gray-500">#{{ user.id }} · {{ user.username || '-' }}</span>
                  </button>
                </div>
              </div>
              <div v-if="manualState.inviter" class="mt-2 rounded-md border border-gray-200 px-3 py-2 text-xs text-gray-600 dark:border-dark-700 dark:text-gray-300">
                <div class="font-medium text-gray-900 dark:text-white">{{ t('admin.affiliates.manualSelectedUser') }}</div>
                <div>#{{ manualState.inviter.id }} · {{ manualState.inviter.email || '-' }} · {{ manualState.inviter.username || '-' }}</div>
              </div>
            </div>

            <div>
              <label class="input-label">{{ t('admin.affiliates.manualInvitee') }}</label>
              <div class="relative">
                <input
                  v-model="manualState.inviteeQuery"
                  type="text"
                  class="input"
                  :placeholder="t('admin.affiliates.manualSearchPlaceholder')"
                  @input="onManualUserSearch('invitee')"
                  @focus="manualState.inviteeDropdown = true"
                />
                <div
                  v-if="manualState.inviteeDropdown && (manualState.inviteeLoading || manualState.inviteeResults.length > 0 || manualState.inviteeQuery)"
                  class="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                >
                  <div v-if="manualState.inviteeLoading" class="px-3 py-2 text-sm text-gray-500">
                    {{ t('common.loading') }}
                  </div>
                  <div v-else-if="manualState.inviteeResults.length === 0" class="px-3 py-2 text-sm text-gray-500">
                    {{ t('admin.affiliates.manualNoOptions') }}
                  </div>
                  <button
                    v-for="user in manualState.inviteeResults"
                    :key="user.id"
                    type="button"
                    class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                    @click="selectManualUser('invitee', user)"
                  >
                    <span class="block font-medium text-gray-900 dark:text-white">{{ user.email || '-' }}</span>
                    <span class="text-xs text-gray-500">#{{ user.id }} · {{ user.username || '-' }}</span>
                  </button>
                </div>
              </div>
              <div v-if="manualState.invitee" class="mt-2 rounded-md border border-gray-200 px-3 py-2 text-xs text-gray-600 dark:border-dark-700 dark:text-gray-300">
                <div class="font-medium text-gray-900 dark:text-white">{{ t('admin.affiliates.manualSelectedUser') }}</div>
                <div>#{{ manualState.invitee.id }} · {{ manualState.invitee.email || '-' }} · {{ manualState.invitee.username || '-' }}</div>
              </div>
            </div>
          </div>
        </div>

        <template #footer>
          <button type="button" class="btn btn-secondary" :disabled="manualState.submitting" @click="closeManualDialog">
            {{ t('common.cancel') }}
          </button>
          <button type="button" class="btn btn-primary" :disabled="manualState.submitting" @click="submitManualRelation">
            {{ manualState.submitting ? t('admin.affiliates.manualSubmitting') : t('admin.affiliates.manualSubmit') }}
          </button>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { affiliatesAPI, type AffiliateInvitee, type AffiliateInviterEntry, type SimpleUser } from '@/api/admin/affiliates'
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

type ManualUserRole = 'inviter' | 'invitee'

const manualState = reactive<{
  show: boolean
  submitting: boolean
  inviter: SimpleUser | null
  invitee: SimpleUser | null
  inviterQuery: string
  inviteeQuery: string
  inviterResults: SimpleUser[]
  inviteeResults: SimpleUser[]
  inviterLoading: boolean
  inviteeLoading: boolean
  inviterDropdown: boolean
  inviteeDropdown: boolean
  inviterTimer: number | null
  inviteeTimer: number | null
}>({
  show: false,
  submitting: false,
  inviter: null,
  invitee: null,
  inviterQuery: '',
  inviteeQuery: '',
  inviterResults: [],
  inviteeResults: [],
  inviterLoading: false,
  inviteeLoading: false,
  inviterDropdown: false,
  inviteeDropdown: false,
  inviterTimer: null,
  inviteeTimer: null,
})

function debounceTimer(slot: { searchTimer: number | null }, delayMs: number, run: () => void) {
  if (slot.searchTimer != null) window.clearTimeout(slot.searchTimer)
  slot.searchTimer = window.setTimeout(run, delayMs)
}

function openManualDialog() {
  manualState.show = true
}

function resetManualDialog() {
  manualState.submitting = false
  manualState.inviter = null
  manualState.invitee = null
  manualState.inviterQuery = ''
  manualState.inviteeQuery = ''
  manualState.inviterResults = []
  manualState.inviteeResults = []
  manualState.inviterLoading = false
  manualState.inviteeLoading = false
  manualState.inviterDropdown = false
  manualState.inviteeDropdown = false
  if (manualState.inviterTimer != null) window.clearTimeout(manualState.inviterTimer)
  if (manualState.inviteeTimer != null) window.clearTimeout(manualState.inviteeTimer)
  manualState.inviterTimer = null
  manualState.inviteeTimer = null
}

function closeManualDialog() {
  manualState.show = false
  resetManualDialog()
}

function onManualUserSearch(role: ManualUserRole) {
  const timerKey = role === 'inviter' ? 'inviterTimer' : 'inviteeTimer'
  if (manualState[timerKey] != null) window.clearTimeout(manualState[timerKey])
  manualState[timerKey] = window.setTimeout(() => {
    void searchManualUsers(role)
  }, 300)
}

async function searchManualUsers(role: ManualUserRole) {
  const queryKey = role === 'inviter' ? 'inviterQuery' : 'inviteeQuery'
  const loadingKey = role === 'inviter' ? 'inviterLoading' : 'inviteeLoading'
  const resultsKey = role === 'inviter' ? 'inviterResults' : 'inviteeResults'
  const selectedKey = role
  const query = manualState[queryKey].trim()
  if (manualState[selectedKey] && query !== manualState[selectedKey]?.email) {
    manualState[selectedKey] = null
  }
  if (!query) {
    manualState[resultsKey] = []
    return
  }
  manualState[loadingKey] = true
  try {
    manualState[resultsKey] = await affiliatesAPI.lookupUsers(query)
  } catch {
    manualState[resultsKey] = []
  } finally {
    manualState[loadingKey] = false
  }
}

function selectManualUser(role: ManualUserRole, user: SimpleUser) {
  if (role === 'inviter') {
    manualState.inviter = user
    manualState.inviterQuery = user.email || user.username || String(user.id)
    manualState.inviterDropdown = false
  } else {
    manualState.invitee = user
    manualState.inviteeQuery = user.email || user.username || String(user.id)
    manualState.inviteeDropdown = false
  }
}

async function submitManualRelation() {
  if (!manualState.inviter || !manualState.invitee) {
    appStore.showError(t('admin.affiliates.manualMissingUserError'))
    return
  }
  if (manualState.inviter.id === manualState.invitee.id) {
    appStore.showError(t('admin.affiliates.manualSameUserError'))
    return
  }
  manualState.submitting = true
  try {
    await affiliatesAPI.createInviteRelation({
      inviter_user_id: manualState.inviter.id,
      invitee_user_id: manualState.invitee.id,
      overwrite: true,
    })
    appStore.showSuccess(t('admin.affiliates.manualSuccess'))
    manualState.show = false
    resetManualDialog()
    await loadInviters()
    if (inviteesState.inviter) {
      inviteesState.items = await affiliatesAPI.listInviterInvitees(inviteesState.inviter.user_id)
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    manualState.submitting = false
  }
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
