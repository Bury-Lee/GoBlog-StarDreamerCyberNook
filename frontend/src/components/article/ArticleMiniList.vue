<template>
  <section class="mini-list sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">{{ title }}</span>
      <router-link v-if="moreTo" class="sd-link mini-list__more" :to="moreTo">更多</router-link>
    </header>
    <div class="sd-panel__body mini-list__body">
      <el-skeleton v-if="loading" :rows="4" animated />
      <EmptyState v-else-if="!list.length" text="暂无内容" compact />
      <ul v-else class="mini-list__items">
        <li v-for="(item, index) in list" :key="item.id" class="mini-list__item" @click="goDetail(item.id)">
          <span class="mini-list__index" :class="{ 'is-top': index < 3 }">{{ index + 1 }}</span>
          <span class="mini-list__title sd-clamp-2" v-html="sanitizeHtml(item.title)" />
          <span class="mini-list__count sd-dim">{{ formatNumber(countOf(item)) }}</span>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, type RouteLocationRaw } from 'vue-router'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchArticleList, searchArticles } from '@/api/article'
import type { ArticleListQuery, ArticleListResponse, ArticleSearchListResponse } from '@/api/types'
import { formatNumber } from '@/utils/format'
import { sanitizeHtml } from '@/utils/html'

const props = withDefaults(
  defineProps<{
    title: string
    params: Record<string, unknown>
    countField?: 'lookCount' | 'diggCount' | 'commentCount' | 'collectCount'
    limit?: number
    moreTo?: RouteLocationRaw
  }>(),
  {
    countField: 'lookCount',
    limit: 6,
    moreTo: undefined,
  },
)

const router = useRouter()
const list = ref<Array<ArticleListResponse | ArticleSearchListResponse>>([])
const loading = ref(true)

function countOf(item: ArticleListResponse | ArticleSearchListResponse): number {
  return Number((item as unknown as Record<string, number>)[props.countField] || 0)
}

function goDetail(id: number): void {
  router.push({ name: 'article-detail', params: { id } })
}

async function load(): Promise<void> {
  loading.value = true
  try {
    const base = { page: 1, limit: props.limit, ...props.params } as Record<string, unknown>
    if (base.type === 'other') {
      const data = await fetchArticleList(base as unknown as ArticleListQuery)
      list.value = data?.list ?? []
    } else {
      const data = await searchArticles(base as never)
      list.value = data?.list ?? []
    }
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.mini-list__body {
  padding: 8px 12px 12px;
}

.mini-list__more {
  font-size: 12px;
}

.mini-list__items {
  display: flex;
  flex-direction: column;
}

.mini-list__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 4px;
  cursor: pointer;
  border-bottom: 1px dashed rgba(30, 43, 69, 0.7);
  transition: all 0.2s ease;
}

.mini-list__item:last-child {
  border-bottom: none;
}

.mini-list__item:hover {
  background: rgba(34, 211, 238, 0.06);
}

.mini-list__item:hover .mini-list__title {
  color: var(--sd-cyan);
}

.mini-list__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  font-size: 11px;
  border-radius: 6px;
  color: var(--sd-text-dim);
  background: rgba(30, 43, 69, 0.6);
}

.mini-list__index.is-top {
  color: #041016;
  background: linear-gradient(135deg, #22d3ee, #818cf8);
}

.mini-list__title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  line-height: 1.55;

  :deep(em) {
    font-style: normal;
    color: var(--sd-amber);
  }
}

.mini-list__count {
  font-size: 12px;
  flex-shrink: 0;
}
</style>
