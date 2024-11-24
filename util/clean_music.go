package util

import (
	"fmt"
	"github.com/dhowden/tag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var IgnoreKeywords = []string{
	"live", "cover", "ktv", "FM", "AKA", "专属", "薇澜", "纯音乐", "赵鹏", "获奖感言", "混音",
	"MZ", "dj", "3D", "remix", "instrument", "绯绯Feifei", "铃声", "伴奏", "改编",
	"翻奏", "现场", "未知歌手", "音悦汇", "灰色轨迹2002", "网络歌手", "原唱", "遛狗天才",
	"淘漉音乐", "沐夕MuXi", "沉默的曾大炮", "民谣老潘", "演奏曲", "解忧音乐厅", "键盘", "闯王",
	"词曲", "翻唱", "工作室", "网友", "演奏", "精选集", "全球音乐吧", "刘京小提琴", "弹唱",
	"抖音", "一坨坨坨子", "yayun811224", "NJ浩瀚", "krav8888888", "Kecs1", "前奏", "1056声音日记",
	"慧缘", "演唱会", "许巍", "车载",
}

// CleanupMusicFiles 清理包含指定关键字的音乐文件，并在特定情况下删除文件夹
func CleanupMusicFiles(baseDir string, keywords []string) error {
	var directoriesToDelete []string // 存储需要删除的目录
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// 检查文件夹名称是否包含关键字
		if d.IsDir() {
			if d.Name() == "@eaDir" {
				return filepath.SkipDir
			}

			isEmpty, err := isEmptyDir(path)
			if err != nil {
				return fmt.Errorf("error checking if directory is empty: %w", err)
			}

			// 如果是空目录，记录路径以便稍后删除
			if isEmpty {
				fmt.Printf("Marking empty directory for deletion: %s\n", path)
				directoriesToDelete = append(directoriesToDelete, path)
				return nil
			}
			return nil
		}

		if !isMusicFile(d.Name()) {
			return nil
		}

		name := strings.ReplaceAll(d.Name(), filepath.Ext(d.Name()), "")
		err = removeFileIfNotMatch(path, name)
		if err != nil {
			return err
		}

		// 如果是文件，检查是否包含关键字
		if containsKeywords(path, keywords) {
			fmt.Printf("Deleting file: %s\n", path)
			return os.Remove(path)
		}

		return nil
	})

	// 遍历完成后，删除标记的目录，并递归删除其父目录（如果为空）
	for _, dir := range directoriesToDelete {
		if err := recursiveDeleteEmptyDirs(dir); err != nil {
			fmt.Printf("Error deleting directory %s: %v\n", dir, err)
		}
	}

	return err
}

func recursiveDeleteEmptyDirs(dir string) error {
	isEmpty, err := isEmptyDir(dir)
	if err != nil {
		fmt.Printf("Error checking if directory is empty: %v\n", err)
		return err
	}

	if !isEmpty {
		fmt.Printf("Directory is not empty: %s\n", dir)
		return nil
	}

	// 删除当前目录
	fmt.Printf("Deleting empty directory: %s\n", dir)
	if err := os.RemoveAll(dir); err != nil {
		fmt.Printf("Error deleting directory %s: %v\n", dir, err)
		return err
	}

	// 获取父目录路径
	parentDir := filepath.Dir(dir)
	if parentDir == "." || parentDir == "/" {
		fmt.Printf("Reached root or current directory, stopping: %s\n", parentDir)
		return nil
	}

	// 递归检查父目录
	return recursiveDeleteEmptyDirs(parentDir)
}

// isEmptyDir 检查目录是否为空（忽略隐藏文件）
func isEmptyDir(dir string) (bool, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	if len(files) == 2 && containsFile(files, "cover.jpg") && containsFile(files, "@eaDir") {
		return true, nil
	}

	for _, file := range files {
		// 忽略隐藏文件和系统文件（如 .DS_Store 或 Thumbs.db）
		if !strings.HasPrefix(file.Name(), ".") && file.Name() != "Thumbs.db" {
			return false, nil
		}
	}
	return true, nil
}

// containsKeywords 检查文件或文件夹名称是否包含指定关键字
func containsKeywords(name string, keywords []string) bool {
	lowerName := strings.ToLower(name)
	for _, keyword := range keywords {
		if strings.Contains(lowerName, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// containsFile 检查目录中是否包含指定文件
func containsFile(files []os.DirEntry, target string) bool {
	for _, file := range files {
		if file.Name() == target {
			return true
		}
	}
	return false
}

// containsSingleMusicFile 检查目录中是否仅包含一个音乐文件
func containsSingleMusicFile(files []os.DirEntry) bool {
	musicCount := 0
	for _, file := range files {
		if !file.IsDir() && isMusicFile(file.Name()) {
			musicCount++
		}
	}
	return musicCount == 1
}

// isMusicFile 判断文件是否为音乐文件
func isMusicFile(filename string) bool {
	extensions := []string{".mp3", ".flac", ".wav", ".aac", ".m4a"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

func getMusicFileMeta(filePath string) (tag.Metadata, error) {
	// 打开音频文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	// 读取音频文件标签
	tagData, err := tag.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("error reading tags from file %s: %w", filePath, err)
	}
	return tagData, nil
}

// 比较文件名和音乐文件元数据中的标题
func isMusicFileNameMatch(filePath, musicName string) (bool, error) {
	// 读取音频文件标签
	tagData, err := getMusicFileMeta(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading tags from file %s: %w", filePath, err)
	}

	// 获取音乐文件中的标题信息
	trackTitle := tagData.Title()

	// 比较文件名和音乐文件中的标题
	if strings.EqualFold(trackTitle, musicName) {
		return true, nil
	}

	return false, nil
}

// 删除不匹配的文件
func removeFileIfNotMatch(filePath, musicName string) error {
	match, err := isMusicFileNameMatch(filePath, musicName)
	if err != nil {
		return err
	}
	if !match {
		fmt.Printf("File '%s' does not match music name, deleting file.\n", filePath)
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("error deleting file %s: %w", filePath, err)
		}
		fmt.Printf("Deleted file: %s\n", filePath)
	}
	return nil
}
