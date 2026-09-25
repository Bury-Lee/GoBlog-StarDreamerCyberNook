import { http } from './request'
import type {
  FeedbackCreatePayload,
  FeedbackHandlePayload,
  FeedbackItem,
  FeedbackListQuery,
  ListData,
} from './types'

/** 提交反馈(登录可选,可匿名) */
export function createFeedback(payload: FeedbackCreatePayload): Promise<unknown> {
  return http.post<unknown>('/feedback', payload)
}

/** 反馈墙列表(全站公开) */
export function fetchFeedbackList(params?: FeedbackListQuery): Promise<ListData<FeedbackItem>> {
  return http.get<ListData<FeedbackItem>>('/feedback', params as Record<string, unknown>)
}

/** 处理反馈(管理员);返回更新后的条目,用于列表增量替换 */
export function handleFeedback(id: number, payload: FeedbackHandlePayload): Promise<FeedbackItem> {
  return http.put<FeedbackItem>(`/feedback/${id}`, payload)
}
