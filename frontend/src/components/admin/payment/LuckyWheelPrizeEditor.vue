<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('luckyWheel.adminPrizesTitle') }}</h2>
      <div class="flex flex-wrap gap-2">
        <button class="btn btn-secondary btn-sm" @click="$emit('add')">{{ t('luckyWheel.adminAddPrize') }}</button>
      </div>
    </div>
    <VueDraggable
      v-model="localPrizes"
      :animation="200"
      handle=".drag-handle"
      class="space-y-3"
      data-test="prize-draggable"
      @end="onDragEnd"
    >
      <div
        v-for="(prize, index) in localPrizes"
        :key="prize.id || index"
        class="flex items-start gap-3 rounded-2xl border border-gray-100 p-4 dark:border-dark-700"
      >
        <div class="drag-handle mt-3 flex cursor-grab items-center text-gray-300 hover:text-gray-500 active:cursor-grabbing dark:text-dark-600 dark:hover:text-dark-400" data-test="drag-handle">
          <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path d="M7 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4z"/>
          </svg>
        </div>
        <div class="grid flex-1 gap-3 md:grid-cols-[1fr,1.2fr,0.8fr,auto,auto]">
          <input :value="prize.id" class="input" placeholder="prize_id" @input="$emit('update-prize', index, 'id', ($event.target as HTMLInputElement).value)" />
          <input :value="prize.name" class="input" placeholder="奖项名称" @input="$emit('update-prize', index, 'name', ($event.target as HTMLInputElement).value)" />
          <input :value="prize.reward_amount" class="input" type="number" min="0.01" step="0.01" placeholder="奖励金额" @input="$emit('update-prize', index, 'reward_amount', Number(($event.target as HTMLInputElement).value || 0))" />
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input :checked="prize.enabled" type="checkbox" @change="$emit('update-prize', index, 'enabled', ($event.target as HTMLInputElement).checked)" />
            启用
          </label>
          <button class="btn btn-secondary btn-sm text-red-500" @click="$emit('remove', index)">删除</button>
        </div>
      </div>
    </VueDraggable>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { useI18n } from 'vue-i18n'
import type { LuckyWheelPrize } from '@/types/payment'

const props = defineProps<{
  prizes: LuckyWheelPrize[]
}>()

const { t } = useI18n()

const emit = defineEmits<{
  (e: 'add'): void
  (e: 'remove', index: number): void
  (e: 'update-prize', index: number, field: 'id' | 'name' | 'reward_amount' | 'enabled', value: string | number | boolean): void
  (e: 'reorder', prizes: LuckyWheelPrize[]): void
}>()

const localPrizes = ref<LuckyWheelPrize[]>([])

watch(
  () => props.prizes,
  (value) => {
    localPrizes.value = value.map((prize) => ({ ...prize }))
  },
  { immediate: true, deep: true },
)

function onDragEnd() {
  emit('reorder', localPrizes.value.map((prize) => ({ ...prize })))
}
</script>
