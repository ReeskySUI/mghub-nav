package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
	if categories == nil {
		categories = []*model.Category{}
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

	if items == nil {

		items = []*model.NavItem{}

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
	if len([]rune(req.Description)) > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "描述不能超过80个字符"})
		return
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
	if len([]rune(req.Description)) > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "描述不能超过80个字符"})
		return
	}
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

		if items == nil {

			items = []*model.NavItem{}

		}

		c.JSON(http.StatusOK, gin.H{"data": items})

		return

	}
	items, err := h.store.SearchNavItems(keyword)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return

	}

	if items == nil {

		items = []*model.NavItem{}

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

	if users == nil {

		users = []*model.User{}

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

// ==================== 图片上传与空间管理 ====================

// UploadImage 上传图标图片（内容哈希去重，相同图片只保存一份）
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

	// 读取文件内容（同时用于类型判断和哈希去重）
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}

	// 检查文件类型
	contentType := http.DetectContentType(data)
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
		if ext == ".svg" || ext == ".webp" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".ico" {
			allowed = true
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型: " + contentType})
		return
	}

	// 计算内容哈希，查找是否已存在相同图片
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	// 1. 先查 uploads_meta 表
	existing, err := h.store.FindUploadByHash(hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询图片索引失败"})
		return
	}
	if existing != "" {
		c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + existing, "name": existing, "deduped": true})
		return
	}

	// 2. 兜底扫描 uploads 目录（兼容历史未登记文件）
	if existing == "" {
		existing = h.findFileByHashInDir(hash)
		if existing != "" {
			_ = h.store.RegisterUpload(hash, existing, file.Size)
			c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + existing, "name": existing, "deduped": true})
			return
		}
	}

	// 3. 保存新文件
	if err := os.MkdirAll(h.cfg.Upload.Dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".png"
	}
	filename := fmt.Sprintf("icon_%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(h.cfg.Upload.Dir, filename)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
		return
	}
	if err := h.store.RegisterUpload(hash, filename, file.Size); err != nil {
		// 登记失败不影响上传成功
		log.Printf("登记上传文件失败: %v", err)
	}

	// 返回可访问的 URL
	c.JSON(http.StatusOK, gin.H{
		"url":  "/uploads/" + filename,
		"name": filename,
	})
}

// findFileByHashInDir 扫描上传目录，查找内容哈希相同的已有文件
func (h *APIHandler) findFileByHashInDir(hash string) string {
	entries, err := os.ReadDir(h.cfg.Upload.Dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(h.cfg.Upload.Dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) == hash {
			return e.Name()
		}
	}
	return ""
}

// countUploadRefs 统计某上传文件的引用次数（导航项图标 + 站点设置 Logo/Favicon）
func (h *APIHandler) countUploadRefs(filename string) int {
	url := "/uploads/" + filename
	count := 0
	items, err := h.store.ListNavItems()
	if err == nil {
		for _, it := range items {
			if strings.Contains(it.Icon, url) {
				count++
			}
		}
	}
	settings, err := h.store.GetAllSettings()
	if err == nil {
		if strings.Contains(settings.LogoLight, url) {
			count++
		}
		if strings.Contains(settings.LogoDark, url) {
			count++
		}
		if strings.Contains(settings.Favicon, url) {
			count++
		}
	}
	return count
}

// ListUploads 列出上传的图片（含引用次数）
func (h *APIHandler) ListUploads(c *gin.Context) {
	type uploadInfo struct {
		Name      string    `json:"name"`
		Size      int64     `json:"size"`
		CreatedAt time.Time `json:"created_at"`
		RefCount  int       `json:"ref_count"`
	}

	metas, err := h.store.ListUploads()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 合并目录中未被登记的历史文件
	entries, _ := os.ReadDir(h.cfg.Upload.Dir)
	known := make(map[string]bool)
	var result []uploadInfo

	result = []uploadInfo{}

	for _, m := range metas {
		known[m.Filename] = true
		// 文件可能已被手动删除
		if _, err := os.Stat(filepath.Join(h.cfg.Upload.Dir, m.Filename)); err != nil {
			continue
		}
		result = append(result, uploadInfo{
			Name:      m.Filename,
			Size:      m.Size,
			CreatedAt: m.CreatedAt,
			RefCount:  h.countUploadRefs(m.Filename),
		})
	}
	for _, e := range entries {
		if e.IsDir() || known[e.Name()] {
			continue
		}
		// 跳过 .gitkeep 等隐藏/占位文件
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, uploadInfo{
			Name:      e.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			RefCount:  h.countUploadRefs(e.Name()),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// DeleteUpload 删除上传的图片（被引用时禁止删除）
func (h *APIHandler) DeleteUpload(c *gin.Context) {
	filename := c.Param("name")
	// 防路径穿越
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件名"})
		return
	}

	refs := h.countUploadRefs(filename)
	if refs > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("该图片正被 %d 处引用，请先解除引用再删除", refs)})
		return
	}

	path := filepath.Join(h.cfg.Upload.Dir, filename)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			// 文件已不存在，清理登记即可
			_ = h.store.DeleteUploadMeta(filename)
			c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	_ = h.store.DeleteUploadMeta(filename)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
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

	if visions == nil {

		visions = []*model.HomeVision{}

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

	if links == nil {

		links = []*model.SiteLink{}

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

	if ids == nil {

		ids = []int64{}

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
