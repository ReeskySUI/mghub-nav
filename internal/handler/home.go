package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/middleware"
	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

// HomeHandler 首页处理器
type HomeHandler struct {
	store *store.Store
}

// NewHomeHandler 创建首页处理器
func NewHomeHandler(s *store.Store) *HomeHandler {
	return &HomeHandler{store: s}
}

// Index 首页
func (h *HomeHandler) Index(c *gin.Context) {
	if middleware.GetCurrentSite(c) != "home" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	settings, err := h.store.GetAllSettings()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "加载站点设置失败"})
		return
	}

	// 从 home_visions 表读取愿景要点（动态）
	visions, err := h.store.ListVisions()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "加载愿景要点失败"})
		return
	}
	visionTexts := make([]string, 0, len(visions))
	for _, v := range visions {
		visionTexts = append(visionTexts, v.Content)
	}

	// 读取头部链接和 CTA 链接
	headerLinks, _ := h.store.ListLinks("header")
	ctaLinks, _ := h.store.ListLinks("cta")

	session := middleware.GetCurrentUser(c)
	isLoggedIn := session != nil

	c.HTML(http.StatusOK, "home.html", gin.H{
		"site_title":    settings.SiteTitle,
		"site_subtitle": settings.SiteSubtitle,
		"logo_text":     settings.LogoText,
		"logo_light":    settings.LogoLight,
		"logo_dark":     settings.LogoDark,
		"favicon":       settings.Favicon,
		"home_badge":      settings.HomeBadge,
		"home_badge_icon": settings.HomeBadgeIcon,
		"custom_primary":  settings.CustomPrimary,
		"custom_accent":   settings.CustomAccent,
		"custom_bg":       settings.CustomBg,
		"custom_bg_color": settings.CustomBgColor,
		"custom_theme_enabled": settings.CustomThemeEnabled,
		"announcement": activeHomeAnnouncement(h.store),
		"hero_title":    settings.HomeHeroTitle,
		"hero_sub":      settings.HomeHeroSub,
		"vision":        visionTexts,
		"footer_text":   settings.FooterText,
		"header_links":  headerLinks,
		"cta_links":     ctaLinks,
		"default_theme": settings.DefaultTheme,
		"logged_in":     isLoggedIn,
		"current_user":  session,
	})
}


func activeHomeAnnouncement(store *store.Store) *model.Announcement {
	a, err := store.GetActiveAnnouncement()
	if err != nil || a == nil || !a.ShowOnHome {
		return nil
	}
	return a
}
