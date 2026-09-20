<template>
  <section class="tag-cloud sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">标签云</span>
      <el-button text size="small" @click="reload">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </header>
    <div class="sd-panel__body">
      <el-skeleton v-if="loading" :rows="3" animated />
      <EmptyState v-else-if="!tags.length" text="暂无标签" compact />
      <div v-else class="tag-cloud__items">
        <span
          v-for="item in tags"
          :key="item.name"
          class="sd-tag"
          :class="{ 'sd-tag--purple': item.count > 2 }"
          @click="goTag(item.name)"
        >
          # {{ item.name }}
          <em class="tag-cloud__count">{{ item.count }}</em>
        </span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { fetchArticleList } from '@/api/article'

const router = useRouter()
const loading = ref(true)
const rawTags = ref<string[]>([])

const tags = computed(() => {
  const map = new Map<string, number>()
  rawTags.value.forEach((tag) => {
    const key = String(tag || '').trim()
    if (!key) return
    map.set(key, (map.get(key) || 0) + 1)
  })
  return Array.from(map.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 24)
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

.tag-cloud__count {
  font-style: normal;
  font-size: 11px;
  opacity: 0.7;
}
</style>
