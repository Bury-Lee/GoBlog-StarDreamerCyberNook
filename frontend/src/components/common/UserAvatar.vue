<template>
  <router-link
    v-if="userId > 0"
    class="user-avatar-link"
    :to="{ name: 'user-home', params: { id: userId } }"
    @click.stop
  >
    <span
      class="user-avatar"
      :style="{ width: `${size}px`, height: `${size}px`, fontSize: `${Math.max(11, size * 0.4)}px` }"
    >
      <img v-if="src && !failed" :src="resolvedSrc" :alt="name" @error="failed = true" />
      <span v-else class="user-avatar__fallback">{{ initial }}</span>
    </span>
  </router-link>
  <span
    v-else
    class="user-avatar"
    :style="{ width: `${size}px`, height: `${size}px`, fontSize: `${Math.max(11, size * 0.4)}px` }"
  >
    <img v-if="src && !failed" :src="resolvedSrc" :alt="name" @error="failed = true" />
    <span v-else class="user-avatar__fallback">{{ initial }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { resolveAssetUrl } from '@/api/request'

const props = withDefaults(
  defineProps<{
    src?: string
    name?: string
    size?: number
    /** 提供用户ID时,头像可点击跳转到该用户主页 */
    userId?: number
  }>(),
  {
    src: '',
    name: '',
    size: 36,
    userId: 0,
  },
)

const failed = ref(false)

watch(
  () => props.src,
  () => {
    failed.value = false
  },
)

const resolvedSrc = computed(() => resolveAssetUrl(props.src))

const initial = computed(() => {
  const name = (props.name || '').trim()
  if (!name) return '?'
  return name.slice(0, 1).toUpperCase()
})
</script>

<style scoped lang="scss">
.user-avatar-link {
  display: inline-flex;
  flex-shrink: 0;
}

.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
  border-radius: 50%;
  border: 1px solid rgba(34, 211, 238, 0.35);
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.25), rgba(168, 85, 247, 0.25));

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.user-avatar__fallback {
  font-weight: 600;
  color: #e0f2fe;
  text-shadow: 0 0 10px rgba(34, 211, 238, 0.8);
}
</style>
