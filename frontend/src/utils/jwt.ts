export interface JwtPayload {
  exp?: number
  iat?: number
  ID?: number
  name?: string
  role?: number
  [key: string]: unknown
}

export function decodeJwtPayload(token: string): JwtPayload | null {
  if (!token) return null
  const parts = token.split('.')
  if (parts.length !== 3) return null
  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
    const binary = atob(padded)
    const json = decodeURIComponent(
      binary
        .split('')
        .map((char) => `%${char.charCodeAt(0).toString(16).padStart(2, '0')}`)
        .join(''),
    )
    return JSON.parse(json) as JwtPayload
  } catch {
    return null
  }
}

export function getTokenExpiry(token: string): number | null {
  const payload = decodeJwtPayload(token)
  if (!payload?.exp) return null
  return payload.exp * 1000
}

export function isTokenExpiring(token: string, thresholdSeconds = 120): boolean {
  const expiry = getTokenExpiry(token)
  if (expiry === null) return false
  return Date.now() + thresholdSeconds * 1000 >= expiry
}

export function isTokenExpired(token: string, skewSeconds = 5): boolean {
  const expiry = getTokenExpiry(token)
  if (expiry === null) return false
  return Date.now() >= expiry - skewSeconds * 1000
}
