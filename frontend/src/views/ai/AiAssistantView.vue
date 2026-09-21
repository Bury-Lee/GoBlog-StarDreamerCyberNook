<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="sd-panel ai-main">
        <header class="sd-panel__header">
          <span class="sd-panel__title">AI 助手</span>
        </header>
        <div class="ai-main__body">
          <AiChatPanel />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import AiChatPanel from '@/components/ai/AiChatPanel.vue'
</script>

<style scoped lang="scss">
.ai-main {
  display: flex;
  flex-direction: column;
  min-height: 640px;
}

.ai-main__body {
  flex: 1;
  min-height: 0;
  display: flex;
  padding: 16px 18px;
}
</style>
