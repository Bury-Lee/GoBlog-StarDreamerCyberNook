<template>
  <div class="image-uploader">
    <div class="image-uploader__preview" :style="{ width: `${width}px`, height: `${height}px` }">
      <img v-if="modelValue" :src="preview" alt="preview" />
      <div v-else class="image-uploader__placeholder">
        <el-icon><Picture /></el-icon>
      </div>
      <div v-if="modelValue" class="image-uploader__mask">
        <el-icon class="image-uploader__mask-btn" @click="clear"><Delete /></el-icon>
      </div>
    </div>
    <div class="image-uploader__side">
      <el-upload
        :show-file-list="false"
        :http-request="doUpload"
        :before-upload="beforeUpload"
        :accept="accept"
        :disabled="uploading"
      >
        <el-button size="small" :loading="uploading">
          <el-icon class="mr-4"><Upload /></el-icon>
          {{ modelValue ? '重新上传' : '上传图片' }}
        </el-button>
      </el-upload>
      <el-input
        v-model="manualUrl"
        size="small"
        placeholder="或直接粘贴图片地址"
        class="image-uploader__input"
        @keyup.enter="applyManual"
        @blur="applyManual"
      />
      <p class="image-uploader__tip sd-dim">{{ tip }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Picture, Upload } from '@element-plus/icons-vue'
import type { UploadRequestOptions } from 'element-plus'
import { uploadImage } from '@/api/ops'
import { imageUrl, resolveAssetUrl } from '@/api/request'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    width?: number
    height?: number
    accept?: string
    tip?: string
  }>(),
  {
    modelValue: '',
    width: 120,
    height: 90,
    accept: 'image/jpeg,image/png,image/gif,image/webp,image/bmp,image/tiff',
    tip: '支持 jpg/png/gif/webp/bmp/tiff,不支持 svg',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const uploading = ref(false)
const manualUrl = ref('')

watch(
  () => props.modelValue,
  (value) => {
    if (value && value.startsWith('/api/image')) {
      manualUrl.value = ''
    } else {
      manualUrl.value = value || ''
    }
  },
  { immediate: true },
)

const preview = computed(() => resolveAssetUrl(props.modelValue))

function beforeUpload(file: File): boolean {
  const maxSize = 20 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('图片不能超过 20MB')
    return false
  }
  return true
}

async function doUpload(options: UploadRequestOptions): Promise<void> {
  uploading.value = true
  try {
    const id = await uploadImage(options.file as File)
    emit('update:modelValue', imageUrl(id))
    ElMessage.success('图片上传成功')
  } catch {
    // 错误提示已由请求层处理
  } finally {
    uploading.value = false
  }
}

function applyManual(): void {
  const value = manualUrl.value.trim()
  if (value === props.modelValue) return
  emit('update:modelValue', value)
}

function clear(): void {
  manualUrl.value = ''
  emit('update:modelValue', '')
}
</script>

<style scoped lang="scss">
.image-uploader {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.image-uploader__preview {
  position: relative;
  overflow: hidden;
  border: 1px dashed var(--sd-border-strong);
  border-radius: 10px;
  background: rgba(9, 14, 26, 0.7);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.image-uploader__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--sd-text-dim);
  font-size: 24px;
}

.image-uploader__mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(7, 11, 20, 0.65);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.image-uploader__preview:hover .image-uploader__mask {
  opacity: 1;
}

.image-uploader__mask-btn {
  font-size: 20px;
  color: var(--sd-red);
  cursor: pointer;
}

.image-uploader__side {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 220px;
  flex: 1;
}

.image-uploader__input {
  max-width: 260px;
}

.image-uploader__tip {
  font-size: 12px;
  margin: 0;
}

.mr-4 {
  margin-right: 4px;
}
</style>
