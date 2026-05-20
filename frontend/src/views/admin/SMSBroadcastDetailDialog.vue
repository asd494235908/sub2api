<template>
  <BaseDialog :show="show" :title="t('admin.smsBroadcasts.details.title')" width="full" @close="close">
    <div class="flex h-full min-h-0 flex-col">
      <div class="mb-4 flex flex-shrink-0 items-start justify-between gap-3 border-b border-gray-200 pb-4 dark:border-dark-700">
        <div class="min-w-0 space-y-2">
          <div class="text-base font-semibold text-gray-900 dark:text-gray-100">
            {{ campaign?.title || '-' }}
          </div>
          <div class="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-dark-400">
            <span class="rounded-full bg-gray-100 px-2 py-1 dark:bg-dark-700">
              {{ t('admin.smsBroadcasts.details.templateId') }}: {{ campaign?.template_id || '-' }}
            </span>
            <span v-if="campaignVarsSummary" class="rounded-full bg-gray-100 px-2 py-1 dark:bg-dark-700">
              {{ campaignVarsSummary }}
            </span>
          </div>
        </div>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="refresh">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div v-if="loading && !campaign" class="flex flex-1 items-center justify-center py-16 text-sm text-gray-500 dark:text-dark-400">
        {{ t('common.loading') }}
      </div>

      <div v-else class="flex min-h-0 flex-1 flex-col gap-4">
        <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
          <div v-for="item in summaryItems" :key="item.key" class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
            <div class="text-xs text-gray-500 dark:text-dark-400">{{ item.label }}</div>
            <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">{{ item.value }}</div>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-3 border-b border-gray-200 pb-3 dark:border-dark-700">
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <span>{{ t('admin.smsBroadcasts.details.statusFilter') }}</span>
            <select v-model="statusFilter" class="input min-w-36" :disabled="loading">
              <option v-for="option in statusOptions" :key="option.value || 'all'" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>
        </div>

        <div class="min-h-0 flex-1 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <div class="min-h-0 flex-1 overflow-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="sticky top-0 z-10 bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.recipient') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.phoneNumber') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.rawPhone') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.renderedBody') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.status') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.failureReason') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.sentAt') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
                <tr v-if="recipientLoading">
                  <td colspan="7" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">
                    {{ t('common.loading') }}
                  </td>
                </tr>
                <tr v-else-if="recipients.length === 0">
                  <td colspan="7" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">
                    {{ t('admin.smsBroadcasts.details.noRecipients') }}
                  </td>
                </tr>
                <tr v-for="recipient in recipients" :key="`${recipient.user_id}-${recipient.phone_number}-${recipient.sent_at || ''}`" class="hover:bg-gray-50 dark:hover:bg-dark-800">
                  <td class="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">用户 #{{ recipient.user_id }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ recipient.phone_number || '-' }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ recipient.raw_phone || '-' }}</td>
                  <td class="max-w-[360px] px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                    <div class="max-w-[360px] truncate" :title="recipient.rendered_body || ''">
                      {{ recipient.rendered_body || '-' }}
                    </div>
                  </td>
                  <td class="px-4 py-3 text-sm">
                    <span class="badge" :class="statusBadgeClass(recipient.status)">{{ recipient.status || '-' }}</span>
                  </td>
                  <td class="max-w-[240px] px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                    <div class="max-w-[240px] truncate" :title="recipient.error_message || ''">
                      {{ recipient.error_message || '-' }}
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                    {{ recipient.sent_at ? formatDateTime(recipient.sent_at) : '-' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <Pagination
          v-if="total > 0"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import { adminAPI } from '@/api/admin'
import type { SMSBroadcastCampaign, SMSBroadcastRecipientPreview } from '@/api/admin/smsBroadcasts'

interface Props {
  show: boolean
  campaignId: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const recipientLoading = ref(false)
const campaign = ref<SMSBroadcastCampaign | null>(null)
const recipients = ref<SMSBroadcastRecipientPreview[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const statusFilter = ref('')

const summaryItems = computed(() => [
  { key: 'total', label: t('admin.smsBroadcasts.details.totalRecipients'), value: String(campaign.value?.total_recipients ?? 0) },
  { key: 'sent', label: t('admin.smsBroadcasts.details.sentCount'), value: String(campaign.value?.sent_count ?? 0) },
  { key: 'failed', label: t('admin.smsBroadcasts.details.failedCount'), value: String(campaign.value?.failed_count ?? 0) },
  { key: 'skipped', label: t('admin.smsBroadcasts.details.skippedCount'), value: String(campaign.value?.skipped_count ?? 0) }
])

const campaignVarsSummary = computed(() => {
  const rows = campaign.value?.template_var_rows ?? []
  if (rows.length > 0) {
    return rows.map((row) => `${row.key}=${templateVarDisplayValue(row)}`).join(' · ')
  }
  const vars = campaign.value?.template_vars
  if (!vars) return ''
  const entries = Object.entries(vars)
  if (entries.length === 0) return ''
  return entries.map(([key, value]) => `${key}=${value}`).join(' · ')
})

const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'queued', label: t('admin.smsBroadcasts.details.statusQueued') },
  { value: 'succeeded', label: t('admin.smsBroadcasts.details.statusSucceeded') },
  { value: 'failed', label: t('admin.smsBroadcasts.details.statusFailed') },
  { value: 'canceled', label: t('admin.smsBroadcasts.details.statusCanceled') }
])

function close() {
  emit('close')
}

async function loadCampaign() {
  if (!props.campaignId) return
  campaign.value = await adminAPI.smsBroadcasts.getById(props.campaignId)
}

async function loadRecipients() {
  if (!props.campaignId) return
  recipientLoading.value = true
  try {
    const res = await adminAPI.smsBroadcasts.getRecipients(
      props.campaignId,
      page.value,
      pageSize.value,
      statusFilter.value || undefined
    )
    recipients.value = res.items || []
    total.value = res.total || 0
    page.value = res.page || page.value
    pageSize.value = res.page_size || pageSize.value
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('admin.smsBroadcasts.details.failedToLoadRecipients'))
    recipients.value = []
    total.value = 0
  } finally {
    recipientLoading.value = false
  }
}

async function refresh() {
  if (!props.show || !props.campaignId) return
  loading.value = true
  try {
    await Promise.all([loadCampaign(), loadRecipients()])
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('admin.smsBroadcasts.details.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function handlePageChange(next: number) {
  page.value = next
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
}

function statusBadgeClass(status?: string) {
  if (status === 'succeeded') return 'badge-success'
  if (status === 'failed' || status === 'canceled') return 'badge-danger'
  if (status === 'queued') return 'badge-warning'
  return 'badge-gray'
}

function templateVarDisplayValue(row: { value?: string; source?: string }) {
  if (row.source) return variableSourceLabel(row.source)
  return row.value || ''
}

function variableSourceLabel(source: string) {
  if (source === 'phone_number') return t('admin.smsBroadcasts.form.variableSourcePhone')
  if (source === 'email') return t('admin.smsBroadcasts.form.variableSourceEmail')
  if (source === 'username') return t('admin.smsBroadcasts.form.variableSourceUsername')
  return source
}

watch(
  () => [props.show, props.campaignId] as const,
  async ([open]) => {
    if (!open || !props.campaignId) return
    page.value = 1
    pageSize.value = 20
    statusFilter.value = ''
    loading.value = true
    try {
      await Promise.all([loadCampaign(), loadRecipients()])
    } finally {
      loading.value = false
    }
  },
  { immediate: true }
)

watch(
  () => [page.value, pageSize.value, statusFilter.value] as const,
  () => {
    if (!props.show || !props.campaignId) return
    void loadRecipients()
  }
)
</script>
