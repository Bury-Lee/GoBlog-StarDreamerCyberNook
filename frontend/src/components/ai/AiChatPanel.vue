<template>
  <div class="ai-panel" :class="{ 'ai-panel--compact': compact }">
    <div ref="scrollRef" class="ai-panel__body">
      <div v-if="!userStore.isLogin" class="ai-panel__hint">
        <el-icon :size="28"><Lock /></el-icon>
        <p>AI 助手需要登录后使用</p>
        <el-button type="primary" size="small" @click="goLogin">去登录</el-button>
      </div>

      <template v-else>
        <div v-if="!messages.length" class="ai-panel__welcome">
          <el-icon :size="30" class="ai-panel__welcome-icon"><MagicStick /></el-icon>
          <p class="sd-neon-text">我是本站 AI 助手</p>
          <p class="sd-dim">可以问我技术问题、文章总结、写作思路等</p>
          <div class="ai-panel__suggests">
            <span v-for="item in suggests" :key="item" class="sd-chip" @click="quickAsk(item)">
              {{ item }}
            </span>
          </div>
        </div>

        <div
          v-for="(message, index) in messages"
          :key="index"
          class="ai-panel__msg"
          :class="message.role === 'user' ? 'ai-panel__msg--user' : 'ai-panel__msg--assistant'"
        >
          <div class="ai-panel__bubble" v-html="render(message.content)" />
        </div>

        <div v-if="loading" class="ai-panel__msg ai-panel__msg--assistant">
          <div class="ai-panel__bubble ai-panel__bubble--loading">
            <span class="ai-panel__dot" />
            <span class="ai-panel__dot" />
            <span class="ai-panel__dot" />
          </div>
        </div>
      </template>
    </div>

    <div v-if="userStore.isLogin" class="ai-panel__footer">
      <el-input
        v-model="input"
        type="textarea"
        :rows="compact ? 2 : 3"
        resize="none"
        maxlength="2000"
        placeholder="输入你的问题,Enter 发送 / Shift + Enter 换行"
        @keydown.enter.exact.prevent="send"
      />
      <div class="ai-panel__actions">
        <span class="sd-dim">{{ messages.length }} 轮对话</span>
        <div>
          <el-button text :disabled="!messages.length || loading" @click="clear">清空</el-button>
          <el-button type="primary" :loading="loading" :disabled="!input.trim()" @click="send">
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, MagicStick } from '@element-plus/icons-vue'
import { askAi } from '@/api/chat'
import type { AiMessage } from '@/api/types'
import { renderMarkdown } from '@/utils/html'
import { useUserStore } from '@/stores'

withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })

const router = useRouter()
const userStore = useUserStore()

const messages = ref<AiMessage[]>([])
const input = ref('')
const loading = ref(false)
const scrollRef = ref<HTMLElement | null>(null)

const suggests = ['帮我总结一篇文章的要点', 'Go 语言的并发模型怎么理解?', '给我一个技术博客选题']

function render(content: string): string {
  return renderMarkdown(content)
}

async function scrollToBottom(): Promise<void> {
  await nextTick()
  const el = scrollRef.value
  if (el) el.scrollTop = el.scrollHeight
}

function goLogin(): void {
  router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
}

function quickAsk(text: string): void {
  input.value = text
  void send()
}

function clear(): void {
  messages.value = []
}

async function send(): Promise<void> {
  const content = input.value.trim()
  if (!content || loading.value) return
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录后再使用 AI 助手')
    return
  }

  const history = messages.value.slice(-10).map((item) => ({ role: item.role, content: item.content }))
  messages.value.push({ role: 'user', content })
  input.value = ''
  loading.value = true
  void scrollToBottom()

  try {
    const result = await askAi({ messages: history, user_input: content })
    if (result?.success && result.content) {
      messages.value.push({ role: 'assistant', content: result.content })
    } else {
      messages.value.push({
        role: 'assistant',
        content: result?.error || 'AI 暂时没有返回内容,请稍后再试',
      })
    }
  } catch {
    messages.value.push({ role: 'assistant', content: 'AI 服务暂时不可用,请稍后再试' })
  } finally {
    loading.value = false
    void scrollToBottom()
  }
}
</script>

<style scoped lang="scss">
.ai-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.ai-panel__body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 6px 4px 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-panel__hint,
.ai-panel__welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px 12px;
  text-align: center;
  color: var(--sd-text-muted);
}

.ai-panel__welcome-icon {
  color: var(--sd-cyan);
  animation: sd-float 3s ease-in-out infinite;
}

.ai-panel__suggests {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  margin-top: 6px;
}

.ai-panel__suggests .sd-chip {
  cursor: pointer;
}

.ai-panel__suggests .sd-chip:hover {
  border-color: rgba(34, 211, 238, 0.6);
  color: var(--sd-text);
}

.ai-panel__msg {
  display: flex;
}

.ai-panel__msg--user {
  justify-content: flex-end;
}

.ai-panel__msg--assistant {
  justify-content: flex-start;
}

.ai-panel__bubble {
  max-width: 86%;
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;

  :deep(p) {
    margin: 0 0 6px;
  }

  :deep(p:last-child) {
    margin-bottom: 0;
  }

  :deep(pre) {
    padding: 10px;
    overflow-x: auto;
    border-radius: 8px;
    background: rgba(7, 11, 20, 0.85);
    border: 1px solid var(--sd-border);
  }

  :deep(code) {
    font-family: var(--sd-font-mono);
    font-size: 12px;
  }

  :deep(a) {
    color: var(--sd-cyan);
  }
}

.ai-panel__msg--user .ai-panel__bubble {
  background: linear-gradient(120deg, rgba(34, 211, 238, 0.22), rgba(129, 140, 248, 0.22));
  border: 1px solid rgba(34, 211, 238, 0.4);
  border-bottom-right-radius: 4px;
}

.ai-panel__msg--assistant .ai-panel__bubble {
  background: rgba(18, 27, 46, 0.85);
  border: 1px solid var(--sd-border);
  border-bottom-left-radius: 4px;
}

.ai-panel__bubble--loading {
  display: flex;
  gap: 5px;
  align-items: center;
}

.ai-panel__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--sd-cyan);
  animation: sd-pulse 1.1s ease-in-out infinite;
}

.ai-panel__dot:nth-child(2) {
  animation-delay: 0.2s;
}

.ai-panel__dot:nth-child(3) {
  animation-delay: 0.4s;
}

.ai-panel__footer {
  border-top: 1px solid var(--sd-border);
  padding-top: 12px;
}

.ai-panel__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 12px;
}

.ai-panel--compact .ai-panel__bubble {
  max-width: 92%;
}
</style>
