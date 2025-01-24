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
			audio, err := LoadAudioFile(tt.inputFile, tt.format)

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
			if audio.SampleRate != tt.expectedRate {
				t.Errorf("expected sample rate %d, got %d", tt.expectedRate, audio.SampleRate)
			}
			if audio.Channels != tt.expectedChan {
				t.Errorf("expected channels %d, got %d", tt.expectedChan, audio.Channels)
			}
			if audio.BitDepth != tt.expectedDepth {
				t.Errorf("expected bit depth %d, got %d", tt.expectedDepth, audio.BitDepth)
			}
			if len(audio.Samples) == 0 {
				t.Error("expected non-empty samples")
			}
		})
	}
}

func TestSaveAudioFile(t *testing.T) {
	// 创建测试用例
	tests := []struct {
		name        string
		audio       *AudioData
		format      string
		expectError bool
	}{
		{
			name: "Save 16-bit WAV",
			audio: &AudioData{
				Samples:    []float64{0.0, 0.5, -0.5, 1.0, -1.0},
				SampleRate: 44100,
				Channels:   1,
				BitDepth:   16,
			},
			format:      "wav",
			expectError: false,
		},
		{
			name: "Save 24-bit WAV",
			audio: &AudioData{
				Samples:    []float64{0.0, 0.5, -0.5, 1.0, -1.0},
				SampleRate: 48000,
				Channels:   1,
				BitDepth:   24,
			},
			format:      "wav",
			expectError: false,
		},
		{
			name: "Invalid Bit Depth",
			audio: &AudioData{
				Samples:    []float64{0.0, 0.5, -0.5},
				SampleRate: 44100,
				Channels:   1,
				BitDepth:   12, // 不支持的位深度
			},
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
			err := SaveAudioFile(tt.audio, outputPath, tt.format)

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
			loadedAudio, err := LoadAudioFile(outputPath, tt.format)
			if err != nil {
				t.Errorf("failed to load saved file: %v", err)
				return
			}

			// 验证音频参数
			if loadedAudio.SampleRate != tt.audio.SampleRate {
				t.Errorf("expected sample rate %d, got %d", tt.audio.SampleRate, loadedAudio.SampleRate)
			}
			if loadedAudio.Channels != tt.audio.Channels {
				t.Errorf("expected channels %d, got %d", tt.audio.Channels, loadedAudio.Channels)
			}
			if loadedAudio.BitDepth != tt.audio.BitDepth {
				t.Errorf("expected bit depth %d, got %d", tt.audio.BitDepth, loadedAudio.BitDepth)
			}
			if len(loadedAudio.Samples) == 0 {
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

	audio := &AudioData{
		Samples:    samples,
		SampleRate: 44100,
		Channels:   1,
		BitDepth:   16,
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
			err := SaveAudioFile(audio, outputPath, format)
			if err != nil {
				t.Fatalf("failed to save audio: %v", err)
			}

			// 重新加载音频
			loadedAudio, err := LoadAudioFile(outputPath, format)
			if err != nil {
				t.Fatalf("failed to load audio: %v", err)
			}

			// 验证基本参数
			if loadedAudio.SampleRate != 44100 {
				t.Errorf("expected sample rate 44100, got %d", loadedAudio.SampleRate)
			}
			if loadedAudio.Channels != 1 {
				t.Errorf("expected 1 channel, got %d", loadedAudio.Channels)
			}
			if loadedAudio.BitDepth != 16 {
				t.Errorf("expected 16-bit depth, got %d", loadedAudio.BitDepth)
			}

			// 验证样本数据（考虑到压缩格式可能会有些许差异，这里只检查长度）
			if len(loadedAudio.Samples) == 0 {
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

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加载 OGG 文件
			audio, err := LoadAudioFile(tt.inputFile, "ogg")
			if err != nil {
				t.Fatalf("failed to load ogg file: %v", err)
			}

			// 验证音频参数
			if audio.SampleRate != tt.expectedRate {
				t.Errorf("expected sample rate %d, got %d", tt.expectedRate, audio.SampleRate)
			}
			if audio.Channels != tt.expectedChan {
				t.Errorf("expected channels %d, got %d", tt.expectedChan, audio.Channels)
			}
			if audio.BitDepth != tt.expectedDepth {
				t.Errorf("expected bit depth %d, got %d", tt.expectedDepth, audio.BitDepth)
			}
			if len(audio.Samples) == 0 {
				t.Error("expected non-empty samples")
			}

			// 保存为 WAV 文件
			outputPath := filepath.Join(tempDir, "output.wav")
			err = SaveAudioFile(audio, outputPath, "wav")
			if err != nil {
				t.Fatalf("failed to save wav file: %v", err)
			}

			// 验证文件是否创建
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("output file was not created")
			}

			// 重新加载 WAV 文件并验证参数
			loadedAudio, err := LoadAudioFile(outputPath, "wav")
			if err != nil {
				t.Fatalf("failed to load wav file: %v", err)
			}

			// 验证音频参数
			if loadedAudio.SampleRate != tt.expectedRate {
				t.Errorf("expected sample rate %d, got %d", tt.expectedRate, loadedAudio.SampleRate)
			}
			if loadedAudio.Channels != tt.expectedChan {
				t.Errorf("expected channels %d, got %d", tt.expectedChan, loadedAudio.Channels)
			}
			if loadedAudio.BitDepth != tt.expectedDepth {
				t.Errorf("expected bit depth %d, got %d", tt.expectedDepth, loadedAudio.BitDepth)
			}
			if len(loadedAudio.Samples) == 0 {
				t.Error("expected non-empty samples")
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
