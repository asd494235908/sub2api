<template>
  <div
    ref="triggerEl"
    :class="['inline-flex min-w-0 flex-wrap gap-1', attrs.class]"
    data-test="supported-models-summary"
    tabindex="0"
    role="button"
    :aria-label="t('payment.planCard.supportedModelsTitle', { count: normalizedModels.length })"
    @mouseenter="openPopover"
    @mouseleave="scheduleClosePopover"
    @focusin="openPopover"
    @focusout="scheduleClosePopover"
    @click.stop="togglePopover"
    @keydown.esc.stop.prevent="closePopover"
  >
    <span
      v-for="model in previewModels"
      :key="model"
      class="max-w-[10rem] truncate rounded bg-gray-200/80 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300"
      :title="model"
    >
      {{ model }}
    </span>
    <span
      v-if="hiddenModelCount > 0"
      class="cursor-help rounded bg-gray-200/80 px-1.5 py-0.5 text-[10px] font-semibold text-gray-600 dark:bg-dark-600 dark:text-gray-300"
    >
      +{{ hiddenModelCount }}
    </span>
  </div>

  <Teleport to="body">
    <div
      v-if="open"
      ref="popoverEl"
      role="tooltip"
      class="fixed z-[99999] w-80 max-w-[min(22rem,calc(100vw-1rem))] rounded-lg border bg-white text-xs shadow-xl dark:bg-dark-800"
      :class="popoverBorderClass"
      :style="popoverStyle"
      data-test="supported-models-popover"
      @mouseenter="openPopover"
      @mouseleave="scheduleClosePopover"
      @click.stop
    >
      <div
        class="flex items-center justify-between gap-2 rounded-t-lg border-b px-3 py-2"
        :class="[popoverHeaderClass, popoverBorderClass]"
      >
        <span class="truncate font-semibold">
          {{ t('payment.planCard.supportedModelsTitle', { count: normalizedModels.length }) }}
        </span>
      </div>
      <div class="max-h-64 overflow-y-auto p-3">
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="model in sortedModels"
            :key="model"
            class="max-w-full truncate rounded bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300"
            :title="model"
          >
            {{ model }}
          </span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, useAttrs } from 'vue'
import { useI18n } from 'vue-i18n'
import { platformBorderClass, platformBadgeLightClass } from '@/utils/platformColors'

const props = withDefaults(
  defineProps<{
    models: string[]
    platform?: string
    limit?: number
  }>(),
  {
    platform: '',
    limit: 3,
  },
)

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()
const { t } = useI18n()

const HOT_MODEL_PRIORITY = [
  'gpt-5.5',
  'gpt-5.4',
  'gpt-image-2',
  'gpt-5.4-mini',
  'claude-opus-4-7',
  'claude-opus-4-6',
  'claude-opus-4-6-thinking',
  'claude-sonnet-4-6',
  'claude-sonnet-4-5',
  'claude-sonnet-4-5-20250929',
  'gemini-3.1-pro-high',
  'gemini-3.1-pro-preview',
  'gemini-3.1-flash-image',
  'gemini-3-pro-preview',
  'gemini-3-flash',
  'gemini-2.5-pro',
  'gemini-2.5-flash',
  'gemini-2.5-flash-image',
]
const HOT_MODEL_RANK = new Map(HOT_MODEL_PRIORITY.map((model, index) => [model, index]))

const normalizedModels = computed(() => props.models.map(model => model.trim()).filter(Boolean))
const sortedModels = computed(() => {
  return normalizedModels.value
    .map((model, index) => ({ model, index, rank: HOT_MODEL_RANK.get(model) ?? Number.MAX_SAFE_INTEGER }))
    .sort((a, b) => a.rank - b.rank || a.index - b.index)
    .map(item => item.model)
})
const previewModels = computed(() => sortedModels.value.slice(0, props.limit))
const hiddenModelCount = computed(() => Math.max(0, normalizedModels.value.length - previewModels.value.length))

const popoverBorderClass = computed(() => (
  props.platform ? platformBorderClass(props.platform) : 'border-gray-200 dark:border-dark-600'
))
const popoverHeaderClass = computed(() => (
  props.platform ? platformBadgeLightClass(props.platform) : 'bg-gray-50 text-gray-700 dark:bg-dark-700/60 dark:text-gray-300'
))

const open = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const popoverEl = ref<HTMLElement | null>(null)
const popoverStyle = ref<Record<string, string>>({ top: '0px', left: '0px' })
let closeTimer: ReturnType<typeof setTimeout> | null = null

function clearCloseTimer() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

function updatePosition() {
  const trigger = triggerEl.value
  if (!trigger) return
  const rect = trigger.getBoundingClientRect()
  const margin = 8
  const popover = popoverEl.value
  const popWidth = popover?.offsetWidth ?? 320
  const popHeight = popover?.offsetHeight ?? 260
  const vw = window.innerWidth
  const vh = window.innerHeight

  let top = rect.bottom + margin
  if (top + popHeight > vh - margin) {
    top = Math.max(margin, rect.top - popHeight - margin)
  }

  let left = rect.left + rect.width / 2 - popWidth / 2
  if (left < margin) left = margin
  if (left + popWidth > vw - margin) left = vw - margin - popWidth

  popoverStyle.value = {
    top: `${Math.round(top)}px`,
    left: `${Math.round(left)}px`,
  }
}

function bindPositionListeners() {
  window.addEventListener('scroll', updatePosition, true)
  window.addEventListener('resize', updatePosition)
  document.addEventListener('click', closePopover)
}

function unbindPositionListeners() {
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('click', closePopover)
}

function openPopover() {
  if (normalizedModels.value.length === 0) return
  clearCloseTimer()
  open.value = true
  nextTick(() => {
    updatePosition()
    bindPositionListeners()
  })
}

function closePopover() {
  clearCloseTimer()
  open.value = false
  unbindPositionListeners()
}

function scheduleClosePopover() {
  clearCloseTimer()
  closeTimer = setTimeout(closePopover, 120)
}

function togglePopover() {
  if (open.value) {
    closePopover()
    return
  }
  openPopover()
}

onBeforeUnmount(() => {
  clearCloseTimer()
  unbindPositionListeners()
})
</script>
