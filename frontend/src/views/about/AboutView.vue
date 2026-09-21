<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="about-hero sd-panel sd-grid-bg">
        <span class="sd-chip">ABOUT THIS SITE</span>
        <h1 class="about-hero__title sd-neon-text">{{ siteStore.siteTitle }}</h1>
        <p class="about-hero__desc">
          {{ siteStore.seo.description || '一个基于 GoBlog 后端构建的技术分享与内容社区' }}
        </p>
        <div class="about-hero__stats">
          <div class="about-hero__stat">
            <span class="about-hero__value sd-neon-text">{{ stats.articles }}</span>
            <span class="sd-dim">篇文章</span>
          </div>
          <div class="about-hero__stat">
            <span class="about-hero__value sd-neon-text">{{ stats.users }}</span>
            <span class="sd-dim">位用户</span>
          </div>
          <div class="about-hero__stat">
            <span class="about-hero__value sd-neon-text">{{ siteStore.about.Version || 'v1' }}</span>
            <span class="sd-dim">后端版本</span>
          </div>
        </div>
      </section>

      <div class="about-grid">
        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">站点信息</span>
          </header>
          <div class="sd-panel__body">
            <ul class="about-list">
              <li>
                <span class="sd-dim">建站时间</span>
                <span>{{ siteStore.about.siteDate || '未填写' }}</span>
              </li>
              <li>
                <span class="sd-dim">备案号</span>
                <span>{{ siteStore.beian || '未填写' }}</span>
              </li>
              <li>
                <span class="sd-dim">运行模式</span>
                <span>{{ siteStore.mode === 2 ? '社区模式' : '博客模式' }}</span>
              </li>
              <li>
                <span class="sd-dim">文章审核</span>
                <span>{{ siteStore.reviewEnabled ? '已开启' : '已关闭' }}</span>
              </li>
            </ul>
          </div>
        </section>

        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">联系与关注</span>
          </header>
          <div class="sd-panel__body">
            <ul class="about-list">
              <li>
                <span class="sd-dim">QQ 群</span>
                <span>{{ siteStore.about.qq || '未填写' }}</span>
              </li>
              <li>
                <span class="sd-dim">微信</span>
                <span>{{ siteStore.about.wechat || '未填写' }}</span>
              </li>
              <li>
                <span class="sd-dim">哔哩哔哩</span>
                <a v-if="siteStore.about.biliBili" class="sd-link" :href="siteStore.about.biliBili" target="_blank" rel="noopener">
                  访问主页
                </a>
                <span v-else>未填写</span>
              </li>
              <li>
                <span class="sd-dim">GitHub</span>
                <a v-if="siteStore.about.gitHub" class="sd-link" :href="siteStore.about.gitHub" target="_blank" rel="noopener">
                  访问仓库
                </a>
                <span v-else>未填写</span>
              </li>
            </ul>
          </div>
        </section>

        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">技术栈</span>
          </header>
          <div class="sd-panel__body">
            <div class="about-tags">
              <span v-for="item in techStack" :key="item" class="sd-tag">{{ item }}</span>
            </div>
          </div>
        </section>

        <section class="sd-panel">
          <header class="sd-panel__header">
            <span class="sd-panel__title">已接入能力</span>
          </header>
          <div class="sd-panel__body">
            <ul class="about-list">
              <li v-for="item in features" :key="item">
                <span class="sd-dim">▹</span>
                <span>{{ item }}</span>
              </li>
            </ul>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { fetchArticleList } from '@/api/article'
import { fetchUserList } from '@/api/user'
import { useSiteStore } from '@/stores'

const siteStore = useSiteStore()

const stats = reactive({ articles: 0, users: 0 })

const techStack = ['Gin', 'GORM', 'Redis', 'JWT', 'Vue 3', 'TypeScript', 'Vite', 'Element Plus']

const features = [
  '文章发布、分类、搜索与审核流转',
  '点赞 / 评论 / 收藏 / 浏览记录',
  '站内消息中心与私聊会话',
  '轮播图、友链、友站推广运营位',
  'AI 助手对话与内容审核',
  '图片上传与静态资源托管',
]

onMounted(async () => {
  try {
    const data = await fetchArticleList({ type: 'other', page: 1, limit: 1 })
    stats.articles = data?.count ?? 0
  } catch {
    stats.articles = 0
  }
  try {
    const data = await fetchUserList({ page: 1, limit: 1 })
    stats.users = data?.count ?? 0
  } catch {
    stats.users = 0
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

.about-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
