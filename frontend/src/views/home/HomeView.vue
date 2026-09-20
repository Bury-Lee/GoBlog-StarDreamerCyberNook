<template>
  <div class="home-view">
    <section class="home-hero sd-grid-bg">
      <div class="sd-container home-hero__inner">
        <div class="home-hero__text">
          <span class="sd-chip home-hero__chip">
            <span class="home-hero__dot" />
            SIGNAL ONLINE · 服务已连接
          </span>
          <h1 class="home-hero__title sd-neon-text">{{ siteStore.siteTitle }}</h1>
          <p class="home-hero__desc">
            {{ siteStore.seo.description || '技术分享 · 内容社区 · 记录每一次思考' }}
          </p>
          <div class="home-hero__search">
            <el-input
              v-model="keyword"
              size="large"
              placeholder="搜索文章标题、摘要或标签…"
              clearable
              @keyup.enter="goSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" size="large" @click="goSearch">搜索</el-button>
          </div>
          <div v-if="hotTags.length" class="home-hero__tags">
            <span class="sd-dim">热门标签:</span>
            <span v-for="tag in hotTags" :key="tag" class="sd-tag" @click="goTag(tag)"># {{ tag }}</span>
          </div>
        </div>
        <div class="home-hero__deco">
          <div class="home-hero__orb" />
          <div class="home-hero__ring" />
        </div>
      </div>
    </section>

    <div class="sd-container home-view__body">
      <el-carousel
        v-if="banners.length"
        class="home-view__banner"
        height="240px"
        :interval="5000"
        arrow="hover"
      >
        <el-carousel-item v-for="banner in banners" :key="banner.id">
          <a
            class="home-view__banner-item"
            :href="banner.href || 'javascript:void(0)'"
            :target="banner.href ? '_blank' : '_self'"
            rel="noopener noreferrer"
          >
            <img :src="resolveAssetUrl(banner.cover)" :alt="`banner-${banner.id}`" />
            <span class="home-view__banner-mask" />
          </a>
        </el-carousel-item>
      </el-carousel>

      <div class="sd-layout-2col">
        <div class="sd-section-gap">
          <section class="home-view__articles sd-panel">
            <header class="home-view__articles-head">
              <div class="home-view__tabs">
                <button
                  v-for="tab in orderTabs"
                  :key="tab.value"
                  class="home-view__tab"
                  :class="{ 'is-active': order === tab.value }"
                  @click="switchOrder(tab.value)"
                >
                  {{ tab.label }}
                </button>
              </div>
              <router-link class="sd-link home-view__more" :to="{ name: 'articles' }">
                全部文章
                <el-icon><ArrowRight /></el-icon>
              </router-link>
            </header>

            <div v-loading="loading" class="home-view__list">
              <EmptyState v-if="!loading && !list.length" text="还没有文章,快去发布第一篇吧" />
              <ArticleCard
                v-for="article in list"
                :key="article.id"
                :article="article"
                layout="list"
              />
            </div>

            <PaginationBar
              :page="page"
              :limit="limit"
              :count="count"
              layout="prev, pager, next"
              @update:page="changePage"
            />
          </section>
        </div>

        <aside class="sd-section-gap">
          <template v-for="item in sidebarItems" :key="item.title">
            <ArticleMiniList
              v-if="isHot(item.title)"
              :title="item.title"
              :params="{ type: 'other', order: 'look_count desc' }"
              count-field="lookCount"
              :more-to="{ name: 'articles' }"
            />
            <ArticleMiniList
              v-else-if="isComment(item.title)"
              title="热议文章"
              :params="{ type: 'other', order: 'comment_count desc' }"
              count-field="commentCount"
              :more-to="{ name: 'articles' }"
            />
            <FriendLinksPanel v-else-if="isFriend(item.title)" :title="item.title" />
            <TagCloud v-else-if="isTag(item.title)" />
            <ArticleMiniList
              v-else
              :title="item.title"
              :params="{ type: 'other' }"
              count-field="lookCount"
              :more-to="{ name: 'articles' }"
            />
          </template>
        </aside>
      </div>

      <PromotionPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowRight, Search } from '@element-plus/icons-vue'
import ArticleCard from '@/components/article/ArticleCard.vue'
import ArticleMiniList from '@/components/article/ArticleMiniList.vue'
import TagCloud from '@/components/article/TagCloud.vue'
import FriendLinksPanel from '@/components/common/FriendLinksPanel.vue'
import PromotionPanel from '@/components/common/PromotionPanel.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchArticleList } from '@/api/article'
import { fetchBanners } from '@/api/ops'
import { resolveAssetUrl } from '@/api/request'
import type { Banner } from '@/api/types'
import { usePagination } from '@/composables/usePagination'
import { useSiteStore } from '@/stores'

const router = useRouter()
const siteStore = useSiteStore()

const keyword = ref('')
const banners = ref<Banner[]>([])
const order = ref('')
const hotTags = ref<string[]>([])

const orderTabs = [
  { label: '最新发布', value: '' },
  { label: '最多浏览', value: 'look_count desc' },
  { label: '最多点赞', value: 'digg_count desc' },
  { label: '最多收藏', value: 'collect_count desc' },
]

const { list, count, page, limit, loading, load, changePage } = usePagination(
  (params) => fetchArticleList(params),
  () => ({ type: 'other' as const, order: order.value || undefined }),
  { limit: 6 },
)

const DEFAULT_SIDEBAR = [
  { title: '热门文章', enable: true },
  { title: '最新评论', enable: true },
  { title: '标签云', enable: true },
  { title: '友情链接', enable: true },
]

const sidebarItems = computed(() => {
  const enabled = siteStore.indexRightList.filter((item) => item.enable)
  return enabled.length ? enabled : DEFAULT_SIDEBAR
})

function isHot(title: string): boolean {
  return /热门|popular/i.test(title)
}

function isComment(title: string): boolean {
  return /评论|comment|热议/i.test(title)
}

function isFriend(title: string): boolean {
  return /友链|友情|friend|link/i.test(title)
}

function isTag(title: string): boolean {
  return /标签|tag/i.test(title)
}

function switchOrder(value: string): void {
  if (order.value === value) return
  order.value = value
  page.value = 1
  void load()
}

function goSearch(): void {
  router.push({ name: 'search', query: keyword.value.trim() ? { key: keyword.value.trim() } : {} })
}

function goTag(tag: string): void {
  router.push({ name: 'search', query: { tag } })
}

onMounted(async () => {
  try {
    const data = await fetchBanners({ page: 1, limit: 6 })
    banners.value = data?.list ?? []
  } catch {
    banners.value = []
  }
  try {
    const data = await fetchArticleList({ type: 'other', page: 1, limit: 30, order: 'look_count desc' })
    const tags = (data?.list ?? []).flatMap((item) => item.tagList || [])
    hotTags.value = Array.from(new Set(tags)).slice(0, 8)
  } catch {
    hotTags.value = []
  }
})
</script>

<style scoped lang="scss">
.home-view__body {
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 24px 0 48px;
}

.home-hero {
  position: relative;
  overflow: hidden;
  border-bottom: 1px solid var(--sd-border);
  background:
    radial-gradient(800px 320px at 15% 0%, rgba(34, 211, 238, 0.16), transparent 65%),
    radial-gradient(700px 300px at 85% 20%, rgba(168, 85, 247, 0.16), transparent 60%);
}

.home-hero__inner {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding-top: 54px;
  padding-bottom: 54px;
}

.home-hero__text {
  max-width: 620px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.home-hero__chip {
  width: fit-content;
  color: #a5f3fc;
}

.home-hero__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--sd-green);
  box-shadow: 0 0 10px var(--sd-green);
  animation: sd-pulse 1.8s ease-in-out infinite;
}

.home-hero__title {
  font-size: 40px;
  font-weight: 800;
  letter-spacing: 1px;
  line-height: 1.25;
}

.home-hero__desc {
  font-size: 15px;
  color: var(--sd-text-muted);
}

.home-hero__search {
  display: flex;
  gap: 10px;
  margin-top: 6px;
}

.home-hero__tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
}

.home-hero__tags .sd-tag {
  cursor: pointer;
}

.home-hero__deco {
  position: relative;
  width: 220px;
  height: 220px;
  flex-shrink: 0;
}

.home-hero__orb {
  position: absolute;
  inset: 30px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, rgba(34, 211, 238, 0.55), rgba(168, 85, 247, 0.35) 60%, transparent 70%);
  filter: blur(6px);
  animation: sd-float 5s ease-in-out infinite;
}

.home-hero__ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 1px dashed rgba(34, 211, 238, 0.45);
  animation: home-spin 22s linear infinite;
}

@keyframes home-spin {
  to {
    transform: rotate(360deg);
  }
}

.home-view__banner {
  border-radius: var(--sd-radius);
  overflow: hidden;
  border: 1px solid var(--sd-border);
}

.home-view__banner-item {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.home-view__banner-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 40%, rgba(7, 11, 20, 0.55));
}

.home-view__articles {
  padding: 16px 18px 20px;
}

.home-view__articles-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--sd-border);
}

.home-view__tabs {
  display: flex;
  gap: 6px;
}

.home-view__tab {
  padding: 6px 14px;
  font-size: 13px;
  color: var(--sd-text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.home-view__tab:hover {
  color: var(--sd-text);
  background: rgba(34, 211, 238, 0.08);
}

.home-view__tab.is-active {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.5);
  background: rgba(34, 211, 238, 0.1);
}

.home-view__more {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.home-view__list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 160px;
}

@media (max-width: 900px) {
  .home-hero__deco {
    display: none;
  }

  .home-hero__title {
    font-size: 30px;
  }
}
</style>
