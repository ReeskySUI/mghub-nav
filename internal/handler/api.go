package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mghub-portal/internal/auth"
	"mghub-portal/internal/config"
	"mghub-portal/internal/middleware"
	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

// APIHandler API 处理器
type APIHandler struct {
	cfg     *config.Config
	store   *store.Store
	authMgr *auth.Manager
}

// NewAPIHandler 创建 API 处理器
func NewAPIHandler(cfg *config.Config, s *store.Store, mgr *auth.Manager) *APIHandler {
	return &APIHandler{cfg: cfg, store: s, authMgr: mgr}
}

// ==================== 分类 API ====================

// ListCategories 获取分类列表
func (h *APIHandler) ListCategories(c *gin.Context) {
	categories, err := h.store.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// CreateCategory 创建分类
func (h *APIHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	cat, err := h.store.CreateCategory(req.Name, req.Icon, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cat})
}

// UpdateCategory 更新分类
func (h *APIHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req struct {
		Name      string `json:"name" binding:"required"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.store.UpdateCategory(id, req.Name, req.Icon, req.SortOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteCategory 删除分类
func (h *APIHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	if err := h.store.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 导航项 API ====================

// ListNavItems 获取导航项列表
func (h *APIHandler) ListNavItems(c *gin.Context) {
	items, err := h.store.ListNavItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateNavItem 创建导航项
func (h *APIHandler) CreateNavItem(c *gin.Context) {
	var req model.NavItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if req.Name == "" || req.URL == "" || req.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称、URL和分类不能为空"})
		return
	}
	if req.IconType == "" {
		req.IconType = "emoji"
	}
	item, err := h.store.CreateNavItem(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// UpdateNavItem 更新导航项
func (h *APIHandler) UpdateNavItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req model.NavItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	req.ID = id
	if err := h.store.UpdateNavItem(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteNavItem 删除导航项
func (h *APIHandler) DeleteNavItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	if err := h.store.DeleteNavItem(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SearchNavItems 搜索导航项
func (h *APIHandler) SearchNavItems(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		items, _ := h.store.ListNavItems()
		c.JSON(http.StatusOK, gin.H{"data": items})
		return
	}
	items, err := h.store.SearchNavItems(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// ==================== 用户管理 API（超管） ====================

// ListUsers 获取用户列表
func (h *APIHandler) ListUsers(c *gin.Context) {
	users, err := h.store.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// CreateUser 创建用户
func (h *APIHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	// 校验角色
	role := model.Role(req.Role)
	if role != model.RoleAdmin && role != model.RoleMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user, err := h.store.CreateUser(req.Username, hash, req.DisplayName, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateUser 更新用户
func (h *APIHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 不能修改自己的角色和状态（防止误操作把自己锁出去）
	session := middleware.GetCurrentUser(c)
	if session.UserID == id {
		if req.Role != "" || req.Status != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改自己的角色或状态"})
			return
		}
	}

	// 获取当前用户信息，用于空值时保持原值
	user, err := h.store.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	// 超管永不禁用，角色不可修改
	if user.Role == model.RoleSuperAdmin {
		req.Role = string(model.RoleSuperAdmin)
		req.Status = string(model.UserActive)
	}

	// 空值时保持原值
	role := user.Role
	if req.Role != "" {
		role = model.Role(req.Role)
	}
	status := user.Status
	if req.Status != "" {
		status = model.UserStatus(req.Status)
	}

	if err := h.store.UpdateUser(id, req.DisplayName, role, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// ResetUserPassword 重置用户密码
func (h *APIHandler) ResetUserPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	if err := h.store.UpdatePassword(id, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}

// DeleteUser 删除用户
func (h *APIHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	session := middleware.GetCurrentUser(c)
	if session.UserID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}
	if err := h.store.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 修改自己的密码 ====================

// ChangePassword 修改当前用户密码
func (h *APIHandler) ChangePassword(c *gin.Context) {
	session := middleware.GetCurrentUser(c)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少6位"})
		return
	}
	user, err := h.store.GetUserByID(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}
	if !auth.CheckPassword(req.OldPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	if err := h.store.UpdatePassword(user.ID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// ==================== 图片上传 ====================

// UploadImage 上传图标图片
func (h *APIHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未获取到文件"})
		return
	}

	// 检查文件大小
	maxSize := int64(h.cfg.Upload.MaxSizeMB) * 1024 * 1024
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件大小不能超过 %d MB", h.cfg.Upload.MaxSizeMB)})
		return
	}

	// 检查文件类型
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"})
		return
	}
	defer src.Close()

	// 读取文件头判断类型
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	allowed := false
	for _, t := range h.cfg.Upload.AllowedTypes {
		if strings.HasPrefix(contentType, t) || contentType == t {
			allowed = true
			break
		}
	}
	// 扩展名兜底判断（SVG 等文本格式 DetectContentType 可能检测为 text/xml）
	if !allowed {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".svg" || ext == ".webp" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			allowed = true
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型: " + contentType})
		return
	}

	// 确保上传目录存在
	if err := os.MkdirAll(h.cfg.Upload.Dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".png"
	}
	filename := fmt.Sprintf("icon_%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(h.cfg.Upload.Dir, filename)

	// 重新读取并保存文件（因为前面读了文件头）
	src.Seek(0, 0)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
		return
	}

	// 返回可访问的 URL
	c.JSON(http.StatusOK, gin.H{
		"url":  "/uploads/" + filename,
		"name": filename,
	})
}

// ==================== 站点设置 API ====================

// GetSettings 获取站点设置
func (h *APIHandler) GetSettings(c *gin.Context) {
	settings, err := h.store.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// UpdateSettings 更新站点设置
func (h *APIHandler) UpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 只允许更新已知的设置键
	allowedKeys := map[string]bool{
		"site_title": true, "site_subtitle": true, "logo_text": true,
		"logo_light": true, "logo_dark": true, "favicon": true,
		"home_badge": true, "home_hero_title": true, "home_hero_sub": true,
		"footer_text": true, "wiki_url": true, "nav_url": true, "home_url": true,
		"default_theme": true,
	}

	settings := make(map[string]string)
	for k, v := range req {
		if allowedKeys[k] {
			settings[k] = v
		}
	}

	if err := h.store.UpdateSettings(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "设置更新成功"})
}

// ==================== 愿景要点 API ====================

func (h *APIHandler) ListVisions(c *gin.Context) {
	visions, err := h.store.ListVisions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": visions})
}

func (h *APIHandler) CreateVision(c *gin.Context) {
	var req struct {
		Content   string `json:"content"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容不能为空"})
		return
	}
	vision, err := h.store.CreateVision(req.Content, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vision})
}

func (h *APIHandler) UpdateVision(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req struct {
		Content   string `json:"content"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.store.UpdateVision(id, req.Content, req.SortOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *APIHandler) DeleteVision(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	if err := h.store.DeleteVision(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 站点链接 API ====================

func (h *APIHandler) ListLinks(c *gin.Context) {
	linkType := c.Query("type")
	links, err := h.store.ListLinks(linkType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": links})
}

func (h *APIHandler) CreateLink(c *gin.Context) {
	var req model.SiteLink
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if req.Name == "" || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和URL不能为空"})
		return
	}
	if req.LinkType == "" {
		req.LinkType = "header"
	}
	link, err := h.store.CreateLink(req.Name, req.URL, req.Icon, req.LinkType, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": link})
}

func (h *APIHandler) UpdateLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	var req model.SiteLink
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.store.UpdateLink(id, req.Name, req.URL, req.Icon, req.LinkType, req.SortOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *APIHandler) DeleteLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}
	if err := h.store.DeleteLink(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 用户分类权限 ====================

// GetUserCategories 获取用户可见分类
func (h *APIHandler) GetUserCategories(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}
	ids, err := h.store.GetUserCategoryIDs(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ids})
}

// UpdateUserCategories 更新用户可见分类
func (h *APIHandler) UpdateUserCategories(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}
	var req struct {
		CategoryIDs []int64 `json:"category_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.store.UpdateUserCategories(id, req.CategoryIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "分类权限更新成功"})
}
