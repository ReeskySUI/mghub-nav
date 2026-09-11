package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/auth"
	"mghub-portal/internal/model"
)

const (
	SessionCookie = "mghub_session"
	CtxUserKey    = "current_user"
	CtxSiteKey    = "current_site"
)

// AuthRequired 登录认证中间件
func AuthRequired(authMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookie)
		if err != nil || token == "" {
			// 检查 Authorization header（API 调用）
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if token == "" {
			redirectLogin(c)
			return
		}
		session := authMgr.GetSession(token)
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
// 用于区分首页域(mghub.top)和导航页域(nav.mghub.top)
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
