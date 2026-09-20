import { http } from './request'
import type {
  CommentChildQuery,
  CommentCreatePayload,
  CommentModel,
  CommentQuery,
  ListData,
} from './types'

export function createComment(payload: CommentCreatePayload): Promise<unknown> {
  return http.post<unknown>('/comment', payload)
}

export function fetchComments(params: CommentQuery): Promise<ListData<CommentModel>> {
  return http.get<ListData<CommentModel>>('/comment', params as unknown as Record<string, unknown>)
}

export function fetchChildComments(params: CommentChildQuery): Promise<ListData<CommentModel>> {
  return http.get<ListData<CommentModel>>('/commentChild', params as unknown as Record<string, unknown>)
}

export function removeComment(id: number): Promise<unknown> {
  return http.delete<unknown>(`/comment/${id}`)
}

export function diggComment(id: number): Promise<unknown> {
  return http.post<unknown>(`/comment/digg/${id}`)
}
