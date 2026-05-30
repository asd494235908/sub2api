<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.accounts.dataImportHint') }}
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ t('admin.accounts.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportFile') }}</label>
        <div
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200">
              {{ fileLabel || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">JSON (.json)</div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json"
          multiple
          @change="handleFileChange"
        />
        <div
          v-if="selectedFileNames.length > 1"
          class="mt-2 max-h-24 overflow-auto rounded-lg bg-gray-50 p-2 text-xs text-gray-600 dark:bg-dark-800 dark:text-dark-300"
        >
          <div v-for="name in selectedFileNames" :key="name" class="truncate">
            {{ name }}
          </div>
        </div>
      </div>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.accounts.dataImportResult') }}
        </div>
        <div class="text-sm text-gray-700 dark:text-dark-300">
          {{ t('admin.accounts.dataImportResultSummary', result) }}
        </div>
        <div
          v-if="fileFailures.length"
          class="text-sm text-red-600 dark:text-red-400"
        >
          {{ t('admin.accounts.dataImportFileFailureSummary', { count: fileFailures.length, files: failedFileNames }) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">
            {{ t('admin.accounts.dataImportErrors') }}
          </div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-gray-50 p-3 font-mono text-xs dark:bg-dark-800"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              {{ item.fileName ? `[${item.fileName}] ` : '' }}{{ item.kind }} {{ item.name || item.proxy_key || '-' }} — {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.accounts.dataImporting') : t('admin.accounts.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult } from '@/types'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const files = ref<File[]>([])
const result = ref<AdminDataImportResult | null>(null)
const fileFailures = ref<FileImportFailure[]>([])
const resultErrors = ref<DisplayImportError[]>([])

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFileNames = computed(() => files.value.map((item) => item.name))
const fileLabel = computed(() => {
  if (files.value.length === 0) return ''
  if (files.value.length === 1) return files.value[0].name
  return t('admin.accounts.dataImportSelectedFiles', { count: files.value.length })
})
const failedFileNames = computed(() => fileFailures.value.map((item) => item.fileName).join(', '))
const failedFileDetails = computed(() => fileFailures.value.map((item) => `${item.fileName}: ${item.message}`).join('; '))

type FileImportFailure = {
  fileName: string
  message: string
}

type DisplayImportError = {
  fileName?: string
  kind: 'proxy' | 'account' | 'file'
  name?: string
  proxy_key?: string
  message: string
}

const errorItems = computed<DisplayImportError[]>(() => [
  ...fileFailures.value.map((item) => ({
    fileName: item.fileName,
    kind: 'file' as const,
    message: item.message
  })),
  ...resultErrors.value
])

watch(
  () => props.show,
  (open) => {
    if (open) {
      files.value = []
      result.value = null
      fileFailures.value = []
      resultErrors.value = []
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  files.value = Array.from(target.files || [])
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const createEmptyResult = (): AdminDataImportResult => ({
  proxy_created: 0,
  proxy_reused: 0,
  proxy_failed: 0,
  account_created: 0,
  account_failed: 0,
  errors: []
})

const mergeResult = (
  target: AdminDataImportResult,
  source: AdminDataImportResult,
  fileName: string
) => {
  target.proxy_created += source.proxy_created
  target.proxy_reused += source.proxy_reused
  target.proxy_failed += source.proxy_failed
  target.account_created += source.account_created
  target.account_failed += source.account_failed
  if (source.errors?.length) {
    resultErrors.value.push(
      ...source.errors.map((item) => ({
        ...item,
        fileName
      }))
    )
  }
}

const getImportErrorMessage = (error: unknown, parseFallback: string): string => {
  if (error instanceof SyntaxError) {
    return parseFallback
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  if (typeof error === 'object' && error !== null && 'message' in error) {
    const message = (error as { message?: unknown }).message
    if (typeof message === 'string' && message) {
      return message
    }
  }
  return t('admin.accounts.dataImportFailed')
}

const handleImport = async () => {
  if (files.value.length === 0) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  result.value = null
  fileFailures.value = []
  resultErrors.value = []

  try {
    const aggregate = createEmptyResult()
    let completedImports = 0

    for (const selectedFile of files.value) {
      try {
        const text = await readFileAsText(selectedFile)
        const dataPayload = JSON.parse(text)

        const res = await adminAPI.accounts.importData({
          data: dataPayload,
          skip_default_group_bind: true
        })

        completedImports++
        mergeResult(aggregate, res, selectedFile.name)
      } catch (error) {
        const message = getImportErrorMessage(error, t('admin.accounts.dataImportParseFailed'))
        fileFailures.value.push({
          fileName: selectedFile.name,
          message
        })
      }
    }

    result.value = aggregate

    const msgParams: Record<string, unknown> = {
      account_created: aggregate.account_created,
      account_failed: aggregate.account_failed,
      proxy_created: aggregate.proxy_created,
      proxy_reused: aggregate.proxy_reused,
      proxy_failed: aggregate.proxy_failed,
      file_failed: fileFailures.value.length,
      failed_files: failedFileDetails.value || failedFileNames.value
    }

    if (fileFailures.value.length > 0 || aggregate.account_failed > 0 || aggregate.proxy_failed > 0) {
      if (files.value.length === 1 && completedImports === 0 && fileFailures.value[0]?.message === t('admin.accounts.dataImportParseFailed')) {
        appStore.showError(t('admin.accounts.dataImportParseFailed'))
      } else {
        appStore.showError(t('admin.accounts.dataImportCompletedWithErrors', msgParams))
      }
    } else {
      appStore.showSuccess(t('admin.accounts.dataImportSuccess', msgParams))
    }

    if (completedImports > 0) {
      emit('imported')
    }
  } catch (error: any) {
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.accounts.dataImportParseFailed'))
    } else {
      appStore.showError(error?.message || t('admin.accounts.dataImportFailed'))
    }
  } finally {
    importing.value = false
  }
}
</script>
