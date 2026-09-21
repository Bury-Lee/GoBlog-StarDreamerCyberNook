import { beforeEach, describe, expect, it } from 'vitest'
import { API_BASE, imageUrl, resolveAssetUrl } from '@/api/request'
import {
  clearAuth,
  clearTokens,
  getAccessToken,
  getRefreshToken,
  readProfile,
  setAccessToken,
  setTokens,
  writeProfile,
} from '@/utils/storage'

beforeEach(() => {
  localStorage.clear()
})

describe('token / profile 本地存储', () => {
  it('读写与清理 token', () => {
    setTokens('access-1', 'refresh-1')
    expect(getAccessToken()).toBe('access-1')
    expect(getRefreshToken()).toBe('refresh-1')

    setAccessToken('access-2')
    expect(getAccessToken()).toBe('access-2')
    expect(getRefreshToken()).toBe('refresh-1')

    clearTokens()
    expect(getAccessToken()).toBe('')
    expect(getRefreshToken()).toBe('')
  })

  it('profile 序列化、容错与清理', () => {
    writeProfile({ id: 7, nickname: '赛博用户' })
    expect(readProfile<{ id: number; nickname: string }>()?.nickname).toBe('赛博用户')

    localStorage.setItem('goblog.profile', '{ 非法 JSON')
    expect(readProfile()).toBeNull()

    clearAuth()
    expect(readProfile()).toBeNull()
  })
})

describe('资源地址工具', () => {
  it('imageUrl 指向 /api/image', () => {
    expect(imageUrl(42)).toBe(`${API_BASE}/image?id=42`)
    expect(imageUrl('42')).toContain('/api/image?id=42')
  })

  it('resolveAssetUrl 处理外链 / 绝对路径 / 相对路径', () => {
    expect(resolveAssetUrl('https://cdn.example.com/a.png')).toBe('https://cdn.example.com/a.png')
    expect(resolveAssetUrl('//cdn.example.com/a.png')).toBe('//cdn.example.com/a.png')
    expect(resolveAssetUrl('data:image/png;base64,AAA')).toBe('data:image/png;base64,AAA')
    expect(resolveAssetUrl('/api/image?id=1')).toBe('/api/image?id=1')
    expect(resolveAssetUrl('/web/logo.png')).toBe('/web/logo.png')
    expect(resolveAssetUrl('/static/logo.png')).toBe('/web/static/logo.png')
    expect(resolveAssetUrl('images/a.png')).toBe('/web/images/a.png')
    expect(resolveAssetUrl('')).toBe('')
  })
})
