<template>
  <div class="admin-articles">
    <section class="sd-panel admin-articles__filters">
      <div class="admin-articles__tabs">
        <button
          v-for="tab in statusTabs"
          :key="tab.value"
          class="admin-articles__tab"
          :class="{ 'is-active': status === tab.value }"
          @click="switchStatus(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="admin-articles__tools">
        <el-input
          v-model="key"
          class="admin-articles__search"
          placeholder="搜索标题 / 摘要"
          clearable
          @keyup.enter="search"
          @clear="search"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-input
          v-model="userID"
          class="admin-articles__search"
          type="number"
          min="1"
          placeholder="按作者 ID 筛选"
          clearable
          @keyup.enter="search"
          @clear="search"
        />
        <el-tag v-if="categoryID" closable type="info" @close="clearCategory">
          已按分类筛选
        </el-tag>
        <el-select v-model="order" class="admin-articles__select" @change="search">
          <el-option
            v-for="item in ARTICLE_ORDER_OPTIONS"
            :key="item.label"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-button type="danger" plain :disabled="!selection.length" @click="removeSelected">
          批量删除({{ selection.length }})
        </el-button>
      </div>
    </section>

    <section class="sd-panel admin-articles__table">
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="46" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="标题" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <router-link class="sd-link" :to="{ name: 'article-detail', params: { id: row.id } }">
              {{ row.title }}
            </router-link>
            <el-tag v-if="row.adminTop" size="small" type="danger" class="admin-articles__flag">管理置顶</el-tag>
            <el-tag v-else-if="row.userTop" size="small" type="warning" class="admin-articles__flag">置顶</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="作者" width="150">
          <template #default="{ row }">
            <div class="admin-articles__author">
              <UserAvatar :src="row.avatar" :name="row.userNickName" :size="26" />
              <span class="sd-ellipsis">{{ row.userNickName || '未知用户' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="articleStatusType(row.status)" size="small">
              {{ articleStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="数据" width="210">
          <template #default="{ row }">
            <span class="sd-dim admin-articles__counts">
              浏览 {{ formatNumber(row.lookCount) }} · 赞 {{ formatNumber(row.diggCount) }} ·
              评 {{ formatNumber(row.commentCount) }} · 藏 {{ formatNumber(row.collectCount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openReview(row)">审核</el-button>
            <el-button
              link
              type="warning"
              :loading="topLoading === row.id"
              @click="toggleTop(row)"
            >
              {{ row.userTop ? '取消置顶' : '置顶' }}
            </el-button>
            <el-button
              v-if="row.adminTop"
              link
              type="info"
              @click="cancelAdminTop(row)"
            >
              取消管理置顶
            </el-button>
            <el-button link type="danger" @click="removeOne(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <PaginationBar
        :page="page"
        :limit="limit"
        :count="count"
        @update:page="changePage"
        @update:limit="changeLimit"
      />
    </section>

    <el-dialog v-model="reviewVisible" title="文章审核" width="520px">
      <el-alert
        class="admin-articles__alert"
        type="info"
        :closable="false"
        show-icon
        title="通过后文章状态变为「已发布」,驳回后回到「草稿」,结果会以系统通知发送给作者"
      />
      <el-form label-width="80px">
        <el-form-item label="文章">
          <span>{{ reviewTarget?.title }}</span>
        </el-form-item>
        <el-form-item label="审核意见">
          <el-input v-model="reviewMsg" type="textarea" :rows="3" resize="none" placeholder="可填写审核说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button type="danger" :loading="reviewing" @click="submitReview(0)">驳回</el-button>
        <el-button type="primary" :loading="reviewing" @click="submitReview(2)">通过</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import {
  adminCancelTopArticle,
  adminRemoveArticles,
  cancelTopArticle,
  fetchArticleList,
  reviewArticle,
  topArticle,
} from '@/api/article'
import type { ArticleListResponse } from '@/api/types'
import {
  ARTICLE_ORDER_OPTIONS,
  articleStatusLabel,
  articleStatusType,
  formatDate,
  formatNumber,
} from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const route = useRoute()

const statusTabs = [
  { label: '全部', value: -1 },
  { label: '草稿', value: 0 },
  { label: '审核中', value: 1 },
  { label: '已发布', value: 2 },
  { label: '已下线', value: 3 },
]

const status = ref(-1)
const order = ref('')
const userID = ref(String(route.query.userID || ''))
const categoryID = ref<number | undefined>(
  route.query.categoryID ? Number(route.query.categoryID) : undefined,
)
const selection = ref<ArticleListResponse[]>([])
const reviewVisible = ref(false)
const reviewing = ref(false)
const reviewMsg = ref('')
const reviewTarget = ref<ArticleListResponse | null>(null)
const topLoading = ref(0)

const { list, count, page, limit, key, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchArticleList(params),
  () => {
    const uid = Number(userID.value)
    return {
      type: 'admin' as const,
      status: status.value >= 0 ? status.value : undefined,
      userID: Number.isFinite(uid) && uid > 0 ? uid : undefined,
      categoryID: categoryID.value,
      order: order.value || undefined,
    }
  },
  { limit: 10 },
)

function switchStatus(value: number): void {
  status.value = value
  page.value = 1
  void load()
}

function clearCategory(): void {
  categoryID.value = undefined
  page.value = 1
  void load()
}

function onSelectionChange(rows: ArticleListResponse[]): void {
  selection.value = rows
}

function openReview(row: ArticleListResponse): void {
  reviewTarget.value = row
  reviewMsg.value = ''
  reviewVisible.value = true
}

async function submitReview(nextStatus: number): Promise<void> {
  if (!reviewTarget.value) return
  reviewing.value = true
  try {
    await reviewArticle({
      articleID: reviewTarget.value.id,
      status: nextStatus,
      msg: reviewMsg.value || (nextStatus === 2 ? '审核通过' : '审核未通过'),
    })
    ElMessage.success(nextStatus === 2 ? '已通过审核' : '已驳回')
    reviewVisible.value = false
    await load()
  } catch {
    // ignore
  } finally {
    reviewing.value = false
  }
}

async function toggleTop(row: ArticleListResponse): Promise<void> {
  if (row.userTop) {
    try {
      await ElMessageBox.confirm('取消置顶后该文章将回到正常排序,确定吗?', '取消置顶', {
        type: 'warning',
      })
    } catch {
      return
    }
  }
  topLoading.value = row.id
  try {
    if (row.userTop) {
      await cancelTopArticle(row.id)
      ElMessage.success('已取消置顶')
    } else {
      await topArticle(row.id)
      ElMessage.success('置顶成功')
    }
    await load()
  } catch {
    // ignore
  } finally {
    topLoading.value = 0
  }
}

async function cancelAdminTop(row: ArticleListResponse): Promise<void> {
  try {
    await adminCancelTopArticle(row.id, row.userID)
    ElMessage.success('已取消管理员置顶')
    await load()
  } catch {
    // ignore
  }
}

async function removeOne(row: ArticleListResponse): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除《${row.title}》?该操作不可恢复`, '删除文章', { type: 'warning' })
  } catch {
    return
  }
  try {
    await adminRemoveArticles([row.id])
    ElMessage.success('删除成功')
    await load()
  } catch {
    // ignore
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 篇文章?`, '批量删除', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await adminRemoveArticles(selection.value.map((item) => item.id))
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.admin-articles {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.admin-articles__filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  padding: 16px;
}

.admin-articles__tabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.admin-articles__tab {
  padding: 6px 14px;
  font-size: 13px;
  color: var(--sd-text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.admin-articles__tab.is-active {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.5);
  background: rgba(34, 211, 238, 0.1);
}

.admin-articles__tools {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.admin-articles__search {
  width: 180px;
}

.admin-articles__select {
  width: 140px;
}

.admin-articles__table {
  padding: 18px;
}

.admin-articles__flag {
  margin-left: 8px;
}

.admin-articles__author {
  display: flex;
  align-items: center;
  gap: 8px;
}

.admin-articles__counts {
  font-size: 12px;
}

.admin-articles__alert {
  margin-bottom: 14px;
}
</style>
