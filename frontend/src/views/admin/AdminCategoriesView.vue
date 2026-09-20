<template>
  <CrudPanel
    title="分类管理"
    description="管理员可查看全站分类,新建与修改共用同一接口(id = 0 为新建)"
    :columns="columns"
    :fields="fields"
    :list-fn="listCategories"
    :create-fn="createCategory"
    :update-fn="updateCategory"
    :remove-fn="removeCategories"
    :action-width="140"
  />
</template>

<script setup lang="ts">
import CrudPanel, { type CrudColumn, type CrudField } from '@/components/admin/CrudPanel.vue'
import { fetchCategories, removeCategories, saveCategory } from '@/api/article'
import type { ListData, PageParams } from '@/api/types'

const columns: CrudColumn[] = [
  { prop: 'id', label: 'ID', width: 70 },
  { prop: 'title', label: '分类名称', minWidth: 160 },
  { prop: 'articleCount', label: '文章数', width: 100 },
  { prop: 'userID', label: '归属用户', width: 100 },
  { prop: 'nickname', label: '用户昵称', width: 140 },
  { prop: 'createdAt', label: '创建时间', type: 'date', width: 170 },
]

const fields: CrudField[] = [{ prop: 'title', label: '分类名称', required: true }]

async function listCategories(params: PageParams): Promise<ListData<Record<string, unknown>>> {
  const data = await fetchCategories({ ...params, type: 'admin' })
  return data as unknown as ListData<Record<string, unknown>>
}

async function createCategory(payload: Record<string, unknown>): Promise<unknown> {
  return saveCategory({ id: 0, title: String(payload.title || '') })
}

async function updateCategory(id: number, payload: Record<string, unknown>): Promise<unknown> {
  return saveCategory({ id, title: String(payload.title || '') })
}
</script>
