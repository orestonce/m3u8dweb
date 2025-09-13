package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"m3u8dweb/db"
	"m3u8dweb/models"
	_ "embed"
)

var GTemplateFs *embed.FS

// 显示新建下载页面
func NewDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取系统设置，用于填充表单默认值
	var settings models.SystemSettings
	if err := db.GetData("settings", "system", &settings); err != nil {
		// 如果没有设置，使用默认值
		settings = *models.DefaultSettings()
	}

	// 渲染模板
	tmpl, err := template.ParseFS(*GTemplateFs, "templates/new-download.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Settings models.SystemSettings
	}{
		Settings: settings,
	}

	tmpl.Execute(w, data)
}
