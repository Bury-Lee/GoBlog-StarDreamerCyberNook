import { describe, expect, it, vi } from 'vitest'
import { usePagination } from '@/composables/usePagination'

type MockFetcher = ReturnType<typeof vi.fn<(params: Record<string, any>) => Promise<{ list: never[]; count: number }>>>

function createFetcher(): MockFetcher {
  return vi.fn(async (_params: Record<string, any>) => ({ list: [], count: 0 }))
}

describe('usePagination 参数拼装', () => {
  it('保留 baseParams 自带的关键词(搜索页场景,回归:曾被内部空 key 覆盖)', async () => {
    const fetcher = createFetcher()
    const { load } = usePagination(
      fetcher as never,
      () => ({ key: 'GoTenon', type: 'other', tag: 'GO' }),
      { immediate: false },
    )

    await load()

    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(fetcher.mock.calls[0][0]).toMatchObject({
      key: 'GoTenon',
      type: 'other',
      tag: 'GO',
      page: 1,
      limit: 10,
    })
  })

  it('内部关键词非空时优先生效(列表页搜索框场景)', async () => {
    const fetcher = createFetcher()
    const { key, load } = usePagination(fetcher as never, () => ({ type: 'other' }), { immediate: false })

    key.value = '内部关键词'
    await load()

    expect(fetcher.mock.calls[0][0]).toMatchObject({ key: '内部关键词' })
  })

  it('内部关键词为空时不发送 key 字段', async () => {
    const fetcher = createFetcher()
    const { load } = usePagination(fetcher as never, () => ({ type: 'other' }), { immediate: false })

    await load()

    expect(fetcher.mock.calls[0][0]).not.toHaveProperty('key')
  })

  it('分页与每页条数正确拼装', async () => {
    const fetcher = createFetcher()
    const { load, changePage, changeLimit } = usePagination(fetcher as never, () => ({ type: 'other' }), {
      immediate: false,
      limit: 12,
    })

    await load()
    expect(fetcher.mock.calls[0][0]).toMatchObject({ page: 1, limit: 12 })

    changeLimit(24)
    await Promise.resolve()
    expect(fetcher.mock.calls[1][0]).toMatchObject({ page: 1, limit: 24 })

    changePage(3)
    await Promise.resolve()
    expect(fetcher.mock.calls[2][0]).toMatchObject({ page: 3, limit: 24 })
  })
})
