<template>
  <div class="sd-container feedback-view">
    <header class="feedback-view__head">
      <h1 class="feedback-view__title">功能反馈</h1>
      <p class="sd-dim">欢迎提交功能建议或问题反馈，也可以匿名。所有反馈公开展示。</p>
    </header>

    <section class="sd-panel feedback-view__form">
      <div class="sd-panel__header">
        <span class="sd-panel__title">提交反馈</span>
      </div>
      <el-form label-width="80px">
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 160px">
            <el-option :value="1" label="功能建议" />
            <el-option :value="2" label="问题反馈" />
            <el-option :value="3" label="内容举报" />
            <el-option :value="0" label="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="4"
            maxlength="2000"
            show-word-limit
            resize="none"
            placeholder="请描述你的建议或遇到的问题"
          />
        </el-form-item>
        <el-form-item label="联系方式">
          <el-input v-model="form.contact" maxlength="128" placeholder="可选，便于我们回访" />
        </el-form-item>
        <el-form-item label="匿名">
          <el-switch v-model="form.isAnonymous" />
          <span class="sd-dim" style="margin-left: 8px">匿名后不展示提交者</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="submit">提交反馈</el-button>
        </el-form-item>
      </el-form>
    </section>

    <section class="feedback-view__list">
      <div class="sd-panel__header">
        <span class="sd-panel__title">反馈列表 <span class="sd-dim">({{ countText }})</span></span>
      </div>

      <EmptyState v-if="!loading && list.length === 0" text="还没有反馈，来发第一条吧" />
      <ul v-else class="feedback-list">
        <li v-for="item in list" :key="item.id" class="feedback-item sd-panel">
          <div class="feedback-item__head">
            <el-tag size="small" :type="typeTag(item.type)">{{ typeText(item.type) }}</el-tag>
            <el-tag size="small" effect="plain" :type="statusTag(item.status)">{{ statusText(item.status) }}</el-tag>
            <span class="sd-dim feedback-item__meta">
              {{ submitterText(item) }} · {{ formatDate(item.createdAt) }}
            </span>
            <el-button v-if="userStore.isAdmin" class="feedback-item__op" size="small" text @click="openHandle(item)">
              处理
            </el-button>
          </div>
          <p class="feedback-item__content">{{ item.content }}</p>
          <div v-if="item.reply" class="feedback-item__reply">官方回复：{{ item.reply }}</div>
        </li>
      </ul>

      <PaginationBar
        v-model:page="page"
        v-model:limit="limit"
        :count="count"
        @update:page="load"
        @update:limit="onLimit"
      />
    </section>

    <el-dialog v-model="handleVisible" title="处理反馈" width="480px">
      <el-form label-width="64px">
        <el-form-item label="状态">
          <el-select v-model="handleForm.status" style="width: 180px">
            <el-option :value="0" label="待处理" />
            <el-option :value="1" label="已采纳未处理" />
            <el-option :value="2" label="正在处理" />
            <el-option :value="3" label="已处理" />
          </el-select>
        </el-form-item>
        <el-form-item label="回复">
          <el-input v-model="handleForm.reply" type="textarea" :rows="3" maxlength="2000" show-word-limit resize="none" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleVisible = false">取消</el-button>
        <el-button type="primary" :loading="handling" @click="confirmHandle">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { createFeedback, fetchFeedbackList, handleFeedback } from '@/api/feedback'
import type { FeedbackItem, FeedbackStatus, FeedbackType } from '@/api/types'
import { formatDate } from '@/utils/format'
import { useUserStore } from '@/stores'

const userStore = useUserStore()

const list = ref<FeedbackItem[]>([])
const count = ref(0)
const capped = ref(false)
const loading = ref(false)
const page = ref(1)
const limit = ref(10)

const submitting = ref(false)
const form = reactive({
  content: '',
  contact: '',
  type: 1 as FeedbackType,
  isAnonymous: false,
})

const handleVisible = ref(false)
const handling = ref(false)
const handleForm = reactive({ id: 0, status: 0 as FeedbackStatus, reply: '' })

const countText = computed(() => (capped.value ? `>${count.value}` : String(count.value)))

async function load(): Promise<void> {
  loading.value = true
  try {
    const data = await fetchFeedbackList({ page: page.value, limit: limit.value })
    list.value = data.list
    count.value = data.count
    capped.value = Boolean(data.capped)
  } catch {
    list.value = []
    count.value = 0
  } finally {
    loading.value = false
  }
}

function onLimit(): void {
  page.value = 1
  void load()
}

async function submit(): Promise<void> {
  if (!form.content.trim()) {
    ElMessage.warning('请输入反馈内容')
    return
  }
  submitting.value = true
  try {
    await createFeedback({
      content: form.content.trim(),
      contact: form.contact.trim(),
      type: form.type,
      isAnonymous: form.isAnonymous,
    })
    ElMessage.success('感谢反馈，我们会尽快处理')
    form.content = ''
    form.contact = ''
    page.value = 1
    await load()
  } catch {
    // ignore
  } finally {
    submitting.value = false
  }
}

function openHandle(item: FeedbackItem): void {
  handleForm.id = item.id
  handleForm.status = item.status
  handleForm.reply = item.reply || ''
  handleVisible.value = true
}

async function confirmHandle(): Promise<void> {
  handling.value = true
  try {
    const updated = await handleFeedback(handleForm.id, { status: handleForm.status, reply: handleForm.reply })
    // 增量替换列表中对应项
    const index = list.value.findIndex((i) => i.id === updated.id)
    if (index >= 0) list.value.splice(index, 1, updated)
    handleVisible.value = false
    ElMessage.success('处理成功')
  } catch {
    // ignore
  } finally {
    handling.value = false
  }
}

function submitterText(item: FeedbackItem): string {
  if (item.isAnonymous) return '匿名'
  if (item.userID) return `用户 #${item.userID}`
  return '游客'
}

function typeText(type: FeedbackType): string {
  return { 0: '其他', 1: '功能建议', 2: '问题反馈', 3: '内容举报' }[type] || '其他'
}

function typeTag(type: FeedbackType): 'primary' | 'success' | 'warning' | 'info' | 'danger' {
  return { 0: 'info', 1: 'primary', 2: 'warning', 3: 'danger' }[type] as
    | 'primary'
    | 'success'
    | 'warning'
    | 'info'
    | 'danger'
}

function statusText(status: FeedbackStatus): string {
  return { 0: '待处理', 1: '已采纳', 2: '处理中', 3: '已处理' }[status] || '待处理'
}

function statusTag(status: FeedbackStatus): 'primary' | 'success' | 'warning' | 'info' | 'danger' {
  return { 0: 'info', 1: 'primary', 2: 'warning', 3: 'success' }[status] as
    | 'primary'
    | 'success'
    | 'warning'
    | 'info'
    | 'danger'
}

onMounted(load)
</script>

<style scoped lang="scss">
.feedback-view {
  padding: 20px 0 40px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.feedback-view__head {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.feedback-view__title {
  margin: 0;
  font-size: 22px;
}

.feedback-view__form {
  padding: 16px 18px;
}

.feedback-list {
  list-style: none;
  margin: 12px 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.feedback-item {
  padding: 14px 16px;
}

.feedback-item__head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.feedback-item__meta {
  font-size: 12px;
}

.feedback-item__op {
  margin-left: auto;
}

.feedback-item__content {
  margin: 10px 0 0;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.feedback-item__reply {
  margin-top: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.7;
  background: rgba(34, 211, 238, 0.08);
  border: 1px solid rgba(34, 211, 238, 0.25);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
