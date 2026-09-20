import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchSiteConfig } from '@/api/site'
import type { SiteConfig, SiteIndexRightItem } from '@/api/types'
import { resolveAssetUrl } from '@/api/request'

const DEFAULT_SITE: SiteConfig = {
  siteInfo: { title: 'StarDreamer 赛博空间', Logo: '', Beian: '', Mode: 1 },
  project: { title: 'StarDreamer', icon: '', webPath: '' },
  seo: { keywords: '', description: '' },
  about: { Version: '', siteDate: '', qq: '', wechat: '', biliBili: '', gitHub: '' },
  indexRight: { list: [] },
  article: { enableExamination: true },
  login: { QQLogin: false, usernamePassword: true, emailLogin: true, captcha: false },
}

export const useSiteStore = defineStore('site', () => {
  const config = ref<SiteConfig>({ ...DEFAULT_SITE })
  const loaded = ref(false)
  const loading = ref(false)

  const title = computed(() => config.value.project?.title || config.value.siteInfo?.title || 'StarDreamer')
  const siteTitle = computed(() => config.value.siteInfo?.title || title.value)
  const logo = computed(() => resolveAssetUrl(config.value.siteInfo?.Logo || ''))
  const beian = computed(() => config.value.siteInfo?.Beian || '')
  const mode = computed(() => config.value.siteInfo?.Mode ?? 1)
  const seo = computed(() => config.value.seo)
  const about = computed(() => config.value.about)
  const loginOptions = computed(() => config.value.login)
  const captchaEnabled = computed(() => Boolean(config.value.login?.captcha))
  const reviewEnabled = computed(() => Boolean(config.value.article?.enableExamination))
  const indexRightList = computed<SiteIndexRightItem[]>(() => config.value.indexRight?.list ?? [])
  const isCommunityMode = computed(() => mode.value === 2)

  function applyDocumentMeta(): void {
    const project = config.value.project
    const seoInfo = config.value.seo
    document.title = project?.title || config.value.siteInfo?.title || 'StarDreamer'
    const setMeta = (name: string, content: string) => {
      if (!content) return
      let node = document.querySelector(`meta[name="${name}"]`)
      if (!node) {
        node = document.createElement('meta')
        node.setAttribute('name', name)
        document.head.appendChild(node)
      }
      node.setAttribute('content', content)
    }
    setMeta('keywords', seoInfo?.keywords || '')
    setMeta('description', seoInfo?.description || '')
    if (project?.icon) {
      const icon = resolveAssetUrl(project.icon)
      let link = document.querySelector('link[rel="icon"]') as HTMLLinkElement | null
      if (!link) {
        link = document.createElement('link')
        link.rel = 'icon'
        document.head.appendChild(link)
      }
      link.href = icon
    }
  }

  async function loadSite(force = false): Promise<SiteConfig> {
    if (loaded.value && !force) return config.value
    loading.value = true
    try {
      const data = await fetchSiteConfig()
      if (data) {
        config.value = {
          ...DEFAULT_SITE,
          ...data,
          siteInfo: { ...DEFAULT_SITE.siteInfo, ...data.siteInfo },
          project: { ...DEFAULT_SITE.project, ...data.project },
          seo: { ...DEFAULT_SITE.seo, ...data.seo },
          about: { ...DEFAULT_SITE.about, ...data.about },
          indexRight: { list: data.indexRight?.list ?? [] },
          article: { ...DEFAULT_SITE.article, ...data.article },
          login: { ...DEFAULT_SITE.login, ...data.login },
        }
      }
      loaded.value = true
      applyDocumentMeta()
      return config.value
    } catch {
      loaded.value = true
      return config.value
    } finally {
      loading.value = false
    }
  }

  return {
    config,
    loaded,
    loading,
    title,
    siteTitle,
    logo,
    beian,
    mode,
    seo,
    about,
    loginOptions,
    captchaEnabled,
    reviewEnabled,
    indexRightList,
    isCommunityMode,
    loadSite,
    applyDocumentMeta,
  }
})
