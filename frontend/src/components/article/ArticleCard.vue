<template>
  <article
    class="article-card sd-panel sd-panel--hover"
    :class="[`article-card--${layout}`]"
    @click="goDetail"
  >
    <div class="article-card__cover">
      <img v-if="cover" :src="cover" :alt="article.title" loading="lazy" @error="coverFailed = true" />
      <div v-else class="article-card__cover-fallback">
        <span class="sd-neon-text">{{ titleInitial }}</span>
      </div>
      <span v-if="article.adminTop" class="article-card__top article-card__top--admin">管理员置顶</span>
      <span v-else-if="article.userTop" class="article-card__top">置顶</span>
    </div>

    <div class="article-card__body">
      <h3 class="article-card__title" v-html="titleHtml" />
      <p class="article-card__abstract sd-clamp-2" v-html="abstractHtml" />

      <div class="article-card__tags">
        <span v-if="article.categoryTitle" class="sd-tag sd-tag--purple">{{ article.categoryTitle }}</span>
        <span v-for="tag in tags" :key="tag" class="sd-tag" @click.stop="goTag(tag)"># {{ tag }}</span>
      </div>

      <div class="article-card__meta">
        <div class="article-card__author">
          <UserAvatar :src="authorAvatar" :name="authorName" :size="24" />
          <span class="sd-ellipsis">{{ authorName }}</span>
        </div>
        <div class="article-card__stats">
          <span><el-icon><View /></el-icon>{{ formatNumber(article.lookCount) }}</span>
          <span><el-icon><Pointer /></el-icon>{{ formatNumber(article.diggCount) }}</span>
          <span><el-icon><ChatDotRound /></el-icon>{{ formatNumber(article.commentCount) }}</span>
          <span><el-icon><Star /></el-icon>{{ formatNumber(article.collectCount) }}</span>
          <span class="article-card__date sd-dim">{{ fromNow(article.createdAt) }}</span>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ChatDotRound, Pointer, Star, View } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { resolveAssetUrl } from '@/api/request'
import type { ArticleModel } from '@/api/types'
import { fromNow, formatNumber } from '@/utils/format'
import { sanitizeHtml } from '@/utils/html'

export interface ArticleCardData extends ArticleModel {
  userTop?: boolean
  adminTop?: boolean
  categoryTitle?: string | null
  userNickName?: string
  userNickname?: string
  avatar?: string
  userAvatar?: string
}

const props = withDefaults(
  defineProps<{
    article: ArticleCardData
    layout?: 'grid' | 'list'
  }>(),
  { layout: 'grid' },
)

const router = useRouter()
const coverFailed = ref(false)

watch(
  () => props.article.cover,
  () => {
    coverFailed.value = false
  },
)

const cover = computed(() => {
  if (coverFailed.value) return ''
  return resolveAssetUrl(props.article.cover || '')
})

const titleHtml = computed(() => sanitizeHtml(props.article.title))
const abstractHtml = computed(() => sanitizeHtml(props.article.abstract || ''))
const tags = computed(() => (props.article.tagList || []).slice(0, 4))
const authorName = computed(
  () => props.article.userNickName || props.article.userNickname || '匿名用户',
)
const authorAvatar = computed(() => props.article.avatar || props.article.userAvatar || '')
const titleInitial = computed(() => (props.article.title || '文').slice(0, 1))

function goDetail(): void {
  router.push({ name: 'article-detail', params: { id: props.article.id } })
}

function goTag(tag: string): void {
  router.push({ name: 'search', query: { tag } })
}
</script>

<style scoped lang="scss">
.article-card {
  display: flex;
  overflow: hidden;
  cursor: pointer;
}

.article-card--grid {
  flex-direction: column;
}

.article-card--list {
  flex-direction: row;
}

.article-card__cover {
  position: relative;
  flex-shrink: 0;
  overflow: hidden;
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.18), rgba(168, 85, 247, 0.18));

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.4s ease;
  }
}

.article-card:hover .article-card__cover img {
  transform: scale(1.05);
}

.article-card--grid .article-card__cover {
  height: 168px;
}

.article-card--list .article-card__cover {
  width: 220px;
  min-height: 132px;
}

.article-card__cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  min-height: 120px;
  font-size: 40px;
  font-weight: 700;
  opacity: 0.85;
}

.article-card__top {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 2px 10px;
  font-size: 12px;
  border-radius: 999px;
  color: #041016;
  background: linear-gradient(92deg, #22d3ee, #818cf8);
  box-shadow: 0 4px 14px -6px rgba(34, 211, 238, 0.95);
}

.article-card__top--admin {
  background: linear-gradient(92deg, #f472b6, #a855f7);
  color: #fff;
}

.article-card__body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px 16px;
  flex: 1;
  min-width: 0;
}

.article-card__title {
  font-size: 16px;
  line-height: 1.5;
  transition: color 0.2s ease;

  :deep(em) {
    font-style: normal;
    color: var(--sd-amber);
    background: rgba(251, 191, 36, 0.12);
    padding: 0 2px;
    border-radius: 3px;
  }
}

.article-card:hover .article-card__title {
  color: var(--sd-cyan);
}

.article-card__abstract {
  font-size: 13px;
  color: var(--sd-text-muted);
  line-height: 1.7;

  :deep(em) {
    font-style: normal;
    color: var(--sd-amber);
  }
}

.article-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.article-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: auto;
  padding-top: 8px;
  font-size: 12px;
  color: var(--sd-text-dim);
  border-top: 1px dashed rgba(30, 43, 69, 0.8);
}

.article-card__author {
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 40%;
  color: var(--sd-text-muted);
}

.article-card__stats {
  display: flex;
  align-items: center;
  gap: 10px;

  span {
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
}

.article-card__date {
  display: none;
}

@media (min-width: 1500px) {
  .article-card__date {
    display: inline;
  }
}

@media (max-width: 640px) {
  .article-card--list {
    flex-direction: column;
  }

  .article-card--list .article-card__cover {
    width: 100%;
    height: 150px;
  }
}
</style>
