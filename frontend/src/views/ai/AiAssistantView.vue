<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="ai-layout">
        <aside class="ai-side">
          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">AI 助手</span>
            </header>
            <div class="sd-panel__body ai-side__intro">
              <el-icon :size="34" class="ai-side__icon"><MagicStick /></el-icon>
              <p class="ai-side__title sd-neon-text">赛博空间智能助手</p>
              <p class="sd-dim">
                基于站点配置的本地大模型,支持多轮对话。若后端未开启 AI,接口会返回 404。
              </p>
            </div>
          </section>

          <section class="sd-panel">
            <header class="sd-panel__header">
              <span class="sd-panel__title">能做什么</span>
            </header>
            <div class="sd-panel__body">
              <ul class="ai-side__list">
                <li v-for="item in abilities" :key="item">{{ item }}</li>
              </ul>
            </div>
          </section>
        </aside>

        <section class="sd-panel ai-main">
          <header class="sd-panel__header">
            <span class="sd-panel__title">对话窗口</span>
            <span class="sd-chip">POST /api/chat</span>
          </header>
          <div class="ai-main__body">
            <AiChatPanel />
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { MagicStick } from '@element-plus/icons-vue'
import AiChatPanel from '@/components/ai/AiChatPanel.vue'

const abilities = [
  '解释技术概念与代码思路',
  '帮助梳理文章结构与摘要',
  '提供写作选题与内容灵感',
  '辅助排查报错与调试方向',
]
</script>

<style scoped lang="scss">
.ai-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.ai-side {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.ai-side__intro {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  text-align: center;
}

.ai-side__icon {
  color: var(--sd-cyan);
  animation: sd-float 3s ease-in-out infinite;
}

.ai-side__title {
  font-size: 16px;
  font-weight: 700;
}

.ai-side__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;
  color: var(--sd-text-muted);

  li::before {
    content: '▹ ';
    color: var(--sd-purple);
  }
}

.ai-main {
  display: flex;
  flex-direction: column;
  min-height: 620px;
}

.ai-main__body {
  flex: 1;
  min-height: 0;
  display: flex;
  padding: 16px 18px;
}

@media (max-width: 1000px) {
  .ai-layout {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
