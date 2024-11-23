package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// 判断文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// DownloadFile 下载文件
func DownloadFile(url, filePath string) error {
	if fileExists(filePath) {
		return nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	// 创建文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// 保存到文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// sanitizeFileName 清理文件名中的非法字符
func sanitizeFileName(name string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}

// RemoveEmptyDir 判断文件夹是否为空，并删除它（如果为空）
func RemoveEmptyDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", dirPath, err)
	}
	if len(entries) == 0 {
		err = os.Remove(dirPath)
		if err != nil {
			return fmt.Errorf("error removing empty directory %s: %w", dirPath, err)
		}
		fmt.Printf("Removed empty directory: %s\n", dirPath)
	}
	return nil
}
