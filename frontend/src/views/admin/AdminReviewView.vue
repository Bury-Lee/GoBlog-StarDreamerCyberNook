<template>
  <div class="admin-review">
    <section class="sd-panel admin-review__panel">
      <header class="admin-review__head">
        <div>
          <h3 class="admin-review__title">待审核文章</h3>
          <p class="sd-dim admin-review__desc">
            列表返回状态为「审核中」的文章,通过后状态变为已发布,驳回则退回草稿并发送系统通知
          </p>
        </div>
        <div class="admin-review__tools">
          <el-input
            v-model="userID"
            class="admin-review__search"
            type="number"
            min="1"
            placeholder="按作者 ID 筛选"
            clearable
            @keyup.enter="search"
            @clear="search"
          />
          <el-button type="warning" plain :loading="aiReviewing" @click="aiReviewAll">
            <el-icon><MagicStick /></el-icon>
            AI 批量审核
          </el-button>
          <el-button @click="load">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </header>

      <div v-loading="loading" class="admin-review__list">
        <EmptyState v-if="!loading && !list.length" text="没有待审核的文章,清爽~" />
        <article v-for="item in list" :key="item.id" class="review-card">
          <div class="review-card__main">
            <div class="review-card__head">
              <span class="review-card__title">{{ item.title }}</span>
              <el-tag size="small" :type="articleStatusType(item.status)">
                {{ articleStatusLabel(item.status) }}
              </el-tag>
            </div>
            <p class="review-card__abstract sd-clamp-2">{{ item.abstract || excerpt(item.content, 160) }}</p>
            <div class="review-card__meta">
              <span class="review-card__author sd-dim">
                <UserAvatar :src="authors[item.userID]?.avatar" :name="authors[item.userID]?.nickName" :size="22" />
                {{ authors[item.userID]?.nickName || '未知用户' }}
              </span>
              <span class="sd-dim">创建:{{ formatDate(item.createdAt) }}</span>
              <span class="sd-dim">更新:{{ formatDate(item.updatedAt) }}</span>
              <span v-for="tag in item.tagList || []" :key="tag" class="sd-tag"># {{ tag }}</span>
            </div>
          </div>
          <div class="review-card__actions">
            <el-button size="small" @click="preview(item)">预览</el-button>
            <el-button
              size="small"
              type="warning"
              plain
              :loading="aiReviewingId === item.id"
              @click="aiReviewOne(item)"
            >
              AI 审核
            </el-button>
            <el-button size="small" type="danger" plain @click="openReview(item, 0)">驳回</el-button>
            <el-button size="small" type="primary" @click="openReview(item, 2)">通过</el-button>
          </div>
        </article>
      </div>

      <PaginationBar
        :page="page"
        :limit="limit"
        :count="count"
        @update:page="changePage"
        @update:limit="changeLimit"
      />
    </section>

    <el-drawer v-model="previewVisible" title="文章预览" size="620px">
      <div v-if="previewLoading" v-loading="true" class="review-preview__loading" />
      <div v-else-if="previewArticle" class="review-preview">
        <h2 class="review-preview__title">{{ previewArticle.title }}</h2>
        <div class="review-preview__meta sd-dim">
          {{ previewArticle.nickname || previewArticle.username || '未知用户' }} ·
          {{ formatDate(previewArticle.createdAt) }}
        </div>
        <p v-if="previewArticle.articleAddition?.aiAbstract" class="review-preview__ai">
          AI 摘要:{{ previewArticle.articleAddition?.aiAbstract }}
        </p>
        <div class="article-content review-preview__content" v-html="previewHtml" />
      </div>
      <EmptyState v-else text="文章内容加载失败,请关闭后重试" compact />
    </el-drawer>

    <el-dialog v-model="reviewVisible" :title="reviewStatus === 2 ? '通过审核' : '驳回文章'" width="480px">
      <el-form label-width="80px">
        <el-form-item label="文章">
          <span>{{ reviewTarget?.title }}</span>
        </el-form-item>
        <el-form-item label="审核意见">
          <el-input
            v-model="reviewMsg"
            type="textarea"
            :rows="3"
            resize="none"
            :placeholder="reviewStatus === 2 ? '默认:审核通过' : '默认:审核未通过'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button :type="reviewStatus === 2 ? 'primary' : 'danger'" :loading="reviewing" @click="submitReview">
          确认{{ reviewStatus === 2 ? '通过' : '驳回' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, Refresh } from '@element-plus/icons-vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { aiReviewArticles, fetchArticleDetail, fetchReviewArticles, reviewArticle } from '@/api/article'
import { fetchUserBaseInfo } from '@/api/user'
import type { ArticleDetailResponse, ArticleModel, UserBaseInfo } from '@/api/types'
import {
  articleStatusLabel,
  articleStatusType,
  excerpt,
  formatDate,
} from '@/utils/format'
import { processArticleHtml } from '@/utils/html'
import { usePagination } from '@/composables/usePagination'

const userID = ref('')
const previewVisible = ref(false)
const previewArticle = ref<ArticleDetailResponse | null>(null)
const previewHtml = ref('')
const previewLoading = ref(false)
const reviewVisible = ref(false)
const reviewing = ref(false)
const aiReviewing = ref(false)
const aiReviewingId = ref(0)
const reviewMsg = ref('')
const reviewStatus = ref(2)
const reviewTarget = ref<ArticleModel | null>(null)
const authors = ref<Record<number, UserBaseInfo>>({})

const { list, count, page, limit, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchReviewArticles(params),
  () => {
    const uid = Number(userID.value)
    return { userID: Number.isFinite(uid) && uid > 0 ? uid : undefined }
  },
  { limit: 10 },
)

// 审核列表接口不带作者信息,按当前页补齐昵称与头像,失败时显示"未知用户"
async function loadAuthors(): Promise<void> {
  const ids = [...new Set(list.value.map((item) => item.userID))].filter(
    (id) => id > 0 && !authors.value[id],
  )
  if (!ids.length) return
  const results = await Promise.allSettled(ids.map((id) => fetchUserBaseInfo(id, { silent: true })))
  results.forEach((result, index) => {
    if (result.status === 'fulfilled' && result.value) {
      authors.value[ids[index]] = result.value
    }
  })
}

async function reload(): Promise<void> {
  await load()
  void loadAuthors()
}

watch(list, () => {
  void loadAuthors()
})

async function preview(item: ArticleModel): Promise<void> {
  previewVisible.value = true
  previewArticle.value = null
  previewHtml.value = ''
  previewLoading.value = true
  try {
    const data = await fetchArticleDetail(item.id)
    previewArticle.value = data
    previewHtml.value = processArticleHtml(data?.content).html
  } catch {
    previewArticle.value = null
  } finally {
    previewLoading.value = false
  }
}

function openReview(item: ArticleModel, status: number): void {
  reviewTarget.value = item
  reviewStatus.value = status
  reviewMsg.value = ''
  reviewVisible.value = true
}

async function submitReview(): Promise<void> {
  if (!reviewTarget.value) return
  reviewing.value = true
  try {
    await reviewArticle({
      articleID: reviewTarget.value.id,
      status: reviewStatus.value,
      msg: reviewMsg.value || (reviewStatus.value === 2 ? '审核通过' : '审核未通过'),
    })
    ElMessage.success(reviewStatus.value === 2 ? '已通过审核' : '已驳回')
    reviewVisible.value = false
    await load()
  } catch {
    // ignore
  } finally {
    reviewing.value = false
  }
}

async function aiReviewAll(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      'AI 批量审核会直接通过或驳回最多 10 篇文章,结果不可撤销,确定继续吗?',
      'AI 批量审核',
      { type: 'warning', confirmButtonText: '开始审核', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  aiReviewing.value = true
  try {
    const result = await aiReviewArticles({ limit: 10 })
    ElMessage.success(`AI 审核完成:共 ${result?.total ?? 0} 篇,成功 ${result?.count ?? 0} 篇`)
    await reload()
  } catch {
    // 错误提示已由请求层处理
  } finally {
    aiReviewing.value = false
  }
}

async function aiReviewOne(item: ArticleModel): Promise<void> {
  aiReviewingId.value = item.id
  try {
    const result = await aiReviewArticles({ articleID: item.id })
    const first = result?.list?.[0]
    if (first?.error) {
      console.warn('AI 审核失败:', first.error)
      ElMessage.error('AI 审核失败,请稍后重试或改为人工审核')
    } else {
      const text = (first?.aiResult || '已完成').replace(/\s+/g, ' ').slice(0, 60)
      ElMessage.success(`AI 审核完成:${text}`)
    }
    await reload()
  } catch {
    // 错误提示已由请求层处理
  } finally {
    aiReviewingId.value = 0
  }
}
</script>

<style scoped lang="scss">
.admin-review__panel {
  padding: 18px;
}

.admin-review__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.admin-review__title {
  font-size: 17px;
}

.admin-review__desc {
  margin-top: 4px;
  font-size: 12px;
  max-width: 620px;
}

.admin-review__tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-review__search {
  width: 180px;
}

.admin-review__list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 200px;
}

.review-card {
  display: flex;
  gap: 16px;
  padding: 16px;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.55);
  transition: border-color 0.2s ease;
}

.review-card:hover {
  border-color: rgba(34, 211, 238, 0.45);
}

.review-card__main {
  flex: 1;
  min-width: 0;
}

.review-card__head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.review-card__title {
  font-size: 15px;
  font-weight: 600;
}

.review-card__abstract {
  margin-top: 6px;
  font-size: 13px;
  color: var(--sd-text-muted);
}

.review-card__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 10px;
  font-size: 12px;
}

.review-card__author {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.review-preview__loading {
  min-height: 200px;
}

.review-card__actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  justify-content: center;
}

.review-preview__title {
  font-size: 20px;
  margin-bottom: 8px;
}

.review-preview__meta {
  font-size: 12px;
  margin-bottom: 12px;
}

.review-preview__ai {
  padding: 10px 12px;
  margin-bottom: 14px;
  border-radius: 10px;
  border: 1px solid rgba(168, 85, 247, 0.35);
  background: rgba(168, 85, 247, 0.1);
  font-size: 13px;
  color: #e9d5ff;
}

.review-preview__content {
  font-size: 14px;
  line-height: 1.9;
  color: #cbd5e1;
}

.review-preview__content :deep(img) {
  max-width: 100%;
  border-radius: 10px;
}

.review-preview__content :deep(pre) {
  padding: 12px;
  overflow-x: auto;
  border-radius: 10px;
  background: rgba(7, 11, 20, 0.9);
}
</style>
