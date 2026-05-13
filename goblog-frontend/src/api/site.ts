import api from './client';
import type { ApiResponse, Banner } from '../types';

export const siteApi = {
  getSiteInfo: (name: string = 'title') =>
    api.get<ApiResponse<any>>(`/site/${name}`),

  getBannerList: () =>
    api.get<ApiResponse<Banner[]>>('/banner'),

  getQQLoginUrl: () =>
    api.get<ApiResponse<{ url: string }>>('/site/qq_login'),
};
