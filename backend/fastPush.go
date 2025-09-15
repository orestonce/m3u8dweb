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
		case <- ticker.C:
		default:
			if bytes.Equal(bs, lastPushBs) {
				continue
			}
		}
		ch <- bs
		lastPushBs = bs
	}
}
