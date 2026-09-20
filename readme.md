# GoBlog (StarDreamerCyberNook)

> 🖥️ **Built-in frontend**: this repository now ships a Vue 3 + TypeScript frontend under [`frontend/`](frontend/) (public site + admin console). The legacy external frontend repository is `https://github.com/azurekiln2333/azurekiln-goblog-web`.
A blog/community backend project based on `Gin + GORM + Redis + Elasticsearch`, featuring modules for users, articles, comments, messages, follows, chat, and site configuration.

## 🌐 Multilingual Documentation

**Languages:** [🇨🇳 中文](docs/README.zh-CN.md) • [🇺🇸 English](README.md) • [🇯🇵 日本語](docs/README.ja.md) • [🇰🇷 한국어](docs/README.ko.md)

## 👤 Who Is It For?

- **Tech enthusiasts who want full control**: Build a fully self-hosted personal blog or community without relying on third-party platform rules; prioritize customization and data ownership
- **Site owners with a small server**: Have basic server resources such as a VPS, cloud instance, or home server and want a stable text-and-image content service
- **Community operators**: Need a vertical community, official forum, or lightweight content platform with user interaction, content moderation, and SEO optimization
- **Privacy and data security minded users**: Prefer local deployment and self-managed backups instead of hosting core content data on commercial platforms


## 📋 Features Overview

- **Blog & Community Integration**: Articles, comments, collections, messages, follows, and private chat all in one backend
- **Complete Search Capabilities**: Full-text article search, highlighting, and multi-dimensional sorting based on `Elasticsearch`; when ES is disabled, search falls back to a database search table with tag filtering and multiple sort modes
- **Security Hardened**: ORDER BY whitelist, JWT secret strength validation, login/registration rate limiting and brute-force protection, upload magic-number check, XSS sanitization, security response headers, and CORS whitelist
- **High Concurrency Friendly**: Views, likes, collections, and comment counts are written to Redis first, then synchronized to the database in batches by scheduled tasks
- **AI Integration**: Site AI assistant and content moderation for articles, comments, and nicknames
- **Multi-model AI Support**: Now supports multiple AIs (OpenAI interface compatible), added AI article summary and AI article rating features, and supports outputting AI responses in debug mode
- **Complete Operations Features**: Site configuration, SEO, banners, friend links, promotion slots, and log management
- **Built-in Frontend**: Vue 3 + TypeScript + Vite + Pinia + Element Plus frontend covering both the public site and the admin console, with a cyber dark theme

> 📖 **Feature Documentation**: [Blog Feature Documentation](docs/功能文档.md)

> ⚠️ **Note**: The project supports database read-write separation but **does not provide data synchronization between databases**. The repository does not ship any synchronization dependency or tooling (the former Canal / PGSync solutions have been removed); if you need cross-database or cross-store synchronization, implement it yourself outside this project.

## 🏗️ Project Structure

```
go-blog
├─ api/                 # Controller layer
├─ router/              # Route registration
├─ models/              # Data models and ES Mapping
├─ service/             # Business services (including scheduled tasks, ES services, Redis services)
├─ middleware/          # Middleware
├─ core/                # Configuration/log/DB/Redis/ES/AI initialization
├─ conf/                # Configuration structures
├─ flags/               # Command line arguments (migration, index creation, user creation)
├─ init/                # Local dependency services docker-compose and basic configuration
├─ build/               # Cross-platform build scripts (build.sh / build.bat)
├─ frontend/            # Built-in Vue 3 + TypeScript frontend (public site + admin console)
├─ setting.yaml         # Main configuration file
└─ main.go              # Entry point
```

## 🖥️ Built-in Frontend (Vue 3 + TypeScript)

A complete frontend lives in [`frontend/`](frontend/), covering the public site and the admin console. It talks to the same `/api` endpoints and reuses the `token` / `refreshToken` authentication flow.

### Tech Stack

| Layer | Choice |
|-------|--------|
| Framework | Vue 3 (`<script setup>`) + TypeScript |
| Build | Vite 6 |
| State / Routing | Pinia + Vue Router 4 (auth & admin route guards) |
| UI | Element Plus with a custom cyber dark theme |
| HTTP | Axios (auto `token` header, refresh-token retry with request replay, unified error toast) |
| Content | Markdown editor with live preview, `marked` + `DOMPurify` sanitising |

### Feature Coverage

- **Public site**: home (banner, configurable sidebar widgets driven by `site.indexRight`), article list / detail (TOC, reading progress, likes, collections, nested comments), search with highlight and multi-dimension sorting, Markdown / HTML editor with image upload, profile & settings (privacy, notifications, e-mail reset), collections, browsing history, message center (7 types), private chat, AI assistant, about page
- **Admin console**: dashboard, article review, article / category / user management, banners, friend links, friend promotions, image library, logs, site configuration viewer (e-mail / QQ / AI, secrets masked)

### Quick Start

```bash
cd frontend
npm install
npm run dev      # http://localhost:5173, proxies /api and /web to 127.0.0.1:8080
npm run build    # production output: frontend/dist
```

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE` | `/api` | API prefix used by the frontend |
| `VITE_PROXY_TARGET` | `http://127.0.0.1:8080` | Backend address for the dev-server proxy |

> 💡 Keep `http://localhost:5173` / `http://127.0.0.1:5173` in `system.cors_origins` — CORS is only registered in `debug` mode.

### Local Quick Run (SQLite + Docker Redis + LM Studio)

The backend can run fully offline without MySQL / Elasticsearch:

```yaml
es:
  enabled: false            # fall back to the database search table
dbWrite:
  - db_name: data/goblog.db # relative path, resolved from the working directory
    sql_name: sqlite        # pure-Go driver, no CGO required
objectStorage:
  enable: false             # store uploads in ./images instead of S3/RustFS
ai:
  enable: true
  chat_enable: true
  host: http://127.0.0.1:1234/v1   # LM Studio OpenAI-compatible endpoint
```

```bash
# 1. dependency services
docker compose -f init/Redis/docker-compose.yml up -d

# 2. run from the directory that contains setting.yaml, and make sure
#    init/ip2region.xdb, images/ and static/ exist next to it
./main_windows_amd64.exe -db       # migrate the database
./main_windows_amd64.exe -search   # rebuild the fallback search table
./main_windows_amd64.exe           # start the service
```

### Backend Gaps Handled by the UI

| Backend status | Frontend behaviour |
|----------------|--------------------|
| Follow / follower routes are not registered in `router/enter.go` | Follow buttons degrade gracefully with a clear notice |
| `PUT /api/site/:name` is commented out | Site configuration page is read-only |
| `PUT /api/user/updatePassword` is not registered | No password-change entry is rendered |

> ℹ️ Article status and review semantics follow the **source code** rather than `docs/前端API文档.md`: `0=draft, 1=pending, 2=published, 3=offline`; review `status=2` approves and `0` rejects; child comments are queried with the `root` parameter.

## 🔧 Environment Requirements

| Component | Version Requirement | Notes |
|----------|-------------------|-------|
| Go | 1.26.6 | According to project `go.mod` version |
| MySQL | 5.7 | Docker-compose provided in project |
| Redis | 7 | Docker-compose provided in project, dual instances |
| Elasticsearch | 7.17.x | Project uses `olivere/elastic/v7` |
| AI Service | Optional | Default configuration example: `http://localhost:1234/v1` (currently only supports local models) |

## 🚀 Quick Start

### 1️⃣ Start Dependency Services

Execute in the project root directory `go-blog`:

```bash
# Start MySQL
docker compose -f init/MySQL/docker-compose.yml up -d

# Start Redis
docker compose -f init/Redis/docker-compose.yml up -d

# Start Elasticsearch
docker compose -f init/ES/docker-compose.yml up -d
```

After writing the configuration file, execute the release program:

Windows
```bash
# Start project
.\main_windows_amd64.exe
```
Linux
```bash
# Start project
./main_linux_amd64
```
macOS
```bash
# Start project
./main_macos_amd64
```



> 💡 **Tip**: `init/MySQL/docker-compose.yml` uses MySQL `5.7`; keep the same version to avoid compatibility issues.

### 2️⃣ Compile Project

Use the provided build scripts (output goes to `dist/`):

```bash
# Windows
build\build.bat

# Linux / macOS
./build/build.sh
```

Or compile manually:
```bash
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

> ⚠️ **Run the binary from the directory that contains `setting.yaml`** (relative paths such as `init/ip2region.xdb`, `log`, `images`, and the SQLite file are resolved against the working directory).

### 3️⃣ Modify Main Configuration `setting.yaml`

Key fields to confirm first:

- `system.ip`, `system.port`, `system.env`, `system.run_mode`
- `dbWrite` / `dbRead` (write database list / read database list)
- `redisStatic`, `redisDynamic`
- `es.enabled` and `es.url`, `es.username`, `es.password`
- `jwt.accessTokenSecret`, `jwt.refreshTokenSecret`
- `email` (for email verification code login/registration)
- `ai.enable` / `ai.chat_enable` (set to `false` if not using AI / not exposing AI chat)
- `system.cors_origins` (CORS whitelist, only registered in debug mode)

<details>
<summary>📖 Click to view complete configuration example</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  run_mode: debug # debug or release
  cron: true             # Whether to enable scheduled tasks
  scheduled_cleanup: true    # Whether to enable scheduled cleanup of visit records
  cors_origins:          # CORS whitelist (only registered in debug mode)
    - http://localhost:5173
    - http://127.0.0.1:5173

log:
  app: GoBlog
  dir: log
  log_level: debug

ai:
  enable: true
  chat_enable: true # Whether to expose the AI chat endpoint (/api/chat)
  model: local
  host: http://localhost:1234/v1
  api_type: openai
  ApiKey: 123456
  nickName: Nickname
  avatar: "https://example.com/avatar.jpg" # TODO: Adapt avatar source later

email:
  domain: smtp.qq.com
  port: 587 # QQ email common ports are 587 or 465
  sendEmail: xxxxxx@qq.com
  authCode: "xxxxxxx" # Email authorization code
  sendNickname: Nickname

upload:
  size: 20 # Upload file size limit, unit MB
  whiteList: # Upload file whitelist (keys without the leading dot; svg is not allowed)
    "jpg": ~
    "jpeg": ~
    "png": ~
    "gif": ~
    "webp": ~
    "bmp": ~
    "tiff": ~
  uploadDir: images # Image upload directory

jwt:
  accessExpire: 30 # Access token expiration time (minutes), recommended 30 minutes
  refreshExpire: 172 # Refresh token expiration time (hours), recommended one week (172 hours)
  accessTokenSecret: xxxxx # Random string >= 32 chars; release mode refuses to start with placeholders/weak keys
  refreshTokenSecret: xxxxxxx
  issuer: "StarDreamer"

redisStatic: # TTL data: tokens, verification codes, rate limits
  addr: 127.0.0.1:6379
  password: ""
  db: 1

redisDynamic: # Hot data cache: article details, banners
  addr: 127.0.0.1:6380
  password: ""
  db: 2

es:
  enabled: false # true: use Elasticsearch; false: database search fallback
  url: http://127.0.0.1:9200
  username: elastic
  password: es

dbWrite: # Write databases (at least one)
  - user: root
    password: root
    host: 127.0.0.1
    port: 3306
    db_name: db
    sql_name: mysql
  # - user: root # Can continue adding multiple write databases in this format
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db_name: db
  #   sql_name: mysql

dbRead: # Read databases (optional; reads fall back to write databases when empty)
  # - user: root
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db_name: db
  #   sql_name: mysql

site:
  siteInfo:
    title: "StarDreamer Cyberspace" # Site title
    logo: "/static/images/logo.png" # Site Logo path
    beian: "京ICP备XXXXXXXX号" # Filing number
    mode: 1 # Run mode (1: blog mode, 2: community mode, etc., needs to correspond to code enum)

  project:
    title: "StarDreamer" # Project name
    icon: "/static/images/favicon.ico" # Project icon
    webPath: "https://www.example.com" # Project access path

  seo:
    keywords: "tech blog, Go language, AI, sharing"
    description: "A personal site focused on technology sharing and AI exploration."

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "Popular Articles"
        enable: true
      - title: "Latest Comments"
        enable: true
      - title: "Friend Links"
        enable: true
      - title: "Tag Cloud"
        enable: true

  article:
    # Note: Go field name is DisableExamination, but yaml tag is enableExamination
    # If read as enableExamination: true means "enable review" (i.e., not disabled)
    # Business semantics: true = needs review, false = no review needed
    enableExamination: true

  login:
    QQLogin: true # TODO: Not yet implemented
    usernamePassword: true
    emailLogin: true # TODO: Can consider always on (basic login method)
    captcha: false # Whether to enable captcha
```

</details>

### 4️⃣ Initialize Database Structure

```bash
go run main.go -db
```

### 5️⃣ Initialize Search (Optional)

If Elasticsearch is enabled, create the ES index; if ES is disabled, rebuild the database search table instead:

```bash
# ES enabled
go run main.go -es

# ES disabled (database search fallback)
go run main.go -search
```

### 6️⃣ Start Service

```bash
go run main.go -f setting.yaml
```

🌐 Default listening after startup: `http://127.0.0.1:8080`

## 📝 Command Line Parameters

| Parameter | Description |
|-----------|-------------|
| `-f` | Configuration file path (default `setting.yaml`) |
| `-db` | Execute GORM auto migration |
| `-es` | Create/rebuild ES index |
| `-search` | Rebuild the database search table (ES fallback search) |
| `-v` | View version |
| `-t user -s create` | Create user via command line |

**Example:**
```bash
go run main.go -t user -s create
```

## ⚙️ Key Runtime Instructions

- **Initialization order**: Configuration → Log → IP library → DB → Redis → ES → AI → Scheduled tasks → Routes
- **Search modes**: When `es.enabled=true`, search uses Elasticsearch; when disabled, it falls back to the `article_search_models` table (keyword + tag filter, sort by latest/comments/likes/collects). Rebuild the fallback table with `-search`
- **ES behavior**: When ES is enabled, a failed connection aborts startup; keep it disabled if you do not run ES
- **Interface specification**: Current interfaces use `/api` prefix, e.g., `/api/user/login`
- **Static resources**: Static resource directory mapped as `/web`, corresponding to local `static/`
- **Scheduled tasks**: `SyncArticle` / `SyncComment` batch-sync Redis count increments to the database every 10 minutes; `SyncCleanHistory` removes browsing history older than 30 days

## 🔄 Database Synchronization (Not Provided)

The project supports database read-write separation, but **does not provide data synchronization between databases**, and the repository no longer ships any synchronization dependency or tooling (the former Canal / PGSync solutions have been removed).

- Data consistency between the write database and the read databases must be guaranteed by your own operations outside the project (for example, native MySQL replication)
- If you need to synchronize data to another store (such as Elasticsearch), implement it yourself outside this repository

## 🌐 NGINX Horizontal Scaling (Optional)

After recent architecture adjustments, the service layer is now effectively stateless, and business data is centralized in `MySQL / Redis / Elasticsearch`.  
Based on this design, you can use `NGINX` reverse proxy and load balancing to scale out horizontally.

### ✅ Suitable Scenarios

- A single instance is approaching CPU or connection limits, and you need smoother concurrency scaling
- Multiple hosts or container instances are available, and you want one unified public entry point
- You need gradual capacity expansion or node replacement without service interruption

### 🔧 Implementation Notes

1. Start multiple GoBlog instances (same version and config recommended, different ports)
2. Register backend nodes in an NGINX `upstream`
3. Forward inbound traffic to the `upstream` via `proxy_pass`
4. Enable health checks, timeout retries, and keepalive policies as needed

### 📈 Expected Benefits

- Distributes request pressure across instances and increases overall throughput
- Reduces single-node hotspots and improves stability during peak traffic
- Supports rolling updates by node to lower release risk

## ❓ Common Issues

| Issue | Solution |
|-------|----------|
| **MySQL connection failed** | Check if account, password and port in `setting.yaml` match your MySQL instance |
| **ES startup failed** | Confirm `http://127.0.0.1:9200` is accessible and account password matches |
| **Redis warning eviction policy** | Project will check dynamic Redis `maxmemory-policy` |
| **Interface 404** | Note current interfaces have `/api` prefix, should access paths like `/api/user/login` |
| **Frontend proxy `ECONNREFUSED 127.0.0.1:8080`** | The backend is not running, or it was started from a directory without `setting.yaml`. Start the binary from the folder that contains the configuration file |

## 💡 Developer Only Tips

- Code contains some TODOs (e.g., ES upgrade to v8, route prefix switching, scheduled task frequency)
- For production use, recommend adding: authentication, rate limiting, monitoring, environment-specific configuration management, error recovery and idempotent processing

## 📞 Contact Us

If you encounter any problems during deployment or use, please provide feedback through the following methods:

- **🐛 GitHub Issues**: Submit Issue in [project repository](https://github.com/Bury-Lee/go-blog)
- **📧 Email Contact**: [18151161@qq.com](mailto:18151161@qq.com)
- **👥 Discussion Group**: [StarDreamer Discussion Group](https://qun.qq.com/universal-share/share?ac=1&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ&svctype=4&tempid=h5_group_info)

The author is very willing to solve valuable technical problems and welcomes PR submissions to contribute to the project!
