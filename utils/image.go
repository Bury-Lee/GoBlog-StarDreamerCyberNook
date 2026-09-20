package utils

import (
	"StarDreamerCyberNook/global"
	"path/filepath"

	"encoding/base64"
	"io"
	"os"
	"strings"
)

// encodeImageToBase64 将图片文件编码为base64字符串
func EncodeImageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 5*1024*1024) // 5MB buffer
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(buf[:n])
	return encoded, nil
}

// 假设 global.Config.Upload.WhiteList 是 []string 类型
func ImageSuffixJudge(filename string) (string, bool) {
	// 1. 获取扩展名，例如 ".jpg", ".png"
	// filepath.Ext 会自动处理路径问题，如 "C:/path/to/img.jpg" -> ".jpg"
	ext := filepath.Ext(filename)

	// 2. 如果没有扩展名，或者扩展名只是一个点 (例如文件名为 ".")
	if ext == "" || ext == "." {
		return "", false
	}

	suffix := strings.ToLower(ext[1:])

	// 4. 检查白名单
	// 建议确保白名单里的内容也是小写的，或者在这里做不区分大小写的比较
	if _, ok := global.Config.Upload.WhiteList[suffix]; !ok {
		return suffix, false
	}
	return suffix, true
}

func GetContentType(suffix string) string {
	switch strings.ToLower(suffix) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
