package main

import (
	"flac/config"
	"flac/model"
	"flac/util"
	"fmt"
	"sync"
	"time"
)

func main() {
	config.InitConfig()
	flacInfo := config.GetAppConfig().FlacInfo
	page := 1
	i := 0
	baseDir := "Music"

	for {
		// 从第一个接口获取音乐信息
		songs, err := util.FetchMusicInfo(flacInfo.Keywords[i], page, flacInfo.PageSize)
		if err != nil {
			fmt.Printf("Error fetching %s music info: %v\n", flacInfo.Keywords[i], err)
			break
		}

		if len(songs) == 0 {
			i++
			page = 1
			if i == len(flacInfo.Keywords) {
				break
			}
			continue
		}

		// 用于同步等待所有协程完成
		var wg sync.WaitGroup
		// 任务通道
		taskChan := make(chan model.Song, len(songs))
		// 启动 worker 协程
		for w := 0; w < config.GetAppConfig().WorkerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// 遍历每首歌曲并启动协程处理
				for s := range taskChan {
					fmt.Printf("Processing song: %s\n", s.Name)
					if err := util.ProcessSong(s, baseDir, flacInfo.UnlockCode); err != nil {
						time.Sleep(15 * time.Second)
						fmt.Printf("Error processing song '%s': %v\n", s.Name, err)
					} else {
						fmt.Printf("Successfully processed song '%s'\n", s.Name)
					}
				}
			}()
		}

		// 将任务发送到任务通道
		for _, s := range songs {
			taskChan <- s
		}
		close(taskChan)

		// 等待所有协程完成
		wg.Wait()

		// 当前批次处理完成，进入下一页
		page++
		fmt.Printf("Page %d processed.\n", page)
		time.Sleep(5 * time.Second)
	}
	fmt.Println("All songs processed.")
}
