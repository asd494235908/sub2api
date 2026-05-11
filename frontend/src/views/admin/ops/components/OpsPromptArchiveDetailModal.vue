<script setup lang="ts">
import { computed } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PromptArchiveRecord } from '@/api/admin/ops'

const props = defineProps<{
  show: boolean
  record: PromptArchiveRecord | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const close = () => emit('close')

const prettyAttachments = computed(() => props.record?.attachments ?? [])
</script>

<template>
  <BaseDialog :show="show" title="归档详情" width="extra-wide" @close="close">
    <div v-if="record" class="space-y-4 text-sm text-gray-700 dark:text-gray-300">
      <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <div><span class="font-semibold">请求 ID：</span>{{ record.request_id || '-' }}</div>
        <div><span class="font-semibold">Session：</span>{{ record.session_id || '-' }}</div>
        <div><span class="font-semibold">用户：</span>{{ record.username_snapshot || '-' }}</div>
        <div><span class="font-semibold">邮箱：</span>{{ record.email_snapshot || '-' }}</div>
        <div><span class="font-semibold">模型：</span>{{ record.model || '-' }}</div>
        <div><span class="font-semibold">状态：</span>{{ record.status || '-' }}</div>
      </div>

      <div>
        <div class="mb-1 font-semibold">System Prompt</div>
        <pre class="rounded-lg bg-gray-50 p-3 text-xs whitespace-pre-wrap break-words dark:bg-dark-800">{{ record.system_prompt || '-' }}</pre>
      </div>

      <div>
        <div class="mb-1 font-semibold">User Prompt</div>
        <pre class="rounded-lg bg-gray-50 p-3 text-xs whitespace-pre-wrap break-words dark:bg-dark-800">{{ record.user_prompt_text || '-' }}</pre>
      </div>

      <div>
        <div class="mb-1 font-semibold">附件</div>
        <div v-if="prettyAttachments.length === 0" class="text-xs text-gray-500">无附件</div>
        <div v-else class="space-y-2">
          <div
            v-for="(attachment, index) in prettyAttachments"
            :key="`${attachment.sequence || index}-${attachment.object_key || attachment.source_url || 'attachment'}`"
            class="rounded-lg border border-gray-200 p-3 text-xs dark:border-dark-700"
          >
            <div><span class="font-semibold">类型：</span>{{ attachment.kind }}</div>
            <div><span class="font-semibold">MIME：</span>{{ attachment.mime_type || '-' }}</div>
            <div><span class="font-semibold">来源：</span>{{ attachment.source_type }}</div>
            <div><span class="font-semibold">对象键：</span>{{ attachment.object_key || '-' }}</div>
            <div><span class="font-semibold">URL：</span>{{ attachment.source_url || '-' }}</div>
          </div>
        </div>
      </div>

      <div v-if="record.presigned_url">
        <a :href="record.presigned_url" target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline dark:text-blue-400">
          打开 Markdown 归档正文
        </a>
      </div>
    </div>
  </BaseDialog>
</template>
