<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="sd-layout-3col">
        <aside class="sd-panel message-side">
          <header class="sd-panel__header">
            <span class="sd-panel__title">消息中心</span>
            <el-button text size="small" @click="refreshAll">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </header>
          <div class="message-side__body">
            <div
              v-for="item in MESSAGE_TYPES"
              :key="item.value"
              class="message-side__item"
              :class="{ 'is-active': type === item.value }"
              @click="switchType(item.value)"
            >
              <el-icon><component :is="item.icon" /></el-icon>
              <span class="message-side__label">{{ item.label }}</span>
              <el-badge
                v-if="messageStore.countOf(item.value) > 0"
                :value="messageStore.countOf(item.value)"
                :max="99"
              />
            </div>
          </div>
          <div class="message-side__footer">
            <el-button size="small" @click="clearVisible = true">一键已读</el-button>
          </div>
        </aside>

        <main class="sd-panel message-main">
          <header class="sd-panel__header">
            <span class="sd-panel__title">{{ currentTypeLabel }}</span>
            <div class="message-main__tools">
              <el-button
                size="small"
                type="danger"
                plain
                :disabled="!selection.length"
                @click="removeSelected"
              >
                删除选中({{ selection.length }})
              </el-button>
            </div>
          </header>
          <div class="sd-panel__body">
            <div v-loading="loading" class="message-main__list">
              <EmptyState v-if="!loading && !list.length" text="暂时没有消息" />
              <article
                v-for="item in list"
                :key="item.id"
                class="message-item"
                :class="{ 'is-unread': !item.isRead }"
                @click="openMessage(item)"
              >
                <el-checkbox
                  class="message-item__check"
                  :model-value="selectedIds.includes(item.id)"
                  @click.stop
                  @change="() => toggleSelect(item.id)"
                />
                <UserAvatar :src="item.actionUserAvatar" :name="item.actionUserNickname" :size="38" />
                <div class="message-item__main">
                  <div class="message-item__head">
                    <span class="message-item__title">{{ item.title || messageTypeLabel(item.type) }}</span>
                    <span class="sd-dim message-item__time">{{ fromNow(item.createdAt) }}</span>
                  </div>
                  <p class="message-item__content sd-clamp-2">{{ item.content }}</p>
                  <div class="message-item__extra">
                    <span v-if="item.actionUserNickname" class="sd-chip">
                      {{ item.actionUserNickname }}
                    </span>
                    <span v-if="item.articleTitle" class="sd-chip">文章:{{ item.articleTitle }}</span>
                    <span v-if="item.linkTitle" class="sd-chip">{{ item.linkTitle }}</span>
                  </div>
                </div>
                <span v-if="!item.isRead" class="message-item__dot" />
              </article>
            </div>

            <PaginationBar
              :page="page"
              :limit="limit"
              :count="count"
              layout="prev, pager, next"
              @update:page="changePage"
            />
          </div>
        </main>
      </div>
    </div>

    <el-dialog v-model="clearVisible" title="一键已读" width="440px">
      <div class="clear-list">
        <el-checkbox v-model="clearForm.commentMessage">评论 / 回复 / @我的消息</el-checkbox>
        <el-checkbox v-model="clearForm.diggAndCollectMessage">点赞 / 收藏消息</el-checkbox>
        <el-checkbox v-model="clearForm.privateMessage">私信消息</el-checkbox>
        <el-checkbox v-model="clearForm.systemMessage">系统通知</el-checkbox>
      </div>
      <template #footer>
        <el-button @click="clearVisible = false">取消</el-button>
        <el-button type="primary" :loading="clearing" @click="submitClear">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { clearMessages, fetchMessages, removeMessages } from '@/api/message'
import type { MessageModel } from '@/api/types'
import { MESSAGE_TYPES, fromNow, messageTypeLabel } from '@/utils/format'
import { useMessageStore } from '@/stores'

const router = useRouter()
const messageStore = useMessageStore()

const type = ref(1)
const list = ref<MessageModel[]>([])
const count = ref(0)
const page = ref(1)
const limit = ref(10)
const loading = ref(false)
const selection = ref<number[]>([])
const clearVisible = ref(false)
const clearing = ref(false)

const clearForm = reactive({
  commentMessage: false,
  diggAndCollectMessage: false,
  privateMessage: false,
  systemMessage: false,
})

const currentTypeLabel = computed(() => messageTypeLabel(type.value))
const selectedIds = computed(() => selection.value)

async function load(): Promise<void> {
  loading.value = true
  try {
    const data = await fetchMessages({ type: type.value, page: page.value, limit: limit.value })
    list.value = data?.list ?? []
    count.value = data?.count ?? 0
    void messageStore.refresh()
  } catch {
    list.value = []
    count.value = 0
  } finally {
    loading.value = false
  }
}

function switchType(next: number): void {
  if (type.value === next) return
  type.value = next
  page.value = 1
  selection.value = []
  void load()
}

function changePage(next: number): void {
  page.value = next
  void load()
}

function toggleSelect(id: number): void {
  if (selection.value.includes(id)) {
    selection.value = selection.value.filter((item) => item !== id)
  } else {
    selection.value = [...selection.value, id]
  }
}

function openMessage(item: MessageModel): void {
  if (item.articleID) {
    router.push({ name: 'article-detail', params: { id: item.articleID } })
    return
  }
  if (item.type === 5) {
    router.push({ name: 'chat', query: { userID: String(item.ActionUserID) } })
    return
  }
  if (item.linkHref) {
    window.open(item.linkHref, '_blank', 'noopener')
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  if (selection.value.length > 100) {
    ElMessage.warning('一次最多删除 100 条')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 条消息?`, '删除消息', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeMessages(selection.value)
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}

async function submitClear(): Promise<void> {
  clearing.value = true
  try {
    await clearMessages({ ...clearForm })
    ElMessage.success('操作成功')
    clearVisible.value = false
    await load()
  } catch {
    // ignore
  } finally {
    clearing.value = false
  }
}

function refreshAll(): void {
  void load()
  void messageStore.refresh()
}

onMounted(load)
</script>

<style scoped lang="scss">
.message-side {
  padding-bottom: 10px;
}

.message-side__body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px;
}

.message-side__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--sd-text-muted);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.message-side__item:hover {
  background: rgba(34, 211, 238, 0.08);
  color: var(--sd-text);
}

.message-side__item.is-active {
  color: var(--sd-cyan);
  background: linear-gradient(92deg, rgba(34, 211, 238, 0.16), rgba(168, 85, 247, 0.12));
  box-shadow: inset 0 0 0 1px rgba(34, 211, 238, 0.35);
}

.message-side__label {
  flex: 1;
}

.message-side__footer {
  padding: 0 14px 10px;
}

.message-main__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 200px;
}

.message-item {
  position: relative;
  display: flex;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.55);
  cursor: pointer;
  transition: all 0.2s ease;
}

.message-item:hover {
  border-color: rgba(34, 211, 238, 0.45);
}

.message-item.is-unread {
  background: rgba(34, 211, 238, 0.06);
  border-color: rgba(34, 211, 238, 0.3);
}

.message-item__check {
  margin-right: 2px;
}

.message-item__main {
  flex: 1;
  min-width: 0;
}

.message-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.message-item__title {
  font-size: 14px;
  font-weight: 600;
}

.message-item__time {
  font-size: 12px;
  flex-shrink: 0;
}

.message-item__content {
  margin-top: 4px;
  font-size: 13px;
  color: var(--sd-text-muted);
  line-height: 1.7;
}

.message-item__extra {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
  font-size: 12px;
}

.message-item__dot {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--sd-red);
  box-shadow: 0 0 10px var(--sd-red);
}

.clear-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
