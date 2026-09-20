<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="sd-panel my-articles">
        <header class="my-articles__head">
          <div class="my-articles__tabs">
            <button
              v-for="tab in statusTabs"
              :key="tab.value"
              class="my-articles__tab"
              :class="{ 'is-active': status === tab.value }"
              @click="switchStatus(tab.value)"
            >
              {{ tab.label }}
            </button>
          </div>
          <div class="my-articles__tools">
            <el-input
              v-model="key"
              class="my-articles__search"
              placeholder="搜索我的文章"
              clearable
              @keyup.enter="search"
              @clear="search"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" @click="goCreate">
              <el-icon><EditPen /></el-icon>
              写文章
            </el-button>
          </div>
        </header>

        <el-table v-loading="loading" :data="visibleList" border stripe>
          <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <router-link class="sd-link" :to="{ name: 'article-detail', params: { id: row.id } }">
                {{ row.title }}
              </router-link>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="articleStatusType(row.status)" size="small">
                {{ articleStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="数据" width="200">
            <template #default="{ row }">
              <span class="sd-dim my-articles__counts">
                浏览 {{ formatNumber(row.lookCount) }} · 赞 {{ formatNumber(row.diggCount) }} ·
                评 {{ formatNumber(row.commentCount) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="分类" width="120">
            <template #default="{ row }">
              <span class="sd-dim">{{ row.categoryID ? `#${row.categoryID}` : '未分类' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="更新时间" width="160">
            <template #default="{ row }">
              <span class="sd-dim">{{ formatDate(row.updatedAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="goEdit(row.id)">编辑</el-button>
              <el-button link type="warning" @click="toggleTop(row)">
                {{ row.userTop ? '取消置顶' : '置顶' }}
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { EditPen, Search } from '@element-plus/icons-vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { cancelTopArticle, fetchArticleList, removeArticles, topArticle } from '@/api/article'
import type { ArticleListResponse } from '@/api/types'
import { articleStatusLabel, articleStatusType, formatDate, formatNumber } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()

const statusTabs = [
  { label: '全部', value: -1 },
  { label: '草稿', value: 0 },
  { label: '审核中', value: 1 },
  { label: '已发布', value: 2 },
  { label: '已下线', value: 3 },
]

const status = ref(-1)

const { list, count, page, limit, key, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchArticleList(params),
  () => ({
    type: 'self' as const,
    status: status.value > 0 ? status.value : undefined,
  }),
  { limit: 10 },
)

const visibleList = computed(() => {
  if (status.value < 0) return list.value
  return (list.value as ArticleListResponse[]).filter((item) => item.status === status.value)
})

function switchStatus(value: number): void {
  status.value = value
  page.value = 1
  void load()
}

function goCreate(): void {
  router.push({ name: 'article-create' })
}

function goEdit(id: number): void {
  router.push({ name: 'article-edit', params: { id } })
}

async function toggleTop(row: ArticleListResponse): Promise<void> {
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
  }
}

async function removeOne(row: ArticleListResponse): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除《${row.title}》?`, '删除文章', { type: 'warning' })
  } catch {
    return
  }
  try {
    await removeArticles([row.id])
    ElMessage.success('删除成功')
    await load()
  } catch {
    // ignore
  }
}
</script>

<style scoped lang="scss">
.my-articles {
  padding: 18px;
}

.my-articles__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.my-articles__tabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.my-articles__tab {
  padding: 6px 14px;
  font-size: 13px;
  color: var(--sd-text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.my-articles__tab:hover {
  color: var(--sd-text);
  background: rgba(34, 211, 238, 0.08);
}

.my-articles__tab.is-active {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.5);
  background: rgba(34, 211, 238, 0.1);
}

.my-articles__tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.my-articles__search {
  width: 220px;
}

.my-articles__counts {
  font-size: 12px;
}
</style>
