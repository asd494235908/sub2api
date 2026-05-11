<script setup lang="ts">
import { ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { opsAPI, type PromptArchiveSessionDetail } from '@/api/admin/ops'
import { useAppStore } from '@/stores'

const props = defineProps<{
  show: boolean
  sessionId: string
  groupId: number | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const appStore = useAppStore()
const loading = ref(false)
const detail = ref<PromptArchiveSessionDetail | null>(null)

const close = () => emit('close')

async function loadSession() {
  if (!props.show || !props.sessionId || !props.groupId) return
  loading.value = true
  try {
    detail.value = await opsAPI.getPromptArchiveSession(props.sessionId, props.groupId)
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || '加载归档会话详情失败')
  } finally {
    loading.value = false
  }
}

watch(() => [props.show, props.sessionId, props.groupId], () => {
  void loadSession()
}, { immediate: true })
</script>

<template>
  <BaseDialog :show="show" title="会话归档详情" width="extra-wide" @close="close">
    <div v-if="loading" class="py-10 text-center text-sm text-gray-500">加载中...</div>
    <div v-else-if="detail" class="space-y-3">
      <div class="text-xs text-gray-500">Session: {{ detail.session_id }} / Group: {{ detail.group_id }}</div>
      <div
        v-for="(record, index) in detail.records"
        :key="record.id || index"
        class="rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="mb-2 flex items-center justify-between gap-3">
          <div class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ record.model || '-' }} · {{ record.status || '-' }}
          </div>
          <div class="text-xs text-gray-500">{{ record.created_at }}</div>
        </div>
        <div class="text-xs text-gray-700 dark:text-gray-300 whitespace-pre-wrap break-words">
          {{ record.prompt_summary || record.user_prompt_text || '-' }}
        </div>
      </div>
    </div>
  </BaseDialog>
</template>
