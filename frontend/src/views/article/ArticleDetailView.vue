<template>
  <div class="detail-view">
    <div class="detail-view__progress" :style="{ width: `${progress}%` }" />

    <div class="sd-container detail-view__container">
      <div v-loading="loading" class="detail-view__layout">
        <template v-if="article">
          <main class="sd-section-gap">
            <el-alert
              v-if="article.status !== 2"
              class="detail-view__status"
              :type="article.status === 3 ? 'error' : article.status === 1 ? 'warning' : 'info'"
              :closable="false"
              show-icon
              :title="`当前文章状态:${articleStatusLabel(article.status)}(仅作者与管理员可见)`"
            />

            <article class="detail-view__article sd-panel">
              <header class="detail-view__head">
                <h1 class="detail-view__title">{{ article.title }}</h1>
                <div class="detail-view__meta">
                  <router-link
                    class="detail-view__author"
                    :to="{ name: 'user-home', params: { id: article.userID } }"
                  >
                    <UserAvatar :src="article.userAvatar" :name="article.nickname" :size="38" />
                    <span class="detail-view__author-info">
                      <span class="detail-view__author-name">{{ article.nickname || article.username }}</span>
                      <span class="sd-dim detail-view__author-time">
                        发布于 {{ formatDate(article.createdAt) }} · 更新于 {{ fromNow(article.updatedAt) }}
                      </span>
                    </span>
                  </router-link>

                  <div class="detail-view__stats">
                    <span><el-icon><View /></el-icon>{{ formatNumber(article.lookCount) }} 浏览</span>
                    <span><el-icon><Pointer /></el-icon>{{ formatNumber(article.diggCount) }} 点赞</span>
                    <span><el-icon><ChatDotRound /></el-icon>{{ formatNumber(article.commentCount) }} 评论</span>
                    <span><el-icon><Star /></el-icon>{{ formatNumber(article.collectCount) }} 收藏</span>
                  </div>
                </div>

                <div class="detail-view__tags">
                  <span v-if="article.categoryTitle" class="sd-tag sd-tag--purple">
                    {{ article.categoryTitle }}
                  </span>
                  <router-link
                    v-for="tag in article.tagList || []"
                    :key="tag"
                    class="sd-tag"
                    :to="{ name: 'search', query: { tag } }"
                  >
                    # {{ tag }}
                  </router-link>
                </div>

                <el-image
                  v-if="article.cover"
                  class="detail-view__cover"
                  :src="resolveAssetUrl(article.cover)"
                  fit="cover"
                  :preview-src-list="[resolveAssetUrl(article.cover)]"
                  preview-teleported
                />
              </header>

              <div v-if="article.aiAbstract" class="detail-view__ai">
                <div class="detail-view__ai-head">
                  <el-icon><MagicStick /></el-icon>
                  <span>AI 摘要</span>
                  <el-tag v-if="article.aiQuality" size="small" type="success" effect="dark">
                    质量评级 {{ article.aiQuality }}
                  </el-tag>
                </div>
                <p>{{ article.aiAbstract }}</p>
              </div>

              <div class="detail-view__content article-content" v-html="contentHtml" />

              <div class="detail-view__actions">
                <el-button
                  class="detail-view__action"
                  :class="{ 'is-active': digged }"
                  round
                  @click="onDigg"
                >
                  <el-icon><Pointer /></el-icon>
                  {{ digged ? '已点赞' : '点赞' }} {{ formatNumber(article.diggCount) }}
                </el-button>
                <el-button class="detail-view__action" round @click="collectVisible = true">
                  <el-icon><Star /></el-icon>
                  收藏 {{ formatNumber(article.collectCount) }}
                </el-button>
                <el-button class="detail-view__action" round @click="copyLink">
                  <el-icon><Link /></el-icon>
                  复制链接
                </el-button>
                <template v-if="isOwner || userStore.isAdmin">
                  <el-button class="detail-view__action" round @click="goEdit">
                    <el-icon><EditPen /></el-icon>
                    编辑
                  </el-button>
                  <el-button class="detail-view__action" round @click="topThis">
                    <el-icon><Top /></el-icon>
                    {{ userTop ? '取消置顶' : '置顶' }}
                  </el-button>
                  <el-button class="detail-view__action is-danger" round @click="removeThis">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </template>
              </div>
            </article>

            <CommentSection :article-id="article.id" :open-comment="article.openComment" />
          </main>

          <aside class="detail-view__aside">
            <nav v-if="toc.length" class="detail-view__toc sd-panel">
              <header class="sd-panel__header">
                <span class="sd-panel__title">目录</span>
              </header>
              <div class="sd-panel__body">
                <a
                  v-for="item in toc"
                  :key="item.id"
                  class="detail-view__toc-item"
                  :class="`is-level-${item.level}`"
                  :href="`#${item.id}`"
                  @click.prevent="scrollToHeading(item.id)"
                >
                  {{ item.text }}
                </a>
              </div>
            </nav>

            <section class="detail-view__author-card sd-panel">
              <header class="sd-panel__header">
                <span class="sd-panel__title">关于作者</span>
              </header>
              <div class="sd-panel__body detail-view__author-body">
                <UserAvatar :src="article.userAvatar" :name="article.nickname" :size="54" />
                <span class="detail-view__author-card-name">{{ article.nickname || article.username }}</span>
                <p class="sd-dim">作者已发布文章,点击查看主页</p>
                <div class="detail-view__author-actions">
                  <el-button size="small" @click="goAuthorHome">访问主页</el-button>
                  <el-button size="small" type="primary" plain @click="goAuthorArticles">
                    他的文章
                  </el-button>
                </div>
              </div>
            </section>

            <ArticleMiniList
              v-if="relatedTag"
              :title="`相关文章 · ${relatedTag}`"
              :params="{ type: 'other', tag: relatedTag }"
              count-field="lookCount"
              :limit="5"
              :more-to="{ name: 'search', query: { tag: relatedTag } }"
            />
          </aside>
        </template>

        <EmptyState v-else-if="!loading" text="文章不存在或已被删除" />
      </div>
    </div>

    <el-dialog v-model="collectVisible" title="收藏到收藏夹" width="440px">
      <el-select v-model="collectID" class="collect-select" placeholder="选择收藏夹">
        <el-option label="默认收藏夹" :value="0" />
        <el-option
          v-for="folder in folders"
          :key="folder.id"
          :label="folder.title"
          :value="folder.id"
        />
      </el-select>
      <p class="sd-dim collect-tip">已收藏的文章再次收藏会取消收藏</p>
      <template #footer>
        <el-button @click="collectVisible = false">取消</el-button>
        <el-button type="primary" :loading="collecting" @click="submitCollect">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  Delete,
  EditPen,
  Link,
  MagicStick,
  Pointer,
  Star,
  Top,
  View,
} from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ArticleMiniList from '@/components/article/ArticleMiniList.vue'
import CommentSection from '@/components/comment/CommentSection.vue'
import {
  cancelTopArticle,
  collectArticle,
  diggArticle,
  fetchArticleDetail,
  fetchCollectFolders,
  removeArticles,
  reportArticleLook,
  topArticle,
} from '@/api/article'
import { resolveAssetUrl } from '@/api/request'
import type { ArticleDetailResponse, CollectModel } from '@/api/types'
import { articleStatusLabel, formatDate, formatNumber, fromNow } from '@/utils/format'
import { copyToClipboard, processArticleHtml, type TocItem } from '@/utils/html'
import { useUserStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const article = ref<ArticleDetailResponse | null>(null)
const loading = ref(true)
const digged = ref(false)
const userTop = ref(false)
const progress = ref(0)
const toc = ref<TocItem[]>([])
const contentHtml = ref('')
const relatedTag = ref('')
const collectVisible = ref(false)
const collecting = ref(false)
const collectID = ref(0)
const folders = ref<CollectModel[]>([])
let lookTimer: ReturnType<typeof setTimeout> | null = null

const articleID = computed(() => Number(route.params.id))
const isOwner = computed(() => Boolean(article.value && userStore.userId === article.value.userID))

async function loadArticle(): Promise<void> {
  if (!articleID.value) return
  loading.value = true
  try {
    const data = await fetchArticleDetail(articleID.value)
    article.value = data
    const processed = processArticleHtml(data?.content)
    contentHtml.value = processed.html
    toc.value = processed.toc
    relatedTag.value = (data?.tagList || [])[0] || ''
    document.title = `${data?.title || '文章详情'} · ${document.title.split(' · ').pop() || ''}`
    scheduleLookReport()
  } catch {
    article.value = null
  } finally {
    loading.value = false
  }
}

function scheduleLookReport(): void {
  if (!userStore.isLogin || !articleID.value) return
  if (lookTimer) clearTimeout(lookTimer)
  lookTimer = setTimeout(() => {
    void reportArticleLook({ articleID: articleID.value, timeSecond: 15 })
  }, 15000)
}

function onScroll(): void {
  const doc = document.documentElement
  const total = doc.scrollHeight - doc.clientHeight
  progress.value = total > 0 ? Math.min(100, Math.round((doc.scrollTop / total) * 100)) : 0
}

function scrollToHeading(id: string): void {
  const target = document.getElementById(id)
  if (target) target.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function onDigg(): Promise<void> {
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录后再点赞')
    return
  }
  try {
    await diggArticle(articleID.value)
    digged.value = !digged.value
    if (article.value) {
      article.value.diggCount = Math.max(0, article.value.diggCount + (digged.value ? 1 : -1))
    }
    ElMessage.success(digged.value ? '点赞成功' : '已取消点赞')
  } catch {
    // ignore
  }
}

async function openCollect(): Promise<void> {
  collectVisible.value = true
  if (!userStore.isLogin || !userStore.userId) return
  try {
    const data = await fetchCollectFolders({ id: userStore.userId, page: 1, limit: 40 })
    folders.value = data?.list ?? []
  } catch {
    folders.value = []
  }
}

async function submitCollect(): Promise<void> {
  if (!userStore.isLogin) {
    collectVisible.value = false
    ElMessage.warning('请先登录后再收藏')
    return
  }
  collecting.value = true
  try {
    await collectArticle({ articleID: articleID.value, collectID: collectID.value })
    ElMessage.success('操作成功')
    collectVisible.value = false
    await loadArticle()
  } catch {
    // ignore
  } finally {
    collecting.value = false
  }
}

async function copyLink(): Promise<void> {
  try {
    await copyToClipboard(window.location.href)
    ElMessage.success('链接已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败,请手动复制地址栏链接')
  }
}

function goEdit(): void {
  router.push({ name: 'article-edit', params: { id: articleID.value } })
}

async function topThis(): Promise<void> {
  try {
    if (userTop.value) {
      await cancelTopArticle(articleID.value)
      userTop.value = false
      ElMessage.success('已取消置顶')
    } else {
      await topArticle(articleID.value)
      userTop.value = true
      ElMessage.success('置顶成功')
    }
    await loadArticle()
  } catch {
    // ignore
  }
}

async function removeThis(): Promise<void> {
  try {
    await ElMessageBox.confirm('删除后不可恢复,确认删除该文章?', '删除文章', { type: 'warning' })
  } catch {
    return
  }
  try {
    if (userStore.isAdmin && !isOwner.value) {
      await (await import('@/api/article')).adminRemoveArticles([articleID.value])
    } else {
      await removeArticles([articleID.value])
    }
    ElMessage.success('文章已删除')
    router.replace({ name: 'articles' })
  } catch {
    // ignore
  }
}

function goAuthorHome(): void {
  if (article.value) router.push({ name: 'user-home', params: { id: article.value.userID } })
}

function goAuthorArticles(): void {
  if (article.value) router.push({ name: 'user-home', params: { id: article.value.userID }, query: { tab: 'articles' } })
}

watch(collectVisible, (visible) => {
  if (visible) void openCollect()
})

watch(articleID, () => {
  digged.value = false
  collectID.value = 0
  void loadArticle()
})

onMounted(() => {
  void loadArticle()
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
  if (lookTimer) clearTimeout(lookTimer)
})
</script>

<style scoped lang="scss">
.detail-view {
  position: relative;
  padding: 22px 0 48px;
}

.detail-view__progress {
  position: fixed;
  top: 62px;
  left: 0;
  height: 2px;
  z-index: 40;
  background: linear-gradient(90deg, #22d3ee, #a855f7);
  box-shadow: 0 0 12px rgba(34, 211, 238, 0.9);
  transition: width 0.1s linear;
}

.detail-view__layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 20px;
  align-items: start;
  min-height: 300px;
}

.detail-view__status {
  margin-bottom: 0;
}

.detail-view__article {
  padding: 26px 30px 22px;
}

.detail-view__title {
  font-size: 28px;
  line-height: 1.4;
  margin-bottom: 16px;
}

.detail-view__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.detail-view__author {
  display: flex;
  align-items: center;
  gap: 10px;
}

.detail-view__author-info {
  display: flex;
  flex-direction: column;
}

.detail-view__author-name {
  font-size: 14px;
  font-weight: 600;
}

.detail-view__author-time {
  font-size: 12px;
}

.detail-view__stats {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--sd-text-dim);

  span {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
}

.detail-view__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.detail-view__cover {
  display: block;
  width: 100%;
  max-height: 380px;
  margin-top: 18px;
  border-radius: 12px;
  overflow: hidden;
}

.detail-view__ai {
  margin-top: 18px;
  padding: 14px 16px;
  border-radius: 12px;
  border: 1px solid rgba(168, 85, 247, 0.35);
  background: linear-gradient(120deg, rgba(168, 85, 247, 0.12), rgba(34, 211, 238, 0.08));
  font-size: 13px;
  color: #ddd6fe;
}

.detail-view__ai-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-weight: 600;
  color: #e9d5ff;
}

.detail-view__content {
  margin-top: 22px;
  font-size: 15px;
  line-height: 1.9;
  color: #d3dced;
  word-break: break-word;
}

.detail-view__content :deep(h1),
.detail-view__content :deep(h2),
.detail-view__content :deep(h3),
.detail-view__content :deep(h4) {
  margin: 26px 0 12px;
  padding-left: 12px;
  border-left: 3px solid var(--sd-cyan);
  scroll-margin-top: 86px;
}

.detail-view__content :deep(p) {
  margin: 12px 0;
}

.detail-view__content :deep(a) {
  color: var(--sd-cyan);
}

.detail-view__content :deep(img) {
  display: block;
  max-width: 100%;
  margin: 16px auto;
  border-radius: 12px;
}

.detail-view__content :deep(blockquote) {
  margin: 14px 0;
  padding: 10px 16px;
  border-left: 3px solid rgba(168, 85, 247, 0.6);
  background: rgba(30, 43, 69, 0.4);
  border-radius: 0 10px 10px 0;
  color: var(--sd-text-muted);
}

.detail-view__content :deep(pre) {
  padding: 14px 16px;
  overflow-x: auto;
  border-radius: 12px;
  border: 1px solid var(--sd-border);
  background: rgba(7, 11, 20, 0.9);
}

.detail-view__content :deep(code) {
  font-family: var(--sd-font-mono);
  font-size: 13px;
  color: #7dd3fc;
}

.detail-view__content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 14px 0;
}

.detail-view__content :deep(th),
.detail-view__content :deep(td) {
  padding: 8px 12px;
  border: 1px solid var(--sd-border);
}

.detail-view__content :deep(hr) {
  margin: 22px 0;
  border: none;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--sd-border-strong), transparent);
}

.detail-view__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 28px;
  padding-top: 18px;
  border-top: 1px solid var(--sd-border);
}

.detail-view__action {
  background: rgba(30, 43, 69, 0.5);
  border-color: var(--sd-border);
  color: var(--sd-text-muted);
}

.detail-view__action:hover {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.6);
}

.detail-view__action.is-active {
  color: #041016;
  background: linear-gradient(92deg, #22d3ee, #818cf8);
  border-color: transparent;
}

.detail-view__action.is-danger:hover {
  color: var(--sd-red);
  border-color: rgba(248, 113, 113, 0.6);
}

.detail-view__aside {
  display: flex;
  flex-direction: column;
  gap: 18px;
  position: sticky;
  top: 82px;
}

.detail-view__toc-item {
  display: block;
  padding: 5px 8px;
  font-size: 13px;
  color: var(--sd-text-muted);
  border-radius: 6px;
  border-left: 2px solid transparent;
  transition: all 0.2s ease;
}

.detail-view__toc-item:hover {
  color: var(--sd-cyan);
  background: rgba(34, 211, 238, 0.08);
  border-left-color: var(--sd-cyan);
}

.detail-view__toc-item.is-level-2 {
  padding-left: 18px;
  font-size: 12.5px;
}

.detail-view__toc-item.is-level-3 {
  padding-left: 28px;
  font-size: 12px;
}

.detail-view__author-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-align: center;
}

.detail-view__author-card-name {
  font-size: 15px;
  font-weight: 600;
}

.detail-view__author-actions {
  display: flex;
  gap: 8px;
}

.collect-select {
  width: 100%;
}

.collect-tip {
  margin-top: 10px;
  font-size: 12px;
}

@media (max-width: 1100px) {
  .detail-view__layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .detail-view__aside {
    position: static;
  }
}

@media (max-width: 640px) {
  .detail-view__article {
    padding: 18px 16px;
  }

  .detail-view__title {
    font-size: 21px;
  }
}
</style>
