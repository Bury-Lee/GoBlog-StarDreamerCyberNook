export interface ApiResult<T = unknown> {
  code: number
  data: T
  message: string
}

export interface ListData<T> {
  list: T[]
  count: number
  /** 总数是否被后端封顶;为 true 时前端显示 ">count" */
  capped?: boolean
}

export interface PageParams {
  page?: number
  limit?: number
  key?: string
  order?: string
  endId?: number
}

export interface BaseModel {
  id: number
  createdAt: string
  updatedAt: string
}

export interface RemoveRequest {
  IDList: number[]
}

export interface ImageRemoveRequest {
  IDlist: number[]
}

export interface CaptchaResult {
  captchaID: string
  captcha: string
}

export interface CaptchaPayload {
  captchaID?: string
  captchaCode?: string
}

export interface TokenPair {
  AccessToken: string
  RefreshToken: string
}

export type UserRole = 1 | 2 | 3 | 4 | 5 | 6

export interface UserModel extends BaseModel {
  username: string
  nickname: string
  avatar: string
  abstract: string
  registerSource: number
  Age: number
  contactInfo: Record<string, string> | null
  email: string
  openID: string
  role: UserRole | number
  likeTags: string[] | null
  lastLoginTime: string
  IP: string
}

export interface UserDetail {
  id: number
  createdAt: string
  updatedAt: string
  username: string
  nickname: string
  avatar: string
  abstract: string
  Age: number
  likeTags: string[] | null
  contactInfo: Record<string, string> | null
  role: number
  updateUsernameDate: string | null
  openFollow: boolean
  openFans: boolean
  homeStyleID: number
}

export interface UserBaseInfo {
  userID: number
  age: number
  nickName: string
  avatar: string
  lastLoginTime: string
  region: string
  existDay: number
  articleCount: number
  fansCount: number
  followCount: number
}

export interface UserListItem {
  userID: number
  nickname: string
  avatar: string
  abstract: string
}

export interface UserInfoUpdatePayload {
  avatar?: string | null
  abstract?: string | null
  likeTags?: string[] | null
  nickName?: string | null
  Age?: number | null
  contactInfo?: Record<string, string> | null
  openFollow?: boolean | null
  openFans?: boolean | null
  homeStyleID?: number | null
}

export interface AdminUserInfoUpdatePayload {
  userID: number
  username?: string | null
  nickname?: string | null
  avatar?: string | null
  abstract?: string | null
  role?: number | null
}

export interface LoginPayload extends CaptchaPayload {
  type: '用户名' | '邮箱'
  val: string
  pwd: string
}

export interface RegisterPayload extends CaptchaPayload {
  emailID: string
  emailCode: string
  password: string
  nickName?: string
}

export type EmailCodeType = '注册' | '重置密码' | '重置邮箱'

export interface SendEmailPayload extends CaptchaPayload {
  type: EmailCodeType
  email: string
}

export interface SendEmailResult {
  emailID: string
  resetEmailID?: string
}

export interface ResetEmailPayload {
  emailID: string
  emailCode: string
  ResetEmailID: string
  ResetEmailCode: string
}

export interface LoginLogItem extends BaseModel {
  userID: number
  ip: string
  addr: string
  userAgent: string
  userNickname?: string
  userAvatar?: string
}

export interface LoginLogQuery extends PageParams {
  userId?: number
  ip?: string
  addr?: string
  startTime?: string
  endTime?: string
}

export type ArticleStatus = 0 | 1 | 2 | 3

export interface ArticleModel extends BaseModel {
  title: string
  abstract: string
  content: string
  categoryID: number | null
  tagList: string[] | null
  cover: string
  userID: number
  lookCount: number
  diggCount: number
  commentCount: number
  collectCount: number
  openComment: boolean
  status: number
}

export interface ArticleAddition {
  articleID: number
  adminComment: string
  aiQuality: string
  aiAbstract: string
  aiModel: string
}

export interface ArticleListResponse extends ArticleModel {
  userTop: boolean
  adminTop: boolean
  categoryTitle: string | null
  userNickName: string
  avatar: string
}

export interface ArticleSearchListResponse extends ArticleModel {
  adminTop: boolean
  categoryTitle: string | null
  userNickname: string
  userAvatar: string
}

export interface ArticleDetailResponse extends ArticleModel {
  categoryTitle: string | null
  username: string
  nickname: string
  userAvatar: string
  articleAddition?: ArticleAddition | null
}

export interface ArticleInteraction {
  digged: boolean
  collected: boolean
}

export interface ArticleCreatePayload {
  title: string
  abstract?: string
  content: string
  categoryID?: number | null
  tagList?: string[]
  cover?: string
  openComment?: boolean
  status?: number
}

export interface ArticleUpdatePayload {
  id: number
  title?: string | null
  abstract?: string | null
  content?: string | null
  categoryID?: number | null
  tagList?: string[] | null
  cover?: string | null
  openComment?: boolean | null
  status?: number | null
}

export interface ArticleListQuery extends PageParams {
  type: 'other' | 'self' | 'admin'
  userID?: number
  categoryID?: number
  status?: number
}

export interface ArticleSearchQuery extends PageParams {
  tag?: string
  type?: number
}

export interface ArticleLookPayload {
  articleID: number
  timeSecond?: number
}

export interface ArticleReviewPayload {
  articleID: number
  status: number
  msg?: string
}

export interface ArticleAIReviewItem {
  articleID: number
  title: string
  aiResult: string
  status: number
  error?: string
}

export interface ArticleAIReviewResult {
  list: ArticleAIReviewItem[]
  count: number
  total: number
}

export interface ArticleHistoryItem {
  id: number
  lookDate: string
  title: string
  cover: string
  nickname: string
  avatar: string
  userID: number
  articleID: number
}

export interface CategoryModel extends BaseModel {
  title: string
  userID: number
}

export interface CategoryListItem extends CategoryModel {
  articleCount: number
  nickname?: string
  avatar?: string
}

export interface CategoryPayload {
  id?: number
  title: string
}

export interface CollectModel extends BaseModel {
  title: string
  abstract: string
  cover: string
  articleList: unknown[] | null
  userID: number
  isDefault: boolean
  isPublic: boolean
}

export interface CollectFolderDetail extends CollectModel {
  articleCount: number
}

export interface CollectPayload {
  articleID: number
  collectID?: number
}

export interface CollectFolderPayload {
  id?: number
  title?: string | null
  abstract?: string | null
  cover?: string | null
  isPublic?: boolean | null
}

export interface CommentModel extends BaseModel {
  content: string
  userID: number
  user: UserModel
  articleID: number
  path: string
  rootParentID: number | null
  diggCount: number
  digged?: boolean
  childCount?: number
}

export interface CommentCreatePayload {
  articleID: number
  content: string
  parentID?: number
}

export interface CommentQuery extends PageParams {
  articleID: number
}

export interface CommentChildQuery extends PageParams {
  root: number
}

export type MessageType = 1 | 2 | 3 | 4 | 5 | 6 | 7

export interface MessageModel extends BaseModel {
  type: number
  revUserID: number
  ActionUserID: number
  actionUserNickname: string
  actionUserAvatar: string
  title: string
  content: string
  articleID: number
  articleTitle: string
  commentID: number
  linkTitle: string
  linkHref: string
  isRead: boolean
}

export interface MessageConf {
  userID: number
  userModel: UserModel
  openCommentMessage: boolean
  openReplyMessage: boolean
  openDiggMessage: boolean
  openCollectMessage: boolean
  openPrivateMessage: boolean
}

export interface MessageConfUpdatePayload {
  openCommentMessage?: boolean
  openReplyMessage?: boolean
  openDiggMessage?: boolean
  openCollectMessage?: boolean
  openPrivateMessage?: boolean
}

export interface OneKeyReadPayload {
  commentMessage: boolean
  diggAndCollectMessage: boolean
  privateMessage: boolean
  systemMessage: boolean
}

export interface ChatMsg {
  textMsg?: { content: string }
  imageMsg?: { src: string }
  markdownMsg?: { content: string }
}

export interface ChatModel extends BaseModel {
  sendUserID: number
  revUserID: number
  msgType: number
  msg: ChatMsg
}

export interface ChatItem extends ChatModel {
  sendUserNickname: string
  sendUserAvatar: string
  revUserNickname: string
  revUserAvatar: string
  isMe: boolean
}

export interface ChatSendPayload {
  revUserID: number
  msg: ChatMsg
}

export interface ChatSession extends BaseModel {
  uniqueId: string
  userId: number
  lastMessageId: number
  lastMessage: ChatModel
  lastMessageTime: string
  unreadCount: number
  isRead: boolean
}

export interface AiMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface AiChatPayload {
  messages: AiMessage[]
  user_input: string
  image_ID?: number
  model?: string
}

export interface AiChatResult {
  success: boolean
  content: string
  error: string
}

export interface FollowUserItem {
  focusUserID: number
  focusUserNickname: string
  focusUserAvatar: string
  focusUserAbstract: string
  createdAt: string
}

export interface FollowModel extends BaseModel {
  userID: number
  focusUserID: number
  friend: boolean
}

export interface Banner extends BaseModel {
  isShow: boolean
  cover: string
  href: string
}

export interface BannerPayload {
  cover?: string
  href?: string
  isShow?: boolean
}

export interface FriendLink extends BaseModel {
  name: string
  url: string
  logo: string
  is_show: boolean
  sort_order: number
  remark: string
}

export interface FriendPromotion extends BaseModel {
  title: string
  friend_name: string
  avatar: string
  category: string
  description: string
  preview_images: string
  contact_info: string[] | null
  is_show: boolean
  sort_order: number
  position: string
  remark: string
}

export interface ImageItem extends BaseModel {
  filename: string
  path: string
  size: number
  hash: string
}

export interface ImageListItem extends ImageItem {
  webPath: string
}

export interface LogModel extends BaseModel {
  logType: number
  title: string
  content: string
  level: number
  userID: number
  ip: string
  addr: string
  isRead: boolean
  loginStatus: boolean
  loginType: number
  serviceName: string
}

export interface LogQuery extends PageParams {
  logType?: number
  level?: number
  ip?: string
  loginStatus?: boolean
  serviceName?: string
  userID?: number
}

export interface SiteInfoConfig {
  title: string
  Logo: string
  Beian: string
  Mode: number
}

export interface SiteProjectConfig {
  title: string
  icon: string
  webPath: string
}

export interface SiteSeoConfig {
  keywords: string
  description: string
}

export interface SiteAboutConfig {
  Version: string
  siteDate: string
  qq: string
  wechat: string
  biliBili: string
  gitHub: string
}

export interface SiteIndexRightItem {
  title: string
  enable: boolean
}

export interface SiteLoginConfig {
  QQLogin: boolean
  usernamePassword: boolean
  emailLogin: boolean
  captcha: boolean
}

export interface SiteConfig {
  siteInfo: SiteInfoConfig
  project: SiteProjectConfig
  seo: SiteSeoConfig
  about: SiteAboutConfig
  indexRight: { list: SiteIndexRightItem[] }
  article: { enableExamination: boolean }
  login: SiteLoginConfig
}

export interface EmailConfig {
  domain: string
  port: number
  sendEmail: string
  authCode: string
  sendNickname: string
  SSL: boolean
  TLS: boolean
}

export interface QQConfig {
  appID: string
  appKey: string
  redirect: string
}

export interface AIConfig {
  enable: boolean
  chat_enable: boolean
  auto_review: boolean
  model: string
  temperature: number
  max_tokens: number
  host: string
  api_type: string
  nickName: string
  avatar: string
  platform: string
}

// ===== 反馈墙 =====
export type FeedbackType = 0 | 1 | 2 | 3 // 0其他 1功能建议 2问题反馈 3内容举报
export type FeedbackStatus = 0 | 1 | 2 | 3 // 0待处理 1已采纳未处理 2正在处理 3已处理

export interface FeedbackItem {
  id: number
  createdAt: string
  updatedAt?: string
  userID?: number
  isAnonymous: boolean
  content: string
  type: FeedbackType
  status: FeedbackStatus
  reply: string
  /** 仅管理员可见 */
  contact?: string
  /** 仅管理员可见 */
  handlerID?: number
}

export interface FeedbackCreatePayload {
  content: string
  contact?: string
  type?: FeedbackType
  isAnonymous?: boolean
}

export interface FeedbackListQuery extends PageParams {
  status?: FeedbackStatus
  type?: FeedbackType
}

export interface FeedbackHandlePayload {
  status: FeedbackStatus
  reply?: string
}

// ===== 动态 / 日记 =====
export type MomentType = 0 | 1 // 0动态 1日记
export type MomentVisibility = 0 | 1 | 2 // 0公开 1仅好友 2私密

export interface MomentModel extends BaseModel {
  userID: number
  user: UserModel
  type: number
  visibility: number
  content: string
  images: string[] | null
  likeCount: number
  commentCount: number
  repostCount: number
  repostFromID: number | null
  /** 转发时后端返回的被转发原动态(作者隐私字段已裁剪) */
  repostFrom?: MomentModel | null
  status: number
  /** 当前用户是否已点赞(前端本地维护) */
  digged?: boolean
}

export interface MomentCommentModel extends BaseModel {
  momentID: number
  userID: number
  user: UserModel
  content: string
  path: string
  rootParentID: number | null
  diggCount: number
  digged?: boolean
}

export interface MomentCreatePayload {
  content?: string
  images?: string[]
  type?: number
  visibility?: number
  status?: number
}

export interface MomentUpdatePayload extends MomentCreatePayload {
  id: number
}

export interface MomentQuery extends PageParams {
  userID?: number
  type?: number
}

export interface MomentRepostPayload {
  content?: string
  visibility?: number
}

export interface MomentCommentCreatePayload {
  momentID: number
  content: string
  parentID?: number
}

export interface MomentCommentQuery extends PageParams {
  momentID: number
}

export interface MomentCommentChildQuery extends PageParams {
  root: number
}

