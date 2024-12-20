package main

import (
	"flac/config"
	"flac/model"
	"flac/util"
	"flag"
	"fmt"
	"sync"
	"time"
)

func main() {
	config.InitConfig()
	m := flag.Int("m", 0, "model: 0 fetch music; 1 clean data;")
	flag.Parse()

	switch *m {
	case 0:
		fetchAllMusic()
	case 1:
		cleanDirtyData()
	}
}

func fetchAllMusic() {
	flacInfo := config.GetAppConfig().FlacInfo
	page := 1
	i := 0
	baseDir := config.GetAppConfig().SavePath

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
					if err := util.ProcessSong(flacInfo.Keywords[i], s, baseDir, flacInfo.UnlockCode); err != nil {
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

func cleanDirtyData() {
	baseDir := config.GetAppConfig().SavePath
	for i := 0; i < 3; i++ {
		fmt.Printf("clean base dir: %s\n", baseDir)
		err := util.CleanupMusicFiles(baseDir, util.IgnoreKeywords)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("All dirty data cleaned.")
}
