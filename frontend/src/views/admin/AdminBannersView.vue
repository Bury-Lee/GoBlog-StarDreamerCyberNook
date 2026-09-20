<template>
  <CrudPanel
    title="轮播图"
    description="首页顶部轮播,仅显示 isShow = true 的条目"
    :columns="columns"
    :fields="fields"
    :list-fn="fetchBanners"
    :create-fn="createBanner"
    :update-fn="updateBanner"
    :remove-fn="removeBanners"
    :default-form="defaultForm"
  />
</template>

<script setup lang="ts">
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import { createBanner, fetchBanners, removeBanners, updateBanner } from '@/api/ops'

const columns: CrudColumn[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'cover', label: '封面', type: 'image', width: 110 },
  { prop: 'href', label: '跳转链接', minWidth: 240 },
  { prop: 'isShow', label: '显示', type: 'switch', width: 90 },
  { prop: 'createdAt', label: '创建时间', type: 'date', width: 170 },
]

const fields: CrudField[] = [
  { prop: 'cover', label: '封面图', type: 'image', required: true, tip: '建议尺寸 1200x400 以上' },
  { prop: 'href', label: '跳转链接', placeholder: 'https://example.com/article/1' },
  { prop: 'isShow', label: '是否显示', type: 'switch', default: true },
]

function defaultForm(): Record<string, unknown> {
  return { isShow: true, cover: '', href: '' }
}
</script>
