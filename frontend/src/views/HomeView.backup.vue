<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    id="top"
    class="home-shell relative min-h-screen overflow-hidden bg-[#f4f8ff] text-slate-950 transition-colors duration-300 dark:bg-[#050816] dark:text-white"
    :class="{ 'home-dark': isDark }"
  >
    <div
      data-test="home-decor-overlay"
      class="pointer-events-none absolute inset-0 overflow-hidden"
      :class="{ 'home-decor-dark-tuned': isDark }"
    >
      <div class="hero-ambient absolute inset-x-0 top-0 h-[720px]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div data-test="hero-flow-left" class="hero-wave-left absolute -left-[18%] top-[140px] h-[520px] w-[70%]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-wave-left-secondary absolute -left-[6%] top-[248px] h-[240px] w-[54%]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div data-test="hero-flow-right" class="hero-wave-right absolute right-[-8%] top-[80px] h-[560px] w-[52%]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-wave-right-secondary absolute right-[6%] top-[164px] h-[360px] w-[34%]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-mist absolute left-[12%] top-[88px] h-32 w-32 rounded-full" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-mist-alt absolute right-[18%] top-[132px] h-24 w-24 rounded-full" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-sweep absolute inset-x-0 top-[258px] h-[220px]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
      <div class="hero-floor absolute inset-x-0 top-[340px] h-[280px]" :class="{ 'home-dark-hidden-layer': isDark }"></div>
    </div>

    <header ref="headerRef" class="fixed inset-x-0 top-0 z-50 px-4 py-4 sm:px-6">
      <nav
        class="mx-auto flex max-w-7xl items-center justify-between rounded-full border border-white/70 bg-white/78 px-4 py-3 shadow-[0_18px_60px_rgba(147,197,253,0.22)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/82 dark:shadow-[0_24px_80px_rgba(2,6,23,0.68)] sm:px-6"
      >
        <a href="#top" class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200/70 dark:bg-slate-900/90 dark:ring-slate-700/80">
            <img :src="siteLogo || '/favicon.ico'" alt="Logo" class="h-full w-full object-contain p-1.5" />
          </div>
          <div>
            <p class="text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-400 dark:text-slate-500">
              {{ t('home.landing.logoTagline') }}
            </p>
            <p class="text-lg font-black tracking-tight">{{ t('home.landing.brand') }}</p>
          </div>
        </a>

        <div class="hidden items-center gap-8 lg:flex">
          <template v-for="item in navItems" :key="item.label">
            <router-link
              v-if="item.to"
              :to="item.to"
              class="text-sm font-semibold text-slate-600 transition-colors hover:text-slate-950 dark:text-slate-300/95 dark:hover:text-cyan-200"
            >
              {{ item.label }}
            </router-link>
            <a
              v-else
              :href="item.href"
              :target="item.external ? '_blank' : undefined"
              :rel="item.external ? 'noopener noreferrer' : undefined"
              class="text-sm font-semibold text-slate-600 transition-colors hover:text-slate-950 dark:text-slate-300/95 dark:hover:text-cyan-200"
            >
              {{ item.label }}
            </a>
          </template>
        </div>

        <div class="flex items-center gap-2 sm:gap-3">
          <LocaleSwitcher />

          <button
            @click="toggleTheme"
            class="inline-flex h-10 w-10 items-center justify-center rounded-full border border-white/80 bg-white/80 text-slate-500 transition-colors hover:text-slate-950 dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:border-cyan-400/30 dark:hover:text-cyan-200"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            :to="consolePath"
            class="inline-flex items-center rounded-full bg-slate-950 px-5 py-2.5 text-sm font-bold text-white transition-transform hover:-translate-y-0.5 hover:bg-slate-800 dark:border dark:border-cyan-400/30 dark:bg-cyan-300/90 dark:text-slate-950 dark:shadow-[0_12px_30px_rgba(34,211,238,0.18)] dark:hover:bg-cyan-200"
          >
            {{ consoleLabel }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 px-4 pb-16 sm:px-6" :style="mainStyle">
      <section class="mx-auto max-w-7xl pt-4 lg:pt-8">
        <div class="grid gap-10 lg:grid-cols-[minmax(0,1.02fr)_minmax(420px,0.98fr)] lg:items-center">
          <div class="relative z-10 pt-6 lg:pt-10">
            <div class="inline-flex items-center gap-2 rounded-full border border-white/70 bg-white/76 px-4 py-2 text-sm font-semibold text-slate-700 shadow-[0_12px_40px_rgba(148,163,184,0.15)] backdrop-blur-md dark:border-cyan-400/20 dark:bg-slate-900/70 dark:text-cyan-100 dark:shadow-[0_16px_40px_rgba(8,47,73,0.32)]">
              <span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-blue-100 text-xs text-blue-600 dark:bg-cyan-400/15 dark:text-cyan-200">✦</span>
              <span>{{ t('home.landing.hero.badge') }}</span>
            </div>

            <h1 class="mt-7 max-w-3xl text-5xl font-black tracking-[-0.06em] text-slate-950 dark:text-white dark:[text-shadow:0_8px_36px_rgba(8,145,178,0.18)] sm:text-6xl xl:text-7xl">
              {{ t('home.landing.hero.titlePrefix') }}
              <span class="block mt-2">
                {{ t('home.landing.hero.titleLead') }}
                <span class="bg-gradient-to-r from-blue-500 via-sky-500 to-cyan-400 bg-clip-text text-transparent">{{ t('home.landing.hero.titleHighlight') }}</span>
                {{ t('home.landing.hero.titleSuffix') }}
              </span>
            </h1>

            <p class="mt-6 max-w-2xl text-lg font-medium leading-8 text-slate-600 dark:text-slate-200 sm:text-xl">
              {{ t('home.landing.hero.subtitle') }}
            </p>

            <div class="mt-7 flex flex-wrap gap-x-6 gap-y-3 text-sm font-semibold text-slate-700 dark:text-slate-100">
              <div
                v-for="item in heroChecks"
                :key="item.label"
                class="inline-flex items-center gap-2"
              >
                <span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-blue-100 text-[11px] text-blue-600 dark:bg-cyan-400/20 dark:text-cyan-200">●</span>
                <span>{{ item.label }}</span>
              </div>
            </div>

            <div class="mt-9 flex flex-col gap-4 sm:flex-row">
              <router-link
                :to="consolePath"
                class="inline-flex min-w-[148px] items-center justify-center rounded-2xl bg-slate-950 px-6 py-4 text-base font-black text-white transition-transform hover:-translate-y-0.5 hover:bg-slate-800 dark:border dark:border-cyan-400/30 dark:bg-cyan-300/90 dark:text-slate-950 dark:shadow-[0_18px_36px_rgba(34,211,238,0.2)] dark:hover:bg-cyan-200"
              >
                {{ heroPrimaryLabel }}
              </router-link>
              <a
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex min-w-[148px] items-center justify-center rounded-2xl border border-white/80 bg-white/78 px-6 py-4 text-base font-bold text-slate-700 shadow-[0_12px_40px_rgba(148,163,184,0.15)] transition-transform hover:-translate-y-0.5 hover:text-slate-950 dark:border-slate-700/80 dark:bg-slate-900/75 dark:text-slate-100 dark:hover:border-cyan-400/30 dark:hover:text-cyan-100"
              >
                {{ t('home.landing.hero.secondaryCta') }} →
              </a>
            </div>
          </div>

          <div class="relative mx-auto hidden w-full max-w-[620px] lg:block">
            <div class="hero-scene relative h-[560px]">
              <div data-test="hero-core-glow" class="hero-scene-glow absolute inset-x-8 bottom-8 top-14 rounded-[44px]" :class="{ 'home-dark-keep-glow': isDark }"></div>
              <div data-test="hero-beam-right" class="hero-light-column absolute right-10 top-0 h-[420px] w-px"></div>
              <div data-test="hero-beam-left" class="hero-light-column absolute left-[20%] top-14 h-[300px] w-px opacity-50"></div>
              <div class="hero-float-cube absolute left-2 top-14 h-10 w-10"></div>
              <div class="hero-float-cube absolute right-3 top-20 h-12 w-12"></div>
              <div class="hero-float-cube absolute left-[52%] bottom-[160px] h-12 w-12"></div>
              <div class="hero-aura absolute right-[18%] top-24 h-[324px] w-[308px] rounded-full" :class="{ 'home-dark-keep-glow': isDark }"></div>

              <div class="hero-ring absolute inset-x-[12%] bottom-4 h-24 rounded-full"></div>
              <div data-test="hero-plinth" class="hero-pedestal absolute bottom-10 left-1/2 w-[360px] -translate-x-1/2">
                <div class="hero-pedestal-base mx-auto h-8 w-[84%] rounded-[28px]"></div>
                <div class="hero-pedestal-mid mx-auto -mt-2 h-10 w-[72%] rounded-[26px]"></div>
                <div class="hero-pedestal-top mx-auto -mt-2 h-12 w-[60%] rounded-[24px]"></div>
              </div>

              <div data-test="hero-panel" class="hero-panel hero-panel-back absolute right-[12%] top-12 h-[292px] w-[232px] rounded-[34px]"></div>
              <div data-test="hero-panel" class="hero-panel hero-panel-mid absolute right-[20%] top-20 h-[308px] w-[244px] rounded-[34px]"></div>
              <div data-test="hero-panel" class="hero-panel hero-panel-front absolute right-[28%] top-28 flex h-[324px] w-[256px] items-center justify-center rounded-[38px]">
                <div class="hero-panel-edge absolute inset-3 rounded-[30px]"></div>
                <div data-test="hero-core" class="hero-panel-core absolute inset-[18px] rounded-[28px]"></div>
                <div class="hero-core-reflection absolute left-10 top-7 h-[184px] w-[64px] rounded-full"></div>
                <div class="hero-core-shadow absolute bottom-[34px] left-1/2 h-16 w-[148px] -translate-x-1/2 rounded-full"></div>
                <div class="hero-ai-text relative z-10 select-none bg-gradient-to-b from-[#93b5ff] via-[#4d82ff] to-[#2d5ef0] bg-clip-text text-[108px] font-black tracking-[-0.08em] text-transparent">
                  AI
                </div>
              </div>
              <div class="hero-reflection absolute inset-x-0 bottom-0 h-28" :class="{ 'home-dark-hidden-layer': isDark }"></div>
            </div>
          </div>
        </div>

        <section
          class="mx-auto mt-10 overflow-hidden rounded-[30px] border border-white/80 bg-white/84 shadow-[0_24px_80px_rgba(148,163,184,0.16)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/72 dark:shadow-[0_30px_80px_rgba(2,6,23,0.55)]"
          :style="sectionOffsetStyle"
        >
          <div class="grid gap-0 md:grid-cols-[1fr_1fr_1fr_1fr_1.2fr]">
            <div
              v-for="item in heroStats"
              :key="item.label"
              class="border-b border-slate-200/70 px-6 py-6 md:border-b-0 md:border-r md:border-slate-200/70 dark:border-slate-800/80"
            >
              <p class="text-2xl font-black tracking-tight text-slate-950 dark:text-white">{{ item.value }}</p>
              <p class="mt-1 text-sm font-medium text-slate-500 dark:text-slate-300">{{ item.label }}</p>
            </div>

            <a
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="group flex items-center justify-between gap-4 bg-gradient-to-r from-blue-50 via-white to-blue-100/70 px-6 py-6 transition-colors hover:from-blue-100 hover:to-blue-100 dark:from-slate-900/80 dark:via-slate-900/60 dark:to-cyan-500/12"
            >
              <div>
                <p class="text-sm font-black uppercase tracking-[0.2em] text-blue-600 dark:text-cyan-300">{{ t('home.landing.hero.promoTitle') }}</p>
                <p class="mt-2 text-sm text-slate-500 dark:text-slate-200">{{ t('home.landing.hero.promoSubtitle') }}</p>
              </div>
              <span class="inline-flex h-10 w-10 items-center justify-center rounded-full bg-blue-600 text-lg text-white transition-transform group-hover:translate-x-1 dark:bg-cyan-300 dark:text-slate-950">→</span>
            </a>
          </div>
        </section>
      </section>

      <section
        id="overview"
        class="mx-auto mt-10 max-w-7xl rounded-[32px] border border-white/80 bg-white/82 px-6 py-8 shadow-[0_24px_80px_rgba(148,163,184,0.14)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/72 sm:px-8"
        :style="sectionOffsetStyle"
      >
        <div>
          <div>
            <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-600 dark:text-cyan-300">{{ t('home.landing.overview.kicker') }}</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-[2.1rem]">
              {{ t('home.landing.overview.title') }}
            </h2>
            <p class="mt-3 max-w-3xl text-sm leading-7 text-slate-500 dark:text-slate-300">
              {{ t('home.landing.overview.subtitle') }}
            </p>
          </div>
        </div>

        <div class="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="row in benefitRows"
            :key="row.title"
            data-test="benefit-tag"
            class="rounded-[24px] border p-5 shadow-[0_10px_30px_rgba(148,163,184,0.12)] transition-colors duration-300"
            :class="isDark
              ? 'border-slate-800/90 bg-slate-950/92 shadow-[0_22px_48px_rgba(2,6,23,0.52)]'
              : 'border-slate-200/70 bg-white/90'"
          >
            <div
              data-test="benefit-icon-shell"
              class="inline-flex h-11 w-11 items-center justify-center rounded-2xl ring-1 transition-colors duration-300"
              :class="isDark
                ? 'bg-slate-900/90 text-cyan-100 ring-cyan-400/20'
                : 'bg-blue-50 text-blue-600 ring-blue-100'"
            >
              <Icon :name="row.icon" size="md" class="h-5 w-5" />
            </div>
            <h3 class="mt-4 text-lg font-black text-slate-950 dark:text-white">{{ row.title }}</h3>
            <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-100/90">{{ row.description }}</p>
          </article>
        </div>
      </section>

      <section
        id="platforms"
        class="mx-auto mt-10 max-w-7xl"
        :style="sectionOffsetStyle"
      >
        <div>
          <div>
            <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ t('home.landing.platforms.kicker') }}</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">{{ t('home.landing.platforms.title') }}</h2>
            <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300">{{ t('home.landing.platforms.subtitle') }}</p>
          </div>
        </div>

        <div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          <article
            v-for="row in platformRows"
            :key="row.label"
            class="rounded-[24px] border border-white/80 bg-white/82 px-5 py-6 text-center shadow-[0_16px_50px_rgba(148,163,184,0.12)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/78 dark:shadow-[0_18px_42px_rgba(2,6,23,0.42)]"
          >
            <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-100 text-xl text-slate-900 dark:bg-slate-800 dark:text-cyan-100">
              {{ row.icon }}
            </div>
            <p class="mt-4 text-sm font-semibold text-slate-700 dark:text-slate-200">{{ row.label }}</p>
          </article>
        </div>
      </section>

      <section class="mx-auto mt-10 max-w-7xl grid gap-4 lg:grid-cols-3">
        <article class="rounded-[30px] border border-white/80 bg-white/84 p-6 shadow-[0_24px_80px_rgba(148,163,184,0.14)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/72">
          <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ t('home.landing.models.kicker') }}</p>
          <h2 class="mt-3 text-2xl font-black tracking-tight text-slate-950 dark:text-white">{{ t('home.landing.models.cardTitle') }}</h2>
          <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300">{{ t('home.landing.models.cardSubtitle') }}</p>
          <p class="mt-3 rounded-2xl border border-blue-100 bg-blue-50/80 px-4 py-3 text-sm font-medium text-blue-700 dark:border-cyan-400/20 dark:bg-cyan-400/10 dark:text-cyan-100">
            {{ t('home.landing.models.apiNotice') }}
          </p>
          <div class="mt-6 space-y-4">
            <div
              v-for="model in modelChoices"
              :key="model.name"
              class="rounded-[20px] border border-slate-200/70 bg-white px-4 py-4 dark:border-slate-800/80 dark:bg-slate-950/80"
            >
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-lg font-black text-blue-600 dark:text-cyan-300">{{ model.name }}</p>
                  <p class="mt-1 text-sm text-slate-500 dark:text-slate-200">{{ model.description }}</p>
                </div>
                <span
                  v-if="model.badge"
                  class="inline-flex rounded-full bg-emerald-100 px-3 py-1 text-xs font-black text-emerald-700 dark:bg-cyan-400/20 dark:text-cyan-100"
                >
                  {{ model.badge }}
                </span>
              </div>
            </div>
          </div>
        </article>

        <section
          id="pricing"
          class="rounded-[30px] border border-white/80 bg-white/84 p-6 shadow-[0_24px_80px_rgba(148,163,184,0.14)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/72"
          :style="sectionOffsetStyle"
        >
          <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ t('home.landing.pricing.kicker') }}</p>
          <h2 class="mt-3 text-2xl font-black tracking-tight text-slate-950 dark:text-white">{{ t('home.landing.pricing.title') }}</h2>
          <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300">{{ t('home.landing.pricing.subtitle') }}</p>

          <div class="mt-6 grid gap-4">
            <router-link
              v-for="row in pricingRows"
              :key="row.plan"
              to="/purchase"
              data-test="pricing-card"
              :data-plan-highlight="row.featured ? 'true' : 'false'"
              class="rounded-[22px] border px-4 py-4 transition-transform hover:-translate-y-0.5"
              :class="row.featured
                ? 'border-blue-200 bg-gradient-to-r from-blue-50 via-white to-cyan-50 dark:border-cyan-400/20 dark:from-cyan-500/10 dark:via-white/5 dark:to-blue-500/10'
                : 'border-slate-200/70 bg-white dark:border-slate-800/80 dark:bg-slate-950/80'"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <div class="flex items-center gap-2">
                    <p class="text-base font-black text-slate-950 dark:text-white">{{ row.plan }}</p>
                    <span
                      class="inline-flex rounded-full px-2.5 py-1 text-[11px] font-black"
                      :class="row.featured
                        ? 'bg-blue-600 text-white dark:bg-cyan-300 dark:text-slate-950'
                        : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-100'"
                    >
                      {{ row.badge }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-slate-500 dark:text-slate-200">{{ row.description }}</p>
                </div>
                <p class="whitespace-nowrap text-xl font-black text-slate-950 dark:text-white">{{ row.price }}</p>
              </div>
            </router-link>
          </div>

          <router-link
            to="/purchase"
            class="mt-6 inline-flex text-sm font-semibold text-blue-600 transition-colors hover:text-blue-700 dark:text-cyan-300 dark:hover:text-cyan-200"
          >
            {{ t('home.landing.pricing.viewDetails') }} →
          </router-link>
        </section>

        <article class="rounded-[30px] border border-white/80 bg-white/84 p-6 shadow-[0_24px_80px_rgba(148,163,184,0.14)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-900/72">
          <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ t('home.landing.setup.kicker') }}</p>
          <h2 class="mt-3 text-2xl font-black tracking-tight text-slate-950 dark:text-white">{{ t('home.landing.setup.title') }}</h2>
          <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300">{{ t('home.landing.setup.subtitle') }}</p>

          <div class="mt-6 space-y-5">
            <div
              v-for="step in setupSteps"
              :key="step.step"
              class="flex gap-4"
            >
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-50 text-base font-black text-blue-600 dark:bg-cyan-400/15 dark:text-cyan-100">
                {{ step.step }}
              </div>
              <div>
                <p class="text-base font-black text-blue-600 dark:text-cyan-300">{{ step.title }}</p>
                <p class="mt-1 text-sm leading-7 text-slate-500 dark:text-slate-200">{{ step.description }}</p>
              </div>
            </div>
          </div>

          <router-link
            :to="consolePath"
            class="mt-8 inline-flex w-full items-center justify-center rounded-2xl bg-slate-950 px-5 py-4 text-base font-black text-white transition-transform hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
          >
            {{ heroPrimaryLabel }}
          </router-link>
        </article>
      </section>

      <section
        id="community"
        class="relative mx-auto mt-10 max-w-7xl overflow-hidden rounded-[32px] border border-slate-900/10 bg-[#0c1428] px-6 py-8 text-white shadow-[0_30px_80px_rgba(15,23,42,0.25)] sm:px-8"
        :style="sectionOffsetStyle"
      >
        <div class="community-bg pointer-events-none absolute inset-0 rounded-[32px]"></div>
        <div class="relative flex flex-col items-center gap-6 py-2">
          <div class="text-center">
            <p class="text-sm font-black uppercase tracking-[0.22em] text-cyan-300">{{ t('home.landing.community.kicker') }}</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight">{{ t('home.landing.community.title') }}</h2>
            <p class="mt-3 text-base font-medium text-slate-100">{{ t('home.landing.community.lead') }}</p>
            <p class="mt-2 text-sm leading-7 text-slate-200">{{ t('home.landing.community.body') }}</p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div
              v-for="item in communityQrItems"
              :key="item.alt"
              class="overflow-hidden rounded-[28px] border border-white/10 bg-white/5 p-3 text-center shadow-[0_24px_80px_rgba(15,23,42,0.18)] backdrop-blur-xl dark:border-cyan-400/15 dark:bg-slate-900/72"
            >
              <img
                :src="item.src"
                :alt="item.alt"
                class="h-auto w-full max-w-[300px] rounded-[22px] bg-white object-cover"
              />
              <p class="mt-3 text-sm font-semibold text-slate-100">{{ item.label }}</p>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 mt-10 bg-[#0a1020] px-4 pb-8 pt-8 text-white sm:px-6">
      <div class="mx-auto max-w-7xl rounded-[28px] border border-white/8 bg-slate-950/45 px-6 py-8 backdrop-blur-xl sm:px-8">
        <div class="grid gap-8 lg:grid-cols-[1.1fr_1.4fr_0.55fr]">
          <div>
            <div class="flex items-center gap-3">
              <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-white/10">
                <img :src="siteLogo || '/favicon.ico'" alt="Logo" class="h-full w-full object-contain p-1.5" />
              </div>
              <div>
                <p class="text-lg font-black">{{ t('home.landing.brand') }}</p>
                <p class="text-sm text-slate-300">{{ t('home.landing.footer.about') }}</p>
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
            <div v-for="column in footerColumns" :key="column.title">
              <p class="text-sm font-black text-white">{{ column.title }}</p>
              <div class="mt-4 space-y-3">
                <template v-for="item in column.items" :key="item.label">
                  <a
                    v-if="item.href"
                    :href="item.href"
                    :target="item.external ? '_blank' : undefined"
                    :rel="item.external ? 'noopener noreferrer' : undefined"
                    class="block text-sm text-slate-300 transition-colors hover:text-white dark:hover:text-cyan-200"
                  >
                    {{ item.label }}
                  </a>
                  <router-link
                    v-else-if="item.to"
                    :to="item.to"
                    class="block text-sm text-slate-300 transition-colors hover:text-white dark:hover:text-cyan-200"
                  >
                    {{ item.label }}
                  </router-link>
                </template>
              </div>
            </div>
          </div>

          <div class="flex flex-col items-start lg:items-end">
            <p class="text-sm font-black text-white">{{ t('home.landing.footer.follow') }}</p>
            <div class="mt-4 grid grid-cols-2 gap-3">
              <div
                v-for="item in communityQrItems"
                :key="`footer-${item.alt}`"
                class="flex flex-col items-center"
              >
                <img
                  :src="item.src"
                  :alt="item.alt"
                  class="h-28 w-28 rounded-[20px] border border-white/10 bg-white object-cover p-1.5 dark:border-cyan-400/10"
                />
                <p class="mt-2 text-xs text-slate-300">{{ item.label }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-8 flex flex-col items-center justify-center gap-3 border-t border-white/10 pt-6 text-center text-sm text-slate-300">
          <p>
            © 2018 {{ t('home.landing.footer.copyrightOwner') }}
            <a
              href="https://beian.miit.gov.cn/"
              target="_blank"
              rel="noopener noreferrer"
              class="transition-colors hover:text-white hover:underline"
            >
              蜀ICP备17044249号-1
            </a>
          </p>
          <a
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="font-semibold transition-colors hover:text-white"
          >
            {{ t('home.landing.nav.docs') }}
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { TOKEN_DOC_URL } from '@/constants/externalLinks'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const docUrl = TOKEN_DOC_URL
const communityQrItems = computed(() => [
  { src: '/qq.jpg', alt: t('home.landing.community.qrAlt.qq'), label: t('home.landing.community.qrLabels.qq') },
  { src: '/wechat.jpg', alt: t('home.landing.community.qrAlt.wechat'), label: t('home.landing.community.qrLabels.wechat') }
])

const headerRef = ref<HTMLElement | null>(null)
const headerOffset = ref(120)

type NavItem = {
  label: string
  href?: string
  to?: string
  external?: boolean
}

type FooterColumnItem = {
  label: string
  href?: string
  to?: string
  external?: boolean
}

const navItems = computed<NavItem[]>(() => [
  { href: '#top', label: t('home.landing.nav.home') },
  { href: '#pricing', label: t('home.landing.nav.pricing') },
  { href: docUrl, label: t('home.landing.nav.docs'), external: true },
  { href: '#community', label: t('home.landing.nav.community') }
])

const heroChecks = computed(() => [
  { label: t('home.landing.hero.checks.availability') },
  { label: t('home.landing.hero.checks.latency') },
  { label: t('home.landing.hero.checks.security') }
])

const heroStats = computed(() => [
  { value: '1.2M+', label: t('home.landing.hero.stats.calls') },
  { value: '50,000+', label: t('home.landing.hero.stats.developers') },
  { value: '99.9%', label: t('home.landing.hero.stats.sla') },
  { value: '24/7', label: t('home.landing.hero.stats.support') }
])

type BenefitIconName = 'chartBar' | 'shield' | 'dollar' | 'swap' | 'chatBubble' | 'refresh'

const benefitRows = computed<Array<{ icon: BenefitIconName; title: string; description: string }>>(() => [
  { icon: 'chartBar', title: t('home.landing.overview.benefits.stability.title'), description: t('home.landing.overview.benefits.stability.description') },
  { icon: 'shield', title: t('home.landing.overview.benefits.security.title'), description: t('home.landing.overview.benefits.security.description') },
  { icon: 'dollar', title: t('home.landing.overview.benefits.value.title'), description: t('home.landing.overview.benefits.value.description') },
  { icon: 'swap', title: t('home.landing.overview.benefits.easy.title'), description: t('home.landing.overview.benefits.easy.description') },
  { icon: 'chatBubble', title: t('home.landing.overview.benefits.support.title'), description: t('home.landing.overview.benefits.support.description') },
  { icon: 'refresh', title: t('home.landing.overview.benefits.updates.title'), description: t('home.landing.overview.benefits.updates.description') }
])

const platformRows = computed(() => [
  { icon: '⊞', label: 'Windows' },
  { icon: '', label: 'macOS' },
  { icon: '◭', label: 'Linux' },
  { icon: '⬡', label: 'Docker' },
  { icon: '▣', label: t('home.landing.platforms.mobile') },
  { icon: '⌘', label: t('home.landing.platforms.devices') }
])

const modelChoices = computed(() => [
  { name: 'GPT-5.5', description: t('home.landing.models.items.gpt55'), badge: t('home.landing.models.badges.latest') },
  { name: 'GPT-5.4 mini', description: t('home.landing.models.items.gpt54Mini'), badge: '' },
  { name: 'GPT-5.4 nano', description: t('home.landing.models.items.gpt54Nano'), badge: '' },
  { name: 'GPT Image 2', description: t('home.landing.models.items.gptImage2'), badge: '' }
])

const pricingRows = computed(() => [
  { plan: t('home.landing.pricing.plans.standard.plan'), price: '¥19.9/月起', description: t('home.landing.pricing.plans.standard.description'), badge: t('home.landing.pricing.plans.standard.badge'), featured: false },
  { plan: t('home.landing.pricing.plans.pro.plan'), price: '¥135/月', description: t('home.landing.pricing.plans.pro.description'), badge: t('home.landing.pricing.plans.pro.badge'), featured: false },
  { plan: t('home.landing.pricing.plans.enterprise.plan'), price: t('home.landing.pricing.plans.enterprise.price'), description: t('home.landing.pricing.plans.enterprise.description'), badge: t('home.landing.pricing.plans.enterprise.badge'), featured: true }
])

const setupSteps = computed(() => [
  { step: '1', title: t('home.landing.setup.steps.register.title'), description: t('home.landing.setup.steps.register.description') },
  { step: '2', title: t('home.landing.setup.steps.integrate.title'), description: t('home.landing.setup.steps.integrate.description') },
  { step: '3', title: t('home.landing.setup.steps.start.title'), description: t('home.landing.setup.steps.start.description') }
])

const socialDots = ['微', '博', '视', '知']

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const consolePath = computed(() => (
  isAuthenticated.value
    ? dashboardPath.value
    : `/login?redirect=${encodeURIComponent(dashboardPath.value)}`
))
const consoleLabel = computed(() => isAuthenticated.value ? t('home.landing.console.enter') : t('home.landing.console.login'))
const heroPrimaryLabel = computed(() => isAuthenticated.value ? t('home.landing.hero.primaryCtaAuthed') : t('home.landing.hero.primaryCtaGuest'))
const footerColumns = computed<Array<{ title: string; items: FooterColumnItem[] }>>(() => [
  {
    title: t('home.landing.footer.columns.product.title'),
    items: [
      { label: t('home.landing.footer.columns.product.pricing'), href: '#pricing' },
      { label: t('home.landing.footer.columns.product.docs'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.product.platforms'), href: '#platforms' },
      { label: t('home.landing.footer.columns.product.changelog'), href: docUrl, external: true }
    ]
  },
  {
    title: t('home.landing.footer.columns.developer.title'),
    items: [
      { label: t('home.landing.footer.columns.developer.quickstart'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.developer.sdk'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.developer.bestPractices'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.developer.status'), href: docUrl, external: true }
    ]
  },
  {
    title: t('home.landing.footer.columns.company.title'),
    items: [
      { label: t('home.landing.footer.columns.company.about'), href: '#overview' },
      { label: t('home.landing.footer.columns.company.contact'), href: '#community' },
      { label: t('home.landing.footer.columns.company.terms'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.company.privacy'), href: docUrl, external: true }
    ]
  },
  {
    title: t('home.landing.footer.columns.support.title'),
    items: [
      { label: t('home.landing.footer.columns.support.help'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.support.community'), href: '#community' }
    ]
  }
])

const mainStyle = computed(() => ({
  paddingTop: `${headerOffset.value}px`
}))

const sectionOffsetStyle = computed(() => ({
  scrollMarginTop: `${headerOffset.value}px`
}))

function updateHeaderOffset() {
  if (!headerRef.value) {
    return
  }

  const { height } = headerRef.value.getBoundingClientRect()
  headerOffset.value = Math.ceil(height) + 24
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  nextTick(() => {
    updateHeaderOffset()
  })

  window.addEventListener('resize', updateHeaderOffset)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateHeaderOffset)
})
</script>

<style scoped>
:global(html) {
  scroll-behavior: smooth;
}

.home-shell {
  background-image:
    radial-gradient(circle at 12% 22%, rgba(255, 255, 255, 0.95) 0, rgba(255, 255, 255, 0) 26%),
    linear-gradient(180deg, rgba(228, 239, 255, 0.9) 0%, rgba(244, 248, 255, 0.9) 26%, rgba(246, 249, 255, 1) 100%);
}

.hero-ambient {
  background:
    radial-gradient(circle at 68% 22%, rgba(117, 165, 255, 0.34), rgba(117, 165, 255, 0) 34%),
    radial-gradient(circle at 28% 18%, rgba(255, 255, 255, 0.85), rgba(255, 255, 255, 0) 28%),
    linear-gradient(180deg, rgba(227, 238, 255, 0.88) 0%, rgba(244, 248, 255, 0.1) 100%);
}

.hero-wave-left {
  border-radius: 9999px;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.98), rgba(208, 226, 255, 0.88) 38%, rgba(133, 181, 255, 0.22) 70%, rgba(107, 162, 255, 0.04)),
    radial-gradient(circle at 18% 38%, rgba(255, 255, 255, 0.98), rgba(255, 255, 255, 0) 58%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 40px 90px rgba(153, 193, 255, 0.24);
  filter: blur(6px);
  opacity: 0.92;
  transform: perspective(1200px) rotateX(72deg) rotate(-8deg);
}

.hero-wave-left-secondary {
  border-radius: 9999px;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.86), rgba(178, 209, 255, 0.48), rgba(98, 150, 255, 0.02)),
    radial-gradient(circle at 24% 44%, rgba(255, 255, 255, 0.94), rgba(255, 255, 255, 0) 62%);
  filter: blur(14px);
  opacity: 0.72;
  transform: perspective(1200px) rotateX(76deg) rotate(-7deg);
}

.hero-wave-right {
  border-radius: 48px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(151, 188, 255, 0.36) 42%, rgba(112, 160, 255, 0.12)),
    radial-gradient(circle at 60% 38%, rgba(255, 255, 255, 0.86), rgba(255, 255, 255, 0) 58%);
  filter: blur(18px);
  opacity: 0.84;
}

.hero-wave-right-secondary {
  border-radius: 42px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.56), rgba(138, 181, 255, 0.22)),
    radial-gradient(circle at 50% 30%, rgba(255, 255, 255, 0.62), rgba(255, 255, 255, 0) 56%);
  filter: blur(18px);
  opacity: 0.72;
}

.hero-mist,
.hero-mist-alt {
  background: radial-gradient(circle, rgba(255, 255, 255, 0.9), rgba(255, 255, 255, 0));
  filter: blur(14px);
}

.hero-floor {
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0.98), rgba(212, 227, 255, 0.7) 28%, rgba(244, 248, 255, 0) 72%);
  filter: blur(8px);
}

.hero-sweep {
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0) 4%, rgba(255, 255, 255, 0.92) 18%, rgba(180, 210, 255, 0.56) 44%, rgba(120, 170, 255, 0.18) 70%, rgba(255, 255, 255, 0) 100%);
  filter: blur(16px);
  opacity: 0.6;
  transform: perspective(1600px) rotateX(75deg);
}

.hero-scene-glow {
  background:
    radial-gradient(circle at 55% 42%, rgba(116, 168, 255, 0.5), rgba(116, 168, 255, 0) 46%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.62), rgba(255, 255, 255, 0));
  filter: blur(14px);
}

.hero-light-column {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0), rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0));
  box-shadow: 0 0 22px rgba(255, 255, 255, 0.4);
}

.hero-float-cube {
  border: 1px solid rgba(255, 255, 255, 0.72);
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(184, 212, 255, 0.28));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 18px 40px rgba(140, 176, 235, 0.2);
  border-radius: 14px;
  backdrop-filter: blur(10px);
}

.hero-ring {
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0), rgba(123, 169, 255, 0.34) 48%, rgba(255, 255, 255, 0) 72%);
  filter: blur(4px);
}

.hero-aura {
  background:
    radial-gradient(circle at center, rgba(112, 164, 255, 0.34), rgba(112, 164, 255, 0.1) 38%, rgba(112, 164, 255, 0) 72%);
  filter: blur(20px);
  opacity: 0.9;
}

.hero-pedestal-base,
.hero-pedestal-mid,
.hero-pedestal-top,
.hero-panel {
  border: 1px solid rgba(255, 255, 255, 0.72);
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.74), rgba(205, 223, 255, 0.22));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.82),
    inset 0 -16px 32px rgba(139, 177, 255, 0.12),
    0 24px 60px rgba(103, 153, 255, 0.18);
  backdrop-filter: blur(16px);
}

.hero-panel-back {
  opacity: 0.58;
  transform: rotate(16deg) translateZ(0);
}

.hero-panel-mid {
  opacity: 0.76;
  transform: rotate(8deg) translateZ(0);
}

.hero-panel-front {
  position: relative;
  overflow: hidden;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    inset 0 -24px 46px rgba(113, 160, 255, 0.16),
    0 28px 80px rgba(96, 135, 225, 0.25);
}

.hero-panel-front::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.5), rgba(255, 255, 255, 0.06) 35%, rgba(255, 255, 255, 0.24)),
    linear-gradient(90deg, rgba(255, 255, 255, 0.58), rgba(255, 255, 255, 0.04) 22%, rgba(255, 255, 255, 0.28));
}

.hero-panel-edge {
  border: 1px solid rgba(255, 255, 255, 0.56);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.04));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.52),
    inset 0 -10px 18px rgba(115, 158, 255, 0.08);
}

.hero-panel-core {
  border: 1px solid rgba(255, 255, 255, 0.42);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.22), rgba(144, 183, 255, 0.08) 42%, rgba(255, 255, 255, 0.06) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.46),
    inset 0 -24px 36px rgba(102, 150, 255, 0.12);
  backdrop-filter: blur(18px);
}

.hero-core-reflection {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.64), rgba(255, 255, 255, 0.08));
  filter: blur(4px);
  opacity: 0.72;
}

.hero-core-shadow {
  background: radial-gradient(circle at center, rgba(95, 141, 255, 0.26), rgba(95, 141, 255, 0));
  filter: blur(6px);
}

.hero-ai-text {
  text-shadow:
    0 16px 30px rgba(59, 130, 246, 0.18),
    0 0 24px rgba(97, 154, 255, 0.16);
}

.hero-reflection {
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0.84), rgba(198, 221, 255, 0.22) 34%, rgba(255, 255, 255, 0));
  filter: blur(12px);
}

.community-bg {
  background:
    radial-gradient(circle at 28% 0%, rgba(99, 145, 255, 0.22), rgba(99, 145, 255, 0) 28%),
    radial-gradient(circle at 72% 26%, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0) 24%);
}

.home-decor-dark-tuned {
  isolation: isolate;
}

.home-dark {
  background-image:
    radial-gradient(circle at 74% 14%, rgba(34, 211, 238, 0.08) 0, rgba(34, 211, 238, 0) 26%),
    radial-gradient(circle at 22% 18%, rgba(59, 130, 246, 0.05) 0, rgba(59, 130, 246, 0) 24%),
    linear-gradient(180deg, rgba(5, 8, 22, 1) 0%, rgba(8, 13, 30, 1) 46%, rgba(5, 9, 20, 1) 100%);
}

.home-dark-hidden-layer {
  opacity: 0 !important;
}

.home-dark-keep-glow {
  opacity: 1 !important;
}

.home-dark .hero-scene-glow {
  background:
    radial-gradient(circle at 55% 42%, rgba(34, 211, 238, 0.2), rgba(34, 211, 238, 0) 48%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0));
  filter: blur(18px);
}

.home-dark .hero-aura {
  background:
    radial-gradient(circle at center, rgba(59, 130, 246, 0.22), rgba(59, 130, 246, 0.08) 42%, rgba(59, 130, 246, 0) 76%);
  filter: blur(28px);
}

.home-dark .hero-light-column {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0), rgba(125, 211, 252, 0.42), rgba(255, 255, 255, 0));
  box-shadow: 0 0 24px rgba(34, 211, 238, 0.18);
}

.home-dark .hero-float-cube {
  border-color: rgba(71, 85, 105, 0.7);
  background:
    linear-gradient(145deg, rgba(15, 23, 42, 0.86), rgba(30, 41, 59, 0.48));
  box-shadow:
    inset 0 1px 0 rgba(148, 163, 184, 0.14),
    0 16px 36px rgba(2, 6, 23, 0.34);
}

.home-dark .hero-ring {
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0), rgba(34, 211, 238, 0.16) 48%, rgba(255, 255, 255, 0) 72%);
}

.home-dark .hero-pedestal-base,
.home-dark .hero-pedestal-mid,
.home-dark .hero-pedestal-top,
.home-dark .hero-panel {
  border-color: rgba(71, 85, 105, 0.82);
  background:
    linear-gradient(145deg, rgba(15, 23, 42, 0.92), rgba(30, 41, 59, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(148, 163, 184, 0.14),
    inset 0 -16px 28px rgba(14, 165, 233, 0.05),
    0 24px 60px rgba(2, 6, 23, 0.4);
}

.home-dark .hero-panel-front {
  box-shadow:
    inset 0 1px 0 rgba(148, 163, 184, 0.16),
    inset 0 -24px 42px rgba(14, 165, 233, 0.06),
    0 28px 72px rgba(2, 6, 23, 0.46);
}

.home-dark .hero-panel-front::before {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.03) 35%, rgba(255, 255, 255, 0.1)),
    linear-gradient(90deg, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.02) 22%, rgba(255, 255, 255, 0.08));
}

.home-dark .hero-panel-edge {
  border-color: rgba(56, 189, 248, 0.14);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.02));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    inset 0 -10px 18px rgba(14, 165, 233, 0.04);
}

.home-dark .hero-panel-core {
  border-color: rgba(56, 189, 248, 0.16);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.09), rgba(14, 165, 233, 0.06) 42%, rgba(255, 255, 255, 0.03) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    inset 0 -24px 36px rgba(14, 165, 233, 0.08);
}

.home-dark .hero-core-reflection {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0.04));
}

.home-dark .hero-core-shadow {
  background: radial-gradient(circle at center, rgba(14, 165, 233, 0.16), rgba(14, 165, 233, 0));
}

@media (max-width: 1023px) {
  .hero-ambient {
    height: 560px;
  }

  .hero-wave-left {
    left: -28%;
    top: 160px;
    width: 92%;
  }

  .hero-wave-right {
    width: 70%;
  }
}
</style>
