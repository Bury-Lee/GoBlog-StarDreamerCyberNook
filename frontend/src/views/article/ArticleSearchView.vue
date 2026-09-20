<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="search-panel sd-panel">
        <div class="search-panel__row">
          <el-input
            v-model="keyInput"
            size="large"
            placeholder="输入关键词搜索文章标题、摘要、正文…"
            clearable
            @keyup.enter="applySearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-input
            v-model="tagInput"
            size="large"
            class="search-panel__tag"
            placeholder="标签(可选)"
            clearable
            @keyup.enter="applySearch"
          >
            <template #prefix>
              <el-icon><CollectionTag /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" size="large" @click="applySearch">搜索</el-button>
        </div>
        <div class="search-panel__sorts">
          <span class="sd-dim">排序:</span>
          <span
            v-for="item in SEARCH_SORT_OPTIONS"
            :key="item.value"
            class="search-panel__sort"
            :class="{ 'is-active': sortType === item.value }"
            @click="changeSort(item.value)"
          >
            {{ item.label }}
          </span>
        </div>
      </section>

      <div class="search-result__head">
        <span class="sd-dim">
          共找到 <b class="sd-neon-text">{{ count }}</b> 条结果
        </span>
        <span v-if="key || tag" class="sd-chip">
          {{ key ? `关键词:${key}` : '' }}{{ key && tag ? ' · ' : '' }}{{ tag ? `标签:${tag}` : '' }}
        </span>
      </div>

      <div v-loading="loading" class="article-grid">
        <EmptyState v-if="!loading && !list.length" text="没有匹配的文章,换个关键词试试" />
        <ArticleCard v-for="article in list" :key="article.id" :article="article" layout="list" />
      </div>

      <PaginationBar
        :page="page"
        :limit="limit"
        :count="count"
        :page-sizes="[10, 20, 30, 40]"
        @update:page="changePage"
        @update:limit="changeLimit"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CollectionTag, Search } from '@element-plus/icons-vue'
import ArticleCard from '@/components/article/ArticleCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { searchArticles } from '@/api/article'
import { SEARCH_SORT_OPTIONS } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const route = useRoute()
const router = useRouter()

const key = ref(String(route.query.key || ''))
const tag = ref(String(route.query.tag || ''))
const sortType = ref(Number(route.query.type ?? 0))
const keyInput = ref(key.value)
const tagInput = ref(tag.value)

const { list, count, page, limit, loading, load, changePage, changeLimit } = usePagination(
  (params) => searchArticles(params),
  () => ({
    key: key.value || undefined,
    tag: tag.value || undefined,
    type: sortType.value,
  }),
  { limit: 10 },
)

function applySearch(): void {
  key.value = keyInput.value.trim()
  tag.value = tagInput.value.trim()
  page.value = 1
  void router.replace({
    name: 'search',
    query: {
      ...(key.value ? { key: key.value } : {}),
      ...(tag.value ? { tag: tag.value } : {}),
      ...(sortType.value ? { type: String(sortType.value) } : {}),
    },
  })
  void load()
}

function changeSort(value: number): void {
  sortType.value = value
  page.value = 1
  void load()
}

watch(
  () => route.query,
  (query) => {
    const nextKey = String(query.key || '')
    const nextTag = String(query.tag || '')
    const nextType = Number(query.type ?? 0)
    if (nextKey === key.value && nextTag === tag.value && nextType === sortType.value) return
    key.value = nextKey
    tag.value = nextTag
    sortType.value = nextType
    keyInput.value = nextKey
    tagInput.value = nextTag
    page.value = 1
    void load()
  },
)

onMounted(() => {
  void load()
})
</script>

<style scoped lang="scss">
.search-panel {
  padding: 18px;
  margin-bottom: 18px;
}

.search-panel__row {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.search-panel__tag {
  width: 220px;
}

.search-panel__sorts {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
  margin-top: 14px;
  font-size: 13px;
}

.search-panel__sort {
  color: var(--sd-text-muted);
  cursor: pointer;
  transition: color 0.2s ease;
}

.search-panel__sort:hover,
.search-panel__sort.is-active {
  color: var(--sd-cyan);
  text-shadow: 0 0 12px rgba(34, 211, 238, 0.6);
}

.search-result__head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  font-size: 13px;
}

.article-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 14px;
  min-height: 180px;
}

@media (max-width: 720px) {
  .search-panel__tag {
    width: 100%;
  }
}
</style>
