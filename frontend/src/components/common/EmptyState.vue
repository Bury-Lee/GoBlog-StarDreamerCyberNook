<template>
  <div class="empty-state" :style="{ padding: `${compact ? 18 : 40}px 0` }">
    <el-icon class="empty-state__icon" :size="compact ? 26 : 40">
      <component :is="icon" />
    </el-icon>
    <p class="empty-state__text">{{ text }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { DocumentDelete } from '@element-plus/icons-vue'

withDefaults(
  defineProps<{
    text?: string
    icon?: unknown
    compact?: boolean
  }>(),
  {
    text: '暂无数据',
    icon: () => DocumentDelete,
    compact: false,
  },
)
</script>

<style scoped lang="scss">
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  text-align: center;
}

.empty-state__icon {
  color: rgba(100, 116, 139, 0.75);
}

.empty-state__text {
  font-size: 13px;
  color: var(--sd-text-dim);
}
</style>
