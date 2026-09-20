import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { routes } from './routes'
import { useSiteStore, useUserStore } from '@/stores'

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const siteStore = useSiteStore()
  const userStore = useUserStore()

  if (!siteStore.loaded) {
    void siteStore.loadSite()
  }

  if (to.meta.requiresAuth && !userStore.isLogin) {
    ElMessage.warning('请先登录后再操作')
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (userStore.isLogin && !userStore.profile) {
    void userStore.loadProfile()
  }

  if (to.meta.requiresAdmin) {
    if (!userStore.profile) {
      await userStore.loadProfile()
    }
    if (!userStore.isAdmin) {
      ElMessage.error('需要管理员权限')
      return { name: 'home' }
    }
  }

  return true
})

router.afterEach((to) => {
  const siteStore = useSiteStore()
  const base = siteStore.title
  document.title = to.meta.title ? `${to.meta.title} · ${base}` : base
})

export default router
