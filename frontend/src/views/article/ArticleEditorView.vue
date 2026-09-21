<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="editor-head sd-panel">
        <el-input
          v-model="form.title"
          class="editor-head__title"
          size="large"
          maxlength="120"
          show-word-limit
          placeholder="请输入文章标题"
        />
        <div class="editor-head__actions">
          <el-tag v-if="isEdit" :type="articleStatusType(article?.status)" size="small">
            当前状态:{{ articleStatusLabel(article?.status) }}
          </el-tag>
          <el-button @click="router.back()">返回</el-button>
          <template v-if="isEdit">
            <el-button v-if="article?.status !== 0" :loading="saving" @click="save('draft')">
              转为草稿
            </el-button>
            <el-button :loading="saving" @click="save('keep')">保存修改</el-button>
            <el-button
              v-if="article?.status === 0"
              type="primary"
              :loading="saving"
              @click="save('publish')"
            >
              {{ reviewEnabled ? '提交审核' : '发布文章' }}
            </el-button>
          </template>
          <template v-else>
            <el-button :loading="saving" @click="save('draft')">存草稿</el-button>
            <el-button type="primary" :loading="saving" @click="save('publish')">
              {{ reviewEnabled ? '提交审核' : '发布文章' }}
            </el-button>
          </template>
        </div>
      </div>

      <div class="editor-layout">
        <section class="editor-main sd-panel">
          <header class="editor-main__bar">
            <div class="editor-main__tabs">
              <button
                class="editor-main__tab"
                :class="{ 'is-active': mode === 'markdown' }"
                @click="mode = 'markdown'"
              >
                Markdown
              </button>
              <button
                class="editor-main__tab"
                :class="{ 'is-active': mode === 'html' }"
                @click="mode = 'html'"
              >
                HTML 源码
              </button>
            </div>
            <div class="editor-main__tools">
              <el-button size="small" :loading="uploading" @click="pickImage">
                <el-icon><Picture /></el-icon>
                插入图片
              </el-button>
              <el-button size="small" @click="insertCode">
                <el-icon><Document /></el-icon>
                代码块
              </el-button>
              <el-button size="small" @click="previewVisible = !previewVisible">
                <el-icon><View /></el-icon>
                {{ previewVisible ? '隐藏预览' : '显示预览' }}
              </el-button>
            </div>
          </header>

          <div class="editor-main__body" :class="{ 'has-preview': previewVisible }">
            <el-input
              ref="textareaRef"
              v-model="form.content"
              type="textarea"
              resize="none"
              class="editor-main__textarea"
              :placeholder="mode === 'markdown' ? '使用 Markdown 书写正文…' : '粘贴或编写 HTML 正文…'"
            />
            <div v-if="previewVisible" class="editor-main__preview article-content" v-html="previewHtml" />
          </div>
        </section>

        <aside class="editor-side">
          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">文章设置</span>
            </header>
            <div class="sd-panel__body editor-side__body">
              <div class="editor-field">
                <label class="editor-field__label">
                  摘要
                  <span class="sd-dim">(最多 200 字,留空由 AI 生成)</span>
                </label>
                <el-input
                  v-model="form.abstract"
                  type="textarea"
                  :rows="3"
                  resize="none"
                  maxlength="200"
                  show-word-limit
                  placeholder="一句话概括文章内容"
                />
              </div>

              <div class="editor-field">
                <label class="editor-field__label">分类</label>
                <div class="editor-field__row">
                  <el-select v-model="form.categoryID" placeholder="选择分类" clearable class="editor-field__grow">
                    <el-option
                      v-for="item in categories"
                      :key="item.id"
                      :label="item.title"
                      :value="item.id"
                    />
                  </el-select>
                  <el-button @click="createCategory">新建</el-button>
                </div>
              </div>

              <div class="editor-field">
                <label class="editor-field__label">标签</label>
                <div class="editor-tags">
                  <el-tag
                    v-for="tag in form.tagList"
                    :key="tag"
                    closable
                    class="editor-tags__item"
                    @close="removeTag(tag)"
                  >
                    {{ tag }}
                  </el-tag>
                  <el-input
                    v-model="tagInput"
                    class="editor-tags__input"
                    size="small"
                    placeholder="输入后回车添加"
                    maxlength="20"
                    @keyup.enter="addTag"
                  />
                </div>
              </div>

              <div class="editor-field">
                <label class="editor-field__label">封面</label>
                <ImageUploader v-model="form.cover" :width="140" :height="96" />
              </div>

              <div class="editor-field editor-field--inline">
                <label class="editor-field__label">允许评论</label>
                <el-switch v-model="form.openComment" />
              </div>

              <div class="editor-field">
                <label class="editor-field__label">发布说明</label>
                <p class="sd-dim editor-field__tip">
                  {{
                    reviewEnabled
                      ? '站点已开启审核:提交后需管理员审核通过才会公开,草稿不会公开'
                      : '站点未开启审核:提交后立即公开,草稿不会公开'
                  }}
                </p>
              </div>
            </div>
          </section>

          <section class="sd-panel editor-tips">
            <header class="sd-panel__header">
              <span class="sd-panel__title">写作提示</span>
            </header>
            <div class="sd-panel__body">
              <ul class="editor-tips__list">
                <li>正文会自动清理不安全的代码,正常排版不受影响</li>
                <li>图片上传后会自动插入到光标所在位置</li>
                <li>Markdown 模式在保存时会自动转换为网页格式</li>
                <li>编辑已有文章时会展示已保存的正文,可直接修改</li>
              </ul>
            </div>
          </section>
        </aside>
      </div>

      <input
        ref="fileInputRef"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp,image/bmp,image/tiff"
        class="editor-file"
        @change="onFileChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Picture, View } from '@element-plus/icons-vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import { createArticle, fetchArticleDetail, fetchCategories, saveCategory, updateArticle } from '@/api/article'
import { uploadImage } from '@/api/ops'
import { imageUrl } from '@/api/request'
import type { ArticleDetailResponse, ArticleUpdatePayload, CategoryListItem } from '@/api/types'
import { articleStatusLabel, articleStatusType } from '@/utils/format'
import { renderMarkdown, sanitizeHtml } from '@/utils/html'
import { useSiteStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const siteStore = useSiteStore()

const mode = ref<'markdown' | 'html'>('markdown')
const previewVisible = ref(false)
const saving = ref(false)
const uploading = ref(false)
const tagInput = ref('')
const categories = ref<CategoryListItem[]>([])
const article = ref<ArticleDetailResponse | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const textareaRef = ref<{ textarea?: HTMLTextAreaElement } | null>(null)

const form = reactive({
  title: '',
  abstract: '',
  content: '',
  categoryID: undefined as number | undefined,
  tagList: [] as string[],
  cover: '',
  openComment: true,
})

const articleID = computed(() => Number(route.params.id || 0))
const isEdit = computed(() => articleID.value > 0)
const reviewEnabled = computed(() => siteStore.reviewEnabled)

//记录进入页面时的表单快照,离开前提示未保存的修改
const initialSnapshot = ref('')
const dirty = computed(() => initialSnapshot.value !== '' && snapshot() !== initialSnapshot.value)

function snapshot(): string {
  return JSON.stringify({
    title: form.title,
    abstract: form.abstract,
    content: form.content,
    categoryID: form.categoryID ?? 0,
    tagList: form.tagList,
    cover: form.cover,
    openComment: form.openComment,
  })
}

const previewHtml = computed(() => {
  if (!form.content) return '<p class="sd-dim">暂无内容</p>'
  return mode.value === 'markdown' ? renderMarkdown(form.content) : sanitizeHtml(form.content)
})

async function loadCategories(): Promise<void> {
  try {
    const data = await fetchCategories({ type: 'self', page: 1, limit: 40 })
    categories.value = data?.list ?? []
  } catch {
    categories.value = []
  }
}

async function loadArticle(): Promise<void> {
  if (!isEdit.value) return
  try {
    const data = await fetchArticleDetail(articleID.value)
    article.value = data
    form.title = data.title
    form.abstract = data.abstract
    form.content = data.content
    form.categoryID = data.categoryID ?? undefined
    form.tagList = data.tagList || []
    form.cover = data.cover
    form.openComment = data.openComment
    mode.value = 'html'
  } catch {
    ElMessage.error('文章加载失败')
    router.replace({ name: 'my-articles' })
  }
}

function addTag(): void {
  const value = tagInput.value.trim()
  if (!value) return
  if (form.tagList.includes(value)) {
    ElMessage.warning('标签已存在')
    tagInput.value = ''
    return
  }
  if (form.tagList.length >= 8) {
    ElMessage.warning('最多添加 8 个标签')
    return
  }
  form.tagList.push(value)
  tagInput.value = ''
}

function removeTag(tag: string): void {
  form.tagList = form.tagList.filter((item) => item !== tag)
}

async function createCategory(): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt('请输入分类名称', '新建分类', {
      inputPattern: /^.{1,32}$/,
      inputErrorMessage: '分类名称需为 1-32 个字符',
    })
    await saveCategory({ id: 0, title: value })
    ElMessage.success('分类创建成功')
    await loadCategories()
  } catch {
    // 取消或失败
  }
}

function pickImage(): void {
  fileInputRef.value?.click()
}

function insertAtCursor(text: string): void {
  const textarea = textareaRef.value?.textarea
  if (!textarea) {
    form.content += text
    return
  }
  const start = textarea.selectionStart ?? form.content.length
  const end = textarea.selectionEnd ?? form.content.length
  form.content = `${form.content.slice(0, start)}${text}${form.content.slice(end)}`
}

function insertCode(): void {
  const snippet =
    mode.value === 'markdown' ? '\n```go\n// 代码\n```\n' : '\n<pre><code>// 代码</code></pre>\n'
  insertAtCursor(snippet)
}

async function onFileChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploading.value = true
  try {
    const id = await uploadImage(file)
    const url = imageUrl(id)
    insertAtCursor(mode.value === 'markdown' ? `\n![${file.name}](${url})\n` : `\n<img src="${url}" alt="${file.name}" />\n`)
    ElMessage.success('图片已插入')
  } catch {
    // ignore
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function buildContent(): string {
  if (!form.content.trim()) return ''
  return mode.value === 'markdown' ? renderMarkdown(form.content) : sanitizeHtml(form.content)
}

async function save(target: 'draft' | 'publish' | 'keep'): Promise<void> {
  if (!form.title.trim()) {
    ElMessage.warning('请输入文章标题')
    return
  }
  const content = buildContent()
  if (!content) {
    ElMessage.warning('请输入文章正文')
    return
  }
  if (isEdit.value && target === 'draft') {
    try {
      await ElMessageBox.confirm('转为草稿后文章会从站点下线,确定吗?', '转为草稿', {
        type: 'warning',
        confirmButtonText: '转为草稿',
        cancelButtonText: '取消',
      })
    } catch {
      return
    }
  }
  saving.value = true
  try {
    if (isEdit.value) {
      const payload: ArticleUpdatePayload = {
        id: articleID.value,
        title: form.title.trim(),
        abstract: form.abstract.trim() || undefined,
        content,
        categoryID: form.categoryID ?? 0,
        tagList: form.tagList,
        cover: form.cover,
        openComment: form.openComment,
      }
      if (target === 'draft') payload.status = 0
      if (target === 'publish') payload.status = 1
      await updateArticle(payload)
      initialSnapshot.value = snapshot()
      if (target === 'draft') {
        ElMessage.success('已转为草稿')
        router.push({ name: 'my-articles' })
      } else {
        ElMessage.success(target === 'publish' ? (reviewEnabled.value ? '已提交审核' : '文章已发布') : '文章已更新')
        router.push({ name: 'article-detail', params: { id: articleID.value } })
      }
    } else {
      await createArticle({
        title: form.title.trim(),
        abstract: form.abstract.trim() || undefined,
        content,
        categoryID: form.categoryID ?? undefined,
        tagList: form.tagList,
        cover: form.cover,
        openComment: form.openComment,
        status: target === 'publish' ? 1 : 0,
      })
      initialSnapshot.value = snapshot()
      ElMessage.success(
        target === 'publish' ? (reviewEnabled.value ? '已提交审核' : '文章已发布') : '草稿已保存',
      )
      router.push({ name: 'my-articles' })
    }
  } catch {
    // ignore
  } finally {
    saving.value = false
  }
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true
  try {
    await ElMessageBox.confirm('当前修改尚未保存,确定离开吗?', '离开编辑页', {
      type: 'warning',
      confirmButtonText: '离开',
      cancelButtonText: '继续编辑',
    })
    return true
  } catch {
    return false
  }
})

onMounted(async () => {
  await Promise.all([loadCategories(), loadArticle()])
  initialSnapshot.value = snapshot()
})
</script>

<style scoped lang="scss">
.editor-head {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
  padding: 14px 18px;
  margin-bottom: 18px;
}

.editor-head__title {
  flex: 1;
  min-width: 240px;
}

.editor-head__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.editor-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 330px;
  gap: 18px;
  align-items: start;
}

.editor-main {
  overflow: hidden;
}

.editor-main__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  padding: 12px 16px;
  border-bottom: 1px solid var(--sd-border);
}

.editor-main__tabs {
  display: flex;
  gap: 6px;
}

.editor-main__tab {
  padding: 5px 14px;
  font-size: 13px;
  color: var(--sd-text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.editor-main__tab.is-active {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.5);
  background: rgba(34, 211, 238, 0.1);
}

.editor-main__tools {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.editor-main__body {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  min-height: 560px;
}

.editor-main__body.has-preview {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.editor-main__textarea {
  height: 100%;

  :deep(.el-textarea__inner) {
    height: 100%;
    min-height: 560px;
    border-radius: 0;
    border: none;
    padding: 18px;
    font-family: var(--sd-font-mono);
    font-size: 13.5px;
    line-height: 1.8;
    background: rgba(7, 11, 20, 0.6) !important;
    box-shadow: none !important;
  }
}

.editor-main__preview {
  min-height: 560px;
  max-height: 720px;
  overflow-y: auto;
  padding: 18px;
  border-left: 1px solid var(--sd-border);
  font-size: 14px;
  line-height: 1.85;
  color: #cbd5e1;
}

.editor-main__preview :deep(pre) {
  padding: 12px;
  overflow-x: auto;
  border-radius: 10px;
  border: 1px solid var(--sd-border);
  background: rgba(7, 11, 20, 0.9);
}

.editor-main__preview :deep(img) {
  max-width: 100%;
  border-radius: 10px;
}

.editor-main__preview :deep(ul) {
  padding-left: 20px;
  list-style: disc;
}

.editor-main__preview :deep(ol) {
  padding-left: 20px;
  list-style: decimal;
}

.editor-main__preview :deep(li) {
  margin: 3px 0;
}

.editor-main__preview :deep(a) {
  color: var(--sd-cyan);
}

.editor-side {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.editor-side__body {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.editor-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.editor-field--inline {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}

.editor-field__label {
  font-size: 13px;
  color: var(--sd-text-muted);
}

.editor-field__row {
  display: flex;
  gap: 8px;
}

.editor-field__grow {
  flex: 1;
}

.editor-field__tip {
  font-size: 12px;
  line-height: 1.6;
}

.editor-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.editor-tags__input {
  width: 130px;
}

.editor-tips__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
  color: var(--sd-text-muted);

  li::before {
    content: '▹ ';
    color: var(--sd-cyan);
  }
}

.editor-file {
  display: none;
}

@media (max-width: 1100px) {
  .editor-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .editor-main__body.has-preview {
    grid-template-columns: minmax(0, 1fr);
  }

  .editor-main__preview {
    border-left: none;
    border-top: 1px solid var(--sd-border);
  }
}
</style>
