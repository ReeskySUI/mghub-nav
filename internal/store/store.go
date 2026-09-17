package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mghub-portal/internal/model"

	_ "modernc.org/sqlite"
)

// Store 数据访问层
type Store struct {
	db *sql.DB
}

// New 创建 Store 并初始化数据库
func New(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("设置数据库参数失败: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate 创建表结构
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		display_name TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT 'member',
		status TEXT NOT NULL DEFAULT 'active',
		login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		icon TEXT NOT NULL DEFAULT '',
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS nav_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		icon TEXT NOT NULL DEFAULT '',
		icon_type TEXT NOT NULL DEFAULT 'emoji',
		description TEXT NOT NULL DEFAULT '',
		is_public INTEGER NOT NULL DEFAULT 0,
		visible_roles TEXT NOT NULL DEFAULT 'all',
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_nav_items_category ON nav_items(category_id);
	CREATE INDEX IF NOT EXISTS idx_nav_items_sort ON nav_items(sort_order);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS home_visions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS site_links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		icon TEXT NOT NULL DEFAULT '',
		link_type TEXT NOT NULL DEFAULT 'header',
		sort_order INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS user_categories (
		user_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, category_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS uploads_meta (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL UNIQUE,
		filename TEXT NOT NULL,
		size INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// 为旧数据库添加 visible_roles 字段（如果不存在）
	_, _ = s.db.Exec("ALTER TABLE nav_items ADD COLUMN visible_roles TEXT NOT NULL DEFAULT 'all'")

	return nil
}

// ==================== 用户相关 ====================

func (s *Store) CreateUser(username, passwordHash, displayName string, role model.Role) (*model.User, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (username, password_hash, display_name, role) VALUES (?, ?, ?, ?)`,
		username, passwordHash, displayName, role,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetUserByID(id)
}

func (s *Store) GetUserByID(id int64) (*model.User, error) {
	return s.scanUser(s.db.QueryRow(`SELECT * FROM users WHERE id = ?`, id))
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	return s.scanUser(s.db.QueryRow(`SELECT * FROM users WHERE username = ?`, username))
}

func (s *Store) ListUsers() ([]*model.User, error) {
	rows, err := s.db.Query(`SELECT * FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) UpdateUser(id int64, displayName string, role model.Role, status model.UserStatus) error {
	_, err := s.db.Exec(
		`UPDATE users SET display_name = ?, role = ?, status = ? WHERE id = ?`,
		displayName, role, status, id,
	)
	return err
}

func (s *Store) UpdatePassword(id int64, passwordHash string) error {
	_, err := s.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	return err
}

func (s *Store) DeleteUser(id int64) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	if user.Role == model.RoleSuperAdmin {
		return errors.New("超级管理员不可删除")
	}
	_, err = s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (s *Store) UpdateLoginAttempts(id int64, attempts int, lockedUntil *time.Time) error {
	_, err := s.db.Exec(
		`UPDATE users SET login_attempts = ?, locked_until = ? WHERE id = ?`,
		attempts, lockedUntil, id,
	)
	return err
}

func (s *Store) RecordLogin(id int64) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE users SET last_login_at = ?, login_attempts = 0, locked_until = NULL WHERE id = ?`,
		now, id,
	)
	return err
}

func (s *Store) scanUser(scanner interface {
	Scan(...interface{}) error
}) (*model.User, error) {
	var u model.User
	var lockedUntil, lastLoginAt sql.NullTime
	err := scanner.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName,
		&u.Role, &u.Status, &u.LoginAttempts, &lockedUntil,
		&u.CreatedAt, &lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	return &u, nil
}

// ==================== 分类相关 ====================

func (s *Store) CreateCategory(name, icon string, sortOrder int) (*model.Category, error) {
	res, err := s.db.Exec(
		`INSERT INTO categories (name, icon, sort_order) VALUES (?, ?, ?)`,
		name, icon, sortOrder,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetCategoryByID(id)
}

func (s *Store) GetCategoryByID(id int64) (*model.Category, error) {
	var c model.Category
	err := s.db.QueryRow(`SELECT * FROM categories WHERE id = ?`, id).Scan(
		&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("分类不存在")
		}
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListCategories() ([]*model.Category, error) {
	rows, err := s.db.Query(`SELECT * FROM categories ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (s *Store) UpdateCategory(id int64, name, icon string, sortOrder int) error {
	_, err := s.db.Exec(
		`UPDATE categories SET name = ?, icon = ?, sort_order = ? WHERE id = ?`,
		name, icon, sortOrder, id,
	)
	return err
}

func (s *Store) DeleteCategory(id int64) error {
	_, err := s.db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

// ==================== 导航项相关 ====================

func (s *Store) CreateNavItem(item *model.NavItem) (*model.NavItem, error) {
	now := time.Now()
	if item.VisibleRoles == "" {
		item.VisibleRoles = model.VisibleAll
	}
	res, err := s.db.Exec(
		`INSERT INTO nav_items (category_id, name, url, icon, icon_type, description, is_public, visible_roles, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.CategoryID, item.Name, item.URL, item.Icon, item.IconType,
		item.Description, item.IsPublic, item.VisibleRoles, item.SortOrder, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetNavItemByID(id)
}

func (s *Store) GetNavItemByID(id int64) (*model.NavItem, error) {
	var item model.NavItem
	var isPublic int
	err := s.db.QueryRow(
		`SELECT id, category_id, name, url, icon, icon_type, description, is_public, visible_roles, sort_order, created_at, updated_at
		 FROM nav_items WHERE id = ?`, id,
	).Scan(
		&item.ID, &item.CategoryID, &item.Name, &item.URL, &item.Icon,
		&item.IconType, &item.Description, &isPublic, &item.VisibleRoles, &item.SortOrder,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("导航项不存在")
		}
		return nil, err
	}
	item.IsPublic = isPublic == 1
	if item.VisibleRoles == "" {
		item.VisibleRoles = model.VisibleAll
	}
	return &item, nil
}

func (s *Store) ListNavItems() ([]*model.NavItem, error) {
	rows, err := s.db.Query(`
		SELECT n.id, n.category_id, n.name, n.url, n.icon, n.icon_type,
		       n.description, n.is_public, n.visible_roles, n.sort_order, n.created_at, n.updated_at,
		       c.name as category_name
		FROM nav_items n
		LEFT JOIN categories c ON n.category_id = c.id
		ORDER BY c.sort_order ASC, n.sort_order ASC, n.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.NavItem
	for rows.Next() {
		var item model.NavItem
		var isPublic int
		if err := rows.Scan(
			&item.ID, &item.CategoryID, &item.Name, &item.URL, &item.Icon,
			&item.IconType, &item.Description, &isPublic, &item.VisibleRoles, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt, &item.CategoryName,
		); err != nil {
			return nil, err
		}
		item.IsPublic = isPublic == 1
		if item.VisibleRoles == "" {
			item.VisibleRoles = model.VisibleAll
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

func (s *Store) ListNavItemsByCategory(categoryID int64) ([]*model.NavItem, error) {
	rows, err := s.db.Query(
		`SELECT id, category_id, name, url, icon, icon_type, description, is_public, visible_roles, sort_order, created_at, updated_at
		 FROM nav_items WHERE category_id = ? ORDER BY sort_order ASC, id ASC`,
		categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.NavItem
	for rows.Next() {
		var item model.NavItem
		var isPublic int
		if err := rows.Scan(
			&item.ID, &item.CategoryID, &item.Name, &item.URL, &item.Icon,
			&item.IconType, &item.Description, &isPublic, &item.VisibleRoles, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.IsPublic = isPublic == 1
		if item.VisibleRoles == "" {
			item.VisibleRoles = model.VisibleAll
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

func (s *Store) UpdateNavItem(item *model.NavItem) error {
	now := time.Now()
	if item.VisibleRoles == "" {
		item.VisibleRoles = model.VisibleAll
	}
	_, err := s.db.Exec(
		`UPDATE nav_items SET category_id=?, name=?, url=?, icon=?, icon_type=?,
		 description=?, is_public=?, visible_roles=?, sort_order=?, updated_at=? WHERE id=?`,
		item.CategoryID, item.Name, item.URL, item.Icon, item.IconType,
		item.Description, item.IsPublic, item.VisibleRoles, item.SortOrder, now, item.ID,
	)
	return err
}

func (s *Store) DeleteNavItem(id int64) error {
	_, err := s.db.Exec(`DELETE FROM nav_items WHERE id = ?`, id)
	return err
}

func (s *Store) SearchNavItems(keyword string) ([]*model.NavItem, error) {
	pattern := "%" + keyword + "%"
	rows, err := s.db.Query(`
		SELECT n.id, n.category_id, n.name, n.url, n.icon, n.icon_type,
		       n.description, n.is_public, n.visible_roles, n.sort_order, n.created_at, n.updated_at,
		       c.name as category_name
		FROM nav_items n
		LEFT JOIN categories c ON n.category_id = c.id
		WHERE n.name LIKE ? OR n.description LIKE ? OR n.url LIKE ?
		ORDER BY c.sort_order ASC, n.sort_order ASC
	`, pattern, pattern, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.NavItem
	for rows.Next() {
		var item model.NavItem
		var isPublic int
		if err := rows.Scan(
			&item.ID, &item.CategoryID, &item.Name, &item.URL, &item.Icon,
			&item.IconType, &item.Description, &isPublic, &item.VisibleRoles, &item.SortOrder,
			&item.CreatedAt, &item.UpdatedAt, &item.CategoryName,
		); err != nil {
			return nil, err
		}
		item.IsPublic = isPublic == 1
		if item.VisibleRoles == "" {
			item.VisibleRoles = model.VisibleAll
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

// ==================== 站点设置相关 ====================

// GetAllSettings 获取所有站点设置
func (s *Store) GetAllSettings() (*model.SiteSettings, error) {
	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return &model.SiteSettings{
		SiteTitle:     m["site_title"],
		SiteSubtitle:  m["site_subtitle"],
		LogoText:      m["logo_text"],
		LogoLight:     m["logo_light"],
		LogoDark:      m["logo_dark"],
		Favicon:       m["favicon"],
		HomeBadge:     m["home_badge"],
		HomeHeroTitle: m["home_hero_title"],
		HomeHeroSub:   m["home_hero_sub"],
		FooterText:    m["footer_text"],
		WikiURL:       m["wiki_url"],
		HomeURL:       m["home_url"],
		DefaultTheme:  m["default_theme"],
	}, rows.Err()
}

// UpdateSetting 更新单个设置
func (s *Store) UpdateSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		key, value,
	)
	return err
}

// UpdateSettings 批量更新设置
func (s *Store) UpdateSettings(settings map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for k, v := range settings {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// InitDefaultSettings 初始化默认设置（仅在 key 不存在时插入）
func (s *Store) InitDefaultSettings(defaults map[string]string) error {
	for k, v := range defaults {
		var count int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM settings WHERE key = ?`, k).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if err := s.UpdateSetting(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// ==================== 首页愿景要点 ====================

func (s *Store) ListVisions() ([]*model.HomeVision, error) {
	rows, err := s.db.Query(`SELECT id, content, sort_order, created_at FROM home_visions ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.HomeVision
	for rows.Next() {
		var v model.HomeVision
		if err := rows.Scan(&v.ID, &v.Content, &v.SortOrder, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (s *Store) CreateVision(content string, sortOrder int) (*model.HomeVision, error) {
	res, err := s.db.Exec(`INSERT INTO home_visions (content, sort_order) VALUES (?, ?)`, content, sortOrder)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.HomeVision{ID: id, Content: content, SortOrder: sortOrder}, nil
}

func (s *Store) UpdateVision(id int64, content string, sortOrder int) error {
	_, err := s.db.Exec(`UPDATE home_visions SET content=?, sort_order=? WHERE id=?`, content, sortOrder, id)
	return err
}

func (s *Store) DeleteVision(id int64) error {
	_, err := s.db.Exec(`DELETE FROM home_visions WHERE id=?`, id)
	return err
}

// ==================== 站点链接 ====================

func (s *Store) ListLinks(linkType string) ([]*model.SiteLink, error) {
	var rows *sql.Rows
	var err error
	if linkType != "" {
		rows, err = s.db.Query(`SELECT id, name, url, icon, link_type, sort_order FROM site_links WHERE link_type=? ORDER BY sort_order ASC, id ASC`, linkType)
	} else {
		rows, err = s.db.Query(`SELECT id, name, url, icon, link_type, sort_order FROM site_links ORDER BY link_type ASC, sort_order ASC, id ASC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.SiteLink
	for rows.Next() {
		var l model.SiteLink
		if err := rows.Scan(&l.ID, &l.Name, &l.URL, &l.Icon, &l.LinkType, &l.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, &l)
	}
	return list, rows.Err()
}

func (s *Store) CreateLink(name, url, icon, linkType string, sortOrder int) (*model.SiteLink, error) {
	res, err := s.db.Exec(`INSERT INTO site_links (name, url, icon, link_type, sort_order) VALUES (?, ?, ?, ?, ?)`,
		name, url, icon, linkType, sortOrder)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.SiteLink{ID: id, Name: name, URL: url, Icon: icon, LinkType: linkType, SortOrder: sortOrder}, nil
}

func (s *Store) UpdateLink(id int64, name, url, icon, linkType string, sortOrder int) error {
	_, err := s.db.Exec(`UPDATE site_links SET name=?, url=?, icon=?, link_type=?, sort_order=? WHERE id=?`,
		name, url, icon, linkType, sortOrder, id)
	return err
}

func (s *Store) DeleteLink(id int64) error {
	_, err := s.db.Exec(`DELETE FROM site_links WHERE id=?`, id)
	return err
}

// ==================== 用户分类权限 ====================

// GetUserCategoryIDs 获取用户可见的分类ID列表（空列表表示可见所有分类）
func (s *Store) GetUserCategoryIDs(userID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT category_id FROM user_categories WHERE user_id=?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UpdateUserCategories 更新用户可见分类（全量替换）
func (s *Store) UpdateUserCategories(userID int64, categoryIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM user_categories WHERE user_id=?`, userID); err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO user_categories (user_id, category_id) VALUES (?, ?)`, userID, cid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ==================== 上传图片元数据 ====================

// FindUploadByHash 根据内容哈希查找已存在的上传文件（不存在返回空字符串）
func (s *Store) FindUploadByHash(hash string) (string, error) {
	var filename string
	err := s.db.QueryRow(`SELECT filename FROM uploads_meta WHERE hash = ?`, hash).Scan(&filename)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return filename, nil
}

// RegisterUpload 记录上传文件元数据（重复 hash 自动忽略）
func (s *Store) RegisterUpload(hash, filename string, size int64) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO uploads_meta (hash, filename, size) VALUES (?, ?, ?)`,
		hash, filename, size,
	)
	return err
}

// ListUploads 列出全部上传文件元数据
func (s *Store) ListUploads() ([]*model.UploadMeta, error) {
	rows, err := s.db.Query(`SELECT hash, filename, size, created_at FROM uploads_meta ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.UploadMeta
	for rows.Next() {
		var m model.UploadMeta
		if err := rows.Scan(&m.Hash, &m.Filename, &m.Size, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

// DeleteUploadMeta 删除上传文件元数据记录
func (s *Store) DeleteUploadMeta(filename string) error {
	_, err := s.db.Exec(`DELETE FROM uploads_meta WHERE filename = ?`, filename)
	return err
}
