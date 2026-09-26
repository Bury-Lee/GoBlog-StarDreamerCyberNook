<template>
  <div class="moment-comments">
    <div v-if="!userStore.isLogin" class="moment-comments__tip">
      登录后即可参与评论
    </div>

    <div v-else class="moment-comments__composer">
      <div v-if="replyTo" class="moment-comments__reply-tip">
        <span>回复 <b>{{ replyTo.user?.nickname || '匿名用户' }}</b></span>
        <el-button text size="small" @click="replyTo = null">取消</el-button>
      </div>
      <el-input
        v-model="content"
        type="textarea"
        :rows="2"
        resize="none"
        maxlength="500"
        show-word-limit
        :placeholder="replyTo ? '写下你的回复…' : '说点什么…'"
      />
      <div class="moment-comments__composer-actions">
        <el-button type="primary" size="small" :loading="submitting" :disabled="!content.trim()" @click="submit">
          {{ replyTo ? '回复' : '评论' }}
        </el-button>
      </div>
    </div>

    <div v-loading="loading" class="moment-comments__list">
      <EmptyState v-if="!loading && !list.length" text="还没有评论" compact />
      <template v-for="comment in list" :key="comment.id">
        <div class="moment-comment">
          <UserAvatar :src="comment.user?.avatar" :name="comment.user?.nickname" :size="28" :user-id="comment.userID" />
          <div class="moment-comment__main">
            <div class="moment-comment__head">
              <span class="moment-comment__name">{{ comment.user?.nickname || '匿名用户' }}</span>
              <span class="sd-dim moment-comment__time">{{ fromNow(comment.createdAt) }}</span>
            </div>
            <p class="moment-comment__content">{{ comment.content }}</p>
            <div class="moment-comment__actions">
              <el-button text size="small" @click="onReply(comment)">回复</el-button>
              <el-button text size="small" @click="onDigg(comment)">
                <el-icon><Pointer /></el-icon>{{ comment.diggCount || 0 }}
              </el-button>
              <el-button
                v-if="canRemove(comment)"
                text
                size="small"
                class="is-danger"
                @click="onRemove(comment)"
              >
                删除
              </el-button>
            </div>

            <div v-if="childState[comment.id]" class="moment-comment__children">
              <span v-if="childState[comment.id].count === 0" class="sd-dim moment-comment__empty">暂无回复</span>
              <div v-for="child in childState[comment.id].list" :key="child.id" class="moment-comment moment-comment--child">
                <UserAvatar :src="child.user?.avatar" :name="child.user?.nickname" :size="24" :user-id="child.userID" />
                <div class="moment-comment__main">
                  <div class="moment-comment__head">
                    <span class="moment-comment__name">{{ child.user?.nickname || '匿名用户' }}</span>
                    <span class="sd-dim moment-comment__time">{{ fromNow(child.createdAt) }}</span>
                  </div>
                  <p class="moment-comment__content">{{ child.content }}</p>
                  <div class="moment-comment__actions">
                    <el-button text size="small" @click="onReply(child)">回复</el-button>
                    <el-button text size="small" @click="onDigg(child)">
                      <el-icon><Pointer /></el-icon>{{ child.diggCount || 0 }}
                    </el-button>
                    <el-button v-if="canRemove(child)" text size="small" class="is-danger" @click="onRemove(child)">
                      删除
                    </el-button>
                  </div>
                </div>
              </div>
              <el-button
                v-if="childState[comment.id].count > childState[comment.id].list.length"
                text
                size="small"
                :loading="childState[comment.id].loading"
                @click="loadMoreChildren(comment)"
              >
                查看更多回复
              </el-button>
            </div>
            <el-button
              v-else
              text
              size="small"
              @click="loadChildren(comment)"
            >
              查看回复
            </el-button>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Pointer } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import {
  createMomentComment,
  diggMomentComment,
  fetchMomentChildComments,
  fetchMomentComments,
  removeMomentComment,
} from '@/api/moment'
import type { MomentCommentModel } from '@/api/types'
import { fromNow } from '@/utils/format'
import { useUserStore } from '@/stores'

const props = defineProps<{ momentId: number }>()
const emit = defineEmits<{ (e: 'change', delta: number): void }>()

const userStore = useUserStore()

const list = ref<MomentCommentModel[]>([])
const loading = ref(false)
const content = ref('')
const submitting = ref(false)
const replyTo = ref<MomentCommentModel | null>(null)

interface ChildState {
  list: MomentCommentModel[]
  count: number
  page: number
  loading: boolean
}

const childState = reactive<Record<number, ChildState>>({})

function canRemove(comment: MomentCommentModel): boolean {
  return userStore.isAdmin || userStore.userId === comment.userID
}

async function load(): Promise<void> {
  if (!props.momentId) return
  loading.value = true
  try {
    const data = await fetchMomentComments({ momentID: props.momentId, page: 1, limit: 10 })
    list.value = data?.list ?? []
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function submit(): Promise<void> {
  const text = content.value.trim()
  if (!text) return
  submitting.value = true
  const parent = replyTo.value
  try {
    await createMomentComment({ momentID: props.momentId, content: text, parentID: parent?.id ?? 0 })
    ElMessage.success(parent ? '回复成功' : '评论成功')
    emit('change', 1)
    content.value = ''
    replyTo.value = null
    if (parent) {
      const rootID = parent.rootParentID ?? parent.id
      delete childState[rootID]
      await loadChildren({ ...parent, id: rootID } as MomentCommentModel)
    } else {
      await load()
    }
  } catch {
    // 错误提示由请求层处理
  } finally {
    submitting.value = false
  }
}

function onReply(comment: MomentCommentModel): void {
  replyTo.value = comment
}

async function onDigg(comment: MomentCommentModel): Promise<void> {
  if (!userStore.isLogin) return
  try {
    const result = await diggMomentComment(comment.id)
    comment.diggCount = result?.diggCount ?? comment.diggCount
    comment.digged = result?.digged ?? !comment.digged
  } catch {
    // ignore
  }
}

async function onRemove(comment: MomentCommentModel): Promise<void> {
  try {
    await ElMessageBox.confirm(comment.rootParentID ? '确认删除这条回复?' : '确认删除该评论及其回复?', '删除评论', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeMomentComment(comment.id)
    ElMessage.success('删除成功')
    const rootID = comment.rootParentID ?? comment.id
    const removed = comment.rootParentID == null ? (childState[rootID]?.count ?? 0) + 1 : 1
    emit('change', -removed)
    delete childState[rootID]
    await load()
  } catch {
    // ignore
  }
}

async function loadChildren(comment: MomentCommentModel): Promise<void> {
  const rootID = comment.rootParentID ?? comment.id
  childState[rootID] = { list: [], count: 0, page: 1, loading: true }
  try {
    const data = await fetchMomentChildComments({ root: rootID, page: 1, limit: 5 })
    childState[rootID] = { list: data?.list ?? [], count: data?.count ?? 0, page: 1, loading: false }
  } catch {
    delete childState[rootID]
  }
}

async function loadMoreChildren(comment: MomentCommentModel): Promise<void> {
  const rootID = comment.id
  const state = childState[rootID]
  if (!state) return
  state.loading = true
  try {
    const nextPage = state.page + 1
    const data = await fetchMomentChildComments({ root: rootID, page: nextPage, limit: 5 })
    state.list = [...state.list, ...(data?.list ?? [])]
    state.page = nextPage
    state.count = data?.count ?? state.count
  } finally {
    state.loading = false
  }
}

defineExpose({ load })

load()
</script>

<style scoped lang="scss">
.moment-comments {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--sd-border);
}

.moment-comments__tip {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--sd-text-muted);
  background: rgba(30, 43, 69, 0.35);
  border: 1px dashed var(--sd-border-strong);
  border-radius: 8px;
}

.moment-comments__composer-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 6px;
}

.moment-comments__reply-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--sd-cyan);
}

.moment-comments__list {
  margin-top: 10px;
  min-height: 40px;
}

.moment-comment {
  display: flex;
  gap: 10px;
  padding: 8px 0;
}

.moment-comment--child {
  padding: 6px 0;
}

.moment-comment__main {
  flex: 1;
  min-width: 0;
}

.moment-comment__head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.moment-comment__name {
  font-size: 13px;
  font-weight: 600;
}

.moment-comment__time {
  font-size: 12px;
}

.moment-comment__content {
  margin: 4px 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--sd-text-muted);
  white-space: pre-wrap;
  word-break: break-word;
}

.moment-comment__actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.moment-comment__children {
  margin-top: 6px;
  padding-left: 12px;
  border-left: 2px solid rgba(34, 211, 238, 0.35);
}

.moment-comment__empty {
  font-size: 12px;
}

.is-danger {
  color: var(--sd-red);
}
</style>
