<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="chat-layout">
        <aside class="sd-panel chat-side">
          <header class="sd-panel__header">
            <span class="sd-panel__title">会话列表</span>
            <el-button text size="small" @click="loadSessions">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </header>
          <div v-loading="sessionLoading" class="chat-side__body">
            <EmptyState v-if="!sessions.length && !sessionLoading" text="暂无会话" compact />
            <div
              v-for="session in sessions"
              :key="session.id"
              class="chat-side__item"
              :class="{ 'is-active': session.peerID === peerID }"
              @click="selectPeer(session.peerID)"
            >
              <el-badge :value="session.unreadCount" :hidden="!session.unreadCount" :max="99">
                <UserAvatar :src="peerInfo(session.peerID).avatar" :name="peerInfo(session.peerID).nickname" :size="38" />
              </el-badge>
              <div class="chat-side__text">
                <span class="chat-side__name sd-ellipsis">
                  {{ peerLabel(session.peerID) }}
                </span>
                <span class="sd-dim chat-side__time">{{ fromNow(session.lastMessageTime) }}</span>
              </div>
            </div>
          </div>
        </aside>

        <main class="sd-panel chat-main">
          <header class="sd-panel__header">
            <span class="sd-panel__title">
              {{ peerID ? peerLabel(peerID) : '选择会话开始聊天' }}
            </span>
            <div v-if="peerID" class="chat-main__peer">
              <el-button text size="small" @click="goPeerHome">查看主页</el-button>
            </div>
          </header>

          <div ref="scrollRef" v-loading="messageLoading" class="chat-main__body">
            <EmptyState v-if="!peerID" text="从左侧选择一个会话" />
            <EmptyState v-else-if="!messages.length && !messageLoading" text="还没有聊天记录,打个招呼吧" />
            <div v-if="peerID && hasMore" class="chat-main__earlier">
              <el-button text size="small" :loading="loadingEarlier" @click="loadEarlier">
                加载更早的消息
              </el-button>
            </div>
            <div
              v-for="message in messages"
              :key="message.id"
              class="chat-bubble"
              :class="{ 'chat-bubble--me': message.isMe }"
            >
              <UserAvatar
                :src="message.isMe ? message.sendUserAvatar : message.revUserAvatar"
                :name="message.isMe ? message.sendUserNickname : message.revUserNickname"
                :size="34"
              />
              <div class="chat-bubble__body">
                <div class="chat-bubble__meta sd-dim">
                  {{ message.isMe ? '我' : message.revUserNickname }} · {{ fromNow(message.createdAt) }}
                </div>
                <div class="chat-bubble__content">
                  <img
                    v-if="message.msg?.imageMsg?.src"
                    class="chat-bubble__image"
                    :src="resolveAssetUrl(message.msg.imageMsg.src)"
                    alt="image"
                  />
                  <p v-if="message.msg?.textMsg?.content" class="chat-bubble__text">
                    {{ message.msg.textMsg.content }}
                  </p>
                  <div
                    v-if="message.msg?.markdownMsg?.content"
                    class="chat-bubble__markdown"
                    v-html="renderMarkdown(message.msg.markdownMsg.content)"
                  />
                </div>
              </div>
            </div>
          </div>

          <footer v-if="peerID" class="chat-main__footer">
            <el-input
              v-model="draft"
              type="textarea"
              :rows="3"
              resize="none"
              maxlength="1000"
              placeholder="输入消息,Enter 发送 / Shift + Enter 换行(仅好友之间可发送)"
              @keydown.enter.exact.prevent="send"
            />
            <div class="chat-main__actions">
              <div class="chat-main__left">
                <el-checkbox v-model="asMarkdown">Markdown</el-checkbox>
                <el-button size="small" :loading="uploading" @click="pickImage">
                  <el-icon><Picture /></el-icon>
                  发送图片
                </el-button>
              </div>
              <el-button type="primary" :loading="sending" :disabled="!draft.trim()" @click="send">
                发送
              </el-button>
            </div>
          </footer>
        </main>
      </div>

      <input
        ref="fileInputRef"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp,image/bmp,image/tiff"
        class="chat-file"
        @change="onFileChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture, Refresh } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchChatHistory, fetchChatSessions, sendChatMessage } from '@/api/chat'
import { uploadImage } from '@/api/ops'
import { imageUrl, resolveAssetUrl } from '@/api/request'
import { fetchUserBaseInfo } from '@/api/user'
import type { ChatItem, ChatSession } from '@/api/types'
import { fromNow } from '@/utils/format'
import { renderMarkdown } from '@/utils/html'

interface SessionView extends ChatSession {
  peerID: number
}

const route = useRoute()
const router = useRouter()

const sessions = ref<SessionView[]>([])
const sessionLoading = ref(false)
const messages = ref<ChatItem[]>([])
const messageLoading = ref(false)
const peerID = ref(0)
const draft = ref('')
const sending = ref(false)
const uploading = ref(false)
const asMarkdown = ref(false)
const scrollRef = ref<HTMLElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const peerCache = ref<Record<number, { nickname: string; avatar: string }>>({})
const peerStatus = ref<Record<number, 'loading' | 'error'>>({})
const historyPage = ref(1)
const hasMore = ref(false)
const loadingEarlier = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

function peerInfo(id: number): { nickname: string; avatar: string } {
  return peerCache.value[id] || { nickname: '', avatar: '' }
}

function peerLabel(id: number): string {
  const nickname = peerInfo(id).nickname
  if (nickname) return nickname
  return peerStatus.value[id] === 'error' ? '未知用户' : '加载中…'
}

async function ensurePeerInfo(id: number): Promise<void> {
  if (!id || peerCache.value[id]) return
  peerStatus.value = { ...peerStatus.value, [id]: 'loading' }
  try {
    const data = await fetchUserBaseInfo(id)
    peerCache.value = {
      ...peerCache.value,
      [id]: { nickname: data?.nickName || '未知用户', avatar: data?.avatar || '' },
    }
  } catch {
    peerStatus.value = { ...peerStatus.value, [id]: 'error' }
  }
}

async function loadSessions(): Promise<void> {
  sessionLoading.value = true
  try {
    const data = await fetchChatSessions({ page: 1, limit: 40 })
    sessions.value = (data?.list ?? []).map((session) => {
      const parts = String(session.uniqueId || '').split('_')
      const peer = Number(parts[1] || 0)
      return { ...session, peerID: peer || session.userId }
    })
    await Promise.all(sessions.value.map((session) => ensurePeerInfo(session.peerID)))
  } catch {
    sessions.value = []
  } finally {
    sessionLoading.value = false
  }
}

async function loadMessages(silent = false): Promise<void> {
  if (!peerID.value) return
  if (!silent) messageLoading.value = true
  try {
    const data = await fetchChatHistory({ userID: peerID.value, page: 1, limit: 40 })
    const latest = [...(data?.list ?? [])].reverse()
    if (silent) {
      const known = new Set(messages.value.map((item) => item.id))
      const fresh = latest.filter((item) => !known.has(item.id))
      if (!fresh.length) return
      messages.value = [...messages.value, ...fresh]
      hasMore.value = messages.value.length < (data?.count ?? messages.value.length)
      await nextTick()
      scrollToBottom()
      return
    }
    messages.value = latest
    historyPage.value = 1
    hasMore.value = messages.value.length < (data?.count ?? 0)
    await nextTick()
    scrollToBottom()
  } catch {
    if (!silent) messages.value = []
  } finally {
    if (!silent) messageLoading.value = false
  }
}

async function loadEarlier(): Promise<void> {
  if (!peerID.value || loadingEarlier.value) return
  const el = scrollRef.value
  const prevHeight = el?.scrollHeight ?? 0
  const prevTop = el?.scrollTop ?? 0
  loadingEarlier.value = true
  try {
    const next = historyPage.value + 1
    const data = await fetchChatHistory({ userID: peerID.value, page: next, limit: 40 })
    const older = [...(data?.list ?? [])].reverse()
    messages.value = [...older, ...messages.value]
    historyPage.value = next
    hasMore.value = messages.value.length < (data?.count ?? messages.value.length)
    await nextTick()
    if (el) el.scrollTop = el.scrollHeight - prevHeight + prevTop
  } catch {
    ElMessage.warning('加载更早的消息失败,请稍后再试')
  } finally {
    loadingEarlier.value = false
  }
}

function scrollToBottom(): void {
  const el = scrollRef.value
  if (el) el.scrollTop = el.scrollHeight
}

async function selectPeer(id: number): Promise<void> {
  if (!id) return
  peerID.value = id
  await ensurePeerInfo(id)
  await loadMessages()
  await loadSessions()
  void router.replace({ name: 'chat', query: { userID: String(id) } })
}

async function send(): Promise<void> {
  const content = draft.value.trim()
  if (!content || !peerID.value) return
  sending.value = true
  try {
    await sendChatMessage({
      revUserID: peerID.value,
      msg: asMarkdown.value ? { markdownMsg: { content } } : { textMsg: { content } },
    })
    draft.value = ''
    await loadMessages()
    await loadSessions()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '发送失败,仅好友之间可发送消息')
  } finally {
    sending.value = false
  }
}

function pickImage(): void {
  fileInputRef.value?.click()
}

async function onFileChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !peerID.value) return
  uploading.value = true
  try {
    const id = await uploadImage(file)
    await sendChatMessage({
      revUserID: peerID.value,
      msg: { imageMsg: { src: imageUrl(id) } },
    })
    await loadMessages()
  } catch {
    // ignore
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function goPeerHome(): void {
  router.push({ name: 'user-home', params: { id: peerID.value } })
}

onMounted(async () => {
  await loadSessions()
  const queryPeer = Number(route.query.userID || 0)
  const target = queryPeer || sessions.value[0]?.peerID || 0
  if (target) await selectPeer(target)
  pollTimer = setInterval(() => {
    if (peerID.value) void loadMessages(true)
  }, 5000)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped lang="scss">
.chat-layout {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.chat-side {
  padding-bottom: 10px;
}

.chat-side__body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px;
  max-height: 620px;
  overflow-y: auto;
}

.chat-side__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.chat-side__item:hover {
  background: rgba(34, 211, 238, 0.08);
}

.chat-side__item.is-active {
  background: linear-gradient(92deg, rgba(34, 211, 238, 0.16), rgba(168, 85, 247, 0.12));
  box-shadow: inset 0 0 0 1px rgba(34, 211, 238, 0.35);
}

.chat-side__text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.chat-side__name {
  font-size: 13px;
}

.chat-side__time {
  font-size: 11px;
}

.chat-main__body {
  height: 520px;
  overflow-y: auto;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: rgba(7, 11, 20, 0.35);
}

.chat-main__earlier {
  display: flex;
  justify-content: center;
}

.chat-bubble {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.chat-bubble--me {
  flex-direction: row-reverse;
}

.chat-bubble__body {
  max-width: 72%;
}

.chat-bubble--me .chat-bubble__body {
  text-align: right;
}

.chat-bubble__meta {
  font-size: 11px;
  margin-bottom: 4px;
}

.chat-bubble__content {
  display: inline-block;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid var(--sd-border);
  background: rgba(18, 27, 46, 0.9);
  text-align: left;
  font-size: 13.5px;
  line-height: 1.7;
  word-break: break-word;
}

.chat-bubble--me .chat-bubble__content {
  background: linear-gradient(120deg, rgba(34, 211, 238, 0.22), rgba(129, 140, 248, 0.22));
  border-color: rgba(34, 211, 238, 0.4);
}

.chat-bubble__text {
  white-space: pre-wrap;
}

.chat-bubble__image {
  max-width: 220px;
  border-radius: 10px;
}

.chat-bubble__markdown :deep(p) {
  margin: 0 0 6px;
}

.chat-bubble__markdown :deep(pre) {
  padding: 8px;
  overflow-x: auto;
  border-radius: 8px;
  background: rgba(7, 11, 20, 0.85);
}

.chat-main__footer {
  padding: 12px 18px 16px;
  border-top: 1px solid var(--sd-border);
}

.chat-main__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
}

.chat-main__left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.chat-file {
  display: none;
}

@media (max-width: 900px) {
  .chat-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .chat-side__body {
    max-height: 240px;
  }
}
</style>
