import { http } from './request'
import type { AIConfig, EmailConfig, QQConfig, SiteConfig } from './types'

export function fetchSiteConfig(): Promise<SiteConfig> {
  return http.get<SiteConfig>('/site/site')
}

export function fetchEmailConfig(): Promise<EmailConfig> {
  return http.get<EmailConfig>('/site/email', undefined, { silent: true })
}

export function fetchQQConfig(): Promise<QQConfig> {
  return http.get<QQConfig>('/site/qq', undefined, { silent: true })
}

export function fetchAIConfig(): Promise<AIConfig> {
  return http.get<AIConfig>('/site/ai', undefined, { silent: true })
}

export function fetchQqLoginUrl(): Promise<string> {
  return http.get<string>('/site/qq_login', undefined, { silent: true })
}
