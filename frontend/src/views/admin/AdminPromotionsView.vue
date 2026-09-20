<template>
  <CrudPanel
    title="友站推广"
    description="首页/内页推广位,preview_images 使用 JSON 数组字符串,contact_info 每行一个链接"
    :columns="columns"
    :fields="fields"
    :list-fn="fetchFriendPromotions"
    :create-fn="createFriendPromotion"
    :update-fn="updateFriendPromotion"
    :remove-fn="removeFriendPromotions"
    :default-form="defaultForm"
    dialog-width="680px"
  />
</template>

<script setup lang="ts">
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import {
  createFriendPromotion,
  fetchFriendPromotions,
  removeFriendPromotions,
  updateFriendPromotion,
} from '@/api/ops'

const columns: CrudColumn[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'avatar', label: '头像', type: 'image', width: 80 },
  { prop: 'title', label: '标题', minWidth: 160 },
  { prop: 'friend_name', label: '友站名', width: 130 },
  { prop: 'category', label: '分类', width: 100 },
  { prop: 'description', label: '描述', type: 'longtext', minWidth: 200 },
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
    tip: 'JSON 数组字符串,如 ["/api/image?id=1","/api/image?id=2"]',
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

function defaultForm(): Record<string, unknown> {
  return { is_show: true, sort_order: 0, position: 'home', contact_info: [] }
}
</script>
