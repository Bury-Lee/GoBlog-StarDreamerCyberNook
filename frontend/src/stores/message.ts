import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { checkUnreadMessages } from '@/api/message'
import { useUserStore } from './user'

export const useMessageStore = defineStore('message', () => {
  const unread = ref<Record<string, number>>({})
  const loading = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null

  const total = computed(() =>
    Object.values(unread.value).reduce((sum, value) => sum + Number(value || 0), 0),
  )

  function countOf(type: number): number {
    return Number(unread.value[String(type)] || 0)
  }

  async function refresh(): Promise<void> {
    const userStore = useUserStore()
    if (!userStore.isLogin) {
      unread.value = {}
      return
    }
    loading.value = true
    try {
      const data = await checkUnreadMessages()
      unread.value = data || {}
    } catch {
      unread.value = {}
    } finally {
      loading.value = false
    }
  }

  function startPolling(interval = 60000): void {
    stopPolling()
    void refresh()
    timer = setInterval(() => {
      void refresh()
    }, interval)
  }

  function stopPolling(): void {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  function clear(): void {
    unread.value = {}
  }

  return { unread, loading, total, countOf, refresh, startPolling, stopPolling, clear }
})
