<template>
  <div class="app-footer">
    <div class="sd-container">
      <div class="app-footer__inner">
        <div class="app-footer__brand">
          <span class="sd-neon-text app-footer__title">{{ siteStore.title }}</span>
        </div>
        <div class="app-footer__links">
          <router-link class="sd-link" :to="{ name: 'about' }">关于本站</router-link>
          <a
            v-if="siteStore.beian"
            class="sd-link sd-dim"
            href="https://beian.miit.gov.cn/"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ siteStore.beian }}
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSiteStore } from '@/stores'

const siteStore = useSiteStore()
</script>

<style scoped lang="scss">
.app-footer {
  margin-top: 40px;
  padding: 22px 0;
  border-top: 1px solid var(--sd-border);
  background: rgba(7, 11, 20, 0.72);
}

.app-footer__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  font-size: 13px;
}

.app-footer__title {
  font-weight: 600;
}

.app-footer__links {
  display: flex;
  align-items: center;
  gap: 18px;
  flex-wrap: wrap;
}
</style>
