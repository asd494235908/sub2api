<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <section class="relative overflow-hidden rounded-[32px] border border-amber-200/70 bg-[radial-gradient(circle_at_top,#fff7d6_0%,#fff4cf_18%,#ffffff_72%)] p-6 shadow-[0_35px_90px_rgba(245,158,11,0.22)]">
        <div class="absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_center,rgba(251,191,36,0.2),transparent_62%)]" />
        <div class="relative flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
          <div class="max-w-2xl">
            <p class="text-sm font-semibold uppercase tracking-[0.3em] text-amber-500">{{ t('luckyWheel.heroTag') }}</p>
            <h1 class="mt-3 font-serif text-4xl font-black text-slate-900">{{ t('nav.luckyWheel') }}</h1>
            <p class="mt-4 whitespace-pre-line text-sm leading-6 text-slate-600">{{ introText }}</p>

            <div class="mt-5 flex flex-wrap gap-3 text-xs font-semibold text-slate-600">
              <span class="rounded-full bg-white/80 px-4 py-2 shadow-sm">{{ t('luckyWheel.summaryStep') }} {{ summary?.config.multiplier_step?.toFixed(1) ?? '0.1' }}x</span>
              <span class="rounded-full bg-white/80 px-4 py-2 shadow-sm">{{ t('luckyWheel.summaryCap') }} {{ summary?.config.global_max_multiplier?.toFixed(1) ?? '3.0' }}x</span>
              <span class="rounded-full bg-white/80 px-4 py-2 shadow-sm">{{ t('luckyWheel.summarySettle') }}</span>
            </div>
          </div>

          <div class="grid min-w-[280px] grid-cols-2 gap-3">
            <div class="rounded-3xl border border-white/80 bg-white/85 px-5 py-4 shadow-sm">
              <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.remainingDraws') }}</div>
              <div class="mt-2 text-3xl font-black text-amber-500">{{ activeSession?.remaining_draws ?? 0 }}</div>
            </div>
            <div class="rounded-3xl border border-white/80 bg-white/85 px-5 py-4 shadow-sm">
              <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.bestMultiplier') }}</div>
              <div class="mt-2 text-3xl font-black text-emerald-500">{{ formatMultiplier(activeSession?.best_multiplier ?? 0) }}</div>
            </div>
            <button
              class="col-span-2 btn h-14 rounded-2xl border-none bg-gradient-to-r from-amber-500 via-orange-500 to-rose-500 text-base font-bold text-white shadow-[0_18px_40px_rgba(249,115,22,0.35)] hover:from-amber-600 hover:to-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="drawing || !summary?.enabled || !activeSession"
              data-test="draw-button"
              @click="draw"
            >
              <span v-if="drawing">{{ t('common.processing') }}</span>
              <span v-else>{{ t('luckyWheel.drawNow') }}</span>
            </button>
          </div>
        </div>
      </section>

      <div v-if="loading" class="flex justify-center py-16">
        <LoadingSpinner />
      </div>

      <template v-else>
        <div v-if="!summary?.enabled" class="rounded-[28px] border border-dashed border-slate-300 bg-white px-6 py-10 text-center text-sm text-slate-500">
          {{ t('luckyWheel.disabled') }}
        </div>

        <div v-else class="grid gap-6 xl:grid-cols-[1.2fr,0.8fr]">
          <section class="rounded-[32px] border border-amber-100 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
            <div class="mb-6 flex items-center justify-between gap-4">
              <div>
                <h2 class="text-xl font-black text-slate-900">{{ t('luckyWheel.currentSession') }}</h2>
                <p class="mt-1 text-sm text-slate-500">{{ activeSession ? t('luckyWheel.currentSessionHint') : t('luckyWheel.noActiveSession') }}</p>
              </div>
              <button class="btn btn-secondary btn-sm" @click="loadSummary">{{ t('luckyWheel.refresh') }}</button>
            </div>

            <div v-if="!activeSession" class="rounded-[28px] border border-dashed border-slate-300 px-6 py-12 text-center text-sm text-slate-500">
              {{ t('luckyWheel.noActiveSession') }}
            </div>

            <div v-else class="space-y-6">
              <div class="grid gap-6 lg:grid-cols-[1fr,0.85fr]">
                <div>
                  <div class="mx-auto flex max-w-[420px] items-center justify-center">
                    <div class="relative aspect-square w-full max-w-[380px]" data-test="wheel-board">
                      <div
                        data-test="wheel-rotor"
                        class="absolute inset-0 rounded-full transition-transform duration-[4600ms] [transition-timing-function:cubic-bezier(0.18,0.89,0.2,1)]"
                        :class="{ 'is-spinning': drawing }"
                        :style="wheelRotorStyle"
                      >
                        <div
                          class="absolute inset-0 rounded-full border-[18px] border-amber-100 shadow-[inset_0_0_0_10px_rgba(255,255,255,0.7),0_30px_60px_rgba(251,146,60,0.15)]"
                          :style="{ background: wheelGradient }"
                        />
                        <div class="absolute inset-[12%] rounded-full border border-white/60 bg-white/80 backdrop-blur">
                          <div
                            v-for="(segment, index) in wheelSegments"
                            :key="segment.key"
                            class="absolute left-1/2 top-1/2 origin-bottom -translate-x-1/2"
                            :style="segmentStyle(index)"
                          >
                            <span
                              class="block rounded-full px-2 py-1 text-[10px] font-bold shadow-sm"
                              :class="highlightedMultiplier === segment.multiplier ? 'bg-emerald-500 text-white' : 'bg-white/90 text-slate-700'"
                            >
                              {{ formatMultiplier(segment.multiplier) }}
                            </span>
                          </div>
                        </div>
                      </div>
                      <div
                        data-test="wheel-pointer-tip"
                        class="absolute left-1/2 top-[2px] z-10 -translate-x-1/2 -translate-y-[16px]"
                        :class="{ 'is-ticking': pointerTicking }"
                      >
                        <div
                          data-test="wheel-pointer-arrow"
                          class="h-0 w-0 border-x-[12px] border-t-[28px] border-x-transparent border-t-rose-500 drop-shadow-[0_10px_14px_rgba(244,63,94,0.28)]"
                        />
                      </div>
                      <div class="absolute inset-[36%] z-10 flex items-center justify-center rounded-full border-[10px] border-amber-200 bg-[radial-gradient(circle_at_top,#fde68a,#f59e0b_58%,#ea580c)] text-center shadow-[0_18px_35px_rgba(245,158,11,0.35)]">
                        <div>
                          <div class="text-xs font-bold uppercase tracking-[0.28em] text-amber-50 drop-shadow-[0_1px_2px_rgba(124,45,18,0.3)]">{{ t('luckyWheel.bestMultiplier') }}</div>
                          <div class="mt-2 text-3xl font-black text-white drop-shadow-[0_2px_8px_rgba(124,45,18,0.4)]">{{ formatMultiplier(activeSession.best_multiplier || activeSession.min_multiplier) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="space-y-4">
                  <div class="grid grid-cols-2 gap-3">
                    <div class="rounded-3xl bg-slate-50 p-4">
                      <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.rechargeAmount') }}</div>
                      <div class="mt-2 text-2xl font-black text-slate-900">{{ formatAmount(activeSession.source_pay_amount) }}</div>
                    </div>
                    <div class="rounded-3xl bg-slate-50 p-4">
                      <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.tierName') }}</div>
                      <div class="mt-2 text-2xl font-black text-slate-900">{{ activeSession.matched_tier_name }}</div>
                    </div>
                    <div class="rounded-3xl bg-slate-50 p-4">
                      <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.inviteBonus') }}</div>
                      <div class="mt-2 text-2xl font-black text-slate-900">+{{ formatMultiplier(activeSession.invite_bonus_multiplier) }}</div>
                    </div>
                    <div class="rounded-3xl bg-slate-50 p-4">
                      <div class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ t('luckyWheel.goldenBonus') }}</div>
                      <div class="mt-2 text-2xl font-black text-slate-900">+{{ activeSession.golden_window_extra_draws }}</div>
                    </div>
                  </div>

                  <div class="rounded-[28px] border border-slate-100 bg-slate-50/70 p-4">
                    <div class="flex items-center justify-between text-sm font-semibold text-slate-600">
                      <span>{{ t('luckyWheel.progressTitle') }}</span>
                      <span>{{ activeSession.completed_draws }}/{{ activeSession.total_draws }}</span>
                    </div>
                    <div class="mt-3 h-3 overflow-hidden rounded-full bg-white">
                      <div class="h-full rounded-full bg-gradient-to-r from-amber-400 to-rose-500" :style="{ width: `${progressPercent}%` }" />
                    </div>
                  </div>

                  <div class="rounded-[28px] border border-slate-100 bg-white p-4">
                    <h3 class="text-sm font-bold uppercase tracking-[0.24em] text-slate-500">{{ t('luckyWheel.drawHistory') }}</h3>
                    <div v-if="!activeSession.draw_records?.length" class="py-6 text-sm text-slate-400">{{ t('luckyWheel.noDrawHistory') }}</div>
                    <div v-else class="mt-4 space-y-3">
                      <div
                        v-for="record in activeSession.draw_records"
                        :key="record.id"
                        class="flex items-center justify-between rounded-2xl border border-slate-100 bg-slate-50/70 px-4 py-3"
                      >
                        <div>
                          <div class="text-sm font-semibold text-slate-900">{{ t('luckyWheel.drawIndexLabel', { value: record.draw_index }) }}</div>
                          <div class="mt-1 text-xs text-slate-500">
                            {{ formatMultiplier(record.base_multiplier) }} + {{ formatMultiplier(record.invite_bonus_multiplier) }}
                          </div>
                        </div>
                        <div class="text-right">
                          <div class="text-lg font-black text-emerald-500">{{ formatMultiplier(record.final_multiplier) }}</div>
                          <div class="text-xs text-slate-400">{{ formatDateTime(record.created_at) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="space-y-6">
            <div class="rounded-[32px] border border-slate-100 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
              <h2 class="text-xl font-black text-slate-900">{{ rulesTitle }}</h2>
              <div class="mt-4 space-y-3 text-sm leading-6 text-slate-600">
                <p v-for="rule in rulesItems" :key="rule">{{ rule }}</p>
              </div>
            </div>

            <div class="rounded-[32px] border border-slate-100 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.06)]">
              <div class="flex items-center justify-between">
                <h2 class="text-xl font-black text-slate-900">{{ t('luckyWheel.historyTitle') }}</h2>
                <span class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ historySessions.length }}</span>
              </div>
              <div v-if="!historySessions.length" class="py-10 text-center text-sm text-slate-400">{{ t('luckyWheel.noRecords') }}</div>
              <div v-else data-test="history-scroll-container" class="mt-4 max-h-[26rem] overflow-y-auto pr-1">
                <div class="space-y-3">
                  <div
                    v-for="session in historySessions"
                    :key="session.id"
                    class="rounded-3xl border border-slate-100 bg-slate-50/80 p-4"
                  >
                    <div class="flex items-start justify-between gap-4">
                      <div>
                        <div class="text-sm font-bold text-slate-900">{{ session.matched_tier_name }} · {{ formatAmount(session.source_pay_amount) }}</div>
                        <div class="mt-2 text-xs text-slate-500">
                          {{ t('luckyWheel.historyDraws', { value: `${session.completed_draws}/${session.total_draws}` }) }}
                        </div>
                      </div>
                      <div class="text-right">
                        <div class="text-lg font-black text-emerald-500">{{ formatMultiplier(session.best_multiplier) }}</div>
                        <div class="mt-1 text-xs text-slate-400">{{ formatAmount(session.settled_bonus_amount ?? 0) }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </template>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="resultModalOpen && lastDrawResult" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/65 p-4" data-test="result-modal" @click.self="closeResultModal">
          <div class="w-full max-w-md rounded-[32px] border border-amber-200 bg-white p-6 text-center shadow-[0_35px_90px_rgba(15,23,42,0.25)]">
            <div class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-[radial-gradient(circle_at_top,#fef3c7,#f59e0b)] text-4xl text-white shadow-[0_18px_35px_rgba(245,158,11,0.35)]">✦</div>
            <p class="mt-4 text-sm font-semibold uppercase tracking-[0.24em] text-amber-500">{{ t('luckyWheel.resultTitle') }}</p>
            <h2 class="mt-3 text-4xl font-black text-slate-900">{{ formatMultiplier(lastDrawResult.best_multiplier) }}</h2>
            <p class="mt-3 text-sm text-slate-500">{{ t('luckyWheel.resultHint') }}</p>
            <p class="mt-4 text-2xl font-black text-emerald-500">{{ formatAmount(lastDrawResult.settled_bonus_amount ?? 0) }}</p>
            <button class="btn mt-6 w-full rounded-2xl border-none bg-gradient-to-r from-amber-500 to-rose-500 text-white hover:from-amber-600 hover:to-rose-600" data-test="result-confirm" @click="closeResultModal">
              {{ t('luckyWheel.resultConfirm') }}
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
import type { LuckyWheelDrawResult, LuckyWheelSummary } from '@/types/payment'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const drawing = ref(false)
const summary = ref<LuckyWheelSummary | null>(null)
const lastDrawResult = ref<LuckyWheelDrawResult | null>(null)
const resultModalOpen = ref(false)
const highlightedMultiplier = ref<number | null>(null)
const wheelRotationDeg = ref(0)
const spinVersion = ref(0)
const pointerTicking = ref(false)

const WHEEL_SPIN_DURATION_MS = 4600
const POINTER_TARGET_DEG = 0

const activeSession = computed(() => summary.value?.active_session ?? null)
const historySessions = computed(() => summary.value?.history_sessions ?? [])
const introText = computed(() => summary.value?.config.intro_text?.trim() || t('luckyWheel.heroDescription'))
const rulesTitle = computed(() => summary.value?.config.rules_title?.trim() || t('luckyWheel.rulesTitle'))
const rulesItems = computed(() => {
  const items = (summary.value?.config.rules_items ?? []).map((item) => item.trim()).filter(Boolean)
  if (items.length > 0) return items
  return [
    t('luckyWheel.ruleTier20To50'),
    t('luckyWheel.ruleTier51Plus'),
    t('luckyWheel.ruleHighestWins'),
    t('luckyWheel.ruleInviteBonus'),
    t('luckyWheel.ruleGoldenWindow'),
  ]
})

const wheelSegments = computed(() => {
  const session = activeSession.value
  const config = summary.value?.config
  if (!session || !config) return []
  const segments: Array<{ key: string; multiplier: number }> = []
  const step = config.multiplier_step || 0.1
  let current = session.min_multiplier
  while (current <= session.max_multiplier + 1e-9) {
    const multiplier = Number(current.toFixed(2))
    segments.push({ key: `${session.id}-${multiplier}`, multiplier })
    current += step
  }
  return segments
})

const progressPercent = computed(() => {
  const session = activeSession.value
  if (!session || session.total_draws <= 0) return 0
  return Math.min(100, Math.round((session.completed_draws / session.total_draws) * 100))
})

const wheelGradient = computed(() => {
  const segments = wheelSegments.value
  if (!segments.length) return 'conic-gradient(#fde68a, #fdba74, #fca5a5)'
  const palette = ['#f59e0b', '#fb7185', '#f97316', '#fbbf24', '#fda4af', '#fdba74']
  const step = 100 / segments.length
  const stops = segments.map((_, index) => {
    const color = palette[index % palette.length]
    const start = (index * step).toFixed(2)
    const end = ((index + 1) * step).toFixed(2)
    return `${color} ${start}% ${end}%`
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

function getWheelTargetRotation(multiplier: number) {
  const segments = wheelSegments.value
  if (!segments.length) return wheelRotationDeg.value
  const targetIndex = segments.findIndex((segment) => Math.abs(segment.multiplier - multiplier) < 0.001)
  if (targetIndex < 0) return wheelRotationDeg.value
  const segmentAngle = 360 / segments.length
  const pointerOffset = targetIndex * segmentAngle
  return POINTER_TARGET_DEG - pointerOffset
}

function alignWheelToMultiplier(multiplier: number, animated: boolean) {
  const target = getWheelTargetRotation(multiplier)
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
  const extraTurns = 360 * 8
  wheelRotationDeg.value = wheelRotationDeg.value + extraTurns + target - normalizedCurrent
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
    const { data } = await paymentAPI.getLuckyWheelSummary()
    summary.value = data
    highlightedMultiplier.value = data.active_session?.best_multiplier ?? null
    if (data.active_session) {
      alignWheelToMultiplier(data.active_session.best_multiplier || data.active_session.min_multiplier, false)
    }
  } catch (err) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function draw() {
  if (!activeSession.value) return
  drawing.value = true
  pointerTicking.value = true
  try {
    const { data } = await paymentAPI.drawLuckyWheel(activeSession.value.id)
    lastDrawResult.value = data
    highlightedMultiplier.value = null
    alignWheelToMultiplier(data.draw_record.final_multiplier, true)
    await wait(WHEEL_SPIN_DURATION_MS)
    highlightedMultiplier.value = data.draw_record.final_multiplier
    pointerTicking.value = false
    if (data.session) {
      summary.value = {
        ...(summary.value as LuckyWheelSummary),
        active_session: data.settled ? null : data.session,
        pending_sessions: data.settled
          ? (summary.value?.pending_sessions ?? []).filter((item) => item.id !== data.session_id)
          : [data.session, ...(summary.value?.pending_sessions ?? []).filter((item) => item.id !== data.session_id)],
        history_sessions: data.settled
          ? [data.session, ...(summary.value?.history_sessions ?? [])]
          : (summary.value?.history_sessions ?? []),
      }
    }
    if (data.settled) {
      resultModalOpen.value = true
    }
    await loadSummary()
    appStore.showSuccess(
      data.settled
        ? t('luckyWheel.toastSettled', { amount: (data.settled_bonus_amount ?? 0).toFixed(2) })
        : t('luckyWheel.toastDrawn', { amount: data.draw_record.final_multiplier.toFixed(1) }),
    )
  } catch (err) {
    pointerTicking.value = false
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    pointerTicking.value = false
    drawing.value = false
  }
}

function formatMultiplier(value: number) {
  return `${Number(value || 0).toFixed(1)}x`
}

function formatAmount(value: number) {
  return Number(value || 0).toFixed(2)
}

function formatDateTime(value?: string | null) {
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
[data-test='wheel-pointer-tip'] {
  transform-origin: 50% 0%;
}

[data-test='wheel-pointer-tip'].is-ticking {
  animation: wheel-pointer-tick 140ms ease-in-out infinite alternate;
}

@keyframes wheel-pointer-tick {
  from {
    transform: translateX(-50%) rotate(-10deg) scale(0.98);
  }

  to {
    transform: translateX(-50%) rotate(8deg) scale(1.02);
  }
}
</style>
