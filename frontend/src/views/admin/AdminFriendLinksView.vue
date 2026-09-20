<template>
  <CrudPanel
    title="友情链接"
    description="首页与关于页展示的友链,仅显示 is_show = true 的条目"
    :columns="columns"
    :fields="fields"
    :list-fn="fetchFriendLinks"
    :create-fn="createFriendLink"
    :update-fn="updateFriendLink"
    :remove-fn="removeFriendLinks"
    :default-form="defaultForm"
  />
</template>

<script setup lang="ts">
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import {
  createFriendLink,
  fetchFriendLinks,
  removeFriendLinks,
  updateFriendLink,
} from '@/api/ops'

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

function defaultForm(): Record<string, unknown> {
  return { is_show: true, sort_order: 0 }
}
</script>
