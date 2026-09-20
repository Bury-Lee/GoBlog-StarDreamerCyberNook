import DOMPurify, { type Config } from 'dompurify'
import { marked } from 'marked'

export interface TocItem {
  id: string
  text: string
  level: number
}

const SANITIZE_CONFIG: Config = {
  ADD_ATTR: ['target', 'rel'],
  FORBID_TAGS: ['style', 'form', 'input', 'iframe', 'script'],
  FORBID_ATTR: ['onerror', 'onclick', 'onload', 'style'],
}

export function sanitizeHtml(html?: string | null): string {
  if (!html) return ''
  return DOMPurify.sanitize(html, SANITIZE_CONFIG) as unknown as string
}

export function renderMarkdown(markdown: string): string {
  if (!markdown) return ''
  const raw = marked.parse(markdown, { async: false, gfm: true, breaks: true }) as string
  return sanitizeHtml(raw)
}

function slugify(text: string, index: number): string {
  const base = text
    .trim()
    .toLowerCase()
    .replace(/[\s]+/g, '-')
    .replace(/[^\w\u4e00-\u9fa5-]/g, '')
    .slice(0, 40)
  return `sd-h-${index}-${base || 'section'}`
}

export function processArticleHtml(html?: string | null): { html: string; toc: TocItem[] } {
  const safe = sanitizeHtml(html)
  if (!safe) return { html: '', toc: [] }
  if (typeof window === 'undefined' || !window.DOMParser) {
    return { html: safe, toc: [] }
  }
  const doc = new DOMParser().parseFromString(`<div id="sd-root">${safe}</div>`, 'text/html')
  const root = doc.getElementById('sd-root')
  if (!root) return { html: safe, toc: [] }

  const toc: TocItem[] = []
  root.querySelectorAll('h1, h2, h3, h4').forEach((node, index) => {
    const level = Number(node.tagName.slice(1))
    const text = node.textContent?.trim() || `章节 ${index + 1}`
    const id = node.id || slugify(text, index)
    node.setAttribute('id', id)
    if (level <= 3) {
      toc.push({ id, text, level })
    }
  })

  root.querySelectorAll('a[href]').forEach((node) => {
    const href = node.getAttribute('href') || ''
    if (href.startsWith('http')) {
      node.setAttribute('target', '_blank')
      node.setAttribute('rel', 'noopener noreferrer nofollow')
    }
  })

  root.querySelectorAll('img').forEach((node) => {
    node.setAttribute('loading', 'lazy')
  })

  return { html: root.innerHTML, toc }
}

export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text)
  }
  return new Promise((resolve, reject) => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      resolve()
    } catch (error) {
      reject(error)
    } finally {
      document.body.removeChild(textarea)
    }
  })
}
