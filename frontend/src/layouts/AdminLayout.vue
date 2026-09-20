<template>
  <div class="admin-layout">
    <aside class="admin-layout__aside" :class="{ 'is-collapsed': collapsed }">
      <div class="admin-layout__brand">
        <span class="admin-layout__logo">
          <el-icon><Monitor /></el-icon>
        </span>
        <span v-show="!collapsed" class="admin-layout__brand-text sd-neon-text">管理控制台</span>
      </div>
      <el-scrollbar class="admin-layout__menu">
        <el-menu
          :default-active="activeMenu"
          :collapse="collapsed"
          :collapse-transition="false"
          router
          class="admin-layout__el-menu"
        >
          <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.label }}</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </aside>

    <div class="admin-layout__main">
      <header class="admin-layout__header">
        <el-button text class="admin-layout__collapse" @click="collapsed = !collapsed">
          <el-icon :size="18">
            <component :is="collapsed ? Expand : Fold" />
          </el-icon>
        </el-button>
        <div class="admin-layout__crumb">
          <span class="sd-dim">管理后台</span>
          <el-icon class="sd-dim"><ArrowRight /></el-icon>
          <span>{{ currentTitle }}</span>
        </div>
        <div class="admin-layout__actions">
          <el-button text @click="router.push({ name: 'home' })">
            <el-icon><HomeFilled /></el-icon>
            <span class="admin-layout__action-text">返回前台</span>
          </el-button>
          <el-dropdown trigger="click" @command="onCommand">
            <div class="admin-layout__user">
              <UserAvatar :src="userStore.avatar" :name="userStore.nickname" :size="30" />
              <span class="admin-layout__user-name">{{ userStore.nickname }}</span>
              <el-icon class="sd-dim"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile" :icon="User">个人主页</el-dropdown-item>
                <el-dropdown-item command="settings" :icon="Setting">个人设置</el-dropdown-item>
                <el-dropdown-item command="logout" :icon="SwitchButton" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <main class="admin-layout__content">
        <router-view v-slot="{ Component }">
          <transition name="sd-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowDown,
  ArrowRight,
  Bell,
  ChatDotSquare,
  Collection,
  DataAnalysis,
  Document,
  Expand,
  Fold,
  HomeFilled,
  Link,
  List,
  Monitor,
  Picture,
  Promotion,
  Setting,
  SwitchButton,
  User,
} from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useMessageStore, useUserStore } from '@/stores'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const messageStore = useMessageStore()

const collapsed = ref(false)

const menus = [
  { path: '/admin', label: '控制台', icon: DataAnalysis },
  { path: '/admin/review', label: '文章审核', icon: Bell },
  { path: '/admin/articles', label: '文章管理', icon: Document },
  { path: '/admin/categories', label: '分类管理', icon: Collection },
  { path: '/admin/users', label: '用户管理', icon: User },
  { path: '/admin/banners', label: '轮播图管理', icon: Picture },
  { path: '/admin/friend-links', label: '友情链接', icon: Link },
  { path: '/admin/promotions', label: '友站推广', icon: Promotion },
  { path: '/admin/images', label: '图片管理', icon: List },
  { path: '/admin/logs', label: '日志管理', icon: ChatDotSquare },
  { path: '/admin/site', label: '站点配置', icon: Setting },
]

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => (route.meta.title as string) || '控制台')

async function onCommand(command: string): Promise<void> {
  if (command === 'profile') {
    router.push({ name: 'user-home', params: { id: userStore.userId } })
    return
  }
  if (command === 'settings') {
    router.push({ name: 'user-settings' })
    return
  }
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确认退出当前账号?', '退出登录', { type: 'warning' })
    } catch {
      return
    }
    await userStore.logout()
    messageStore.stopPolling()
    messageStore.clear()
    ElMessage.success('已退出登录')
    router.push({ name: 'home' })
  }
}
</script>

<style scoped lang="scss">
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--sd-bg-deep);
}

.admin-layout__aside {
  display: flex;
  flex-direction: column;
  width: 210px;
  flex-shrink: 0;
  border-right: 1px solid var(--sd-border);
  background: linear-gradient(180deg, rgba(14, 21, 36, 0.98), rgba(9, 14, 26, 0.98));
  transition: width 0.2s ease;
}

.admin-layout__aside.is-collapsed {
  width: 64px;
}

.admin-layout__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 62px;
  padding: 0 16px;
  border-bottom: 1px solid var(--sd-border);
}

.admin-layout__logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 10px;
  border: 1px solid rgba(34, 211, 238, 0.4);
  background: rgba(34, 211, 238, 0.1);
  color: var(--sd-cyan);
}

.admin-layout__brand-text {
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.admin-layout__menu {
  flex: 1;
  min-height: 0;
}

.admin-layout__el-menu {
  border-right: none;
  background: transparent;
}

:deep(.admin-layout__el-menu .el-menu-item) {
  height: 44px;
  line-height: 44px;
  margin: 4px 8px;
  border-radius: 10px;
  color: var(--sd-text-muted);
}

:deep(.admin-layout__el-menu .el-menu-item:hover) {
  background: rgba(34, 211, 238, 0.08);
  color: var(--sd-text);
}

:deep(.admin-layout__el-menu .el-menu-item.is-active) {
  background: linear-gradient(92deg, rgba(34, 211, 238, 0.22), rgba(168, 85, 247, 0.18));
  color: #e0f2fe;
  box-shadow: inset 0 0 0 1px rgba(34, 211, 238, 0.35);
}

.admin-layout__main {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.admin-layout__header {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 62px;
  padding: 0 18px;
  border-bottom: 1px solid var(--sd-border);
  background: rgba(8, 12, 22, 0.86);
  backdrop-filter: blur(12px);
  position: sticky;
  top: 0;
  z-index: 20;
}

.admin-layout__crumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.admin-layout__actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.admin-layout__user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 3px 10px 3px 3px;
  border: 1px solid var(--sd-border);
  border-radius: 999px;
  background: rgba(30, 43, 69, 0.35);
  cursor: pointer;
}

.admin-layout__user-name {
  font-size: 13px;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-layout__content {
  flex: 1;
  padding: 20px;
  min-width: 0;
}

@media (max-width: 860px) {
  .admin-layout__aside {
    width: 64px;
  }

  .admin-layout__brand-text,
  .admin-layout__action-text,
  .admin-layout__user-name {
    display: none;
  }
}
</style>
