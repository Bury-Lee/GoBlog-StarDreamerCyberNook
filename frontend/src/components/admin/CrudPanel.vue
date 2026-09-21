<template>
  <div class="crud-panel sd-panel">
    <header class="crud-panel__header">
      <div class="crud-panel__heading">
        <h3 class="crud-panel__title">{{ title }}</h3>
        <p v-if="description" class="crud-panel__desc sd-dim">{{ description }}</p>
      </div>
      <div class="crud-panel__tools">
        <el-input
          v-if="searchable"
          v-model="keyword"
          class="crud-panel__search"
          :placeholder="searchPlaceholder"
          clearable
          @keyup.enter="search"
          @clear="search"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button v-if="createFn" type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
        <el-button type="danger" plain :disabled="!selection.length" @click="removeSelected">
          <el-icon><Delete /></el-icon>
          批量删除{{ selection.length ? '(' + selection.length + ')' : '' }}
        </el-button>
      </div>
    </header>

    <div class="crud-panel__table">
      <el-table
        v-loading="loading"
        :data="list"
        row-key="id"
        border
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="46" reserve-selection />
        <el-table-column
          v-for="column in columns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :width="column.width"
          :min-width="column.minWidth"
          :show-overflow-tooltip="column.showOverflowTooltip !== false"
        >
          <template #default="{ row }">
            <el-image
              v-if="column.type === 'image'"
              :src="imageSource(row, column)"
              class="crud-panel__image"
              fit="cover"
              :preview-src-list="imageSource(row, column) ? [imageSource(row, column)] : []"
              preview-teleported
            >
              <template #error>
                <span class="sd-dim">{{ column.emptyText || '无图' }}</span>
              </template>
            </el-image>
            <el-switch
              v-else-if="column.type === 'switch' && hasToggle"
              :model-value="Boolean(row[column.prop])"
              :loading="isToggling(row)"
              @change="(value: boolean | string | number) => handleToggle(row, column, Boolean(value))"
            />
            <el-tag v-else-if="column.type === 'switch'" :type="row[column.prop] ? 'success' : 'info'" size="small">
              {{ row[column.prop] ? '是' : '否' }}
            </el-tag>
            <el-tag
              v-else-if="column.type === 'tag'"
              :type="(column.tagType?.(row) as never) || 'info'"
              size="small"
            >
              {{ column.formatter ? column.formatter(row) : row[column.prop] }}
            </el-tag>
            <span v-else-if="column.type === 'date'">{{ formatDate(row[column.prop]) }}</span>
            <span v-else-if="column.type === 'longtext'" class="sd-clamp-2 crud-panel__longtext">
              {{ column.formatter ? column.formatter(row) : row[column.prop] }}
            </span>
            <span
              v-else
              :class="{ 'crud-panel__link': !!column.onClick }"
              @click="column.onClick?.(row)"
            >
              {{ column.formatter ? column.formatter(row) : row[column.prop] }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" :width="actionWidth">
          <template #default="{ row }">
            <slot name="actions" :row="row" />
            <el-button v-if="updateFn" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="removeOne(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <PaginationBar
      :page="page"
      :limit="limit"
      :count="count"
      :page-sizes="[10, 20, 30, 40]"
      @update:page="changePage"
      @update:limit="changeLimit"
    />

    <el-dialog v-model="dialogVisible" :title="dialogTitle" :width="dialogWidth" append-to-body>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item
          v-for="field in fields"
          :key="field.prop"
          :label="field.label"
          :prop="field.prop"
        >
          <el-input
            v-if="!field.type || field.type === 'input'"
            v-model="form[field.prop]"
            :placeholder="field.placeholder || `请输入${field.label}`"
            clearable
          />
          <el-input
            v-else-if="field.type === 'textarea'"
            v-model="form[field.prop]"
            type="textarea"
            :rows="3"
            resize="none"
            :placeholder="field.placeholder || `请输入${field.label}`"
          />
          <el-input
            v-else-if="field.type === 'list'"
            v-model="form[field.prop]"
            type="textarea"
            :rows="3"
            resize="none"
            :placeholder="field.placeholder || '每行一个,或用逗号分隔'"
          />
          <el-input-number
            v-else-if="field.type === 'number'"
            v-model="form[field.prop]"
            :min="field.min ?? 0"
            :max="field.max ?? 999999"
          />
          <el-switch v-else-if="field.type === 'switch'" v-model="form[field.prop]" />
          <el-select
            v-else-if="field.type === 'select'"
            v-model="form[field.prop]"
            class="crud-panel__control"
            clearable
            :placeholder="field.placeholder || `请选择${field.label}`"
          >
            <el-option
              v-for="option in field.options || []"
              :key="String(option.value)"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
          <ImageUploader
            v-else-if="field.type === 'image'"
            v-model="form[field.prop]"
            :tip="field.tip"
          />
          <p v-if="field.tip && field.type !== 'image'" class="crud-panel__tip sd-dim">{{ field.tip }}</p>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Delete, Plus, Search } from '@element-plus/icons-vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import ImageUploader from '@/components/common/ImageUploader.vue'
import { resolveAssetUrl } from '@/api/request'
import type { ListData, PageParams } from '@/api/types'
import { formatDate } from '@/utils/format'

export interface CrudColumn {
  prop: string
  label: string
  width?: string | number
  minWidth?: string | number
  type?: 'text' | 'switch' | 'image' | 'tag' | 'date' | 'number' | 'longtext'
  options?: { label: string; value: unknown }[]
  tagType?: (row: Record<string, any>) => string
  formatter?: (row: Record<string, any>) => string
  onClick?: (row: Record<string, any>) => void
  emptyText?: string
  showOverflowTooltip?: boolean
}

export interface CrudField {
  prop: string
  label: string
  type?: 'input' | 'textarea' | 'number' | 'switch' | 'select' | 'image' | 'list'
  options?: { label: string; value: unknown }[]
  placeholder?: string
  required?: boolean
  tip?: string
  min?: number
  max?: number
  default?: unknown
}

const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    columns: CrudColumn[]
    fields: CrudField[]
    listFn: (params: PageParams) => Promise<ListData<Record<string, any>>>
    createFn?: (payload: Record<string, any>) => Promise<unknown>
    updateFn?: (id: number, payload: Record<string, any>) => Promise<unknown>
    removeFn: (ids: number[]) => Promise<unknown>
    onToggle?: (row: Record<string, any>, value: boolean) => Promise<void>
    searchable?: boolean
    searchPlaceholder?: string
    extraParams?: Record<string, unknown>
    defaultForm?: () => Record<string, any>
    dialogWidth?: string
    actionWidth?: number
  }>(),
  {
    description: '',
    searchable: true,
    searchPlaceholder: '搜索关键词',
    extraParams: () => ({}),
    defaultForm: () => ({}),
    dialogWidth: '620px',
    actionWidth: 160,
  },
)

const emit = defineEmits<{
  (e: 'loaded', list: Record<string, any>[]): void
  (e: 'saved'): void
}>()

const list = ref<Record<string, any>[]>([])
const count = ref(0)
const page = ref(1)
const limit = ref(10)
const keyword = ref('')
const loading = ref(false)
const selection = ref<Record<string, any>[]>([])
const dialogVisible = ref(false)
const submitting = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const form = reactive<Record<string, any>>({})
const toggleLoading = ref<Set<number>>(new Set())

const dialogTitle = computed(() => (editingId.value ? `编辑${props.title}` : `新增${props.title}`))
const hasToggle = computed(() => typeof props.onToggle === 'function')

const rules = computed<FormRules>(() => {
  const result: FormRules = {}
  props.fields.forEach((field) => {
    if (field.required) {
      result[field.prop] = [
        { required: true, message: `${field.label}不能为空`, trigger: field.type === 'select' ? 'change' : 'blur' },
      ]
    }
  })
  return result
})

async function load(): Promise<void> {
  loading.value = true
  try {
    const data = await props.listFn({
      page: page.value,
      limit: limit.value,
      key: keyword.value || undefined,
      ...props.extraParams,
    })
    list.value = (data?.list ?? []) as Record<string, any>[]
    count.value = data?.count ?? 0
    emit('loaded', list.value)
  } catch {
    list.value = []
    count.value = 0
  } finally {
    loading.value = false
  }
}

function search(): void {
  page.value = 1
  void load()
}

function changePage(next: number): void {
  page.value = next
  void load()
}

function changeLimit(next: number): void {
  limit.value = next
  page.value = 1
  void load()
}

function onSelectionChange(rows: Record<string, any>[]): void {
  selection.value = rows
}

function imageSource(row: Record<string, any>, column: CrudColumn): string {
  const value = column.formatter ? column.formatter(row) : row[column.prop]
  return resolveAssetUrl(String(value || ''))
}

function isToggling(row: Record<string, any>): boolean {
  return toggleLoading.value.has(Number(row.id))
}

async function handleToggle(row: Record<string, any>, column: CrudColumn, value: boolean): Promise<void> {
  if (!props.onToggle) return
  const previous = Boolean(row[column.prop])
  if (previous === value) return
  row[column.prop] = value
  const id = Number(row.id)
  toggleLoading.value.add(id)
  try {
    await props.onToggle(row, value)
  } catch {
    row[column.prop] = previous
    ElMessage.error('操作失败,已恢复原状态')
  } finally {
    toggleLoading.value.delete(id)
  }
}

function describeRow(row: Record<string, any>): string {
  const label = row.name || row.title
  return label ? `《${label}》` : `ID ${row.id}`
}

function resetForm(): void {
  Object.keys(form).forEach((key) => delete form[key])
  const defaults = props.defaultForm()
  props.fields.forEach((field) => {
    if (field.type === 'list') {
      form[field.prop] = ''
      return
    }
    form[field.prop] =
      defaults[field.prop] !== undefined ? defaults[field.prop] : (field.default ?? defaultForType(field))
  })
  Object.entries(defaults).forEach(([key, value]) => {
    form[key] = value
  })
}

function defaultForType(field: CrudField): unknown {
  if (field.type === 'switch') return false
  if (field.type === 'number') return 0
  if (field.type === 'select') return ''
  return ''
}

function fillForm(row: Record<string, any>): void {
  resetForm()
  props.fields.forEach((field) => {
    const value = row[field.prop]
    if (field.type === 'list') {
      form[field.prop] = Array.isArray(value) ? value.join('\n') : String(value ?? '')
    } else if (value !== undefined && value !== null) {
      form[field.prop] = value
    }
  })
  Object.keys(row).forEach((key) => {
    if (form[key] === undefined) form[key] = row[key]
  })
}

function openCreate(): void {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: Record<string, any>): void {
  editingId.value = Number(row.id)
  fillForm(row)
  dialogVisible.value = true
}

function buildPayload(): Record<string, any> {
  const payload: Record<string, any> = { ...form }
  props.fields.forEach((field) => {
    if (field.type === 'list') {
      const raw = String(payload[field.prop] ?? '')
      payload[field.prop] = raw
        .split(/[\n,]/)
        .map((item) => item.trim())
        .filter(Boolean)
    }
    if (field.type === 'number') {
      payload[field.prop] = Number(payload[field.prop] || 0)
    }
  })
  return payload
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const payload = buildPayload()
  submitting.value = true
  try {
    if (editingId.value) {
      await props.updateFn?.(editingId.value, payload)
      ElMessage.success('更新成功')
    } else {
      await props.createFn?.(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    emit('saved')
    await load()
  } catch {
    // 错误提示已由请求层处理
  } finally {
    submitting.value = false
  }
}

async function removeOne(row: Record<string, any>): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除${describeRow(row)}?`, '删除确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await props.removeFn([Number(row.id)])
    ElMessage.success('删除成功')
    await load()
  } catch {
    // ignore
  }
}

async function removeSelected(): Promise<void> {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 条记录?`, '批量删除', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await props.removeFn(selection.value.map((row) => Number(row.id)))
    ElMessage.success('删除成功')
    selection.value = []
    await load()
  } catch {
    // ignore
  }
}

defineExpose({ refresh: load, openCreate })
void load()
</script>

<style scoped lang="scss">
.crud-panel {
  padding: 18px;
}

.crud-panel__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.crud-panel__title {
  font-size: 16px;
}

.crud-panel__desc {
  margin-top: 4px;
  font-size: 12px;
}

.crud-panel__tools {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.crud-panel__search {
  width: 220px;
}

.crud-panel__table {
  overflow-x: auto;
}

.crud-panel__image {
  width: 64px;
  height: 40px;
  border-radius: 6px;
}

.crud-panel__longtext {
  font-size: 12px;
  color: var(--sd-text-muted);
  max-width: 320px;
}

.crud-panel__link {
  color: var(--sd-cyan);
  cursor: pointer;
}

.crud-panel__link:hover {
  text-decoration: underline;
}

.crud-panel__control {
  width: 100%;
}

.crud-panel__tip {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.6;
}
</style>
