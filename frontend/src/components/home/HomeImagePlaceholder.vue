<template>
  <div
    class="home-placeholder relative overflow-hidden rounded-[2rem] border border-slate-200/80 bg-white/90 p-5 shadow-[0_24px_80px_rgba(15,23,42,0.08)] dark:border-slate-800/80 dark:bg-slate-950/80"
    :class="[aspectClass, compact ? 'min-h-[220px]' : 'min-h-[320px]']"
  >
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.12),transparent_34%),radial-gradient(circle_at_bottom_right,rgba(14,165,233,0.14),transparent_30%)] dark:bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.16),transparent_32%),radial-gradient(circle_at_bottom_right,rgba(59,130,246,0.18),transparent_28%)]"></div>
      <div class="absolute inset-4 rounded-[1.5rem] border border-dashed border-slate-300/80 dark:border-slate-700/80"></div>
      <div class="absolute left-6 right-6 top-8 h-px bg-gradient-to-r from-transparent via-slate-300/80 to-transparent dark:via-slate-700/80"></div>
      <div class="absolute bottom-10 left-8 h-24 w-24 rounded-full bg-blue-200/40 blur-3xl dark:bg-cyan-400/15"></div>
      <div class="absolute right-6 top-10 grid gap-3">
        <span class="h-3 w-16 rounded-full bg-slate-200 dark:bg-slate-800"></span>
        <span class="h-3 w-10 rounded-full bg-slate-200/80 dark:bg-slate-800/80"></span>
        <span class="h-3 w-20 rounded-full bg-slate-200/70 dark:bg-slate-800/70"></span>
      </div>
      <div class="absolute bottom-8 left-8 right-8 grid gap-3 sm:grid-cols-3">
        <span class="h-16 rounded-2xl border border-slate-200/80 bg-white/70 dark:border-slate-800/80 dark:bg-slate-900/70"></span>
        <span class="h-16 rounded-2xl border border-slate-200/80 bg-white/70 dark:border-slate-800/80 dark:bg-slate-900/70"></span>
        <span class="h-16 rounded-2xl border border-slate-200/80 bg-white/70 dark:border-slate-800/80 dark:bg-slate-900/70"></span>
      </div>
    </div>

    <div class="relative z-10 flex h-full flex-col justify-between">
      <div class="max-w-sm">
        <span
          v-if="badge"
          class="inline-flex rounded-full border border-slate-200 bg-white/80 px-3 py-1 text-[11px] font-black uppercase tracking-[0.18em] text-slate-500 dark:border-slate-700/80 dark:bg-slate-900/80 dark:text-slate-300"
        >
          {{ badge }}
        </span>
        <h3 class="mt-4 text-2xl font-black tracking-tight text-slate-950 dark:text-white">
          {{ title }}
        </h3>
        <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-slate-300">
          {{ description }}
        </p>
      </div>

      <div class="pt-10">
        <span class="inline-flex rounded-full bg-slate-950 px-3 py-1.5 text-xs font-semibold text-white dark:bg-cyan-300 dark:text-slate-950">
          {{ footerLabel }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  badge?: string
  title: string
  description: string
  footerLabel: string
  aspectClass?: string
  compact?: boolean
}>(), {
  badge: '',
  aspectClass: 'aspect-[4/3]',
  compact: false
})
</script>
