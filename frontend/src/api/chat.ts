import { http } from './request'
import type { AiChatPayload, AiChatResult, ChatItem, ChatSendPayload, ChatSession, ListData, PageParams } from './types'

export function sendChatMessage(payload: ChatSendPayload): Promise<unknown> {
  return http.post<unknown>('/chat/send', payload)
}

export function fetchChatHistory(params: PageParams & { userID: number }): Promise<ListData<ChatItem>> {
  return http.get<ListData<ChatItem>>('/chat/get', params as unknown as Record<string, unknown>)
}

export function fetchChatSessions(params?: PageParams): Promise<ListData<ChatSession>> {
  return http.get<ListData<ChatSession>>('/chat/session', params as Record<string, unknown>)
}

export function askAi(payload: AiChatPayload): Promise<AiChatResult> {
  return http.post<AiChatResult>('/chat', payload, { timeout: 90000 })
}
