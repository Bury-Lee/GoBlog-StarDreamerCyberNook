<template>
  <div class="not-found sd-grid-bg">
    <div class="not-found__code sd-neon-text">404</div>
    <h1 class="not-found__title">信号丢失 · 页面不存在</h1>
    <p class="sd-dim">你访问的路径不在当前星图中,可能是链接失效或页面已迁移</p>
    <div class="not-found__actions">
      <el-button type="primary" @click="router.push({ name: 'home' })">返回首页</el-button>
      <el-button @click="router.back()">返回上一页</el-button>
      <el-button text @click="router.push({ name: 'articles' })">浏览文章</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped lang="scss">
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 100vh;
  padding: 40px 20px;
  text-align: center;
}

.not-found__code {
  font-size: 96px;
  font-weight: 900;
  line-height: 1;
  font-family: var(--sd-font-mono);
}

.not-found__title {
  font-size: 22px;
}

.not-found__actions {
  display: flex;
  gap: 12px;
  margin-top: 10px;
  flex-wrap: wrap;
  justify-content: center;
}
</style>
