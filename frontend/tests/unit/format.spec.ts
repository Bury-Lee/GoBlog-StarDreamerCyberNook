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
  parseAiQuality,
  roleLabel,
  stripHtml,
} from '@/utils/format'

describe('formatNumber', () => {
  it('按万 / 亿分级,千位不缩写', () => {
    expect(formatNumber(0)).toBe('0')
    expect(formatNumber(999)).toBe('999')
    expect(formatNumber(1500)).toBe('1500')
    expect(formatNumber(9999)).toBe('9999')
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
    expect(logLevelLabel(2)).toBe('警告')
    expect(roleLabel(1)).toBe('管理员')
    expect(roleLabel(6)).toBe('封禁用户')
  })
})

describe('parseAiQuality(把后端整段评级文本拆成标签与简评)', () => {
  it('拆出分数与简评', () => {
    const raw = '评级:8/10分\n简评:文章详细介绍了 GoTenon 的特性,内容准确、结构清晰。'
    const result = parseAiQuality(raw)
    expect(result.score).toBe('8/10')
    expect(result.comment).toBe('文章详细介绍了 GoTenon 的特性,内容准确、结构清晰。')
  })

  it('兼容不同写法(冒号/空格/小数)', () => {
    expect(parseAiQuality('评级：9.5/10分').score).toBe('9.5/10')
    expect(parseAiQuality('评分 7 / 10').score).toBe('7/10')
    expect(parseAiQuality('简评：内容偏短。').comment).toBe('内容偏短。')
  })

  it('短文本直接作为分数标签,长文本没有分数时整体作为简评', () => {
    expect(parseAiQuality('优秀').score).toBe('优秀')
    const long = '这篇文章结构完整,但缺少示例代码,建议补充可运行的最小示例以便读者理解。'
    const result = parseAiQuality(long)
    expect(result.score).toBe('')
    expect(result.comment).toBe(long)
  })

  it('空值返回空结构', () => {
    expect(parseAiQuality('')).toEqual({ score: '', comment: '' })
    expect(parseAiQuality(null)).toEqual({ score: '', comment: '' })
  })
})

describe('HTML 文本处理', () => {  it('stripHtml 去标签并还原实体', () => {
    expect(stripHtml('<p>你好&nbsp;<b>世界</b></p>')).toBe('你好 世界')
    expect(stripHtml('a &amp; b')).toBe('a & b')
    expect(stripHtml(null)).toBe('')
  })

  it('excerpt 截断并追加省略号', () => {
    expect(excerpt('<p>abcdefghij</p>', 5)).toBe('abcde…')
    expect(excerpt('<p>abc</p>', 5)).toBe('abc')
  })
})
