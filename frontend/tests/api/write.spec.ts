import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import type { ArticleListResponse, ArticleModel, CollectModel, CommentModel } from '@/api/types'
import { loginAsTestUser, rawRequest, resetAuth, type TestSession } from '../helpers'

const describeWrite = process.env.TEST_WRITE === '1' ? describe : describe.skip

interface ListPayload<T> {
  list: T[]
  count: number
}

describeWrite('写操作接口(需 TEST_WRITE=1,测试数据会自动清理)', () => {
  let session: TestSession

  beforeAll(async () => {
    session = await loginAsTestUser()
  })

  afterAll(() => {
    resetAuth()
  })

  it(
    '文章: 创建草稿 -> 我的列表可见 -> 删除',
    async () => {
      const title = `[api-test] 草稿 ${Date.now()}`
      const create = await rawRequest('POST', '/article', {
        token: session.token,
        body: {
          title,
          abstract: '由前端自动化测试创建,会自动删除',
          content: '<h2>api-test</h2><p>hello</p>',
          tagList: ['api-test'],
          openComment: true,
          status: 0,
        },
      })
      expect(create.status, `创建文章失败: ${create.body.message}`).toBe(200)
      expect(create.body.code, `创建文章失败: ${create.body.message}`).toBe(200)

      const list = await rawRequest<ListPayload<ArticleListResponse>>('GET', '/article', {
        token: session.token,
        params: { type: 'self', page: 1, limit: 40 },
      })
      const target = list.body.data.list.find((item) => item.title === title)
      expect(target, '新建文章应出现在自己的列表中').toBeTruthy()

      const remove = await rawRequest('DELETE', '/article', {
        token: session.token,
        body: { IDList: [target!.id] },
      })
      expect(remove.body.code).toBe(200)

      const after = await rawRequest<ListPayload<ArticleListResponse>>('GET', '/article', {
        token: session.token,
        params: { type: 'self', page: 1, limit: 40 },
      })
      expect(after.body.data.list.some((item) => item.id === target!.id)).toBe(false)
    },
    60000,
  )

  it(
    '评论: 对已发布文章发表评论 -> 删除',
    async (ctx) => {
      const published = await rawRequest<ListPayload<ArticleModel>>('GET', '/article', {
        params: { type: 'other', page: 1, limit: 1 },
      })
      const article = published.body.data.list[0]
      if (!article) {
        ctx.skip()
        return
      }

      const create = await rawRequest('POST', '/comment', {
        token: session.token,
        body: { articleID: article.id, content: `[api-test] 评论 ${Date.now()}`, parentID: 0 },
      })
      expect(create.status).toBe(200)
      expect(create.body.code).toBe(200)

      const comments = await rawRequest<ListPayload<ArticleModel & { userID: number }>>('GET', '/comment', {
        params: { articleID: article.id, page: 1, limit: 40 },
      })
      const mine = comments.body.data.list.find((item) => item.userID === session.userID)
      expect(mine, '评论应出现在一级评论列表中').toBeTruthy()

      const remove = await rawRequest('DELETE', `/comment/${mine!.id}`, { token: session.token })
      expect(remove.body.code).toBe(200)
    },
    60000,
  )

  it(
    '收藏夹: 创建 -> 列表可见 -> 删除',
    async () => {
      const title = `[api-test] 收藏夹 ${Date.now()}`
      const create = await rawRequest('POST', '/article/collect/folder', {
        token: session.token,
        body: { title, abstract: '由前端自动化测试创建' },
      })
      expect(create.status).toBe(200)
      expect(create.body.code).toBe(200)

      const folders = await rawRequest<ListPayload<CollectModel>>('GET', '/article/collect/folder', {
        token: session.token,
        params: { id: session.userID, page: 1, limit: 40 },
      })
      const target = folders.body.data.list.find((item) => item.title === title)
      expect(target, '新建收藏夹应可见').toBeTruthy()

      const remove = await rawRequest('DELETE', '/article/collect/folder', {
        token: session.token,
        body: { IDList: [target!.id] },
      })
      expect(remove.body.code).toBe(200)
    },
    60000,
  )

  it(
    '评论: 点赞 toggle 生效,且列表立即可见点赞数(回归: 原点赞增量只写Redis,列表不叠加)',
    async (ctx) => {
      const published = await rawRequest<ListPayload<ArticleModel>>('GET', '/article', {
        params: { type: 'other', page: 1, limit: 1 },
      })
      const article = published.body.data.list[0]
      if (!article) {
        ctx.skip()
        return
      }

      const create = await rawRequest('POST', '/comment', {
        token: session.token,
        body: { articleID: article.id, content: `[api-test] 点赞用例 ${Date.now()}`, parentID: 0 },
      })
      expect(create.body.code).toBe(200)

      const list = await rawRequest<ListPayload<CommentModel>>('GET', '/comment', {
        params: { articleID: article.id, page: 1, limit: 40 },
      })
      const target = list.body.data.list.find((item) => item.userID === session.userID)
      expect(target, '评论应出现在列表中').toBeTruthy()
      const before = target!.diggCount

      const digg = await rawRequest<{ digged: boolean; diggCount: number }>(
        'POST',
        `/comment/digg/${target!.id}`,
        { token: session.token },
      )
      expect(digg.body.code).toBe(200)
      expect(digg.body.data.digged).toBe(true)
      expect(digg.body.data.diggCount).toBe(before + 1)

      const afterDigg = await rawRequest<ListPayload<CommentModel>>('GET', '/comment', {
        params: { articleID: article.id, page: 1, limit: 40 },
      })
      const liked = afterDigg.body.data.list.find((item) => item.id === target!.id)
      expect(liked?.diggCount, '列表应叠加Redis点赞增量').toBe(before + 1)

      const undigg = await rawRequest<{ digged: boolean; diggCount: number }>(
        'POST',
        `/comment/digg/${target!.id}`,
        { token: session.token },
      )
      expect(undigg.body.data.digged).toBe(false)
      expect(undigg.body.data.diggCount).toBe(before)

      const remove = await rawRequest('DELETE', `/comment/${target!.id}`, { token: session.token })
      expect(remove.body.code).toBe(200)
    },
    60000,
  )

  it(
    '评论: 删除一级评论时,自身与子评论一起删除(回归: 原来漏删一级评论本身)',
    async (ctx) => {
      const published = await rawRequest<ListPayload<ArticleModel>>('GET', '/article', {
        params: { type: 'other', page: 1, limit: 1 },
      })
      const article = published.body.data.list[0]
      if (!article) {
        ctx.skip()
        return
      }

      const createRoot = await rawRequest('POST', '/comment', {
        token: session.token,
        body: { articleID: article.id, content: `[api-test] 根评论 ${Date.now()}`, parentID: 0 },
      })
      expect(createRoot.body.code).toBe(200)

      const listAfterCreate = await rawRequest<ListPayload<CommentModel>>('GET', '/comment', {
        params: { articleID: article.id, page: 1, limit: 40 },
      })
      const root = listAfterCreate.body.data.list.find((item) => item.userID === session.userID)
      expect(root, '根评论应创建成功').toBeTruthy()

      const createChild = await rawRequest('POST', '/comment', {
        token: session.token,
        body: { articleID: article.id, content: `[api-test] 子评论 ${Date.now()}`, parentID: root!.id },
      })
      expect(createChild.body.code).toBe(200)

      const children = await rawRequest<ListPayload<CommentModel>>('GET', '/commentChild', {
        params: { root: root!.id, page: 1, limit: 20 },
      })
      expect(children.body.data.count).toBeGreaterThan(0)

      const remove = await rawRequest('DELETE', `/comment/${root!.id}`, { token: session.token })
      expect(remove.body.code, `删除失败: ${remove.body.message}`).toBe(200)

      const afterDelete = await rawRequest<ListPayload<CommentModel>>('GET', '/comment', {
        params: { articleID: article.id, page: 1, limit: 40 },
      })
      expect(
        afterDelete.body.data.list.some((item) => item.id === root!.id),
        '一级评论本身应已被删除',
      ).toBe(false)

      const childrenAfterDelete = await rawRequest('GET', '/commentChild', {
        params: { root: root!.id, page: 1, limit: 20 },
      })
      expect(childrenAfterDelete.body.code, '根评论已删除,子评论接口应返回错误').not.toBe(200)
    },
    60000,
  )

  it('非法写操作被拒绝: 未登录发文被拒绝(后端返回 201)', async () => {
    const res = await rawRequest('POST', '/article', {
      body: { title: 'x', content: '<p>x</p>', status: 0 },
    })
    expect(res.body.code).not.toBe(200)
    expect(res.status).toBe(201)
  })

  it('非法写操作被拒绝: 空标题返回非 200', async () => {
    const res = await rawRequest('POST', '/article', {
      token: session.token,
      body: { title: '', content: '<p>x</p>', status: 0 },
    })
    expect(res.body.code).not.toBe(200)
  })
})
