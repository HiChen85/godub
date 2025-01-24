package test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/HiChen85/godub/pkg/converter"
)

func TestOggToWav(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "godub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 将 base64 编码的 OGG 数据写入临时文件
	oggData := make([]byte, 0)
	for _, str := range Audio {
		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to decode base64 data: %v", err)
		}
		oggData = append(oggData, data...)
	}

	oggPath := filepath.Join(tempDir, "test.ogg")
	if err := os.WriteFile(oggPath, oggData, 0644); err != nil {
		t.Fatalf("failed to write ogg file: %v", err)
	}

	// 加载 OGG 文件并转换为指定参数
	audio, err := converter.LoadAudioFileWithParams(oggPath, "ogg", 16000, 1, 16)
	if err != nil {
		t.Fatalf("failed to load ogg file: %v", err)
	}

	// 验证音频参数
	if audio.SampleRate != 16000 {
		t.Errorf("expected sample rate 16000, got %d", audio.SampleRate)
	}
	if audio.Channels != 1 {
		t.Errorf("expected 1 channel, got %d", audio.Channels)
	}
	if audio.BitDepth != 16 {
		t.Errorf("expected bit depth 16, got %d", audio.BitDepth)
	}

	// 保存为 WAV 文件
	wavPath := filepath.Join(tempDir, "test.wav")
	if err := converter.SaveAudioFile(audio, wavPath, "wav"); err != nil {
		t.Fatalf("failed to save wav file: %v", err)
	}

	// 验证生成的 WAV 文件
	loadedAudio, err := converter.LoadAudioFile(wavPath, "wav")
	if err != nil {
		t.Fatalf("failed to load wav file: %v", err)
	}

	// 验证转换后的音频参数
	if loadedAudio.SampleRate != audio.SampleRate {
		t.Errorf("wav sample rate %d does not match original %d", loadedAudio.SampleRate, audio.SampleRate)
	}
	if loadedAudio.Channels != audio.Channels {
		t.Errorf("wav channels %d does not match original %d", loadedAudio.Channels, audio.Channels)
	}
	if loadedAudio.BitDepth != audio.BitDepth {
		t.Errorf("wav bit depth %d does not match original %d", loadedAudio.BitDepth, audio.BitDepth)
	}

	// 允许样本数量有 1% 的误差
	samplesDiff := float64(abs(len(loadedAudio.Samples)-len(audio.Samples))) / float64(len(audio.Samples))
	if samplesDiff > 0.01 {
		t.Errorf("wav samples length %d differs too much from original %d (%.2f%% difference)",
			len(loadedAudio.Samples), len(audio.Samples), samplesDiff*100)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
