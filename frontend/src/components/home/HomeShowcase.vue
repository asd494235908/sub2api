<template>
  <section class="mx-auto mt-10 max-w-7xl" :style="sectionOffsetStyle">
    <div class="mb-10 text-center">
      <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ kicker }}</p>
      <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">{{ title }}</h2>
      <p class="mx-auto mt-4 max-w-3xl text-sm leading-7 text-slate-500 dark:text-slate-300 sm:text-base">{{ subtitle }}</p>
    </div>

    <div class="grid gap-5 lg:grid-cols-3">
      <article
        v-for="stat in stats"
        :key="stat.label"
        class="rounded-[2rem] border border-slate-200/80 bg-white/92 p-6 text-center shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/78"
      >
        <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-slate-950 text-white dark:bg-cyan-300 dark:text-slate-950">
          <Icon :name="stat.icon" size="lg" />
        </div>
        <p class="mt-5 text-4xl font-black tracking-tight text-slate-950 dark:text-white">{{ stat.value }}</p>
        <p class="mt-2 text-sm font-medium text-slate-500 dark:text-slate-300">{{ stat.label }}</p>
      </article>
    </div>

    <div class="mt-5">
      <div class="grid gap-5 md:grid-cols-3">
        <article
          v-for="quote in quotes"
          :key="quote.author"
          class="rounded-[2rem] border border-slate-200/80 bg-white/92 p-6 shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/78"
        >
          <div class="flex gap-1">
            <span v-for="star in 5" :key="star" class="text-yellow-500">★</span>
          </div>
          <p class="mt-4 text-sm leading-7 text-slate-600 dark:text-slate-300">“{{ quote.quote }}”</p>
          <div class="mt-5">
            <p class="text-sm font-black text-slate-950 dark:text-white">{{ quote.author }}</p>
            <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">{{ quote.role }}</p>
          </div>
        </article>
      </div>

    </div>
  </section>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  kicker: string
  title: string
  subtitle: string
  sectionOffsetStyle: Record<string, string>
  stats: Array<{ icon: 'users' | 'shield' | 'sparkles'; value: string; label: string }>
  quotes: Array<{ quote: string; author: string; role: string }>
}>()
</script>
