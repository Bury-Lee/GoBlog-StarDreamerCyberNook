<template>
  <CrudPanel
    title="轮播图"
    description="首页顶部轮播图,关闭显示的条目不会出现在前台"
    :columns="columns"
    :fields="fields"
    :list-fn="listBanners"
    :create-fn="createBanner"
    :update-fn="updateBanner"
    :remove-fn="removeBanners"
    :default-form="defaultForm"
    :on-toggle="toggleBanner"
  />
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import { createBanner, fetchBanners, removeBanners, updateBanner } from '@/api/ops'
import type { Banner, ListData, PageParams } from '@/api/types'

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

function listBanners(params: PageParams): Promise<ListData<Banner>> {
  return fetchBanners({ ...params, all: 1 })
}

async function toggleBanner(row: Record<string, any>, value: boolean): Promise<void> {
  await updateBanner(Number(row.id), {
    cover: String(row.cover || ''),
    href: String(row.href || ''),
    isShow: value,
  })
  ElMessage.success(value ? '已显示' : '已隐藏')
}

function defaultForm(): Record<string, unknown> {
  return { isShow: true, cover: '', href: '' }
}
</script>
