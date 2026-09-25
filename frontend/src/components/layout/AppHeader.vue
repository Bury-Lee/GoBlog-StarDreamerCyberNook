<template>
  <header class="app-header">
    <div class="sd-container app-header__inner">
      <router-link :to="{ name: 'home' }" class="app-header__brand">
        <span class="app-header__logo">
          <img v-if="siteStore.logo" :src="siteStore.logo" alt="logo" @error="logoFailed = true" />
          <el-icon v-if="!siteStore.logo || logoFailed"><Platform /></el-icon>
        </span>
        <span class="app-header__title sd-neon-text">{{ siteStore.title }}</span>
      </router-link>

      <nav class="app-header__nav">
        <template v-for="item in navItems" :key="item.name">
          <el-dropdown v-if="item.children" trigger="hover" @command="goName">
            <span class="app-header__nav-item" :class="{ 'is-active': isGroupActive(item) }">
              {{ item.label }}
              <el-icon class="app-header__nav-caret"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="child in item.children" :key="child.name" :command="child.name">
                  {{ child.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <router-link
            v-else
            :to="{ name: item.name }"
            class="app-header__nav-item"
            :class="{ 'is-active': isActive(item.name) }"
          >
            {{ item.label }}
          </router-link>
        </template>
      </nav>

      <div class="app-header__actions">
        <el-input
          v-model="keyword"
          class="app-header__search"
          placeholder="搜索文章 / 标签"
          clearable
          @keyup.enter="goSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-tooltip content="写文章" placement="bottom">
          <el-button class="app-header__icon-btn" circle @click="goWrite">
            <el-icon><EditPen /></el-icon>
          </el-button>
        </el-tooltip>

        <el-tooltip content="消息中心" placement="bottom">
          <el-badge :value="messageStore.total" :hidden="messageStore.total <= 0" :max="99">
            <el-button class="app-header__icon-btn" circle @click="goMessages">
              <el-icon><Bell /></el-icon>
            </el-button>
          </el-badge>
        </el-tooltip>

        <el-dropdown v-if="userStore.isLogin" trigger="click" @command="onCommand">
          <div class="app-header__user">
            <UserAvatar :src="userStore.avatar" :name="userStore.nickname" :size="34" />
            <span class="app-header__user-name sd-ellipsis">{{ userStore.nickname }}</span>
            <el-icon class="sd-dim"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile" :icon="User">个人主页</el-dropdown-item>
              <el-dropdown-item command="articles" :icon="Document">我的文章</el-dropdown-item>
              <el-dropdown-item command="collections" :icon="Star">我的收藏</el-dropdown-item>
              <el-dropdown-item command="history" :icon="Clock">浏览记录</el-dropdown-item>
              <el-dropdown-item command="settings" :icon="Setting">个人设置</el-dropdown-item>
              <el-dropdown-item command="chat-all" :icon="ChatDotRound">会话列表</el-dropdown-item>
              <el-dropdown-item v-if="userStore.isAdmin" command="admin" :icon="Monitor" divided>
                管理后台
              </el-dropdown-item>
              <el-dropdown-item command="logout" :icon="SwitchButton" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <template v-else>
          <el-button text @click="goLogin">登录</el-button>
          <el-button type="primary" @click="goRegister">注册</el-button>
        </template>

        <el-button class="app-header__icon-btn app-header__menu-btn" circle @click="menuOpen = true">
          <el-icon><Menu /></el-icon>
        </el-button>
      </div>
    </div>

    <el-drawer v-model="menuOpen" title="导航" size="80%" append-to-body>
      <div class="app-header__drawer-nav">
        <router-link
          v-for="item in navItems"
          :key="item.name"
          :to="{ name: item.name }"
          class="app-header__drawer-link"
          :class="{ 'is-active': isActive(item.name) }"
          @click="menuOpen = false"
        >
          {{ item.label }}
        </router-link>
      </div>
      <div class="app-header__drawer-search">
        <el-input
          v-model="keyword"
          class="app-header__drawer-search-input"
          placeholder="搜索文章 / 标签"
          clearable
          @keyup.enter="goSearchFromMenu"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="goSearchFromMenu">搜索</el-button>
      </div>
    </el-drawer>
  </header>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowDown,
  Bell,
  ChatDotRound,
  Clock,
  Document,
  EditPen,
  Menu,
  Monitor,
  Platform,
  Search,
  Setting,
  Star,
  SwitchButton,
  User,
} from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useMessageStore, useSiteStore, useUserStore } from '@/stores'

const router = useRouter()
const route = useRoute()
const siteStore = useSiteStore()
const userStore = useUserStore()
const messageStore = useMessageStore()

const keyword = ref('')
const logoFailed = ref(false)
const menuOpen = ref(false)

interface NavChild {
  name: string
  label: string
}

interface NavItem {
  name: string
  label: string
  children?: NavChild[]
}

const navItems = computed<NavItem[]>(() => [
  { name: 'home', label: '首页' },
  { name: 'articles', label: '文章' },
  {
    name: 'about-group',
    label: '关于',
    children: [
      { name: 'about', label: '关于本站' },
      { name: 'feedback', label: '功能反馈' },
    ],
  },
])

function goName(name: string): void {
  router.push({ name })
}

function isGroupActive(item: NavItem): boolean {
  return Boolean(item.children?.some((child) => isActive(child.name)))
}

function isActive(name: string): boolean {
  if (name === 'articles') {
    return route.name === 'articles' || route.name === 'article-detail'
  }
  return route.name === name
}

function goSearch(): void {
  const key = keyword.value.trim()
  router.push({ name: 'search', query: key ? { key } : {} })
}

function goSearchFromMenu(): void {
  goSearch()
  menuOpen.value = false
}

function goWrite(): void {
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录后再写文章')
    router.push({ name: 'login', query: { redirect: '/write' } })
    return
  }
  router.push({ name: 'article-create' })
}

function goMessages(): void {
  if (!userStore.isLogin) {
    router.push({ name: 'login', query: { redirect: '/messages' } })
    return
  }
  router.push({ name: 'messages' })
}

function goLogin(): void {
  router.push({ name: 'login' })
}

function goRegister(): void {
  router.push({ name: 'register' })
}

async function onCommand(command: string): Promise<void> {
  const id = userStore.userId
  switch (command) {
    case 'chat-all':
      router.push({ name: 'chat' })
      break
    case 'profile':
      router.push({ name: 'user-home', params: { id } })
      break
    case 'articles':
      router.push({ name: 'my-articles' })
      break
    case 'collections':
      router.push({ name: 'collections' })
      break
    case 'history':
      router.push({ name: 'history' })
      break
    case 'settings':
      router.push({ name: 'user-settings' })
      break
    case 'admin':
      router.push({ name: 'admin-dashboard' })
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确认退出当前账号?', '退出登录', { type: 'warning' })
      } catch {
        return
      }
      await userStore.logout()
      messageStore.clear()
      messageStore.stopPolling()
      ElMessage.success('已退出登录')
      router.push({ name: 'home' })
      break
    default:
      break
  }
}
</script>

<style scoped lang="scss">
.app-header {
  position: sticky;
  top: 0;
  z-index: 30;
  border-bottom: 1px solid var(--sd-border);
  background: rgba(8, 12, 22, 0.86);
  backdrop-filter: blur(14px);
}

.app-header__inner {
  display: flex;
  align-items: center;
  gap: 18px;
  height: 62px;
}

.app-header__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.app-header__logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid rgba(34, 211, 238, 0.4);
  background: rgba(34, 211, 238, 0.08);
  box-shadow: 0 0 16px -6px rgba(34, 211, 238, 0.9);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.app-header__title {
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 0.5px;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-header__nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.app-header__nav-item {
  position: relative;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 14px;
  color: var(--sd-text-muted);
  transition: all 0.2s ease;
}

.app-header__nav-item:hover {
  color: var(--sd-text);
  background: rgba(34, 211, 238, 0.08);
}

.app-header__nav-item.is-active {
  color: var(--sd-cyan);
  background: rgba(34, 211, 238, 0.12);
}

.app-header__actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.app-header__search {
  width: 210px;
}

.app-header__icon-btn {
  background: rgba(30, 43, 69, 0.5);
  border-color: var(--sd-border);
  color: var(--sd-text-muted);
}

.app-header__icon-btn:hover {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.6);
}

.app-header__user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 3px 10px 3px 3px;
  border: 1px solid var(--sd-border);
  border-radius: 999px;
  background: rgba(30, 43, 69, 0.35);
  cursor: pointer;
  transition: all 0.2s ease;
}

.app-header__user:hover {
  border-color: rgba(34, 211, 238, 0.6);
}

.app-header__user-name {
  max-width: 96px;
  font-size: 13px;
}

.app-header__menu-btn {
  display: none;
}

.app-header__drawer-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 18px;
}

.app-header__drawer-link {
  padding: 10px 12px;
  border-radius: 10px;
  font-size: 15px;
  color: var(--sd-text-muted);
  background: rgba(30, 43, 69, 0.35);
  transition: all 0.2s ease;
}

.app-header__drawer-link.is-active {
  color: var(--sd-cyan);
  background: rgba(34, 211, 238, 0.12);
}

.app-header__drawer-search {
  display: flex;
  gap: 8px;
}

.app-header__drawer-search-input {
  flex: 1;
}

@media (max-width: 1180px) {
  .app-header__search {
    display: none;
  }
}

@media (max-width: 860px) {
  .app-header__nav,
  .app-header__user-name {
    display: none;
  }

  .app-header__menu-btn {
    display: inline-flex;
  }

  .app-header__inner {
    gap: 10px;
  }
}
</style>
