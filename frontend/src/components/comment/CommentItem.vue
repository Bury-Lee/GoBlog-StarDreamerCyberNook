<template>
  <div class="comment-item" :class="{ 'comment-item--child': isChild }">
    <UserAvatar :src="comment.user?.avatar" :name="comment.user?.nickname" :size="isChild ? 28 : 36" />
    <div class="comment-item__main">
      <div class="comment-item__head">
        <span class="comment-item__name">{{ comment.user?.nickname || '匿名用户' }}</span>
        <span class="comment-item__time sd-dim">{{ fromNow(comment.createdAt) }}</span>
      </div>
      <p class="comment-item__content">{{ comment.content }}</p>
      <div class="comment-item__actions">
        <span
          class="comment-item__action"
          :class="{ 'is-active': comment.diggCount > 0 }"
          @click="emit('digg', comment)"
        >
          <el-icon><Pointer /></el-icon>
          {{ comment.diggCount > 0 ? comment.diggCount : '点赞' }}
        </span>
        <span v-if="!isChild" class="comment-item__action" @click="emit('reply', comment)">
          <el-icon><ChatLineRound /></el-icon>
          回复
        </span>
        <span v-if="canManage" class="comment-item__action comment-item__action--danger" @click="emit('remove', comment)">
          <el-icon><Delete /></el-icon>
          删除
        </span>
        <slot name="extra" />
      </div>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ChatLineRound, Delete, Pointer } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import type { CommentModel } from '@/api/types'
import { fromNow } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    comment: CommentModel
    isChild?: boolean
    currentUserId?: number
    isAdmin?: boolean
  }>(),
  {
    isChild: false,
    currentUserId: 0,
    isAdmin: false,
  },
)

const emit = defineEmits<{
  (e: 'digg', comment: CommentModel): void
  (e: 'reply', comment: CommentModel): void
  (e: 'remove', comment: CommentModel): void
}>()

const canManage = computed(() => {
  if (props.isAdmin) return true
  return props.currentUserId > 0 && props.comment.userID === props.currentUserId
})
</script>

<style scoped lang="scss">
.comment-item {
  display: flex;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px dashed rgba(30, 43, 69, 0.7);
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-item--child {
  padding: 10px 0;
  border-bottom: none;
}

.comment-item__main {
  flex: 1;
  min-width: 0;
}

.comment-item__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.comment-item__name {
  font-size: 13px;
  font-weight: 600;
  color: var(--sd-text);
}

.comment-item__time {
  font-size: 12px;
}

.comment-item__content {
  font-size: 14px;
  line-height: 1.8;
  color: #cbd5e1;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-item__actions {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 8px;
  font-size: 12px;
  color: var(--sd-text-dim);
}

.comment-item__action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  transition: color 0.2s ease;
}

.comment-item__action:hover {
  color: var(--sd-cyan);
}

.comment-item__action.is-active {
  color: var(--sd-cyan);
}

.comment-item__action--danger:hover {
  color: var(--sd-red);
}
</style>
