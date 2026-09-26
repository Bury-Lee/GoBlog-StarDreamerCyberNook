import { http } from './request'
import type {
  ListData,
  MomentCommentChildQuery,
  MomentCommentCreatePayload,
  MomentCommentModel,
  MomentCommentQuery,
  MomentCreatePayload,
  MomentModel,
  MomentQuery,
  MomentRepostPayload,
  MomentUpdatePayload,
} from './types'

export function createMoment(payload: MomentCreatePayload): Promise<MomentModel> {
  return http.post<MomentModel>('/moment', payload)
}

export function fetchMoments(params: MomentQuery): Promise<ListData<MomentModel>> {
  return http.get<ListData<MomentModel>>('/moment', params as unknown as Record<string, unknown>)
}

export function fetchMomentDetail(id: number): Promise<MomentModel> {
  return http.get<MomentModel>(`/moment/${id}`)
}

export function updateMoment(payload: MomentUpdatePayload): Promise<unknown> {
  return http.put<unknown>('/moment', payload)
}

export function removeMoment(id: number): Promise<unknown> {
  return http.delete<unknown>(`/moment/${id}`)
}

export interface MomentDiggResult {
  digged: boolean
  likeCount: number
}

export function diggMoment(id: number): Promise<MomentDiggResult> {
  return http.post<MomentDiggResult>(`/moment/digg/${id}`)
}

export interface MomentInteractionResult {
  digged: boolean
}

export function fetchMomentInteraction(id: number): Promise<MomentInteractionResult> {
  return http.get<MomentInteractionResult>(`/moment/interaction/${id}`)
}

export function repostMoment(id: number, payload: MomentRepostPayload = {}): Promise<MomentModel> {
  return http.post<MomentModel>(`/moment/repost/${id}`, payload)
}

export function createMomentComment(payload: MomentCommentCreatePayload): Promise<MomentCommentModel> {
  return http.post<MomentCommentModel>('/moment/comment', payload)
}

export function fetchMomentComments(params: MomentCommentQuery): Promise<ListData<MomentCommentModel>> {
  return http.get<ListData<MomentCommentModel>>('/moment/comment', params as unknown as Record<string, unknown>)
}

export function fetchMomentChildComments(
  params: MomentCommentChildQuery,
): Promise<ListData<MomentCommentModel>> {
  return http.get<ListData<MomentCommentModel>>(
    '/moment/comment/child',
    params as unknown as Record<string, unknown>,
  )
}

export function removeMomentComment(id: number): Promise<unknown> {
  return http.delete<unknown>(`/moment/comment/${id}`)
}

export interface MomentCommentDiggResult {
  digged: boolean
  diggCount: number
}

export function diggMomentComment(id: number): Promise<MomentCommentDiggResult> {
  return http.post<MomentCommentDiggResult>(`/moment/comment/digg/${id}`)
}
