package handler

import (
	"net/http"

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

	// 从站点设置读取 Wiki / 首页地址
	wikiURL := settings.WikiURL
	homeURL := settings.HomeURL
	if wikiURL == "" {
		wikiURL = "#"
	}
	if homeURL == "" {
		homeURL = "/"
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

	// 根据导航项可见范围过滤
	// all: 所有登录用户可见
	// member: 成员及以上可见（等同于 all，因为能登录的都是成员及以上）
	// admin: 仅管理员及以上可见
	var visibleItems []*model.NavItem
	for _, item := range items {
		if isAdmin {
			// 管理员和超管可见所有导航项
			visibleItems = append(visibleItems, item)
		} else {
			// 普通成员只能看到 all 或 member 的导航项
			if item.VisibleRoles == model.VisibleAll || item.VisibleRoles == model.VisibleMember {
				visibleItems = append(visibleItems, item)
			}
		}
	}

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
		for _, item := range visibleItems {
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
	for _, item := range visibleItems {
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
		"custom_primary": settings.CustomPrimary,
		"custom_accent":  settings.CustomAccent,
		"custom_bg":      settings.CustomBg,
		"custom_theme_enabled": settings.CustomThemeEnabled,
		"categories":    categoriesWithItems,
		"all_items":     allowedItems,
		"logged_in":     true,
		"is_admin":      isAdmin,
		"current_user":  session,
	})
}
