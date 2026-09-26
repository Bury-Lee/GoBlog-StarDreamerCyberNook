<template>
  <article class="moment-card sd-panel">
    <header class="moment-card__head">
      <UserAvatar :src="moment.user?.avatar" :name="moment.user?.nickname" :size="40" :user-id="moment.userID" />
      <div class="moment-card__meta">
        <span class="moment-card__name">{{ moment.user?.nickname || '匿名用户' }}</span>
        <span class="sd-dim moment-card__time">{{ fromNow(moment.createdAt) }}</span>
      </div>
      <div class="moment-card__tags">
        <el-tag size="small" effect="plain">{{ momentTypeLabel(moment.type) }}</el-tag>
        <el-tag
          v-if="isOwner"
          size="small"
          effect="plain"
          :type="momentVisibilityType(moment.visibility)"
        >
          {{ momentVisibilityLabel(moment.visibility) }}
        </el-tag>
      </div>
    </header>

    <p v-if="moment.content" class="moment-card__content">{{ moment.content }}</p>

    <div v-if="images.length" class="moment-card__images" :class="`is-${Math.min(images.length, 3)}`">
      <el-image
        v-for="(img, index) in images"
        :key="index"
        class="moment-card__image"
        :src="img"
        fit="cover"
        :preview-src-list="images"
        :initial-index="index"
        preview-teleported
      />
    </div>

    <div v-if="moment.repostFrom" class="moment-card__repost">
      <div class="moment-card__repost-head">
        <UserAvatar
          :src="moment.repostFrom.user?.avatar"
          :name="moment.repostFrom.user?.nickname"
          :size="22"
          :user-id="moment.repostFrom.userID"
        />
        <span class="moment-card__repost-name">{{ moment.repostFrom.user?.nickname || '匿名用户' }}</span>
      </div>
      <p v-if="moment.repostFrom.content" class="moment-card__repost-content">{{ moment.repostFrom.content }}</p>
      <div v-if="repostImages.length" class="moment-card__repost-images">
        <el-image
          v-for="(img, index) in repostImages"
          :key="index"
          class="moment-card__repost-image"
          :src="img"
          fit="cover"
          :preview-src-list="repostImages"
          :initial-index="index"
          preview-teleported
        />
      </div>
    </div>

    <footer class="moment-card__actions">
      <el-button text size="small" :class="{ 'is-active': digged }" :loading="digging" @click="onDigg">
        <el-icon><Pointer /></el-icon>{{ formatNumber(likeCount) }}
      </el-button>
      <el-button text size="small" @click="showComments = !showComments">
        <el-icon><ChatDotRound /></el-icon>{{ formatNumber(commentCount) }}
      </el-button>
      <el-button text size="small" @click="onRepost">
        <el-icon><Share /></el-icon>{{ formatNumber(repostCount) }}
      </el-button>
      <span class="moment-card__spacer" />
      <el-button v-if="isOwner || userStore.isAdmin" text size="small" @click="emit('edit', moment)">
        <el-icon><EditPen /></el-icon>编辑
      </el-button>
      <el-button v-if="isOwner || userStore.isAdmin" text size="small" class="is-danger" @click="onRemove">
        <el-icon><Delete /></el-icon>删除
      </el-button>
    </footer>

    <MomentCommentList v-if="showComments" :moment-id="moment.id" @change="onCommentChange" />
  </article>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Delete, EditPen, Pointer, Share } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import MomentCommentList from './MomentCommentList.vue'
import { diggMoment, fetchMomentInteraction, removeMoment, repostMoment } from '@/api/moment'
import { resolveAssetUrl } from '@/api/request'
import type { MomentModel } from '@/api/types'
import {
  fromNow,
  formatNumber,
  momentTypeLabel,
  momentVisibilityLabel,
  momentVisibilityType,
} from '@/utils/format'
import { useUserStore } from '@/stores'

const props = defineProps<{ moment: MomentModel }>()
const emit = defineEmits<{
  (e: 'changed'): void
  (e: 'edit', moment: MomentModel): void
}>()

const userStore = useUserStore()

const digged = ref(false)
const digging = ref(false)
const likeCount = ref(props.moment.likeCount || 0)
const commentCount = ref(props.moment.commentCount || 0)
const repostCount = ref(props.moment.repostCount || 0)
const showComments = ref(false)

const isOwner = computed(() => userStore.userId === props.moment.userID)
const images = computed(() => (props.moment.images || []).map((src) => resolveAssetUrl(src)))
const repostImages = computed(() =>
  (props.moment.repostFrom?.images || []).map((src) => resolveAssetUrl(src)),
)

watch(
  () => props.moment,
  (value) => {
    likeCount.value = value.likeCount || 0
    commentCount.value = value.commentCount || 0
    repostCount.value = value.repostCount || 0
  },
)

onMounted(() => {
  if (!userStore.isLogin) return
  void fetchMomentInteraction(props.moment.id)
    .then((data) => {
      digged.value = Boolean(data?.digged)
    })
    .catch(() => {
      digged.value = false
    })
})

async function onDigg(): Promise<void> {
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录后再点赞')
    return
  }
  digging.value = true
  try {
    const result = await diggMoment(props.moment.id)
    digged.value = result.digged
    likeCount.value = result.likeCount
  } catch {
    // ignore
  } finally {
    digging.value = false
  }
}

function onCommentChange(delta: number): void {
  commentCount.value = Math.max(0, commentCount.value + delta)
}

async function onRepost(): Promise<void> {
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录后再转发')
    return
  }
  let content = ''
  try {
    const result = await ElMessageBox.prompt('转发理由(可选)', '转发动态', {
      inputType: 'textarea',
      inputValue: '',
      confirmButtonText: '转发',
      cancelButtonText: '取消',
    })
    content = result.value || ''
  } catch {
    return
  }
  try {
    await repostMoment(props.moment.id, { content })
    repostCount.value += 1
    ElMessage.success('转发成功')
    emit('changed')
  } catch {
    // ignore
  }
}

async function onRemove(): Promise<void> {
  try {
    await ElMessageBox.confirm('确认删除该动态?删除后不可恢复', '删除动态', { type: 'warning' })
  } catch {
    return
  }
  try {
    await removeMoment(props.moment.id)
    ElMessage.success('删除成功')
    emit('changed')
  } catch {
    // ignore
  }
}
</script>

<style scoped lang="scss">
.moment-card {
  padding: 16px 18px;
}

.moment-card__head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.moment-card__meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.moment-card__name {
  font-size: 14px;
  font-weight: 600;
}

.moment-card__time {
  font-size: 12px;
}

.moment-card__tags {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

.moment-card__content {
  margin: 12px 0 0;
  font-size: 14px;
  line-height: 1.8;
  color: var(--sd-text);
  white-space: pre-wrap;
  word-break: break-word;
}

.moment-card__images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-top: 12px;
  max-width: 480px;
}

.moment-card__images.is-1 {
  grid-template-columns: minmax(0, 220px);
}

.moment-card__images.is-2 {
  grid-template-columns: repeat(2, minmax(0, 220px));
}

.moment-card__image {
  aspect-ratio: 1;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--sd-border);
  background: rgba(9, 14, 26, 0.7);
  cursor: zoom-in;
}

.moment-card__repost {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid var(--sd-border);
  border-radius: 10px;
  background: rgba(9, 14, 26, 0.55);
}

.moment-card__repost-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.moment-card__repost-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--sd-cyan);
}

.moment-card__repost-content {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--sd-text-muted);
  white-space: pre-wrap;
  word-break: break-word;
}

.moment-card__repost-images {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.moment-card__repost-image {
  width: 72px;
  height: 72px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--sd-border);
  cursor: zoom-in;
}

.moment-card__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px dashed var(--sd-border);
}

.moment-card__actions :deep(.el-button.is-active) {
  color: var(--sd-cyan);
}

.moment-card__spacer {
  flex: 1;
}

.is-danger {
  color: var(--sd-red);
}
</style>
