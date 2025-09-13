package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"m3u8dweb/db"
	"m3u8dweb/models"
	"net/http"
	"strconv"
	"time"
)

// 显示所有任务
func AllTasksHandler(w http.ResponseWriter, r *http.Request) {
	// 加载所有任务
	tasks, err := models.GetAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 渲染模板
	tmpl, err := template.ParseFS(*GTemplateFs,"templates/all-tasks.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TasksJSON []models.DownloadTask
	}{
		TasksJSON: tasks,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

// 处理任务API请求
func TaskAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createTask(w, r)
	case http.MethodGet:
		getTasks(w, r)
	case http.MethodPut:
		updateTask(w, r)
	case http.MethodDelete:
		deleteTask(w, r)
	default:
		http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
	}
}

// 创建新任务
func createTask(w http.ResponseWriter, r *http.Request) {
	var formData struct {
		URL            string              `json:"url"`
		Filename       string              `json:"filename"`
		UseHttpHeaders bool                `json:"use_http_headers"`
		HttpHeaders    map[string][]string `json:"http_headers"`
		Advanced       bool                `json:"advanced"`
		models.SystemSettings
	}

	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		http.Error(w, "解析请求失败", http.StatusBadRequest)
		return
	}

	// 创建新任务
	task := models.NewDownloadTask(formData.URL, formData.Filename)

	if formData.UseHttpHeaders {
		task.HeaderMap = formData.HttpHeaders
	}

	// 如果没有高级设置，系统设置
	if formData.Advanced == false {
		// 获取保存路径（从设置中获取）
		var settings models.SystemSettings
		db.GetData("settings", "system", &settings)
		if settings.SaveLocation == "" {
			settings = *models.DefaultSettings()
		}
		formData.SystemSettings = settings
	}
	task.AdvancedSettings = &formData.SystemSettings

	// 保存任务到数据库
	if err := db.SaveData("tasks", task.ID, task); err != nil {
		http.Error(w, "保存任务失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "id": task.ID})
}

// 获取所有任务
func getTasks(w http.ResponseWriter, r *http.Request) {
	tasks, _ := models.GetAllTasks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// 更新任务
func updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "解析请求失败", http.StatusBadRequest)
		return
	}

	// 获取现有任务
	task, ok := models.GetTaskById(id)
	if ok == false {
		http.Error(w, "任务不存在", http.StatusNotFound)
		return
	}

	// 应用更新
	if status, ok := updates["status"].(string); ok {
		task.Status = status
	}

	// 保存更新
	if ok = models.UpdateTask(task); ok == false {
		http.Error(w, "更新任务失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// 删除任务
func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	if ok := models.DeleteTask(id); ok == false {
		http.Error(w, "删除任务失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// 辅助函数：格式化文件大小
func formatSize(bytes int64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= TB:
		return strconv.FormatFloat(float64(bytes)/TB, 'f', 2, 64) + " TB"
	case bytes >= GB:
		return strconv.FormatFloat(float64(bytes)/GB, 'f', 2, 64) + " GB"
	case bytes >= MB:
		return strconv.FormatFloat(float64(bytes)/MB, 'f', 2, 64) + " MB"
	case bytes >= KB:
		return strconv.FormatFloat(float64(bytes)/KB, 'f', 2, 64) + " KB"
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}

// 辅助函数：根据状态返回CSS类
func statusClass(status string) string {
	switch status {
	case models.StatusDownloading:
		return "bg-primary"
	case models.StatusCompleted:
		return "bg-success"
	case models.StatusFailed:
		return "bg-danger"
	case models.StatusPaused:
		return "bg-warning"
	default: // 等待中
		return "bg-secondary"
	}
}

// 辅助函数：格式化时间
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}
