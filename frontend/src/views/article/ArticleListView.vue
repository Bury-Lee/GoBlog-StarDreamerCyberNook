<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="page-head sd-panel">
        <div class="page-head__text">
          <h1 class="page-head__title">全部文章</h1>
          <p class="sd-dim">共 {{ count }} 篇已发布文章</p>
        </div>
        <div class="page-head__filters">
          <el-select v-model="order" class="page-head__select" @change="onOrderChange">
            <el-option
              v-for="item in ARTICLE_ORDER_OPTIONS"
              :key="item.label"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-input
            v-model="key"
            class="page-head__search"
            placeholder="搜索标题 / 摘要"
            clearable
            @keyup.enter="search"
            @clear="search"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <div v-loading="loading" class="article-grid">
        <EmptyState v-if="!loading && !list.length" text="没有找到符合条件的文章" />
        <ArticleCard v-for="article in list" :key="article.id" :article="article" />
      </div>

      <PaginationBar
        :page="page"
        :limit="limit"
        :count="count"
        :page-sizes="[12, 24, 36, 40]"
        @update:page="changePage"
        @update:limit="changeLimit"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import ArticleCard from '@/components/article/ArticleCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchArticleList } from '@/api/article'
import { ARTICLE_ORDER_OPTIONS } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const order = ref('')

const { list, count, page, limit, key, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchArticleList(params),
  () => ({ type: 'other' as const, order: order.value || undefined }),
  { limit: 12 },
)

function onOrderChange(): void {
  page.value = 1
  void load()
}
</script>

<style scoped lang="scss">
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  padding: 18px 20px;
  margin-bottom: 20px;
}

.page-head__title {
  font-size: 20px;
}

.page-head__filters {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.page-head__select {
  width: 150px;
}

.page-head__search {
  width: 240px;
}

.article-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 18px;
  min-height: 200px;
}

@media (max-width: 640px) {
  .page-head__search,
  .page-head__select {
    width: 100%;
  }
}
</style>
