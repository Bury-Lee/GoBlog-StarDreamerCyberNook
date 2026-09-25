<template>
  <div class="sd-page">
    <div class="sd-container">
      <div class="sd-layout-3col">
        <aside class="sd-panel collect-side">
          <header class="sd-panel__header">
            <span class="sd-panel__title">收藏夹</span>
            <el-button v-if="isSelf" text size="small" @click="openCreate">
              <el-icon><Plus /></el-icon>
              新建收藏夹
            </el-button>
          </header>
          <div v-loading="folderLoading" class="collect-side__body">
            <EmptyState v-if="!folders.length && !folderLoading" :text="folderError || '暂无收藏夹'" compact />
            <div
              v-for="folder in folders"
              :key="folder.id"
              class="collect-side__item"
              :class="{ 'is-active': folder.id === activeFolderID }"
              @click="selectFolder(folder)"
            >
              <el-icon><Folder /></el-icon>
              <span class="sd-ellipsis collect-side__name">{{ folder.title }}</span>
              <el-tag v-if="folder.isDefault" size="small" type="info">默认</el-tag>
              <el-tag v-else-if="!folder.isPublic" size="small" type="warning">私密</el-tag>
              <template v-if="isSelf">
                <el-icon class="collect-side__op" @click.stop="openEdit(folder)"><EditPen /></el-icon>
                <el-icon
                  v-if="!folder.isDefault"
                  class="collect-side__op collect-side__op--danger"
                  @click.stop="removeFolder(folder)"
                >
                  <Delete />
                </el-icon>
              </template>
            </div>
          </div>
        </aside>

        <main class="sd-panel collect-main">
          <header class="sd-panel__header">
            <span class="sd-panel__title">
              {{ activeFolder ? activeFolder.title : '收藏内容' }}
              <span class="sd-dim collect-main__count">({{ count }})</span>
            </span>
            <el-input
              v-model="key"
              class="collect-main__search"
              placeholder="搜索收藏的文章"
              clearable
              @keyup.enter="search"
              @clear="search"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </header>
          <div class="sd-panel__body">
            <div v-loading="articleLoading" class="collect-main__list">
              <EmptyState v-if="!articleLoading && !articles.length" :text="articleError || '这个收藏夹还是空的'" />
              <ArticleCard
                v-for="item in articles"
                :key="item.id"
                :article="item as never"
                layout="list"
              />
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

    <el-dialog v-model="dialogVisible" :title="editingFolder ? '编辑收藏夹' : '新建收藏夹'" width="480px">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="folderForm.title" maxlength="32" show-word-limit placeholder="收藏夹名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="folderForm.abstract" type="textarea" :rows="3" resize="none" placeholder="可选" />
        </el-form-item>
        <el-form-item label="封面">
          <ImageUploader v-model="folderForm.cover" :width="120" :height="90" />
        </el-form-item>
        <el-form-item v-if="editingFolder" label="公开">
          <el-switch v-model="folderForm.isPublic" />
          <span class="sd-dim" style="margin-left: 8px">关闭后仅自己可见</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitFolder">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, EditPen, Folder, Plus, Search } from '@element-plus/icons-vue'
import ArticleCard from '@/components/article/ArticleCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import {
  createCollectFolder,
  fetchCollectArticles,
  fetchCollectFolders,
  removeCollectFolders,
  updateCollectFolder,
} from '@/api/article'
import type { ArticleModel, CollectModel } from '@/api/types'
import { useUserStore } from '@/stores'

const route = useRoute()
const userStore = useUserStore()

const folders = ref<CollectModel[]>([])
const folderLoading = ref(false)
const folderError = ref('')
const activeFolderID = ref(0)
const articles = ref<ArticleModel[]>([])
const articleLoading = ref(false)
const articleError = ref('')
const count = ref(0)
const page = ref(1)
const limit = ref(8)
const key = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editingFolder = ref<CollectModel | null>(null)

const folderForm = reactive({
  title: '',
  abstract: '',
  cover: '',
  isPublic: true,
})

const ownerID = computed(() => Number(route.query.user || userStore.userId || 0))
const isSelf = computed(() => ownerID.value === userStore.userId)
const activeFolder = computed(() => folders.value.find((item) => item.id === activeFolderID.value) || null)

async function loadFolders(): Promise<void> {
  if (!ownerID.value) return
  folderLoading.value = true
  folderError.value = ''
  try {
    const data = await fetchCollectFolders({ id: ownerID.value, page: 1, limit: 40 })
    folders.value = data?.list ?? []
    const queryFolder = Number(route.query.folder || 0)
    const target = folders.value.find((item) => item.id === queryFolder) || folders.value[0]
    if (target) {
      activeFolderID.value = target.id
      await loadArticles()
    }
  } catch (error) {
    folders.value = []
    folderError.value = error instanceof Error ? error.message : '收藏夹未公开或不存在'
  } finally {
    folderLoading.value = false
  }
}

async function loadArticles(): Promise<void> {
  if (!activeFolderID.value) return
  articleLoading.value = true
  articleError.value = ''
  try {
    const data = await fetchCollectArticles({
      id: activeFolderID.value,
      page: page.value,
      limit: limit.value,
      key: key.value || undefined,
    })
    articles.value = data?.list ?? []
    count.value = data?.count ?? 0
  } catch (error) {
    articles.value = []
    count.value = 0
    articleError.value = error instanceof Error ? error.message : '收藏内容加载失败'
  } finally {
    articleLoading.value = false
  }
}

function selectFolder(folder: CollectModel): void {
  if (activeFolderID.value === folder.id) return
  activeFolderID.value = folder.id
  page.value = 1
  key.value = ''
  void loadArticles()
}

function search(): void {
  page.value = 1
  void loadArticles()
}

function changePage(next: number): void {
  page.value = next
  void loadArticles()
}

function openCreate(): void {
  editingFolder.value = null
  folderForm.title = ''
  folderForm.abstract = ''
  folderForm.cover = ''
  folderForm.isPublic = true
  dialogVisible.value = true
}

function openEdit(folder: CollectModel): void {
  editingFolder.value = folder
  folderForm.title = folder.title
  folderForm.abstract = folder.abstract
  folderForm.cover = folder.cover
  folderForm.isPublic = folder.isPublic
  dialogVisible.value = true
}

async function submitFolder(): Promise<void> {
  if (!folderForm.title.trim()) {
    ElMessage.warning('请输入收藏夹名称')
    return
  }
  submitting.value = true
  try {
    if (editingFolder.value) {
      await updateCollectFolder({
        id: editingFolder.value.id,
        title: folderForm.title.trim(),
        abstract: folderForm.abstract,
        cover: folderForm.cover,
        isPublic: folderForm.isPublic,
      })
      ElMessage.success('收藏夹已更新')
    } else {
      await createCollectFolder({
        title: folderForm.title.trim(),
        abstract: folderForm.abstract,
        cover: folderForm.cover,
      })
      ElMessage.success('收藏夹创建成功')
    }
    dialogVisible.value = false
    await loadFolders()
  } catch {
    // ignore
  } finally {
    submitting.value = false
  }
}

async function removeFolder(folder: CollectModel): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `删除后「${folder.title}」里的收藏记录会一并删除,文章本身不受影响。确认删除吗?`,
      '删除收藏夹',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await removeCollectFolders([folder.id])
    ElMessage.success('删除成功')
    if (activeFolderID.value === folder.id) activeFolderID.value = 0
    await loadFolders()
  } catch {
    // ignore
  }
}

onMounted(loadFolders)
</script>

<style scoped lang="scss">
.collect-side {
  padding: 0 0 8px;
}

.collect-side__body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px;
  min-height: 200px;
}

.collect-side__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 10px;
  font-size: 13px;
  color: var(--sd-text-muted);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.collect-side__item:hover {
  background: rgba(34, 211, 238, 0.08);
  color: var(--sd-text);
}

.collect-side__item.is-active {
  color: var(--sd-cyan);
  background: linear-gradient(92deg, rgba(34, 211, 238, 0.16), rgba(168, 85, 247, 0.12));
  box-shadow: inset 0 0 0 1px rgba(34, 211, 238, 0.35);
}

.collect-side__name {
  flex: 1;
  min-width: 0;
}

.collect-side__op {
  font-size: 14px;
  opacity: 0.6;
}

.collect-side__op:hover {
  opacity: 1;
  color: var(--sd-cyan);
}

.collect-side__op--danger:hover {
  color: var(--sd-red);
}

.collect-main__count {
  font-size: 12px;
}

.collect-main__search {
  width: 220px;
}

.collect-main__list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 160px;
}
</style>
