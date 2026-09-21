import { describe, expect, it } from 'vitest'
import { fetchArticleList, searchArticles } from '@/api/article'
import { fetchBanners, fetchFriendLinks, fetchFriendPromotions } from '@/api/ops'
import { fetchSiteConfig } from '@/api/site'
import { fetchCaptcha, fetchUserList } from '@/api/user'
import type { SiteConfig } from '@/api/types'
import { rawRequest, callWithRetry } from '../helpers'

describe('匿名接口 · HTTP 200 直连校验(原生 fetch)', () => {
  it('GET /api/heartbeat 返回 200', async () => {
    const res = await rawRequest('GET', '/heartbeat')
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
  })

  it('GET /api/site/site 返回 200 且带站点配置', async () => {
    const res = await rawRequest<SiteConfig>('GET', '/site/site')
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(res.body.data.project.title).toBeTruthy()
    expect(typeof res.body.data.login.usernamePassword).toBe('boolean')
  })

  it('GET /api/captcha 返回 200 且带 captchaID / base64 图片', async () => {
    const res = await rawRequest<{ captchaID: string; captcha: string }>('GET', '/captcha', {
      params: { target: '注册' },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(res.body.data.captchaID).toBeTruthy()
    expect(res.body.data.captcha).toMatch(/^data:image\/png;base64,/)
  })

  it('GET /api/article 列表返回 200 且为 { list, count } 结构', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/article', {
      params: { type: 'other', page: 1, limit: 5 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
    expect(typeof res.body.data.count).toBe('number')
  })

  it('GET /api/article/search 返回 200(ES 关闭时走数据库降级)', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/article/search', {
      params: { key: '', page: 1, limit: 5, type: 0 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
  })

  it('GET /api/article/search 非法排序类型返回非 200', async () => {
    const res = await rawRequest('GET', '/article/search', { params: { type: 99 } })
    expect(res.status).toBeGreaterThanOrEqual(400)
    expect(res.body.code).not.toBe(200)
  })

  it('GET /api/article/search 标签过滤生效', async (ctx) => {
    const published = await rawRequest<{ list: Array<{ tagList: string[] | null }> }>('GET', '/article', {
      params: { type: 'other', page: 1, limit: 20 },
    })
    const tag = published.body.data.list.flatMap((item) => item.tagList || [])[0]
    if (!tag) {
      ctx.skip()
      return
    }

    const res = await rawRequest<{ list: Array<{ tagList: string[] | null }> }>('GET', '/article/search', {
      params: { tag, page: 1, limit: 20 },
    })
    expect(res.body.code).toBe(200)
    expect(res.body.data.list.length).toBeGreaterThan(0)
    expect(res.body.data.list.every((item) => (item.tagList || []).includes(tag))).toBe(true)
  })

  it('GET /api/banner /api/friendLink /api/friendPromotion /api/user/list 均返回 200', async () => {
    for (const path of ['/banner', '/friendLink', '/friendPromotion', '/user/list']) {
      const res = await rawRequest<{ list: unknown[] }>('GET', path, { params: { page: 1, limit: 5 } })
      expect(res.status, `${path} 状态码`).toBe(200)
      expect(res.body.code, `${path} 业务码`).toBe(200)
      expect(Array.isArray(res.body.data.list), `${path} list`).toBe(true)
    }
  })
})

describe('前端 API 封装层(axios)调用后端', () => {
  it('fetchSiteConfig() 返回站点配置', async () => {
    const config = await callWithRetry(() => fetchSiteConfig())
    expect(config.project.title).toBeTruthy()
    expect(config.siteInfo.Mode).toBeGreaterThan(0)
  })

  it('fetchCaptcha() 返回验证码', async () => {
    const captcha = await callWithRetry(() => fetchCaptcha('注册'))
    expect(captcha.captchaID).toBeTruthy()
    expect(captcha.captcha.length).toBeGreaterThan(32)
  })

  it('fetchArticleList({ type: "other" }) 返回列表与总数', async () => {
    const data = await callWithRetry(() => fetchArticleList({ type: 'other', page: 1, limit: 5 }))
    expect(Array.isArray(data.list)).toBe(true)
    expect(typeof data.count).toBe('number')
  })

  it('searchArticles({ key }) 关键词确实参与过滤(回归:曾被 usePagination 吞掉)', async (ctx) => {
    const published = await fetchArticleList({ type: 'other', page: 1, limit: 1 })
    const first = published.list[0]
    if (!first) {
      ctx.skip()
      return
    }

    const keyword = first.title.replace(/<[^>]+>/g, '').slice(0, 2)
    const hit = await callWithRetry(() => searchArticles({ key: keyword, page: 1, limit: 5 }))
    expect(hit.count, `关键词「${keyword}」应有命中`).toBeGreaterThan(0)

    const miss = await callWithRetry(() =>
      searchArticles({ key: 'zzz-绝对不存在的关键词-zzz', page: 1, limit: 5 }),
    )
    expect(miss.count).toBe(0)
  })

  it('fetchBanners() / fetchFriendLinks() / fetchFriendPromotions() / fetchUserList() 正常返回', async () => {
    const [banners, links, promotions, users] = await callWithRetry(() =>
      Promise.all([
        fetchBanners({ page: 1, limit: 5 }),
        fetchFriendLinks({ page: 1, limit: 5 }),
        fetchFriendPromotions({ page: 1, limit: 5 }),
        fetchUserList({ page: 1, limit: 5 }),
      ]),
    )
    expect(Array.isArray(banners.list)).toBe(true)
    expect(Array.isArray(links.list)).toBe(true)
    expect(Array.isArray(promotions.list)).toBe(true)
    expect(Array.isArray(users.list)).toBe(true)
  })
})
