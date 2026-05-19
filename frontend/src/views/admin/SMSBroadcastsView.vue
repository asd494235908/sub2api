<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <button class="btn btn-secondary" :disabled="loading" @click="load">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
          <button data-test="sms-broadcast-create" class="btn btn-primary" @click="openCreateDialog">
            <Icon name="plus" size="sm" class="mr-1.5" />
            {{ t('admin.smsBroadcasts.create') }}
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="items" :loading="loading">
          <template #cell-title="{ row }">
            <div>
              <div class="font-medium text-gray-900 dark:text-gray-100">{{ campaignTitle(row) }}</div>
              <div class="mt-1 max-w-xl truncate text-xs text-gray-500 dark:text-dark-400">{{ campaignBody(row) }}</div>
            </div>
          </template>

          <template #cell-status="{ value, row }">
            <span class="badge" :class="statusClass(value || row.status)">{{ statusLabel(value || row.status) }}</span>
          </template>

          <template #cell-summary="{ row }">
            <div class="whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
              {{ countOf(row, 'total_recipients') }} /
              {{ countOf(row, 'sent_count') }} /
              {{ countOf(row, 'failed_count') }}
            </div>
          </template>

          <template #cell-created_at="{ row }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(row.created_at) || '-' }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                v-if="canCancel(row)"
                class="btn btn-secondary btn-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                :title="t('common.cancel')"
                @click="handleCancel(row)"
              >
                <Icon name="x" size="sm" class="mr-1" />
                {{ t('common.cancel') }}
              </button>
              <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
            </div>
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <BaseDialog :show="showCreate" :title="t('admin.smsBroadcasts.create')" width="wide" @close="closeCreate">
      <form id="sms-broadcast-form" class="space-y-4" @submit.prevent="submitCreate">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.smsBroadcasts.form.title') }}</label>
            <input v-model="form.title" data-test="sms-title" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.smsBroadcasts.form.templateId') }}</label>
            <input
              v-model="form.template_id"
              data-test="sms-template-id"
              class="input"
              required
              :placeholder="t('admin.smsBroadcasts.form.templateIdPlaceholder')"
            />
          </div>
        </div>

        <div>
          <div class="mb-3 flex items-center justify-between gap-3">
            <label class="input-label mb-0">{{ t('admin.smsBroadcasts.form.audience') }}</label>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="selectedUsers.length === 0" @click="clearSelectedUsers">
              {{ t('admin.smsBroadcasts.form.clearSelected') }}
            </button>
          </div>

          <div class="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
            <div class="rounded-lg border border-gray-200 dark:border-dark-700">
              <div class="border-b border-gray-200 p-3 dark:border-dark-700">
                <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
                  <div class="min-w-0">
                    <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {{ t('admin.smsBroadcasts.form.availableUsers') }}
                    </div>
                    <div v-if="audiencePagination.total > 0" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                      {{ t('admin.smsBroadcasts.form.usersLoadedSummary', { shown: audienceUsers.length, total: audiencePagination.total }) }}
                    </div>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <button
                      type="button"
                      data-test="sms-add-all-phone-users"
                      class="btn btn-secondary btn-sm"
                      :disabled="audienceBusy"
                      @click="addAllPhoneUsers"
                    >
                      {{ addingAllPhoneUsers ? t('admin.smsBroadcasts.form.addingAllPhoneUsers') : t('admin.smsBroadcasts.form.addAllPhoneUsers') }}
                    </button>
                    <button type="button" class="btn btn-secondary btn-sm" :disabled="audienceBusy" @click="loadAudienceUsers({ reset: true })">
                      <Icon name="refresh" size="sm" :class="audienceLoading ? 'animate-spin' : ''" />
                    </button>
                  </div>
                </div>
                <div class="grid grid-cols-1 gap-2 md:grid-cols-4">
                  <input
                    v-model="audienceFilters.search"
                    data-test="sms-audience-search"
                    class="input md:col-span-1"
                    :placeholder="t('admin.smsBroadcasts.form.searchPlaceholder')"
                    @input="debounceAudienceSearch"
                  />
                  <select v-model="audienceFilters.role" data-test="sms-audience-role" class="input" :disabled="audienceBusy" @change="loadAudienceUsers({ reset: true })">
                    <option value="">{{ t('admin.smsBroadcasts.form.allRoles') }}</option>
                    <option value="user">{{ t('admin.smsBroadcasts.form.roleUser') }}</option>
                    <option value="admin">{{ t('admin.smsBroadcasts.form.roleAdmin') }}</option>
                  </select>
                  <select v-model="audienceFilters.status" data-test="sms-audience-status" class="input" :disabled="audienceBusy" @change="loadAudienceUsers({ reset: true })">
                    <option value="">{{ t('admin.smsBroadcasts.form.allStatus') }}</option>
                    <option value="active">{{ t('admin.smsBroadcasts.form.statusActive') }}</option>
                    <option value="disabled">{{ t('admin.smsBroadcasts.form.statusDisabled') }}</option>
                  </select>
                  <label class="flex items-center gap-2">
                    <span class="shrink-0 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.smsBroadcasts.form.pageSize') }}</span>
                    <select
                      v-model.number="audiencePagination.page_size"
                      data-test="sms-audience-page-size"
                      class="input"
                      :disabled="audienceBusy"
                      @change="handleAudiencePageSizeChange"
                    >
                      <option v-for="size in audiencePageSizeOptions" :key="size" :value="size">{{ size }}</option>
                    </select>
                  </label>
                </div>
              </div>
              <div class="max-h-72 overflow-y-auto p-2">
                <div v-if="audienceLoading" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
                  {{ t('common.loading') }}
                </div>
                <div v-else-if="audienceUsers.length === 0" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.smsBroadcasts.form.noUsers') }}
                </div>
                <template v-else>
                  <button
                    v-for="user in audienceUsers"
                    :key="user.id"
                    type="button"
                    :data-test="`sms-add-user-${user.id}`"
                    class="mb-2 flex w-full items-center justify-between gap-3 rounded-lg border border-gray-200 p-3 text-left transition-colors hover:border-primary-300 hover:bg-primary-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-700 dark:hover:border-primary-700 dark:hover:bg-primary-900/20"
                    :disabled="!canAddUser(user)"
                    @click="addSelectedUser(user)"
                  >
                    <span class="min-w-0">
                      <span class="block truncate text-sm font-medium text-gray-900 dark:text-gray-100">{{ userLabel(user) }}</span>
                      <span class="mt-0.5 block truncate text-xs text-gray-500 dark:text-dark-400">{{ user.email }} · {{ user.phone_number || t('admin.smsBroadcasts.form.noPhone') }}</span>
                    </span>
                    <Icon name="plus" size="sm" />
                  </button>
                  <div class="pt-1 text-center">
                    <button
                      v-if="audienceHasMore"
                      type="button"
                      data-test="sms-load-more-users"
                      class="btn btn-secondary btn-sm w-full justify-center"
                      :disabled="audienceBusy"
                      @click="loadMoreAudienceUsers"
                    >
                      {{ audienceBusy ? t('common.loading') : t('admin.smsBroadcasts.form.loadMoreUsers') }}
                    </button>
                    <div v-else-if="audiencePagination.total > 0" class="py-2 text-xs text-gray-500 dark:text-dark-400">
                      {{ t('admin.smsBroadcasts.form.allUsersLoaded') }}
                    </div>
                  </div>
                </template>
              </div>
            </div>

            <div class="rounded-lg border border-gray-200 dark:border-dark-700">
              <div class="flex items-center justify-between gap-3 border-b border-gray-200 p-3 dark:border-dark-700">
                <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {{ t('admin.smsBroadcasts.form.selectedUsers') }}
                </div>
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('admin.smsBroadcasts.form.selectedCount', { count: selectedUsers.length }) }}
                </div>
              </div>
              <div class="max-h-72 overflow-y-auto p-2">
                <div v-if="selectedUsers.length === 0" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.smsBroadcasts.form.noSelectedUsers') }}
                </div>
                <template v-else>
                  <div
                    v-for="user in selectedUsers"
                    :key="user.id"
                    class="mb-2 flex items-center justify-between gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-700"
                  >
                    <span class="min-w-0">
                      <span class="block truncate text-sm font-medium text-gray-900 dark:text-gray-100">{{ userLabel(user) }}</span>
                      <span class="mt-0.5 block truncate text-xs text-gray-500 dark:text-dark-400">{{ user.email }} · {{ user.phone_number }}</span>
                    </span>
                    <button
                      type="button"
                      :data-test="`sms-remove-user-${user.id}`"
                      class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                      @click="removeSelectedUser(user.id)"
                    >
                      <Icon name="x" size="sm" />
                    </button>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>

        <div>
          <div class="mb-3 flex items-center justify-between gap-3">
            <label class="input-label mb-0">{{ t('admin.smsBroadcasts.form.variables') }}</label>
            <button type="button" data-test="sms-add-var" class="btn btn-secondary btn-sm" @click="addVariable">
              <Icon name="plus" size="sm" class="mr-1.5" />
              {{ t('admin.smsBroadcasts.form.addVariable') }}
            </button>
          </div>
          <div class="space-y-2">
            <div
              v-for="(variable, index) in form.variables"
              :key="variable.id"
              class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] gap-2"
            >
              <input
                v-model="variable.key"
                :data-test="`sms-var-key-${index}`"
                class="input"
                :placeholder="t('admin.smsBroadcasts.form.variableKey')"
                :required="hasVariableValue(variable)"
              />
              <input
                v-model="variable.value"
                :data-test="`sms-var-value-${index}`"
                class="input"
                :placeholder="t('admin.smsBroadcasts.form.variableValue')"
              />
              <button
                type="button"
                :data-test="`sms-remove-var-${index}`"
                class="btn btn-secondary"
                :disabled="form.variables.length === 1"
                @click="removeVariable(index)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </div>

      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeCreate">{{ t('common.cancel') }}</button>
          <button type="submit" form="sms-broadcast-form" class="btn btn-primary" :disabled="creating">
            {{ creating ? t('common.saving') : t('admin.smsBroadcasts.sendNow') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import type { SMSBroadcastAudience, SMSBroadcastCampaign, SMSBroadcastTemplateVarRow } from '@/api/admin/smsBroadcasts'
import type { AdminUser } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const creating = ref(false)
const audienceLoading = ref(false)
const addingAllPhoneUsers = ref(false)
const showCreate = ref(false)
const items = ref<SMSBroadcastCampaign[]>([])
const audienceUsers = ref<AdminUser[]>([])
const selectedUsers = ref<AdminUser[]>([])
let audienceAbortController: AbortController | null = null
let audienceSearchTimeout: ReturnType<typeof setTimeout> | null = null
let variableID = 0

const form = reactive({
  title: '',
  template_id: '',
  variables: [{ id: variableID++, key: '', value: '' }]
})

const audienceFilters = reactive({
  status: 'active' as '' | 'active' | 'disabled',
  role: 'user' as '' | 'user' | 'admin',
  search: ''
})

const audiencePagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const audiencePageSizeOptions = [20, 50, 100]
const audienceBusy = computed(() => audienceLoading.value || addingAllPhoneUsers.value)

const audienceHasMore = computed(() => (
  audiencePagination.total > audienceUsers.value.length &&
  audiencePagination.page < audiencePagination.pages
))

const columns = computed<Column[]>(() => {
  const baseColumns: Column[] = [
    { key: 'title', label: t('admin.smsBroadcasts.columns.title') },
    { key: 'status', label: t('admin.smsBroadcasts.columns.status') },
    { key: 'summary', label: t('admin.smsBroadcasts.columns.summary') },
    { key: 'created_at', label: t('admin.smsBroadcasts.columns.createdAt') }
  ]
  if (items.value.some(canCancel)) {
    baseColumns.push({ key: 'actions', label: t('common.actions'), sortable: false })
  }
  return baseColumns
})

async function load() {
  loading.value = true
  try {
    const data = await adminAPI.smsBroadcasts.list(1, 20)
    items.value = data.items
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  showCreate.value = true
  resetAudienceSelection()
  void loadAudienceUsers({ reset: true })
}

function closeCreate() {
  showCreate.value = false
  audienceAbortController?.abort()
}

async function loadAudienceUsers(options: { reset?: boolean; page?: number } = {}) {
  if (addingAllPhoneUsers.value) return
  audienceAbortController?.abort()
  const controller = new AbortController()
  audienceAbortController = controller
  audienceLoading.value = true
  const nextPage = options.page ?? (options.reset ? 1 : audiencePagination.page)
  try {
    const result = await adminAPI.users.list(nextPage, audiencePagination.page_size, {
      ...audienceListFilters()
    }, { signal: controller.signal })
    if (controller.signal.aborted) return
    audiencePagination.page = result.page ?? nextPage
    audiencePagination.total = result.total ?? result.items.length
    audiencePagination.pages = result.pages ?? Math.ceil(audiencePagination.total / audiencePagination.page_size)
    audienceUsers.value = options.reset ? result.items : mergeUsersByID(audienceUsers.value, result.items)
  } catch (error: any) {
    const errorInfo = error as { name?: string; code?: string }
    if (errorInfo?.name === 'AbortError' || errorInfo?.name === 'CanceledError' || errorInfo?.code === 'ERR_CANCELED') {
      return
    }
    appStore.showError(error.response?.data?.detail || t('admin.smsBroadcasts.form.failedToLoadUsers'))
  } finally {
    if (audienceAbortController === controller) {
      audienceLoading.value = false
    }
  }
}

function loadMoreAudienceUsers() {
  if (audienceBusy.value || !audienceHasMore.value) return
  void loadAudienceUsers({ page: audiencePagination.page + 1 })
}

function handleAudiencePageSizeChange() {
  void loadAudienceUsers({ reset: true })
}

async function addAllPhoneUsers() {
  if (audienceBusy.value) return
  audienceAbortController?.abort()
  audienceAbortController = null
  addingAllPhoneUsers.value = true
  const fetchedUsers: AdminUser[] = []
  try {
    let page = 1
    let pages = 1
    let total = 0
    do {
      const result = await adminAPI.users.list(page, audiencePagination.page_size, audienceListFilters(), {})
      fetchedUsers.push(...result.items)
      total = result.total ?? fetchedUsers.length
      pages = Math.max(1, result.pages ?? Math.ceil(total / audiencePagination.page_size))
      page += 1
    } while (page <= pages)

    const mergedAudience = mergeUsersByID(audienceUsers.value, fetchedUsers)
    audienceUsers.value = mergedAudience
    audiencePagination.page = pages
    audiencePagination.total = total
    audiencePagination.pages = pages
    selectedUsers.value = mergeUsersByID(selectedUsers.value, fetchedUsers.filter(hasValidPhone))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.smsBroadcasts.form.failedToAddAllPhoneUsers'))
  } finally {
    addingAllPhoneUsers.value = false
  }
}

async function submitCreate() {
  const vars = buildVarsPayload()
  if (vars === null) return
  const audience = buildAudiencePayload()
  if (!audience) return
  creating.value = true
  try {
    await adminAPI.smsBroadcasts.create({
      title: form.title,
      template_id: form.template_id.trim(),
      vars,
      audience
    })
    closeCreate()
    await load()
  } finally {
    creating.value = false
  }
}

function addVariable() {
  form.variables.push({ id: variableID++, key: '', value: '' })
}

function removeVariable(index: number) {
  if (form.variables.length === 1) return
  form.variables.splice(index, 1)
}

function hasVariableValue(variable: { key: string; value: string }) {
  return variable.value.trim() !== ''
}

function buildVarsPayload(): SMSBroadcastTemplateVarRow[] | null {
  const vars: SMSBroadcastTemplateVarRow[] = []
  const seen = new Set<string>()
  for (const variable of form.variables) {
    const key = variable.key.trim()
    const value = variable.value.trim()
    if (!key && !value) continue
    if (!key || !value || seen.has(key)) return null
    seen.add(key)
    vars.push({ key, value })
  }
  return vars
}

function buildAudiencePayload(): SMSBroadcastAudience | null {
  const userIDs = selectedUsers.value.map((user) => user.id)
  if (userIDs.length === 0) {
    appStore.showError(t('admin.smsBroadcasts.form.selectedUsersRequired'))
    return null
  }
  return { user_ids: userIDs }
}

function resetAudienceSelection() {
  audienceFilters.status = 'active'
  audienceFilters.role = 'user'
  audienceFilters.search = ''
  selectedUsers.value = []
  audienceUsers.value = []
  resetAudiencePagination()
}

function debounceAudienceSearch() {
  if (audienceSearchTimeout) {
    clearTimeout(audienceSearchTimeout)
  }
  audienceSearchTimeout = setTimeout(() => loadAudienceUsers({ reset: true }), 300)
}

function resetAudiencePagination() {
  audiencePagination.page = 1
  audiencePagination.page_size = 20
  audiencePagination.total = 0
  audiencePagination.pages = 0
}

function audienceListFilters() {
  return {
    status: audienceFilters.status || undefined,
    role: audienceFilters.role || undefined,
    search: audienceFilters.search.trim() || undefined,
    include_subscriptions: false,
    has_phone: true
  }
}

function mergeUsersByID(current: AdminUser[], incoming: AdminUser[]) {
  const seen = new Set<number>()
  const merged: AdminUser[] = []
  for (const user of [...current, ...incoming]) {
    if (seen.has(user.id)) continue
    seen.add(user.id)
    merged.push(user)
  }
  return merged
}

function hasValidPhone(user: AdminUser) {
  return Boolean(user.phone_number?.trim())
}

function isUserSelected(userID: number) {
  return selectedUsers.value.some((user) => user.id === userID)
}

function canAddUser(user: AdminUser) {
  return hasValidPhone(user) && !isUserSelected(user.id)
}

function addSelectedUser(user: AdminUser) {
  if (!canAddUser(user)) return
  selectedUsers.value.push(user)
}

function removeSelectedUser(userID: number) {
  selectedUsers.value = selectedUsers.value.filter((user) => user.id !== userID)
}

function clearSelectedUsers() {
  selectedUsers.value = []
}

function userLabel(user: AdminUser) {
  return user.username || user.email || `#${user.id}`
}

async function handleCancel(row: SMSBroadcastCampaign) {
  if (!row.id) return
  await adminAPI.smsBroadcasts.cancel(row.id)
  await load()
}

function campaignTitle(row: SMSBroadcastCampaign) {
  return row.title || '-'
}

function campaignBody(row: SMSBroadcastCampaign) {
  if (row.template_id) return row.template_id
  return row.body || ''
}

function countOf(row: SMSBroadcastCampaign, key: keyof SMSBroadcastCampaign) {
  return Number(row[key] ?? 0)
}

function canCancel(row: SMSBroadcastCampaign) {
  return row.status === 'queued' || row.status === 'running'
}

function statusLabel(value?: string) {
  if (!value) return '-'
  return t(`admin.smsBroadcasts.statusLabels.${value}`)
}

function statusClass(value?: string) {
  if (value === 'running' || value === 'queued') return 'badge-warning'
  if (value === 'succeeded') return 'badge-success'
  if (value === 'failed' || value === 'canceled') return 'badge-danger'
  return 'badge-gray'
}

onMounted(load)
onUnmounted(() => {
  audienceAbortController?.abort()
  if (audienceSearchTimeout) {
    clearTimeout(audienceSearchTimeout)
  }
})
</script>
