# MGHUB-Nav - 自托管导航与首页系统

一个轻量级的自托管导航与首页系统，包含公网首页和成员导航页，基于 Go + Gin + SQLite 构建，单二进制部署。适用于个人/小团队的自托管服务统一入口。

## 功能特性

### 首页（公网可访问）
- 展示项目愿景与设计理念（愿景要点、徽章、主副标题均可后台配置）
- 跳转 Wiki 和导航页的按钮（链接可后台增删改）
- 公网可访问，无需登录；登录后再回首页，右上角按钮由「登录」变为「进入导航」（首页与导航页子域共享登录态）
- 可选择是否在首页显示公告
- 支持浅色 / 深色 / 跟随系统主题

### 导航页（需登录）
- 用户登录认证（账号密码 + bcrypt，含登录失败锁定）
- 按用户角色显示导航项（普通成员 / 管理员 / 超级管理员）
- 按用户分类权限过滤导航项（可为每个用户指定可见分类）
- 导航项实时搜索（名称、描述、URL、分类）+ 分类筛选
- **桌面端左侧钢琴键分类页签**：彩色竖向页签，点击按下后只显示该分类；移动端自动回退为顶部下拉筛选器
- 网格 / 列表布局切换（记忆用户偏好）；网格卡片图标占两行、标题右对齐，下一行显示公网/内网与分类标签，描述限两行、卡片高度统一
- 公网 / 内网服务标签区分
- 顶部公告横幅（维护、宕机通知，可关闭）
- 支持浅色 / 深色 / 跟随系统主题

### 管理后台（/admin）
- **概览**：分类数量、导航项数量、用户数量统计
- **分类管理**：添加 / 编辑 / 删除分类，支持 emoji 图标；表格支持搜索、筛选、点击表头排序
- **导航管理**：添加 / 编辑 / 删除导航项，支持图标上传（图片 / emoji / 从图片库选择 / 直接填 URL）、可见范围、公网内网标记、排序；表格支持搜索、按分类 / 类型 / 可见范围筛选、点击表头排序
- **图片管理**（管理员）：相册式浏览已上传图片，显示引用次数，支持上传新图、一键复制引用 URL、删除（被引用时拒绝并提示）；相同内容图片按哈希自动去重
- **公告管理**（仅超管）：发布 / 编辑 / 停用 / 删除公告，支持 info / warn / error 级别、生效与失效时间、是否允许关闭、显示位置（导航页 / 首页）
- **站点设置**（仅超管）：
  - 基本信息：站点标题、副标题、Logo 文字、页脚
  - Logo 上传：浅色 Logo、深色 Logo、favicon、首页徽章图标
  - 首页内容：徽章、主标题、副标题、愿景要点（动态增删改）
  - 链接管理：头部 / CTA / 页脚链接（增删改）
  - 默认主题（浅色 / 深色 / 自动）与实验性外观（见下）
- **用户管理**（仅超管）：
  - 添加 / 编辑 / 禁用 / 删除用户、重置密码
  - 角色分配（普通成员 / 管理员；超级管理员唯一且不可禁用、不可改角色）
  - 用户分类权限管理（指定用户可见的导航分类）

## 技术栈

- **后端**：Go 1.21+ + Gin
- **数据库**：SQLite（modernc.org/sqlite 纯 Go 驱动，无 CGO 依赖）
- **前端**：原生 HTML/CSS/JS（Go template 嵌入二进制）
- **认证**：bcrypt 密码哈希 + Session Cookie
- **部署**：单二进制 + systemd + nginx 反向代理

## 主题配色

### 浅色模式（6:3:1）
| 颜色 | HEX | 占比 | 用途 |
|------|-----|------|------|
| 主蓝 | `#0175CB` | 60% | 导航栏、主按钮、链接、标题 |
| 橙 | `#FE8835` | 30% | CTA 按钮、标签、强调、hover |
| 米白 | `#FCF2DA` | 10% | 卡片背景、辅助背景、输入框 focus |

### 深色模式（6:3:1）
| 颜色 | HEX | 占比 | 用途 |
|------|-----|------|------|
| 深棕 | `#4C221B` | 60% | 页面背景、卡片背景 |
| 浅米 | `#F6E0C9` | 30% | 文字、标题、边框 |
| 深紫 | `#423A4F` | 10% | 辅助背景、标签、强调 |

## 快速开始

### 编译

```bash
# 克隆项目
git clone https://github.com/ReeskySUI/mghub-nav.git
cd mghub-nav

# Windows 本地编译
CGO_ENABLED=0 go build -o portal.exe .

# Linux 交叉编译（部署到 VPS）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o portal .

# macOS 交叉编译
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o portal-darwin .
```

### 运行

```bash
# 首次运行会自动创建数据库和默认超级管理员
./portal

# 指定配置文件
./portal /path/to/config.yaml
```

默认超级管理员：
- 用户名：`admin`
- 密码：`008800`

> **重要**：首次登录后请立即修改密码！

### 本地测试

修改 hosts 文件添加（将 example.com 替换为你的域名）：
```
127.0.0.1 example.com
127.0.0.1 nav.example.com
```

然后访问：
- 首页：http://example.com:8080
- 导航页：http://nav.example.com:8080
- 管理后台：http://nav.example.com:8080/admin

> 本地开发时，直接访问 http://localhost:8080 会默认路由到导航页域。

## 配置文件（config.yaml）

```yaml
server:
  # 监听地址：127.0.0.1 仅本地访问（配合 nginx 反向代理），0.0.0.0 监听所有网卡
  listen_address: "127.0.0.1"
  port: 8080
  # 首页域名（用于 Host 路由区分）
  home_host: "example.com"
  # 导航页域名
  nav_host: "nav.example.com"
  # 运行模式：debug / release
  mode: "release"

database:
  # SQLite 数据库文件路径
  path: "./data/portal.db"

auth:
  # Session 密钥（生产环境请修改）
  session_secret: "portal-change-me-in-production"
  # Session 过期时间（小时）
  session_max_age: 168
  # 登录失败锁定阈值
  max_login_attempts: 5
  # 锁定时长（分钟）
  lockout_minutes: 15

upload:
  # 上传文件保存目录
  dir: "./web/static/uploads"
  # 允许的图片类型
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "image/webp"
    - "image/svg+xml"
  # 单文件最大尺寸（MB）
  max_size_mb: 2

site:
  title: "My Portal"
  subtitle: "自托管导航中心"
  # 首页愿景描述
  vision:
    - "以极低成本获得稳定、可控、可自托管的数字服务"
    - "成员共享硬件、带宽与运维能力，避免重复造轮子"
    - "所有服务透明可查，架构可演进、可迁移、可回溯"
  # Wiki 地址
  wiki_url: "https://wiki.example.com"
  # 导航页地址
  nav_url: "https://nav.example.com"
  # 是否开启首页：true 展示首页愿景页；false 时访问首页域名直接跳转到导航页（只跑导航页场景）
  # 注意：该项缺省（不写）时默认为 true，只有显式写 false 才关闭首页
  home_enabled: true
```

## 部署到 Linux（VPS）

### 1. 上传文件

```bash
# 创建项目目录
sudo mkdir -p /opt/portal
sudo mkdir -p /opt/portal/data
sudo mkdir -p /opt/portal/web/static/uploads

# 上传二进制和配置
sudo cp portal /opt/portal/
sudo cp config.yaml /opt/portal/
sudo chmod +x /opt/portal/portal

# 创建专用用户
sudo useradd -r -s /bin/false portal
sudo chown -R portal:portal /opt/portal
```

### 2. 配置 systemd 服务

创建 `/etc/systemd/system/portal.service`：

```ini
[Unit]
Description=Self-Hosted Portal
After=network.target

[Service]
Type=simple
User=portal
Group=portal
WorkingDirectory=/opt/portal
ExecStart=/opt/portal/portal /opt/portal/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable portal
sudo systemctl start portal
sudo systemctl status portal
```

### 3. 配置 Nginx 反向代理

创建 `/etc/nginx/conf.d/portal.conf`：

```nginx
# 首页
# 80>>443 跳转 有统一的跳转可以不用添加这两个
server {
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# 导航页 80>>443 跳转 有统一的跳转可以不用添加这两个
server {
    listen 80;
    server_name nav.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name nav.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

重载 Nginx：
```bash
sudo nginx -t
sudo systemctl reload nginx
```

> **关键：两个域名必须反代到同一个 portal 进程**，且 nginx 都要带 `proxy_set_header Host $host;`，portal 靠 Host 头区分首页域 / 导航页域。

### 关于登录态跨子域共享

登录成功后，portal 会按请求域名自动推导 Cookie 的 Domain：在 `nav.example.com` 登录会写入 `Domain=example.com` 的 Cookie，于是 `example.com`（首页）与 `nav.example.com`（导航页）共享登录态，首页右上角才能在登录后显示「进入导航」。因此**首页与导航页需在同一注册域下的不同子域**（如 `example.com` 与 `nav.example.com`）。通过 `localhost` / IP 直连时不设置 Domain，仅当前主机生效。

### 进阶：SNI 分流与 Cloudflare

- **同一台机器 443 端口同时跑网站和代理（Xray / Reality 等）**：可用 nginx `stream` 模块的 `ssl_preread` 按 SNI 分流，把网站域名（首页、导航页及其他站点）统一转发到本地 http 层 nginx（如 `127.0.0.1:8443`），把伪装域名转发给代理端口（如 `127.0.0.1:1443`）。`map` 中多个网站域名可以指向**同一个** 8443 上游，也支持通配匹配；http 层 nginx 再按 `server_name` 反代到 portal 的 8080。
- **Cloudflare 橙云代理**：DNS 解析到 Cloudflare 后，源站 nginx 需正常提供 443 证书（可用 Cloudflare 源站证书），Cloudflare SSL/TLS 模式建议设为 **Full（严格）**；若源站 443 握手异常会报 `Error 525 SSL handshake failed`，可用 `openssl s_client -connect <源站IP>:443 -servername <域名>` 排查 SNI 是否正确命中网站上游。

## 用户角色与权限

| 角色 | 导航页 | 管理后台 | 用户管理 | 说明 |
|------|--------|----------|----------|------|
| 超级管理员 | ✅ | ✅（全部） | ✅ | 唯一，可管理所有用户和角色、站点设置、公告；永不禁用、不可改角色 |
| 管理员 | ✅ | ✅（部分） | ❌ | 可管理分类、导航项、图片管理；**不可**管理用户、站点设置、公告 |
| 普通成员 | ✅ | ❌ | ❌ | 只能访问导航页，按权限查看导航项 |

### 导航项可见范围

每个导航项可设置可见范围：
- **所有登录用户可见**：普通成员、管理员、超管都能看到
- **成员及以上可见**：管理员和超管可见（普通成员不可见）
- **仅管理员可见**：只有管理员和超管可见

### 用户分类权限

超级管理员可以为每个用户指定可见的导航分类：
- 不勾选任何分类：用户可见所有分类
- 勾选特定分类：用户只能看到勾选分类下的导航项

### 实验性外观（仅超管）

管理后台「站点设置 → 实验性外观」可自定义品牌色与页面背景，作用于首页和导航页：
- **启用自定义外观**（Switch 开关）：关闭时使用内置配色；开启后无论当前是浅色 / 深色 / 自动模式，都应用以下自定义外观
- **预设色板**：一键填入内置浅色（#0175CB / #FE8835 / #FCF2DA）或深色暖调（#4C221B / #F6E0C9 / #423A4F）三套 6:3:1 配色
- **主色 / 强调色 / 背景色 HEX**：对应 6:3:1 的三个颜色，可手动输入
- **页面背景图 URL**：可选，填 `/uploads/xxx.jpg`（先在图片管理上传并复制地址）或外链；优先级高于背景色
- **内容透明度**：滑块调整内容区不透明度（30%–100%），让背景图透出来
- **背景毛玻璃**：Switch 开关，开启后对背景层做模糊（backdrop-filter），形成毛玻璃质感

### 公告（仅超管）

用于机器故障、运维停机等场景通知成员：
- **级别**：info（提示）/ warn（警告）/ error（严重），横幅配色不同
- **显示位置**：可勾选显示在导航页、首页
- **生效时间**：可设置开始 / 结束时间，留空则长期有效；仅「启用」且在时间范围内的公告展示
- **是否可关闭**：允许成员手动收起横幅，或强制常驻
- 导航页与首页顶部各有一条横幅位，取当前生效的公告

## 项目结构

```
mghub-nav/
├── main.go                  # 入口文件，路由注册
├── config.yaml              # 配置文件
├── go.mod                   # Go 模块定义
├── go.sum                   # 依赖锁定
├── internal/
│   ├── auth/                # 认证模块（密码哈希、Session 管理）
│   │   └── auth.go
│   ├── config/              # 配置加载
│   │   └── config.go
│   ├── handler/             # HTTP 处理器
│   │   ├── home.go          # 首页处理器
│   │   ├── nav.go           # 导航页处理器
│   │   ├── auth_handler.go  # 登录/登出处理器
│   │   ├── admin.go         # 管理后台处理器
│   │   ├── api.go           # REST API 处理器
│   │   └── helpers.go       # 首页/导航页地址、Cookie Domain 推导
│   ├── middleware/          # 中间件
│   │   └── middleware.go    # 认证、权限、Host 路由中间件
│   ├── model/               # 数据模型
│   │   └── model.go         # 用户、分类、导航项等结构体定义
│   └── store/               # 数据库操作
│       └── store.go         # SQLite CRUD 操作
└── web/
    ├── templates/           # HTML 模板
    │   ├── home.html        # 首页模板
    │   ├── nav.html         # 导航页模板
    │   ├── login.html       # 登录页模板
    │   ├── admin.html       # 管理后台模板
    │   └── error.html       # 错误页模板
    └── static/              # 静态资源
        ├── css/
        │   └── style.css    # 全局样式（含浅色/深色主题）
        └── js/
            ├── theme.js     # 主题切换逻辑
            ├── nav.js       # 导航页交互（搜索、布局切换）
            └── admin.js     # 管理后台交互（CRUD 操作）
```

## API 接口

所有 API 接口都需要登录认证（Session Cookie），管理类接口需要管理员或超管权限。

### 分类管理
- `GET /api/categories` - 获取分类列表
- `POST /api/categories` - 创建分类（管理员）
- `PUT /api/categories/:id` - 更新分类（管理员）
- `DELETE /api/categories/:id` - 删除分类（管理员）

### 导航项管理
- `GET /api/nav-items` - 获取导航项列表
- `GET /api/nav-items/search?q=关键词` - 搜索导航项
- `POST /api/nav-items` - 创建导航项（管理员）
- `PUT /api/nav-items/:id` - 更新导航项（管理员）
- `DELETE /api/nav-items/:id` - 删除导航项（管理员）

### 用户管理（仅超管）
- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建用户
- `PUT /api/users/:id` - 更新用户
- `DELETE /api/users/:id` - 删除用户
- `POST /api/users/:id/reset-password` - 重置密码
- `GET /api/users/:id/categories` - 获取用户可见分类
- `PUT /api/users/:id/categories` - 更新用户可见分类

### 站点设置（仅超管）
- `GET /api/settings` - 获取站点设置
- `POST /api/settings` - 更新站点设置

### 愿景要点（仅超管可写）
- `GET /api/visions` - 获取愿景列表
- `POST /api/visions` - 创建愿景
- `PUT /api/visions/:id` - 更新愿景
- `DELETE /api/visions/:id` - 删除愿景

### 站点链接（仅超管可写）
- `GET /api/links` - 获取链接列表
- `POST /api/links` - 创建链接
- `PUT /api/links/:id` - 更新链接
- `DELETE /api/links/:id` - 删除链接

### 公告（仅超管可写）
- `GET /api/announcements` - 获取公告列表（所有登录用户可读）
- `POST /api/announcements` - 创建公告
- `PUT /api/announcements/:id` - 更新公告
- `DELETE /api/announcements/:id` - 删除公告

### 其他
- `POST /api/change-password` - 修改当前用户密码
- `POST /api/upload` - 上传图片（管理员，内容哈希去重，可在图片管理与导航项图标中复用）
- `GET /api/uploads` - 上传图片列表（管理员，相册视图，含引用次数与可复制 URL）
- `DELETE /api/uploads/:name` - 删除图片（管理员，被引用时拒绝）

## 修改 Go Module 名称

本项目的 Go module 名为 `mghub-portal`。如果你想修改为自己的名称：

1. 修改 `go.mod` 中的 module 名
2. 全局替换所有 `import` 语句中的 `mghub-portal` 为新名称
3. 重新编译

```bash
# 示例：修改为 my-portal
sed -i 's/mghub-portal/my-portal/g' go.mod main.go internal/**/*.go
go mod tidy
go build -o portal .
```

## 常见问题

### 1. 登录后提示 Session 过期

检查 `config.yaml` 中的 `session_secret` 是否修改，以及浏览器是否禁用了 Cookie。

### 2. 上传图片失败

检查 `upload.dir` 目录是否存在且有写入权限，以及文件大小是否超过 `max_size_mb` 限制。

### 3. 首页和导航页显示相同内容

检查 `config.yaml` 中的 `home_host` 和 `nav_host` 是否配置正确，以及 Nginx 的 `proxy_set_header Host $host` 是否配置。

### 4. 如何备份数据

直接复制 `data/portal.db` 文件即可。建议定期备份。

### 5. 如何重置超级管理员密码

删除数据库文件后重启服务，会重新创建默认超管（admin/008800）。注意：这会清除所有数据！

### 6. 上传的图片无法删除

图片正在被导航项、Logo 或 favicon 引用时不可删除（后台会提示引用次数）。先解除引用（修改导航项图标 / 站点设置中的 Logo），再回到"图片管理"删除即可。相同内容的图片重复上传会自动去重，只保留一份。

### 7. 访问首页域名被自动跳到导航页

两种原因：① `config.yaml` 中显式设置了 `home_enabled: false`（删掉该行或改为 `true`，缺省即开启）；② nginx 没有把首页域名的 `Host` 头透传（缺少 `proxy_set_header Host $host;`），导致 portal 无法识别为首页域。可用 `curl -I https://你的首页域` 观察是否返回 302 到导航页域。

### 8. 在导航页登录后，首页右上角仍显示「登录」

说明登录 Cookie 没有在两个子域间共享。确认首页与导航页位于同一注册域的子域（如 `example.com` 与 `nav.example.com`），且由同一个 portal 进程服务；清理浏览器旧 Cookie 后重新登录即可。本地用 `localhost` / IP 访问时不支持跨子域共享，属正常现象。

## License

MIT
