<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="about-hero sd-panel sd-grid-bg">
        <h1 class="about-hero__title sd-neon-text">{{ siteStore.siteTitle }}</h1>
        <p v-if="siteStore.seo.description" class="about-hero__desc">{{ siteStore.seo.description }}</p>
        <div v-if="stats.length" class="about-hero__stats">
          <div v-for="item in stats" :key="item.label" class="about-hero__stat">
            <span class="about-hero__value sd-neon-text">{{ item.value }}</span>
            <span class="sd-dim">{{ item.label }}</span>
          </div>
        </div>
      </section>

      <div v-if="siteRows.length || contactRows.length" class="about-grid">
        <section v-if="siteRows.length" class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">站点信息</span>
          </header>
          <div class="sd-panel__body">
            <ul class="about-list">
              <li v-for="row in siteRows" :key="row.label">
                <span class="sd-dim">{{ row.label }}</span>
                <span>{{ row.value }}</span>
              </li>
            </ul>
          </div>
        </section>

        <section v-if="contactRows.length" class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">联系与关注</span>
          </header>
          <div class="sd-panel__body">
            <ul class="about-list">
              <li v-for="row in contactRows" :key="row.label">
                <span class="sd-dim">{{ row.label }}</span>
                <a
                  v-if="row.href"
                  class="sd-link about-list__link"
                  :href="row.href"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {{ row.value }}
                </a>
                <span v-else>{{ row.value }}</span>
              </li>
            </ul>
          </div>
        </section>
      </div>

      <section v-else class="sd-panel">
        <div class="sd-panel__body">
          <EmptyState text="站点暂未填写更多信息" />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchArticleList } from '@/api/article'
import { fetchUserList } from '@/api/user'
import { useSiteStore } from '@/stores'

const siteStore = useSiteStore()

const counts = reactive<{ articles: number | null; users: number | null }>({
  articles: null,
  users: null,
})

const stats = computed(() => {
  const items: Array<{ label: string; value: string | number }> = []
  if (counts.articles !== null) items.push({ label: '篇文章', value: counts.articles })
  if (counts.users !== null) items.push({ label: '位用户', value: counts.users })
  const version = siteStore.about.Version
  if (version) items.push({ label: '后端版本', value: version })
  return items
})

const siteRows = computed(() => {
  const rows: Array<{ label: string; value: string }> = []
  const about = siteStore.about
  if (about.siteDate) rows.push({ label: '建站时间', value: about.siteDate })
  if (siteStore.beian) rows.push({ label: '备案号', value: siteStore.beian })
  rows.push({ label: '运行模式', value: siteStore.mode === 2 ? '社区模式' : '博客模式' })
  rows.push({ label: '文章审核', value: siteStore.reviewEnabled ? '已开启' : '已关闭' })
  return rows
})

const contactRows = computed(() => {
  const rows: Array<{ label: string; value: string; href?: string }> = []
  const about = siteStore.about
  if (about.qq) rows.push({ label: 'QQ', value: about.qq })
  if (about.wechat) rows.push({ label: '微信', value: about.wechat })
  if (about.biliBili) rows.push({ label: '哔哩哔哩', value: about.biliBili, href: about.biliBili })
  if (about.gitHub) rows.push({ label: 'GitHub', value: about.gitHub, href: about.gitHub })
  return rows
})

onMounted(async () => {
  try {
    const data = await fetchArticleList({ type: 'other', page: 1, limit: 1 })
    counts.articles = data?.count ?? 0
  } catch {
    counts.articles = null
  }
  try {
    const data = await fetchUserList({ page: 1, limit: 1 })
    counts.users = data?.count ?? 0
  } catch {
    counts.users = null
  }
})
</script>

<style scoped lang="scss">
.about-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 44px 24px 34px;
  margin-bottom: 20px;
  text-align: center;
}

.about-hero__title {
  font-size: 34px;
  font-weight: 800;
}

.about-hero__desc {
  max-width: 620px;
  color: var(--sd-text-muted);
}

.about-hero__stats {
  display: flex;
  gap: 46px;
  margin-top: 12px;
}

.about-hero__stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
}

.about-hero__value {
  font-size: 24px;
  font-weight: 700;
}

.about-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 18px;
}

.about-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 13px;

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    padding-bottom: 10px;
    border-bottom: 1px dashed rgba(30, 43, 69, 0.7);
  }

  li:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
}

.about-list__link {
  word-break: break-all;
  text-align: right;
}
</style>
