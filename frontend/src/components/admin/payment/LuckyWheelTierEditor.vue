<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('luckyWheel.adminTiersTitle') }}</h2>
      <button class="btn btn-secondary btn-sm" @click="$emit('add-tier')">{{ t('luckyWheel.adminAddTier') }}</button>
    </div>
    <div class="space-y-4">
      <div v-for="(tier, tierIndex) in tiers" :key="tier.id || tierIndex" class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
        <div class="grid gap-3 md:grid-cols-[1fr,1fr,1fr,1fr,auto]">
          <input :value="tier.id" class="input" placeholder="tier_id" @input="$emit('update-tier', tierIndex, 'id', ($event.target as HTMLInputElement).value)" />
          <input :value="tier.name" class="input" placeholder="区间名称" @input="$emit('update-tier', tierIndex, 'name', ($event.target as HTMLInputElement).value)" />
          <input :value="tier.min_amount" class="input" type="number" min="0" step="0.01" placeholder="最小金额" @input="$emit('update-tier', tierIndex, 'min_amount', Number(($event.target as HTMLInputElement).value || 0))" />
          <input :value="tier.max_amount ?? ''" class="input" type="number" min="0" step="0.01" placeholder="最大金额，留空表示无上限" @input="$emit('update-tier', tierIndex, 'max_amount', ($event.target as HTMLInputElement).value)" />
          <button class="btn btn-secondary btn-sm text-red-500" @click="$emit('remove-tier', tierIndex)">删除</button>
        </div>
        <div class="mt-4 grid gap-3 md:grid-cols-2">
          <div
            v-for="prize in prizes"
            :key="`${tier.id}-${prize.id}`"
            class="rounded-xl border border-orange-100 bg-orange-50/50 p-3 dark:border-orange-500/10 dark:bg-orange-500/5"
          >
            <div class="mb-2 text-sm font-semibold text-gray-900 dark:text-white">{{ prize.name || prize.id || '未命名奖项' }}</div>
            <input
              :value="tier.prize_weights[prize.id] ?? 0"
              class="input"
              type="number"
              min="0"
              step="0.01"
              @input="$emit('update-weight', tierIndex, prize.id, ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
        <div v-if="warnings[tierIndex]?.length" class="mt-3 rounded-xl border border-amber-200 bg-amber-50/80 p-3 text-sm text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
          <div v-for="warning in warnings[tierIndex]" :key="warning">• {{ warning }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { LuckyWheelPrize, LuckyWheelTier } from '@/types/payment'

defineProps<{
  prizes: LuckyWheelPrize[]
  tiers: LuckyWheelTier[]
  warnings: string[][]
}>()

const { t } = useI18n()

defineEmits<{
  (e: 'add-tier'): void
  (e: 'remove-tier', index: number): void
  (e: 'update-tier', index: number, field: 'id' | 'name' | 'min_amount' | 'max_amount', value: string | number): void
  (e: 'update-weight', tierIndex: number, prizeID: string, value: string): void
}>()
</script>
