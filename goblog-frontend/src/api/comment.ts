import api from './client';
import type { ApiResponse, Comment } from '../types';

export const commentApi = {
  getCommentList: (articleId: number, page = 1, pageSize = 20) =>
    api.get<ApiResponse<{ list: Comment[]; total: number }>>('/comment', {
      params: { articleID: articleId, page, pageSize },
    }),

  getChildComments: (parentId: number) =>
    api.get<ApiResponse<Comment[]>>('/commentChild', {
      params: { parentID: parentId },
    }),

  createComment: (data: {
    content: string;
    article_id: number;
    parent_id?: number;
  }) => api.post<ApiResponse>('/comment', data),

  deleteComment: (id: number) =>
    api.delete<ApiResponse>(`/comment/${id}`),

  diggComment: (id: number) =>
    api.post<ApiResponse>(`/comment/digg/${id}`),
};
