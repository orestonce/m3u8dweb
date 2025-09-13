package backend

import (
	"github.com/orestonce/m3u8d"
	"log"
	"m3u8dweb/models"
	"os"
	"time"
)

func RunBackendDownloader() {
	var env = &m3u8d.DownloadEnv{}
	var id string // 正在下载的任务的id

	for {
		time.Sleep(time.Second)

		var envStatus m3u8d.GetStatus_Resp
		if id != "" {
			envStatus = env.GetStatus()
			task, ok := models.GetTaskById(id)

			now := time.Now()
			if ok == false || task.Status == models.StatusPaused {
				env.CloseEnv() // 前端已经删除任务、暂停任务，后端也停止任务
				id = ""
			} else if envStatus.IsDownloading { // 后端正在下载
				task.Status = models.StatusDownloading
				task.Progress = envStatus.Percent
				task.StatusBar = envStatus.Title + " " + envStatus.StatusBar
				task.ErrMsg = ""
			} else { // 没下载：要么顺利完成，要么出错了
				if envStatus.ErrMsg == "" {
					task.Status = models.StatusCompleted
					task.Progress = 100
					task.CompletedAt = now
					fi, _ := os.Stat(envStatus.SaveFileTo)
					task.Size = fi.Size()
					task.StatusBar = ""
				} else {
					task.Status = models.StatusFailed
					task.ErrMsg = envStatus.ErrMsg
				}
				env.CloseEnv()
				id = ""
			}
			task.UpdatedAt = now
			models.UpdateTask(task)
			continue
		}

		allTask, _ := models.GetAllTasks()
		for _, task := range allTask {
			if task.Status != models.StatusWaiting {
				continue
			}
			if task.AdvancedSettings == nil {
				task.Status = models.StatusFailed
				task.UpdatedAt = time.Now()
				models.UpdateTask(task)
				log.Println("task.AdvancedSettings == nil", task.ID)
				continue
			}
			errMsg := env.StartDownload(m3u8d.StartDownload_Req{
				M3u8Url:           task.URL,
				Insecure:          task.AdvancedSettings.AllowInsecureHTTPS,
				SaveDir:           task.AdvancedSettings.SaveLocation,
				FileName:          task.Filename,
				SkipTsExpr:        task.AdvancedSettings.SkipTSInfo,
				SetProxy:          task.AdvancedSettings.GetProxyString(),
				HeaderMap:         task.HeaderMap,
				SkipRemoveTs:      task.AdvancedSettings.KeepTSFiles,
				ProgressBarShow:   false,
				ThreadCount:       task.AdvancedSettings.DownloadThreads,
				SkipCacheCheck:    false,
				SkipMergeTs:       task.AdvancedSettings.NoMergeTS,
				DebugLog:          task.AdvancedSettings.DebugLog,
				TsTempDir:         task.AdvancedSettings.TempDirectory,
				UseServerSideTime: task.AdvancedSettings.UseServerFileTime,
				WithSkipLog:       task.AdvancedSettings.LogSkippedTS,
			})
			if errMsg != "" {
				task.Status = models.StatusFailed
				task.UpdatedAt = time.Now()
				models.UpdateTask(task)
			}
			id = task.ID
			task.Status = models.StatusDownloading
			task.UpdatedAt = time.Now()
			models.UpdateTask(task)
			break
		}
	}
}
