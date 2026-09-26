import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

export function formatDate(value?: string | number | Date | null, pattern = 'YYYY-MM-DD HH:mm'): string {
  if (!value) return '-'
  const date = dayjs(value)
  if (!date.isValid()) return '-'
  return date.format(pattern)
}

export function formatDateShort(value?: string | number | Date | null): string {
  return formatDate(value, 'YYYY-MM-DD')
}

export function fromNow(value?: string | number | Date | null): string {
  if (!value) return '-'
  const date = dayjs(value)
  if (!date.isValid()) return '-'
  return date.fromNow()
}

export function formatNumber(value?: number | null): string {
  const num = Number(value || 0)
  if (num >= 100000000) return `${(num / 100000000).toFixed(1)}亿`
  if (num >= 10000) return `${(num / 10000).toFixed(1)}万`
  return String(num)
}

export function formatFileSize(size?: number | null): string {
  const bytes = Number(size || 0)
  if (bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

export function formatSeconds(seconds?: number | null): string {
  const total = Math.max(0, Math.floor(Number(seconds || 0)))
  const m = Math.floor(total / 60)
  const s = total % 60
  if (m <= 0) return `${s} 秒`
  return `${m} 分 ${s} 秒`
}

export const ARTICLE_STATUS_OPTIONS = [
  { label: '草稿', value: 0, type: 'info' as const },
  { label: '审核中', value: 1, type: 'warning' as const },
  { label: '已发布', value: 2, type: 'success' as const },
  { label: '已下线', value: 3, type: 'danger' as const },
]

export function articleStatusLabel(status?: number | null): string {
  return ARTICLE_STATUS_OPTIONS.find((item) => item.value === status)?.label || '未知'
}

export function articleStatusType(status?: number | null): 'info' | 'warning' | 'success' | 'danger' {
  return ARTICLE_STATUS_OPTIONS.find((item) => item.value === status)?.type || 'info'
}

export const ROLE_OPTIONS = [
  { label: '管理员', value: 1, type: 'danger' as const },
  { label: '超级会员', value: 2, type: 'warning' as const },
  { label: '会员', value: 3, type: 'success' as const },
  { label: '用户', value: 4, type: 'primary' as const },
  { label: '访客', value: 5, type: 'info' as const },
  { label: '封禁用户', value: 6, type: 'info' as const },
]

export function roleLabel(role?: number | null): string {
  return ROLE_OPTIONS.find((item) => item.value === role)?.label || '未知'
}

export function roleType(role?: number | null): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  return ROLE_OPTIONS.find((item) => item.value === role)?.type || 'info'
}

export const MESSAGE_TYPES = [
  { label: '评论通知', value: 1, icon: 'ChatDotRound' },
  { label: '回复通知', value: 2, icon: 'ChatLineRound' },
  { label: '点赞通知', value: 3, icon: 'Pointer' },
  { label: '收藏通知', value: 4, icon: 'Star' },
  { label: '私信通知', value: 5, icon: 'Message' },
  { label: '系统通知', value: 6, icon: 'Bell' },
  { label: '@我', value: 7, icon: 'At' },
]

export function messageTypeLabel(type?: number | null): string {
  return MESSAGE_TYPES.find((item) => item.value === type)?.label || '未知消息'
}

export const MOMENT_TYPE_OPTIONS = [
  { label: '动态', value: 0 },
  { label: '日记', value: 1 },
]

export function momentTypeLabel(type?: number | null): string {
  return MOMENT_TYPE_OPTIONS.find((item) => item.value === type)?.label || '动态'
}

export const MOMENT_VISIBILITY_OPTIONS = [
  { label: '公开', value: 0, type: 'success' as const },
  { label: '仅好友', value: 1, type: 'warning' as const },
  { label: '私密', value: 2, type: 'info' as const },
]

export function momentVisibilityLabel(visibility?: number | null): string {
  return MOMENT_VISIBILITY_OPTIONS.find((item) => item.value === visibility)?.label || '公开'
}

export function momentVisibilityType(visibility?: number | null): 'success' | 'warning' | 'info' {
  return MOMENT_VISIBILITY_OPTIONS.find((item) => item.value === visibility)?.type || 'success'
}

export const LOG_TYPE_OPTIONS = [
  { label: '登录日志', value: 1 },
  { label: '操作日志', value: 2 },
  { label: '运行时日志', value: 3 },
]

export function logTypeLabel(type?: number | null): string {
  return LOG_TYPE_OPTIONS.find((item) => item.value === type)?.label || '未知类型'
}

export const LOG_LEVEL_OPTIONS = [
  { label: '信息', value: 1, type: 'info' as const },
  { label: '警告', value: 2, type: 'warning' as const },
  { label: '错误', value: 3, type: 'danger' as const },
]

export function logLevelLabel(level?: number | null): string {
  return LOG_LEVEL_OPTIONS.find((item) => item.value === level)?.label || '信息'
}

export function logLevelType(level?: number | null): 'info' | 'warning' | 'danger' {
  return LOG_LEVEL_OPTIONS.find((item) => item.value === level)?.type || 'info'
}

export function chatMsgTypeLabel(msgType?: number | null): string {
  const type = Number(msgType || 0)
  const labels: string[] = []
  if (type & 1) labels.push('文本')
  if (type & 2) labels.push('图片')
  if (type & 4) labels.push('Markdown')
  return labels.join(' + ') || '空消息'
}

export const SEARCH_SORT_OPTIONS = [
  { label: '最新发布', value: 0 },
  { label: '猜你喜欢', value: 1 },
  { label: '最多回复', value: 2 },
  { label: '最多点赞', value: 3 },
  { label: '最多收藏', value: 4 },
]

export const ARTICLE_ORDER_OPTIONS = [
  { label: '最新发布', value: '' },
  { label: '最多浏览', value: 'look_count desc' },
  { label: '最多点赞', value: 'digg_count desc' },
  { label: '最多评论', value: 'comment_count desc' },
  { label: '最多收藏', value: 'collect_count desc' },
]

export interface AiQualityParts {
  score: string
  comment: string
}

export function parseAiQuality(raw?: string | null): AiQualityParts {
  const text = (raw || '').trim()
  if (!text) return { score: '', comment: '' }
  const scoreMatch = text.match(/(\d+(?:\.\d+)?)\s*\/\s*(\d+)/)
  const score = scoreMatch ? `${scoreMatch[1]}/${scoreMatch[2]}` : text.length <= 12 ? text : ''
  const commentMatch = text.match(/简评[:：]?\s*([\s\S]+)$/)
  const comment = commentMatch ? commentMatch[1].trim() : score ? '' : text
  return { score, comment }
}

export function stripHtml(html?: string | null): string {  if (!html) return ''
  return html
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/\s+/g, ' ')
    .trim()
}

export function excerpt(html?: string | null, max = 120): string {
  const text = stripHtml(html)
  if (text.length <= max) return text
  return `${text.slice(0, max)}…`
}

export function debounce<T extends (...args: never[]) => void>(fn: T, wait = 300) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), wait)
  }
}

export function toNumber(value: unknown, fallback = 0): number {
  const num = Number(value)
  return Number.isFinite(num) ? num : fallback
}
