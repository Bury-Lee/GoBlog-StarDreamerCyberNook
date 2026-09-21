import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { fetchCollectFolders, fetchCategories } from '@/api/article'
import { askAi, fetchChatSessions } from '@/api/chat'
import { fetchMessageConf } from '@/api/message'
import { fetchUserDetail } from '@/api/user'
import type { MessageConf } from '@/api/types'
import {
  callWithRetry,
  loginAsTestUser,
  makeExpiredJwt,
  rawRequest,
  resetAuth,
  setTestTokens,
  testCredentials,
  type TestSession,
} from '../helpers'

const describeAuth = testCredentials() ? describe : describe.skip

describeAuth('登录后接口(需要 token)', () => {
  let session: TestSession

  beforeAll(async () => {
    session = await loginAsTestUser()
  })

  afterAll(() => {
    resetAuth()
  })

  it('POST /api/user/login 返回双令牌', () => {
    expect(session.token.split('.')).toHaveLength(3)
    expect(session.refreshToken).toBeTruthy()
  })

  it('GET /api/user/detail 返回 200 且 id 与登录用户一致', async () => {
    const res = await rawRequest<{ id: number }>('GET', '/user/detail', { token: session.token })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(res.body.data.id).toBe(session.userID)
  })

  it('GET /api/user/detail 未携带 token 时被拒绝(后端鉴权失败返回 FailCode=201)', async () => {
    const res = await rawRequest('GET', '/user/detail')
    expect(res.body.code).not.toBe(200)
    expect(res.status).toBe(201)
  })

  it('POST /api/user/token 使用 refreshToken 换取新 AccessToken', async () => {
    const res = await rawRequest<string>('POST', '/user/token', { refreshToken: session.refreshToken })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(typeof res.body.data).toBe('string')
    expect(res.body.data.split('.')).toHaveLength(3)
  })

  it('GET /api/msg/check 返回未读消息映射', async () => {
    const res = await rawRequest<Record<string, number>>('GET', '/msg/check', { token: session.token })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(typeof res.body.data).toBe('object')
  })

  it('GET /api/msg/conf 返回通知配置', async () => {
    const res = await rawRequest<MessageConf>('GET', '/msg/conf', { token: session.token })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(typeof res.body.data.openCommentMessage).toBe('boolean')
  })

  it('GET /api/msg?type=1 返回消息列表', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/msg', {
      token: session.token,
      params: { type: 1, page: 1, limit: 5 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
  })

  it('GET /api/user/loginlog 返回登录日志', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/user/loginlog', {
      token: session.token,
      params: { page: 1, limit: 5 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(res.body.data.count).toBeGreaterThan(0)
  })

  it('GET /api/article?type=self 返回自己的文章', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/article', {
      token: session.token,
      params: { type: 'self', page: 1, limit: 5 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
  })

  it('GET /api/article/category?type=self 返回分类列表', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/article/category', {
      token: session.token,
      params: { type: 'self', page: 1, limit: 10 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
  })

  it('GET /api/chat/session 返回会话列表', async () => {
    const res = await rawRequest<{ list: unknown[]; count: number }>('GET', '/chat/session', {
      token: session.token,
      params: { page: 1, limit: 5 },
    })
    expect(res.status).toBe(200)
    expect(res.body.code).toBe(200)
    expect(Array.isArray(res.body.data.list)).toBe(true)
  })

  it('封装层调用: 分类 / 收藏夹 / 消息配置 / 会话列表', async () => {
    const [categories, folders, conf, sessions] = await callWithRetry(() =>
      Promise.all([
        fetchCategories({ type: 'self', page: 1, limit: 10 }),
        fetchCollectFolders({ id: session.userID, page: 1, limit: 10 }),
        fetchMessageConf(),
        fetchChatSessions({ page: 1, limit: 5 }),
      ]),
    )
    expect(Array.isArray(categories.list)).toBe(true)
    expect(Array.isArray(folders.list)).toBe(true)
    expect(typeof conf.openPrivateMessage).toBe('boolean')
    expect(Array.isArray(sessions.list)).toBe(true)
  })

  it('封装层自动携带 token(登录后调用 fetchUserDetail)', async () => {
    const detail = await callWithRetry(() => fetchUserDetail())
    expect(detail.id).toBe(session.userID)
  })

  it('封装层 refreshToken() 能刷新 AccessToken', async () => {
    const { refreshToken: refreshAccessToken } = await import('@/api/user')
    const token = await callWithRetry(() => refreshAccessToken(session.refreshToken))
    expect(typeof token).toBe('string')
    expect(token.split('.')).toHaveLength(3)
  })

  it('GET /api/msg/check 未登录时被拒绝(后端返回 201 而非 401)', async () => {
    const res = await rawRequest('GET', '/msg/check')
    expect(res.body.code).not.toBe(200)
    expect(res.status).toBe(201)
  })

  it('AccessToken 过期时,请求拦截器会自动刷新并正常完成请求', async () => {
    const expired = makeExpiredJwt({ ID: session.userID, name: 'admin', role: 1 })
    setTestTokens(expired, session.refreshToken)

    const detail = await callWithRetry(() => fetchUserDetail())
    expect(detail.id).toBe(session.userID)

    const { getAccessToken } = await import('@/utils/storage')
    const current = getAccessToken()
    expect(current).not.toBe(expired)
    expect(current.split('.')).toHaveLength(3)
  })

  it.runIf(process.env.TEST_AI === '1')(
    'POST /api/chat AI 助手返回 200(需要 LM Studio 已启动)',
    async () => {
      const res = await rawRequest<{ success: boolean; content: string }>('POST', '/chat', {
        token: session.token,
        body: { messages: [], user_input: '用一句话介绍你自己' },
      })
      expect(res.status).toBe(200)
      expect(res.body.code).toBe(200)
      expect(res.body.data.content.length).toBeGreaterThan(0)
    },
    90000,
  )

  it.runIf(process.env.TEST_AI === '1')(
    '封装层 askAi() 返回内容(需要 LM Studio 已启动)',
    async () => {
      const result = await askAi({ messages: [], user_input: '回复:测试通过' })
      expect(result.success).toBe(true)
      expect(result.content.length).toBeGreaterThan(0)
    },
    90000,
  )
})
