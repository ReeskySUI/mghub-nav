package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/middleware"
	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	store *store.Store
}

// NewAdminHandler 创建管理后台处理器
func NewAdminHandler(s *store.Store) *AdminHandler {
	return &AdminHandler{store: s}
}

// Dashboard 管理后台首页
func (h *AdminHandler) Dashboard(c *gin.Context) {
	session := middleware.GetCurrentUser(c)
	isSuperAdmin := session.Role == model.RoleSuperAdmin

	settings, _ := h.store.GetAllSettings()
	categories, _ := h.store.ListCategories()
	items, _ := h.store.ListNavItems()
	users, _ := h.store.ListUsers()

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"site_title":     settings.SiteTitle,
		"logo_text":      settings.LogoText,
		"default_theme":  settings.DefaultTheme,
		"settings":       settings,
		"categories":     categories,
		"nav_items":      items,
		"users":          users,
		"current_user":   session,
		"is_super_admin": isSuperAdmin,
		"category_count": len(categories),
		"item_count":     len(items),
		"user_count":     len(users),
	})
}
