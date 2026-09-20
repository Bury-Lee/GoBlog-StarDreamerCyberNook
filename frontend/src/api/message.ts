import { http } from './request'
import type {
  ListData,
  MessageConf,
  MessageConfUpdatePayload,
  MessageModel,
  OneKeyReadPayload,
  PageParams,
} from './types'

export function fetchMessageConf(): Promise<MessageConf> {
  return http.get<MessageConf>('/msg/conf', undefined, { silent: true })
}

export function updateMessageConf(payload: MessageConfUpdatePayload): Promise<unknown> {
  return http.post<unknown>('/msg/conf/update', payload)
}

export function checkUnreadMessages(): Promise<Record<string, number>> {
  return http.get<Record<string, number>>('/msg/check', undefined, { silent: true })
}

export function fetchMessages(
  params: PageParams & { type: number },
): Promise<ListData<MessageModel>> {
  return http.get<ListData<MessageModel>>('/msg', params as unknown as Record<string, unknown>)
}

export function removeMessages(messageID: number[]): Promise<unknown> {
  return http.delete<unknown>('/msg', { messageID })
}

export function clearMessages(payload: OneKeyReadPayload): Promise<unknown> {
  return http.post<unknown>('/msg/clear', payload)
}
