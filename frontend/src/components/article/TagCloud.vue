<template>
  <section class="tag-cloud sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">{{ title }}</span>
      <el-button text size="small" @click="reload">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </header>
    <div class="sd-panel__body">
      <el-skeleton v-if="loading" :rows="3" animated />
      <EmptyState v-else-if="!tags.length" text="暂无标签" compact />
      <template v-else>
        <div class="tag-cloud__items">
          <span v-for="tag in tags" :key="tag" class="sd-tag" @click="goTag(tag)"># {{ tag }}</span>
        </div>
        <p class="sd-dim tag-cloud__hint">来自最近发布的文章</p>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchArticleList } from '@/api/article'

withDefaults(defineProps<{ title?: string }>(), { title: '标签云' })

const router = useRouter()
const loading = ref(true)
const rawTags = ref<string[]>([])

const tags = computed(() => {
  const result: string[] = []
  rawTags.value.forEach((tag) => {
    const name = String(tag || '').trim()
    if (name && !result.includes(name)) result.push(name)
  })
  return result.slice(0, 24)
})

function goTag(tag: string): void {
  router.push({ name: 'search', query: { tag } })
}

async function reload(): Promise<void> {
  loading.value = true
  try {
    const data = await fetchArticleList({ type: 'other', page: 1, limit: 40 })
    rawTags.value = (data?.list ?? []).flatMap((item) => item.tagList || [])
  } catch {
    rawTags.value = []
  } finally {
    loading.value = false
  }
}

onMounted(reload)
</script>

<style scoped lang="scss">
.tag-cloud__items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-cloud__items .sd-tag {
  cursor: pointer;
}

.tag-cloud__hint {
  margin-top: 12px;
  font-size: 12px;
}
</style>
