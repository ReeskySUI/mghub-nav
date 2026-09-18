package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/auth"
	"mghub-portal/internal/config"
	"mghub-portal/internal/handler"
	"mghub-portal/internal/middleware"
	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

//go:embed web/templates/*.html
var templateFS embed.FS

//go:embed web/static/css/*.css web/static/js/*.js
var staticFS embed.FS

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := store.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	if err := initSuperAdmin(db); err != nil {
		log.Fatalf("初始化超级管理员失败: %v", err)
	}

	if err := initDefaultSettings(db, cfg); err != nil {
		log.Fatalf("初始化默认设置失败: %v", err)
	}

	if err := initDefaultContent(db, cfg); err != nil {
		log.Fatalf("初始化默认内容失败: %v", err)
	}

	authMgr := auth.New(db, cfg.Auth.SessionMaxAge, cfg.Auth.MaxLoginAttempts, cfg.Auth.LockoutMinutes)

	if err := os.MkdirAll(cfg.Upload.Dir, 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower":     strings.ToLower,
		"hasPrefix": strings.HasPrefix,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "web/templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.Use(middleware.HostRouter(cfg.Server.HomeHost, cfg.Server.NavHost))

	staticSub, _ := fs.Sub(staticFS, "web/static")
	r.StaticFS("/static", http.FS(staticSub))
	r.StaticFS("/uploads", http.Dir(cfg.Upload.Dir))

	homeHandler := handler.NewHomeHandler(db)
	navHandler := handler.NewNavHandler(db)
	authHandler := handler.NewAuthHandler(db, authMgr)
	adminHandler := handler.NewAdminHandler(db)
	apiHandler := handler.NewAPIHandler(cfg, db, authMgr)

	// 根路径：根据 Host 分发
	r.GET("/", func(c *gin.Context) {
		site := middleware.GetCurrentSite(c)
		if site == "home" {
			if !cfg.Site.HomeEnabled {
				c.Redirect(http.StatusFound, "https://"+cfg.Server.NavHost)
				return
			}
			homeHandler.Index(c)
			return
		}
		token, _ := c.Cookie(middleware.SessionCookie)
		if token == "" {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		session := authMgr.GetSession(token)
		if session == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		c.Set(middleware.CtxUserKey, session)
		navHandler.Index(c)
	})

	navGroup := r.Group("/")
	{
		navGroup.GET("/login", authHandler.LoginPage)
		navGroup.POST("/login", authHandler.LoginSubmit)
		navGroup.POST("/logout", authHandler.Logout)
		navGroup.GET("/admin", middleware.AuthRequired(authMgr), middleware.AdminRequired(), adminHandler.Dashboard)
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthRequired(authMgr))
	{
		apiGroup.GET("/categories", apiHandler.ListCategories)
		apiGroup.POST("/categories", middleware.AdminRequired(), apiHandler.CreateCategory)
		apiGroup.PUT("/categories/:id", middleware.AdminRequired(), apiHandler.UpdateCategory)
		apiGroup.DELETE("/categories/:id", middleware.AdminRequired(), apiHandler.DeleteCategory)

		apiGroup.GET("/nav-items", apiHandler.ListNavItems)
		apiGroup.GET("/nav-items/search", apiHandler.SearchNavItems)
		apiGroup.POST("/nav-items", middleware.AdminRequired(), apiHandler.CreateNavItem)
		apiGroup.PUT("/nav-items/:id", middleware.AdminRequired(), apiHandler.UpdateNavItem)
		apiGroup.DELETE("/nav-items/:id", middleware.AdminRequired(), apiHandler.DeleteNavItem)

		apiGroup.GET("/users", middleware.SuperAdminRequired(), apiHandler.ListUsers)
		apiGroup.POST("/users", middleware.SuperAdminRequired(), apiHandler.CreateUser)
		apiGroup.PUT("/users/:id", middleware.SuperAdminRequired(), apiHandler.UpdateUser)
		apiGroup.DELETE("/users/:id", middleware.SuperAdminRequired(), apiHandler.DeleteUser)
		apiGroup.POST("/users/:id/reset-password", middleware.SuperAdminRequired(), apiHandler.ResetUserPassword)
		apiGroup.GET("/users/:id/categories", middleware.SuperAdminRequired(), apiHandler.GetUserCategories)
		apiGroup.PUT("/users/:id/categories", middleware.SuperAdminRequired(), apiHandler.UpdateUserCategories)

		apiGroup.POST("/change-password", apiHandler.ChangePassword)
		apiGroup.POST("/upload", middleware.AdminRequired(), apiHandler.UploadImage)

		// 上传图片管理
		apiGroup.GET("/uploads", middleware.AdminRequired(), apiHandler.ListUploads)
		apiGroup.DELETE("/uploads/:name", middleware.AdminRequired(), apiHandler.DeleteUpload)

		// 站点设置（仅超管）
		apiGroup.GET("/settings", apiHandler.GetSettings)
		apiGroup.POST("/settings", middleware.SuperAdminRequired(), apiHandler.UpdateSettings)

		// 愿景要点
		apiGroup.GET("/visions", apiHandler.ListVisions)
		apiGroup.POST("/visions", middleware.SuperAdminRequired(), apiHandler.CreateVision)
		apiGroup.PUT("/visions/:id", middleware.SuperAdminRequired(), apiHandler.UpdateVision)
		apiGroup.DELETE("/visions/:id", middleware.SuperAdminRequired(), apiHandler.DeleteVision)

		// 站点链接
		apiGroup.GET("/links", apiHandler.ListLinks)
		apiGroup.POST("/links", middleware.SuperAdminRequired(), apiHandler.CreateLink)
		apiGroup.PUT("/links/:id", middleware.SuperAdminRequired(), apiHandler.UpdateLink)
		apiGroup.DELETE("/links/:id", middleware.SuperAdminRequired(), apiHandler.DeleteLink)
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
			return
		}
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "页面不存在"})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.ListenAddress, cfg.Server.Port)
	log.Printf("Portal 启动中...")
	log.Printf("  首页域: %s", cfg.Server.HomeHost)
	log.Printf("  导航域: %s", cfg.Server.NavHost)
	log.Printf("  监听地址: %s", addr)
	log.Printf("  默认超管: admin / 008800 (请尽快修改密码)")
	log.Printf("服务已启动，按 Ctrl+C 停止")

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func initSuperAdmin(db *store.Store) error {
	users, err := db.ListUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Role == model.RoleSuperAdmin {
			// 超管永不禁用，确保状态正常
			if u.Status != model.UserActive {
				if err := db.UpdateUser(u.ID, u.DisplayName, u.Role, model.UserActive); err != nil {
					return fmt.Errorf("恢复超管状态失败: %w", err)
				}
				log.Printf("已恢复超管 %s 的正常状态", u.Username)
			}
			return nil
		}
	}
	// 没有 super_admin，检查 admin 用户是否存在
	existing, err := db.GetUserByUsername("admin")
	if err == nil && existing != nil {
		// admin 存在但 role 不是 super_admin，升级为超级管理员并确保状态正常
		err = db.UpdateUser(existing.ID, existing.DisplayName, model.RoleSuperAdmin, model.UserActive)
		if err != nil {
			return fmt.Errorf("升级超级管理员失败: %w", err)
		}
		log.Println("已将 admin 用户升级为超级管理员")
		return nil
	}
	// admin 不存在，创建
	hash, err := auth.HashPassword("008800")
	if err != nil {
		return err
	}
	_, err = db.CreateUser("admin", hash, "超级管理员", model.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}
	log.Println("已创建默认超级管理员: admin / 008800")
	return nil
}

func initDefaultSettings(db *store.Store, cfg *config.Config) error {
	defaults := map[string]string{
		"site_title":      cfg.Site.Title,
		"site_subtitle":   cfg.Site.Subtitle,
		"logo_text":       "M",
		"home_badge":      cfg.Site.Subtitle + " · 小社群共建",
		"home_hero_title": "一体化导航中心",
		"home_hero_sub":   "以极低成本，获得稳定、可控、可自托管的数字服务",
		"home_vision_1":   "以极低成本获得稳定、可控、可自托管的数字服务",
		"home_vision_2":   "成员共享硬件、带宽与运维能力，避免重复造轮子",
		"home_vision_3":   "所有服务透明可查，架构可演进、可迁移、可回溯",
		"footer_text":     "成员共建 · 架构可演进、可迁移、可回溯",
		"wiki_url":        cfg.Site.WikiURL,
		"nav_url":         cfg.Site.NavURL,
		"home_url":        "https://" + cfg.Server.HomeHost,
		"default_theme":   "auto",
	}
	return db.InitDefaultSettings(defaults)
}

func initDefaultContent(db *store.Store, cfg *config.Config) error {
	// 默认愿景要点（仅在表为空时初始化）
	visions, err := db.ListVisions()
	if err != nil {
		return err
	}
	if len(visions) == 0 {
		defaultVisions := []string{
			"以极低成本获得稳定、可控、可自托管的数字服务",
			"成员共享硬件、带宽与运维能力，避免重复造轮子",
			"所有服务透明可查，架构可演进、可迁移、可回溯",
		}
		for i, v := range defaultVisions {
			if _, err := db.CreateVision(v, i); err != nil {
				return err
			}
		}
		log.Println("已初始化默认愿景要点")
	}

	// 默认链接（仅在表为空时初始化）
	links, err := db.ListLinks("")
	if err != nil {
		return err
	}
	if len(links) == 0 {
		defaultLinks := []struct {
			name, url, icon, linkType string
			sortOrder                  int
		}{
			{"Wiki 文档", cfg.Site.WikiURL, "📖", "header", 0},
			{"浏览 Wiki", cfg.Site.WikiURL, "📖", "cta", 0},
			{"成员登录", cfg.Site.NavURL + "/login", "🔑", "cta", 1},
		}
		for _, l := range defaultLinks {
			if _, err := db.CreateLink(l.name, l.url, l.icon, l.linkType, l.sortOrder); err != nil {
				return err
			}
		}
		log.Println("已初始化默认链接")
	}
	return nil
}
