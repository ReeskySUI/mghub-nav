# MGHUB-Nav - 自托管导航与首页系统

一个轻量级的自托管导航与首页系统，包含公网首页和成员导航页，基于 Go + Gin + SQLite 构建，单二进制部署。适用于个人/小团队的自托管服务统一入口。

## 功能特性

### 首页（公网可访问）
- 展示项目愿景与设计理念（内容可后台配置）
- 跳转 Wiki 和导航页的按钮（链接可后台配置）
- 公网可访问，无需登录
- 支持浅色/深色主题，可手动切换或跟随系统

### 导航页（需登录）
- 用户登录认证（账号密码 + bcrypt，含登录失败锁定）
- 按用户角色显示导航项（普通成员 / 管理员 / 超级管理员）
- 按用户分类权限过滤导航项（可为每个用户指定可见分类）
- 导航项实时搜索（支持名称、描述、URL、分类）
- 网格 / 列表布局切换（记忆用户偏好）
- 公网/内网服务标签区分
- 支持浅色/深色主题

### 管理后台（/admin）
- **概览**：分类数量、导航项数量、用户数量统计
- **分类管理**：添加/编辑/删除分类，支持 emoji 图标
- **导航管理**：添加/编辑/删除导航项，支持图标上传（图片/emoji）、可见范围设置、排序
- **站点设置**：
  - 基本信息：站点标题、副标题、Logo 文字、页脚
  - Logo 上传：浅色 Logo、深色 Logo、favicon
  - 首页内容：徽章、主标题、副标题、愿景要点（动态增删改）
  - 链接与主题：首页按钮链接、默认主题（浅色/深色/自动）
- **用户管理**（仅超管）：
  - 添加/编辑/禁用/删除用户
  - 重置密码
  - 角色分配（普通成员/管理员）
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
sudo useradd -r -s /sbin/nologin portal
sudo chown -R portal:portal /opt/portal
```

### 2. systemd 服务

创建 `/etc/systemd/system/portal.service`：

```ini
[Unit]
Description=Self-Hosted Portal Service
After=network.target

[Service]
Type=simple
User=portal
Group=portal
WorkingDirectory=/opt/portal
ExecStart=/opt/portal/portal /opt/portal/config.yaml
Restart=always
RestartSec=5
Environment=GIN_MODE=release

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

### 3. nginx 反向代理

确保 nginx 已安装并配置好 SSL 证书（可用 certbot 或现有的 SNI 分流架构）。

创建 `/etc/nginx/conf.d/portal.conf`：

```nginx
# 首页 example.com
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate     /path/to/example.com.crt;
    ssl_certificate_key /path/to/example.com.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# 导航页 nav.example.com
server {
    listen 443 ssl http2;
    server_name nav.example.com;

    ssl_certificate     /path/to/nav.example.com.crt;
    ssl_certificate_key /path/to/nav.example.com.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

重载 nginx：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 4. 防火墙

```bash
# RockyLinux / CentOS
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# Ubuntu / Debian
sudo ufw allow https
```

## 用户角色说明

| 角色 | 权限 |
|------|------|
| 超级管理员 (super_admin) | 全部权限：导航管理、分类管理、用户管理（含角色分配和分类权限）、站点设置 |
| 管理员 (admin) | 导航管理、分类管理、站点设置；不能管理用户 |
| 普通成员 (member) | 登录后查看导航页、搜索、切换布局 |

### 角色规则
- 超级管理员唯一且不可删除，永不禁用
- 超级管理员用户名不可修改
- 管理员不能访问用户管理页面
- 登录失败 5 次后锁定 15 分钟

### 用户分类权限
- 可为每个用户指定可见的导航分类
- 不指定（空）表示该用户可见所有分类
- 管理员和超管不受分类权限限制，可见所有分类

## 项目结构

```
mghub-nav/
├── main.go                    # 入口：路由注册、初始化、Host 分发
├── config.yaml                # 配置文件
├── go.mod / go.sum           # 依赖管理
├── .gitignore
├── internal/
│   ├── config/
│   │   └── config.go          # 配置加载
│   ├── model/
│   │   └── model.go           # 数据模型（User/Category/NavItem/Session/SiteSetting/HomeVision/SiteLink/UserCategory）
│   ├── store/
│   │   └── store.go           # SQLite 数据访问层（8 张表 CRUD）
│   ├── auth/
│   │   └── auth.go            # 认证：bcrypt 密码哈希 + Session 管理 + 登录失败锁定
│   ├── middleware/
│   │   └── middleware.go      # 中间件：AuthRequired/AdminRequired/SuperAdminRequired/HostRouter
│   └── handler/
│       ├── home.go            # 首页处理器
│       ├── nav.go             # 导航页处理器（含分类权限过滤）
│       ├── auth_handler.go    # 登录/登出处理器
│       ├── admin.go           # 管理后台页面处理器
│       └── api.go             # REST API（分类/导航项/用户/上传/站点设置/愿景/链接/用户分类权限）
├── web/
│   ├── templates/             # HTML 模板（embed 嵌入二进制）
│   │   ├── home.html          # 首页
│   │   ├── nav.html           # 导航页
│   │   ├── login.html         # 登录页
│   │   ├── admin.html         # 管理后台（概览/分类/导航/站点设置/用户管理）
│   │   └── error.html         # 错误页
│   └── static/                # 静态资源（embed 嵌入二进制）
│       ├── css/
│       │   └── style.css      # 主题样式（浅色/深色变量、组件样式）
│       ├── js/
│       │   ├── nav.js         # 导航页交互（搜索/布局切换/渲染）
│       │   ├── admin.js       # 管理后台交互（CRUD/表单/上传）
│       │   └── theme.js       # 主题切换模块（浅色/深色/auto 三态）
│       └── uploads/           # 用户上传的图标/Logo（运行时生成）
└── data/                      # 运行时生成
    └── portal.db              # SQLite 数据库
```

## 数据库表结构

| 表名 | 说明 |
|------|------|
| users | 用户表（用户名、密码哈希、角色、状态、登录失败计数） |
| categories | 导航分类表（名称、图标、排序） |
| nav_items | 导航项表（名称、URL、图标、描述、分类、可见范围、排序） |
| sessions | Session 表（Token、用户 ID、过期时间） |
| settings | 站点设置表（key-value 存储，含 Logo/favicon/首页内容等） |
| home_visions | 首页愿景要点表（内容、排序） |
| site_links | 首页链接表（名称、URL、主题、排序） |
| user_categories | 用户分类权限表（user_id + category_id 联合主键） |

## 维护

### 查看日志

```bash
sudo journalctl -u portal -f
```

### 重启服务

```bash
sudo systemctl restart portal
```

### 备份数据库

```bash
sudo cp /opt/portal/data/portal.db /backup/portal-$(date +%Y%m%d).db
```

### 更新版本

```bash
sudo systemctl stop portal
sudo cp new-portal /opt/portal/portal
sudo chmod +x /opt/portal/portal
sudo systemctl start portal
```

### 重置超管密码

如果忘记超管密码，停止服务后删除数据库中的 admin 用户，重启服务会自动重新创建默认超管（admin/008800）：

```bash
sudo systemctl stop portal
sqlite3 /opt/portal/data/portal.db "DELETE FROM users WHERE username='admin';"
sudo systemctl start portal
```

## 安全建议

1. **修改默认密码**：首次登录 admin/008800 后立即修改
2. **修改 session_secret**：config.yaml 中改为随机字符串
3. **HTTPS**：确保 nginx 配置了 SSL，生产环境将 Cookie 标记为 Secure
4. **定期备份**：SQLite 单文件，定期备份 data/portal.db
5. **限制注册**：本系统不开放公开注册，用户由超管在后台添加
6. **上传目录**：确保 uploads 目录不可执行脚本（nginx 配置中仅作为静态文件）
7. **文件上传**：限制上传文件大小和类型，当前支持 png/jpg/webp/gif/svg

## 开发说明

### 本地开发

```bash
# 安装依赖
go mod download

# 开发模式运行
go run .
```

### 修改 Go Module 名称

本项目的 Go module 名称为 `mghub-portal`（go.mod 中定义）。如果你希望修改为自己的名称：

1. 修改 `go.mod` 中的 `module mghub-portal` 为 `module your-module-name`
2. 全局替换所有 Go 文件中的 import 路径 `mghub-portal/` 为 `your-module-name/`
3. 重新编译

### 前端开发

前端模板和静态资源通过 Go embed 嵌入二进制，开发时修改后需要重新编译才能生效。

### API 接口

所有 API 接口位于 `/api/` 路径下，需要登录认证：

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/categories | 分类列表 | 登录用户 |
| POST | /api/categories | 创建分类 | 管理员 |
| PUT | /api/categories/:id | 更新分类 | 管理员 |
| DELETE | /api/categories/:id | 删除分类 | 管理员 |
| GET | /api/nav-items | 导航项列表 | 登录用户 |
| POST | /api/nav-items | 创建导航项 | 管理员 |
| PUT | /api/nav-items/:id | 更新导航项 | 管理员 |
| DELETE | /api/nav-items/:id | 删除导航项 | 管理员 |
| GET | /api/users | 用户列表 | 超管 |
| POST | /api/users | 创建用户 | 超管 |
| PUT | /api/users/:id | 更新用户 | 超管 |
| DELETE | /api/users/:id | 删除用户 | 超管 |
| POST | /api/users/:id/reset-password | 重置密码 | 超管 |
| GET | /api/users/:id/categories | 用户分类权限 | 超管 |
| PUT | /api/users/:id/categories | 更新用户分类权限 | 超管 |
| GET | /api/settings | 站点设置 | 登录用户 |
| POST | /api/settings | 保存站点设置 | 管理员 |
| GET | /api/visions | 愿景要点列表 | 登录用户 |
| POST | /api/visions | 创建愿景要点 | 管理员 |
| PUT | /api/visions/:id | 更新愿景要点 | 管理员 |
| DELETE | /api/visions/:id | 删除愿景要点 | 管理员 |
| GET | /api/links | 链接列表 | 登录用户 |
| POST | /api/links | 创建链接 | 管理员 |
| PUT | /api/links/:id | 更新链接 | 管理员 |
| DELETE | /api/links/:id | 删除链接 | 管理员 |
| POST | /api/upload | 上传图片 | 管理员 |
| POST | /api/change-password | 修改自己的密码 | 登录用户 |

## License

MIT
