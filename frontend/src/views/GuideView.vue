<template>
  <div
    id="top"
    class="relative min-h-screen overflow-hidden bg-[#f6f7fb] text-slate-950 dark:bg-slate-950 dark:text-white"
  >
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute inset-0 grid-pattern opacity-70 dark:opacity-15"></div>
      <div class="absolute inset-x-0 top-0 h-80 bg-gradient-to-b from-blue-100/70 to-transparent dark:from-cyan-500/10"></div>
      <div class="absolute -left-20 top-32 h-72 w-72 rounded-full bg-cyan-200/40 blur-3xl dark:bg-cyan-400/10"></div>
      <div class="absolute -right-24 top-96 h-72 w-72 rounded-full bg-amber-200/40 blur-3xl dark:bg-amber-300/10"></div>
    </div>

    <header class="fixed inset-x-0 top-0 z-50 px-4 py-4 sm:px-6">
      <nav
        class="mx-auto flex max-w-7xl items-center justify-between rounded-2xl border border-slate-200/80 bg-white/92 px-4 py-3 shadow-lg shadow-slate-200/30 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/82 dark:shadow-black/30"
      >
        <a href="#top" class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-white shadow-sm ring-1 ring-slate-200/70 dark:bg-slate-900 dark:ring-white/10">
            <span class="text-lg font-black tracking-tight text-slate-950 dark:text-white">G</span>
          </div>
          <div>
            <p class="text-[11px] font-semibold uppercase tracking-[0.35em] text-slate-400 dark:text-slate-500">
              Docs Hub
            </p>
            <p class="text-lg font-black tracking-tight">GPAPI Guide</p>
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
            to="/home"
            class="inline-flex items-center rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:border-slate-300 hover:text-slate-950 dark:border-white/10 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-white/20 dark:hover:text-white"
          >
            返回首页
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 px-4 pb-12 pt-[120px] sm:px-6">
      <section class="mx-auto max-w-7xl pb-10 pt-6">
        <div class="rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8">
          <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
            文档中心
          </p>
          <h1 class="mt-3 text-4xl font-black tracking-[-0.04em] text-slate-950 dark:text-white sm:text-5xl">
            {{ entrySection?.title }}
          </h1>
          <p class="mt-4 max-w-4xl text-base leading-8 text-slate-600 dark:text-slate-300 sm:text-lg">
            以总入口文档作为首页，将所有客户端教程按文档顺序连续编排到同一个阅读页面中。页面结构强调目录、章节和正文本身，便于快速定位、搜索与连续阅读。
          </p>

          <div class="mt-6 flex flex-wrap gap-3 text-sm text-slate-500 dark:text-slate-300">
            <span
              v-for="tag in heroTags"
              :key="tag"
              class="rounded-full border border-slate-200 bg-white px-3 py-1.5 dark:border-white/10 dark:bg-slate-950/60"
            >
              {{ tag }}
            </span>
          </div>
        </div>
      </section>

      <section class="mx-auto grid max-w-7xl gap-8 lg:grid-cols-[280px,minmax(0,1fr)]">
        <aside class="lg:sticky lg:top-[112px] lg:self-start">
          <div class="rounded-[1.5rem] border border-slate-200/80 bg-white/92 p-5 shadow-sm dark:border-white/10 dark:bg-slate-900/70">
            <p class="text-sm font-black uppercase tracking-[0.2em] text-blue-700 dark:text-cyan-300">
              文档目录
            </p>
            <div class="mt-5 space-y-5">
              <section
                v-for="doc in docSections"
                :key="`nav-${doc.id}`"
                class="border-l border-slate-200 pl-4 dark:border-white/10"
              >
                <a
                  :href="`#${doc.id}`"
                  class="block text-sm font-black text-slate-900 transition-colors hover:text-blue-700 dark:text-white dark:hover:text-cyan-300"
                >
                  {{ doc.title }}
                </a>
                <p class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400">
                  {{ doc.summary }}
                </p>
                <div v-if="doc.headings.length > 0" class="mt-3 space-y-2">
                  <a
                    v-for="heading in doc.headings"
                    :key="heading.id"
                    :href="`#${heading.id}`"
                    class="block text-sm leading-6 text-slate-600 transition-colors hover:text-slate-950 dark:text-slate-300 dark:hover:text-white"
                    :class="heading.level === 3 ? 'pl-4 text-[13px]' : ''"
                  >
                    {{ heading.text }}
                  </a>
                </div>
              </section>
            </div>
          </div>
        </aside>

        <div class="space-y-8">
          <section class="rounded-[1.5rem] border border-slate-200/80 bg-slate-950 text-white shadow-sm dark:border-white/10">
            <div class="border-b border-white/10 px-6 py-4">
              <p class="text-sm font-black uppercase tracking-[0.2em] text-cyan-300">
                快速接入模板
              </p>
            </div>
            <div class="space-y-3 px-6 py-5 font-mono text-sm leading-7 text-slate-200">
              <div><span class="text-slate-500">provider</span> = <span class="text-emerald-300">"OpenAI Compatible"</span></div>
              <div><span class="text-slate-500">base_url</span> = <span class="text-amber-300">"https://token.gepinkeji.com/v1"</span></div>
              <div><span class="text-slate-500">api_key</span> = <span class="text-cyan-300">"sk-xxxx"</span></div>
              <div><span class="text-slate-500">model</span> = <span class="text-violet-300">"GPT-5.4"</span></div>
            </div>
          </section>

          <section
            v-for="doc in docSections"
            :id="doc.id"
            :key="doc.id"
            class="rounded-[1.75rem] border border-slate-200/80 bg-white/92 p-6 shadow-sm dark:border-white/10 dark:bg-slate-900/70 sm:p-8"
          >
        <div class="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <p class="text-sm font-black uppercase tracking-[0.22em] text-blue-700 dark:text-cyan-300">
              {{ doc.isEntry ? '总入口文档' : '客户端子文档' }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">
              {{ doc.title }}
            </h2>
            <p class="mt-4 text-base leading-8 text-slate-600 dark:text-slate-300">
              {{ doc.summary }}
            </p>
          </div>

          <a
            href="#top"
            class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-slate-50 px-5 py-3 text-sm font-semibold text-slate-700 transition-colors hover:border-slate-300 hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-slate-950/60 dark:text-slate-200 dark:hover:border-white/20 dark:hover:text-white"
          >
            返回顶部
          </a>
        </div>

        <div
          class="markdown-body prose prose-slate max-w-none"
          :class="isDark ? 'markdown-body-dark' : 'markdown-body-light'"
          v-html="doc.html"
        ></div>
          </section>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { marked, type Tokens } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { guideDocSections, type GuideDocHeading } from './guide/generated-guide-content'

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

marked.setOptions({
  breaks: true,
  gfm: true,
})

marked.use({
  renderer: {
    code({ text }: Tokens.Code) {
      return `<pre><code>${escapeHtml(text)}</code></pre>`
    },
  },
})

function injectHeadingIds(html: string, headings: GuideDocHeading[]) {
  let index = 0

  return html.replace(/<h([23])>(.*?)<\/h\1>/g, (match, level, content) => {
    const heading = headings[index]
    if (!heading || heading.level !== Number(level)) {
      return match
    }

    index += 1
    return `<h${level} id="${heading.id}">${content}</h${level}>`
  })
}

const docSections = computed(() =>
  guideDocSections.map((doc) => {
    const rawHtml = marked.parse(doc.markdown) as string

    return {
      ...doc,
      html: DOMPurify.sanitize(injectHeadingIds(rawHtml, doc.headings)),
    }
  }),
)

const entrySection = computed(() => docSections.value.find((doc) => doc.isEntry))

const navItems = [
  { label: '目录', href: '#doc-guide' },
  { label: '总入口', href: '#doc-guide' },
  { label: '客户端章节', href: '#doc-codex' },
]

const heroTags = ['总入口', '文档目录', '连续阅读', '锚点导航', '图片直显']

const isDark = ref(document.documentElement.classList.contains('dark'))
let themeObserver: MutationObserver | null = null

function syncThemeFromDocument() {
  isDark.value = document.documentElement.classList.contains('dark')
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  syncThemeFromDocument()

  themeObserver = new MutationObserver(() => {
    syncThemeFromDocument()
  })

  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
})

onBeforeUnmount(() => {
  themeObserver?.disconnect()
  themeObserver = null
})
</script>

<style scoped>
.markdown-body {
  background: transparent;
  color: var(--guide-markdown-text);
}

.markdown-body-light {
  --guide-markdown-text: #475569;
  --guide-markdown-strong: #0f172a;
  --guide-markdown-link: #2563eb;
  --guide-markdown-link-hover: #1d4ed8;
  --guide-markdown-heading: #0f172a;
  --guide-markdown-pre-bg: rgba(248, 250, 252, 0.95);
  --guide-markdown-pre-border: rgba(148, 163, 184, 0.28);
  --guide-markdown-inline-code-bg: rgba(148, 163, 184, 0.14);
  --guide-markdown-inline-code-text: #be123c;
  --guide-markdown-table-bg: rgba(255, 255, 255, 0.82);
  --guide-markdown-table-border: rgba(148, 163, 184, 0.28);
  --guide-markdown-thead-bg: rgba(241, 245, 249, 0.95);
  --guide-markdown-tbody-bg: rgba(255, 255, 255, 0.9);
  --guide-markdown-cell-border: rgba(148, 163, 184, 0.25);
  --guide-markdown-cell-text: #334155;
  --guide-markdown-blockquote-bg: rgba(96, 165, 250, 0.08);
  --guide-markdown-blockquote-text: #334155;
  --guide-markdown-hr-border: rgba(148, 163, 184, 0.3);
}

.markdown-body-dark {
  --guide-markdown-text: #cbd5e1;
  --guide-markdown-strong: #f8fafc;
  --guide-markdown-link: #7dd3fc;
  --guide-markdown-link-hover: #bae6fd;
  --guide-markdown-heading: #f8fafc;
  --guide-markdown-pre-bg: rgba(15, 23, 42, 0.72);
  --guide-markdown-pre-border: rgba(71, 85, 105, 0.85);
  --guide-markdown-inline-code-bg: rgba(51, 65, 85, 0.9);
  --guide-markdown-inline-code-text: #fda4af;
  --guide-markdown-table-bg: rgba(15, 23, 42, 0.7);
  --guide-markdown-table-border: rgba(71, 85, 105, 0.9);
  --guide-markdown-thead-bg: rgba(30, 41, 59, 0.95);
  --guide-markdown-tbody-bg: rgba(15, 23, 42, 0.72);
  --guide-markdown-cell-border: rgba(71, 85, 105, 0.85);
  --guide-markdown-cell-text: #e2e8f0;
  --guide-markdown-blockquote-bg: rgba(30, 64, 175, 0.18);
  --guide-markdown-blockquote-text: #cbd5e1;
  --guide-markdown-hr-border: rgba(71, 85, 105, 0.8);
}

.markdown-body :deep(p),
.markdown-body :deep(ul),
.markdown-body :deep(ol),
.markdown-body :deep(li) {
  color: inherit;
}

.markdown-body :deep(strong) {
  color: var(--guide-markdown-strong);
}

.markdown-body :deep(a) {
  color: var(--guide-markdown-link);
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  color: var(--guide-markdown-link-hover);
  text-decoration: underline;
}

.markdown-body :deep(h1) {
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.markdown-body :deep(h2) {
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.markdown-body :deep(h3) {
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  font-size: 1.2rem;
  font-weight: 800;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  color: var(--guide-markdown-heading);
}

.markdown-body :deep(img) {
  border-radius: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.3);
  box-shadow: 0 10px 30px -18px rgba(15, 23, 42, 0.4);
}

.markdown-body :deep(pre) {
  border-radius: 1rem;
  border: 1px solid var(--guide-markdown-pre-border);
  background: var(--guide-markdown-pre-bg);
  padding: 1rem 1.25rem;
  color: var(--guide-markdown-strong);
}

.markdown-body :deep(pre code) {
  display: block;
  background: transparent;
  color: inherit;
  padding: 0;
}

.markdown-body :deep(code:not(pre code)) {
  border-radius: 0.5rem;
  background: var(--guide-markdown-inline-code-bg);
  color: var(--guide-markdown-inline-code-text);
  padding: 0.15rem 0.45rem;
}

.markdown-body :deep(code) {
  word-break: break-word;
}

.markdown-body :deep(table) {
  display: block;
  width: 100%;
  overflow-x: auto;
  border-radius: 1rem;
  border: 1px solid var(--guide-markdown-table-border);
  background: var(--guide-markdown-table-bg);
}

.markdown-body :deep(thead tr) {
  background: var(--guide-markdown-thead-bg);
}

.markdown-body :deep(tbody tr) {
  background: var(--guide-markdown-tbody-bg);
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border-color: var(--guide-markdown-cell-border);
  color: var(--guide-markdown-cell-text);
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #60a5fa;
  background: var(--guide-markdown-blockquote-bg);
  padding: 1rem 1.25rem;
  border-radius: 0 1rem 1rem 0;
  color: var(--guide-markdown-blockquote-text);
}

.markdown-body :deep(hr) {
  border-color: var(--guide-markdown-hr-border);
}
</style>
