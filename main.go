package main

import (
	"flac/util"
	"fmt"
	"time"
)

func main() {
	keyword := []string{"邓紫棋", "郁可唯", "周杰伦", "朴树", "许巍"}
	page := 1
	size := 30
	unlockCode := "5767"
	baseDir := "Music"
	i := 0

	for {
		// 从第一个接口获取音乐信息
		songs, err := util.FetchMusicInfo(keyword[i], page, size)
		if err != nil {
			fmt.Printf("Error fetching music info: %v\n", err)
			break
		}
		if len(songs) == 0 {
			i++
			page = 1
		}
		if len(songs) == 0 && i == len(keyword) {
			break
		}
		// 遍历每首歌曲并启动协程处理
		for _, s := range songs {
			fmt.Printf("Processing song: %s\n", s.Name)
			if err := util.ProcessSong(s, baseDir, unlockCode); err != nil {
				fmt.Printf("Error processing song '%s': %v\n", s.Name, err)
			} else {
				fmt.Printf("Successfully processed song '%s'\n", s.Name)
			}
		}

		// 当前批次处理完成，进入下一页
		page++
		fmt.Printf("Page %d processed.\n", page)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("All songs processed.")
}
