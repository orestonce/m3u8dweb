package models

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strconv"
)

// 代理类型
const (
	ProxyTypeNone   = "none"
	ProxyTypeHTTP   = "http"
	ProxyTypeHTTPS  = "https" // 新增HTTPS代理类型
	ProxyTypeSOCKS5 = "socks5"
)

// 系统设置模型
type SystemSettings struct {
	SaveLocation       string              `json:"save_location"`
	TempDirectory      string              `json:"temp_directory"`
	DownloadThreads    int                 `json:"download_threads"`
	SkipTSInfo         string              `json:"skip_ts_info"`
	ProxyType          string              `json:"proxy_type"`
	ProxyHost          string              `json:"proxy_host"`
	ProxyPort          int                 `json:"proxy_port"`
	ProxyUsername      string              `json:"proxy_username"`
	ProxyPassword      string              `json:"proxy_password"`
	AllowInsecureHTTPS bool                `json:"allow_insecure_https"`
	KeepTSFiles        bool                `json:"keep_ts_files"`
	UseServerFileTime  bool                `json:"use_server_file_time"`
	NoMergeTS          bool                `json:"no_merge_ts"`
	DebugLog           bool                `json:"debug_log"`
	LogSkippedTS       bool                `json:"log_skipped_ts"`
}

// 创建默认设置
func DefaultSettings() *SystemSettings {
	ep, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(ep)

	return &SystemSettings{
		SaveLocation:    dir,
		TempDirectory:   dir,
		DownloadThreads: 8,
		ProxyType:       ProxyTypeNone,
	}
}

func (obj SystemSettings) GetProxyString() string {
	//if obj.ProxyPort == 0 {
	//	if obj.ProxyType == ProxyTypeHTTP {
	//		obj.ProxyPort = 80
	//	} else {
	//		obj.ProxyPort = 443
	//	}
	//}
	//http://[用户名]:[密码]@[代理服务器地址]:[代理端口]
	//总之，最标准的拼接格式是 http://user:pass@host:port，但务必注意对特殊字符进行编码，并意识到其潜在的安全风险。
	buf := bytes.NewBuffer(nil)
	switch obj.ProxyType {
	case ProxyTypeHTTP:
		buf.WriteString("http://")
	case ProxyTypeHTTPS:
		buf.WriteString("https://")
	case ProxyTypeSOCKS5:
		buf.WriteString("socks5://")
	default:
		return ""
	}
	if obj.ProxyUsername != "" || obj.ProxyPassword != "" {
		if obj.ProxyUsername != "" {
			un := base64.URLEncoding.EncodeToString([]byte(obj.ProxyUsername))
			buf.WriteString(un)
		}
		if obj.ProxyPassword != "" {
			pw := base64.URLEncoding.EncodeToString([]byte(obj.ProxyPassword))
			buf.WriteString(":")
			buf.WriteString(pw)
		}
		buf.WriteString("@")
	}
	buf.WriteString(obj.ProxyHost)
	if obj.ProxyPort > 0 {
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(obj.ProxyPort))
	}
	return buf.String()
}
