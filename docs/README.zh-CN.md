# GoBlog（StarDreamerCyberNook）

> 🖥️ **内置前端**：本仓库已在 [`frontend/`](../frontend/) 目录内置 Vue 3 + TypeScript 前端（用户站 + 管理后台），原有外部前端仓库为 `https://github.com/azurekiln2333/azurekiln-goblog-web`。
一个基于 `Gin + GORM + Redis + Elasticsearch` 的博客/社区后端项目，包含用户、文章、评论、消息、关注、聊天、站点配置等模块。

## 🌐 多语言文档

**其他语言：**:[🇨🇳 中文](README.zh-CN.md) • [🇺🇸 English](../README.md) • [🇯🇵 日本語](README.ja.md) • [🇰🇷 한국어](README.ko.md)

## 👤 适合人群:

- 追求独立与自主的技术爱好者：希望拥有完全可控的个人博客或社区平台，不依赖第三方内容平台的规则限制，追求个性化定制和数据主权
- 拥有小型服务器的站长：具备基础服务器资源（如VPS、云主机或家用服务器），希望搭建稳定可靠的图文内容服务
- 社区运营者：需要搭建垂直领域的交流社区、官方论坛或轻量化内容平台，支持用户互动、内容审核和SEO优化
- 注重隐私与数据安全的用户：不愿将核心内容数据托管于商业平台，希望实现本地化部署和自主备份

## 📋 功能速览

- **博客与社区一体化**：文章、评论、收藏、消息、关注、私聊集中在同一套后端
- **搜索能力完整**：基于 `Elasticsearch` 提供文章全文搜索、高亮和多维排序
- **高并发友好**：浏览、点赞、收藏、评论计数先写 Redis，再由定时任务批量同步数据库
- **AI 已接入业务**：支持站点 AI 助手，以及文章、评论、昵称等内容审核
- **多模型 AI 接入**：现已支持多种 AI（支持 OpenAI 接口的模型），加入 AI 文章摘要功能和 AI 文章评级功能，调试模式下输出 AI 回复内容的功能
- **运营能力齐全**：支持站点配置、SEO、轮播图、友情链接、推广位和日志管理
- **内置前端**：`frontend/` 内提供 Vue 3 + TypeScript + Vite + Pinia + Element Plus 前端，覆盖用户站与管理后台，赛博暗色主题

> 📖 **功能文档**：[博客功能文档](功能文档.md)

> ⚠️ **注意**：项目本体支持数据库读写分离，但**不提供数据库之间的数据同步**。仓库不再包含任何同步依赖或工具（原 Canal / PGSync 方案已移除），如需跨库或跨存储同步，请自行在项目之外实现。

## 🏗️ 项目结构

```
go-blog
├─ api/                 # 控制器层
├─ router/              # 路由注册
├─ models/              # 数据模型与 ES Mapping
├─ service/             # 业务服务（含定时任务、ES服务、Redis服务）
├─ middleware/          # 中间件
├─ core/                # 配置/日志/DB/Redis/ES/AI 初始化
├─ conf/                # 配置结构体
├─ flags/               # 命令行参数（迁移、建索引、建用户）
├─ init/                # 本地依赖服务 docker-compose 与基础配置
├─ frontend/            # 内置 Vue 3 + TypeScript 前端（用户站 + 管理后台）
├─ setting.yaml         # 主配置文件
└─ main.go              # 入口
```

## 🖥️ 内置前端（Vue 3 + TypeScript）

完整前端位于 [`frontend/`](../frontend/) 目录，同时覆盖用户站与管理后台，复用同一套 `/api` 接口与 `token` / `refreshToken` 鉴权流程。

### 技术栈

| 分层     | 选型                                                        |
| ------ | --------------------------------------------------------- |
| 框架     | Vue 3（`<script setup>`）+ TypeScript                        |
| 构建     | Vite 6                                                    |
| 状态 / 路由 | Pinia + Vue Router 4（登录与管理员路由守卫）                           |
| UI     | Element Plus + 自定义赛博暗色主题                                   |
| 请求     | Axios（自动携带 `token`、401 自动刷新令牌并重放请求、统一错误提示）                  |
| 内容     | Markdown 编辑器（实时预览），`marked` + `DOMPurify` 过滤             |

### 功能覆盖

- **用户站**：首页（轮播图、按 `site.indexRight` 配置渲染的侧栏组件）、文章列表 / 详情（目录、阅读进度、点赞、收藏、多级评论）、搜索（高亮 + 多维排序）、Markdown / HTML 双模编辑器（图片上传）、个人主页与设置（隐私、通知、邮箱重置）、收藏夹、浏览记录、消息中心（7 类）、私聊、AI 助手、关于页
- **管理后台**：仪表盘、文章审核、文章 / 分类 / 用户管理、轮播图、友情链接、友站推广、图库、日志管理、站点配置查看（邮件 / QQ / AI，敏感字段打码）

### 快速启动

```bash
cd frontend
npm install
npm run dev      # http://localhost:5173，/api 与 /web 自动代理到 127.0.0.1:8080
npm run build    # 产物目录：frontend/dist
```

| 环境变量                  | 默认值                      | 说明           |
| --------------------- | ------------------------ | ------------ |
| `VITE_API_BASE`       | `/api`                   | 前端使用的接口前缀    |
| `VITE_PROXY_TARGET`   | `http://127.0.0.1:8080`  | dev server 代理的后端地址 |

> 💡 请保持 `system.cors_origins` 中包含 `http://localhost:5173` 与 `http://127.0.0.1:5173`（CORS 仅在 `debug` 模式注册）。

### 本地快速跑通（SQLite + Docker Redis + LM Studio）

无需 MySQL / Elasticsearch 也能完整运行：

```yaml
es:
  enabled: false            # 降级为数据库搜索表
dbWrite:
  - db_name: data/goblog.db # 相对路径，基于运行目录解析
    sql_name: sqlite        # 纯 Go 驱动，无需 CGO
objectStorage:
  enable: false             # 图片存本地 ./images，不依赖 RustFS
ai:
  enable: true
  chat_enable: true
  host: http://127.0.0.1:1234/v1   # LM Studio 的 OpenAI 兼容地址
```

```bash
# 1. 启动依赖服务
docker compose -f init/Redis/docker-compose.yml up -d

# 2. 在 setting.yaml 所在目录运行，并确保同目录存在
#    init/ip2region.xdb、images/、static/
./main_windows_amd64.exe -db       # 迁移数据库
./main_windows_amd64.exe -search   # 重建降级搜索表
./main_windows_amd64.exe           # 启动服务
```

### 前端对后端现状的降级处理

| 后端现状                                        | 前端行为                     |
| ------------------------------------------- | ------------------------ |
| `router/enter.go` 未注册关注 / 粉丝路由              | 关注按钮给出明确提示，不阻断页面         |
| `PUT /api/site/:name` 处于注释状态                | 站点配置页只读                  |
| `PUT /api/user/updatePassword` 未注册          | 不渲染修改密码入口                |

> ℹ️ 文章状态与审核语义以**源码**为准（与 `docs/前端API文档.md` 部分描述不一致）：`0=草稿, 1=审核中, 2=已发布, 3=已下线`；审核 `status=2` 通过、`0` 驳回；子评论使用 `root` 参数查询。

## 🔧 环境要求

| 组件            | 版本要求   | 备注                                            |
| ------------- | ------ | --------------------------------------------- |
| Go            | 1.26.6 | 按项目 `go.mod` 版本                               |
| MySQL         | 5.7    | 项目内提供 docker-compose                          |
| Redis         | 7      | 项目内提供 docker-compose，双实例                      |
| Elasticsearch | 7.17.x | 项目使用 `olivere/elastic/v7`                     |
| AI 服务         | 可选     | 默认配置示例：`http://localhost:1234/v1` (目前仅支持本地模型) |

## 🚀 快速启动

### 1️⃣ 启动依赖服务

在项目根目录 `go-blog` 下分别执行：

```bash
# 启动 MySQL
docker compose -f init/MySQL/docker-compose.yml up -d

# 启动 Redis
docker compose -f init/Redis/docker-compose.yml up -d

# 启动 Elasticsearch
docker compose -f init/ES/docker-compose.yml up -d
```

在写完配置文件后执行发行版程序:

Windows

```bash
# 启动项目
.\main_windows_amd64.exe
```

Linux

```bash
# 启动项目
./main_linux_amd64
```

macOS

```bash
# 启动项目
./main_macos_amd64
```

> 💡 **提示**：`init/MySQL/docker-compose.yml` 中使用的是 MySQL `5.7`，建议保持一致以避免兼容性问题。

### 2️⃣ 编译项目

在Windows上编译:
```bash
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

### 3️⃣ 修改主配置 `setting.yaml`

建议优先确认以下关键字段：

- `system.ip`、`system.port`、`system.env`、`system.run_mode`
- `dbWrite` / `dbRead`（写库列表 / 读库列表）
- `redisStatic`、`redisDynamic`
- `es.url`、`es.username`、`es.password`
- `jwt.accessTokenSecret`、`jwt.refreshTokenSecret`
- `email`（用于邮箱验证码登录/注册）
- `ai.enable`（不使用 AI 可设为 `false`）

<details>
<summary>📖 点击查看完整配置示例</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  run_mode: debug # Gin 运行模式：debug 或 release
  cron: true             # 是否启用定时任务,分布式环境下建议仅保留几个开启,其他节点关闭定时任务(反正开了没有锁也执行不了)
  scheduled_cleanup: true    # 是否启用浏览记录清理功能

log:
  app: GoBlog
  dir: log
  log_level: debug
ai:
  enable: false
  model: local
  temperature: 0.7 # 预留字段
  max_tokens: 1024 # 预留字段
  host: http://localhost:1234/v1
  ApiKey: 123456 
  nickName: 昵称
  avatar: "https://example.com/avatar.jpg" # TODO：后续适配头像来源

email:
  domain: smtp.qq.com
  port: 587 # QQ 邮箱常用端口为 587 或 465
  sendEmail: 1287167895@qq.com
  authCode: "xxxxxxx" # 邮箱授权码
  sendNickname: 昵称

upload:
  size: 20 # 上传文件大小限制，单位 MB
  whiteList: # 上传文件白名单
    "jpg": ~
    "jpeg": ~
    "png": ~
    "gif": ~
    "webp": ~
    "bmp": ~
    "tiff": ~
  uploadDir: images # 图片上传目录

jwt:
  accessExpire: 30 # 访问令牌过期时间（分钟），推荐 30 分钟
  refreshExpire: 172 # 刷新令牌过期时间（小时），推荐一周（172 小时）
  accessTokenSecret: xxxxx
  refreshTokenSecret: xxxxxxx
  issuer: "StarDreamer"

redisStatic:
  addr: 127.0.0.1:6379
  password: ""
  db: 1

redisDynamic:
  addr: 127.0.0.1:6380
  password: ""
  db: 2

es:
  enabled: false # true: ES, false: database search fallback
  url: http://127.0.0.1:9200
  username: elastic
  password: es

dbWrite: # 写库列表，至少配置一个
  - user: root
    password: root
    host: 127.0.0.1
    port: 5432
    db_name: db
    sql_name: postgresql
  # - user: root # 可按此格式继续添加多个写库
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db_name: db
  #   sql_name: mysql

dbRead: # 读库列表，可为空；为空时读请求也走写库
  # - user: root
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db_name: db
  #   sql_name: mysql

site:
  siteInfo:
    title: "星梦网络空间" # 站点标题
    logo: "/static/images/logo.png" # 站点 Logo 路径
    beian: "京ICP备XXXXXXXX号" # 备案号
    mode: 1 # 运行模式（1: 博客模式, 2: 社区模式等，需对应代码枚举）

  project:
    title: "StarDreamer" # 项目名称
    icon: "/static/images/favicon.ico" # 项目图标
    webPath: "https://www.example.com" # 项目访问路径

  seo:
    keywords: "技术博客, Go语言, 人工智能, 分享"
    description: "一个专注于技术分享与人工智能探索的个人站点。"

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "热门文章"
        enable: true
      - title: "最新评论"
        enable: true
      - title: "友情链接"
        enable: true
      - title: "标签云"
        enable: true

  article:
    # 说明：Go 字段名是 DisableExamination，但 yaml tag 是 enableExamination
    # 若以 enableExamination 读取：true 表示“启用审核”（即不禁用）
    # 业务语义：true = 需要审核，false = 无需审核
    enableExamination: true

  login:
    QQLogin: true # TODO：尚未实现
    usernamePassword: true
    emailLogin: true # TODO：可考虑固定开启（基础登录方式）
    captcha: false # 是否启用验证码


objectStorage:
  enable: true                     # 启用对象存储，true表示使用RustFS
  accessKey: "admin"               # 与docker-compose.yml中的RUSTFS_ACCESS_KEY一致
  secretKey: "123456"              # 与docker-compose.yml中的RUSTFS_SECRET_KEY一致
  bucket: "test"              # 存储桶名称，你需要提前在RustFS中创建
  host: "http://127.0.0.1:9000"    # RustFS服务地址，使用S3 API端口
  uri: ""                          # (TODO:自定义URI路径)
  region: "us-east-1"              # 区域，默认值
  prefix: "uploads/"               # 文件前缀，如 uploads/images/
  size: 10                   # 文件大小限制(字节)，默认10MB (10*1024*1024)
```

</details>

### 4️⃣ 初始化数据库结构

```bash
go run main.go -db
```

### 5️⃣ 初始化 ES 索引

```bash
go run main.go -es
```

### 6️⃣ 启动服务

```bash
go run main.go -f setting.yaml
```

🌐 启动后默认监听：`http://127.0.0.1:8080`

## 📝 命令行参数

| 参数                  | 说明                        |
| ------------------- | ------------------------- |
| `-f`                | 配置文件路径（默认 `setting.yaml`） |
| `-db`               | 执行 GORM 自动迁移              |
| `-es`               | 创建/重建 ES 索引               |
| `-search`           | 重建数据库搜索表（ES 降级搜索）          |
| `-v`                | 查看版本                      |
| `-t user -s create` | 命令行创建用户                   |

**示例：**

```bash
go run main.go -t user -s create
```

## ⚙️ 关键运行说明

- **初始化顺序**：配置 → 日志 → IP库 → DB → Redis → ES → AI → 定时任务 → 路由
- **ES 为强依赖**：ES 初始化失败会导致服务启动中断
- **接口规范**：当前接口统一使用 `/api` 前缀，例如 `/api/user/login`
- **静态资源**：静态资源目录映射为 `/web`，对应本地 `static/`
- **定时任务**：当前定时任务 `SyncArticle/SyncComment` 为 10 分钟级批量同步，用于刷新 Redis 中的计数增量

- **搜索模式**：`es.enabled=true` 时使用 Elasticsearch；关闭时降级到 `article_search_models` 表（支持标签过滤与多种排序），可用 `-search` 重建
- **清理任务**：`SyncCleanHistory` 每 10 分钟清理超过 30 天的浏览记录

## 🔄 数据库同步（不提供依赖）

项目本体支持数据库读写分离，但**不提供数据库之间的数据同步**，仓库也不再包含任何同步依赖或工具（原 Canal / PGSync 方案已移除）。

- 主库与从库之间的数据一致性，请在项目之外自行保证（例如使用 MySQL 自身的主从复制）
- 如需把数据同步到其他存储（例如 Elasticsearch），请在项目之外自行实现

## 🌐 NGINX 水平扩展（可选）

经过近期架构调整，项目服务层已具备无状态特性，业务数据统一落在 `MySQL / Redis / Elasticsearch`。  
在此基础上，可通过 `NGINX` 反向代理与负载均衡能力实现横向扩展。

### ✅ 适用场景

- 单实例 CPU 或连接数已接近瓶颈，需要平滑提升并发处理能力
- 已有多台主机或多个容器实例，希望对外统一一个访问入口
- 需要在不中断服务的情况下，逐步扩容或替换后端节点

### 🔧 实施要点

1. 启动多个 GoBlog 实例（建议统一版本与配置，仅端口不同）
2. 在 NGINX `upstream` 中注册多个后端节点
3. 通过 `proxy_pass` 将入口流量转发至 `upstream`
4. 按需启用健康检查、超时重试、连接保持等策略

### 📈 效果预期

- 多实例分担请求压力，提升整体吞吐量
- 降低单点负载，改善高峰期响应稳定性
- 支持按节点滚动发布，降低变更风险

## ❓ 常见问题

| 问题               | 解决方案                                                     |
| ---------------- | -------------------------------------------------------- |
| **MySQL 连接失败**   | 检查 `setting.yaml` 中的账号、密码、端口是否与实际 MySQL 一致               |
| **ES 启动失败**      | 确认 `http://127.0.0.1:9200` 可访问，账号密码匹配                    |
| **Redis 警告淘汰策略** | 项目会检查动态 Redis 的 `maxmemory-policy`                       |
| **接口 404**       | 注意当前接口带 `/api` 前缀，应访问 `/api/user/login` 这类路径             |
| **前端代理报 `ECONNREFUSED 127.0.0.1:8080`** | 后端未启动，或在没有 `setting.yaml` 的目录下启动。请在配置文件所在目录运行可执行文件 |

## 💡 仅开发者提示

- 代码中含部分 TODO（如 ES 升级 v8、路由前缀切换、定时任务频率）
- 若用于生产，建议补充：限流、监控、配置分环境管理、错误恢复与幂等处理

## 📞 联系我们

如果在部署或使用过程中遇到任何问题，欢迎通过以下方式反馈：

- **🐛 GitHub Issues**: 在 [项目仓库](https://github.com/Bury-Lee/go-blog) 提交 Issue
- **📧 邮箱联系**: <18151161@qq.com>
- **👥 交流群**: [星梦的交流群](https://qun.qq.com/universal-share/share?ac=1\&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g\&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D\&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ\&svctype=4\&tempid=h5_group_info)

作者非常乐意解决有价值的技术问题，也欢迎提交 PR 参与项目贡献！
