package model

import "time"

// Role 用户角色
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleMember     Role = "member"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserDisabled UserStatus = "disabled"
)

// VisibleRole 导航项可见角色
type VisibleRole string

const (
	VisibleAll    VisibleRole = "all"    // 所有登录用户可见
	VisibleMember VisibleRole = "member" // 成员及以上可见
	VisibleAdmin  VisibleRole = "admin"  // 仅管理员可见
)

// User 用户表
type User struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	PasswordHash  string     `json:"-"`
	DisplayName   string     `json:"display_name"`
	Role          Role       `json:"role"`
	Status        UserStatus `json:"status"`
	LoginAttempts int        `json:"-"`
	LockedUntil   *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

// Category 导航分类
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// NavItem 导航项
type NavItem struct {
	ID           int64       `json:"id"`
	CategoryID   int64       `json:"category_id"`
	Name         string      `json:"name"`
	URL          string      `json:"url"`
	Icon         string      `json:"icon"`         // emoji 或 /uploads/xxx.png
	IconType     string      `json:"icon_type"`    // emoji / image
	Description  string      `json:"description"`
	IsPublic     bool        `json:"is_public"`    // true=公网服务, false=内网服务
	VisibleRoles VisibleRole `json:"visible_roles"` // all / member / admin
	SortOrder    int         `json:"sort_order"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	// 关联字段
	CategoryName string `json:"category_name,omitempty"`
}

// Session 登录会话（内存存储，重启后需重新登录）
type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SiteSetting 站点设置（key-value 存储）
type SiteSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SiteSettings 站点设置集合（方便模板使用）
type SiteSettings struct {
	SiteTitle     string `json:"site_title"`
	SiteSubtitle  string `json:"site_subtitle"`
	LogoText      string `json:"logo_text"`
	LogoLight     string `json:"logo_light"`     // 浅色模式 logo 图片 URL
	LogoDark      string `json:"logo_dark"`      // 深色模式 logo 图片 URL
	Favicon       string `json:"favicon"`         // favicon 图片 URL
	HomeBadge     string `json:"home_badge"`
	HomeHeroTitle string `json:"home_hero_title"`
	HomeHeroSub   string `json:"home_hero_sub"`
	FooterText    string `json:"footer_text"`
	WikiURL       string `json:"wiki_url"`
	HomeURL       string `json:"home_url"`
	DefaultTheme  string `json:"default_theme"` // light / dark / auto
}

// HomeVision 首页愿景要点（动态管理）
type HomeVision struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// SiteLink 站点链接（动态管理）
type SiteLink struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`       // 链接名称，如 "Wiki 文档"
	URL       string `json:"url"`        // 链接地址
	Icon      string `json:"icon"`       // emoji 图标
	LinkType  string `json:"link_type"`  // header / cta / footer
	SortOrder int    `json:"sort_order"`
}

// UserCategory 用户可见分类关联
type UserCategory struct {
	UserID     int64 `json:"user_id"`
	CategoryID int64 `json:"category_id"`
}

// UploadMeta 上传图片元数据（用于去重与空间管理）
type UploadMeta struct {
	Hash      string    `json:"hash"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
