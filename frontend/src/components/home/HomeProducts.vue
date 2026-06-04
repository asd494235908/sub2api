<template>
  <section class="overflow-hidden bg-gradient-to-b from-white to-zinc-50 py-32 dark:from-[#040816] dark:to-[#071022]" :style="sectionOffsetStyle">
    <div class="relative mx-auto mb-48 max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="relative z-10 mb-16 text-center">
        <span class="mb-6 inline-block text-sm font-semibold uppercase tracking-widest text-zinc-500 dark:text-slate-400">{{ kicker }}</span>
        <p class="mx-auto mb-8 max-w-3xl text-xl font-medium tracking-wide text-zinc-600 dark:text-slate-300">
          {{ subtitle }}
        </p>
        <h2 class="flex flex-wrap items-center justify-center gap-6 font-serif text-5xl leading-tight tracking-tight text-zinc-900 dark:text-white md:text-7xl lg:text-8xl">
          {{ title }}
          <span class="inline-block relative">
            <div class="group relative mx-4 h-56 w-56 md:h-72 md:w-72 lg:h-96 lg:w-96">
              <div class="absolute inset-0 rounded-full border-2 border-dashed border-zinc-200 opacity-60 animate-[spin_40s_linear_infinite] dark:border-slate-700"></div>
              <div class="absolute inset-[15%] rounded-full border border-zinc-100 animate-[spin_25s_linear_infinite_reverse] dark:border-slate-800">
                <div class="absolute left-1/2 top-0 h-2 w-2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-blue-500 shadow-[0_0_10px_#3b82f6]"></div>
                <div class="absolute bottom-0 left-1/2 h-2 w-2 -translate-x-1/2 translate-y-1/2 rounded-full bg-zinc-300 dark:bg-slate-600"></div>
              </div>
              <div class="absolute inset-[25%] flex items-center justify-center overflow-hidden rounded-full border border-zinc-200 bg-white shadow-[0_0_50px_rgba(0,0,0,0.03)] transition-transform duration-700 group-hover:scale-105 dark:border-slate-700 dark:bg-slate-900">
                <div
                  class="absolute inset-0 opacity-[0.05]"
                  style="background-image: linear-gradient(to right, #000 1px, transparent 1px), linear-gradient(to bottom, #000 1px, transparent 1px); background-size: 25px 25px;"
                ></div>
                <svg viewBox="0 0 800 400" class="absolute inset-0 h-full w-full scale-150 object-contain opacity-[0.08] animate-[pulse_4s_ease-in-out_infinite]">
                  <path fill="currentColor" d="M150,150 L160,140 L180,145 L190,160 L180,175 L160,170 Z M300,100 L320,90 L350,110 L340,130 L310,120 Z M500,200 L530,190 L560,210 L540,240 L510,230 Z M650,80 L680,70 L700,90 L670,110 Z M200,300 L230,290 L250,310 L220,330 Z" />
                </svg>
                <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
                  <div class="absolute h-full w-full rounded-full border border-blue-400/20 animate-[ping_3s_linear_infinite]"></div>
                  <div class="absolute h-3/4 w-3/4 rounded-full border border-blue-400/10 animate-[ping_4s_linear_infinite_1s]"></div>
                </div>
                <div class="relative z-10 rounded-full border border-zinc-100 bg-white p-5 shadow-sm dark:border-slate-700 dark:bg-slate-950">
                  <Icon name="globe" size="xl" class="text-zinc-800 dark:text-white" />
                </div>
              </div>
            </div>
          </span>
        </h2>
      </div>

      <div class="mx-auto mt-24 grid max-w-5xl grid-cols-1 gap-8 md:grid-cols-2">
        <article
          v-for="card in cards.slice(0, 2)"
          :key="card.title"
          class="rounded-3xl border-2 border-zinc-200/60 bg-white p-8 shadow-lg transition-all duration-500 hover:border-zinc-300 hover:shadow-2xl dark:border-slate-800/80 dark:bg-slate-900/78 dark:hover:border-slate-600"
        >
          <div class="flex gap-5">
            <div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl border border-blue-200/50 bg-gradient-to-br from-blue-100 to-blue-50 text-blue-600 dark:border-slate-700 dark:from-slate-800 dark:to-slate-900 dark:text-cyan-200">
              <Icon :name="card.icon" size="lg" />
            </div>
            <div>
              <div class="mb-3 inline-block rounded-lg bg-black px-3 py-1 text-xs font-bold text-white dark:bg-cyan-300 dark:text-slate-950">
                {{ card.title }}
              </div>
              <p class="text-sm leading-relaxed text-zinc-700 dark:text-slate-300">
                {{ card.description }}
              </p>
            </div>
          </div>
        </article>
      </div>
    </div>

    <div class="relative mx-auto max-w-7xl border-t-2 border-zinc-200/60 bg-gradient-to-b from-zinc-50 to-white px-4 py-24 sm:px-6 lg:px-8 dark:border-slate-800 dark:from-[#071022] dark:to-[#040816]">
      <div class="relative z-10 text-center">
        <div class="mb-8 inline-flex h-20 w-20 items-center justify-center rounded-2xl bg-black text-white shadow-2xl transition-all duration-500 hover:rotate-6 hover:scale-110 dark:bg-cyan-300 dark:text-slate-950">
          <Icon name="sparkles" size="xl" />
        </div>
        <h3 class="mb-4 font-serif text-4xl font-bold leading-tight text-zinc-900 dark:text-white md:text-5xl">{{ sideTitle }}</h3>
        <p class="mb-12 text-xl font-medium text-zinc-600 dark:text-slate-300">{{ sideSubtitle }}</p>

        <div class="mx-auto grid w-full max-w-5xl grid-cols-2 justify-items-center gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
          <div
            v-for="(item, index) in items"
            :key="item.name"
            class="flex min-h-16 w-full max-w-[13rem] items-center justify-center rounded-2xl px-4 py-3 text-center text-sm font-semibold leading-snug shadow-sm transition-colors sm:text-base"
            :class="index === 0
              ? 'bg-black text-white shadow-lg dark:bg-cyan-300 dark:text-slate-950'
              : 'border-2 border-zinc-200 bg-white text-black dark:border-slate-700 dark:bg-slate-950 dark:text-white'"
          >
            {{ item.name }}
          </div>
        </div>
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
  sideKicker: string
  sideTitle: string
  sideSubtitle: string
  sectionOffsetStyle: Record<string, string>
  cards: Array<{ icon: 'cpu' | 'sparkles' | 'globe' | 'shield'; title: string; description: string }>
  items: Array<{ name: string; description: string; badge: string }>
}>()
</script>
