<template>
  <div v-if="count > 0" class="pagination-bar">
    <el-pagination
      background
      :current-page="page"
      :page-size="limit"
      :page-sizes="pageSizes"
      :total="count"
      :layout="layout"
      :pager-count="5"
      :hide-on-single-page="hideOnSinglePage"
      @current-change="onPageChange"
      @size-change="onSizeChange"
    />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    page: number
    limit: number
    count: number
    pageSizes?: number[]
    layout?: string
    hideOnSinglePage?: boolean
  }>(),
  {
    pageSizes: () => [10, 20, 30, 40],
    layout: 'total, sizes, prev, pager, next, jumper',
    hideOnSinglePage: false,
  },
)

const emit = defineEmits<{
  (e: 'update:page', value: number): void
  (e: 'update:limit', value: number): void
}>()

function onPageChange(value: number): void {
  emit('update:page', value)
}

function onSizeChange(value: number): void {
  emit('update:limit', value)
}
</script>

<style scoped lang="scss">
.pagination-bar {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
