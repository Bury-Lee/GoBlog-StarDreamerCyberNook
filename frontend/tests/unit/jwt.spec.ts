import { describe, expect, it } from 'vitest'
import { decodeJwtPayload, getTokenExpiry, isTokenExpired, isTokenExpiring } from '@/utils/jwt'
import { makeExpiredJwt, makeFreshJwt, makeJwt } from '../helpers'

describe('JWT 解析工具', () => {
  it('解析 payload 中的业务字段', () => {
    const token = makeJwt({ ID: 1, name: 'admin', role: 1, exp: 1893456000 })
    const payload = decodeJwtPayload(token)
    expect(payload?.ID).toBe(1)
    expect(payload?.name).toBe('admin')
    expect(payload?.role).toBe(1)
  })

  it('getTokenExpiry 返回毫秒时间戳', () => {
    const token = makeJwt({ exp: 1893456000 })
    expect(getTokenExpiry(token)).toBe(1893456000 * 1000)
  })

  it('非法 token 返回 null', () => {
    expect(decodeJwtPayload('')).toBeNull()
    expect(decodeJwtPayload('not-a-jwt')).toBeNull()
    expect(decodeJwtPayload('a.b.c')).toBeNull()
    expect(getTokenExpiry('a.b.c')).toBeNull()
  })

  it('isTokenExpired 判断过期', () => {
    expect(isTokenExpired(makeExpiredJwt())).toBe(true)
    expect(isTokenExpired(makeFreshJwt({}, 3600))).toBe(false)
    expect(isTokenExpired('bad-token')).toBe(false)
  })

  it('isTokenExpiring 提前识别即将过期', () => {
    expect(isTokenExpiring(makeFreshJwt({}, 60), 120)).toBe(true)
    expect(isTokenExpiring(makeFreshJwt({}, 3600), 120)).toBe(false)
    expect(isTokenExpiring(makeExpiredJwt(), 120)).toBe(true)
  })
})
