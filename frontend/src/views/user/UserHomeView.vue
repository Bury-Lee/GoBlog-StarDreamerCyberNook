<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="profile-card sd-panel sd-grid-bg">
        <div class="profile-card__bg" />
        <div class="profile-card__main">
          <UserAvatar :src="info?.avatar" :name="info?.nickName" :size="88" />
          <div class="profile-card__text">
            <h1 class="profile-card__name">
              {{ info?.nickName || '用户' }}
              <el-tag v-if="isSelf" size="small" type="success" effect="dark">我自己</el-tag>
            </h1>
            <p class="profile-card__abstract sd-dim">
              {{ detail?.abstract || '这位用户很低调,还没有填写简介' }}
            </p>
            <div class="profile-card__meta">
              <span v-if="info?.region"><el-icon><Location /></el-icon>{{ info.region }}</span>
              <span v-if="info?.age"><el-icon><Calendar /></el-icon>{{ info.age }} 岁</span>
              <span>
                <el-icon><Timer /></el-icon>
                {{ info?.existDay === 0 ? '今天刚加入' : `入驻 ${info?.existDay || 0} 天` }}
              </span>
              <span v-if="info?.lastLoginTime"><el-icon><Clock /></el-icon>最近活跃 {{ fromNow(info.lastLoginTime) }}</span>
            </div>
          </div>
          <div class="profile-card__actions">
            <el-button v-if="isSelf" type="primary" @click="goSettings">
              <el-icon><Setting /></el-icon>
              编辑资料
            </el-button>
            <template v-else>
              <el-button type="primary" plain :loading="followLoading" @click="onFollow">
                {{ followed ? '已关注' : '关注' }}
              </el-button>
              <el-button @click="goChat">发私信</el-button>
            </template>
          </div>
        </div>

        <div class="profile-card__stats">
          <button type="button" class="profile-card__stat" @click="switchTab('articles')">
            <span class="profile-card__stat-value sd-neon-text">{{ info?.articleCount || 0 }}</span>
            <span class="sd-dim">文章</span>
          </button>
          <button type="button" class="profile-card__stat" @click="switchTab('social')">
            <span class="profile-card__stat-value sd-neon-text">{{ info?.fansCount || 0 }}</span>
            <span class="sd-dim">粉丝</span>
          </button>
          <button type="button" class="profile-card__stat" @click="switchTab('social')">
            <span class="profile-card__stat-value sd-neon-text">{{ info?.followCount || 0 }}</span>
            <span class="sd-dim">关注</span>
          </button>
        </div>
      </section>

      <section class="profile-body sd-panel">
        <el-tabs v-model="activeTab" class="profile-body__tabs" @tab-change="onTabChange">
          <el-tab-pane label="文章" name="articles">
            <div v-loading="articles.loading" class="profile-body__articles">
              <EmptyState v-if="!articles.loading && !articles.list.length" text="还没有发布文章" />
              <ArticleCard
                v-for="item in articles.list"
                :key="item.id"
                :article="item"
                layout="list"
              />
            </div>
            <PaginationBar
              :page="articles.page"
              :limit="articles.limit"
              :count="articles.count"
              layout="prev, pager, next"
              @update:page="changeArticlePage"
            />
          </el-tab-pane>

          <el-tab-pane label="收藏夹" name="collect">
            <div v-loading="collect.loading" class="profile-body__folders">
              <EmptyState
                v-if="!collect.loading && !collect.list.length"
                :text="collect.error || '暂无公开的收藏夹'"
              />
              <div
                v-for="folder in collect.list"
                :key="folder.id"
                class="profile-folder"
                @click="openFolder(folder)"
              >
                <div class="profile-folder__cover">
                  <img v-if="folder.cover" :src="resolveAssetUrl(folder.cover)" alt="cover" />
                  <el-icon v-else :size="22"><Star /></el-icon>
                </div>
                <div class="profile-folder__text">
                  <span class="profile-folder__title sd-ellipsis">{{ folder.title }}</span>
                  <span class="sd-dim sd-clamp-2">{{ folder.abstract || '暂无描述' }}</span>
                </div>
                <el-tag v-if="folder.isDefault" size="small" type="info">默认</el-tag>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="关注 / 粉丝" name="social">
            <div v-loading="socialLoading" class="profile-body__social">
              <div class="profile-body__social-grid">
                <div class="profile-body__social-col">
                  <h4 class="profile-body__social-title">关注({{ follows.count }})</h4>
                  <EmptyState
                    v-if="!socialLoading && !follows.list.length"
                    :text="follows.error || '暂无数据'"
                    compact
                  />
                  <button
                    v-for="item in follows.list"
                    :key="item.id"
                    type="button"
                    class="profile-body__social-item"
                    @click="goUser(item.id)"
                  >
                    <UserAvatar :src="item.avatar" :name="item.nickname" :size="30" />
                    <span class="sd-ellipsis">{{ item.nickname || `用户 #${item.id}` }}</span>
                  </button>
                  <el-button
                    v-if="follows.list.length < follows.count"
                    text
                    size="small"
                    class="profile-body__social-more"
                    :loading="follows.loading"
                    @click="loadMoreFollows"
                  >
                    加载更多
                  </el-button>
                </div>
                <div class="profile-body__social-col">
                  <h4 class="profile-body__social-title">粉丝({{ followers.count }})</h4>
                  <EmptyState
                    v-if="!socialLoading && !followers.list.length"
                    :text="followers.error || '暂无数据'"
                    compact
                  />
                  <button
                    v-for="item in followers.list"
                    :key="item.id"
                    type="button"
                    class="profile-body__social-item"
                    @click="goUser(item.id)"
                  >
                    <UserAvatar :src="item.avatar" :name="item.nickname" :size="30" />
                    <span class="sd-ellipsis">{{ item.nickname || `用户 #${item.id}` }}</span>
                  </button>
                  <el-button
                    v-if="followers.list.length < followers.count"
                    text
                    size="small"
                    class="profile-body__social-more"
                    :loading="followers.loading"
                    @click="loadMoreFollowers"
                  >
                    加载更多
                  </el-button>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="浏览记录" name="history">
            <div v-loading="history.loading">
              <EmptyState v-if="!history.loading && !history.list.length" text="暂无公开的浏览记录" />
              <div v-for="item in history.list" :key="item.id" class="history-item" @click="goArticle(item.articleID)">
                <span class="history-item__title sd-ellipsis">{{ item.title }}</span>
                <span class="sd-dim">{{ formatDate(item.lookDate) }}</span>
              </div>
            </div>
            <PaginationBar
              :page="history.page"
              :limit="history.limit"
              :count="history.count"
              layout="prev, pager, next"
              @update:page="changeHistoryPage"
            />
          </el-tab-pane>
        </el-tabs>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Calendar, Clock, Location, Setting, Star, Timer } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import ArticleCard from '@/components/article/ArticleCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchArticleHistory, fetchArticleList, fetchCollectFolders } from '@/api/article'
import { checkFollow, followUser, fetchFollowList, fetchFollowerList, unfollowUser } from '@/api/follow'
import { fetchUserBaseInfo, fetchUserDetail } from '@/api/user'
import { resolveAssetUrl } from '@/api/request'
import type {
  ArticleHistoryItem,
  ArticleListResponse,
  CollectModel,
  FollowModel,
  UserBaseInfo,
  UserDetail,
} from '@/api/types'
import { formatDate, fromNow } from '@/utils/format'
import { useUserStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const info = ref<UserBaseInfo | null>(null)
const detail = ref<UserDetail | null>(null)
const activeTab = ref(String(route.query.tab || 'articles'))
const followed = ref(false)
const followLoading = ref(false)

const targetID = computed(() => Number(route.params.id || 0))
const isSelf = computed(() => userStore.userId === targetID.value)

const articles = reactive({
  list: [] as ArticleListResponse[],
  count: 0,
  page: 1,
  limit: 8,
  loading: false,
})

const collect = reactive({
  list: [] as CollectModel[],
  loading: false,
  error: '',
})

interface SocialUserItem {
  id: number
  nickname: string
  avatar: string
  createdAt: string
}

const follows = reactive({ list: [] as SocialUserItem[], count: 0, page: 1, limit: 10, error: '', loading: false })
const followers = reactive({ list: [] as SocialUserItem[], count: 0, page: 1, limit: 10, error: '', loading: false })
const socialLoading = ref(false)

const history = reactive({
  list: [] as ArticleHistoryItem[],
  count: 0,
  page: 1,
  limit: 10,
  loading: false,
})

async function loadProfile(): Promise<void> {
  if (!targetID.value) return
  try {
    info.value = await fetchUserBaseInfo(targetID.value)
  } catch {
    info.value = null
  }
  if (isSelf.value) {
    detail.value = userStore.profile
  } else {
    detail.value = null
  }
  await loadFollowState()
}

async function loadFollowState(): Promise<void> {
  if (!userStore.isLogin || isSelf.value || !targetID.value) {
    followed.value = false
    return
  }
  try {
    const data = await checkFollow(targetID.value)
    followed.value = data?.followed ?? false
  } catch {
    followed.value = false
  }
}

async function loadArticles(): Promise<void> {
  articles.loading = true
  try {
    const data = await fetchArticleList({
      type: 'other',
      userID: targetID.value,
      page: articles.page,
      limit: articles.limit,
    })
    articles.list = data?.list ?? []
    articles.count = data?.count ?? 0
  } catch {
    articles.list = []
    articles.count = 0
  } finally {
    articles.loading = false
  }
}

function changeArticlePage(next: number): void {
  articles.page = next
  void loadArticles()
}

async function loadCollect(): Promise<void> {
  collect.loading = true
  collect.error = ''
  try {
    const data = await fetchCollectFolders({ id: targetID.value, page: 1, limit: 20 })
    collect.list = data?.list ?? []
  } catch (error) {
    collect.list = []
    collect.error = error instanceof Error ? error.message : '该用户的收藏夹未公开'
  } finally {
    collect.loading = false
  }
}

async function loadSocial(): Promise<void> {
  socialLoading.value = true
  follows.page = 1
  followers.page = 1
  await Promise.all([loadFollows(), loadFollowers()])
  socialLoading.value = false
}

async function loadFollows(): Promise<void> {
  const isFirst = follows.page === 1
  follows.loading = true
  try {
    const data = await fetchFollowList({ userID: targetID.value, page: follows.page, limit: follows.limit })
    const items = (data?.list ?? []).map((item) => ({
      id: item.focusUserID,
      nickname: item.focusUserNickname,
      avatar: item.focusUserAvatar,
      createdAt: item.createdAt,
    }))
    follows.list = isFirst ? items : [...follows.list, ...items]
    follows.count = data?.count ?? 0
    follows.error = ''
  } catch (error) {
    const message = error instanceof Error ? error.message : '关注列表加载失败'
    follows.error = message
    if (isFirst) {
      follows.list = []
      follows.count = 0
    } else {
      follows.page = Math.max(1, follows.page - 1)
      ElMessage.warning(message)
    }
  } finally {
    follows.loading = false
  }
}

async function loadFollowers(): Promise<void> {
  const isFirst = followers.page === 1
  followers.loading = true
  try {
    const data = await fetchFollowerList({ userID: targetID.value, page: followers.page, limit: followers.limit })
    const items = await hydrateFollowers(data?.list ?? [])
    followers.list = isFirst ? items : [...followers.list, ...items]
    followers.count = data?.count ?? 0
    followers.error = ''
  } catch (error) {
    const message = error instanceof Error ? error.message : '粉丝列表加载失败'
    followers.error = message
    if (isFirst) {
      followers.list = []
      followers.count = 0
    } else {
      followers.page = Math.max(1, followers.page - 1)
      ElMessage.warning(message)
    }
  } finally {
    followers.loading = false
  }
}

function loadMoreFollows(): void {
  follows.page += 1
  void loadFollows()
}

function loadMoreFollowers(): void {
  followers.page += 1
  void loadFollowers()
}

// 粉丝列表接口只返回关注记录(不含粉丝的用户信息),这里按需补齐昵称与头像
async function hydrateFollowers(list: FollowModel[]): Promise<SocialUserItem[]> {
  const items: SocialUserItem[] = list
    .filter((item) => item.userID > 0)
    .map((item) => ({
      id: item.userID,
      nickname: '',
      avatar: '',
      createdAt: item.createdAt,
    }))
  const details = await Promise.allSettled(
    items.map((item) => fetchUserBaseInfo(item.id, { silent: true })),
  )
  details.forEach((result, index) => {
    if (result.status !== 'fulfilled') return
    items[index].nickname = result.value?.nickName || ''
    items[index].avatar = result.value?.avatar || ''
  })
  return items
}

async function loadHistory(): Promise<void> {
  history.loading = true
  try {
    const data = await fetchArticleHistory({
      userID: isSelf.value ? 0 : targetID.value,
      page: history.page,
      limit: history.limit,
    })
    history.list = data?.list ?? []
    history.count = data?.count ?? 0
  } catch {
    history.list = []
    history.count = 0
  } finally {
    history.loading = false
  }
}

function changeHistoryPage(next: number): void {
  history.page = next
  void loadHistory()
}

function onTabChange(name: string | number): void {
  const tab = String(name)
  if (tab === 'articles') void loadArticles()
  if (tab === 'collect') void loadCollect()
  if (tab === 'social') void loadSocial()
  if (tab === 'history') void loadHistory()
}

function switchTab(name: string): void {
  if (activeTab.value !== name) {
    activeTab.value = name
    if (route.query.tab !== name) {
      void router.replace({ query: { ...route.query, tab: name } })
    }
  }
  onTabChange(name)
}

function goUser(id: number): void {
  if (!id) return
  router.push({ name: 'user-home', params: { id } })
}

function openFolder(folder: CollectModel): void {
  router.push({ name: 'collections', query: { folder: String(folder.id), user: String(targetID.value) } })
}

function goArticle(id: number): void {
  router.push({ name: 'article-detail', params: { id } })
}

function goSettings(): void {
  router.push({ name: 'user-settings' })
}

function goChat(): void {
  router.push({ name: 'chat', query: { userID: String(targetID.value) } })
}

async function onFollow(): Promise<void> {
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录')
    return
  }
  followLoading.value = true
  try {
    if (followed.value) {
      await unfollowUser(targetID.value)
      followed.value = false
      if (info.value) info.value.followCount = Math.max(0, (info.value.followCount || 0) - 1)
      ElMessage.success('已取消关注')
    } else {
      await followUser(targetID.value)
      followed.value = true
      if (info.value) info.value.followCount = (info.value.followCount || 0) + 1
      ElMessage.success('关注成功')
    }
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : '关注操作失败')
  } finally {
    followLoading.value = false
  }
}

watch(targetID, () => {
  articles.page = 1
  history.page = 1
  follows.page = 1
  followers.page = 1
  followed.value = false
  void loadProfile()
  onTabChange(activeTab.value)
})

onMounted(() => {
  void loadProfile()
  onTabChange(activeTab.value)
})
</script>

<style scoped lang="scss">
.profile-card {
  position: relative;
  overflow: hidden;
  padding: 26px 28px 18px;
  margin-bottom: 20px;
}

.profile-card__bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(600px 200px at 10% 0%, rgba(34, 211, 238, 0.18), transparent 65%),
    radial-gradient(500px 200px at 90% 10%, rgba(168, 85, 247, 0.18), transparent 60%);
  pointer-events: none;
}

.profile-card__main {
  position: relative;
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.profile-card__text {
  flex: 1;
  min-width: 220px;
}

.profile-card__name {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 24px;
}

.profile-card__abstract {
  margin-top: 6px;
  font-size: 13px;
}

.profile-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 10px;
  font-size: 12px;
  color: var(--sd-text-dim);

  span {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
}

.profile-card__actions {
  display: flex;
  gap: 10px;
}

.profile-card__stats {
  position: relative;
  display: flex;
  gap: 46px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--sd-border);
}

.profile-card__stat {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: 0;
  border: none;
  background: none;
  color: inherit;
  font: inherit;
  font-size: 12px;
  text-align: left;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.profile-card__stat:hover {
  transform: translateY(-2px);
}

.profile-card__stat:hover .profile-card__stat-value {
  filter: brightness(1.2) drop-shadow(0 0 8px rgba(34, 211, 238, 0.6));
}

.profile-card__stat:hover .sd-dim {
  color: var(--sd-cyan);
}

.profile-card__stat-value {
  font-size: 22px;
  font-weight: 700;
}

.profile-body {
  padding: 8px 20px 20px;
}

.profile-body__articles {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 120px;
}

.profile-body__folders {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
  min-height: 120px;
}

.profile-folder {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.6);
  cursor: pointer;
  transition: all 0.25s ease;
}

.profile-folder:hover {
  border-color: rgba(34, 211, 238, 0.5);
  transform: translateY(-2px);
}

.profile-folder__cover {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 54px;
  height: 54px;
  flex-shrink: 0;
  overflow: hidden;
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.2), rgba(168, 85, 247, 0.2));
  color: var(--sd-cyan);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.profile-folder__text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
  font-size: 12px;
}

.profile-folder__title {
  font-size: 14px;
  font-weight: 600;
}

.profile-body__social {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.profile-body__social-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 18px;
}

.profile-body__social-col {
  padding: 14px;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.5);
}

.profile-body__social-title {
  margin-bottom: 10px;
  font-size: 14px;
}

.profile-body__social-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 6px 4px;
  border: none;
  border-radius: 8px;
  background: none;
  color: inherit;
  font: inherit;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.profile-body__social-item:hover {
  background: rgba(34, 211, 238, 0.08);
  color: var(--sd-cyan);
}

.profile-body__social-more {
  margin-top: 8px;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 6px;
  font-size: 13px;
  border-bottom: 1px dashed rgba(30, 43, 69, 0.7);
  cursor: pointer;
}

.history-item:hover .history-item__title {
  color: var(--sd-cyan);
}

.history-item__title {
  flex: 1;
  min-width: 0;
}
</style>
