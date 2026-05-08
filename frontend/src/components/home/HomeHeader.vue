<template>
  <header
    ref="headerRef"
    class="fixed inset-x-0 top-0 z-50 transition-all duration-300"
    :class="scrolled
      ? 'border-b border-zinc-100/80 bg-white/80 py-3 backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/84'
      : 'bg-transparent py-5'"
  >
    <nav class="mx-auto max-w-[1400px] px-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-10">
          <a href="#top" class="flex items-center gap-3">
            <div class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-white ring-1 ring-slate-200/70 dark:bg-slate-900/90 dark:ring-slate-700/80">
              <img :src="siteLogo || '/favicon.ico'" alt="Logo" class="h-full w-full object-contain p-1" />
            </div>
            <span class="text-lg font-bold tracking-tighter text-black dark:text-white">{{ brand }}</span>
          </a>

          <div class="hidden items-center space-x-10 lg:flex">
            <a
              v-for="item in navItems"
              :key="item.label"
              :href="item.href"
              :target="item.external ? '_blank' : undefined"
              :rel="item.external ? 'noopener noreferrer' : undefined"
              class="text-[13px] font-medium text-zinc-500 transition-colors hover:text-black dark:text-slate-300 dark:hover:text-white"
            >
              {{ item.label }}
            </a>
          </div>
        </div>

        <div class="hidden items-center space-x-6 lg:flex">
          <LocaleSwitcher />

          <button
            @click="$emit('toggle-theme')"
            class="inline-flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 bg-white/70 text-zinc-500 transition-colors hover:text-black dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:text-white"
            :title="themeTitle"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>

          <router-link
            :to="consolePath"
            class="rounded-full bg-black px-6 py-2 text-[13px] font-medium text-white shadow-sm transition-all hover:bg-zinc-800 active:scale-95 dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950 dark:hover:bg-cyan-200"
          >
            {{ consoleLabel }}
          </router-link>
        </div>

        <div class="lg:hidden">
          <button
            @click="mobileMenuOpen = !mobileMenuOpen"
            class="rounded-full border border-zinc-200 bg-white/80 p-2 text-zinc-600 transition-colors hover:text-black dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:text-white"
          >
            <Icon v-if="mobileMenuOpen" name="x" size="md" />
            <Icon v-else name="menu" size="md" />
          </button>
        </div>
      </div>
    </nav>

    <div
      v-if="mobileMenuOpen"
      class="absolute left-0 right-0 top-full border-b border-zinc-100 bg-white shadow-xl animate-in slide-in-from-top duration-300 dark:border-slate-800/80 dark:bg-slate-950"
    >
      <div class="flex flex-col space-y-4 px-6 py-4">
        <a
          v-for="item in navItems"
          :key="`mobile-${item.label}`"
          :href="item.href"
          :target="item.external ? '_blank' : undefined"
          :rel="item.external ? 'noopener noreferrer' : undefined"
          class="border-b border-zinc-50 py-2 text-[15px] font-medium text-zinc-600 hover:text-black dark:border-slate-900 dark:text-slate-300 dark:hover:text-white"
          @click="mobileMenuOpen = false"
        >
          {{ item.label }}
        </a>

        <div class="flex items-center gap-3 pt-4">
          <LocaleSwitcher />
          <button
            @click="$emit('toggle-theme')"
            class="inline-flex h-10 w-10 items-center justify-center rounded-full border border-zinc-200 bg-white text-zinc-500 dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300"
            :title="themeTitle"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
        </div>

        <router-link
          :to="consolePath"
          class="w-full rounded-full bg-black px-6 py-3 text-center text-[15px] font-medium text-white dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950"
          @click="mobileMenuOpen = false"
        >
          {{ consoleLabel }}
        </router-link>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  siteLogo: string
  logoTagline: string
  brand: string
  navItems: Array<{ label: string; href: string; external?: boolean }>
  themeTitle: string
  isDark: boolean
  consolePath: string
  consoleLabel: string
}>()

defineEmits<{
  (event: 'toggle-theme'): void
}>()

const headerRef = defineModel<HTMLElement | null>('headerRef', { default: null })
const mobileMenuOpen = ref(false)
const scrolled = ref(false)

function handleScroll() {
  scrolled.value = window.scrollY > 10
}

onMounted(() => {
  handleScroll()
  window.addEventListener('scroll', handleScroll)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>
