import { http } from './request'
import { getRefreshToken } from '@/utils/storage'
import type {
  AdminUserInfoUpdatePayload,
  CaptchaPayload,
  CaptchaResult,
  ListData,
  LoginLogItem,
  LoginLogQuery,
  LoginPayload,
  PageParams,
  RegisterPayload,
  ResetEmailPayload,
  SendEmailPayload,
  SendEmailResult,
  TokenPair,
  UserBaseInfo,
  UserDetail,
  UserInfoUpdatePayload,
  UserListItem,
} from './types'

export function fetchCaptcha(target: string): Promise<CaptchaResult> {
  return http.get<CaptchaResult>('/captcha', { target }, { silent: true })
}

export function sendEmailCode(payload: SendEmailPayload): Promise<SendEmailResult> {
  return http.post<SendEmailResult>('/user/send_email', payload)
}

export function login(payload: LoginPayload): Promise<TokenPair> {
  return http.post<TokenPair>('/user/login', payload)
}

export function registerByEmail(payload: RegisterPayload): Promise<TokenPair> {
  return http.post<TokenPair>('/user/email', payload)
}

export function logout(): Promise<unknown> {
  return http.delete<unknown>('/user/logout', undefined, { silent: true })
}

export function refreshToken(value?: string): Promise<string> {
  const refresh = value || getRefreshToken()
  return http.post<string>('/user/token', null, { headers: { refreshToken: refresh } })
}

export function fetchUserDetail(): Promise<UserDetail> {
  return http.get<UserDetail>('/user/detail')
}

export function fetchUserBaseInfo(id: number): Promise<UserBaseInfo> {
  return http.get<UserBaseInfo>(`/user/info/${id}`)
}

export function fetchUserList(params?: PageParams): Promise<ListData<UserListItem>> {
  return http.get<ListData<UserListItem>>('/user/list', params as Record<string, unknown>)
}

export function updateUserInfo(payload: UserInfoUpdatePayload): Promise<unknown> {
  return http.put<unknown>('/user/update', payload)
}

export function adminUpdateUserInfo(payload: AdminUserInfoUpdatePayload): Promise<unknown> {
  return http.put<unknown>('/user/admin/update', payload)
}

export function fetchLoginLogs(params?: LoginLogQuery): Promise<ListData<LoginLogItem>> {
  return http.get<ListData<LoginLogItem>>('/user/loginlog', params as Record<string, unknown>)
}

export function resetEmail(payload: ResetEmailPayload): Promise<unknown> {
  return http.put<unknown>('/user/resetEmail', payload)
}

export function fetchQqLoginUrl(): Promise<string> {
  return http.get<string>('/site/qq_login', undefined, { silent: true })
}

export type { CaptchaPayload }
