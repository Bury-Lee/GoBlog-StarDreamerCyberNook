import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { login as loginApi, logout as logoutApi, registerByEmail, fetchUserDetail, updateUserInfo } from '@/api/user'
import type {
  LoginPayload,
  RegisterPayload,
  TokenPair,
  UserDetail,
  UserInfoUpdatePayload,
} from '@/api/types'
import {
  clearAuth,
  clearTokens,
  getAccessToken,
  getRefreshToken,
  readProfile,
  setTokens,
  writeProfile,
} from '@/utils/storage'

export const useUserStore = defineStore('user', () => {
  const accessToken = ref(getAccessToken())
  const refreshTokenValue = ref(getRefreshToken())
  const profile = ref<UserDetail | null>(readProfile<UserDetail>())
  const profileLoading = ref(false)

  const isLogin = computed(() => Boolean(accessToken.value))
  const isAdmin = computed(() => profile.value?.role === 1)
  const userId = computed(() => profile.value?.id ?? 0)
  const nickname = computed(() => profile.value?.nickname || profile.value?.username || '未登录')
  const avatar = computed(() => profile.value?.avatar || '')

  function applyTokens(tokens: TokenPair | null | undefined): void {
    if (!tokens?.AccessToken) return
    accessToken.value = tokens.AccessToken
    refreshTokenValue.value = tokens.RefreshToken || refreshTokenValue.value
    setTokens(tokens.AccessToken, tokens.RefreshToken)
  }

  function setProfile(next: UserDetail | null): void {
    profile.value = next
    writeProfile(next)
  }

  async function loadProfile(force = false): Promise<UserDetail | null> {
    if (!accessToken.value) return null
    if (profile.value && !force) return profile.value
    profileLoading.value = true
    try {
      const detail = await fetchUserDetail()
      setProfile(detail)
      return detail
    } catch {
      return profile.value
    } finally {
      profileLoading.value = false
    }
  }

  async function login(payload: LoginPayload): Promise<void> {
    const tokens = await loginApi(payload)
    applyTokens(tokens)
    await loadProfile(true)
  }

  async function register(payload: RegisterPayload): Promise<void> {
    const tokens = await registerByEmail(payload)
    applyTokens(tokens)
    await loadProfile(true)
  }

  async function updateProfile(payload: UserInfoUpdatePayload): Promise<void> {
    await updateUserInfo(payload)
    await loadProfile(true)
  }

  function patchProfile(patch: Partial<UserDetail>): void {
    if (!profile.value) return
    setProfile({ ...profile.value, ...patch })
  }

  async function logout(): Promise<void> {
    try {
      await logoutApi()
    } catch {
      // ignore
    } finally {
      clearAuth()
      accessToken.value = ''
      refreshTokenValue.value = ''
      profile.value = null
    }
  }

  function forceLogout(): void {
    clearTokens()
    writeProfile(null)
    accessToken.value = ''
    refreshTokenValue.value = ''
    profile.value = null
  }

  return {
    accessToken,
    refreshTokenValue,
    profile,
    profileLoading,
    isLogin,
    isAdmin,
    userId,
    nickname,
    avatar,
    applyTokens,
    setProfile,
    loadProfile,
    login,
    register,
    updateProfile,
    patchProfile,
    logout,
    forceLogout,
  }
})
