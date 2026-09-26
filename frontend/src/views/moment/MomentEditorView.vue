<template>
  <div class="sd-page">
    <div class="sd-container">
      <section class="moment-editor sd-panel">
        <header class="moment-editor__head">
          <div class="moment-editor__who">
            <UserAvatar :src="userStore.avatar" :name="userStore.nickname" :size="42" />
            <div class="moment-editor__titles">
              <h1 class="moment-editor__title">{{ isEdit ? '编辑动态' : '发布动态' }}</h1>
              <span class="sd-dim">{{ userStore.nickname }} · {{ momentTypeLabel(form.type) }}</span>
            </div>
          </div>
          <div class="moment-editor__head-actions">
            <el-button @click="router.back()">返回</el-button>
            <el-button
              v-if="!isEdit"
              :loading="saving"
              :disabled="!canSubmit"
              @click="submit(0)"
            >
              存草稿
            </el-button>
            <el-button
              type="primary"
              :loading="saving"
              :disabled="!canSubmit"
              @click="submit(isEdit ? form.status : 2)"
            >
              {{ isEdit ? '保存修改' : '发布' }}
            </el-button>
          </div>
        </header>

        <el-input
          v-model="form.content"
          class="moment-editor__content"
          type="textarea"
          :rows="8"
          resize="none"
          maxlength="2000"
          show-word-limit
          placeholder="分享此刻的想法…"
        />

        <div class="moment-editor__images">
          <div v-for="(img, index) in form.images" :key="img" class="moment-editor__image">
            <img :src="resolveAssetUrl(img)" alt="preview" />
            <el-icon class="moment-editor__image-remove" @click="form.images.splice(index, 1)">
              <CircleCloseFilled />
            </el-icon>
          </div>
          <el-upload
            v-if="form.images.length < 9"
            :show-file-list="false"
            :http-request="doUpload"
            :before-upload="beforeUpload"
            accept="image/jpeg,image/png,image/gif,image/webp,image/bmp,image/tiff"
            :disabled="uploading"
          >
            <div class="moment-editor__add" :class="{ 'is-loading': uploading }">
              <el-icon><Plus /></el-icon>
              <span class="sd-dim">{{ form.images.length }}/9</span>
            </div>
          </el-upload>
        </div>

        <div class="moment-editor__options">
          <div class="moment-editor__option">
            <span class="sd-dim">类型</span>
            <el-select v-model="form.type" size="small" class="moment-editor__select">
              <el-option
                v-for="item in MOMENT_TYPE_OPTIONS"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </div>
          <div class="moment-editor__option">
            <span class="sd-dim">可见性</span>
            <el-select v-model="form.visibility" size="small" class="moment-editor__select">
              <el-option
                v-for="item in MOMENT_VISIBILITY_OPTIONS"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { CircleCloseFilled, Plus } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { createMoment, fetchMomentDetail, updateMoment } from '@/api/moment'
import { imageUrl, resolveAssetUrl } from '@/api/request'
import { uploadImage } from '@/api/ops'
import { MOMENT_TYPE_OPTIONS, MOMENT_VISIBILITY_OPTIONS, momentTypeLabel } from '@/utils/format'
import { useUserStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isEdit = computed(() => route.name === 'moment-edit')
const momentId = computed(() => Number(route.params.id || 0))

const form = reactive({
  content: '',
  images: [] as string[],
  type: 0,
  visibility: 0,
  status: 2,
})
const saving = ref(false)
const uploading = ref(false)

const canSubmit = computed(() => Boolean(form.content.trim()) || form.images.length > 0)

onMounted(async () => {
  if (!isEdit.value || !momentId.value) return
  try {
    const moment = await fetchMomentDetail(momentId.value)
    form.content = moment.content || ''
    form.images = [...(moment.images || [])]
    form.type = moment.type ?? 0
    form.visibility = moment.visibility ?? 0
    form.status = moment.status ?? 2
  } catch {
    ElMessage.error('动态不存在或无权查看')
    router.replace({ name: 'home' })
  }
})

function beforeUpload(file: File): boolean {
  if (file.size > 20 * 1024 * 1024) {
    ElMessage.error('图片不能超过 20MB')
    return false
  }
  return true
}

async function doUpload(options: UploadRequestOptions): Promise<void> {
  if (form.images.length >= 9) {
    ElMessage.warning('最多上传 9 张图片')
    return
  }
  uploading.value = true
  try {
    const id = await uploadImage(options.file as File)
    form.images.push(imageUrl(id))
  } catch {
    // 错误提示由请求层处理
  } finally {
    uploading.value = false
  }
}

function backToMoments(): void {
  router.push({ name: 'user-home', params: { id: userStore.userId }, query: { tab: 'moments' } })
}

async function submit(status: number): Promise<void> {
  if (!canSubmit.value) return
  saving.value = true
  try {
    if (isEdit.value && momentId.value) {
      await updateMoment({
        id: momentId.value,
        content: form.content,
        images: form.images,
        type: form.type,
        visibility: form.visibility,
        status,
      })
      ElMessage.success('动态已更新')
    } else {
      await createMoment({
        content: form.content,
        images: form.images,
        type: form.type,
        visibility: form.visibility,
        status,
      })
      ElMessage.success(status === 0 ? '已存草稿' : '发布成功')
    }
    backToMoments()
  } catch {
    // 错误提示由请求层处理
  } finally {
    saving.value = false
  }
}
</script>

<style scoped lang="scss">
.moment-editor {
  padding: 20px 22px;
}

.moment-editor__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.moment-editor__who {
  display: flex;
  align-items: center;
  gap: 12px;
}

.moment-editor__titles {
  display: flex;
  flex-direction: column;
}

.moment-editor__title {
  font-size: 20px;
}

.moment-editor__head-actions {
  display: flex;
  gap: 10px;
}

.moment-editor__content :deep(textarea) {
  font-size: 15px;
  line-height: 1.8;
}

.moment-editor__images {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
}

.moment-editor__image {
  position: relative;
  width: 100px;
  height: 100px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--sd-border);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.moment-editor__image-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  font-size: 18px;
  color: var(--sd-red);
  cursor: pointer;
  background: rgba(7, 11, 20, 0.72);
  border-radius: 50%;
}

.moment-editor__add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 100px;
  height: 100px;
  border: 1px dashed var(--sd-border-strong);
  border-radius: 10px;
  color: var(--sd-text-muted);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: var(--sd-cyan);
    border-color: rgba(34, 211, 238, 0.6);
  }
}

.moment-editor__options {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  margin-top: 18px;
  padding-top: 14px;
  border-top: 1px dashed var(--sd-border);
}

.moment-editor__option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.moment-editor__select {
  width: 120px;
}
</style>
