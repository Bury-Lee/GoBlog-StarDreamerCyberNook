import { describe, expect, it } from 'vitest'
import {
  articleStatusLabel,
  articleStatusType,
  chatMsgTypeLabel,
  excerpt,
  formatDate,
  formatFileSize,
  formatNumber,
  logLevelLabel,
  logTypeLabel,
  messageTypeLabel,
  roleLabel,
  stripHtml,
} from '@/utils/format'

describe('formatNumber', () => {
  it('按千 / 万 / 亿分级', () => {
    expect(formatNumber(0)).toBe('0')
    expect(formatNumber(999)).toBe('999')
    expect(formatNumber(1500)).toBe('1.5k')
    expect(formatNumber(12000)).toBe('1.2万')
    expect(formatNumber(230000000)).toBe('2.3亿')
  })

  it('空值按 0 处理', () => {
    expect(formatNumber(null)).toBe('0')
    expect(formatNumber(undefined)).toBe('0')
  })
})

describe('formatDate / formatFileSize', () => {
  it('格式化本地时间', () => {
    expect(formatDate('2026-09-21 06:40:22', 'YYYY-MM-DD HH:mm')).toBe('2026-09-21 06:40')
    expect(formatDate('2026-09-21 06:40:22', 'YYYY-MM-DD')).toBe('2026-09-21')
  })

  it('空值返回占位符', () => {
    expect(formatDate(null)).toBe('-')
    expect(formatDate('not-a-date')).toBe('-')
  })

  it('文件大小换算', () => {
    expect(formatFileSize(0)).toBe('0 B')
    expect(formatFileSize(1024)).toBe('1.0 KB')
    expect(formatFileSize(2048)).toBe('2.0 KB')
    expect(formatFileSize(5 * 1024 * 1024)).toBe('5.0 MB')
  })
})

describe('枚举文案', () => {
  it('文章状态(以源码为准: 0草稿/1审核中/2已发布/3已下线)', () => {
    expect(articleStatusLabel(0)).toBe('草稿')
    expect(articleStatusLabel(1)).toBe('审核中')
    expect(articleStatusLabel(2)).toBe('已发布')
    expect(articleStatusLabel(3)).toBe('已下线')
    expect(articleStatusType(2)).toBe('success')
    expect(articleStatusType(3)).toBe('danger')
  })

  it('消息类型(1评论/2回复/3点赞/4收藏/5私信/6系统/7@我)', () => {
    expect(messageTypeLabel(1)).toBe('评论通知')
    expect(messageTypeLabel(4)).toBe('收藏通知')
    expect(messageTypeLabel(7)).toBe('@我')
    expect(messageTypeLabel(99)).toBe('未知消息')
  })

  it('私聊位标志组合', () => {
    expect(chatMsgTypeLabel(0)).toBe('空消息')
    expect(chatMsgTypeLabel(1)).toBe('文本')
    expect(chatMsgTypeLabel(3)).toBe('文本 + 图片')
    expect(chatMsgTypeLabel(7)).toBe('文本 + 图片 + Markdown')
  })

  it('日志类型 / 级别 / 角色', () => {
    expect(logTypeLabel(3)).toBe('运行时日志')
    expect(logLevelLabel(2)).toBe('Warn')
    expect(roleLabel(1)).toBe('管理员')
    expect(roleLabel(6)).toBe('封禁用户')
  })
})

describe('HTML 文本处理', () => {
  it('stripHtml 去标签并还原实体', () => {
    expect(stripHtml('<p>你好&nbsp;<b>世界</b></p>')).toBe('你好 世界')
    expect(stripHtml('a &amp; b')).toBe('a & b')
    expect(stripHtml(null)).toBe('')
  })

  it('excerpt 截断并追加省略号', () => {
    expect(excerpt('<p>abcdefghij</p>', 5)).toBe('abcde…')
    expect(excerpt('<p>abc</p>', 5)).toBe('abc')
  })
})
