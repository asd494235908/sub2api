<template>
  <section
    id="overview"
    class="relative overflow-hidden bg-gradient-to-b from-white via-zinc-50/30 to-white px-4 py-32 sm:px-6 lg:px-8 dark:from-[#040816] dark:via-[#071022] dark:to-[#040816]"
    :style="sectionOffsetStyle"
  >
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(15,23,42,0.02),transparent_65%)] dark:bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.02),transparent_65%)]"></div>

    <div class="relative z-10 mx-auto max-w-7xl">
      <div class="mb-20 text-center">
        <span class="inline-block mb-4 text-sm font-semibold uppercase tracking-widest text-zinc-500 dark:text-slate-400">
          {{ kicker }}
        </span>
        <h2 class="mb-6 font-serif text-5xl leading-tight text-black dark:text-white md:text-6xl">
          {{ title }}
        </h2>
        <p class="mx-auto max-w-2xl text-lg text-zinc-600 dark:text-slate-300">
          {{ subtitle }}
        </p>
      </div>

      <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
        <article
          v-for="(card, index) in featureCards"
          :key="card.title"
          class="group relative overflow-hidden rounded-[2.5rem] border p-8 transition-all duration-500 hover:shadow-2xl"
          :class="index === 0
            ? 'md:col-span-2 border-zinc-100 bg-zinc-50 hover:border-zinc-200 dark:border-slate-800 dark:bg-slate-900/70 dark:hover:border-slate-700'
            : index === 1
              ? 'border-zinc-100 bg-white hover:border-zinc-200 dark:border-slate-800 dark:bg-slate-900/70 dark:hover:border-slate-700'
              : 'border-zinc-100 bg-zinc-900 text-white hover:border-zinc-800 dark:border-slate-800'"
        >
          <div class="relative z-10" :class="index === 0 ? 'max-w-md' : ''">
            <div
              class="mb-6 flex h-14 w-14 items-center justify-center rounded-2xl"
              :class="index === 0
                ? 'bg-black text-white dark:bg-cyan-300 dark:text-slate-950'
                : index === 1
                  ? 'bg-zinc-100 text-black dark:bg-slate-800 dark:text-white'
                  : 'bg-white/10 text-white'"
            >
              <Icon :name="card.icon" size="lg" />
            </div>
            <h3 class="mb-4 text-3xl font-bold leading-tight" :class="index === 2 ? 'text-white' : 'text-black dark:text-white'">
              {{ card.title }}
            </h3>
            <p class="text-sm leading-relaxed" :class="index === 2 ? 'text-zinc-300' : 'text-zinc-600 dark:text-slate-300'">
              {{ card.description }}
            </p>
          </div>

          <div
            v-if="index === 0"
            class="pointer-events-none absolute right-[-10%] top-10 h-[300px] w-[300px] rounded-full bg-blue-500/5 blur-[80px] dark:bg-cyan-400/10"
          ></div>
          <div
            v-else-if="index === 1"
            class="absolute bottom-0 left-0 right-0 p-8"
          >
            <div class="flex flex-col gap-2">
              <div
                v-for="bar in [45, 72, 88]"
                :key="bar"
                class="h-2 overflow-hidden rounded-full bg-zinc-50 dark:bg-slate-800"
              >
                <div class="h-full bg-zinc-200 dark:bg-slate-600" :style="{ width: `${bar}%` }"></div>
              </div>
            </div>
          </div>
          <div
            v-else
            class="mt-8 rounded-2xl border border-white/10 bg-white/5 p-6 font-mono text-xs text-zinc-300"
          >
            <code class="block leading-7">
              import OpenAI from "openai"<br />
              const client = new OpenAI()<br />
              await client.responses.create()
            </code>
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
  featureCards: Array<{ icon: 'shield' | 'swap' | 'chartBar'; title: string; description: string; emphasis?: boolean }>
}>()
</script>
