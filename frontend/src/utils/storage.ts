const ACCESS_TOKEN_KEY = 'goblog.access_token'
const REFRESH_TOKEN_KEY = 'goblog.refresh_token'
const PROFILE_KEY = 'goblog.profile'

export function getAccessToken(): string {
  return localStorage.getItem(ACCESS_TOKEN_KEY) || ''
}

export function getRefreshToken(): string {
  return localStorage.getItem(REFRESH_TOKEN_KEY) || ''
}

export function setTokens(accessToken: string, refreshToken?: string): void {
  if (accessToken) localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  if (refreshToken) localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function setAccessToken(accessToken: string): void {
  if (accessToken) localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function readProfile<T>(): T | null {
  const raw = localStorage.getItem(PROFILE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function writeProfile<T>(profile: T | null): void {
  if (profile === null || profile === undefined) {
    localStorage.removeItem(PROFILE_KEY)
    return
  }
  localStorage.setItem(PROFILE_KEY, JSON.stringify(profile))
}

export function clearAuth(): void {
  clearTokens()
  writeProfile(null)
}
