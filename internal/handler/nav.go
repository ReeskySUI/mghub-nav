package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/middleware"
	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

// NavHandler 导航页处理器
type NavHandler struct {
	store *store.Store
}

// NewNavHandler 创建导航页处理器
func NewNavHandler(s *store.Store) *NavHandler {
	return &NavHandler{store: s}
}

// Index 导航页
func (h *NavHandler) Index(c *gin.Context) {
	if middleware.GetCurrentSite(c) != "nav" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	settings, err := h.store.GetAllSettings()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "加载站点设置失败"})
		return
	}

	session := middleware.GetCurrentUser(c)
	isAdmin := session != nil && (session.Role == model.RoleAdmin || session.Role == model.RoleSuperAdmin)

	// 从 site_links 表获取链接
	headerLinks, _ := h.store.ListLinks("header")
	wikiURL := ""
	homeURL := ""
	for _, l := range headerLinks {
		if wikiURL == "" && (strings.Contains(l.Name, "wiki") || strings.Contains(l.Name, "Wiki") || strings.Contains(l.Name, "文档")) {
			wikiURL = l.URL
		}
		if homeURL == "" && (strings.Contains(l.Name, "首页") || strings.Contains(l.Name, "home") || strings.Contains(l.Name, "Home")) {
			homeURL = l.URL
		}
	}
	if wikiURL == "" {
		wikiURL = "https://wiki.mghub.top"
	}
	if homeURL == "" {
		homeURL = "https://mghub.top"
	}

	categories, err := h.store.ListCategories()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "加载分类失败"})
		return
	}
	items, err := h.store.ListNavItems()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "加载导航项失败"})
		return
	}

	// 根据用户分类权限过滤（空列表表示可见所有分类；管理员/超管不受限制）
	allowedCategoryIDs, _ := h.store.GetUserCategoryIDs(session.UserID)
	allowedSet := make(map[int64]bool)
	for _, id := range allowedCategoryIDs {
		allowedSet[id] = true
	}
	hasRestriction := !isAdmin && len(allowedCategoryIDs) > 0

	type CategoryWithItems struct {
		Category *model.Category
		Items    []*model.NavItem
	}
	var categoriesWithItems []CategoryWithItems
	for _, cat := range categories {
		if hasRestriction && !allowedSet[cat.ID] {
			continue // 用户无权限查看此分类
		}
		var catItems []*model.NavItem
		for _, item := range items {
			if item.CategoryID == cat.ID {
				catItems = append(catItems, item)
			}
		}
		categoriesWithItems = append(categoriesWithItems, CategoryWithItems{
			Category: cat,
			Items:    catItems,
		})
	}

	// 过滤 all_items，只包含用户有权限的分类下的导航项
	var allowedItems []*model.NavItem
	for _, item := range items {
		if !hasRestriction || allowedSet[item.CategoryID] {
			allowedItems = append(allowedItems, item)
		}
	}

	c.HTML(http.StatusOK, "nav.html", gin.H{
		"site_title":    settings.SiteTitle,
		"site_subtitle": settings.SiteSubtitle,
		"logo_text":     settings.LogoText,
		"logo_light":    settings.LogoLight,
		"logo_dark":     settings.LogoDark,
		"favicon":       settings.Favicon,
		"footer_text":   settings.FooterText,
		"wiki_url":      wikiURL,
		"home_url":      homeURL,
		"default_theme": settings.DefaultTheme,
		"categories":    categoriesWithItems,
		"all_items":     allowedItems,
		"logged_in":     true,
		"is_admin":      isAdmin,
		"current_user":  session,
	})
}
