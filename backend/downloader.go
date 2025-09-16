package backend

import (
	"github.com/orestonce/m3u8d"
	"m3u8dweb/models"
	"os"
	"time"
)

var env = &m3u8d.DownloadEnv{}

func RunBackendDownloader() {
	for {
		time.Sleep(time.Second)

		envStatus := env.GetStatus()
		cleanEnv := false
		foundTask := false

		models.UpdateTaskV2(envStatus.TaskId, func(task *models.DownloadTask) {
			foundTask = true
			if task.Status != models.StatusDownloading {
				cleanEnv = true
				return
			}

			now := time.Now()
			if envStatus.IsDownloading { // 后端正在下载
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
				cleanEnv = true
			}
			task.UpdatedAt = now
		})
		if cleanEnv || foundTask == false { // 后端停止任务，别再下载了
			env.CloseEnv()
			lookupNextTask()
		}
	}
}

func lookupNextTask() {
	allTask, _ := models.GetAllTasks()

	for _, roTask := range allTask {
		if roTask.Status != models.StatusWaiting && roTask.Status != models.StatusDownloading {
			continue
		}
		if roTask.AdvancedSettings == nil {
			models.UpdateTaskV2(roTask.ID, func(task *models.DownloadTask) {
				task.Status = models.StatusFailed
				task.UpdatedAt = time.Now()
				task.ErrMsg = "task.AdvancedSettings == nil"
			})
			continue
		}
		env.StartDownload(m3u8d.StartDownload_Req{
			M3u8Url:           roTask.URL,
			Insecure:          roTask.AdvancedSettings.AllowInsecureHTTPS,
			SaveDir:           roTask.AdvancedSettings.SaveLocation,
			FileName:          roTask.Filename,
			SkipTsExpr:        roTask.AdvancedSettings.SkipTSInfo,
			SetProxy:          roTask.AdvancedSettings.GetProxyString(),
			HeaderMap:         roTask.HeaderMap,
			SkipRemoveTs:      roTask.AdvancedSettings.KeepTSFiles,
			ProgressBarShow:   false,
			ThreadCount:       roTask.AdvancedSettings.DownloadThreads,
			SkipCacheCheck:    false,
			SkipMergeTs:       roTask.AdvancedSettings.NoMergeTS,
			DebugLog:          roTask.AdvancedSettings.DebugLog,
			TsTempDir:         roTask.AdvancedSettings.TempDirectory,
			UseServerSideTime: roTask.AdvancedSettings.UseServerFileTime,
			WithSkipLog:       roTask.AdvancedSettings.LogSkippedTS,
			TaskId:            roTask.ID,
		})
		models.UpdateTaskV2(roTask.ID, func(task *models.DownloadTask) {
			task.Status = models.StatusDownloading
			task.UpdatedAt = time.Now()
		})
		return
	}
}
