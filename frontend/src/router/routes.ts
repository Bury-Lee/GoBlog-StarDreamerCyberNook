import type { RouteRecordRaw } from 'vue-router'

const DefaultLayout = () => import('@/layouts/DefaultLayout.vue')
const AdminLayout = () => import('@/layouts/AdminLayout.vue')
const BlankLayout = () => import('@/layouts/BlankLayout.vue')

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: DefaultLayout,
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/views/home/HomeView.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'articles',
        name: 'articles',
        component: () => import('@/views/article/ArticleListView.vue'),
        meta: { title: '文章' },
      },
      {
        path: 'search',
        name: 'search',
        component: () => import('@/views/article/ArticleSearchView.vue'),
        meta: { title: '搜索' },
      },
      {
        path: 'article/:id',
        name: 'article-detail',
        component: () => import('@/views/article/ArticleDetailView.vue'),
        meta: { title: '文章详情' },
      },
      {
        path: 'write',
        name: 'article-create',
        component: () => import('@/views/article/ArticleEditorView.vue'),
        meta: { title: '写文章', requiresAuth: true },
      },
      {
        path: 'article/:id/edit',
        name: 'article-edit',
        component: () => import('@/views/article/ArticleEditorView.vue'),
        meta: { title: '编辑文章', requiresAuth: true },
      },
      {
        path: 'me/articles',
        name: 'my-articles',
        component: () => import('@/views/article/MyArticlesView.vue'),
        meta: { title: '我的文章', requiresAuth: true },
      },
      {
        path: 'u/:id',
        name: 'user-home',
        component: () => import('@/views/user/UserHomeView.vue'),
        meta: { title: '用户主页' },
      },
      {
        path: 'settings',
        name: 'user-settings',
        component: () => import('@/views/user/UserSettingsView.vue'),
        meta: { title: '个人设置', requiresAuth: true },
      },
      {
        path: 'collections',
        name: 'collections',
        component: () => import('@/views/user/CollectionsView.vue'),
        meta: { title: '我的收藏', requiresAuth: true },
      },
      {
        path: 'history',
        name: 'history',
        component: () => import('@/views/user/HistoryView.vue'),
        meta: { title: '浏览记录', requiresAuth: true },
      },
      {
        path: 'messages',
        name: 'messages',
        component: () => import('@/views/message/MessageCenterView.vue'),
        meta: { title: '消息中心', requiresAuth: true },
      },
      {
        path: 'chat',
        name: 'chat',
        component: () => import('@/views/chat/ChatView.vue'),
        meta: { title: '私聊', requiresAuth: true },
      },
      {
        path: 'about',
        name: 'about',
        component: () => import('@/views/about/AboutView.vue'),
        meta: { title: '关于' },
      },
      {
        path: 'feedback',
        name: 'feedback',
        component: () => import('@/views/feedback/FeedbackWallView.vue'),
        meta: { title: '功能反馈' },
      },
    ],
  },
  {
    path: '/auth',
    component: BlankLayout,
    children: [
      {
        path: 'login',
        name: 'login',
        component: () => import('@/views/auth/LoginView.vue'),
        meta: { title: '登录' },
      },
      {
        path: 'register',
        name: 'register',
        component: () => import('@/views/auth/RegisterView.vue'),
        meta: { title: '注册' },
      },
    ],
  },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'admin-dashboard',
        component: () => import('@/views/admin/AdminDashboardView.vue'),
        meta: { title: '控制台' },
      },
      {
        path: 'review',
        name: 'admin-review',
        component: () => import('@/views/admin/AdminReviewView.vue'),
        meta: { title: '文章审核' },
      },
      {
        path: 'articles',
        name: 'admin-articles',
        component: () => import('@/views/admin/AdminArticlesView.vue'),
        meta: { title: '文章管理' },
      },
      {
        path: 'categories',
        name: 'admin-categories',
        component: () => import('@/views/admin/AdminCategoriesView.vue'),
        meta: { title: '分类管理' },
      },
      {
        path: 'users',
        name: 'admin-users',
        component: () => import('@/views/admin/AdminUsersView.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'banners',
        name: 'admin-banners',
        component: () => import('@/views/admin/AdminBannersView.vue'),
        meta: { title: '轮播图管理' },
      },
      {
        path: 'friend-links',
        name: 'admin-friend-links',
        component: () => import('@/views/admin/AdminFriendLinksView.vue'),
        meta: { title: '友情链接' },
      },
      {
        path: 'promotions',
        name: 'admin-promotions',
        component: () => import('@/views/admin/AdminPromotionsView.vue'),
        meta: { title: '友站推广' },
      },
      {
        path: 'images',
        name: 'admin-images',
        component: () => import('@/views/admin/AdminImagesView.vue'),
        meta: { title: '图片管理' },
      },
      {
        path: 'logs',
        name: 'admin-logs',
        component: () => import('@/views/admin/AdminLogsView.vue'),
        meta: { title: '日志管理' },
      },
      {
        path: 'site',
        name: 'admin-site',
        component: () => import('@/views/admin/AdminSiteConfigView.vue'),
        meta: { title: '站点配置' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/error/NotFoundView.vue'),
    meta: { title: '页面不存在' },
  },
]
