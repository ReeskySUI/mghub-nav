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
		c.Redirect(http.StatusFound, h.postLoginURL(c))
		return
	}
	settings, _ := h.store.GetAllSettings()
	c.HTML(http.StatusOK, "login.html", gin.H{
		"site_title":    settings.SiteTitle,
		"logo_text":     settings.LogoText,
		"logo_light":    settings.LogoLight,
		"logo_dark":     settings.LogoDark,
		"favicon":       settings.Favicon,
		"home_url":      resolveHomeURL(settings.HomeURL, c.Request.Host),
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

	// 先清除旧版本可能残留的 host-only Cookie（Domain 为空时写入），
	// 避免同名 Cookie 并存导致请求带回已失效的旧 token
	c.SetCookie(middleware.SessionCookie, "", -1, "/", "", false, true)
	// 写入首页/导航页共享的跨子域 Cookie
	c.SetCookie(
		middleware.SessionCookie,
		token,
		3600*24*7,
		"/",
		cookieDomain(c.Request.Host),
		false, // 本地测试时不强制 secure
		true,
	)
	c.Redirect(http.StatusFound, h.postLoginURL(c))
}

// postLoginURL 登录成功后的落地地址：home 域跳到导航页，nav 域留根路径
func (h *AuthHandler) postLoginURL(c *gin.Context) string {
	if middleware.GetCurrentSite(c) == "home" {
		settings, _ := h.store.GetAllSettings()
		return resolveNavURL(settings.NavURL, c.Request.Host)
	}
	return "/"
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 注销请求中携带的所有同名会话
	for _, ck := range c.Request.Cookies() {
		if ck.Name == middleware.SessionCookie && ck.Value != "" {
			h.authMgr.Logout(ck.Value)
		}
	}
	// 同时清理 host-only 与跨子域两种 Cookie
	c.SetCookie(middleware.SessionCookie, "", -1, "/", "", false, true)
	c.SetCookie(middleware.SessionCookie, "", -1, "/", cookieDomain(c.Request.Host), false, true)
	c.Redirect(http.StatusFound, "/login")
}
