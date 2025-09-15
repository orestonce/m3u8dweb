package backend

import (
	"bytes"
	"encoding/json"
	"time"
)

type FastPushData struct {
	ID        string `json:"id,omitempty"`
	StatusBar string `json:"status_bar,omitempty"`
	Progress  int    `json:"progress,omitempty"`
}

func RunFastPushThread(ch chan <- []byte) {
	var lastPushBs []byte
	ticker := time.NewTicker(time.Second * 5)

	for {
		time.Sleep(time.Millisecond * 100)

		// 以下逻辑只读写内存，100毫秒执行一次问题不大
		envStatus := env.GetStatus()

		var data FastPushData
		data.ID = getId()
		if envStatus.IsDownloading {
			data.StatusBar = envStatus.Title + " " + envStatus.StatusBar
		}
		data.Progress = envStatus.Percent


		bs, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		select {
		case <- ticker.C:	//5秒时间到了，不管有没有变更，都推送一下，方便websocket保活
		default:
			if bytes.Equal(bs, lastPushBs) {	// 数据一样，不推了
				continue
			}
		}
		ch <- bs	// 写给推送线程对应的channel
		lastPushBs = bs
	}
}
