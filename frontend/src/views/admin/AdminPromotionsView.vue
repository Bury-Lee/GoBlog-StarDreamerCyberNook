<template>
  <CrudPanel
    title="友站推广"
    description="首页与内页的推广位,关闭显示的条目不会出现在前台"
    :columns="columns"
    :fields="fields"
    :list-fn="listPromotions"
    :create-fn="createFriendPromotion"
    :update-fn="updateFriendPromotion"
    :remove-fn="removeFriendPromotions"
    :default-form="defaultForm"
    :on-toggle="togglePromotion"
    dialog-width="680px"
  />
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import {
  createFriendPromotion,
  fetchFriendPromotions,
  removeFriendPromotions,
  updateFriendPromotion,
} from '@/api/ops'
import type { FriendPromotion, ListData, PageParams } from '@/api/types'

function firstPreviewImage(row: Record<string, any>): string {
  try {
    const parsed = JSON.parse(String(row.preview_images || ''))
    if (Array.isArray(parsed) && parsed.length) return String(parsed[0])
  } catch {
    return ''
  }
  return ''
}

const columns: CrudColumn[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'avatar', label: '头像', type: 'image', width: 80 },
  { prop: 'title', label: '标题', minWidth: 160 },
  { prop: 'friend_name', label: '友站名', width: 130 },
  { prop: 'category', label: '分类', width: 100 },
  { prop: 'description', label: '描述', type: 'longtext', minWidth: 200 },
  {
    prop: 'preview_images',
    label: '预览图',
    type: 'image',
    width: 90,
    emptyText: '-',
    formatter: firstPreviewImage,
  },
  { prop: 'position', label: '位置', width: 100 },
  { prop: 'sort_order', label: '排序', width: 80 },
  { prop: 'is_show', label: '显示', type: 'switch', width: 80 },
]

const fields: CrudField[] = [
  { prop: 'title', label: '推广标题', required: true },
  { prop: 'friend_name', label: '友站名称' },
  { prop: 'avatar', label: '头像', type: 'image' },
  { prop: 'category', label: '分类', placeholder: '如:技术 / 设计' },
  { prop: 'description', label: '描述', type: 'textarea' },
  {
    prop: 'preview_images',
    label: '预览图',
    type: 'textarea',
    placeholder: '例如 ["/api/image?id=1", "/api/image?id=2"]',
    tip: '需要展示多张预览图时,按上面的数组格式填写图片地址',
  },
  { prop: 'contact_info', label: '联系方式', type: 'list', tip: '每行一个链接或账号' },
  {
    prop: 'position',
    label: '展示位置',
    type: 'select',
    default: 'home',
    options: [
      { label: '首页', value: 'home' },
      { label: '内页', value: 'page' },
      { label: '两者都显示', value: 'both' },
    ],
  },
  { prop: 'sort_order', label: '排序值', type: 'number', default: 0 },
  { prop: 'is_show', label: '是否显示', type: 'switch', default: true },
  { prop: 'remark', label: '备注', type: 'textarea' },
]

function listPromotions(params: PageParams): Promise<ListData<FriendPromotion>> {
  return fetchFriendPromotions({ ...params, all: 1 })
}

async function togglePromotion(row: Record<string, any>, value: boolean): Promise<void> {
  await updateFriendPromotion(Number(row.id), {
    title: String(row.title || ''),
    friend_name: String(row.friend_name || ''),
    avatar: String(row.avatar || ''),
    category: String(row.category || ''),
    description: String(row.description || ''),
    preview_images: String(row.preview_images || ''),
    contact_info: Array.isArray(row.contact_info) ? row.contact_info : [],
    is_show: value,
    sort_order: Number(row.sort_order || 0),
    position: String(row.position || ''),
    remark: String(row.remark || ''),
  })
  ElMessage.success(value ? '已显示' : '已隐藏')
}

function defaultForm(): Record<string, unknown> {
  return { is_show: true, sort_order: 0, position: 'home', contact_info: [] }
}
</script>
