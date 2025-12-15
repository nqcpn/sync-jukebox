package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// ffprobeOutput 定义了我们关心的 ffprobe JSON 输出结构
type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Tags     struct {
			Title  string `json:"title"`
			Artist string `json:"artist"`
			Album  string `json:"album"`
		} `json:"tags"`
	} `json:"format"`
}

// getAudioMetadata 使用 ffprobe 读取音频文件的元数据
func getAudioMetadata(filePath string) (title, artist, album string, durationMs int, err error) {
	// ffprobe -v quiet -print_format json -show_format "path/to/file"
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		filePath,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return "", "", "", 0, fmt.Errorf("ffprobe error: %v, details: %s", err, stderr.String())
	}

	var ffData ffprobeOutput
	if err = json.Unmarshal(out.Bytes(), &ffData); err != nil {
		return "", "", "", 0, fmt.Errorf("error parsing ffprobe output: %w", err)
	}

	// 解析时长（字符串转为毫秒）
	durationFloat, _ := strconv.ParseFloat(ffData.Format.Duration, 64)
	durationMs = int(durationFloat * 1000)

	// 优先使用元数据中的标题，如果为空，则使用文件名
	title = ffData.Format.Tags.Title
	artist = ffData.Format.Tags.Artist
	album = ffData.Format.Tags.Album

	return title, artist, album, durationMs, nil
}
