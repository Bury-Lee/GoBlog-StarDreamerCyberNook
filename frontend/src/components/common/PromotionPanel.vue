<template>
  <section v-if="promotions.length || loading" class="promotion-panel sd-panel">
    <header class="sd-panel__header">
      <span class="sd-panel__title">{{ title }}</span>
      <span class="sd-chip">友站推荐</span>
    </header>
    <div class="sd-panel__body">
      <el-skeleton v-if="loading" :rows="3" animated />
      <div v-else class="promotion-panel__grid">
        <article v-for="item in promotions" :key="item.id" class="promotion-panel__card">
          <UserAvatar :src="item.avatar" :name="item.friend_name || item.title" :size="42" />
          <div class="promotion-panel__main">
            <div class="promotion-panel__head">
              <span class="promotion-panel__title-text sd-ellipsis">{{ item.title }}</span>
              <span v-if="item.category" class="sd-tag sd-tag--purple">{{ item.category }}</span>
            </div>
            <p class="promotion-panel__desc sd-clamp-2">{{ item.description }}</p>
            <div v-if="contacts(item).length" class="promotion-panel__contacts">
              <a
                v-for="(link, index) in contacts(item)"
                :key="index"
                class="sd-link promotion-panel__contact"
                :href="link"
                target="_blank"
                rel="noopener noreferrer"
              >
                <el-icon><Link /></el-icon>
                {{ shortLink(link) }}
              </a>
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Link } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { fetchFriendPromotions } from '@/api/ops'
import type { FriendPromotion } from '@/api/types'

const props = withDefaults(
  defineProps<{
    title?: string
    position?: 'home' | 'page' | 'both'
  }>(),
  { title: '友站推广', position: 'home' },
)

const promotions = ref<FriendPromotion[]>([])
const loading = ref(true)

function contacts(item: FriendPromotion): string[] {
  return Array.isArray(item.contact_info) ? item.contact_info.filter(Boolean) : []
}

function shortLink(link: string): string {
  return link.replace(/^https?:\/\//, '').replace(/\/$/, '').slice(0, 32)
}

onMounted(async () => {
  try {
    const data = await fetchFriendPromotions({ page: 1, limit: 12 })
    promotions.value = (data?.list ?? []).filter(
      (item) => item.position === props.position || item.position === 'both' || !item.position,
    )
  } catch {
    promotions.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss">
.promotion-panel__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px;
}

.promotion-panel__card {
  display: flex;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--sd-border);
  border-radius: 12px;
  background: rgba(9, 14, 26, 0.6);
  transition: all 0.25s ease;
}

.promotion-panel__card:hover {
  border-color: rgba(34, 211, 238, 0.45);
  transform: translateY(-2px);
}

.promotion-panel__main {
  flex: 1;
  min-width: 0;
}

.promotion-panel__head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.promotion-panel__title-text {
  font-size: 14px;
  font-weight: 600;
}

.promotion-panel__desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--sd-text-muted);
}

.promotion-panel__contacts {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
  font-size: 12px;
}

.promotion-panel__contact {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
</style>
