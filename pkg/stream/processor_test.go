package stream

import (
	"encoding/base64"
	"github.com/HiChen85/godub/test"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFFmpegProcessor(t *testing.T) {
	// 解码测试用的 OGG 数据
	oggData := make([]byte, 0)
	for _, str := range test.Audio {
		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to decode base64 data: %v", err)
		}
		oggData = append(oggData, data...)
	}

	// 创建处理器
	processor := NewFFmpegProcessor(FormatOGG)

	// 设置音频参数
	params := AudioParams{
		SampleRate: 16000,
		Channels:   1,
		BitDepth:   16,
	}

	// 处理音频流
	frames, err := processor.Process(oggData, params)
	if err != nil {
		t.Fatalf("failed to process ogg stream: %v", err)
	}

	// 验证是否生成了音频帧
	if len(frames) == 0 {
		t.Error("no audio frames generated")
	}

	// 创建临时目录用于测试文件写入
	tempDir, err := os.MkdirTemp("", "godub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试写入文件
	wavPath := filepath.Join(tempDir, "test.wav")
	if err := WriteWAVFile(frames, params, wavPath); err != nil {
		t.Fatalf("failed to write wav file: %v", err)
	}

	// 验证文件是否创建
	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		t.Error("output file was not created")
	}

	// 验证文件大小
	info, err := os.Stat(wavPath)
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}

	// WAV 文件至少应该有文件头（44字节）
	if info.Size() <= 44 {
		t.Error("output file is too small")
	}

	// 计算预期的文件大小
	var expectedDataSize int
	for _, frame := range frames {
		expectedDataSize += len(frame)
	}
	expectedFileSize := 44 + expectedDataSize // 44 是 WAV 文件头的大小

	if info.Size() != int64(expectedFileSize) {
		t.Errorf("unexpected file size: expected %d, got %d", expectedFileSize, info.Size())
	}
}

func TestFFmpegProcessorInvalidData(t *testing.T) {
	// 创建处理器
	processor := NewFFmpegProcessor(FormatOGG)

	// 设置音频参数
	params := AudioParams{
		SampleRate: 16000,
		Channels:   1,
		BitDepth:   16,
	}

	// 测试无效的输入数据
	invalidData := []byte("invalid ogg data")
	_, err := processor.Process(invalidData, params)
	if err == nil {
		t.Error("expected error for invalid input data, got nil")
	}
}

func TestWriteWAVFileEmptyFrames(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "godub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试参数
	params := AudioParams{
		SampleRate: 16000,
		Channels:   1,
		BitDepth:   16,
	}

	// 测试写入空帧
	wavPath := filepath.Join(tempDir, "empty.wav")
	if err := WriteWAVFile(nil, params, wavPath); err != nil {
		t.Fatalf("failed to write empty wav file: %v", err)
	}

	// 验证文件大小（应该只有 WAV 文件头）
	info, err := os.Stat(wavPath)
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}

	if info.Size() != 44 {
		t.Errorf("unexpected file size for empty wav: expected 44, got %d", info.Size())
	}
}

func TestProcessOggAndPlayWav(t *testing.T) {
	// 解码测试用的 OGG 数据
	oggData := make([]byte, 0)
	for _, str := range test.Audio {
		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			t.Fatalf("failed to decode base64 data: %v", err)
		}
		oggData = append(oggData, data...)
	}

	// 创建处理器
	processor := NewFFmpegProcessor(FormatOGG)

	// 设置音频参数
	params := AudioParams{
		SampleRate: 16000,
		Channels:   1,
		BitDepth:   16,
	}

	// 处理音频流
	frames, err := processor.Process(oggData, params)
	if err != nil {
		t.Fatalf("failed to process ogg stream: %v", err)
	}

	// 将处理后的音频写入到当前目录的 output.wav
	fileName := "test_output" + time.Now().Format("2006_01_02_15_04_05") + ".wav"
	outputPath := "../../test/testdata/output/" + fileName

	if err := WriteWAVFile(frames, params, outputPath); err != nil {
		t.Fatalf("failed to write wav file: %v", err)
	}

	t.Logf("Generated WAV file at: %s", outputPath)
	t.Logf("Please verify the audio by playing the file")
}
