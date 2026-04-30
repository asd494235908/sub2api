import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const authStoreMock = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  user: null as null | { email?: string },
  checkAuth: vi.fn()
}))

const appStoreMock = vi.hoisted(() => ({
  cachedPublicSettings: null as null | Record<string, string>,
  siteName: 'PrecisionAPI',
  siteLogo: '',
  siteSubtitle: 'One Key, All AI Models',
  docUrl: 'https://docs.example.com',
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn()
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreMock,
  useAppStore: () => appStoreMock
}))

const messages: Record<string, string> = {}
const defaultMessages: Record<string, string> = {
  'home.switchToLight': '切换浅色模式',
  'home.switchToDark': '切换深色模式',
  'home.landing.brand': '格品API',
  'home.landing.logoTagline': '企业级 OpenAI 中转',
  'home.landing.nav.home': '首页',
  'home.landing.nav.pricing': '定价',
  'home.landing.nav.docs': 'API文档',
  'home.landing.nav.community': '技术社群',
  'home.landing.console.login': '立即使用',
  'home.landing.console.enter': '控制台',
  'home.landing.hero.badge': '稳定 · 安全 · 高效的 OpenAI 中转服务',
  'home.landing.hero.titlePrefix': '格品API',
  'home.landing.hero.titleLead': '让 ',
  'home.landing.hero.titleHighlight': 'AI',
  'home.landing.hero.titleSuffix': '开发更简单',
  'home.landing.hero.subtitle': '企业级 OpenAI API 中转服务 | 国内极速响应 | 一站式接入',
  'home.landing.hero.primaryCtaGuest': '立即开始接入',
  'home.landing.hero.primaryCtaAuthed': '进入控制台',
  'home.landing.hero.secondaryCta': '查看文档',
  'home.landing.hero.promoTitle': '开始构建你的 AI 应用',
  'home.landing.hero.promoSubtitle': '简单几步，快速接入强大的 AI 能力',
  'home.landing.hero.checks.availability': '99.9% 可用性',
  'home.landing.hero.checks.latency': '极速响应 < 200ms',
  'home.landing.hero.checks.security': '企业级安全防护',
  'home.landing.hero.stats.calls': 'API 调用次数',
  'home.landing.hero.stats.developers': '开发者信赖',
  'home.landing.hero.stats.sla': '可用性保障',
  'home.landing.hero.stats.support': '技术支持',
  'home.landing.overview.kicker': '为什么选择 格品API',
  'home.landing.overview.title': '快速、稳定、安全的 AI 接入体验',
  'home.landing.overview.subtitle': '我们提供稳定、快速、安全的 OpenAI API 中转服务，让业务专注于产品创新。',
  'home.landing.overview.benefits.stability.title': '高速稳定',
  'home.landing.overview.benefits.stability.description': '全链路加速与弹性调度，保障高峰期也能稳定返回。',
  'home.landing.overview.benefits.security.title': '企业级安全',
  'home.landing.overview.benefits.security.description': '全程 TLS 加密与独立通道设计，兼顾安全与合规。',
  'home.landing.overview.benefits.value.title': '高性价比',
  'home.landing.overview.benefits.value.description': '按量与套餐并存，覆盖从试用到生产的大部分场景。',
  'home.landing.overview.benefits.easy.title': '简单易用',
  'home.landing.overview.benefits.easy.description': '兼容 OpenAI 接口规范，替换 Base URL 即可接入。',
  'home.landing.overview.benefits.support.title': '专业支持',
  'home.landing.overview.benefits.support.description': '7x24 小时技术支持与维护通知，协助快速排障。',
  'home.landing.overview.benefits.updates.title': '持续更新',
  'home.landing.overview.benefits.updates.description': '紧跟 OpenAI 最新模型能力，快速同步可用模型。',
  'home.landing.platforms.kicker': '支持的平台',
  'home.landing.platforms.title': '支持的操作系统与设备',
  'home.landing.platforms.subtitle': '支持主流平台，随时随地接入 API',
  'home.landing.platforms.mobile': '移动端',
  'home.landing.platforms.devices': '开发设备',
  'home.landing.models.kicker': '丰富的模型选择',
  'home.landing.models.cardTitle': '接入 OpenAI 前沿模型',
  'home.landing.models.cardSubtitle': '覆盖当前 API 可用的主流推理、轻量与图像场景。',
  'home.landing.models.apiNotice': '最新支持 GPT-5.5，适合复杂推理、编码与专业工作流。',
  'home.landing.models.viewAll': '查看所有模型',
  'home.landing.models.badges.latest': '当前最新',
  'home.landing.models.items.gpt55': 'OpenAI 最新支持模型，适合复杂推理、编码与专业工作流。',
  'home.landing.models.items.gpt54Mini': '更高性价比的轻量版本，适合编码、工具调用与低延迟任务。',
  'home.landing.models.items.gpt54Nano': '面向高频与基础任务的超低成本版本。',
  'home.landing.models.items.gptImage2': 'OpenAI 当前图像生成模型，适合图片生成与编辑场景。',
  'home.landing.pricing.kicker': '灵活的计费方案',
  'home.landing.pricing.title': '灵活套餐，按需选择',
  'home.landing.pricing.subtitle': '按量计费、轻量订阅与企业级方案都能覆盖。',
  'home.landing.pricing.viewDetails': '查看详细价格',
  'home.landing.pricing.plans.standard.plan': '标准套餐',
  'home.landing.pricing.plans.standard.description': '适合个人开发者的轻度项目',
  'home.landing.pricing.plans.standard.badge': '轻量版',
  'home.landing.pricing.plans.pro.plan': '专业套餐',
  'home.landing.pricing.plans.pro.description': '适合中型团队和稳定项目',
  'home.landing.pricing.plans.pro.badge': '推荐',
  'home.landing.pricing.plans.enterprise.plan': '企业套餐',
  'home.landing.pricing.plans.enterprise.description': '适合大型企业和定制化需求',
  'home.landing.pricing.plans.enterprise.badge': '企业版',
  'home.landing.pricing.plans.enterprise.price': '定制价格',
  'home.landing.setup.kicker': '三步快速接入',
  'home.landing.setup.title': '轻松完成模型接入',
  'home.landing.setup.subtitle': '简单几步，就能开始使用。',
  'home.landing.setup.steps.register.title': '注册账号',
  'home.landing.setup.steps.register.description': '创建账户并获取专属 API Key。',
  'home.landing.setup.steps.integrate.title': '接入 API',
  'home.landing.setup.steps.integrate.description': '替换文档中的接入地址，配置应用即可。',
  'home.landing.setup.steps.start.title': '开始调用',
  'home.landing.setup.steps.start.description': '快速调用 API，构建你的 AI 应用。',
  'home.landing.community.kicker': '加入开发者社区',
  'home.landing.community.title': '社群二维码',
  'home.landing.community.lead': '扫码加入 QQ 群或微信',
  'home.landing.community.body': '获取接入支持、模型上新通知与问题答疑。',
  'home.landing.community.qrLabels.qq': 'QQ 群',
  'home.landing.community.qrLabels.wechat': '微信',
  'home.landing.community.qrAlt.qq': 'QQ群二维码',
  'home.landing.community.qrAlt.wechat': '微信二维码',
  'home.landing.footer.about': '企业级 OpenAI API 中转服务，让 AI 开发更简单。',
  'home.landing.footer.follow': '关注我们',
  'home.landing.footer.copyrightOwner': '成都格品科技有限公司版权所有'
  ,'home.landing.footer.columns.product.title': '产品'
  ,'home.landing.footer.columns.product.pricing': '定价'
  ,'home.landing.footer.columns.product.docs': 'API 文档'
  ,'home.landing.footer.columns.product.platforms': '支持的平台'
  ,'home.landing.footer.columns.product.changelog': '更新日志'
  ,'home.landing.footer.columns.developer.title': '开发者'
  ,'home.landing.footer.columns.developer.quickstart': '快速开始'
  ,'home.landing.footer.columns.developer.sdk': 'SDK & 工具'
  ,'home.landing.footer.columns.developer.bestPractices': '最佳实践'
  ,'home.landing.footer.columns.developer.status': 'API 状态'
  ,'home.landing.footer.columns.company.title': '公司'
  ,'home.landing.footer.columns.company.about': '关于我们'
  ,'home.landing.footer.columns.company.contact': '联系我们'
  ,'home.landing.footer.columns.company.terms': '服务条款'
  ,'home.landing.footer.columns.company.privacy': '隐私政策'
  ,'home.landing.footer.columns.support.title': '支持'
  ,'home.landing.footer.columns.support.help': '帮助中心'
  ,'home.landing.footer.columns.support.community': '技术社群'
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key
    })
  }
})

const TOKEN_DOC_URL = 'https://token.gepinkeji.com/tokenDoc'

function createWrapper() {
  return mount(HomeView, {
    global: {
      stubs: {
        LocaleSwitcher: {
          template: '<div data-test="locale-switcher">LocaleSwitcher</div>'
        },
        Icon: {
          props: ['name'],
          template: '<span data-test="icon" :data-icon-name="name" />'
        },
        'router-link': {
          props: ['to'],
          computed: {
            href(): string {
              if (typeof this.to === 'string') {
                return this.to
              }

              const path = this.to?.path || ''
              const query = this.to?.query

              if (!query) {
                return path
              }

              const search = new URLSearchParams(query).toString()
              return search ? `${path}?${search}` : path
            }
          },
          template: '<a :href="href"><slot /></a>'
        }
      }
    }
  })
}

describe('HomeView', () => {
  beforeEach(() => {
    Object.keys(messages).forEach((key) => {
      delete messages[key]
    })
    Object.assign(messages, defaultMessages)

    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(() => ({
        matches: false,
        media: '',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn()
      }))
    })

    authStoreMock.isAuthenticated = false
    authStoreMock.isAdmin = false
    authStoreMock.user = null
    authStoreMock.checkAuth.mockReset()

    appStoreMock.cachedPublicSettings = null
    appStoreMock.siteName = 'PrecisionAPI'
    appStoreMock.siteLogo = ''
    appStoreMock.siteSubtitle = 'One Key, All AI Models'
    appStoreMock.docUrl = 'https://docs.example.com'
    appStoreMock.publicSettingsLoaded = true
    appStoreMock.fetchPublicSettings.mockReset()
  })

  it('renders iframe mode when custom home content is an external url', () => {
    appStoreMock.cachedPublicSettings = {
      home_content: 'https://landing.example.com'
    }

    const wrapper = createWrapper()

    const iframe = wrapper.get('iframe')
    expect(iframe.attributes('src')).toBe('https://landing.example.com')
  })

  it('applies the dedicated homepage dark theme class when dark mode is active', () => {
    document.documentElement.classList.add('dark')

    const wrapper = createWrapper()
    const homeRoot = wrapper.get('#top')
    const themeToggle = wrapper.get('button[title="切换浅色模式"]')
    const firstBenefitCard = wrapper.findAll('[data-test="benefit-tag"]')[0]
    const firstBenefitIconShell = wrapper.get('[data-test="benefit-icon-shell"]')
    const firstBenefitIcon = firstBenefitCard.get('[data-test="icon"]')
    const decorOverlay = wrapper.get('[data-test="home-decor-overlay"]')
    const leftFlow = wrapper.get('[data-test="hero-flow-left"]')
    const rightFlow = wrapper.get('[data-test="hero-flow-right"]')
    const coreGlow = wrapper.get('[data-test="hero-core-glow"]')

    expect(homeRoot.classes()).toContain('home-dark')
    expect(decorOverlay.classes()).toContain('home-decor-dark-tuned')
    expect(leftFlow.classes()).toContain('home-dark-hidden-layer')
    expect(rightFlow.classes()).toContain('home-dark-hidden-layer')
    expect(coreGlow.classes()).toContain('home-dark-keep-glow')
    expect(firstBenefitIconShell.classes()).toContain('bg-slate-900/90')
    expect(firstBenefitIconShell.classes()).toContain('ring-cyan-400/20')
    expect(firstBenefitIcon.attributes('data-icon-name')).toBe('chartBar')
    expect(leftFlow.exists()).toBe(true)
    expect(rightFlow.exists()).toBe(true)
    expect(firstBenefitCard.classes()).toContain('bg-slate-950/92')
    expect(themeToggle.exists()).toBe(true)

    document.documentElement.classList.remove('dark')
  })

  it('renders the new Chinese hero and navigation content on default home page', () => {
    const wrapper = createWrapper()
    const pricingAnchorLinks = wrapper.findAll('a[href="#pricing"]')

    expect(wrapper.text()).toContain('首页')
    expect(wrapper.text()).toContain('定价')
    expect(wrapper.text()).toContain('API文档')
    expect(wrapper.text()).toContain('技术社群')
    expect(wrapper.text()).toContain('立即使用')
    expect(wrapper.text()).toContain('稳定 · 安全 · 高效的 OpenAI 中转服务')
    expect(wrapper.text()).toContain('格品API')
    expect(wrapper.text()).toContain('让 AI 开发更简单')
    expect(wrapper.text()).toContain('企业级 OpenAI API 中转服务 | 国内极速响应 | 一站式接入')
    expect(wrapper.text()).toContain('立即开始接入')
    expect(wrapper.text()).toContain('查看文档')
    expect(pricingAnchorLinks.length).toBeGreaterThanOrEqual(2)
    expect(wrapper.find('[data-test="locale-switcher"]').exists()).toBe(true)
  })

  it('keeps auth-dependent console routes in the redesigned wrapper', () => {
    authStoreMock.isAuthenticated = true

    const wrapper = createWrapper()
    const consoleLinks = wrapper.findAll('a[href="/dashboard"]')

    expect(wrapper.text()).toContain('控制台')
    expect(wrapper.text()).toContain('进入控制台')
    expect(consoleLinks.length).toBeGreaterThan(0)
  })

  it('renders homepage copy from english locale messages', () => {
    Object.assign(messages, {
      'home.landing.nav.home': 'Home',
      'home.landing.nav.pricing': 'Pricing',
      'home.landing.nav.docs': 'Docs',
      'home.landing.nav.community': 'Community',
      'home.landing.console.login': 'Get Started',
      'home.landing.hero.badge': 'Reliable, secure, and fast OpenAI relay service',
      'home.landing.hero.titleSuffix': 'builds AI faster',
      'home.landing.hero.subtitle': 'Enterprise OpenAI API relay | Fast mainland access | One-stop integration',
      'home.landing.hero.primaryCtaGuest': 'Start Now',
      'home.landing.hero.secondaryCta': 'Read Docs',
      'home.landing.community.lead': 'Scan to join via QQ or WeChat',
      'home.landing.community.body': 'Get onboarding help, model updates, and fast troubleshooting.',
      'home.landing.community.qrLabels.qq': 'QQ Group',
      'home.landing.community.qrLabels.wechat': 'WeChat',
      'home.landing.community.qrAlt.qq': 'QQ group QR code',
      'home.landing.community.qrAlt.wechat': 'WeChat QR code'
    })

    const wrapper = createWrapper()

    expect(wrapper.text()).toContain('Home')
    expect(wrapper.text()).toContain('Pricing')
    expect(wrapper.text()).toContain('Docs')
    expect(wrapper.text()).toContain('Community')
    expect(wrapper.text()).toContain('Reliable, secure, and fast OpenAI relay service')
    expect(wrapper.text()).toContain('builds AI faster')
    expect(wrapper.text()).toContain('Enterprise OpenAI API relay | Fast mainland access | One-stop integration')
    expect(wrapper.text()).toContain('Start Now')
    expect(wrapper.text()).toContain('Read Docs')
    expect(wrapper.text()).toContain('Scan to join via QQ or WeChat')
    expect(wrapper.text()).toContain('QQ Group')
    expect(wrapper.text()).toContain('WeChat')
  })

  it('sends unauthenticated console entry through login with dashboard redirect', () => {
    const wrapper = createWrapper()
    const hrefs = wrapper.findAll('a').map((link) => link.attributes('href'))

    expect(hrefs.filter((href) => href === '/login?redirect=%2Fdashboard').length).toBeGreaterThan(0)
  })

  it('renders the requested landing page sections and commercial copy', () => {
    const wrapper = createWrapper()
    const pricingSection = wrapper.get('#pricing')
    const pricingCards = wrapper.findAll('[data-test="pricing-card"]')
    const pricingAnchorLinks = wrapper.findAll('a[href="#pricing"]')
    const benefitTags = wrapper.findAll('[data-test="benefit-tag"]')
    const overviewSection = wrapper.get('#overview')

    expect(wrapper.text()).toContain('快速、稳定、安全的 AI 接入体验')
    expect(wrapper.text()).toContain('Windows')
    expect(wrapper.text()).toContain('接入 OpenAI 前沿模型')
    expect(wrapper.text()).toContain('覆盖当前 API 可用的主流推理、轻量与图像场景。')
    expect(wrapper.text()).toContain('最新支持 GPT-5.5，适合复杂推理、编码与专业工作流。')
    expect(wrapper.text()).toContain('GPT-5.5')
    expect(wrapper.text()).toContain('GPT-5.4 mini')
    expect(wrapper.text()).toContain('GPT-5.4 nano')
    expect(wrapper.text()).toContain('GPT Image 2')
    expect(wrapper.text()).toContain('当前最新')
    expect(wrapper.text()).toContain('标准套餐')
    expect(wrapper.text()).toContain('专业套餐')
    expect(wrapper.text()).toContain('企业套餐')
    expect(wrapper.text()).toContain('¥19.9/月起')
    expect(wrapper.text()).toContain('¥135/月')
    expect(wrapper.text()).toContain('定制价格')
    expect(wrapper.text()).toContain('支持的操作系统与设备')
    expect(wrapper.text()).toContain('轻松完成模型接入')
    expect(wrapper.text()).toContain('加入开发者社区')
    expect(wrapper.text()).not.toContain('了解更多')
    expect(wrapper.text()).not.toContain('查看完整支持列表')
    expect(wrapper.text()).not.toContain('查看所有模型')
    expect(wrapper.text()).not.toContain('◉')
    expect(wrapper.text()).not.toContain('◆')
    expect(wrapper.text()).not.toContain('△')
    expect(wrapper.text()).not.toContain('↗')
    expect(wrapper.text()).not.toContain('☁')
    expect(wrapper.text()).not.toContain('◌')
    expect(wrapper.text()).not.toContain('社群交流')
    expect(wrapper.text()).not.toContain('问题解答')
    expect(wrapper.text()).not.toContain('更新同步')
    expect(pricingCards).toHaveLength(3)
    expect(pricingAnchorLinks.length).toBeGreaterThanOrEqual(2)
    expect(pricingCards.every((card) => card.attributes('href') === '/purchase')).toBe(true)
    expect(pricingSection.classes()).toContain('rounded-[30px]')
    expect(pricingSection.find('table').exists()).toBe(false)
    expect(pricingSection.get('[data-plan-highlight="true"]').text()).toContain('企业套餐')
    expect(benefitTags).toHaveLength(6)
    expect(overviewSection.find('table').exists()).toBe(false)
  })

  it('renders layered hero glass panels and flow-light background structure', () => {
    const wrapper = createWrapper()

    expect(wrapper.find('[data-test="hero-flow-left"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="hero-flow-right"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="hero-beam-left"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="hero-beam-right"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test="hero-panel"]').length).toBe(3)
    expect(wrapper.find('[data-test="hero-core"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="hero-plinth"]').exists()).toBe(true)
  })

  it('renders homepage documentation links as external links that open in a new window', () => {
    const wrapper = createWrapper()

    const docLinks = wrapper.findAll(`a[href="${TOKEN_DOC_URL}"]`)

    expect(docLinks.some((link) => link.text().includes('API文档'))).toBe(true)
    expect(docLinks.some((link) => link.text().includes('查看文档'))).toBe(true)
    expect(docLinks.every((link) => link.attributes('target') === '_blank')).toBe(true)
    expect(docLinks.every((link) => link.attributes('rel') === 'noopener noreferrer')).toBe(true)
  })

  it('renders the ICP record number as a MIIT link', () => {
    const wrapper = createWrapper()

    const icpLink = wrapper.get('a[href="https://beian.miit.gov.cn/"]')
    expect(icpLink.text()).toContain('蜀ICP备17044249号-1')
    expect(icpLink.attributes('target')).toBe('_blank')
    expect(icpLink.attributes('rel')).toBe('noopener noreferrer')
  })

  it('renders the redesigned footer sections and support entry points', () => {
    const wrapper = createWrapper()
    const pricingLinks = wrapper.findAll('a[href="#pricing"]')
    const qqImages = wrapper.findAll('img[alt="QQ群二维码"]')
    const wechatImages = wrapper.findAll('img[alt="微信二维码"]')
    const communityQr = qqImages[0]
    const copyrightRow = wrapper.get('footer .border-t')

    expect(wrapper.text()).toContain('产品')
    expect(wrapper.text()).toContain('开发者')
    expect(wrapper.text()).toContain('公司')
    expect(wrapper.text()).toContain('支持')
    expect(wrapper.text()).toContain('关注我们')
    expect(wrapper.text()).toContain('扫码加入 QQ 群或微信')
    expect(wrapper.text()).toContain('QQ 群')
    expect(wrapper.text()).toContain('微信')
    expect(wrapper.text()).toContain('获取接入支持、模型上新通知与问题答疑')
    expect(wrapper.text()).toContain('成都格品科技有限公司版权所有')
    expect(wrapper.text()).not.toContain('工单系统')
    expect(wrapper.text()).not.toContain('状态看板')
    expect(pricingLinks.some((link) => link.text().includes('定价'))).toBe(true)
    expect(qqImages.length).toBeGreaterThan(0)
    expect(wechatImages.length).toBeGreaterThan(0)
    expect(qqImages.every((image) => image.attributes('src') === '/qq.jpg')).toBe(true)
    expect(wechatImages.every((image) => image.attributes('src') === '/wechat.jpg')).toBe(true)
    expect(communityQr.classes()).toContain('max-w-[300px]')
    expect(copyrightRow.classes()).toContain('justify-center')
    expect(copyrightRow.classes()).toContain('text-center')
  })

  it('computes main and anchor offset from the fixed header height', () => {
    const originalGetBoundingClientRect = HTMLElement.prototype.getBoundingClientRect
    HTMLElement.prototype.getBoundingClientRect = vi.fn(function (this: HTMLElement) {
      if (this.tagName === 'HEADER') {
        return {
          width: 1200,
          height: 96,
          top: 0,
          left: 0,
          right: 1200,
          bottom: 96,
          x: 0,
          y: 0,
          toJSON: () => ({})
        } as DOMRect
      }

      return originalGetBoundingClientRect.call(this)
    })

    const wrapper = createWrapper()
    const main = wrapper.get('main')
    const overviewSection = wrapper.get('#overview')

    expect(main.attributes('style')).toContain('padding-top: 120px;')
    expect(overviewSection.attributes('style')).toContain('scroll-margin-top: 120px;')

    HTMLElement.prototype.getBoundingClientRect = originalGetBoundingClientRect
  })
})
