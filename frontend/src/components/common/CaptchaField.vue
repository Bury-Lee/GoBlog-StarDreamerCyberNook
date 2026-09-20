<template>
  <div class="captcha-field">
    <el-input
      v-model="code"
      :placeholder="placeholder"
      maxlength="8"
      clearable
      @input="onInput"
    >
      <template #prefix>
        <el-icon><Key /></el-icon>
      </template>
    </el-input>
    <div class="captcha-field__image" :class="{ 'is-loading': loading }" @click="onRefresh">
      <img v-if="image" :src="image" alt="captcha" />
      <span v-else class="sd-dim">{{ loading ? '加载中' : '点击获取' }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Key } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    image?: string
    loading?: boolean
    placeholder?: string
  }>(),
  {
    modelValue: '',
    image: '',
    loading: false,
    placeholder: '图形验证码',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'refresh'): void
}>()

const code = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value),
})

function onInput(value: string): void {
  emit('update:modelValue', value)
}

function onRefresh(): void {
  emit('refresh')
}
</script>

<style scoped lang="scss">
.captcha-field {
  display: flex;
  gap: 10px;
  width: 100%;
}

.captcha-field__image {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 118px;
  height: 32px;
  overflow: hidden;
  border: 1px solid var(--sd-border);
  border-radius: 8px;
  background: rgba(9, 14, 26, 0.85);
  cursor: pointer;
  font-size: 12px;
  flex-shrink: 0;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.captcha-field__image:hover {
  border-color: rgba(34, 211, 238, 0.6);
}
</style>
