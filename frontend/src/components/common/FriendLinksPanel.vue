<template>
  <section class="friend-links sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">{{ title }}</span>
    </header>
    <div class="sd-panel__body">
      <el-skeleton v-if="loading" :rows="3" animated />
      <EmptyState v-else-if="!links.length" text="暂无友链" compact />
      <div v-else class="friend-links__items">
        <a
          v-for="link in links"
          :key="link.id"
          class="friend-links__item"
          :href="link.url || '#'"
          target="_blank"
          rel="noopener noreferrer"
          :title="link.remark || link.name"
        >
          <UserAvatar :src="link.logo" :name="link.name" :size="30" />
          <span class="friend-links__name sd-ellipsis">{{ link.name }}</span>
        </a>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import EmptyState from '@/components/common/EmptyState.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { fetchFriendLinks } from '@/api/ops'
import type { FriendLink } from '@/api/types'

withDefaults(defineProps<{ title?: string }>(), { title: '友情链接' })

const links = ref<FriendLink[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const data = await fetchFriendLinks({ page: 1, limit: 20 })
    links.value = data?.list ?? []
  } catch {
    links.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.friend-links__items {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.friend-links__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.friend-links__item:hover {
  background: rgba(34, 211, 238, 0.08);
}

.friend-links__name {
  font-size: 13px;
  color: var(--sd-text-muted);
}

.friend-links__item:hover .friend-links__name {
  color: var(--sd-cyan);
}
</style>
