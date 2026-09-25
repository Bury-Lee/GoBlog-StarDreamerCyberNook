# GoBlog 前端 API 文档

> 基于 `Gin + GORM + Redis + Elasticsearch` 的后端。所有接口前缀为 `/api`。
> 静态资源通过 `/web` 暴露。

---

## 1. 通用约定

### 1.1 基础 URL

```
开发环境: http://localhost:{端口}/api
```

### 1.2 统一响应格式

所有接口返回 JSON，结构如下：

```json
{
  "code": 200,
  "data": {},
  "message": "成功"
}
```

### 1.3 状态码说明

| code | 含义 |
|------|------|
| 200 | 操作成功 |
| 201 | 一般操作失败 |
| 400 | 请求参数错误 / 数据格式错误 |
| 401 | 未登录 / 认证失败 / Token 过期 |
| 403 | 权限不足（非管理员） |
| 404 | 资源不存在 |
| 422 | 参数校验失败 |
| 500 | 服务器内部错误 / 数据库错误 |
| 503 | 服务暂时不可用 |

### 1.4 认证方式

登录成功后返回 `AccessToken` 和 `RefreshToken`：

- **AccessToken**：请求头携带 `token: <AccessToken>`
- **RefreshToken**：用于刷新 AccessToken，请求头携带 `refreshToken: <RefreshToken>` 或 `Authorization: Bearer <RefreshToken>`

AccessToken 过期后调用刷新接口获取新 Token。

### 1.5 通用分页参数

适用于所有列表类接口的 Query 参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码（1-20） |
| `limit` | int | 10 | 每页条数（1-40） |
| `key` | string | - | 模糊搜索关键词 |
| `order` | string | - | 排序字段（如 `created_at desc`）。各接口有白名单限制，非法字段会被忽略并回退默认排序 |
| `endId` | uint | - | 游标分页：上一页末尾 ID |

**列表响应格式**：

```json
{
  "code": 200,
  "data": {
    "list": [...],
    "count": 100
  },
  "message": "成功"
}
```

---

## 2. 心跳检测

### GET /api/heartbeat

无需认证。

**响应示例**：
```json
{
  "code": 200,
  "data": {},
  "message": "存活"
}
```

---

## 3. 验证码

### GET /api/captcha

获取图形验证码。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 业务标识：`注册` / `重置密码` / `重置邮箱` / `用户名` / `邮箱` |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "captchaID": "xxx",
    "captcha": "base64图片字符串"
  },
  "message": "成功"
}
```

- `captchaID`：后续请求需携带，用于校验验证码
- `captcha`：Base64 编码的图片，前端可直接渲染
- 有效期：3 分钟

---

## 4. 用户模块

### 4.1 发送邮箱验证码

**POST /api/user/send_email**

> 需要先通过图形验证码（CaptchaMiddleware），根据站点配置可能必须校验 captcha。

**请求体**：
```json
{
  "type": "注册",
  "email": "user@example.com"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | `注册` / `重置密码` / `重置邮箱` |
| `email` | string | 是 | 目标邮箱 |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "emailID": "abc123...",
    "resetEmailID": "def456..."
  },
  "message": "邮箱验证码有效期为: 2 分钟"
}
```

- `resetEmailID` 仅 `type = 重置邮箱` 时返回（原邮箱的验证码ID）
- 验证码为 **8 位字母数字**，有效期 2 分钟
- 同一 IP **1 分钟只能发送 1 次**（超限返回 422「请求过于频繁」）
- 验证码连续输错 5 次会立即作废，需重新获取

---

### 4.2 邮箱注册

**POST /api/user/email**

> 需要先通过邮箱验证码校验（EmailVerifyMiddleware）。

**请求体**：
```json
{
  "emailID": "abc123...",
  "emailCode": "123456",
  "password": "mypassword",
  "nickName": "用户昵称"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `emailID` | string | 是 | 发邮件时返回的 emailID |
| `emailCode` | string | 是 | 邮箱收到的 6 位验证码 |
| `password` | string | 是 | 登录密码 |
| `nickName` | string | 否 | 展示昵称，若启用 AI 会审核 |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "AccessToken": "eyJ...",
    "RefreshToken": "eyJ..."
  },
  "message": "成功"
}
```

注册成功即自动登录，返回 Token。

---

### 4.3 登录

**POST /api/user/login**

> 根据站点配置可能需要先通过图形验证码（CaptchaMiddleware）。

**请求体**：
```json
{
  "type": "用户名",
  "val": "myusername",
  "pwd": "mypassword"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | `用户名` 或 `邮箱` |
| `val` | string | 是 | 用户名或邮箱地址 |
| `pwd` | string | 是 | 密码 |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "AccessToken": "eyJ...",
    "RefreshToken": "eyJ..."
  },
  "message": "成功"
}
```

---

### 4.4 退出登录

**DELETE /api/user/logout**

> 需要认证。Token 将加入黑名单。

**请求头**：
```
token: <AccessToken>
refreshToken: <RefreshToken>
```

**响应示例**：
```json
{
  "code": 200,
  "data": {},
  "message": "退出登录成功"
}
```

---

### 4.5 刷新 AccessToken

**POST /api/user/token**

无需认证。

**请求头**（二选一）：
```
refreshToken: <RefreshToken>
```
或
```
Authorization: Bearer <RefreshToken>
```

**响应示例**：
```json
{
  "code": 200,
  "data": "新的AccessToken字符串",
  "message": "成功"
}
```

---

### 4.6 获取当前用户详情

**GET /api/user/detail**

> 需要认证。

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-06-01T00:00:00Z",
    "username": "系统用户名",
    "nickname": "展示昵称",
    "avatar": "/web/avatar.png",
    "abstract": "个人简介",
    "Age": 25,
    "likeTags": ["Go", "Rust"],
    "contactInfo": { "github": "xxx" },
    "role": 1,
    "updateUsernameDate": null,
    "openFollow": true,
    "openFans": false,
    "homeStyleID": 1
  },
  "message": "成功"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `role` | int | 1=普通用户, 2=管理员 |
| `openFollow` | bool | 是否公开关注列表 |
| `openFans` | bool | 是否公开粉丝列表 |

> **变更**：`openCollect` 字段已从该响应移除。收藏夹公开改为**文件夹级** `CollectModel.isPublic`，不再使用用户级开关。

---

### 4.7 查看指定用户基本信息

**GET /api/user/info/:id**

无需认证。

`id`：用户 ID。

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "userID": 1,
    "age": 25,
    "nickName": "展示昵称",
    "avatar": "/web/avatar.png",
    "lastLoginTime": "2024-06-01T12:00:00Z",
    "region": "广东·广州",
    "existDay": 365,
    "articleCount": 42,
    "fansCount": 128,
    "followCount": 64
  },
  "message": "成功"
}
```

---

### 4.8 用户列表

**GET /api/user/list**

无需认证。支持通用分页参数，模糊搜索昵称和简介。

**响应列表项**（仅返回公开字段，不包含邮箱、OpenID、联系方式等隐私信息）：

```json
{
  "userID": 2,
  "nickname": "用户一",
  "avatar": "/web/avatar.png",
  "abstract": "个人简介"
}
```

---

### 4.9 更新个人信息

**PUT /api/user/update**

> 需要认证。支持部分更新（传 null 或不传表示不修改）。

**请求体**（所有字段可选）：
```json
{
  "avatar": "/web/new-avatar.png",
  "abstract": "新简介",
  "likeTags": ["Go", "TypeScript"],
  "nickName": "新昵称",
  "Age": 26,
  "contactInfo": { "github": "new" },
  "openFollow": false,
  "openFans": true,
  "homeStyleID": 2
}
```

> 若启用 AI 审核，昵称/简介/标签/联系方式会被 AI 审核，不通过则拒绝更新。
> **变更**：`openCollect` 已废弃并不再接收，收藏夹公开请使用 `PUT /api/article/collect/folder` 的 `isPublic`。

---

### 4.10 管理员更新用户信息

**PUT /api/user/admin/update**

> 需要管理员权限。

**请求体**：
```json
{
  "userID": 1,
  "username": "新系统用户名",
  "nickname": "新昵称",
  "avatar": "/web/avatar.png",
  "abstract": "新简介",
  "role": 1
}
```

| 字段 | 说明 |
|------|------|
| `role` | 1=普通用户, 2=管理员 |

所有字段除 `userID` 外均可选。

---

### 4.11 获取登录日志

**GET /api/user/loginlog**

> 需要认证。

> 服务端按登录角色决定查询范围：管理员可查全部（含用户信息），普通用户只能查自己的记录（`type` 参数已移除）。

**Query 参数**（均可选）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ip` | string | 否 | 按 IP 筛选 |
| `addr` | string | 否 | 按地址筛选 |
| `startTime` | string | 否 | 开始时间 `2006-01-02 15:04:05` |
| `endTime` | string | 否 | 结束时间 `2006-01-02 15:04:05` |

支持通用分页参数。

---

### 4.12 重置邮箱

**PUT /api/user/resetEmail**

> 需要认证 + 原邮箱验证码 + 新邮箱验证码（双重验证）。

流程：
1. 调用 `POST /api/user/send_email` 时 type 选 `重置邮箱`，会向新旧邮箱各发验证码
2. 拿到两个 `emailID` 和验证码后调用此接口

**请求体**：
```json
{
  "emailID": "新邮箱的emailID",
  "emailCode": "新邮箱验证码",
  "resetEmailID": "原邮箱的emailID",
  "resetEmailCode": "原邮箱验证码"
}
```

---

### 4.13 修改密码

**PUT /api/user/updatePassword**

> 需要认证 + 邮箱验证（仅限已绑定邮箱的用户）。

> ⚠️ 该路由目前**未注册**在 `user_router.go` 中，需确认是否可用。

**请求体**：
```json
{
  "pwd": "新密码"
}
```

---

## 5. 站点配置

### 5.1 查询站点配置

**GET /api/site/:name**

| name 值 | 说明 | 权限 |
|---------|------|------|
| `site` | 站点基本信息 | 公开 |
| `email` | 邮件配置 | 管理员 |
| `qq` | QQ 登录配置 | 管理员 |
| `ai` | AI 配置 | 管理员 |

**ai 配置字段**：`enable`（启用 AI）、`chat_enable`（开放对话接口）、`auto_review`（定时任务自动 AI 审核待审核文章）、`model`、`temperature`、`max_tokens`、`host`、`api_type`、`nickName`、`avatar`、`platform`；`ApiKey` 返回时打码为 `******`。

**site 响应示例**：
```json
{
  "code": 200,
  "data": {
    "project": { "title": "博客名", "icon": "favicon.ico", "webPath": "/path/to/index.html" },
    "seo": { "keywords": "...", "description": "..." },
    "login": { "emailLogin": true, "usernamePassword": true, "captcha": true },
    "article": { "enableExamination": false },
    "siteInfo": { "mode": 1 }
  },
  "message": "成功"
}
```

---

### 5.2 查询 QQ 登录配置

**GET /api/site/qq_login**

公开接口。

---

### 5.3 更新站点配置

**PUT /api/site/:name**

> 需要管理员权限。

name 取值：`site` / `email` / `qq` / `objectStorage` / `ai`

**请求体**：对应配置对象的 JSON（与 GET 返回的结构一致）。

> 注：敏感字段（密码/密钥等）发送 `"******"` 表示不修改。

---

## 6. 图片模块

### 6.1 上传图片

**POST /api/image**

> 需要认证 + 上传频率限制。上传成功返回图片 ID，用于后续引用。

**请求**：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | 是 | 图片文件 |

**响应示例**：
```json
{
  "code": 200,
  "data": 42,
  "message": "上传成功"
}
```

- `data` 即为图片 ID
- 相同文件（MD5 相同）不会重复存储，直接复用已有 ID
- 前端引用图片：`/api/image?id=<图片ID>`
- 仅允许 `jpg/jpeg/png/gif/webp/bmp/tiff`（**不支持 svg**）；服务端会做魔数校验，文件内容与后缀不符返回 422「文件内容与图片格式不符」
- 同一 IP 每分钟最多上传 20 次

---

### 6.2 获取图片

**GET /api/image**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `id` | 图片 ID |

直接返回图片二进制流（`Content-Type` 为对应图片类型）。

---

### 6.3 图片列表

**GET /api/images**

> 需要管理员权限。支持通用分页，文件名模糊搜索。

---

### 6.4 删除图片

**DELETE /api/image**

> 需要管理员权限。

**请求体**：
```json
{
  "IDlist": [1, 2, 3]
}
```

---

## 7. 轮播图

### 7.1 轮播图列表

**GET /api/banner**

公开接口。支持通用分页。结果会缓存到 Redis。

---

### 7.2 创建轮播图

**POST /api/banner**

> 需要管理员权限。

**请求体**：
```json
{
  "cover": "/api/image?id=1",
  "href": "https://example.com/article/1",
  "isShow": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `cover` | string | 封面图 URL |
| `href` | string | 点击跳转链接 |
| `isShow` | bool | 是否显示 |

---

### 7.3 更新轮播图

**PUT /api/banner/:id**

> 需要管理员权限。`id` 为轮播图 ID。请求体同创建。

---

### 7.4 删除轮播图

**DELETE /api/banner**

> 需要管理员权限。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

---

## 8. 友情链接 & 友站推广

### 8.1 友情链接

#### GET /api/friendLink
公开接口。支持通用分页。

#### POST /api/friendLink
> 需要管理员权限。

**请求体**：
```json
{
  "name": "友站名",
  "url": "https://example.com",
  "logo": "/api/image?id=1",
  "is_show": true,
  "sort_order": 0,
  "remark": "备注"
}
```

#### PUT /api/friendLink/:id
> 管理员权限。请求体同创建。

#### DELETE /api/friendLink
> 管理员权限。请求体：`{ "IDList": [1, 2] }`

---

### 8.2 友站推广位

#### GET /api/friendPromotion
公开接口。

#### POST /api/friendPromotion
> 管理员权限。

**请求体**：
```json
{
  "title": "推广标题",
  "friend_name": "友站名",
  "avatar": "/api/image?id=1",
  "category": "技术",
  "description": "描述",
  "preview_images": "/api/image?id=2",
  "contact_info": ["https://github.com/xxx"],
  "is_show": true,
  "sort_order": 0,
  "position": "sidebar",
  "remark": "备注"
}
```

#### PUT /api/friendPromotion/:id
> 管理员权限。请求体同创建。

#### DELETE /api/friendPromotion
> 管理员权限。请求体：`{ "IDList": [1, 2] }`

---

## 9. 文章模块

### 9.1 创建文章

**POST /api/article**

> 需要认证。

**请求体**：
```json
{
  "title": "文章标题",
  "abstract": "文章摘要（可选，最大200字符）",
  "content": "文章正文（HTML）",
  "categoryID": 1,
  "tagList": ["Go", "后端"],
  "cover": "/api/image?id=1",
  "openComment": true,
  "status": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 标题 |
| `content` | string | 是 | 正文（会经过 XSS 过滤） |
| `abstract` | string | 否 | 摘要，最大 200 字符 |
| `categoryID` | uint | 否 | 分类 ID（必须是自己的分类） |
| `tagList` | []string | 否 | 标签列表 |
| `cover` | string | 否 | 封面图 URL |
| `openComment` | bool | 否 | 是否开启评论 |
| `status` | int | 否 | 0=草稿, 1=待审核, 3=已发布。普通用户仅可设 0 或 1 |

> 若启用 AI + 审核，文章会自动审核并将状态改为已发布/草稿。AI 还会自动生成摘要和评级。

---

### 9.2 全量更新文章

**PUT /api/article**

> 需要认证（仅限自己的文章）。

**请求体**：
```json
{
  "id": 1,
  "title": "新标题",
  "abstract": "新摘要",
  "content": "新正文",
  "categoryID": 1,
  "tagList": ["Go"],
  "cover": "/api/image?id=2",
  "openComment": false
}
```

所有字段必填（除 `categoryID` 可传 0 或 null 清空分类）。

---

### 9.3 增量更新文章

**PUT /api/article/inc**

> 需要认证（仅限自己的文章）。支持部分更新。

**请求体**（所有字段可选，传 null 不修改）：
```json
{
  "id": 1,
  "title": "新标题",
  "abstract": "新摘要",
  "content": "新正文",
  "categoryID": 1,
  "tagList": ["Go"],
  "cover": "/api/image?id=2",
  "openComment": false
}
```

- 只更新传入的非 null 字段
- 若内容变更 + AI 启用，会重新审核并刷新 AI 摘要/评级

---

### 9.4 文章列表

**GET /api/article**

支持通用分页。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | `other`（查别人）/ `self`（查自己）/ `admin` |
| `userID` | uint | 否 | 用户 ID（type=other 时指定） |
| `categoryID` | uint | 否 | 分类筛选 |
| `status` | int | 否 | 文章状态。type=self/admin 时可用 |
| `order` | string | 否 | `look_count desc` / `digg_count desc` / `comment_count desc` / `collect_count desc`（及 asc 变体） |

**type 说明**：
- `other`：只返回已发布文章
- `self`：需要登录，返回自己的文章（可按状态筛选）
- `admin`：需要管理员，返回全部文章

**响应列表项**：
```json
{
  "id": 1,
  "title": "标题",
  "abstract": "摘要",
  "cover": "...",
  "userTop": false,
  "adminTop": false,
  "categoryTitle": "分类名",
  "userNickName": "作者昵称",
  "avatar": "作者头像",
  "lookCount": 100,
  "diggCount": 10,
  "commentCount": 5,
  "collectCount": 3,
  "tagList": ["Go"],
  "createdAt": "2024-01-01T00:00:00Z"
}
```

---

### 9.5 文章详情

**GET /api/article/:id**

> Redis 缓存。未登录只能看已发布文章，登录后可看自己的所有状态文章，管理员可看全部。

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "title": "标题",
    "abstract": "摘要",
    "content": "正文HTML",
    "cover": "...",
    "tagList": ["Go"],
    "lookCount": 100,
    "diggCount": 10,
    "commentCount": 5,
    "collectCount": 3,
    "categoryTitle": "分类名",
    "username": "系统用户名",
    "nickname": "昵称",
    "userAvatar": "头像",
    "articleAddition": {
      "id": 1,
      "articleID": 1,
      "adminComment": "管理员附加评论",
      "aiQuality": "AI评级",
      "aiAbstract": "AI自动摘要",
      "aiModel": "local"
    },
    "openComment": true,
    "status": 3,
    "createdAt": "2024-01-01T00:00:00Z"
  },
  "message": "成功"
}
```

> **变更**：AI 点评（`aiQuality`/`aiAbstract`）与管理员附加评论迁移到文章扩展附录表 `article_additions`（通过 `article_id` 外键一对一同文章关联），详情接口以 **`articleAddition`** 对象返回，并记录生成点评的 AI 模型名 `aiModel`。文章列表/搜索不再返回该对象。

---

### 9.6 搜索文章

**GET /api/article/search**

搜索有两条链路：`es.enabled=true` 时走 Elasticsearch；关闭时自动降级为数据库搜索表（`article_search_models`）。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | 否 | 搜索关键词；为空时返回全部已发布文章 |
| `tag` | string | 否 | 按标签筛选（两种链路均支持） |
| `type` | int8 | 否 | 排序类型：0=最新, 1=猜你喜欢, 2=最多回复, 3=最多点赞, 4=最多收藏；非法值返回 422 |

支持通用分页。

- 两条链路都只返回**已发布**文章（草稿/已下线不会出现在搜索结果中）
- `title` 和 `abstract` 会包含高亮标签 `<em>关键词</em>`
- 降级链路按文章创建时间排序（`type=1` 退化为最新），编辑文章不会使其跳到搜索结果最前
- 批量导入数据后可用命令行 `-search` 重建降级搜索表

---

### 9.7 文章置顶

**POST /api/article/top/:id**

> 需要认证。普通用户最多置顶 5 篇。

`id` 为文章 ID。

---

### 9.8 取消置顶（用户）

**DELETE /api/article/top**

> 需要认证。

**请求体**：
```json
{
  "articleID": 1
}
```

---

### 9.9 管理员取消置顶

**DELETE /api/article/admingTop**

> 需要管理员权限。

**请求体**：
```json
{
  "userID": 1,
  "articleID": 1
}
```

---

### 9.10 审核文章列表

**GET /api/article/review**

> 需要管理员权限。查询状态为草稿（待审核）的文章。

**Query 参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `userID` | uint | 可选，按用户筛选 |

支持通用分页。

---

### 9.11 审核文章

**POST /api/article/review/:id**

> 需要管理员权限。`id` 为文章 ID。

**请求体**：
```json
{
  "articleID": 1,
  "status": 3,
  "msg": "审核通过"
}
```

- `status`：3=通过（已发布），1=不通过（退回草稿）
- 审核结果会以系统通知发送给文章作者

---

### 9.11.1 AI 审核文章（单个 / 批量）

**POST /api/article/ai/review**

> 需要管理员权限，且 `ai.enable=true`，否则返回「AI功能未启用」。

**请求体**：
```json
{
  "articleID": 1,
  "IDList": [2, 3],
  "limit": 10
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `articleID` | uint | 可选，单个文章 ID |
| `IDList` | []uint | 可选，批量文章 ID；与 `articleID` 都为空时审核全部待审核文章 |
| `limit` | int | 批量上限，默认 10，最大 20 |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "list": [
      { "articleID": 1, "title": "标题", "aiResult": "通过", "status": 2 }
    ],
    "count": 1,
    "total": 1
  },
  "message": "成功"
}
```

- 判定规则：AI 返回「通过」→ 已发布，「拒绝」→ 退回草稿，其它 → 保持审核中等待人工处理
- 开启 `ai.auto_review` 后，定时任务会每 10 分钟自动执行同样的批量审核（全部待审核文章），AI 服务不可用时自动跳过

---

### 9.12 点赞/取消点赞文章

**POST /api/article/digg/:id**

> 需要认证。`id` 为文章 ID。点赞和取消点赞同一接口（toggle）。

- 已点赞 → 取消点赞
- 未点赞 → 点赞

---

### 9.13 创建浏览记录

**POST /api/article/look**

> 需要认证。同一天同一用户同一文章只记一次（Redis + 数据库双重去重），重复请求不会重复增加浏览量。

**请求体**：
```json
{
  "articleID": 1,
  "timeSecond": 120
}
```

---

### 9.14 浏览记录列表

**GET /api/article/history**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `userID` | 为 0 查自己的，否则查指定用户（需对方开启公开） |

支持通用分页。

---

### 9.15 删除浏览记录

**DELETE /api/article/history**

> 需要认证。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

---

### 9.16 删除文章（用户）

**DELETE /api/article**

> 需要认证。只能删除自己的文章。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

---

### 9.17 删除文章（管理员）

**DELETE /api/article/admin**

> 需要管理员权限。会清理 Redis 缓存并发送系统通知。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

---

### 9.18 文章分类

#### 创建/更新分类
**POST /api/article/category**

> 需要认证。`id` 为 0 时创建，否则更新。

**请求体**：
```json
{
  "id": 0,
  "title": "技术文章"
}
```

#### 分类列表
**GET /api/article/category**

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | `self` / `other` / `admin` |
| `userID` | uint | 否 | type=other 时必填 |

支持通用分页和标题模糊搜索。

#### 删除分类
**DELETE /api/article/category**

> 需要认证。管理员可删任意分类，普通用户只能删自己的。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

---

### 9.19 文章收藏

#### 收藏/取消收藏
**POST /api/article/collect**

> 需要认证。toggle 操作。

**请求体**：
```json
{
  "articleID": 1,
  "collectID": 0
}
```

- `collectID` 为 0 时使用默认收藏夹（自动创建）

#### 创建收藏夹
**POST /api/article/collect/folder**

> 需要认证。新建收藏夹默认 `isPublic: true`（创建后可通过更新接口改为私有）。

**请求体**：
```json
{
  "title": "收藏夹名称",
  "abstract": "描述",
  "cover": "/api/image?id=1"
}
```

#### 更新收藏夹
**PUT /api/article/collect/folder**

> 需要认证（只能更新自己的）。所有字段可选，仅更新传入的字段。

**请求体**（所有字段可选）：
```json
{
  "id": 1,
  "title": "新名称",
  "abstract": "新描述",
  "cover": "/api/image?id=2",
  "isPublic": false
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 收藏夹 ID（必填） |
| `title` | string | 收藏夹名称 |
| `abstract` | string | 收藏夹简介 |
| `cover` | string | 收藏夹封面 |
| `isPublic` | bool | 是否公开该收藏夹（**收藏夹级**隐私开关） |

> 传空对象 `{ "id": 1 }` 会返回 `没有需要更新的字段`。

#### 删除收藏夹
**DELETE /api/article/collect/folder**

> 需要认证。默认收藏夹不可删除。管理员可删任意收藏夹。

**请求体**：
```json
{
  "IDList": [1, 2]
}
```

#### 收藏夹列表
**GET /api/article/collect/folder**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `id` | 用户 ID（必填），查该用户的收藏夹 |

> 非本人只返回 `isPublic = true` 的收藏夹。

支持通用分页和标题/摘要模糊搜索。

#### 收藏夹详情
**GET /api/article/collect/folder/:id**

> 新增接口。获取单个收藏夹详情 + 文章数量。

**路径参数**：

| 参数 | 说明 |
|------|------|
| `id` | 收藏夹 ID |

**响应 `data`**：
```json
{
  "id": 1,
  "title": "收藏夹名称",
  "abstract": "描述",
  "cover": "/api/image?id=1",
  "userID": 3,
  "isDefault": false,
  "isPublic": true,
  "articleCount": 12
}
```

> 非本人（含未登录）访问时，仅当该收藏夹 `isPublic = true` 才可访问，否则统一返回 **404**。

#### 收藏夹内文章列表
**GET /api/article/collect/list**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `id` | 收藏夹 ID（必填） |
| `key` | 按文章标题/摘要模糊搜索（可选） |

支持通用分页，按收藏时间倒序返回，只返回**已发布**文章。

> 非本人（含未登录）访问他人收藏夹时，仅当该收藏夹 `isPublic = true` 才可访问，否则统一返回 **404**（不暴露收藏夹是否存在）。

---

## 10. 评论模块

### 10.1 发表评论

**POST /api/comment**

> 需要认证。只能对已发布文章评论。

**请求体**：
```json
{
  "articleID": 1,
  "content": "评论内容",
  "parentID": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `articleID` | uint | 是 | 文章 ID |
| `content` | string | 是 | 评论内容 |
| `parentID` | uint | 否 | 父评论 ID，0 表示一级评论 |

> 若启用 AI 审核，评论内容会先过审。发表评论会自动向文章作者/父评论作者发送通知。

---

### 10.2 一级评论列表

**GET /api/comment**

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `articleID` | uint | 是 | 文章 ID |

支持通用分页。默认按点赞数降序排列。预加载用户信息。

---

### 10.3 子评论列表

**GET /api/commentChild**

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `articleID` | uint | 是 | 文章 ID |
| `parentID` | uint | 是 | 父评论 ID |

支持通用分页。默认按点赞数降序排列。

---

### 10.4 删除评论

**DELETE /api/comment/:id**

> 需要认证。只能删自己的，管理员可删任意评论。

- 一级评论：级联删除其下所有子评论
- 子评论：仅删除自身

---

### 10.5 评论点赞

**POST /api/comment/digg/:id**

> 需要认证。toggle 操作。`id` 为评论 ID。

---

## 11. 消息模块

### 11.1 消息配置

**GET /api/msg/conf**

> 需要认证。返回当前用户的消息通知开关。

**POST /api/msg/conf/update**

> 需要认证。

**请求体**：
```json
{
  "openCommentMessage": true,
  "openReplyMessage": true,
  "openDiggMessage": false,
  "openCollectMessage": true,
  "openPrivateMessage": true
}
```

---

### 11.2 检查未读消息

**GET /api/msg/check**

> 需要认证。

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "1": 3,
    "2": 0,
    "3": 1
  },
  "message": "成功"
}
```

- key 为消息类型编号，value 为未读数量

---

### 11.3 消息列表

**GET /api/msg**

> 需要认证。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | int | 是 | 消息类型（见下表） |

消息类型枚举：

| type 值 | 含义 |
|---------|------|
| 1 | @ 提醒 |
| 2 | 评论通知 |
| 3 | 回复通知 |
| 4 | 点赞通知 |
| 5 | 收藏通知 |
| 6 | 私聊通知 |
| 7 | 系统通知 |

> 查询后自动将列表中的未读消息标记为已读。

支持通用分页。

---

### 11.4 删除消息

**DELETE /api/msg**

> 需要认证。

**请求体**：
```json
{
  "messageID": [1, 2, 3]
}
```

一次最多删除 100 条。

---

### 11.5 一键已读

**POST /api/msg/clear**

> 需要认证。

**请求体**：
```json
{
  "commentMessage": true,
  "diggAndCollectMessage": false,
  "privateMessage": false,
  "systemMessage": true
}
```

- `commentMessage`：含 @提醒 + 评论 + 回复
- `diggAndCollectMessage`：含点赞 + 收藏
- `privateMessage`：私聊消息
- `systemMessage`：系统通知

---

## 12. 私聊模块

### 12.1 发送消息

**POST /api/chat/send**

> 需要认证。仅限好友之间发送。

**请求体**：
```json
{
  "revUserID": 2,
  "msg": {
    "textMsg": { "content": "你好" },
    "markdownMsg": null,
    "imageMsg": null
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `revUserID` | uint | 接收者用户 ID |
| `msg.textMsg` | object | 文本消息 `{"content": "文本"}` |
| `msg.markdownMsg` | object | Markdown 消息 `{"content": "md"}` |
| `msg.imageMsg` | object | 图片消息 `{"url": "...", "alt": "..."}` |

三种消息类型可组合使用。

---

### 12.2 聊天记录

**GET /api/chat/get**

> 需要认证。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `userID` | uint | 是 | 对方的用户 ID |

支持通用分页。查询后会自动将该会话标记为已读。

**响应列表项**：
```json
{
  "id": 1,
  "sendUserID": 1,
  "revUserID": 2,
  "msg": { "textMsg": { "content": "你好" } },
  "msgType": 1,
  "sendUserNickname": "我的昵称",
  "sendUserAvatar": "我的头像",
  "revUserNickname": "对方昵称",
  "revUserAvatar": "对方头像",
  "isMe": true,
  "createdAt": "2024-01-01T00:00:00Z"
}
```

---

### 12.3 会话列表

**GET /api/chat/session**

> 需要认证。返回当前用户的所有会话。

支持通用分页。按最后消息时间降序排列。

---

## 13. AI 助手

### POST /api/chat

> **需要认证**。需同时满足 `ai.enable=true` 且 `ai.chat_enable=true`，否则该路由不会注册（返回 404）。

**请求体**：
```json
{
  "messages": [
    { "role": "user", "content": "之前我问过的问题" },
    { "role": "assistant", "content": "之前的回答" }
  ],
  "user_input": "新的问题"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `messages` | array | 历史对话（只含 user/assistant 角色，不允许 system） |
| `user_input` | string | 当前用户输入 |

**响应示例**：
```json
{
  "code": 200,
  "data": {
    "success": true,
    "content": "AI 回复内容",
    "error": ""
  },
  "message": "成功"
}
```

---

## 14. 关注 & 粉丝

> ⚠️ 当前 `router/enter.go` **未注册** `UserFollowRouter`，以下接口暂时全部返回 404；如需启用，请在 `InitRouter()` 中补上 `UserFollowRouter(nr)`。

### 14.1 关注用户

**POST /api/user/follow**

> 需要认证。

**请求体**：
```json
{
  "focusUserID": 2
}
```

---

### 14.2 取消关注

**POST /api/user/follow/unfollow**

> 需要认证。请求体同关注。

---

### 14.3 关注列表

**GET /api/user/follow/list**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `userID` | 为 0 查自己的，否则查指定用户（需对方开启 `openFollow`） |

支持通用分页。

---

### 14.4 粉丝列表

**GET /api/user/follower/list**

**Query 参数**：

| 参数 | 说明 |
|------|------|
| `userID` | 为 0 查自己的，否则查指定用户（需对方开启 `openFans`） |

支持通用分页。

---

## 15. 日志（管理员）

### 15.1 日志列表

**GET /api/logs**

> 需要管理员权限。

**Query 参数**（均可选）：

| 参数 | 类型 | 说明 |
|------|------|------|
| `logType` | int | 日志类型 |
| `level` | string | 日志级别 |
| `ip` | string | 客户端 IP |
| `loginStatus` | bool | 登录状态 |
| `serviceName` | string | 服务名 |
| `userID` | uint | 用户 ID |

支持通用分页。预加载关联用户信息。

---

### 15.2 日志详情

**GET /api/logs/:id**

> 需要管理员权限。首次读取标记为已读。

---

### 15.3 删除日志

**DELETE /api/logs**

> 需要管理员权限。

**请求体**：
```json
{
  "IDList": [1, 2, 3]
}
```

---

## 16. 测试接口（仅 Debug 模式）

仅在 `RunMode = debug` 时可用。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/t/test` | 测试接口 |
| POST | `/api/t/print` | 打印测试 |

---

## 附录 A：文章状态枚举

| 值 | 含义 |
|----|------|
| 0 | 草稿 |
| 1 | 审核中 |
| 2 | 已发布 |
| 3 | 已下线 |

---

## 附录 B：中间件 & 权限说明

| 中间件 | 说明 |
|--------|------|
| `AuthMiddleware` | JWT 认证 + 用户状态校验（被封禁用户会被拒绝） |
| `AdminMiddleware` | 管理员权限校验（同样校验封禁状态） |
| `CaptchaMiddleware` | 图形验证码校验（验证码一次性使用） |
| `EmailVerifyMiddleware` | 邮箱验证码校验（失败 5 次作废，成功后原子消费） |
| `ResetEmailVerifyMiddleware` | 重置邮箱时的原邮箱验证 |
| `EmailSendLimitMiddleware` | 邮箱发送频率限制（同 IP 1 分钟 1 次） |
| `ImgPostLimitMiddleware` | 图片上传频率限制（同 IP 每分钟 20 次） |
| `ActLimitMiddleware` | IP 级请求频率限制（同 IP 每分钟 64 次） |
| `SecurityHeaders` | 安全响应头（nosniff / X-Frame-Options / Referrer-Policy） |
| `LogMiddleware` | 操作日志记录 |
| `CORS` | 跨域（仅 Debug 模式，来源白名单 `system.cors_origins`） |

---

## 附录 C：认证 Token 使用

1. 登录/注册成功后获取 `AccessToken` + `RefreshToken`
2. 每次请求在 Header 中携带：`token: <AccessToken>`
3. AccessToken 过期后，使用 RefreshToken 换取新 Token：
   - Header：`refreshToken: <RefreshToken>` 或 `Authorization: Bearer <RefreshToken>`
   - 接口：`POST /api/user/token`
4. 退出登录会写入 Token 黑名单（以 RefreshToken 为键）；AccessToken 默认 30 分钟过期，前端退出时应同时清除本地保存的两个 Token

---

## 附录 D：近期接口变更（2026-09）

| 变更 | 说明 |
|------|------|
| 登录日志 | 移除 `type` 参数，按角色自动判定查询范围 |
| 用户列表 | 仅返回 `userID/nickname/avatar/abstract`，不再返回邮箱等隐私字段 |
| 排序参数 | 所有列表接口的 `order` 改为白名单校验，非法字段被忽略 |
| 搜索 | 支持 ES / 数据库降级双链路，`tag`、`type` 在降级链路同样生效 |
| 收藏夹文章列表 | 修复 422，支持分页与关键词搜索，仅返回已发布文章 |
| 图片上传 | 增加魔数校验，移除 svg 支持 |
| AI 对话 | 需登录，且受 `ai.chat_enable` 独立开关控制 |
| 邮箱验证码 | 8 位字母数字，发送频率 1 次/分钟，失败 5 次作废 |
| 安全加固 | JWT 密钥强度校验、登录/注册限流、安全响应头、CORS 白名单 |
| 文章状态 | 枚举修正：0=草稿、1=审核中、2=已发布、3=已下线 |
| 收藏夹更新 | 修复更新不生效问题；新增 `isPublic` 收藏夹级开关 |
| 收藏夹详情 | 新增 `GET /api/article/collect/folder/:id` |
| 收藏夹权限 | 非本人访问仅取决于该收藏夹 `isPublic`；用户级 `openCollect` 已移除 |
| AI 点评 | `aiQuality`/`aiAbstract` 迁到扩展附录表，详情以 `articleAddition` 对象返回（含 `aiModel`）；新增配置 `ai.auto_comment` 定时补全 |

---

*文档更新时间：2026-09-25*
