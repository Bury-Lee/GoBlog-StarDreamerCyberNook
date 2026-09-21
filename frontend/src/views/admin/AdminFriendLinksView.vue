<template>
  <CrudPanel
    title="友情链接"
    description="首页与关于页展示的友链,关闭显示的条目不会出现在前台"
    :columns="columns"
    :fields="fields"
    :list-fn="listFriendLinks"
    :create-fn="createFriendLink"
    :update-fn="updateFriendLink"
    :remove-fn="removeFriendLinks"
    :default-form="defaultForm"
    :on-toggle="toggleFriendLink"
  />
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import {
  createFriendLink,
  fetchFriendLinks,
  removeFriendLinks,
  updateFriendLink,
} from '@/api/ops'
import type { FriendLink, ListData, PageParams } from '@/api/types'

const columns: CrudColumn[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'logo', label: 'Logo', type: 'image', width: 90 },
  { prop: 'name', label: '名称', width: 150 },
  { prop: 'url', label: '链接', minWidth: 220 },
  { prop: 'sort_order', label: '排序', width: 80 },
  { prop: 'is_show', label: '显示', type: 'switch', width: 80 },
  { prop: 'remark', label: '备注', minWidth: 140 },
]

const fields: CrudField[] = [
  { prop: 'name', label: '站点名称', required: true },
  { prop: 'url', label: '站点链接', required: true, placeholder: 'https://example.com' },
  { prop: 'logo', label: 'Logo', type: 'image' },
  { prop: 'sort_order', label: '排序值', type: 'number', default: 0, tip: '数值越小越靠前' },
  { prop: 'is_show', label: '是否显示', type: 'switch', default: true },
  { prop: 'remark', label: '备注', type: 'textarea' },
]

function listFriendLinks(params: PageParams): Promise<ListData<FriendLink>> {
  return fetchFriendLinks({ ...params, all: 1 })
}

async function toggleFriendLink(row: Record<string, any>, value: boolean): Promise<void> {
  await updateFriendLink(Number(row.id), {
    name: String(row.name || ''),
    url: String(row.url || ''),
    logo: String(row.logo || ''),
    is_show: value,
    sort_order: Number(row.sort_order || 0),
    remark: String(row.remark || ''),
  })
  ElMessage.success(value ? '已显示' : '已隐藏')
}

function defaultForm(): Record<string, unknown> {
  return { is_show: true, sort_order: 0 }
}
</script>
