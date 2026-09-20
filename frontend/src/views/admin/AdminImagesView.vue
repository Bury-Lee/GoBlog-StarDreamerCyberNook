<template>
  <div class="admin-images">
    <section class="sd-panel admin-images__upload">
      <header class="sd-panel__header">
        <span class="sd-panel__title">上传图片</span>
        <span class="sd-dim">同一文件 MD5 相同会复用已有记录</span>
      </header>
      <div class="sd-panel__body">
        <el-upload
          drag
          multiple
          :show-file-list="false"
          :http-request="doUpload"
          accept="image/jpeg,image/png,image/gif,image/webp,image/bmp,image/tiff"
        >
          <el-icon class="admin-images__icon"><UploadFilled /></el-icon>
          <div class="admin-images__hint">将图片拖到此处,或<em>点击上传</em></div>
          <template #tip>
            <div class="sd-dim admin-images__tip">
              支持 jpg/png/gif/webp/bmp/tiff,不支持 svg;单文件建议不超过 20MB
            </div>
          </template>
        </el-upload>
      </div>
    </section>

    <section class="sd-panel admin-images__list">
      <header class="admin-images__head">
        <div class="admin-images__filters">
          <el-input
            v-model="key"
            class="admin-images__search"
            placeholder="搜索文件名"
            clearable
            @keyup.enter="search"
            @clear="search"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <span class="sd-dim">共 {{ count }} 张</span>
        </div>
        <div class="admin-images__actions">
          <el-button :disabled="!selection.length" @click="copySelected">复制链接</el-button>
          <el-button type="danger" plain :disabled="!selection.length" @click="removeSelected">
            批量删除({{ selection.length }})
          </el-button>
        </div>
      </header>

      <div v-loading="loading" class="admin-images__grid">
        <EmptyState v-if="!loading && !list.length" text="图库还是空的,上传第一张图片吧" />
        <div
          v-for="item in list"
          :key="item.id"
          class="admin-image-card"
          :class="{ 'is-selected': selectedIds.includes(item.id) }"
          @click="toggleSelect(item.id)"
        >
          <el-image :src="item.webPath || imageUrl(item.id)" fit="cover" class="admin-image-card__img" />
          <div class="admin-image-card__info">
            <span class="sd-ellipsis admin-image-card__name" :title="item.filename">{{ item.filename }}</span>
            <span class="sd-dim">{{ formatFileSize(item.size) }}</span>
          </div>
          <el-icon v-if="selectedIds.includes(item.id)" class="admin-image-card__check"><CircleCheckFilled /></el-icon>
        </div>
      </div>

      <PaginationBar
        :page="page"
        :limit="limit"
        :count="count"
        :page-sizes="[12, 24, 36, 40]"
        @update:page="changePage"
        @update:limit="changeLimit"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { CircleCheckFilled, Search, UploadFilled } from '@element-plus/icons-vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { fetchImages, removeImages, uploadImage } from '@/api/ops'
import { imageUrl } from '@/api/request'
import { copyToClipboard } from '@/utils/html'
import { formatFileSize } from '@/utils/format'
import { usePagination } from '@/composables/usePagination'

const selection = ref<number[]>([])
const selectedIds = computed(() => selection.value)

const { list, count, page, limit, key, loading, load, search, changePage, changeLimit } = usePagination(
  (params) => fetchImages(params),
  () => ({}),
  { limit: 12 },
)

function toggleSelect(id: number): void {
  if (selection.value.includes(id)) {
    selection.value = selection.value.filter((item) => item !== id)
  } else {
    selection.value = [...selection.value, id]
  }
}

async function doUpload(options: UploadRequestOptions): Promise<void> {
  try {
    await uploadImage(options.file as File)
    ElMessage.success('上传成功')
    await load()
  } catch {
    // ignore
  }
}

async function copySelected(): Promise<void> {
  const links = selection.value.map((id) => `${window.location.origin}${imageUrl(id)}`).join('\n')
  try {
    await copyToClipboard(links)
    ElMessage.success('已复制选中图片链接')
  } catch {
    ElMessage.warning('复制失败')
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 张图片?`, '删除图片', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await removeImages(selection.value)
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}
</script>

<style scoped lang="scss">
.admin-images {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.admin-images__icon {
  font-size: 44px;
  color: var(--sd-text-dim);
}

.admin-images__hint {
  font-size: 13px;
  color: var(--sd-text-muted);

  em {
    color: var(--sd-cyan);
    font-style: normal;
  }
}

.admin-images__tip {
  margin-top: 8px;
  font-size: 12px;
}

.admin-images__list {
  padding: 18px;
}

.admin-images__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.admin-images__filters,
.admin-images__actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.admin-images__search {
  width: 220px;
}

.admin-images__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 14px;
  min-height: 160px;
}

.admin-image-card {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.6);
  cursor: pointer;
  transition: all 0.2s ease;
}

.admin-image-card:hover {
  border-color: rgba(34, 211, 238, 0.5);
}

.admin-image-card.is-selected {
  border-color: var(--sd-cyan);
  box-shadow: 0 0 0 2px rgba(34, 211, 238, 0.35);
}

.admin-image-card__img {
  display: block;
  width: 100%;
  height: 110px;
}

.admin-image-card__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 10px;
  font-size: 12px;
}

.admin-image-card__name {
  color: var(--sd-text-muted);
}

.admin-image-card__check {
  position: absolute;
  top: 8px;
  right: 8px;
  font-size: 20px;
  color: var(--sd-cyan);
  filter: drop-shadow(0 0 6px rgba(34, 211, 238, 0.9));
}
</style>
