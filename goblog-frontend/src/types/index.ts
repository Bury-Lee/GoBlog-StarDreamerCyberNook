export interface ApiResponse<T = any> {
  code: number;
  data: T;
  message: string;
}

export interface User {
  id: number;
  user_name: string;
  nick_name: string;
  avatar?: string;
  email?: string;
  role: number;
  created_at?: string;
}

export interface Article {
  id: number;
  title: string;
  abstract?: string;
  content: string;
  cover?: string;
  user_id: number;
  category_id?: number;
  tag_list?: string[];
  look_count: number;
  digg_count: number;
  collect_count: number;
  comment_count: number;
  open_comment: boolean;
  status: ArticleStatus;
  created_at: string;
  updated_at: string;
  user?: User;
  category?: Category;
}

export enum ArticleStatus {
  Draft = 1,
  Published = 2,
  Deleted = 0
}

export interface Category {
  id: number;
  name: string;
  created_at: string;
}

export interface Comment {
  id: number;
  content: string;
  user_id: number;
  article_id: number;
  parent_id?: number;
  created_at: string;
  user?: User;
  children?: Comment[];
}

export interface Banner {
  id: number;
  title: string;
  image: string;
  link?: string;
  created_at: string;
}

export interface LoginResponse {
  AccessToken: string;
  RefreshToken: string;
}

export interface PageInfo {
  page: number;
  pageSize: number;
}

export interface ArticleListRequest {
  page?: number;
  pageSize?: number;
  type?: 'other' | 'self' | 'admin';
  userID?: number;
  categoryID?: number;
  status?: ArticleStatus;
}

export interface ArticleFormData {
  title: string;
  content: string;
  abstract?: string;
  categoryID?: number;
  tagList?: string[];
  cover?: string;
  openComment?: boolean;
  status?: ArticleStatus;
}
