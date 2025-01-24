package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAudioFile(t *testing.T) {
	// 创建测试用例
	tests := []struct {
		name          string
		inputFile     string
		format        string
		expectError   bool
		expectedRate  int
		expectedChan  int
		expectedDepth int
	}{
		{
			name:          "Load MP3 File",
			inputFile:     "../../test/testdata/test.mp3",
			format:        "mp3",
			expectError:   false,
			expectedRate:  44100,
			expectedChan:  1,
			expectedDepth: 16,
		},
		{
			name:        "Non-existent File",
			inputFile:   "non_existent.mp3",
			format:      "mp3",
			expectError: true,
		},
		{
			name:        "Invalid Format",
			inputFile:   "../../test/testdata/test.txt",
			format:      "txt",
			expectError: true,
		},
	}

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			samples, rate, channels, depth, err := LoadAudioFile(tt.inputFile, tt.format)

			// 检查错误情况
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			// 检查正常情况
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// 验证音频参数
			if rate != tt.expectedRate {
				t.Errorf("expected sample rate %d, got %d", tt.expectedRate, rate)
			}
			if channels != tt.expectedChan {
				t.Errorf("expected channels %d, got %d", tt.expectedChan, channels)
			}
			if depth != tt.expectedDepth {
				t.Errorf("expected bit depth %d, got %d", tt.expectedDepth, depth)
			}
			if len(samples) == 0 {
				t.Error("expected non-empty samples")
			}
		})
	}
}

func TestSaveAudioFile(t *testing.T) {
	// 创建测试用例
	tests := []struct {
		name        string
		samples     []float64
		sampleRate  int
		channels    int
		bitDepth    int
		format      string
		expectError bool
	}{
		{
			name:        "Save 16-bit WAV",
			samples:     []float64{0.0, 0.5, -0.5, 1.0, -1.0},
			sampleRate:  44100,
			channels:    1,
			bitDepth:    16,
			format:      "wav",
			expectError: false,
		},
		{
			name:        "Save 24-bit WAV",
			samples:     []float64{0.0, 0.5, -0.5, 1.0, -1.0},
			sampleRate:  48000,
			channels:    1,
			bitDepth:    24,
			format:      "wav",
			expectError: false,
		},
		{
			name:        "Invalid Bit Depth",
			samples:     []float64{0.0, 0.5, -0.5},
			sampleRate:  44100,
			channels:    1,
			bitDepth:    12, // 不支持的位深度
			format:      "wav",
			expectError: true,
		},
	}

	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "goudub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "output."+tt.format)
			err := SaveAudioFile(tt.samples, tt.sampleRate, tt.channels, tt.bitDepth, outputPath, tt.format)

			// 检查错误情况
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			// 检查正常情况
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// 验证文件是否创建
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("output file was not created")
			}

			// 尝试重新加载文件并验证参数
			samples, rate, channels, depth, err := LoadAudioFile(outputPath, tt.format)
			if err != nil {
				t.Errorf("failed to load saved file: %v", err)
				return
			}

			// 验证音频参数
			if rate != tt.sampleRate {
				t.Errorf("expected sample rate %d, got %d", tt.sampleRate, rate)
			}
			if channels != tt.channels {
				t.Errorf("expected channels %d, got %d", tt.channels, channels)
			}
			if depth != tt.bitDepth {
				t.Errorf("expected bit depth %d, got %d", tt.bitDepth, depth)
			}
			if len(samples) == 0 {
				t.Error("expected non-empty samples")
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// 生成测试音频数据
	samples := make([]float64, 44100) // 1秒的音频
	for i := range samples {
		// 生成一个440Hz的正弦波
		samples[i] = float64(0.5 * float64(i%440) / 440.0)
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "goudub_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试不同格式的往返转换
	formats := []string{"wav", "mp3", "ogg"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			// 保存音频
			outputPath := filepath.Join(tempDir, "test."+format)
			err := SaveAudioFile(samples, 44100, 1, 16, outputPath, format)
			if err != nil {
				t.Fatalf("failed to save audio: %v", err)
			}

			// 重新加载音频
			loadedSamples, rate, channels, depth, err := LoadAudioFile(outputPath, format)
			if err != nil {
				t.Fatalf("failed to load audio: %v", err)
			}

			// 验证基本参数
			if rate != 44100 {
				t.Errorf("expected sample rate 44100, got %d", rate)
			}
			if channels != 1 {
				t.Errorf("expected 1 channel, got %d", channels)
			}
			if depth != 16 {
				t.Errorf("expected 16-bit depth, got %d", depth)
			}

			// 验证样本数据（考虑到压缩格式可能会有些许差异，这里只检查长度）
			if len(loadedSamples) == 0 {
				t.Error("loaded samples is empty")
			}
		})
	}
}

func TestOggToWavConversion(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "goudub_test")
	defer os.RemoveAll(tempDir)
	if err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// 测试用例
	tests := []struct {
		name          string
		inputFile     string
		expectedRate  int
		expectedChan  int
		expectedDepth int
	}{
		{
			name:          "Convert OGG to WAV",
			inputFile:     "../../test/testdata/news.ogg",
			expectedRate:  48000, // 实际采样率
			expectedChan:  1,     // 单声道
			expectedDepth: 16,    // 标准CD质量
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 加载OGG文件
			samples, rate, channels, depth, err := LoadAudioFile(tt.inputFile, "ogg")
			if err != nil {
				t.Fatalf("failed to load OGG file: %v", err)
			}

			// 验证加载的音频参数
			if rate != tt.expectedRate {
				t.Errorf("expected sample rate %d, got %d", tt.expectedRate, rate)
			}
			if channels != tt.expectedChan {
				t.Errorf("expected channels %d, got %d", tt.expectedChan, channels)
			}
			if depth != tt.expectedDepth {
				t.Errorf("expected bit depth %d, got %d", tt.expectedDepth, depth)
			}
			if len(samples) == 0 {
				t.Error("expected non-empty samples")
			}

			// 2. 保存为WAV
			outputPath := filepath.Join(tempDir, "news_converted.wav")
			err = SaveAudioFile(samples, rate, channels, depth, outputPath, "wav")
			if err != nil {
				t.Fatalf("failed to save WAV file: %v", err)
			}

			t.Logf("Converted file saved to: %s", outputPath)

			// 3. 重新加载WAV文件并验证参数
			wavSamples, wavRate, wavChannels, wavDepth, err := LoadAudioFile(outputPath, "wav")
			if err != nil {
				t.Fatalf("failed to load converted WAV file: %v", err)
			}

			// 验证转换后的音频参数
			if wavRate != rate {
				t.Errorf("WAV sample rate %d does not match original %d", wavRate, rate)
			}
			if wavChannels != channels {
				t.Errorf("WAV channels %d does not match original %d", wavChannels, channels)
			}
			if wavDepth != depth {
				t.Errorf("WAV bit depth %d does not match original %d", wavDepth, depth)
			}

			// 由于重采样和编解码的原因，样本数可能会有细微差异
			samplesDiff := float64(abs(len(wavSamples)-len(samples))) / float64(len(samples))
			if samplesDiff > 0.001 { // 允许0.1%的误差
				t.Errorf("WAV samples length %d differs too much from original %d (%.2f%% difference)",
					len(wavSamples), len(samples), samplesDiff*100)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
