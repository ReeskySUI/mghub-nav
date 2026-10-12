package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/auth"
	"mghub-portal/internal/model"
)

const (
	SessionCookie = "portal_session"
	CtxUserKey    = "current_user"
	CtxSiteKey    = "current_site"
)

// ResolveSession 从请求携带的所有同名 Cookie（及 Bearer 头）中返回第一个
// 仍然有效的会话。浏览器可能同时持有旧版本的 host-only Cookie 与新版本的
// 跨子域 Cookie（同名），只取第一个会拿到已失效的旧值而误判未登录，因此逐个校验。
func ResolveSession(c *gin.Context, authMgr *auth.Manager) *model.Session {
	if authHeader := c.GetHeader("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		if t := strings.TrimPrefix(authHeader, "Bearer "); t != "" {
			if sess := authMgr.GetSession(t); sess != nil {
				return sess
			}
		}
	}
	for _, ck := range c.Request.Cookies() {
		if ck.Name == SessionCookie && ck.Value != "" {
			if sess := authMgr.GetSession(ck.Value); sess != nil {
				return sess
			}
		}
	}
	return nil
}

// AuthRequired 登录认证中间件
func AuthRequired(authMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := ResolveSession(c, authMgr)
		if session == nil {
			redirectLogin(c)
			return
		}
		c.Set(CtxUserKey, session)
		c.Next()
	}
}

// AdminRequired 管理员权限中间件（admin 或 super_admin）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetCurrentUser(c)
		if session == nil {
			redirectLogin(c)
			return
		}
		if session.Role != model.RoleAdmin && session.Role != model.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SuperAdminRequired 超级管理员权限中间件
func SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetCurrentUser(c)
		if session == nil {
			redirectLogin(c)
			return
		}
		if session.Role != model.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要超级管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// HostRouter 基于 Host 头的站点路由中间件
// 用于区分首页域和导航页域
func HostRouter(homeHost, navHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		// 去掉端口
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		// 本地开发环境默认归为 nav 域
		if host == "localhost" || host == "127.0.0.1" || host == "" {
			c.Set(CtxSiteKey, "nav")
		} else if strings.HasPrefix(host, "nav.") || host == navHost {
			c.Set(CtxSiteKey, "nav")
		} else {
			c.Set(CtxSiteKey, "home")
		}
		c.Next()
	}
}

// GetCurrentUser 从上下文获取当前用户
func GetCurrentUser(c *gin.Context) *model.Session {
	v, exists := c.Get(CtxUserKey)
	if !exists {
		return nil
	}
	session, ok := v.(*model.Session)
	if !ok {
		return nil
	}
	return session
}

// GetCurrentSite 获取当前站点类型（home / nav）
func GetCurrentSite(c *gin.Context) string {
	v, exists := c.Get(CtxSiteKey)
	if !exists {
		return "home"
	}
	site, _ := v.(string)
	return site
}

// redirectLogin 重定向到登录页（API 请求返回 JSON）
func redirectLogin(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		c.Abort()
		return
	}
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}
