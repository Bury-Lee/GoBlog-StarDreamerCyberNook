import { http } from './request'
import type {
  ArticleAIReviewResult,
  ArticleCreatePayload,
  ArticleDetailResponse,
  ArticleHistoryItem,
  ArticleListQuery,
  ArticleListResponse,
  ArticleLookPayload,
  ArticleModel,
  ArticleReviewPayload,
  ArticleSearchListResponse,
  ArticleSearchQuery,
  ArticleUpdatePayload,
  CategoryListItem,
  CategoryModel,
  CategoryPayload,
  CollectFolderPayload,
  CollectModel,
  CollectPayload,
  ListData,
  PageParams,
  RemoveRequest,
} from './types'

export function createArticle(payload: ArticleCreatePayload): Promise<unknown> {
  return http.post<unknown>('/article', payload)
}

export function updateArticle(payload: ArticleUpdatePayload): Promise<unknown> {
  return http.put<unknown>('/article/inc', payload)
}

export function replaceArticle(payload: ArticleUpdatePayload): Promise<unknown> {
  return http.put<unknown>('/article', payload)
}

export function fetchArticleList(params: ArticleListQuery): Promise<ListData<ArticleListResponse>> {
  return http.get<ListData<ArticleListResponse>>('/article', params as unknown as Record<string, unknown>)
}

export function fetchArticleDetail(id: number): Promise<ArticleDetailResponse> {
  return http.get<ArticleDetailResponse>(`/article/${id}`)
}

export function searchArticles(params: ArticleSearchQuery): Promise<ListData<ArticleSearchListResponse>> {
  return http.get<ListData<ArticleSearchListResponse>>('/article/search', params as Record<string, unknown>)
}

export function removeArticles(idList: number[]): Promise<unknown> {
  const payload: RemoveRequest = { IDList: idList }
  return http.delete<unknown>('/article', payload)
}

export function adminRemoveArticles(idList: number[]): Promise<unknown> {
  const payload: RemoveRequest = { IDList: idList }
  return http.delete<unknown>('/article/admin', payload)
}

export function topArticle(articleID: number): Promise<unknown> {
  return http.post<unknown>('/article/top/0', { articleID })
}

export function cancelTopArticle(articleID: number): Promise<unknown> {
  return http.delete<unknown>('/article/top', { articleID })
}

export function adminCancelTopArticle(articleID: number, userID?: number): Promise<unknown> {
  return http.delete<unknown>('/article/admingTop', { articleID, userID })
}

export function fetchReviewArticles(params?: PageParams & { userID?: number }): Promise<ListData<ArticleModel>> {
  return http.get<ListData<ArticleModel>>('/article/review', params as Record<string, unknown>)
}

export function reviewArticle(payload: ArticleReviewPayload): Promise<unknown> {
  return http.post<unknown>(`/article/review/${payload.articleID}`, payload)
}

export function aiReviewArticles(payload?: {
  articleID?: number
  IDList?: number[]
  limit?: number
}): Promise<ArticleAIReviewResult> {
  return http.post<ArticleAIReviewResult>('/article/ai/review', payload ?? {})
}

export function diggArticle(id: number): Promise<unknown> {
  return http.post<unknown>(`/article/digg/${id}`)
}

export function reportArticleLook(payload: ArticleLookPayload): Promise<unknown> {
  return http.post<unknown>('/article/look', payload, { silent: true })
}

export function fetchArticleHistory(
  params?: PageParams & { userID?: number },
): Promise<ListData<ArticleHistoryItem>> {
  return http.get<ListData<ArticleHistoryItem>>('/article/history', params as Record<string, unknown>)
}

export function removeArticleHistory(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/article/history', { IDList: idList } satisfies RemoveRequest)
}

export function collectArticle(payload: CollectPayload): Promise<unknown> {
  return http.post<unknown>('/article/collect', payload)
}

export function fetchCollectFolders(
  params: PageParams & { id: number },
): Promise<ListData<CollectModel>> {
  return http.get<ListData<CollectModel>>('/article/collect/folder', params as unknown as Record<string, unknown>)
}

export function createCollectFolder(payload: CollectFolderPayload): Promise<unknown> {
  return http.post<unknown>('/article/collect/folder', payload)
}

export function updateCollectFolder(payload: CollectFolderPayload): Promise<unknown> {
  return http.put<unknown>('/article/collect/folder', payload)
}

export function removeCollectFolders(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/article/collect/folder', { IDList: idList } satisfies RemoveRequest)
}

export function fetchCollectArticles(
  params: PageParams & { id: number },
): Promise<ListData<ArticleModel>> {
  return http.get<ListData<ArticleModel>>('/article/collect/list', params as unknown as Record<string, unknown>)
}

export function saveCategory(payload: CategoryPayload): Promise<unknown> {
  return http.post<unknown>('/article/category', payload)
}

export function fetchCategories(
  params: PageParams & { type: 'self' | 'other' | 'admin'; userID?: number },
): Promise<ListData<CategoryListItem>> {
  return http.get<ListData<CategoryListItem>>('/article/category', params as unknown as Record<string, unknown>)
}

export function removeCategories(idList: number[]): Promise<unknown> {
  return http.delete<unknown>('/article/category', { IDList: idList } satisfies RemoveRequest)
}

export type { CategoryModel }
