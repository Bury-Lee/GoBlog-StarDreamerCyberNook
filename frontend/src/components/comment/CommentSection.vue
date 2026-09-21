<template>
  <section class="comment-section sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">评论区</span>
      <span class="sd-chip">{{ count }} 条评论</span>
    </header>

    <div class="sd-panel__body">
      <div v-if="!openComment" class="comment-section__closed">
        <el-icon><Lock /></el-icon>
        <span>该文章已关闭评论</span>
      </div>

      <template v-else>
        <div v-if="!userStore.isLogin" class="comment-section__login-tip">
          <span>登录后即可参与评论</span>
          <el-button type="primary" size="small" @click="goLogin">去登录</el-button>
        </div>

        <div v-else class="comment-section__composer">
          <div v-if="replyTo" class="comment-section__reply-tip">
            <span>回复 <b>{{ replyTo.user?.nickname || '匿名用户' }}</b> 的评论</span>
            <el-button text size="small" @click="replyTo = null">取消</el-button>
          </div>
          <el-input
            v-model="content"
            type="textarea"
            :rows="3"
            resize="none"
            maxlength="500"
            show-word-limit
            :placeholder="replyTo ? '写下你的回复…' : '友善的评论是交流的开始…'"
          />
          <div class="comment-section__composer-actions">
            <el-button type="primary" :loading="submitting" :disabled="!content.trim()" @click="submit">
              {{ replyTo ? '回复' : '发表评论' }}
            </el-button>
          </div>
        </div>

        <div class="sd-divider" />

        <div v-loading="loading" class="comment-section__list">
          <EmptyState v-if="!loading && !list.length" text="还没有评论,来抢沙发吧" />
          <template v-for="comment in list" :key="comment.id">
            <CommentItem
              :comment="comment"
              :current-user-id="userStore.userId"
              :is-admin="userStore.isAdmin"
              @digg="onDigg"
              @reply="onReply"
              @remove="onRemove"
            >
              <div v-if="childState[comment.id]" class="comment-section__children">
                <CommentItem
                  v-for="child in childState[comment.id].list"
                  :key="child.id"
                  :comment="child"
                  is-child
                  :current-user-id="userStore.userId"
                  :is-admin="userStore.isAdmin"
                  @digg="onDigg"
                  @reply="onReply"
                  @remove="onRemove"
                />
                <div v-if="childState[comment.id].count > childState[comment.id].list.length" class="comment-section__more">
                  <el-button text size="small" :loading="childState[comment.id].loading" @click="loadMoreChildren(comment)">
                    查看更多回复
                  </el-button>
                </div>
              </div>
            </CommentItem>
            <div v-if="!childState[comment.id] && comment.diggCount >= 0" class="comment-section__expand">
              <el-button text size="small" @click="loadChildren(comment)">
                <el-icon><ChatLineRound /></el-icon>
                查看回复
              </el-button>
            </div>
          </template>
        </div>

        <PaginationBar
          :page="page"
          :limit="limit"
          :count="count"
          layout="prev, pager, next"
          @update:page="changePage"
        />
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatLineRound, Lock } from '@element-plus/icons-vue'
import CommentItem from './CommentItem.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { createComment, diggComment, fetchChildComments, fetchComments, removeComment } from '@/api/comment'
import type { CommentModel } from '@/api/types'
import { useUserStore } from '@/stores'

const props = withDefaults(
  defineProps<{
    articleId: number
    openComment?: boolean
  }>(),
  { openComment: true },
)

const router = useRouter()
const userStore = useUserStore()

const list = ref<CommentModel[]>([])
const count = ref(0)
const page = ref(1)
const limit = ref(10)
const loading = ref(false)
const content = ref('')
const submitting = ref(false)
const replyTo = ref<CommentModel | null>(null)

interface ChildState {
  list: CommentModel[]
  count: number
  page: number
  loading: boolean
}

const childState = reactive<Record<number, ChildState>>({})

async function load(): Promise<void> {
  if (!props.articleId) return
  loading.value = true
  try {
    const data = await fetchComments({ articleID: props.articleId, page: page.value, limit: limit.value })
    list.value = data?.list ?? []
    count.value = data?.count ?? 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function changePage(next: number): void {
  page.value = next
  void load()
}

function goLogin(): void {
  router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
}

async function submit(): Promise<void> {
  const text = content.value.trim()
  if (!text) return
  if (!userStore.isLogin) {
    goLogin()
    return
  }
  submitting.value = true
  try {
    await createComment({
      articleID: props.articleId,
      content: text,
      parentID: replyTo.value?.id ?? 0,
    })
    ElMessage.success(replyTo.value ? '回复成功' : '评论发表成功')
    const parent = replyTo.value
    content.value = ''
    replyTo.value = null
    if (parent) {
      const rootID = parent.rootParentID ?? parent.id
      delete childState[rootID]
      await loadChildren({ ...parent, id: rootID } as CommentModel)
    } else {
      page.value = 1
      await load()
    }
  } catch {
    // 错误提示已由请求层处理
  } finally {
    submitting.value = false
  }
}

function onReply(comment: CommentModel): void {
  replyTo.value = comment
}

async function onDigg(comment: CommentModel): Promise<void> {
  if (!userStore.isLogin) {
    goLogin()
    return
  }
  try {
    const result = await diggComment(comment.id)
    if (typeof result?.diggCount === 'number') {
      comment.diggCount = result.diggCount
    }
  } catch {
    // ignore
  }
}

async function onRemove(comment: CommentModel): Promise<void> {
  try {
    await ElMessageBox.confirm(
      comment.rootParentID ? '确认删除这条回复?' : '确认删除该评论及其所有回复?',
      '删除评论',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await removeComment(comment.id)
    ElMessage.success('删除成功')
    const rootID = comment.rootParentID ?? comment.id
    delete childState[rootID]
    await load()
  } catch {
    // ignore
  }
}

async function loadChildren(comment: CommentModel): Promise<void> {
  const rootID = comment.rootParentID ?? comment.id
  childState[rootID] = { list: [], count: 0, page: 1, loading: true }
  try {
    const data = await fetchChildComments({ root: rootID, page: 1, limit: 5 })
    childState[rootID] = {
      list: data?.list ?? [],
      count: data?.count ?? 0,
      page: 1,
      loading: false,
    }
  } catch {
    delete childState[rootID]
  }
}

async function loadMoreChildren(comment: CommentModel): Promise<void> {
  const rootID = comment.id
  const state = childState[rootID]
  if (!state) return
  state.loading = true
  try {
    const nextPage = state.page + 1
    const data = await fetchChildComments({ root: rootID, page: nextPage, limit: 5 })
    state.list = [...state.list, ...(data?.list ?? [])]
    state.page = nextPage
    state.count = data?.count ?? state.count
  } finally {
    state.loading = false
  }
}

watch(
  () => props.articleId,
  () => {
    page.value = 1
    Object.keys(childState).forEach((key) => delete childState[Number(key)])
    void load()
  },
  { immediate: true },
)
</script>

<style scoped lang="scss">
.comment-section__closed,
.comment-section__login-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 10px;
  font-size: 13px;
  color: var(--sd-text-muted);
  background: rgba(30, 43, 69, 0.35);
  border: 1px dashed var(--sd-border-strong);
}

.comment-section__closed {
  justify-content: center;
}

.comment-section__reply-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--sd-cyan);
}

.comment-section__composer-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.comment-section__list {
  min-height: 80px;
}

.comment-section__children {
  margin-top: 10px;
  padding: 4px 12px 4px 14px;
  border-left: 2px solid rgba(34, 211, 238, 0.35);
  border-radius: 0 10px 10px 0;
  background: rgba(9, 14, 26, 0.55);
}

.comment-section__more {
  display: flex;
  justify-content: flex-start;
}

.comment-section__expand {
  margin: -6px 0 4px 48px;
}
</style>
