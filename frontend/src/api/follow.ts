import { http } from './request'
import type { FollowModel, FollowUserItem, ListData, PageParams } from './types'

export function followUser(focusUserID: number): Promise<unknown> {
  return http.post<unknown>('/user/follow', { focusUserID }, { silent: true })
}

export function unfollowUser(focusUserID: number): Promise<unknown> {
  return http.post<unknown>('/user/follow/unfollow', { focusUserID }, { silent: true })
}

export function fetchFollowList(
  params: PageParams & { userID?: number },
): Promise<ListData<FollowUserItem>> {
  return http.get<ListData<FollowUserItem>>('/user/follow/list', params as Record<string, unknown>, {
    silent: true,
  })
}

export function fetchFollowerList(
  params: PageParams & { userID?: number },
): Promise<ListData<FollowModel>> {
  return http.get<ListData<FollowModel>>('/user/follower/list', params as Record<string, unknown>, {
    silent: true,
  })
}
