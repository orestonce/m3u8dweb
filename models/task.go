package models

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"m3u8dweb/db"
	"sort"
	"sync"
	"time"
)

// 下载任务状态
const (
	StatusWaiting     = "等待中"
	StatusDownloading = "下载中"
	StatusPaused      = "已暂停"
	StatusCompleted   = "已完成"
	StatusFailed      = "失败"
)

// 下载任务模型
type DownloadTask struct {
	ID               string              `json:"id"`         //front
	URL              string              `json:"url"`        //front
	Filename         string              `json:"filename"`   //front
	Size             int64               `json:"size"`       //front
	Progress         int                 `json:"progress"`   //front
	Status           string              `json:"status"`     //front
	StatusBar        string              `json:"status_bar"` // front
	ErrMsg           string              `json:"err_msg,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CompletedAt      time.Time           `json:"completed_at"`
	HeaderMap        map[string][]string `json:"header_map,omitempty"`
	AdvancedSettings *SystemSettings     `json:"advanced_settings,omitempty"`
}

type TaskBackendInfo struct {
	SavePath          string
	TempPath          string
	Threads           int
	M3u8Url           string
	Insecure          bool                // "是否允许不安全的请求(默认为false)"
	SaveDir           string              // "文件保存路径(默认为当前路径)"
	FileName          string              // 文件名
	SkipTsExpr        string              // 跳过ts信息，ts编号从1开始，可以以逗号","为分隔符跳过多部分ts，例如: 1,92-100 表示跳过第1号ts、跳过92到100号ts
	SetProxy          string              //代理
	HeaderMap         map[string][]string // 自定义http头信息
	SkipRemoveTs      bool                // 不删除ts文件
	ProgressBarShow   bool                // 在控制台打印进度条
	ThreadCount       int                 // 线程数
	SkipCacheCheck    bool                // 不缓存已下载的m3u8的文件信息
	SkipMergeTs       bool                // 不合并ts为mp4
	DebugLog          bool                // 调试日志
	TsTempDir         string              // 临时ts文件目录
	UseServerSideTime bool                // 使用服务端提供的文件时间
	WithSkipLog       bool                // 在mp4旁记录跳过ts文件的信息
}

// 创建新任务
func NewDownloadTask(url, filename string) *DownloadTask {
	// 生成唯一ID
	id, _ := generateID()

	return &DownloadTask{
		ID:        id,
		URL:       url,
		Filename:  filename,
		Status:    StatusWaiting,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 生成唯一ID
func generateID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GetTaskById(id string) (task DownloadTask, ok bool) {
	err := db.GetData("tasks", id, &task)

	if err != nil {
		return task, false
	}
	if task.ID == "" || task.ID != id {
		return task, false
	}
	return task, true
}

var gTaskLocker sync.Mutex

func UpdateTaskV2(id string, cb func(task *DownloadTask) ) {
	gTaskLocker.Lock()
	defer gTaskLocker.Unlock()

	if id == "" {
		return
	}

	var taskInDb DownloadTask
	err := db.GetData("tasks", id, &taskInDb)
	if err != nil {
		log.Println("get task data error", id, err)
		return
	}

	cb(&taskInDb)

	err = db.SaveData("tasks", taskInDb.ID, taskInDb)
	if err != nil {
		log.Println("save task data error", id, err)
	}
}

func DeleteTask(id string) (ok bool) {
	gTaskLocker.Lock()
	defer gTaskLocker.Unlock()

	return db.DeleteData("tasks", id) == nil
}

func GetAllTasks() (tasks []DownloadTask, err error) {
	// 获取所有任务ID
	keys, err := db.GetAllKeys("tasks")
	if err != nil {
		return nil, err
	}

	//加载所有任务
	for _, key := range keys {
		var task DownloadTask
		err = db.GetData("tasks", string(key), &task)

		if err != nil {
			return nil, err
		}
		if task.ID != "" {
			tasks = append(tasks, task)
		}
	}

	scoreMap := map[string]int{
		StatusDownloading: 10,
		StatusWaiting:     9,
		StatusPaused:      8,
		StatusCompleted:   7,
		StatusFailed:      6,
	}

	sort.Slice(tasks, func(i, j int) bool {
		as, bs := scoreMap[tasks[i].Status], scoreMap[tasks[j].Status]
		if as != bs {
			return as > bs
		}
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}
