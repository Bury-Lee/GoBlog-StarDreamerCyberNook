import { http } from './request'
import type {
  Banner,
  BannerPayload,
  FriendLink,
  FriendPromotion,
  ImageItem,
  ImageListItem,
  ListData,
  LogModel,
  LogQuery,
  PageParams,
  RemoveRequest,
} from './types'

export function fetchBanners(params?: PageParams): Promise<ListData<Banner>> {
  return http.get<ListData<Banner>>('/banner', params as Record<string, unknown>)
}

export function createBanner(payload: BannerPayload): Promise<unknown> {
  return http.post<unknown>('/banner', payload)
}

export function updateBanner(id: number, payload: BannerPayload): Promise<unknown> {
  return http.put<unknown>(`/banner/${id}`, payload)
}

export function removeBanners(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/banner', { IDList: idList } satisfies RemoveRequest)
}

export interface FriendLinkPayload {
  name?: string
  url?: string
  logo?: string
  is_show?: boolean
  sort_order?: number
  remark?: string
}

export function fetchFriendLinks(params?: PageParams): Promise<ListData<FriendLink>> {
  return http.get<ListData<FriendLink>>('/friendLink', params as Record<string, unknown>)
}

export function createFriendLink(payload: FriendLinkPayload): Promise<unknown> {
  return http.post<unknown>('/friendLink', payload)
}

export function updateFriendLink(id: number, payload: FriendLinkPayload): Promise<unknown> {
  return http.put<unknown>(`/friendLink/${id}`, payload)
}

export function removeFriendLinks(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/friendLink', { IDList: idList } satisfies RemoveRequest)
}

export interface FriendPromotionPayload {
  title?: string
  friend_name?: string
  avatar?: string
  category?: string
  description?: string
  preview_images?: string
  contact_info?: string[]
  is_show?: boolean
  sort_order?: number
  position?: string
  remark?: string
}

export function fetchFriendPromotions(params?: PageParams): Promise<ListData<FriendPromotion>> {
  return http.get<ListData<FriendPromotion>>('/friendPromotion', params as Record<string, unknown>)
}

export function createFriendPromotion(payload: FriendPromotionPayload): Promise<unknown> {
  return http.post<unknown>('/friendPromotion', payload)
}

export function updateFriendPromotion(id: number, payload: FriendPromotionPayload): Promise<unknown> {
  return http.put<unknown>(`/friendPromotion/${id}`, payload)
}

export function removeFriendPromotions(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/friendPromotion', { IDList: idList } satisfies RemoveRequest)
}

export function uploadImage(file: File): Promise<number> {
  const form = new FormData()
  form.append('file', file)
  return http.upload<number>('/image', form)
}

export function fetchImages(params?: PageParams): Promise<ListData<ImageListItem>> {
  return http.get<ListData<ImageListItem>>('/images', params as Record<string, unknown>)
}

export function removeImages(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/image', { IDlist: idList })
}

export function fetchLogs(params?: LogQuery): Promise<ListData<LogModel>> {
  return http.get<ListData<LogModel>>('/logs', params as Record<string, unknown>)
}

export function fetchLogDetail(id: number): Promise<unknown> {
  return http.get<unknown>(`/logs/${id}`)
}

export function removeLogs(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/logs', { IDList: idList } satisfies RemoveRequest)
}

export type { ImageItem }
