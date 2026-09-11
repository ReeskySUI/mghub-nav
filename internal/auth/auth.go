package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"mghub-portal/internal/model"
	"mghub-portal/internal/store"
)

// Manager 认证管理器
type Manager struct {
	store       *store.Store
	sessions    map[string]*model.Session
	sessionLock sync.RWMutex
	maxAge      time.Duration
	maxAttempts int
	lockoutTime time.Duration
}

// New 创建认证管理器
func New(s *store.Store, sessionMaxAgeHours, maxAttempts, lockoutMinutes int) *Manager {
	m := &Manager{
		store:       s,
		sessions:    make(map[string]*model.Session),
		maxAge:      time.Duration(sessionMaxAgeHours) * time.Hour,
		maxAttempts: maxAttempts,
		lockoutTime: time.Duration(lockoutMinutes) * time.Minute,
	}
	// 启动会话清理协程
	go m.cleanupSessions()
	return m
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验密码
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Login 登录验证，返回 session token
func (m *Manager) Login(username, password string) (string, error) {
	user, err := m.store.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 检查账号状态
	if user.Status == model.UserDisabled {
		return "", errors.New("账号已被禁用")
	}

	// 检查是否被锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remaining := time.Until(*user.LockedUntil).Minutes()
		return "", errors.New("登录失败次数过多，请 " + formatMinutes(remaining) + " 后再试")
	}

	// 校验密码
	if !CheckPassword(password, user.PasswordHash) {
		// 登录失败，增加计数
		attempts := user.LoginAttempts + 1
		var lockedUntil *time.Time
		if attempts >= m.maxAttempts {
			t := time.Now().Add(m.lockoutTime)
			lockedUntil = &t
			attempts = 0
		}
		_ = m.store.UpdateLoginAttempts(user.ID, attempts, lockedUntil)
		if lockedUntil != nil {
			return "", errors.New("登录失败次数过多，请 " + formatMinutes(m.lockoutTime.Minutes()) + " 后再试")
		}
		return "", errors.New("用户名或密码错误")
	}

	// 登录成功
	_ = m.store.RecordLogin(user.ID)

	// 创建 session
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	session := &model.Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(m.maxAge),
	}
	m.sessionLock.Lock()
	m.sessions[token] = session
	m.sessionLock.Unlock()

	return token, nil
}

// Logout 登出
func (m *Manager) Logout(token string) {
	m.sessionLock.Lock()
	delete(m.sessions, token)
	m.sessionLock.Unlock()
}

// GetSession 获取会话（同时检查过期）
func (m *Manager) GetSession(token string) *model.Session {
	m.sessionLock.RLock()
	session, ok := m.sessions[token]
	m.sessionLock.RUnlock()
	if !ok {
		return nil
	}
	if time.Now().After(session.ExpiresAt) {
		m.sessionLock.Lock()
		delete(m.sessions, token)
		m.sessionLock.Unlock()
		return nil
	}
	return session
}

// cleanupSessions 定期清理过期会话
func (m *Manager) cleanupSessions() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.sessionLock.Lock()
		for token, s := range m.sessions {
			if now.After(s.ExpiresAt) {
				delete(m.sessions, token)
			}
		}
		m.sessionLock.Unlock()
	}
}

// generateToken 生成随机 token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// formatMinutes 格式化分钟数
func formatMinutes(minutes float64) string {
	if minutes < 1 {
		return "不到1分钟"
	}
	return formatInt(int(minutes)) + " 分钟"
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
