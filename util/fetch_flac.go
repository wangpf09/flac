package util

import (
	"encoding/json"
	"flac/config"
	"flac/model"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProcessSong 处理单首歌曲的下载和保存
func ProcessSong(currentSinger string, song model.Song, baseDir string, unlockCode string) error {
	// 组合保存路径
	artist := sanitizeFileName(strings.Join(song.Singers, " & "))
	album := sanitizeFileName(song.AlbumName)
	songName := sanitizeFileName(song.Name)
	basePath := filepath.Join(baseDir, artist, album)

	// 下载音乐文件
	musicFilePath := filepath.Join(basePath, songName+".flac")
	if fileExists(musicFilePath) {
		return nil
	}

	// 创建目录
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", basePath, err)
	}

	// 获取音乐文件下载地址
	downloadURL, quality, err := GetMusicDownloadURL(song, unlockCode)
	if err != nil {
		return fmt.Errorf("error fetching file URL: %w", err)
	}
	if quality != "flac" {
		musicFilePath = filepath.Join(basePath, songName+".mp3")
	}

	err = DownloadFile(downloadURL, musicFilePath)
	if err != nil {
		// 检查并删除空目录
		err = RemoveEmptyDir(basePath)
		if err != nil {
			fmt.Printf("Error removing empty directory %s: %v\n", baseDir, err)
		}
		return fmt.Errorf("error downloading song file: %w", err)
	}

	meta, err := getMusicFileMeta(musicFilePath)
	if err != nil {
		err := os.Remove(musicFilePath)
		if err != nil {
			return fmt.Errorf("error deleting file %s: %w", musicFilePath, err)
		}
		return fmt.Errorf("error getting music file metadata: %w", err)
	}

	errorInfo := make(map[string]string)
	force := false
	if meta.Album() != song.AlbumName {
		errorInfo["album_name"] = meta.Album()
	}

	if meta.Title() != song.Name {
		errorInfo["title"] = meta.Title()
	}

	if !containsKeywords(meta.Artist(), song.Singers) {
		errorInfo["artist"] = meta.Artist()
	}

	if !strings.Contains(meta.Artist(), currentSinger) {
		errorInfo["cartist"] = meta.Artist()
		force = true
	}

	if containsKeywords(musicFilePath, IgnoreKeywords) {
		errorInfo["keywords"] = musicFilePath
		force = true
	}

	if len(errorInfo) > 3 || force {
		err := os.Remove(musicFilePath)
		if err != nil {
			return fmt.Errorf("error deleting file %s: %w", musicFilePath, err)
		}
		for k, v := range errorInfo {
			fmt.Printf("deleting meta not match file %s %s: %s\n", musicFilePath, k, v)
		}
		fmt.Printf("album %s name %s singers %s\n", song.AlbumName, song.Name, song.Singers)
	}

	// 下载封面图片
	if song.PicURL != "" {
		coverFilePath := filepath.Join(basePath, "cover.jpg")
		err = DownloadFile(song.PicURL, coverFilePath)
		if err != nil {
			// 检查并删除空目录
			err = RemoveEmptyDir(basePath)
			if err != nil {
				fmt.Printf("Error removing empty directory %s: %v\n", baseDir, err)
			}
			return fmt.Errorf("error downloading cover image: %w", err)
		}
	}
	fmt.Printf("Successfully processed song '%s'\n", musicFilePath)
	time.Sleep(3 * time.Second)
	return nil
}

// FetchMusicInfo 从第一个接口获取音乐信息
func FetchMusicInfo(keyword string, page, size int) ([]model.Song, error) {
	// 构造请求 URL
	flacInfo := config.GetAppConfig().FlacInfo
	encodedKeyword := url.QueryEscape(keyword)
	apiURL := fmt.Sprintf("%s/%s?keyword=%s&page=%d&size=%d", flacInfo.Baseurl, flacInfo.SearchApi,
		encodedKeyword, page, size)

	// 创建 HTTP 请求
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v\n", err)
	}
	// 添加请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible;)")
	req.Header.Set("Accept", "application/json")
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %v\n", err)
	}
	defer resp.Body.Close()
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Request failed. Status: %d, Body: %s\n", resp.StatusCode, body)
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}
	var apiResponse model.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}
	if !apiResponse.Success {
		return nil, fmt.Errorf("API responded with an error: %s", apiResponse.Message)
	}
	return apiResponse.Result.List, nil
}

// RetryWithExponentialBackoff 封装重试和指数退避逻辑
func RetryWithExponentialBackoff(attempts int, baseDelay time.Duration, operation func() error) error {
	for attempt := 0; attempt < attempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		// 如果是最后一次尝试，直接返回错误
		if attempt == attempts-1 {
			return fmt.Errorf("operation failed after %d attempts: %w", attempts, err)
		}

		// 计算退避时间
		delay := baseDelay * time.Duration(1<<attempt) // 指数退避
		fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt+1, err, delay)
		time.Sleep(delay)
	}
	return nil // 理论上不会到达这里
}

// FetchFileURL 从第二个接口获取音乐文件的下载地址
func FetchFileURL(uri, unlockCode string) (string, error) {
	var result string

	// 定义操作逻辑
	operation := func() error {
		client := &http.Client{}
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}

		// 添加请求头
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible;)")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("unlockcode", unlockCode)

		// 发起请求
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		// 检查 HTTP 响应状态
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, body)
		}

		// 解析响应体
		var fileURLResponse model.FileURLResponse
		err = json.NewDecoder(resp.Body).Decode(&fileURLResponse)
		if err != nil {
			return fmt.Errorf("error decoding JSON: %w", err)
		}

		// 检查业务逻辑成功状态
		if !fileURLResponse.Success {
			return fmt.Errorf("%s API responded with an error: %s", uri, fileURLResponse.Message)
		}

		// 设置结果
		result = fileURLResponse.Result
		return nil
	}

	// 使用封装的重试逻辑
	err := RetryWithExponentialBackoff(3, 2*time.Second, operation)
	if err != nil {
		return "", err
	}

	return result, nil
}

// GetMusicDownloadURL 获取音乐文件下载地址，支持回退机制
func GetMusicDownloadURL(song model.Song, unlockCode string) (string, string, error) {
	flacInfo := config.GetAppConfig().FlacInfo
	for _, quality := range flacInfo.Quality {
		fileURL := fmt.Sprintf("%s/%s/%s/%s", flacInfo.Baseurl, flacInfo.UrlQq, song.ID, quality)
		downloadURL, err := FetchFileURL(fileURL, unlockCode)
		if err == nil && downloadURL != "" {
			fmt.Printf("Successfully fetched %s URL for song '%s': %s\n", quality, song.Name, downloadURL)
			return downloadURL, quality, nil
		}
		fmt.Printf("error fetching file URL for song '%s' quality: %s\n", song.Name, quality)
	}
	return "", "", fmt.Errorf("error fetching file URL for song '%s'", song.Name)
}
