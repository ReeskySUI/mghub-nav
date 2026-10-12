package handler

import (
	"net"
	"strings"
)

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


// resolveNavURL 解析导航页地址。
// 优先使用后台配置的有效 navURL；若未配置或仍是 example.com 占位值，
// 则根据当前请求 Host 推导：example.com -> https://nav.example.com
// 本地 IP / localhost 无法推导时回退到配置值或 "/"
func resolveNavURL(configured, requestHost string) string {
	valid := configured != "" && !strings.Contains(configured, "example.com")
	if valid {
		return configured
	}
	host := requestHost
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if host != "" && host != "localhost" && host != "127.0.0.1" && !strings.HasPrefix(host, "nav.") {
		return "https://nav." + host
	}
	if configured != "" {
		return configured
	}
	return "/"
}


// cookieDomain 推导 session cookie 的 domain，使首页域与导航页域共享登录态。
// nav.mghub.top -> mghub.top；mghub.top -> mghub.top；
// localhost / 127.0.0.1 等本地地址返回空串（IP 无法设置父域 cookie）。
func cookieDomain(requestHost string) string {
	host := requestHost
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if host == "" || host == "localhost" || net.ParseIP(host) != nil {
		return ""
	}
	if strings.HasPrefix(host, "nav.") {
		return host[4:]
	}
	return host
}
