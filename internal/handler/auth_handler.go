package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/auth"
	"mghub-portal/internal/middleware"
	"mghub-portal/internal/store"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	store   *store.Store
	authMgr *auth.Manager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(s *store.Store, mgr *auth.Manager) *AuthHandler {
	return &AuthHandler{store: s, authMgr: mgr}
}

// LoginPage 登录页
func (h *AuthHandler) LoginPage(c *gin.Context) {
	session := middleware.GetCurrentUser(c)
	if session != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	settings, _ := h.store.GetAllSettings()
	c.HTML(http.StatusOK, "login.html", gin.H{
		"site_title":    settings.SiteTitle,
		"logo_text":     settings.LogoText,
		"logo_light":    settings.LogoLight,
		"logo_dark":     settings.LogoDark,
		"favicon":       settings.Favicon,
		"home_url":      "/",
		"default_theme": settings.DefaultTheme,
		"error":         c.Query("error"),
	})
}

// LoginSubmit 登录提交
func (h *AuthHandler) LoginSubmit(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.Redirect(http.StatusFound, "/login?error=用户名和密码不能为空")
		return
	}

	token, err := h.authMgr.Login(username, password)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error="+err.Error())
		return
	}

	c.SetCookie(
		middleware.SessionCookie,
		token,
		3600*24*7,
		"/",
		"",
		false, // 本地测试时不强制 secure
		true,
	)
	c.Redirect(http.StatusFound, "/")
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(middleware.SessionCookie)
	if err == nil && token != "" {
		h.authMgr.Logout(token)
	}
	c.SetCookie(middleware.SessionCookie, "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
