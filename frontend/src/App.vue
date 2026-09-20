<template>
  <router-view v-slot="{ Component }">
    <transition name="sd-fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useSiteStore, useUserStore } from '@/stores'

const siteStore = useSiteStore()
const userStore = useUserStore()

onMounted(async () => {
  await siteStore.loadSite()
  if (userStore.isLogin) {
    void userStore.loadProfile()
  }
})
</script>
