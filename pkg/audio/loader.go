package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HiChen85/godub/pkg/converter"
)

// FromFile 从文件加载音频
func FromFile(path string, format ...string) (*AudioSegment, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}

	var audioFormat string
	if len(format) > 0 {
		audioFormat = format[0]
	} else {
		audioFormat = strings.TrimPrefix(filepath.Ext(path), ".")
	}

	// 使用转换器加载音频文件
	audio, err := converter.LoadAudioFile(path, audioFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio file: %w", err)
	}

	return NewAudioSegment(audio.Samples, audio.SampleRate, audio.Channels, audio.BitDepth)
}

// FromMP3 从MP3文件加载音频
func FromMP3(path string) (*AudioSegment, error) {
	return FromFile(path, "mp3")
}

// FromWAV 从WAV文件加载音频
func FromWAV(path string) (*AudioSegment, error) {
	return FromFile(path, "wav")
}

// FromOGG 从OGG文件加载音频
func FromOGG(path string) (*AudioSegment, error) {
	return FromFile(path, "ogg")
}

// FromFLV 从FLV文件加载音频
func FromFLV(path string) (*AudioSegment, error) {
	return FromFile(path, "flv")
}
