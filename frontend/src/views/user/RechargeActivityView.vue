<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6 p-4 sm:p-6">
      <div v-if="loading" class="flex min-h-[320px] items-center justify-center">
        <LoadingSpinner />
      </div>

      <template v-else>
        <section class="grid gap-6 lg:grid-cols-[minmax(0,1fr),360px]">
          <div class="space-y-5">
            <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <div class="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-semibold text-rose-600 dark:text-rose-400">{{ t('rechargeActivity.heroTag') }}</p>
                  <h1 class="mt-2 text-2xl font-bold text-slate-950 dark:text-white">{{ t('nav.rechargeActivity') }}</h1>
                  <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-300">
                    {{ summary?.config.intro_text || t('rechargeActivity.heroDescription') }}
                  </p>
                </div>
                <button class="btn btn-secondary btn-sm" type="button" @click="loadSummary">{{ t('rechargeActivity.refresh') }}</button>
              </div>
            </div>

            <div class="grid gap-5 md:grid-cols-[340px,minmax(0,1fr)]">
              <div class="rounded-lg border border-amber-100 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <div class="mx-auto flex max-w-[320px] items-center justify-center">
                  <div class="relative aspect-square w-full" data-test="recharge-activity-wheel">
                    <div
                      class="absolute inset-0 rounded-full transition-transform duration-[4600ms] [transition-timing-function:cubic-bezier(0.18,0.89,0.2,1)]"
                      :style="wheelRotorStyle"
                      data-test="recharge-activity-wheel-rotor"
                    >
                      <div
                        class="absolute inset-0 rounded-full border-[16px] border-amber-100 shadow-[inset_0_0_0_10px_rgba(255,255,255,0.7),0_26px_56px_rgba(251,146,60,0.14)] dark:border-amber-900/30"
                        :style="{ background: wheelGradient }"
                      />
                      <div class="absolute inset-[13%] rounded-full border border-white/60 bg-white/85 backdrop-blur dark:border-dark-600 dark:bg-dark-800/85">
                        <div
                          v-for="(segment, index) in wheelSegments"
                          :key="segment.key"
                          class="absolute left-1/2 top-1/2 origin-bottom -translate-x-1/2"
                          :style="segmentStyle(index)"
                        >
                          <span
                            class="block max-w-[5.5rem] truncate rounded-full px-2 py-1 text-[10px] font-bold shadow-sm"
                            :class="highlightedPrizeID === segment.prize.id ? 'bg-rose-500 text-white' : 'bg-white/90 text-slate-700 dark:bg-dark-700 dark:text-slate-100'"
                          >
                            {{ segment.prize.name }}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div
                      class="absolute left-1/2 top-[2px] z-10 -translate-x-1/2 -translate-y-[16px]"
                      :class="{ 'is-ticking': pointerTicking }"
                      data-test="recharge-activity-pointer"
                    >
                      <div class="h-0 w-0 border-x-[12px] border-t-[28px] border-x-transparent border-t-rose-500 drop-shadow-[0_10px_14px_rgba(244,63,94,0.28)]" />
                    </div>
                    <div class="absolute inset-[35%] z-10 flex items-center justify-center rounded-full border-[10px] border-amber-200 bg-[radial-gradient(circle_at_top,#fde68a,#f97316_62%,#e11d48)] text-center shadow-[0_18px_35px_rgba(249,115,22,0.28)]">
                      <div>
                        <div class="text-xs font-bold uppercase text-amber-50">{{ t('rechargeActivity.pendingChances') }}</div>
                        <div class="mt-1 text-3xl font-black text-white">{{ pendingChanceCount }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <div class="flex items-center justify-between gap-3">
                  <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.currentChance') }}</h2>
                  <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600 dark:bg-dark-700 dark:text-slate-300">
                    {{ activeChance ? formatOrderType(activeChance.source_order_type) : t('rechargeActivity.noChanceShort') }}
                  </span>
                </div>

                <div v-if="activeChance" class="mt-5 space-y-4">
                  <div class="grid grid-cols-2 gap-3">
                    <div class="rounded-lg bg-slate-50 p-3 dark:bg-dark-700/70">
                      <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.payAmount') }}</div>
                      <div class="mt-1 text-xl font-black text-slate-950 dark:text-white">{{ activeChance.source_pay_amount.toFixed(2) }}</div>
                    </div>
                    <div class="rounded-lg bg-slate-50 p-3 dark:bg-dark-700/70">
                      <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.sourceOrder') }}</div>
                      <div class="mt-1 text-xl font-black text-slate-950 dark:text-white">#{{ activeChance.source_order_id }}</div>
                    </div>
                  </div>
                  <button
                    class="btn h-12 w-full border-none bg-gradient-to-r from-amber-500 to-rose-500 font-bold text-white shadow-[0_14px_30px_rgba(249,115,22,0.25)] hover:from-amber-600 hover:to-rose-600 disabled:opacity-60"
                    data-test="draw-recharge-activity"
                    type="button"
                    :disabled="drawing || !summary?.enabled"
                    @click="drawNow"
                  >
                    {{ drawing ? t('common.loading') : t('rechargeActivity.drawNow') }}
                  </button>
                </div>

                <div v-else class="mt-5 rounded-lg border border-dashed border-slate-200 p-5 text-sm text-slate-500 dark:border-dark-600 dark:text-slate-400">
                  {{ summary?.enabled === false ? t('rechargeActivity.disabled') : t('rechargeActivity.noChance') }}
                </div>
              </div>
            </div>
          </div>

          <aside class="space-y-5">
            <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.prizeList') }}</h2>
              <div class="mt-4 space-y-3">
                <div
                  v-for="prize in enabledPrizes"
                  :key="prize.id"
                  class="rounded-lg border border-slate-100 p-3 dark:border-dark-700"
                  data-test="recharge-activity-prize"
                >
                  <div class="font-semibold text-slate-900 dark:text-white">{{ prize.name }}</div>
                  <div v-if="prize.reward_description" class="mt-1 text-sm leading-5 text-slate-600 dark:text-slate-300">{{ prize.reward_description }}</div>
                </div>
              </div>
            </div>

            <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ summary?.config.rules_title || t('rechargeActivity.rulesTitle') }}</h2>
              <ul class="mt-4 space-y-2 text-sm leading-6 text-slate-600 dark:text-slate-300">
                <li v-for="item in rulesItems" :key="item">{{ item }}</li>
              </ul>
            </div>
          </aside>
        </section>

        <section class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <h2 class="text-lg font-bold text-slate-950 dark:text-white">{{ t('rechargeActivity.historyTitle') }}</h2>
          <div v-if="historyRecords.length" class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="record in historyRecords"
              :key="record.id"
              class="rounded-lg border border-slate-100 p-4 dark:border-dark-700"
              data-test="recharge-activity-history-item"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="font-semibold text-slate-900 dark:text-white">{{ record.prize_name }}</div>
                <span class="rounded-full px-2 py-1 text-xs font-semibold" :class="fulfillmentBadgeClass(record.fulfillment_status)">
                  {{ fulfillmentStatusLabel(record.fulfillment_status) }}
                </span>
              </div>
              <div v-if="record.reward_description" class="mt-2 text-sm leading-5 text-slate-600 dark:text-slate-300">{{ record.reward_description }}</div>
              <div class="mt-2 text-xs text-slate-500 dark:text-slate-400">#{{ record.source_order_id }} · {{ formatDateTime(record.created_at) }}</div>
            </div>
          </div>
          <div v-else class="mt-4 text-sm text-slate-500 dark:text-slate-400">{{ t('rechargeActivity.noRecords') }}</div>
        </section>
      </template>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="resultModalOpen && lastDrawRecord" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/65 p-4" data-test="recharge-activity-result-modal" @click.self="closeResultModal">
          <div class="w-full max-w-md rounded-lg border border-amber-200 bg-white p-6 text-center shadow-[0_35px_90px_rgba(15,23,42,0.25)] dark:border-dark-700 dark:bg-dark-800">
            <p class="text-sm font-semibold uppercase tracking-[0.24em] text-amber-500">{{ t('rechargeActivity.resultTitle') }}</p>
            <h2 class="mt-3 text-3xl font-black text-slate-950 dark:text-white">{{ lastDrawRecord.prize_name }}</h2>
            <p v-if="lastDrawRecord.reward_description" class="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-300">{{ lastDrawRecord.reward_description }}</p>
            <p class="mt-4 text-xs font-semibold text-slate-500 dark:text-slate-400">{{ fulfillmentStatusLabel(lastDrawRecord.fulfillment_status) }}</p>
            <button class="btn btn-primary mt-6 w-full" data-test="recharge-activity-result-confirm" @click="closeResultModal">
              {{ t('common.confirm') }}
            </button>
          </div>
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { RechargeActivityDrawRecord, RechargeActivitySummary } from '@/types/payment'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const drawing = ref(false)
const summary = ref<RechargeActivitySummary | null>(null)
const lastDrawRecord = ref<RechargeActivityDrawRecord | null>(null)
const resultModalOpen = ref(false)
const highlightedPrizeID = ref<string | null>(null)
const wheelRotationDeg = ref(0)
const spinVersion = ref(0)
const pointerTicking = ref(false)

const WHEEL_SPIN_DURATION_MS = 4600
const POINTER_TARGET_DEG = 0

const activeChance = computed(() => summary.value?.pending_chances?.[0] ?? null)
const pendingChanceCount = computed(() => summary.value?.pending_chances?.length ?? 0)
const enabledPrizes = computed(() => (summary.value?.config.prizes ?? []).filter((prize) => prize.enabled))
const historyRecords = computed(() => summary.value?.history_records ?? [])
const rulesItems = computed(() => summary.value?.config.rules_items?.length ? summary.value.config.rules_items : [t('rechargeActivity.defaultRule')])

const wheelSegments = computed(() => {
  const prizes = enabledPrizes.value.length ? enabledPrizes.value : [{ id: 'empty', name: t('rechargeActivity.noChanceShort'), reward_amount: 0, reward_description: '', probability: 100, min_pay_amount: 0, enabled: true, sort_order: 0 }]
  return prizes.map((prize) => ({ key: prize.id, prize }))
})

const wheelGradient = computed(() => {
  const segments = wheelSegments.value
  const palette = ['#f59e0b', '#fb7185', '#f97316', '#fbbf24', '#14b8a6', '#60a5fa']
  const step = 100 / Math.max(segments.length, 1)
  const stops = segments.map((_, index) => {
    const color = palette[index % palette.length]
    return `${color} ${(index * step).toFixed(2)}% ${((index + 1) * step).toFixed(2)}%`
  })
  return `conic-gradient(${stops.join(',')})`
})

const wheelRotorStyle = computed(() => ({
  transform: `rotate(${wheelRotationDeg.value}deg)`,
  transitionDuration: spinVersion.value > 0 ? `${WHEEL_SPIN_DURATION_MS}ms` : '0ms',
}))

function segmentStyle(index: number) {
  const count = Math.max(wheelSegments.value.length, 1)
  const angle = (360 / count) * index
  return {
    transform: `translate(-50%, -100%) rotate(${angle}deg)`,
    height: '45%',
  }
}

function getWheelTargetRotation(prizeID: string) {
  const targetIndex = wheelSegments.value.findIndex((segment) => segment.prize.id === prizeID)
  if (targetIndex < 0) return wheelRotationDeg.value
  const segmentAngle = 360 / Math.max(wheelSegments.value.length, 1)
  return POINTER_TARGET_DEG - targetIndex * segmentAngle
}

function alignWheelToPrize(prizeID: string, animated: boolean) {
  const target = getWheelTargetRotation(prizeID)
  if (!animated) {
    const previousVersion = spinVersion.value
    spinVersion.value = 0
    wheelRotationDeg.value = target
    requestAnimationFrame(() => {
      spinVersion.value = previousVersion
    })
    return
  }
  const normalizedCurrent = ((wheelRotationDeg.value % 360) + 360) % 360
  wheelRotationDeg.value = wheelRotationDeg.value + 360 * 8 + target - normalizedCurrent
  spinVersion.value += 1
}

function wait(ms: number) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms)
  })
}

async function loadSummary() {
  loading.value = true
  try {
    const { data } = await paymentAPI.getRechargeActivitySummary()
    summary.value = data
    const firstPrize = enabledPrizes.value[0]
    if (firstPrize) alignWheelToPrize(firstPrize.id, false)
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function drawNow() {
  if (!activeChance.value) return
  drawing.value = true
  pointerTicking.value = true
  try {
    const { data } = await paymentAPI.drawRechargeActivity(activeChance.value.id)
    lastDrawRecord.value = data.record
    highlightedPrizeID.value = null
    alignWheelToPrize(data.record.prize_id, true)
    await wait(WHEEL_SPIN_DURATION_MS)
    highlightedPrizeID.value = data.record.prize_id
    pointerTicking.value = false
    resultModalOpen.value = true
    await loadSummary()
    appStore.showSuccess(t('rechargeActivity.toastWin', { name: data.record.prize_name }))
  } catch (err) {
    pointerTicking.value = false
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    pointerTicking.value = false
    drawing.value = false
  }
}

function fulfillmentStatusLabel(value?: string) {
  return value === 'fulfilled' ? t('rechargeActivity.fulfillmentFulfilled') : t('rechargeActivity.fulfillmentPending')
}

function fulfillmentBadgeClass(value?: string) {
  return value === 'fulfilled'
    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
    : 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
}

function formatOrderType(value: string) {
  return value === 'subscription' ? t('rechargeActivity.subscriptionOrder') : t('rechargeActivity.balanceOrder')
}

function formatDateTime(value?: string) {
  if (!value) return '--'
  return new Date(value).toLocaleString()
}

function closeResultModal() {
  resultModalOpen.value = false
}

onMounted(() => {
  loadSummary()
})
</script>

<style scoped>
[data-test='recharge-activity-pointer'] {
  transform-origin: 50% 0%;
}

[data-test='recharge-activity-pointer'].is-ticking {
  animation: recharge-activity-pointer-tick 140ms ease-in-out infinite alternate;
}

@keyframes recharge-activity-pointer-tick {
  from {
    transform: translateX(-50%) rotate(-10deg) scale(0.98);
  }

  to {
    transform: translateX(-50%) rotate(8deg) scale(1.02);
  }
}
</style>
