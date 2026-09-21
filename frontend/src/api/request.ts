import axios, {
  type AxiosError,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import { ElMessage } from 'element-plus'
import { clearAuth, getAccessToken, getRefreshToken, setAccessToken } from '@/utils/storage'
import { isTokenExpired, isTokenExpiring } from '@/utils/jwt'
import type { ApiResult } from './types'

export const API_BASE = import.meta.env.VITE_API_BASE || '/api'

export interface SdRequestConfig extends AxiosRequestConfig {
  silent?: boolean
}

export class ApiError extends Error {
  code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

type RetriableConfig = InternalAxiosRequestConfig & { _retry?: boolean; silent?: boolean }

const instance = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
})

instance.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  const token = await resolveAccessToken()
  if (token) {
    config.headers.set('token', token)
  }
  return config
})

let refreshPromise: Promise<string> | null = null

async function requestNewAccessToken(): Promise<string> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) throw new ApiError(401, '登录状态已失效')
  const response = await axios.post<ApiResult<string>>(
    `${API_BASE}/user/token`,
    null,
    { headers: { refreshToken }, timeout: 15000 },
  )
  const payload = response.data
  if (!payload || payload.code !== 200 || !payload.data) {
    throw new ApiError(401, payload?.message || '登录状态已失效')
  }
  setAccessToken(payload.data)
  return payload.data
}

function refreshAccessTokenOnce(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = requestNewAccessToken().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

async function resolveAccessToken(): Promise<string> {
  const token = getAccessToken()
  if (!token) return ''
  if (!isTokenExpiring(token) || !getRefreshToken()) return token
  try {
    return await refreshAccessTokenOnce()
  } catch {
    return token
  }
}

async function redirectToLogin(): Promise<void> {
  const { default: router } = await import('@/router')
  const current = router.currentRoute.value
  if (current.name === 'login' || current.name === 'register') return
  router.push({ name: 'login', query: { redirect: current.fullPath } })
}

const AUTH_ERROR_CODES = [201, 401, 422]

interface RetryOutcome {
  ok: boolean
  data?: unknown
}

async function tryRefreshAndRetry(config: RetriableConfig | undefined): Promise<RetryOutcome | null> {
  if (!config || config._retry) return null
  if (!getRefreshToken()) return null
  const token = getAccessToken()
  if (!token || !isTokenExpired(token)) return null
  config._retry = true
  try {
    const fresh = await refreshAccessTokenOnce()
    config.headers.set('token', fresh)
    const data = await instance.request(config)
    return { ok: true, data }
  } catch {
    return { ok: false }
  }
}

async function handleUnauthorized(config: RetriableConfig | undefined, message?: string): Promise<unknown> {
  const outcome = await tryRefreshAndRetry(config)
  if (outcome?.ok) return outcome.data
  clearAuth()
  void redirectToLogin()
  throw new ApiError(401, message || '登录状态已过期,请重新登录')
}

instance.interceptors.response.use(
  async (response) => {
    const payload = response.data as ApiResult<unknown> | string | null
    if (payload === null || payload === undefined || payload === '') {
      return {} as never
    }
    if (typeof payload === 'string') {
      return payload as never
    }
    if (payload.code === 200) {
      return payload.data as never
    }

    const config = response.config as RetriableConfig
    if (AUTH_ERROR_CODES.includes(payload.code)) {
      const outcome = await tryRefreshAndRetry(config)
      if (outcome?.ok) return outcome.data as never
      if (outcome && !outcome.ok) {
        clearAuth()
        void redirectToLogin()
        throw new ApiError(payload.code, payload.message || '登录状态已过期,请重新登录')
      }
    }

    if (!config.silent) {
      ElMessage.error(payload.message || '请求失败')
    }
    return Promise.reject(new ApiError(payload.code, payload.message || '请求失败'))
  },
  async (error: AxiosError<ApiResult<unknown>>) => {
    const config = error.config as RetriableConfig | undefined
    const status = error.response?.status
    const payload = error.response?.data
    const serverMessage = payload && typeof payload === 'object' ? payload.message : ''

    if (status && AUTH_ERROR_CODES.includes(status)) {
      const outcome = await tryRefreshAndRetry(config)
      if (outcome?.ok) return outcome.data as never
      if (outcome && !outcome.ok) {
        clearAuth()
        void redirectToLogin()
        throw new ApiError(401, '登录状态已过期,请重新登录')
      }
    }

    let message = serverMessage
    if (!message) {
      if (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT') {
        message = '请求超时,请稍后重试'
      } else if (status === 404) {
        message = '接口不存在或服务未启用'
      } else if (status === 429) {
        message = '请求过于频繁,请稍后再试'
      } else if (status && status >= 500) {
        message = '服务器开小差了,请稍后重试'
      } else if (!error.response) {
        message = '无法连接后端服务,请确认服务已启动'
      } else {
        message = error.message || '请求失败'
      }
    }

    if (!config?.silent) {
      ElMessage.error(message)
    }
    return Promise.reject(new ApiError(status ?? 0, message))
  },
)

export const http = {
  get<T>(url: string, params?: Record<string, unknown>, options?: SdRequestConfig): Promise<T> {
    return instance.get(url, { params, ...options }) as unknown as Promise<T>
  },
  post<T>(url: string, data?: unknown, options?: SdRequestConfig): Promise<T> {
    return instance.post(url, data, options) as unknown as Promise<T>
  },
  put<T>(url: string, data?: unknown, options?: SdRequestConfig): Promise<T> {
    return instance.put(url, data, options) as unknown as Promise<T>
  },
  delete<T>(url: string, data?: unknown, options?: SdRequestConfig): Promise<T> {
    return instance.delete(url, { data, ...options }) as unknown as Promise<T>
  },
  upload<T>(url: string, form: FormData, options?: SdRequestConfig): Promise<T> {
    return instance.post(url, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
      ...options,
    }) as unknown as Promise<T>
  },
}

export function imageUrl(id: number | string): string {
  return `${API_BASE}/image?id=${id}`
}

export function resolveAssetUrl(path: string): string {
  if (!path) return ''
  if (/^(https?:)?\/\//.test(path) || path.startsWith('data:')) return path
  if (path.startsWith('/api') || path.startsWith('/web')) return path
  if (path.startsWith('/')) return `/web${path}`
  return `/web/${path}`
}
