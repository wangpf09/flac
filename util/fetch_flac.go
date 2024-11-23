package util

import (
	"encoding/json"
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
func ProcessSong(song model.Song, baseDir string, unlockCode string) error {
	// 组合保存路径
	artist := sanitizeFileName(strings.Join(song.Singers, " & "))
	album := sanitizeFileName(song.AlbumName)
	songName := sanitizeFileName(song.Name)
	basePath := filepath.Join(baseDir, artist, album)

	// 创建目录
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", basePath, err)
	}

	// 获取音乐文件下载地址
	fileURL := fmt.Sprintf("https://api.flac.life/url/qq/%s/flac", song.ID)
	downloadURL, err := FetchFileURL(fileURL, unlockCode)
	if err != nil {
		return fmt.Errorf("error fetching file URL: %w", err)
	}

	// 下载音乐文件
	musicFilePath := filepath.Join(basePath, songName+".flac")
	err = DownloadFile(downloadURL, musicFilePath)
	if err != nil {
		return fmt.Errorf("error downloading song file: %w", err)
	}

	// 下载封面图片
	if song.PicURL != "" {
		coverFilePath := filepath.Join(basePath, "cover.jpg")
		err = DownloadFile(song.PicURL, coverFilePath)
		if err != nil {
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
	encodedKeyword := url.QueryEscape(keyword)
	apiURL := fmt.Sprintf("https://api.flac.life/search/qq?keyword=%s&page=%d&size=%d", encodedKeyword, page, size)
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

// FetchFileURL 从第二个接口获取音乐文件的下载地址
func FetchFileURL(uri, unlockCode string) (string, error) {
	// 创建 HTTP 请求
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v\n", err)
	}
	// 添加请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible;)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("unlockcode", unlockCode)
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error making request: %v\n", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP status %d: %s", resp.StatusCode, body)
	}
	var fileURLResponse model.FileURLResponse
	err = json.NewDecoder(resp.Body).Decode(&fileURLResponse)
	if err != nil {
		return "", fmt.Errorf("error decoding JSON: %w", err)
	}
	if !fileURLResponse.Success {
		return "", fmt.Errorf("API responded with an error: %s", fileURLResponse.Message)
	}
	return fileURLResponse.Result, nil
}
