import { describe, expect, it } from 'vitest'
import { processArticleHtml, renderMarkdown, sanitizeHtml } from '@/utils/html'

describe('sanitizeHtml', () => {
  it('移除 script / 事件属性等危险内容', () => {
    const dirty = '<p onclick="alert(1)">hi</p><script>alert(2)</script><img src=x onerror=alert(3)>'
    const clean = sanitizeHtml(dirty)
    expect(clean).not.toContain('<script')
    expect(clean).not.toContain('onclick')
    expect(clean).not.toContain('onerror')
    expect(clean).toContain('hi')
  })

  it('保留搜索结果高亮的 em 标签', () => {
    expect(sanitizeHtml('<em>Go</em>语言')).toContain('<em>Go</em>')
  })

  it('空值返回空字符串', () => {
    expect(sanitizeHtml(null)).toBe('')
    expect(sanitizeHtml('')).toBe('')
  })
})

describe('renderMarkdown', () => {
  it('Markdown 转 HTML', () => {
    const html = renderMarkdown('# 标题\n\n**加粗**\n\n```go\nfmt.Println("hi")\n```')
    expect(html).toContain('<h1')
    expect(html).toContain('<strong>加粗</strong>')
    expect(html).toContain('<code')
  })

  it('渲染时同样会过滤脚本', () => {
    const html = renderMarkdown('<script>alert(1)</script>\n\n正文')
    expect(html).not.toContain('<script')
    expect(html).toContain('正文')
  })
})

describe('processArticleHtml', () => {
  it('为标题生成 id、输出目录(仅 h1-h3),并给外链加 target', () => {
    const { html, toc } = processArticleHtml(
      '<h2>第一节</h2><p>x</p><h3>小节</h3><h4>更深的标题</h4><a href="https://example.com">link</a>',
    )
    expect(toc).toHaveLength(2)
    expect(toc[0].text).toBe('第一节')
    expect(toc[0].level).toBe(2)
    expect(toc[1].level).toBe(3)
    expect(html).toContain(`id="${toc[0].id}"`)
    expect(html).toContain('sd-h-2-')
    expect(toc.some((item) => item.text === '更深的标题')).toBe(false)
    expect(html).toContain('target="_blank"')
    expect(html).toContain('rel="noopener noreferrer nofollow"')
  })

  it('图片补充懒加载属性', () => {
    const { html } = processArticleHtml('<p><img src="/api/image?id=1"></p>')
    expect(html).toContain('loading="lazy"')
  })

  it('空内容返回空结构', () => {
    expect(processArticleHtml('')).toEqual({ html: '', toc: [] })
  })
})
