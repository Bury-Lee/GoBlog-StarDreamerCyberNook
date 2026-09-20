<template>
  <div class="admin-dashboard">
    <section class="stat-grid">
      <div v-for="card in statCards" :key="card.label" class="stat-card sd-panel sd-panel--hover">
        <span class="stat-card__icon" :style="{ color: card.color }">
          <el-icon :size="22"><component :is="card.icon" /></el-icon>
        </span>
        <div class="stat-card__text">
          <span class="stat-card__value">{{ card.value }}</span>
          <span class="sd-dim">{{ card.label }}</span>
        </div>
      </div>
    </section>

    <div class="admin-dashboard__grid">
      <section class="sd-panel">
        <header class="sd-panel__header">
          <span class="sd-panel__title">最新日志</span>
          <router-link class="sd-link admin-dashboard__more" :to="{ name: 'admin-logs' }">全部日志</router-link>
        </header>
        <div class="sd-panel__body">
          <el-table :data="logs" size="small">
            <el-table-column label="级别" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="logLevelType(row.level)">{{ logLevelLabel(row.level) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="serviceName" label="服务" width="130" show-overflow-tooltip />
            <el-table-column label="时间" width="160">
              <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
            </el-table-column>
          </el-table>
        </div>
      </section>

      <section class="sd-panel">
        <header class="sd-panel__header">
          <span class="sd-panel__title">快捷入口</span>
        </header>
        <div class="sd-panel__body admin-dashboard__links">
          <router-link
            v-for="link in quickLinks"
            :key="link.name"
            class="quick-link"
            :to="{ name: link.name }"
          >
            <el-icon :size="18"><component :is="link.icon" /></el-icon>
            <span>{{ link.label }}</span>
          </router-link>
        </div>
      </section>
    </div>

    <section class="sd-panel">
      <header class="sd-panel__header">
        <span class="sd-panel__title">待审核文章</span>
        <router-link class="sd-link admin-dashboard__more" :to="{ name: 'admin-review' }">前往审核</router-link>
      </header>
      <div class="sd-panel__body">
        <el-table :data="pending" size="small">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
          <el-table-column label="作者 ID" width="100">
            <template #default="{ row }">{{ row.userID }}</template>
          </el-table-column>
          <el-table-column label="更新时间" width="170">
            <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  Bell,
  ChatDotSquare,
  Document,
  Link,
  Picture,
  User,
} from '@element-plus/icons-vue'
import { fetchArticleList, fetchReviewArticles } from '@/api/article'
import { fetchLogs } from '@/api/ops'
import { fetchUserList } from '@/api/user'
import { fetchImages } from '@/api/ops'
import type { ArticleModel, LogModel } from '@/api/types'
import { formatDate, logLevelLabel, logLevelType } from '@/utils/format'

const logs = ref<LogModel[]>([])
const pending = ref<ArticleModel[]>([])
const counts = ref({ articles: 0, users: 0, images: 0, logs: 0, pending: 0 })

const statCards = ref([
  { label: '文章总数', value: 0, icon: Document, color: '#22d3ee' },
  { label: '注册用户', value: 0, icon: User, color: '#818cf8' },
  { label: '图库图片', value: 0, icon: Picture, color: '#a855f7' },
  { label: '操作日志', value: 0, icon: ChatDotSquare, color: '#fbbf24' },
  { label: '待审核', value: 0, icon: Bell, color: '#f87171' },
])

const quickLinks = [
  { name: 'admin-review', label: '文章审核', icon: Bell },
  { name: 'admin-articles', label: '文章管理', icon: Document },
  { name: 'admin-users', label: '用户管理', icon: User },
  { name: 'admin-images', label: '图片管理', icon: Picture },
  { name: 'admin-banners', label: '轮播图管理', icon: Picture },
  { name: 'admin-friend-links', label: '友情链接', icon: Link },
  { name: 'admin-logs', label: '日志管理', icon: ChatDotSquare },
  { name: 'admin-site', label: '站点配置', icon: Link },
]

async function loadStats(): Promise<void> {
  const [articles, users, images, logs, review] = await Promise.allSettled([
    fetchArticleList({ type: 'admin', page: 1, limit: 1 }),
    fetchUserList({ page: 1, limit: 1 }),
    fetchImages({ page: 1, limit: 1 }),
    fetchLogs({ page: 1, limit: 1 }),
    fetchReviewArticles({ page: 1, limit: 1 }),
  ])

  counts.value = {
    articles: articles.status === 'fulfilled' ? articles.value.count : 0,
    users: users.status === 'fulfilled' ? users.value.count : 0,
    images: images.status === 'fulfilled' ? images.value.count : 0,
    logs: logs.status === 'fulfilled' ? logs.value.count : 0,
    pending: review.status === 'fulfilled' ? review.value.count : 0,
  }

  statCards.value[0].value = counts.value.articles
  statCards.value[1].value = counts.value.users
  statCards.value[2].value = counts.value.images
  statCards.value[3].value = counts.value.logs
  statCards.value[4].value = counts.value.pending
}

async function loadPanels(): Promise<void> {
  try {
    const data = await fetchLogs({ page: 1, limit: 6 })
    logs.value = data?.list ?? []
  } catch {
    logs.value = []
  }
  try {
    const data = await fetchReviewArticles({ page: 1, limit: 5 })
    pending.value = data?.list ?? []
  } catch {
    pending.value = []
  }
}

onMounted(() => {
  void loadStats()
  void loadPanels()
})
</script>

<style scoped lang="scss">
.admin-dashboard {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
}

.stat-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(30, 43, 69, 0.55);
  border: 1px solid var(--sd-border);
}

.stat-card__text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
}

.stat-card__value {
  font-size: 24px;
  font-weight: 700;
}

.admin-dashboard__grid {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
  gap: 18px;
}

.admin-dashboard__more {
  font-size: 12px;
}

.admin-dashboard__links {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
}

.quick-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  font-size: 13px;
  color: var(--sd-text-muted);
  border: 1px solid var(--sd-border);
  border-radius: 10px;
  background: rgba(9, 14, 26, 0.55);
  transition: all 0.2s ease;
}

.quick-link:hover {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.5);
  transform: translateY(-2px);
}

@media (max-width: 1100px) {
  .admin-dashboard__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
