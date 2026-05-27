<template>
  <footer class="relative z-10 mt-10 bg-[#0a1020] px-4 pb-8 pt-8 text-white sm:px-6">
    <div class="mx-auto max-w-7xl rounded-[28px] border border-white/8 bg-slate-950/45 px-6 py-8 backdrop-blur-xl sm:px-8">
      <div class="grid gap-8 lg:grid-cols-[1.1fr_1.4fr_0.55fr]">
        <div>
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-white/10">
              <img :src="siteLogo || '/favicon.ico'" alt="Logo" class="h-full w-full object-contain p-1.5" />
            </div>
            <div>
              <p class="text-lg font-black">{{ brand }}</p>
              <p class="text-sm text-slate-300">{{ about }}</p>
            </div>
          </div>

          <div class="mt-5 flex gap-3 text-slate-400">
            <span
              v-for="dot in socialDots"
              :key="dot"
              class="inline-flex h-9 w-9 items-center justify-center rounded-full border border-white/10 bg-white/5 text-sm dark:border-cyan-400/10 dark:bg-slate-900/70"
            >
              {{ dot }}
            </span>
          </div>
        </div>

        <div class="grid gap-8 sm:grid-cols-4">
          <div v-for="column in columns" :key="column.title">
            <p class="text-sm font-black text-white">{{ column.title }}</p>
            <div class="mt-4 space-y-3">
              <a
                v-for="item in column.items"
                :key="item.label"
                :href="item.href"
                :target="item.external ? '_blank' : undefined"
                :rel="item.external ? 'noopener noreferrer' : undefined"
                class="block text-sm text-slate-300 transition-colors hover:text-white dark:hover:text-cyan-200"
              >
                {{ item.label }}
              </a>
            </div>
          </div>
        </div>

        <div v-if="contactItems.length" class="flex flex-col items-start lg:items-end">
          <p class="text-sm font-black text-white">{{ followLabel }}</p>
          <div class="mt-4 grid gap-3">
            <div
              v-for="item in contactItems"
              :key="item.label"
              class="min-w-[11rem] rounded-[20px] border border-white/10 bg-white/5 px-4 py-3 text-left dark:border-cyan-400/10 dark:bg-slate-900/70"
            >
              <p class="text-xs font-black uppercase tracking-[0.2em] text-slate-400">{{ item.label }}</p>
              <div class="mt-2 space-y-1">
                <p
                  v-for="value in item.values"
                  :key="value"
                  data-test="footer-contact-line"
                  class="break-all text-sm font-semibold leading-6 text-white"
                >
                  {{ value }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-8 flex flex-col items-center justify-center gap-3 border-t border-white/10 pt-6 text-center text-sm text-slate-300">
        <p>
          © 2018 {{ copyrightOwner }}
          <a
            href="https://beian.miit.gov.cn/"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-white hover:underline"
          >
            {{ filingLabel }}
          </a>
        </p>
        <a
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="font-semibold transition-colors hover:text-white"
        >
          {{ docsLabel }}
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps<{
  siteLogo: string
  brand: string
  about: string
  socialDots: string[]
  columns: Array<{ title: string; items: Array<{ label: string; href: string; external?: boolean }> }>
  followLabel: string
  contactItems: Array<{ label: string; values: string[] }>
  copyrightOwner: string
  filingLabel: string
  docUrl: string
  docsLabel: string
}>()
</script>
