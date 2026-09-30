package handler

import "strings"

// resolveHomeURL 解析首页地址。
// 优先使用后台配置的有效 homeURL；若未配置或仍是 example.com 占位值，
// 则根据当前请求 Host 推导：nav.example.com -> https://example.com
// 本地 IP / localhost 无法推导时回退到配置值或 "/"
func resolveHomeURL(configured, requestHost string) string {
	valid := configured != "" && !strings.Contains(configured, "example.com")
	if valid {
		return configured
	}
	host := requestHost
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if strings.HasPrefix(host, "nav.") {
		return "https://" + host[4:]
	}
	if configured != "" {
		return configured
	}
	return "/"
}
