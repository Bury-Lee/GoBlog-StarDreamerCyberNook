import { ref, type Ref } from 'vue'
import type { ListData, PageParams } from '@/api/types'

export interface UsePaginationOptions {
  limit?: number
  immediate?: boolean
}

export function usePagination<T, P extends object = PageParams>(
  fetcher: (params: P & PageParams) => Promise<ListData<T>>,
  baseParams: () => P,
  options: UsePaginationOptions = {},
) {
  const list = ref([]) as Ref<T[]>
  const count = ref(0)
  const page = ref(1)
  const limit = ref(options.limit ?? 10)
  const key = ref('')
  const loading = ref(false)
  const error = ref('')

  async function load(): Promise<void> {
    loading.value = true
    error.value = ''
    try {
      const params = {
        ...baseParams(),
        page: page.value,
        limit: limit.value,
        key: key.value || undefined,
      } as P & PageParams
      const data = await fetcher(params)
      list.value = data?.list ?? []
      count.value = data?.count ?? 0
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载失败'
      list.value = []
      count.value = 0
    } finally {
      loading.value = false
    }
  }

  function search(): void {
    page.value = 1
    void load()
  }

  function reset(): void {
    key.value = ''
    page.value = 1
    limit.value = options.limit ?? 10
    void load()
  }

  function changePage(next: number): void {
    page.value = next
    void load()
  }

  function changeLimit(next: number): void {
    limit.value = next
    page.value = 1
    void load()
  }

  if (options.immediate !== false) {
    void load()
  }

  return {
    list,
    count,
    page,
    limit,
    key,
    loading,
    error,
    load,
    refresh: load,
    search,
    reset,
    changePage,
    changeLimit,
  }
}
