<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="search-panel sd-panel">
        <div class="search-panel__row">
          <el-input
            v-model="keyInput"
            size="large"
            :placeholder="mode === 'user' ? '输入用户ID精确搜索，或昵称模糊搜索' : '输入关键词搜索文章标题、摘要、正文…'"
            clearable
            @keyup.enter="applySearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-input
            v-if="mode === 'article'"
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
        <div class="search-panel__tabs">
          <span
            class="search-panel__tab"
            :class="{ 'is-active': mode === 'article' }"
            @click="switchMode('article')"
          >
            文章
          </span>
          <span
            class="search-panel__tab"
            :class="{ 'is-active': mode === 'user' }"
            @click="switchMode('user')"
          >
            用户
          </span>
        </div>
        <div v-if="mode === 'article'" class="search-panel__sorts">
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

      <template v-if="mode === 'article'">
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
      </template>

      <template v-else>
        <div class="search-result__head">
          <span class="sd-dim">
            共找到 <b class="sd-neon-text">{{ userCapped ? `>${userCount}` : userCount }}</b> 个用户
          </span>
          <span v-if="key" class="sd-chip">{{ isPureDigits(key) ? `用户ID:${key}` : `昵称:${key}` }}</span>
        </div>

        <div v-loading="userLoading" class="user-result">
          <EmptyState v-if="!userLoading && !userList.length" text="没有匹配的用户,试试输入用户ID或昵称" />
          <router-link
            v-for="u in userList"
            :key="u.userID"
            class="user-result__item sd-panel"
            :to="{ name: 'user-home', params: { id: u.userID } }"
          >
            <UserAvatar :src="u.avatar" :name="u.nickname" :size="42" />
            <div class="user-result__main">
              <div class="user-result__name">{{ u.nickname || `用户 #${u.userID}` }}</div>
              <div class="sd-dim user-result__desc">{{ u.abstract || '这个人很懒，什么都没写' }}</div>
            </div>
          </router-link>
        </div>

        <PaginationBar
          :page="userPage"
          :limit="userLimit"
          :count="userCount"
          :page-sizes="[10, 20, 30, 40]"
          @update:page="changeUserPage"
          @update:limit="changeUserLimit"
        />
      </template>
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
import UserAvatar from '@/components/common/UserAvatar.vue'
import { searchArticles } from '@/api/article'
import { fetchUserList } from '@/api/user'
import type { UserListItem } from '@/api/types'
import { SEARCH_SORT_OPTIONS } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const route = useRoute()
const router = useRouter()

const key = ref(String(route.query.key || ''))
const tag = ref(String(route.query.tag || ''))
const sortType = ref(Number(route.query.type ?? 0))
const keyInput = ref(key.value)
const tagInput = ref(tag.value)

// 搜索类型:文章 / 用户
const mode = ref<'article' | 'user'>('article')
const userList = ref<UserListItem[]>([])
const userCount = ref(0)
const userCapped = ref(false)
const userPage = ref(1)
const userLimit = ref(10)
const userLoading = ref(false)

// 纯数字按用户ID精确搜索,否则按昵称模糊
function isPureDigits(value: string): boolean {
  return /^\d+$/.test(value.trim())
}

async function loadUsers(): Promise<void> {
  const kw = key.value.trim()
  if (!kw) {
    userList.value = []
    userCount.value = 0
    return
  }
  userLoading.value = true
  try {
    const params = isPureDigits(kw)
      ? { userID: Number(kw), page: userPage.value, limit: userLimit.value }
      : { key: kw, page: userPage.value, limit: userLimit.value }
    const data = await fetchUserList(params)
    userList.value = data.list
    userCount.value = data.count
    userCapped.value = Boolean(data.capped)
  } catch {
    userList.value = []
    userCount.value = 0
  } finally {
    userLoading.value = false
  }
}

function switchMode(next: 'article' | 'user'): void {
  if (mode.value === next) return
  mode.value = next
  if (next === 'user') {
    userPage.value = 1
    void loadUsers()
  } else {
    void load()
  }
}

function changeUserPage(next: number): void {
  userPage.value = next
  void loadUsers()
}

function changeUserLimit(next: number): void {
  userLimit.value = next
  userPage.value = 1
  void loadUsers()
}

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
  userPage.value = 1
  void router.replace({
    name: 'search',
    query: {
      ...(key.value ? { key: key.value } : {}),
      ...(tag.value ? { tag: tag.value } : {}),
      ...(sortType.value ? { type: String(sortType.value) } : {}),
    },
  })
  if (mode.value === 'user') {
    void loadUsers()
  } else {
    void load()
  }
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

.search-panel__tabs {
  display: flex;
  gap: 18px;
  margin-top: 14px;
  border-bottom: 1px solid var(--sd-border);
}

.search-panel__tab {
  padding: 6px 2px;
  font-size: 14px;
  color: var(--sd-text-muted);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
}

.search-panel__tab:hover,
.search-panel__tab.is-active {
  color: var(--sd-cyan);
  border-bottom-color: var(--sd-cyan);
}

.user-result {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 160px;
}

.user-result__item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;
  text-decoration: none;
  color: inherit;
  transition: border-color 0.2s ease;
}

.user-result__item:hover {
  border-color: var(--sd-cyan);
}

.user-result__main {
  min-width: 0;
}

.user-result__name {
  font-size: 15px;
  font-weight: 600;
}

.user-result__desc {
  margin-top: 4px;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
