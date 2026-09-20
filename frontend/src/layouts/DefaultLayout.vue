<template>
  <div class="default-layout">
    <CyberBackdrop />
    <AppHeader />
    <main class="default-layout__main">
      <router-view v-slot="{ Component }">
        <transition name="sd-fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
    <AppFooter />

    <div class="default-layout__float">
      <el-tooltip content="AI 助手" placement="left">
        <el-button class="default-layout__float-btn default-layout__float-btn--ai" circle @click="aiOpen = true">
          <el-icon :size="18"><MagicStick /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip content="回到顶部" placement="left">
        <el-button class="default-layout__float-btn" circle @click="scrollTop">
          <el-icon :size="18"><Top /></el-icon>
        </el-button>
      </el-tooltip>
    </div>

    <el-drawer v-model="aiOpen" title="AI 助手" size="440px" append-to-body class="ai-drawer">
      <AiChatPanel compact />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { MagicStick, Top } from '@element-plus/icons-vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
import CyberBackdrop from '@/components/layout/CyberBackdrop.vue'
import AiChatPanel from '@/components/ai/AiChatPanel.vue'
import { useMessageStore, useUserStore } from '@/stores'

const userStore = useUserStore()
const messageStore = useMessageStore()
const aiOpen = ref(false)

function scrollTop(): void {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function syncPolling(): void {
  if (userStore.isLogin) {
    messageStore.startPolling()
  } else {
    messageStore.stopPolling()
    messageStore.clear()
  }
}

onMounted(syncPolling)
watch(() => userStore.isLogin, syncPolling)

onBeforeUnmount(() => {
  messageStore.stopPolling()
})
</script>

<style scoped lang="scss">
.default-layout {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.default-layout__main {
  position: relative;
  z-index: 1;
  flex: 1;
}

.default-layout__float {
  position: fixed;
  right: 22px;
  bottom: 40px;
  z-index: 25;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.default-layout__float-btn {
  width: 42px;
  height: 42px;
  border: 1px solid var(--sd-border-strong);
  background: rgba(14, 21, 36, 0.92);
  color: var(--sd-text-muted);
  backdrop-filter: blur(8px);
}

.default-layout__float-btn:hover {
  color: var(--sd-cyan);
  border-color: rgba(34, 211, 238, 0.7);
  box-shadow: 0 0 18px -6px rgba(34, 211, 238, 0.9);
}

.default-layout__float-btn--ai {
  color: #f0abfc;
  border-color: rgba(168, 85, 247, 0.6);
  box-shadow: 0 0 20px -8px rgba(168, 85, 247, 0.95);
}

.default-layout__float-btn--ai:hover {
  color: #fae8ff;
  border-color: rgba(232, 121, 249, 0.9);
}

:deep(.ai-drawer .el-drawer__body) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-top: 8px;
}
</style>
