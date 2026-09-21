import { API_BASE } from '@/api/request'
import { setTokens, clearAuth } from '@/utils/storage'

const FALLBACK_BASE = 'http://127.0.0.1:8080/api'

export interface RawResponse<T = unknown> {
  status: number
  body: {
    code: number
    data: T
    message: string
  }
}

export function resolveBase(): string {
  return /^https?:\/\//.test(API_BASE) ? API_BASE.replace(/\/$/, '') : FALLBACK_BASE
}

interface RawRequestOptions {
  params?: Record<string, unknown>
  body?: unknown
  token?: string
  refreshToken?: string
}

const RATE_LIMIT_MARK = '请求过于频繁'
const MAX_ATTEMPTS = 4
const RETRY_DELAY_MS = 20000

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}

export function isRateLimitedMessage(message: string | undefined): boolean {
  return Boolean(message && message.includes(RATE_LIMIT_MARK))
}

export async function callWithRetry<T>(task: () => Promise<T>): Promise<T> {
  let lastError: unknown
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt += 1) {
    try {
      return await task()
    } catch (error) {
      lastError = error
      const message = error instanceof Error ? error.message : ''
      if (attempt < MAX_ATTEMPTS && isRateLimitedMessage(message)) {
        await sleep(RETRY_DELAY_MS)
        continue
      }
      throw error
    }
  }
  throw lastError
}

async function sendRawRequest<T>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  path: string,
  options: RawRequestOptions,
): Promise<RawResponse<T>> {
  const url = new URL(`${resolveBase()}${path}`)
  Object.entries(options.params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null) url.searchParams.set(key, String(value))
  })

  const headers: Record<string, string> = {}
  if (options.body !== undefined) headers['Content-Type'] = 'application/json'
  if (options.token) headers.token = options.token
  if (options.refreshToken) headers.refreshToken = options.refreshToken

  const response = await fetch(url, {
    method,
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  })

  let body: RawResponse<T>['body'] = { code: 0, data: {} as T, message: '' }
  try {
    body = (await response.json()) as RawResponse<T>['body']
  } catch {
    body = { code: response.status, data: {} as T, message: '非 JSON 响应' }
  }
  return { status: response.status, body }
}

export async function rawRequest<T = unknown>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  path: string,
  options: RawRequestOptions = {},
): Promise<RawResponse<T>> {
  let result: RawResponse<T> | null = null
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt += 1) {
    result = await sendRawRequest<T>(method, path, options)
    if (attempt < MAX_ATTEMPTS && isRateLimitedMessage(result.body.message)) {
      await sleep(RETRY_DELAY_MS)
      continue
    }
    break
  }
  return result as RawResponse<T>
}

export interface TestSession {
  token: string
  refreshToken: string
  userID: number
  nickname: string
}

export function testCredentials(): { username: string; password: string } | null {
  const username = process.env.TEST_USERNAME
  const password = process.env.TEST_PASSWORD
  if (!username || !password) return null
  return { username, password }
}

export async function loginAsTestUser(): Promise<TestSession> {
  const credentials = testCredentials()
  if (!credentials) throw new Error('未配置 TEST_USERNAME / TEST_PASSWORD')

  const { body } = await rawRequest<{ AccessToken: string; RefreshToken: string }>('POST', '/user/login', {
    body: { type: '用户名', val: credentials.username, pwd: credentials.password },
  })
  if (body.code !== 200 || !body.data?.AccessToken) {
    throw new Error(`登录失败: code=${body.code} message=${body.message}`)
  }

  setTokens(body.data.AccessToken, body.data.RefreshToken)

  const detail = await rawRequest<{ id: number; nickname: string }>('GET', '/user/detail', {
    token: body.data.AccessToken,
  })

  return {
    token: body.data.AccessToken,
    refreshToken: body.data.RefreshToken,
    userID: detail.body.data?.id ?? 0,
    nickname: detail.body.data?.nickname ?? '',
  }
}

export function resetAuth(): void {
  clearAuth()
}

export function setTestTokens(accessToken: string, refreshToken: string): void {
  setTokens(accessToken, refreshToken)
}

function base64Url(value: Record<string, unknown>): string {
  return btoa(JSON.stringify(value)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

export function makeJwt(payload: Record<string, unknown>): string {
  return `${base64Url({ alg: 'HS256', typ: 'JWT' })}.${base64Url(payload)}.test-signature`
}

export function makeExpiredJwt(payload: Record<string, unknown> = {}): string {
  return makeJwt({ ...payload, exp: Math.floor(Date.now() / 1000) - 60 })
}

export function makeFreshJwt(payload: Record<string, unknown> = {}, seconds = 3600): string {
  return makeJwt({ ...payload, exp: Math.floor(Date.now() / 1000) + seconds })
}
