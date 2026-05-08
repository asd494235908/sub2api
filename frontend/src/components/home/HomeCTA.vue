<template>
  <section class="mx-auto mt-10 max-w-7xl" :style="sectionOffsetStyle">
    <div class="relative overflow-hidden rounded-[2.5rem] bg-[linear-gradient(135deg,#020617,#0f172a_46%,#111827)] px-6 py-10 text-white shadow-[0_30px_100px_rgba(2,6,23,0.36)] sm:px-10 sm:py-14">
      <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.28),transparent_30%),radial-gradient(circle_at_bottom_right,rgba(34,211,238,0.22),transparent_28%)]"></div>

      <div class="relative z-10 max-w-3xl">
        <span class="inline-flex rounded-full border border-white/15 bg-white/10 px-4 py-2 text-sm font-semibold text-cyan-100">
          {{ badge }}
        </span>
        <h2 class="mt-6 text-4xl font-black tracking-tight sm:text-5xl">
          {{ title }}
        </h2>
        <p class="mt-5 max-w-2xl text-base leading-8 text-slate-300 sm:text-lg">
          {{ subtitle }}
        </p>

        <div class="mt-8 flex flex-col gap-4 sm:flex-row">
          <router-link
            :to="consolePath"
            class="inline-flex items-center justify-center rounded-full bg-white px-8 py-4 text-base font-black text-slate-950 transition-transform hover:-translate-y-0.5 hover:bg-slate-100"
          >
            {{ primaryLabel }}
          </router-link>
          <a
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center justify-center rounded-full border border-white/20 bg-white/5 px-8 py-4 text-base font-bold text-white transition-transform hover:-translate-y-0.5 hover:bg-white/10"
          >
            {{ secondaryLabel }}
          </a>
        </div>

        <div class="mt-8 flex flex-wrap gap-3 text-sm text-slate-300">
          <span
            v-for="item in trustItems"
            :key="item"
            class="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-2"
          >
            <span class="text-cyan-300">●</span>
            {{ item }}
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  badge: string
  title: string
  subtitle: string
  consolePath: string
  primaryLabel: string
  secondaryLabel: string
  docUrl: string
  trustItems: string[]
  sectionOffsetStyle: Record<string, string>
}>()
</script>
