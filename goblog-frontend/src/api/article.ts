import api from './client';
import type { ApiResponse, Article, ArticleListRequest, ArticleFormData, Category } from '../types';

export const articleApi = {
  getArticleList: (params: ArticleListRequest) =>
    api.get<ApiResponse<{ list: Article[]; total: number }>>('/article', { params }),

  getArticleDetail: (id: number) =>
    api.get<ApiResponse<Article>>(`/article/${id}`),

  searchArticles: (params: { keyword?: string; tag?: string; type?: number; page?: number; pageSize?: number }) =>
    api.get<ApiResponse<{ list: Article[]; total: number }>>('/article/search', { params }),

  createArticle: (data: ArticleFormData) => api.post<ApiResponse>('/article', data),

  updateArticle: (data: ArticleFormData & { id: number }) => api.put<ApiResponse>('/article', data),

  deleteArticle: (id: number) =>
    api.delete<ApiResponse>(`/article/${id}`),

  diggArticle: (id: number) =>
    api.post<ApiResponse>(`/article/digg/${id}`),

  collectArticle: (id: number) =>
    api.post<ApiResponse>('/article/collect', { articleID: id }),

  lookArticle: (id: number) =>
    api.post<ApiResponse>(`/article/look/${id}`),

  getCategories: (userID?: number, type: string = 'other') =>
    api.get<ApiResponse<Category[]>>('/article/category', { params: { userID, type } }),

  createCategory: (name: string) =>
    api.post<ApiResponse>('/article/category', { name }),

  getArticleHistory: () =>
    api.get<ApiResponse<Article[]>>('/article/history'),
};
