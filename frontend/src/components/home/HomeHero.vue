<template>
  <section
    data-test="home-hero"
    class="relative flex flex-col items-center justify-between overflow-hidden px-4 pb-[7vh] pt-[max(3rem,6vh)] sm:px-6 lg:px-8"
    :style="heroStyle"
  >
    <div class="absolute inset-0 pointer-events-none"></div>

    <div class="absolute inset-0 pointer-events-none overflow-hidden">
      <div
        class="absolute inset-0 opacity-[0.03] dark:opacity-[0.06]"
        style="background-image: linear-gradient(to right, #000 1px, transparent 1px), linear-gradient(to bottom, #000 1px, transparent 1px); background-size: 60px 60px;"
      ></div>

      <div
        data-test="hero-scan-light"
        class="hero-scan-light absolute inset-x-0 top-[-22%] h-[300px] bg-gradient-to-b from-transparent via-blue-500/[0.03] to-transparent dark:via-cyan-400/[0.06]"
      ></div>

      <div
        v-for="stream in heroStreams"
        :key="stream.id"
        data-test="hero-stream"
        class="hero-stream absolute left-0 top-0 h-[420px] w-px origin-top rotate-45 bg-gradient-to-b from-transparent via-zinc-400/80 to-transparent dark:via-cyan-200/70"
        :style="stream.style"
      ></div>

      <div class="hero-glow hero-glow-left absolute left-[10%] top-[-10%] h-[800px] w-[800px] rounded-full bg-blue-50/25 dark:bg-cyan-500/12"></div>
      <div class="hero-glow hero-glow-right absolute bottom-[-10%] right-[10%] h-[900px] w-[900px] rounded-full bg-zinc-100/40 dark:bg-blue-500/12"></div>

      <div
        v-for="particle in heroParticles"
        :key="particle.id"
        data-test="hero-particle"
        class="hero-particle absolute h-[1.5px] w-[1.5px] rounded-full bg-zinc-300 dark:bg-cyan-100/70"
        :style="particle.style"
      ></div>
    </div>

    <div class="relative z-10 mx-auto flex w-full max-w-7xl flex-1 flex-col items-center justify-center gap-10 py-6 lg:min-h-0 lg:flex-row lg:items-center lg:gap-16 lg:py-10">
      <div class="flex w-full flex-1 flex-col justify-center text-center lg:text-left">
        <div class="mb-8 flex flex-col items-center justify-center gap-5 md:flex-row lg:justify-start">
          <div class="hero-logo-float relative">
            <div class="absolute inset-0 rounded-full bg-blue-400/10 blur-2xl dark:bg-cyan-400/18"></div>
            <div class="relative flex h-16 w-16 items-center justify-center rounded-full border border-zinc-200/80 bg-white shadow-sm dark:border-slate-700/80 dark:bg-slate-900">
              <span class="text-lg font-black tracking-tight text-black dark:text-white">AI</span>
            </div>
          </div>
          <h1 class="flex min-h-[1.15em] items-center text-6xl font-serif tracking-[-0.08em] text-black sm:text-7xl md:text-8xl lg:text-[clamp(5.5rem,8vw,8rem)] dark:text-white">
            <span class="inline-block bg-gradient-to-b from-black to-zinc-700 bg-clip-text text-transparent dark:from-white dark:to-slate-300">
              {{ titlePrefix }}
            </span>
          </h1>
        </div>

        <p class="mb-5 text-[12px] font-semibold uppercase tracking-[0.28em] text-zinc-400 dark:text-slate-400">
          {{ badge }}
        </p>

        <h2 class="mx-auto max-w-4xl text-4xl font-black leading-[1.04] tracking-[-0.055em] text-zinc-900 sm:text-5xl lg:mx-0 lg:max-w-3xl lg:text-[clamp(3.6rem,4.8vw,5.8rem)] dark:text-white">
          {{ titleLead }}<span class="text-blue-600 dark:text-cyan-300">{{ titleHighlight }}</span>{{ titleSuffix }}
        </h2>

        <p class="mx-auto mt-7 max-w-2xl text-base font-light leading-relaxed tracking-wide text-zinc-500 sm:text-lg lg:mx-0 lg:text-xl dark:text-slate-300">
          <span class="opacity-60">—</span> {{ subtitle }} <span class="opacity-60">—</span>
        </p>

        <div class="mt-10 grid max-w-xl grid-cols-2 gap-4 md:gap-6 lg:mx-0">
          <div
            v-for="item in stats"
            :key="item.label"
            class="group relative rounded-2xl border border-zinc-100/80 bg-white/60 px-5 py-5 transition-all duration-500 hover:-translate-y-1 hover:border-zinc-200 hover:shadow-[0_8px_30px_rgb(0,0,0,0.02)] md:px-6 md:py-6 dark:border-slate-800/80 dark:bg-slate-900/65"
          >
            <div class="origin-left text-2xl font-serif text-black transition-transform duration-500 group-hover:scale-105 md:text-3xl dark:text-white">
              {{ item.value }}
            </div>
            <div class="mt-1 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-400 dark:text-slate-400">
              {{ item.label }}
            </div>
          </div>
        </div>

        <div class="mt-10 flex flex-col items-center gap-4 sm:flex-row lg:items-start">
          <router-link
            :to="consolePath"
            class="group inline-flex items-center justify-center rounded-full bg-black px-8 py-4 text-base font-semibold text-white shadow-xl transition-all duration-300 hover:scale-105 hover:bg-zinc-800 dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950 dark:hover:bg-cyan-200"
          >
            {{ primaryLabel }}
          </router-link>
          <a
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center justify-center rounded-full border-2 border-zinc-200 bg-white px-8 py-4 text-base font-semibold text-black transition-all duration-300 hover:scale-105 hover:border-zinc-400 dark:border-slate-700 dark:bg-slate-950 dark:text-white dark:hover:border-cyan-400/40"
          >
            {{ secondaryLabel }}
          </a>
        </div>
      </div>

      <div
        data-test="hero-visual-stack"
        class="relative flex w-full flex-1 items-center justify-center lg:min-h-[clamp(30rem,68vh,46rem)]"
      >
        <div class="relative w-full max-w-[min(92vw,44rem)] lg:max-w-[min(48vw,44rem)]">
          <div class="relative aspect-[4/3] w-full lg:aspect-square">
            <div class="relative z-20 h-full w-full overflow-hidden rounded-[2rem] border border-zinc-200/60 bg-zinc-50 shadow-2xl dark:border-slate-800/80 dark:bg-slate-900">
              <picture class="block h-full w-full">
                <source srcset="/home-hero-main.avif" type="image/avif" />
                <source srcset="/home-hero-main.webp" type="image/webp" />
                <img
                  src="/home-hero-main.png"
                  alt="Homepage hero visual"
                  width="1275"
                  height="1234"
                  class="block h-full w-full object-cover"
                  fetchpriority="high"
                />
              </picture>
              <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.08),rgba(15,23,42,0.14))]"></div>
              <div class="pointer-events-none absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-slate-950/20 via-slate-900/10 to-transparent"></div>
            </div>

            <div class="hero-card-float absolute -right-[6%] -top-[8%] z-30 hidden w-[46%] overflow-hidden rounded-2xl border border-zinc-200/50 bg-white shadow-xl md:block dark:border-slate-800/80 dark:bg-slate-900">
              <div class="relative aspect-video">
                <picture class="block h-full w-full">
                  <source srcset="/home-hero-side-top.avif" type="image/avif" />
                  <source srcset="/home-hero-side-top.webp" type="image/webp" />
                  <img
                    src="/home-hero-side-top.png"
                    alt="Hero supporting visual top"
                    width="1254"
                    height="1254"
                    class="block h-full w-full object-cover"
                  />
                </picture>
                <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.02),rgba(15,23,42,0.12))]"></div>
              </div>
            </div>

            <div class="hero-card-float-alt absolute -bottom-[7%] -left-[5%] z-10 hidden w-[38%] overflow-hidden rounded-2xl border border-zinc-200/50 bg-white shadow-lg md:block dark:border-slate-800/80 dark:bg-slate-900">
              <div class="relative aspect-square">
                <picture class="block h-full w-full">
                  <source srcset="/home-hero-side-bottom.avif" type="image/avif" />
                  <source srcset="/home-hero-side-bottom.webp" type="image/webp" />
                  <img
                    src="/home-hero-side-bottom.png"
                    alt="Hero supporting visual bottom"
                    width="1254"
                    height="1254"
                    class="block h-full w-full object-cover"
                  />
                </picture>
                <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.04),rgba(15,23,42,0.12))]"></div>
              </div>
            </div>

            <div class="pointer-events-none absolute inset-0 rounded-full bg-blue-500/5 blur-[120px] dark:bg-cyan-400/10"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="relative z-10 mt-8 hidden w-full border-t border-zinc-100/80 pt-8 md:block dark:border-slate-800/80">
      <div class="mx-auto flex max-w-6xl items-center justify-center">
        <div class="hero-scroll-indicator flex flex-col items-center gap-3 text-zinc-300 dark:text-slate-600">
          <div class="h-12 w-px bg-gradient-to-b from-zinc-200 to-transparent dark:from-slate-700"></div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  badge: string
  titlePrefix: string
  titleLead: string
  titleHighlight: string
  titleSuffix: string
  subtitle: string
  checks: string[]
  consolePath: string
  primaryLabel: string
  secondaryLabel: string
  docUrl: string
  stats: Array<{ value: string; label: string }>
  headerOffset?: number
}>()

const heroStyle = computed(() => ({
  minHeight: `calc(100vh - ${props.headerOffset ?? 120}px)`
}))

const heroStreams = [
  {
    id: 'stream-1',
    style: {
      left: '8%',
      top: '-6%',
      animationDuration: '15s',
      animationDelay: '0s'
    }
  },
  {
    id: 'stream-2',
    style: {
      left: '34%',
      top: '-12%',
      animationDuration: '20s',
      animationDelay: '4s'
    }
  },
  {
    id: 'stream-3',
    style: {
      left: '62%',
      top: '-8%',
      animationDuration: '24s',
      animationDelay: '8s'
    }
  }
]

const heroParticles = [
  { id: 'p1', style: { left: '8%', top: '18%', animationDuration: '20s', animationDelay: '0s' } },
  { id: 'p2', style: { left: '16%', top: '62%', animationDuration: '22s', animationDelay: '2s' } },
  { id: 'p3', style: { left: '28%', top: '34%', animationDuration: '24s', animationDelay: '5s' } },
  { id: 'p4', style: { left: '44%', top: '70%', animationDuration: '26s', animationDelay: '3s' } },
  { id: 'p5', style: { left: '58%', top: '20%', animationDuration: '21s', animationDelay: '7s' } },
  { id: 'p6', style: { left: '72%', top: '52%', animationDuration: '25s', animationDelay: '1s' } },
  { id: 'p7', style: { left: '84%', top: '26%', animationDuration: '23s', animationDelay: '6s' } },
  { id: 'p8', style: { left: '90%', top: '76%', animationDuration: '27s', animationDelay: '4s' } }
]
</script>

<style scoped>
.hero-scan-light {
  animation: hero-scan 12s linear infinite;
}

.hero-stream {
  animation: hero-stream 18s linear infinite;
  opacity: 0;
}

.hero-glow {
  animation: hero-glow 16s ease-in-out infinite;
  filter: blur(160px);
}

.hero-glow-right {
  animation-duration: 18s;
}

.hero-particle {
  animation: hero-particle linear infinite;
  opacity: 0;
}

.hero-logo-float {
  animation: hero-logo-float 5s ease-in-out infinite;
}

.hero-card-float {
  animation: hero-card-float 7s ease-in-out infinite;
}

.hero-card-float-alt {
  animation: hero-card-float-alt 9s ease-in-out infinite;
}

.hero-scroll-indicator {
  animation: hero-scroll 2.4s ease-in-out infinite;
}

@keyframes hero-scan {
  0% {
    transform: translateY(-120%);
  }

  100% {
    transform: translateY(220%);
  }
}

@keyframes hero-stream {
  0% {
    transform: translate3d(-20vw, -12vh, 0) rotate(45deg);
    opacity: 0;
  }

  18% {
    opacity: 0.12;
  }

  82% {
    opacity: 0.12;
  }

  100% {
    transform: translate3d(120vw, 95vh, 0) rotate(45deg);
    opacity: 0;
  }
}

@keyframes hero-glow {
  0%, 100% {
    transform: translate3d(0, 0, 0);
    opacity: 0.3;
  }

  35% {
    transform: translate3d(30px, -30px, 0);
    opacity: 0.5;
  }

  70% {
    transform: translate3d(-10px, 20px, 0);
    opacity: 0.34;
  }
}

@keyframes hero-particle {
  0% {
    transform: translateY(-6%);
    opacity: 0;
  }

  15% {
    opacity: 0.2;
  }

  85% {
    opacity: 0.2;
  }

  100% {
    transform: translateY(105vh);
    opacity: 0;
  }
}

@keyframes hero-logo-float {
  0%, 100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(-8px);
  }
}

@keyframes hero-card-float {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }

  50% {
    transform: translate3d(10px, -20px, 0);
  }
}

@keyframes hero-card-float-alt {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }

  50% {
    transform: translate3d(-10px, 20px, 0);
  }
}

@keyframes hero-scroll {
  0%, 100% {
    transform: translateY(0);
    opacity: 0.75;
  }

  50% {
    transform: translateY(8px);
    opacity: 1;
  }
}

@media (max-width: 1023px) {
  [data-test='home-hero'] {
    min-height: auto !important;
  }
}
</style>
