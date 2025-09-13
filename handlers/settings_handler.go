package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"m3u8dweb/db"
	"m3u8dweb/models"
)

// 显示系统设置页面
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	var settings models.SystemSettings
	if err := db.GetData("settings", "system", &settings); err != nil {
		// 使用默认设置
		settings = *models.DefaultSettings()
	}

	tmpl, err := template.ParseFS(*GTemplateFs,"templates/settings.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, settings)
}

// 处理设置API请求
func SettingsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getSettings(w, r)
	} else if r.Method == http.MethodPost {
		saveSettings(w, r)
	} else {
		http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
	}
}

// 获取设置
func getSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SystemSettings
	if err := db.GetData("settings", "system", &settings); err != nil {
		settings = *models.DefaultSettings()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// 保存设置
func saveSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "解析设置失败", http.StatusBadRequest)
		return
	}

	if err := db.SaveData("settings", "system", settings); err != nil {
		http.Error(w, "保存设置失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
