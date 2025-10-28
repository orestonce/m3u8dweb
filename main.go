package main

import (
	"embed"
	"flag"
	"log"
	"m3u8dweb/backend"
	"m3u8dweb/config"
	"m3u8dweb/db"
	"m3u8dweb/handlers"
	"net/http"
	"runtime"
)

//go:embed templates/*
var fs embed.FS

// 声明将在编译时注入的变量
var (
	version   = "VERSION_NO"     // 版本号
	buildTime = "BUILD_TIME"     // 编译时间
	commit    = "COMMIT_ID_HASH" // Git提交哈希
)

// basicAuth 中间件实现HTTP基本认证
func basicAuth(username, password string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="m3u8dweb"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("未授权访问\n"))
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func main() {
	handlers.GTemplateFs = &fs

	// 打印版本信息
	log.Printf("m3u8dweb - 版本: %s, 构建时间: %s, 提交: %s, GOOS：%s, GOARCH: %s", version, buildTime, commit, runtime.GOOS, runtime.GOARCH)

	// 定义命令行参数
	listenAddr := flag.String("listen", ":8080", "HTTP服务监听地址")
	dbPath := flag.String("db", "downloads.db", "数据库文件路径")
	authUser := flag.String("auth-user", "", "Basic认证用户名")
	authPass := flag.String("auth-pass", "", "Basic认证密码")

	certFile := flag.String("cert-file", "", "https证书文件")
	keyFile := flag.String("key-file", "", "https私钥文件")

	// 解析命令行参数
	flag.Parse()

	// 初始化数据库
	if err := db.InitDB(*dbPath); err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	defer db.CloseDB()

	// 初始化默认配置
	config.InitDefaultSettings()

	go backend.RunBackendDownloader()
	ch := make(chan []byte, 1)
	go backend.RunFastPushThread(ch)
	go handlers.BroadcastTaskThread(ch)

	// 注册路由，根据是否提供认证信息决定是否启用BasicAuth
	registerHandler := func(path string, handler http.HandlerFunc) {
		if *authUser != "" && *authPass != "" {
			http.HandleFunc(path, basicAuth(*authUser, *authPass, handler))
		} else {
			http.HandleFunc(path, handler)
		}
	}

	// 注册页面路由
	registerHandler("/", handlers.AllTasksHandler)
	registerHandler("/new-download", handlers.NewDownloadHandler)
	registerHandler("/settings", handlers.SettingsHandler)

	// 注册API路由
	registerHandler("/api/tasks", handlers.TaskAPIHandler)
	registerHandler("/api/settings", handlers.SettingsAPIHandler)

	//快速推送
	registerHandler("/ws/progress", handlers.TaskWebSocketHandler)

	log.Printf("服务器启动在 %s 端口", *listenAddr)
	if *certFile != "" && *keyFile != "" {
		log.Fatal(http.ListenAndServeTLS(*listenAddr, *certFile, *keyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(*listenAddr, nil))
	}
}
