<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="sd-panel history-panel">
        <header class="history-panel__head">
          <div>
            <h1 class="history-panel__title">浏览记录</h1>
            <p class="sd-dim">最多保留最近 30 天的浏览历史</p>
          </div>
          <div class="history-panel__tools">
            <el-input
              v-model="key"
              class="history-panel__search"
              placeholder="搜索标题"
              clearable
              @keyup.enter="search"
              @clear="search"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="danger" plain :disabled="!selection.length" @click="removeSelected">
              删除选中({{ selection.length }})
            </el-button>
          </div>
        </header>

        <el-table
          v-loading="loading"
          :data="list"
          border
          stripe
          @selection-change="onSelectionChange"
        >
          <el-table-column type="selection" width="46" />
          <el-table-column label="文章" min-width="260">
            <template #default="{ row }">
              <div class="history-row" @click="goArticle(row.articleID)">
                <img v-if="row.cover" :src="resolveAssetUrl(row.cover)" class="history-row__cover" alt="cover" />
                <div class="history-row__text">
                  <span class="history-row__title">{{ row.title }}</span>
                  <span class="sd-dim history-row__meta">
                    {{ row.nickname }} · {{ formatDate(row.lookDate) }}
                  </span>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="goArticle(row.articleID)">查看</el-button>
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchArticleHistory, removeArticleHistory } from '@/api/article'
import { resolveAssetUrl } from '@/api/request'
import type { ArticleHistoryItem } from '@/api/types'
import { formatDate } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()
const selection = ref<ArticleHistoryItem[]>([])

const { list, count, page, limit, key, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchArticleHistory(params),
  () => ({ userID: 0 }),
  { limit: 10 },
)

function onSelectionChange(rows: ArticleHistoryItem[]): void {
  selection.value = rows
}

function goArticle(id: number): void {
  router.push({ name: 'article-detail', params: { id } })
}

async function removeOne(row: ArticleHistoryItem): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除「${row.title}」这条浏览记录吗?`, '删除记录', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeArticleHistory([row.id])
    ElMessage.success('删除成功')
    await load()
  } catch {
    // ignore
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 条记录?`, '删除记录', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeArticleHistory(selection.value.map((item) => item.id))
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}
</script>

<style scoped lang="scss">
.history-panel {
  padding: 18px;
}

.history-panel__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.history-panel__title {
  font-size: 19px;
}

.history-panel__tools {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.history-panel__search {
  width: 220px;
}

.history-row {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.history-row__cover {
  width: 68px;
  height: 44px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
}

.history-row__text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.history-row__title {
  font-size: 14px;
}

.history-row:hover .history-row__title {
  color: var(--sd-cyan);
}

.history-row__meta {
  font-size: 12px;
}
</style>
