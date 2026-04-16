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
    class="relative min-h-screen overflow-hidden bg-[#f6f7fb] text-slate-950 dark:bg-slate-950 dark:text-white"
  >
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 grid-pattern opacity-70 dark:opacity-15"></div>
      <div class="absolute inset-x-0 top-0 h-80 bg-gradient-to-b from-blue-100/70 to-transparent dark:from-cyan-500/10"></div>
      <div class="absolute -left-20 top-32 h-72 w-72 rounded-full bg-cyan-200/40 blur-3xl dark:bg-cyan-400/10"></div>
      <div class="absolute -right-24 top-96 h-72 w-72 rounded-full bg-amber-200/40 blur-3xl dark:bg-amber-300/10"></div>
    </div>

    <header ref="headerRef" class="fixed inset-x-0 top-0 z-50 px-4 py-4 sm:px-6">
      <nav
        class="mx-auto flex max-w-7xl items-center justify-between rounded-2xl border border-slate-200/80 bg-white/92 px-4 py-3 shadow-lg shadow-slate-200/30 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/82 dark:shadow-black/30"
      >
        <a href="#top" class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200/70 dark:bg-slate-900 dark:ring-white/10">
            <img :src="siteLogo || '/favicon.ico'" alt="Logo" class="h-full w-full object-contain p-1.5" />
          </div>
          <div>
            <p class="text-[11px] font-semibold uppercase tracking-[0.35em] text-slate-400 dark:text-slate-500">
              企业级 OpenAI 中转
            </p>
            <p class="text-lg font-black tracking-tight">GPAPI</p>
          </div>
        </a>

        <div class="hidden items-center gap-6 lg:flex">
          <a
            v-for="item in navItems"
            :key="item.href"
            :href="item.href"
            class="text-sm font-semibold text-slate-600 transition-colors hover:text-slate-950 dark:text-slate-300 dark:hover:text-white"
          >
            {{ item.label }}
          </a>
        </div>

        <div class="flex items-center gap-2 sm:gap-3">
          <LocaleSwitcher />

          <button
            @click="toggleTheme"
            class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-500 transition-colors hover:border-slate-300 hover:text-slate-900 dark:border-white/10 dark:bg-slate-900 dark:text-slate-300 dark:hover:border-white/20 dark:hover:text-white"
            :title="isDark ? '切换浅色模式' : '切换深色模式'"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            :to="consolePath"
            class="inline-flex items-center rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
          >
            {{ consoleLabel }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 px-4 pb-12 sm:px-6" :style="mainStyle">
      <section class="mx-auto max-w-7xl py-8 lg:py-10">
        <div class="grid gap-6 xl:grid-cols-[1.08fr,0.92fr]">
          <div class="rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8">
            <div class="inline-flex items-center gap-2 rounded-full bg-amber-50 px-4 py-2 text-sm font-bold text-amber-700 dark:bg-amber-400/10 dark:text-amber-200">
              <span>🎉</span>
              <span>注册即送 ￥5.00 体验额度 | 无需绑卡 | 即刻上手</span>
            </div>

            <h1 class="mt-6 max-w-4xl text-4xl font-black tracking-[-0.05em] text-slate-950 dark:text-white sm:text-5xl lg:text-6xl">
              GPAPI - 让 AI 开发告别网络烦恼
            </h1>
            <p class="mt-5 max-w-3xl text-lg font-semibold leading-8 text-slate-700 dark:text-slate-200 sm:text-xl">
              企业级 OpenAI 私有代理中转 | 国内极速响应 | 一行代码接入
            </p>

            <ul class="mt-8 space-y-3">
              <li
                v-for="advantage in heroAdvantages"
                :key="advantage.title"
                class="rounded-2xl border border-slate-200/80 bg-slate-50 px-4 py-4 text-sm leading-7 text-slate-700 dark:border-white/10 dark:bg-slate-950/60 dark:text-slate-200"
              >
                <span class="font-black text-slate-950 dark:text-white">{{ advantage.icon }} {{ advantage.title }}</span>
                <span> —— {{ advantage.description }}</span>
              </li>
            </ul>

            <div class="mt-8 flex flex-col gap-4 sm:flex-row">
              <router-link
                :to="consolePath"
                class="inline-flex items-center justify-center rounded-xl bg-blue-600 px-6 py-3.5 text-base font-black text-white transition-colors hover:bg-blue-700"
              >
                {{ heroPrimaryLabel }}
              </router-link>
              <router-link
                to="/guide"
                class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-slate-50 px-6 py-3.5 text-base font-semibold text-slate-700 transition-colors hover:border-slate-300 hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-slate-950/60 dark:text-slate-200 dark:hover:border-white/20 dark:hover:text-white"
              >
                📖 查看接入文档
              </router-link>
            </div>

            <div class="mt-8 flex flex-wrap gap-3 text-sm text-slate-500 dark:text-slate-300">
              <span
                v-for="tag in heroTags"
                :key="tag"
                class="rounded-full border border-slate-200 bg-white px-3 py-1.5 dark:border-white/10 dark:bg-slate-950/60"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <div class="space-y-5">
            <article class="rounded-[1.75rem] border border-slate-200/80 bg-slate-950 text-white shadow-sm dark:border-white/10">
              <div class="border-b border-white/10 px-6 py-4">
                <p class="text-sm font-black uppercase tracking-[0.2em] text-cyan-300">
                  接入示例
                </p>
              </div>
              <div class="space-y-3 px-6 py-5 font-mono text-sm leading-7 text-slate-200">
                <div><span class="text-slate-500">from</span> <span class="text-cyan-300">openai</span> <span class="text-slate-500">import</span> <span class="text-emerald-300">OpenAI</span></div>
                <div class="pt-2"><span class="text-slate-500">client</span> = <span class="text-violet-300">OpenAI</span>(</div>
                <div class="pl-4"><span class="text-slate-400">api_key=</span><span class="text-amber-300">"sk-gpapi-你的密钥"</span>,</div>
                <div class="pl-4"><span class="text-slate-400">base_url=</span><span class="text-emerald-300">"https://api.gpapi.com/v1"</span></div>
                <div>)</div>
              </div>
            </article>

          </div>
        </div>
      </section>

      <section
        id="overview"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <div class="max-w-3xl">
          <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
            为什么选择 GPAPI？
          </p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
            选 GPAPI 中转站的六大好处
          </h2>
        </div>

        <div class="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="row in benefitRows"
            :key="row.title"
            data-test="benefit-tag"
            class="rounded-[1.35rem] border border-slate-200/80 bg-slate-50 p-5 shadow-sm dark:border-white/10 dark:bg-slate-950/60"
          >
            <span class="inline-flex rounded-full bg-blue-100 px-3 py-1 text-xs font-black uppercase tracking-[0.18em] text-blue-700 dark:bg-cyan-400/15 dark:text-cyan-200">
              {{ row.title }}
            </span>
            <p class="mt-4 text-sm leading-7 text-slate-600 dark:text-slate-300">
              {{ row.description }}
            </p>
          </article>
        </div>
      </section>

      <section class="mx-auto mt-8 max-w-7xl">
        <article
          class="rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
          :style="sectionOffsetStyle"
        >
          <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
            全平台支持
          </p>
          <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
            支持的操作系统与设备
          </h2>

          <div class="mt-8 overflow-hidden rounded-[1.5rem] border border-slate-200/80 dark:border-white/10">
            <div class="overflow-x-auto">
              <table class="min-w-full border-collapse">
                <thead>
                  <tr class="bg-slate-50 dark:bg-slate-950/60">
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">平台</th>
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">支持方式</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="row in platformRows"
                    :key="row.platform"
                    class="border-t border-slate-200/80 bg-white dark:border-white/10 dark:bg-slate-900/70"
                  >
                    <th class="whitespace-nowrap px-5 py-4 text-left text-sm font-bold text-slate-900 dark:text-white">
                      {{ row.platform }}
                    </th>
                    <td class="px-5 py-4 text-sm leading-7 text-slate-600 dark:text-slate-300">
                      {{ row.support }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="mt-6 rounded-[1.25rem] border border-sky-200 bg-sky-50 p-5 dark:border-cyan-400/20 dark:bg-cyan-500/10">
            <p class="text-sm leading-7 text-sky-900 dark:text-cyan-100">
              📌 说明：GPAPI 为标准 RESTful API 服务，只要设备能发起 HTTPS 请求即可使用，不限操作系统。
            </p>
          </div>
        </article>
      </section>

      <section
        id="models"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
          模型支持矩阵
        </p>
        <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
          当前已接入模型
        </h2>
        <p class="mt-4 text-base leading-8 text-slate-600 dark:text-slate-300">
          支持 GPT 全系列，热门模型可以直接接入，无需额外适配。
        </p>
        <div class="mt-5 flex flex-wrap gap-3">
          <span
            v-for="model in featuredModels"
            :key="model"
            class="inline-flex rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-sm font-semibold text-slate-700 dark:border-white/10 dark:bg-slate-950/60 dark:text-slate-200"
          >
            {{ model }}
          </span>
        </div>

        <div class="mt-8">
          <div class="overflow-hidden rounded-[1.5rem] border border-slate-200/80 dark:border-white/10">
            <div class="overflow-x-auto">
              <table class="min-w-full border-collapse">
                <thead>
                  <tr class="bg-slate-50 dark:bg-slate-950/60">
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">模型系列</th>
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">具体模型</th>
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">适用场景</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="row in modelRows"
                    :key="row.series"
                    class="border-t border-slate-200/80 bg-white dark:border-white/10 dark:bg-slate-900/70"
                  >
                    <th class="whitespace-nowrap px-5 py-4 text-left text-sm font-bold text-slate-900 dark:text-white">
                      {{ row.series }}
                    </th>
                    <td class="px-5 py-4 text-sm leading-7 text-slate-600 dark:text-slate-300">
                      {{ row.models }}
                    </td>
                    <td class="px-5 py-4 text-sm leading-7 text-slate-600 dark:text-slate-300">
                      {{ row.scene }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <section
        id="ecosystem"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
          支持
        </p>
        <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
          已支持以下软件 / 客户端
        </h2>

        <div class="mt-8">
          <div class="w-full overflow-hidden rounded-[1.5rem] border border-slate-200/80 dark:border-white/10">
            <div class="overflow-x-auto">
              <table class="min-w-full border-collapse">
                <thead>
                  <tr class="bg-slate-50 dark:bg-slate-950/60">
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">软件 / 客户端</th>
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">接入方式</th>
                    <th class="px-5 py-4 text-left text-sm font-black text-slate-700 dark:text-slate-200">状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="row in ecosystemRows"
                    :key="row.tool"
                    class="border-t border-slate-200/80 bg-white dark:border-white/10 dark:bg-slate-900/70"
                  >
                    <th class="whitespace-nowrap px-5 py-4 text-left text-sm font-bold text-slate-900 dark:text-white">
                      {{ row.tool }}
                    </th>
                    <td class="px-5 py-4 text-sm leading-7 text-slate-600 dark:text-slate-300">
                      {{ row.integration }}
                    </td>
                    <td class="px-5 py-4 text-sm leading-7 text-blue-600 dark:text-cyan-300">
                      {{ row.tutorial }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <section
        id="pricing"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
          套餐选择
        </p>
        <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
          灵活套餐，按需选择
        </h2>
        <p class="mt-4 max-w-3xl text-base leading-8 text-slate-600 dark:text-slate-300">
          当前页面即可横向看完全部套餐，不必来回滚动对比。
        </p>

        <div class="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          <router-link
            v-for="row in pricingRows"
            :key="row.plan"
            to="/purchase"
            data-test="pricing-card"
            :data-plan-highlight="row.featured ? 'true' : 'false'"
            class="relative block overflow-hidden rounded-[1.5rem] border p-5 shadow-sm transition-transform duration-200 hover:-translate-y-1"
            :class="row.featured
              ? 'border-blue-300 bg-gradient-to-br from-blue-50 via-white to-cyan-50 shadow-blue-100/70 dark:border-cyan-400/30 dark:from-cyan-500/10 dark:via-slate-900/90 dark:to-slate-900/80'
              : 'border-slate-200/80 bg-white dark:border-white/10 dark:bg-slate-950/60'"
          >
            <div
              class="pointer-events-none absolute inset-x-6 top-0 h-20 rounded-b-[2rem] opacity-80 blur-2xl"
              :class="row.featured ? 'bg-blue-200/70 dark:bg-cyan-400/10' : 'bg-slate-200/50 dark:bg-white/5'"
            ></div>

            <div class="relative flex items-start justify-between gap-4">
              <div>
                <p
                  class="text-xs font-black uppercase tracking-[0.24em]"
                  :class="row.featured ? 'text-blue-700 dark:text-cyan-300' : 'text-slate-500 dark:text-slate-400'"
                >
                  {{ row.tier }}
                </p>
                <h3 class="mt-2 text-xl font-black tracking-tight text-slate-950 dark:text-white">
                  {{ row.plan }}
                </h3>
              </div>

              <span
                class="inline-flex rounded-full px-3 py-1 text-xs font-black"
                :class="row.featured
                  ? 'bg-blue-600 text-white dark:bg-cyan-300 dark:text-slate-950'
                  : row.badge === '限时特惠'
                    ? 'bg-amber-100 text-amber-800 dark:bg-amber-400/15 dark:text-amber-200'
                    : 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200'"
              >
                {{ row.badge }}
              </span>
            </div>

            <div class="relative mt-6">
              <p class="text-2xl font-black tracking-tight text-slate-950 dark:text-white">
                {{ row.price }}
              </p>
              <p class="mt-2 text-sm font-medium text-slate-500 dark:text-slate-400">
                {{ row.priceNote }}
              </p>
            </div>

            <div
              class="relative mt-5 rounded-[1.25rem] border px-4 py-4"
              :class="row.featured
                ? 'border-blue-200 bg-white/80 dark:border-cyan-400/20 dark:bg-slate-950/70'
                : 'border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-slate-900/70'"
            >
              <p class="text-sm leading-7 text-slate-700 dark:text-slate-200">
                {{ row.includes }}
              </p>
            </div>

            <div class="relative mt-5">
              <p class="text-xs font-black uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">
                适合对象
              </p>
              <p class="mt-2 text-sm leading-7 text-slate-700 dark:text-slate-200">
                {{ row.audience }}
              </p>
            </div>

            <ul class="relative mt-5 flex flex-wrap gap-2">
              <li
                v-for="feature in row.features"
                :key="feature"
                class="rounded-full px-3 py-1.5 text-xs font-semibold"
                :class="row.featured
                  ? 'bg-blue-100 text-blue-700 dark:bg-cyan-400/15 dark:text-cyan-200'
                  : 'bg-slate-100 text-slate-600 dark:bg-white/10 dark:text-slate-300'"
              >
                {{ feature }}
              </li>
            </ul>

            <div
              v-if="row.featured"
              class="relative mt-5 rounded-[1.15rem] border border-blue-200 bg-blue-600 px-4 py-3.5 text-white shadow-lg shadow-blue-200/50 dark:border-cyan-400/20 dark:bg-cyan-300 dark:text-slate-950 dark:shadow-cyan-500/10"
            >
              <p class="text-sm leading-7">
                最适合大多数开发者，兼顾成本、额度和稳定响应。
              </p>
            </div>
          </router-link>
        </div>
      </section>

      <section
        id="docs"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <div class="grid gap-8 xl:grid-cols-[0.9fr,1.1fr]">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
              技术文档 - 接入操作手册
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
              三步完成模型接入
            </h2>

            <div class="mt-8 space-y-4">
              <article
                v-for="step in docsSteps"
                :key="step.title"
                class="rounded-[1.35rem] border border-slate-200/80 bg-slate-50 p-5 dark:border-white/10 dark:bg-slate-950/60"
              >
                <p class="text-sm font-black uppercase tracking-[0.18em] text-blue-700 dark:text-cyan-300">{{ step.title }}</p>
                <p class="mt-3 whitespace-pre-line text-sm leading-7 text-slate-600 dark:text-slate-300">{{ step.description }}</p>
              </article>
            </div>
          </div>

          <div class="space-y-5">
            <article class="rounded-[1.5rem] border border-slate-200/80 bg-slate-950 shadow-sm dark:border-white/10">
              <div class="border-b border-white/10 px-6 py-4">
                <p class="text-sm font-black uppercase tracking-[0.2em] text-cyan-300">
                  Python 示例
                </p>
              </div>
              <div class="space-y-3 px-6 py-5 font-mono text-sm leading-7 text-slate-200">
                <div><span class="text-slate-500">from</span> <span class="text-cyan-300">openai</span> <span class="text-slate-500">import</span> <span class="text-emerald-300">OpenAI</span></div>
                <div class="pt-2"><span class="text-cyan-300">client</span> = <span class="text-violet-300">OpenAI</span>(</div>
                <div class="pl-4"><span class="text-slate-400">api_key=</span><span class="text-amber-300">"sk-gpapi-你的密钥"</span>,</div>
                <div class="pl-4"><span class="text-slate-400">base_url=</span><span class="text-emerald-300">"https://api.gpapi.com/v1"</span></div>
                <div>)</div>
                <div class="pt-2"><span class="text-cyan-300">response</span> = <span class="text-cyan-300">client.chat.completions.create</span>(</div>
                <div class="pl-4"><span class="text-slate-400">model=</span><span class="text-emerald-300">"gpt-4o"</span>,</div>
                <div class="pl-4"><span class="text-slate-400">messages=</span>[{ <span class="text-slate-400">"role"</span>: <span class="text-emerald-300">"user"</span>, <span class="text-slate-400">"content"</span>: <span class="text-emerald-300">"你好，GPAPI！"</span> }]</div>
                <div>)</div>
                <div class="pt-2"><span class="text-cyan-300">print</span>(<span class="text-cyan-300">response.choices[0].message.content</span>)</div>
              </div>
            </article>

            <router-link
              to="/guide"
              class="inline-flex w-full items-center justify-center rounded-xl bg-blue-600 px-6 py-3.5 text-base font-black text-white transition-colors hover:bg-blue-700"
            >
              查看完整 API 文档
            </router-link>
          </div>
        </div>
      </section>

      <section
        id="community"
        class="mx-auto mt-8 max-w-7xl rounded-[1.75rem] border border-slate-200/80 bg-slate-950 p-6 text-white shadow-2xl shadow-slate-950/10 dark:border-white/10 sm:p-8"
        :style="sectionOffsetStyle"
      >
        <div class="grid gap-6 lg:grid-cols-2">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.22em] text-cyan-300">
              技术社群
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight">
              交流群 / 公众号 / 接入支持一站式同步
            </h2>
            <p class="mt-4 text-base leading-8 text-slate-300">
              注册后即可通过控制台获取最新接入通知、模型上新公告、社群答疑与商务协作信息。
            </p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <article
              v-for="card in communityCards"
              :key="card.title"
              class="rounded-[1.35rem] border border-white/10 bg-white/5 p-5"
            >
              <p class="text-sm font-black uppercase tracking-[0.18em] text-cyan-300">{{ card.title }}</p>
              <p class="mt-3 text-sm leading-7 text-slate-300">{{ card.description }}</p>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 px-4 pb-10 pt-4 sm:px-6">
      <div
        class="mx-auto flex max-w-7xl flex-col items-center justify-center gap-4 rounded-[1.5rem] border border-slate-200/80 bg-white/70 px-6 py-5 text-center backdrop-blur-sm dark:border-white/10 dark:bg-slate-900/60 sm:flex-row"
      >
        <p class="text-sm text-slate-500 dark:text-slate-300">
          &copy; 2018 成都格品科技有限公司版权所有
          <a
            href="https://beian.miit.gov.cn/"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-700 hover:underline dark:hover:text-white"
          >
            蜀ICP备17044249号-1
          </a>
        </p>
        <router-link
          to="/guide"
          class="text-sm font-semibold text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-300 dark:hover:text-white"
        >
          API文档
        </router-link>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const authStore = useAuthStore()
const appStore = useAppStore()

const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const headerRef = ref<HTMLElement | null>(null)
const headerOffset = ref(120)

const navItems = [
  { href: '#top', label: '首页' },
  { href: '#pricing', label: '定价' },
  { href: '/guide', label: 'API文档' },
  { href: '#community', label: '技术社群' }
]

const heroAdvantages = [
  {
    icon: '🔒',
    title: '公司正规运营',
    description: '可开票、签合同、SLA 99.9% 可用性保障'
  },
  {
    icon: '⚡',
    title: '私有代理加速',
    description: '非公开线路，首 Token 延迟 < 80ms'
  },
  {
    icon: '🔄',
    title: '100% 官方兼容',
    description: '替换 Base URL 即刻接入，SDK 零改动'
  }
]

const heroTags = [
  '企业级 OpenAI 私有代理中转',
  '国内极速响应',
  '支持开票 / SLA',
  '无需绑卡'
]

const benefitRows = [
  { title: '网络稳定', description: '自建私有代理通道，减少断流与丢包。' },
  { title: '响应极速', description: '智能路由加速，首 Token 延迟更低。' },
  { title: '数据安全', description: '全程 TLS 加密，不记录 Prompt 明文。' },
  { title: '成本可控', description: '按量计费透明，支持余额预警。' },
  { title: '技术无忧', description: '专业运维值守，故障响应更快。' },
  { title: '合规安心', description: '支持合同、开票与企业结算。' }
]

const platformRows = [
  { platform: 'Windows', support: '✅ 所有支持 HTTP 请求的开发环境 / 客户端' },
  { platform: 'macOS', support: '✅ 原生终端、各类 AI 客户端完美兼容' },
  { platform: 'Linux', support: '✅ 服务器部署首选，支持 Docker 容器化调用' }
]

const featuredModels = [
  'GPT-5.2',
  'GPT-5.1',
  'GPT-5 mini',
  'GPT-4.1',
  'GPT-4.1 mini',
  'GPT-4o',
  'GPT-4o mini'
]

const modelRows = [
  { series: 'GPT-5 系列', models: 'GPT-5.2、GPT-5.1、GPT-5、GPT-5 mini', scene: '复杂推理、代码生成、Agent 工作流' },
  { series: 'GPT-4.1 系列', models: 'GPT-4.1、GPT-4.1 mini', scene: '长上下文任务、工具调用、稳定生产场景' },
  { series: 'GPT-4o 系列', models: 'GPT-4o、GPT-4o mini', scene: '多模态对话、低延迟通用交互' },
  { series: '图像生成', models: 'GPT Image 1.5', scene: '营销海报、商品图、插画生成' },
  { series: '音频能力', models: 'GPT Audio、GPT Audio mini', scene: '语音输入输出、实时音频交互' }
]

const ecosystemRows = [
  { tool: 'Codex', integration: '填入 GPAPI Base URL 与 API Key 即可接入', tutorial: '已支持' },
  { tool: 'VS Code', integration: '通过 OpenAI 兼容配置接入 GPAPI', tutorial: '已支持' },
  { tool: 'Chatbox', integration: '选择 OpenAI 兼容方式并填写端点与密钥', tutorial: '已支持' },
  { tool: 'OpenCode', integration: '自定义 API Provider 指向 GPAPI', tutorial: '已支持' },
  { tool: 'Cursor', integration: '配置 OpenAI 兼容接口即可使用', tutorial: '已支持' },
  { tool: 'OpenClaw', integration: '使用 OpenAI-compatible 模式连接 GPAPI', tutorial: '已支持' }
]

const communityCards = [
  { title: '交流群答疑', description: '接入调试、套餐咨询、故障反馈、最佳实践都可以在社群中快速同步。' },
  { title: '公众号更新', description: '模型上新、公告通知、优惠活动、维护窗口等信息第一时间获取。' }
]

const pricingRows = [
  {
    tier: '低门槛试用',
    plan: '个人体验版',
    price: '按量计费',
    priceNote: '先测通，再决定是否升级',
    includes: '￥0.00/月基础费 + 实际调用扣费',
    audience: '轻度使用、学习测试',
    features: ['无需包年', '适合学习'],
    badge: '按量灵活',
    featured: false
  },
  {
    tier: '轻量订阅',
    plan: '基础订阅',
    price: '￥19.9/月',
    priceNote: '每日 $5.00 额度',
    includes: '适合入门和低频稳定调用，开通后即可按日使用固定额度。',
    audience: '个人开发者、日常试用项目',
    features: ['每日 $5.00', '订阅制'],
    badge: '轻量入门',
    featured: false
  },
  {
    tier: '主力方案',
    plan: '标准订阅',
    price: '￥99/月',
    priceNote: '每日 $22.00 额度',
    includes: '适合稳定生产环境，覆盖更高频率的团队协作与业务请求。',
    audience: '小团队、正式业务应用',
    features: ['每日 $22.00', '稳定生产'],
    badge: '最受欢迎',
    featured: true
  },
  {
    tier: '高配方案',
    plan: '高级订阅',
    price: '￥299/月',
    priceNote: '每日 $68.00 额度',
    includes: '面向高并发和高消耗场景，日额度更充足，适合持续运行的业务。',
    audience: '高频调用团队、增长型业务',
    features: ['每日 $68.00', '高额度'],
    badge: '高性能',
    featured: false
  },
  {
    tier: '高阶定制',
    plan: '商务定制版',
    price: '商务洽谈',
    priceNote: '可按吞吐、部署形态与资源隔离单独规划',
    includes: '独享带宽、私有化部署与商务定制。',
    audience: '中大型企业、高并发场景',
    features: ['私有化部署', '独享带宽'],
    badge: '限时特惠',
    featured: false
  }
]

const docsSteps = [
  {
    title: 'Step 1：获取 API 密钥',
    description: '登录 GPAPI 控制台 → 左侧菜单「API 密钥」→ 点击「创建新密钥」\n复制保存形如 sk-gpapi-xxxxxxxxxxxxxxxx 的密钥串'
  },
  {
    title: 'Step 2：确定接入端点',
    description: 'GPAPI 网关地址：https://api.gpapi.com/v1\n该地址与 OpenAI 官方格式 100% 兼容'
  },
  {
    title: 'Step 3：修改代码 / 配置',
    description: '将现有 OpenAI SDK 的 base_url 指向 https://api.gpapi.com/v1\n保留原有调用方式即可开始发送请求'
  }
]

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const consolePath = computed(() => isAuthenticated.value ? dashboardPath.value : '/login')
const consoleLabel = computed(() => isAuthenticated.value ? '控制台' : '登录 / 注册')
const heroPrimaryLabel = computed(() => isAuthenticated.value ? '🚀 进入控制台' : '🚀 免费注册领 ￥5')

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
.grid-pattern {
  background-image:
    linear-gradient(to right, rgba(148, 163, 184, 0.12) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(148, 163, 184, 0.12) 1px, transparent 1px);
  background-size: 32px 32px;
}
</style>
