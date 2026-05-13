import api from './client';
import type { ApiResponse, LoginResponse, User } from '../types';

export const userApi = {
  login: (data: { type: string; val: string; pwd: string }) =>
    api.post<ApiResponse<LoginResponse>>('/user/login', data),

  register: (data: { email: string; code: string; password: string; nickname?: string }) =>
    api.post<ApiResponse>('/user/register', data),

  sendEmail: (email: string, type: string = '注册') =>
    api.post<ApiResponse>('/user/send_email', { email, type }),

  getUserDetail: () =>
    api.get<ApiResponse<User>>('/user/detail'),

  updateUserInfo: (data: Partial<User>) =>
    api.put<ApiResponse>('/user/update', data),

  getUserInfo: (id: number) =>
    api.get<ApiResponse<User>>(`/user/info/${id}`),

  refreshToken: (refreshToken: string) =>
    api.post<ApiResponse<LoginResponse>>('/user/token', { refreshToken }),

  logout: () =>
    api.delete<ApiResponse>('/user/logout'),
};
