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
    class="home-shell relative min-h-screen overflow-hidden bg-[#f6f9ff] text-slate-950 transition-colors duration-300 dark:bg-[#040816] dark:text-white"
    :class="{ 'home-dark': isDark }"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="hero-grid absolute inset-0 opacity-60 dark:opacity-20"></div>
      <div class="hero-blur hero-blur-left absolute -left-20 top-20 h-[28rem] w-[28rem] rounded-full"></div>
      <div class="hero-blur hero-blur-right absolute -right-10 top-40 h-[24rem] w-[24rem] rounded-full"></div>
      <div class="hero-ribbon absolute inset-x-0 top-0 h-[36rem]"></div>
    </div>

    <HomeHeader
      v-model:header-ref="headerRef"
      :site-logo="siteLogo"
      :logo-tagline="t('home.landing.logoTagline')"
      :brand="t('home.landing.brand')"
      :nav-items="navItems"
      :theme-title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
      :is-dark="isDark"
      :console-path="consolePath"
      :console-label="consoleLabel"
      @toggle-theme="toggleTheme"
    />

    <main class="relative z-10 px-4 pb-16 sm:px-6" :style="mainStyle">
      <HomeHero
        :badge="t('home.landing.hero.badge')"
        :title-prefix="t('home.landing.hero.titlePrefix')"
        :title-lead="t('home.landing.hero.titleLead')"
        :title-highlight="t('home.landing.hero.titleHighlight')"
        :title-suffix="t('home.landing.hero.titleSuffix')"
        :subtitle="t('home.landing.hero.subtitle')"
        :checks="heroChecks"
        :console-path="consolePath"
        :primary-label="heroPrimaryLabel"
        :secondary-label="t('home.landing.hero.secondaryCta')"
        :doc-url="docUrl"
        :stats="heroStats"
        :header-offset="headerOffset"
      />

      <HomeFeatures
        :kicker="t('home.landing.overview.kicker')"
        :title="t('home.landing.overview.title')"
        :subtitle="t('home.landing.overview.subtitle')"
        :section-offset-style="sectionOffsetStyle"
        :feature-cards="featureCards"
      />

      <HomeProducts
        :kicker="t('home.landing.products.kicker')"
        :title="t('home.landing.products.title')"
        :subtitle="t('home.landing.products.subtitle')"
        :side-kicker="t('home.landing.models.kicker')"
        :side-title="t('home.landing.models.cardTitle')"
        :side-subtitle="t('home.landing.models.cardSubtitle')"
        :section-offset-style="sectionOffsetStyle"
        :cards="productCards"
        :items="modelChoices"
      />

      <section
        id="pricing"
        class="mx-auto mt-10 max-w-7xl rounded-[2rem] border border-slate-200/80 bg-white/92 px-6 py-8 shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/78 sm:px-8"
        :style="sectionOffsetStyle"
      >
        <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400">{{ t('home.landing.pricing.kicker') }}</p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white">{{ t('home.landing.pricing.title') }}</h2>
            <p class="mt-3 max-w-3xl text-sm leading-7 text-slate-500 dark:text-slate-300 sm:text-base">{{ t('home.landing.pricing.subtitle') }}</p>
          </div>
          <router-link
            to="/purchase"
            class="inline-flex items-center text-sm font-semibold text-blue-600 transition-colors hover:text-blue-700 dark:text-cyan-300 dark:hover:text-cyan-200"
          >
            {{ t('home.landing.pricing.viewDetails') }} →
          </router-link>
        </div>

        <div class="mt-8 grid gap-5 lg:grid-cols-3">
          <router-link
            v-for="row in pricingRows"
            :key="row.plan"
            to="/purchase"
            data-test="pricing-card"
            :data-plan-highlight="row.featured ? 'true' : 'false'"
            class="rounded-[1.8rem] border px-5 py-5 transition-transform hover:-translate-y-1"
            :class="row.featured
              ? 'border-blue-200 bg-gradient-to-br from-blue-50 via-white to-cyan-50 dark:border-cyan-400/20 dark:from-cyan-500/10 dark:via-slate-900/80 dark:to-blue-500/10'
              : 'border-slate-200/80 bg-slate-50/90 dark:border-slate-800/80 dark:bg-slate-950/85'"
          >
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <p class="text-base font-black text-slate-950 dark:text-white">{{ row.plan }}</p>
                  <span
                    class="inline-flex rounded-full px-2.5 py-1 text-[11px] font-black"
                    :class="row.featured
                      ? 'bg-blue-600 text-white dark:bg-cyan-300 dark:text-slate-950'
                      : 'bg-slate-200 text-slate-600 dark:bg-slate-800 dark:text-slate-100'"
                  >
                    {{ row.badge }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300">{{ row.description }}</p>
              </div>
              <p class="whitespace-nowrap text-xl font-black text-slate-950 dark:text-white">{{ row.price }}</p>
            </div>
          </router-link>
        </div>
      </section>

      <HomeShowcase
        :kicker="t('home.landing.showcase.kicker')"
        :title="t('home.landing.showcase.title')"
        :subtitle="t('home.landing.showcase.subtitle')"
        :section-offset-style="sectionOffsetStyle"
        :stats="showcaseStats"
        :quotes="showcaseQuotes"
        :quote-open="t('home.landing.showcase.quoteMarks.open')"
        :quote-close="t('home.landing.showcase.quoteMarks.close')"
      />

      <HomeCTA
        :badge="t('home.landing.cta.badge')"
        :title="t('home.landing.cta.title')"
        :subtitle="t('home.landing.cta.subtitle')"
        :console-path="consolePath"
        :primary-label="heroPrimaryLabel"
        :secondary-label="t('home.landing.hero.secondaryCta')"
        :doc-url="docUrl"
        :trust-items="ctaTrustItems"
        :section-offset-style="sectionOffsetStyle"
      />
    </main>

    <HomeFooter
      id="footer"
      :site-logo="siteLogo"
      :brand="t('home.landing.brand')"
      :about="t('home.landing.footer.about')"
      :social-dots="socialDots"
      :columns="footerColumns"
      :follow-label="t('home.landing.footer.follow')"
      :contact-items="footerContactItems"
      :qr-items="communityQrItems"
      :copyright-owner="t('home.landing.footer.copyrightOwner')"
      :filing-label="t('home.landing.footer.filing')"
      :doc-url="docUrl"
      :docs-label="t('home.landing.nav.docs')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import HomeCTA from '@/components/home/HomeCTA.vue'
import HomeFeatures from '@/components/home/HomeFeatures.vue'
import HomeFooter from '@/components/home/HomeFooter.vue'
import HomeHeader from '@/components/home/HomeHeader.vue'
import HomeHero from '@/components/home/HomeHero.vue'
import HomeProducts from '@/components/home/HomeProducts.vue'
import HomeShowcase from '@/components/home/HomeShowcase.vue'
import { TOKEN_DOC_URL } from '@/constants/externalLinks'
import type { HomeLink } from '@/types'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const docUrl = TOKEN_DOC_URL
const communityQrItems = computed(() => [
  { src: '/qq.jpg', alt: t('home.landing.community.qrAlt.qq'), label: t('home.landing.community.qrLabels.qq') },
  { src: '/wechat.jpg', alt: t('home.landing.community.qrAlt.wechat'), label: t('home.landing.community.qrLabels.wechat') }
])
function splitHomepageContact(value: string | undefined) {
  return (value || '')
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

const footerContactItems = computed(() =>
  [
    { label: t('home.landing.community.qrLabels.qq'), values: splitHomepageContact(appStore.cachedPublicSettings?.qq_group) },
    { label: t('home.landing.community.qrLabels.wechat'), values: splitHomepageContact(appStore.cachedPublicSettings?.wechat_contact) }
  ].filter((item) => item.values.length)
)

const headerRef = ref<HTMLElement | null>(null)
const headerOffset = ref(120)

const defaultHomeLinks = computed<HomeLink[]>(() => [
  {
    id: 'gpshop',
    label: t('home.landing.nav.gpshop'),
    label_zh: t('home.landing.nav.gpshop'),
    label_en: 'GePin Shop',
    url: 'https://card.gepinkeji.com',
    enabled: true,
    sort_order: 0
  },
  {
    id: 'gpci',
    label: t('home.landing.nav.gpci'),
    label_zh: t('home.landing.nav.gpci'),
    label_en: 'GePin Image',
    url: 'https://chat.gepinkeji.com/',
    enabled: true,
    sort_order: 1
  }
])

function resolveHomeLinkLabel(link: HomeLink) {
  if (locale.value.startsWith('zh')) {
    return link.label_zh || link.label || link.label_en || ''
  }
  return link.label_en || link.label || link.label_zh || ''
}

const visibleHomeLinks = computed(() => {
  const configuredLinks = appStore.cachedPublicSettings?.home_links || []
  const links = configuredLinks.length > 0 ? configuredLinks : defaultHomeLinks.value

  return links
    .filter((link) => link.enabled)
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
    .map((link) => ({
      href: link.url,
      label: resolveHomeLinkLabel(link),
      external: true
    }))
    .filter((link) => link.label && link.href)
})

const navItems = computed(() => [
  { href: '#top', label: t('home.landing.nav.home') },
  { href: '#pricing', label: t('home.landing.nav.pricing') },
  ...visibleHomeLinks.value,
  { href: docUrl, label: t('home.landing.nav.docs'), external: true },
  { href: '#footer', label: t('home.landing.nav.community') }
])

const heroChecks = computed(() => [
  t('home.landing.hero.checks.availability'),
  t('home.landing.hero.checks.latency'),
  t('home.landing.hero.checks.security')
])

const heroStats = computed(() => [
  { value: '1.2M+', label: t('home.landing.hero.stats.calls') },
  { value: '50,000+', label: t('home.landing.hero.stats.developers') },
  { value: '99.9%', label: t('home.landing.hero.stats.sla') },
  { value: '24/7', label: t('home.landing.hero.stats.support') }
])

const featureCards = computed<Array<{ icon: 'shield' | 'swap' | 'chartBar'; title: string; description: string; emphasis?: boolean }>>(() => [
  {
    icon: 'shield',
    title: t('home.landing.overview.benefits.security.title'),
    description: t('home.landing.overview.benefits.security.description'),
    emphasis: true
  },
  {
    icon: 'swap',
    title: t('home.landing.overview.benefits.easy.title'),
    description: t('home.landing.overview.benefits.easy.description')
  },
  {
    icon: 'chartBar',
    title: t('home.landing.overview.benefits.stability.title'),
    description: t('home.landing.overview.benefits.stability.description')
  }
])

const productCards = computed<Array<{ icon: 'cpu' | 'sparkles' | 'globe' | 'shield'; title: string; description: string }>>(() => [
  {
    icon: 'cpu',
    title: t('home.landing.products.cards.infrastructure.title'),
    description: t('home.landing.products.cards.infrastructure.description')
  },
  {
    icon: 'sparkles',
    title: t('home.landing.products.cards.integration.title'),
    description: t('home.landing.products.cards.integration.description')
  },
  {
    icon: 'globe',
    title: t('home.landing.products.cards.delivery.title'),
    description: t('home.landing.products.cards.delivery.description')
  },
  {
    icon: 'shield',
    title: t('home.landing.products.cards.security.title'),
    description: t('home.landing.products.cards.security.description')
  }
])

const modelChoices = computed(() => [
  { name: 'GPT-5.5', description: t('home.landing.models.items.gpt55'), badge: t('home.landing.models.badges.latest') },
  { name: 'GPT-5.4', description: t('home.landing.models.items.gpt54'), badge: '' },
  { name: 'GPT-5.4 nano', description: t('home.landing.models.items.gpt54nano'), badge: '' },
  { name: 'GPT Image 2', description: t('home.landing.models.items.gptImage2'), badge: '' }
])

const pricingRows = computed(() => [
  { plan: t('home.landing.pricing.plans.standard.plan'), price: t('home.landing.pricing.plans.standard.price'), description: t('home.landing.pricing.plans.standard.description'), badge: t('home.landing.pricing.plans.standard.badge'), featured: false },
  { plan: t('home.landing.pricing.plans.pro.plan'), price: t('home.landing.pricing.plans.pro.price'), description: t('home.landing.pricing.plans.pro.description'), badge: t('home.landing.pricing.plans.pro.badge'), featured: false },
  { plan: t('home.landing.pricing.plans.enterprise.plan'), price: t('home.landing.pricing.plans.enterprise.price'), description: t('home.landing.pricing.plans.enterprise.description'), badge: t('home.landing.pricing.plans.enterprise.badge'), featured: true }
])

const showcaseStats = computed<Array<{ icon: 'users' | 'shield' | 'sparkles'; value: string; label: string }>>(() => [
  { icon: 'users', value: '50,000+', label: t('home.landing.showcase.stats.developers') },
  { icon: 'shield', value: '99.9%', label: t('home.landing.showcase.stats.sla') },
  { icon: 'sparkles', value: '24/7', label: t('home.landing.showcase.stats.support') }
])

const showcaseQuotes = computed(() => [
  { quote: t('home.landing.showcase.quotes.team.quote'), author: t('home.landing.showcase.quotes.team.author'), role: t('home.landing.showcase.quotes.team.role') },
  { quote: t('home.landing.showcase.quotes.engineer.quote'), author: t('home.landing.showcase.quotes.engineer.author'), role: t('home.landing.showcase.quotes.engineer.role') },
  { quote: t('home.landing.showcase.quotes.ops.quote'), author: t('home.landing.showcase.quotes.ops.author'), role: t('home.landing.showcase.quotes.ops.role') }
])

const socialDots = computed(() => [
  t('home.landing.community.socialDots.wechat'),
  t('home.landing.community.socialDots.blog'),
  t('home.landing.community.socialDots.video'),
  t('home.landing.community.socialDots.qa')
])

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

const footerColumns = computed(() => [
  {
    title: t('home.landing.footer.columns.product.title'),
    items: [
      { label: t('home.landing.footer.columns.product.pricing'), href: '#pricing' },
      { label: t('home.landing.footer.columns.product.docs'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.product.platforms'), href: '#overview' },
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
    title: t('home.landing.footer.columns.links.title'),
    items: visibleHomeLinks.value
  },
  {
    title: t('home.landing.footer.columns.company.title'),
    items: [
      { label: t('home.landing.footer.columns.company.about'), href: '#overview' },
      { label: t('home.landing.footer.columns.company.contact'), href: '#footer' },
      { label: t('home.landing.footer.columns.company.terms'), href: docUrl, external: true },
      { label: t('home.landing.footer.columns.company.privacy'), href: docUrl, external: true }
    ]
  }
])

const ctaTrustItems = computed(() => [
  t('home.landing.cta.trust.freeTrial'),
  t('home.landing.cta.trust.noCard'),
  t('home.landing.cta.trust.fastSupport')
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
    radial-gradient(circle at 14% 20%, rgba(255, 255, 255, 0.96) 0, rgba(255, 255, 255, 0) 28%),
    linear-gradient(180deg, rgba(239, 245, 255, 0.96) 0%, rgba(246, 249, 255, 1) 32%, rgba(248, 250, 255, 1) 100%);
}

.hero-grid {
  background-image:
    linear-gradient(to right, rgba(15, 23, 42, 0.05) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(15, 23, 42, 0.05) 1px, transparent 1px);
  background-size: 72px 72px;
}

.hero-blur {
  filter: blur(70px);
  opacity: 0.9;
}

.hero-blur-left {
  background: rgba(96, 165, 250, 0.18);
}

.hero-blur-right {
  background: rgba(125, 211, 252, 0.2);
}

.hero-ribbon {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.74) 0%, rgba(255, 255, 255, 0.2) 58%, rgba(255, 255, 255, 0) 100%),
    radial-gradient(circle at 50% 20%, rgba(96, 165, 250, 0.14), rgba(96, 165, 250, 0) 40%);
}

.home-dark {
  background-image:
    radial-gradient(circle at 78% 14%, rgba(34, 211, 238, 0.08) 0, rgba(34, 211, 238, 0) 24%),
    radial-gradient(circle at 18% 20%, rgba(59, 130, 246, 0.08) 0, rgba(59, 130, 246, 0) 26%),
    linear-gradient(180deg, rgba(4, 8, 22, 1) 0%, rgba(7, 12, 26, 1) 40%, rgba(4, 8, 20, 1) 100%);
}
</style>
