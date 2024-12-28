package common

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServerLocale string

const (
	Localhost            ServerLocale = "localhost:8080"
	Datacenter_Google_JP ServerLocale = "trading.innoroot.com"
	Datacenter_Hikari_JP ServerLocale = "hikari.lolo.finance"
)

var trustedDomains = []ServerLocale{
	Localhost,
	Datacenter_Google_JP,
	Datacenter_Hikari_JP,
}

func GetHostName(c *gin.Context) (ServerLocale, error) {
	// 嘗試從請求中獲取 X-Forwarded-Host 標頭
	xForwardedHost := c.GetHeader("X-Forwarded-Host")

	var domainName string

	if xForwardedHost != "" {
		// 分割 X-Forwarded-Host 值，取第一個
		hosts := strings.Split(xForwardedHost, ",") // 可能有多個值，以逗號分隔
		domainName = strings.TrimSpace(hosts[0])    // 獲取第一個並移除空白
	} else {
		// 否則使用請求的主機名
		domainName = c.Request.Host
	}

	// 檢查 domain name 是否是可信的值
	isTrusted, rtn := isTrustedDomain(domainName)

	if !isTrusted {
		return "", fmt.Errorf("不受信任的域名:%s", domainName)
	}

	return rtn, nil // 返回獲取的域名
}

func isTrustedDomain(domain string) (bool, ServerLocale) {
	domain = strings.ToLower(domain) // 將輸入的域名轉換為小寫
	for _, trusted := range trustedDomains {
		if strings.ToLower(string(trusted)) == domain { // 將可信域名轉換為小寫進行比較
			return true, trusted
		}
	}
	return false, Localhost
}
