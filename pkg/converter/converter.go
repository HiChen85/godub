package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// FFProbeOutput ffprobe输出的JSON结构
type FFProbeOutput struct {
	Streams []struct {
		CodecType  string `json:"codec_type"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
		BitsPerRaw string `json:"bits_per_raw_sample"`
	} `json:"streams"`
}

// LoadAudioFile 从文件加载音频数据
func LoadAudioFile(path string, format string) ([]float64, int, int, int, error) {
	// 创建临时WAV文件
	tempDir, err := os.MkdirTemp("", "goudub")
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	wavPath := filepath.Join(tempDir, "temp.wav")
	var converted bool

	// 首先获取源文件信息
	infoCmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	output, err := infoCmd.Output()
	if err == nil {
		var probeData FFProbeOutput
		if err := json.Unmarshal(output, &probeData); err == nil {
			for _, stream := range probeData.Streams {
				if stream.CodecType == "audio" && stream.BitsPerRaw != "" {
					if bitDepth, err := strconv.Atoi(stream.BitsPerRaw); err == nil {
						// 使用源文件的位深度
						cmd := exec.Command("ffmpeg", "-i", path,
							"-acodec", fmt.Sprintf("pcm_s%dle", bitDepth),
							"-f", "wav",
							wavPath)
						if err := cmd.Run(); err == nil {
							converted = true
							break
						}
					}
				}
			}
		}
	}

	// 如果无法获取源文件信息或转换失败，使用默认设置
	if !converted {
		cmd := exec.Command("ffmpeg", "-i", path,
			"-acodec", "pcm_s16le",
			"-f", "wav",
			wavPath)
		if err := cmd.Run(); err != nil {
			return nil, 0, 0, 0, fmt.Errorf("failed to convert audio to wav: %w", err)
		}
	}

	// 读取WAV文件信息
	infoCmd = exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", wavPath)
	output, err = infoCmd.Output()
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to get audio info: %w", err)
	}

	// 解析ffprobe输出的JSON数据
	var probeData FFProbeOutput
	if err := json.Unmarshal(output, &probeData); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to parse audio info: %w", err)
	}

	// 获取音频流信息
	var audioStream *struct {
		CodecType  string `json:"codec_type"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
		BitsPerRaw string `json:"bits_per_raw_sample"`
	}

	for i := range probeData.Streams {
		if probeData.Streams[i].CodecType == "audio" {
			audioStream = &probeData.Streams[i]
			break
		}
	}

	if audioStream == nil {
		return nil, 0, 0, 0, fmt.Errorf("no audio stream found")
	}

	// 解析音频参数
	sampleRate, err := strconv.Atoi(audioStream.SampleRate)
	if err != nil {
		sampleRate = 44100 // 默认采样率
	}

	channels := audioStream.Channels
	if channels <= 0 {
		channels = 2 // 默认声道数
	}

	bitDepth := 16 // 默认位深度
	if audioStream.BitsPerRaw != "" {
		if bd, err := strconv.Atoi(audioStream.BitsPerRaw); err == nil {
			bitDepth = bd
		}
	}

	// 读取WAV文件数据
	data, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to read wav file: %w", err)
	}

	// 计算每个样本的字节数
	bytesPerSample := bitDepth / 8
	samples := make([]float64, len(data)/bytesPerSample)

	// 将字节数据转换为float64样本
	for i := 0; i < len(samples); i++ {
		var sample int32
		switch bitDepth {
		case 8:
			sample = int32(data[i]) - 128
			samples[i] = float64(sample) / 128.0
		case 16:
			sample = int32(int16(data[i*2]) | int16(data[i*2+1])<<8)
			samples[i] = float64(sample) / 32768.0
		case 24:
			sample = int32(data[i*3]) | int32(data[i*3+1])<<8 | int32(data[i*3+2])<<16
			if sample&0x800000 != 0 {
				sample |= ^0xffffff
			}
			samples[i] = float64(sample) / 8388608.0
		case 32:
			sample = int32(data[i*4]) | int32(data[i*4+1])<<8 | int32(data[i*4+2])<<16 | int32(data[i*4+3])<<24
			samples[i] = float64(sample) / 2147483648.0
		default:
			return nil, 0, 0, 0, fmt.Errorf("unsupported bit depth: %d", bitDepth)
		}
	}

	return samples, sampleRate, channels, bitDepth, nil
}

// SaveAudioFile 将音频数据保存到文件
func SaveAudioFile(samples []float64, sampleRate, channels, bitDepth int, path string, format string) error {
	// 验证位深度
	switch bitDepth {
	case 8, 16, 24, 32:
		// 支持的位深度
	default:
		return fmt.Errorf("unsupported bit depth: %d", bitDepth)
	}

	// 创建临时WAV文件
	tempDir, err := os.MkdirTemp("", "goudub")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	wavPath := filepath.Join(tempDir, "temp.wav")

	// 创建WAV文件
	f, err := os.Create(wavPath)
	if err != nil {
		return fmt.Errorf("failed to create wav file: %w", err)
	}
	defer f.Close()

	// 计算数据大小
	dataSize := len(samples) * bitDepth / 8

	// 写入WAV文件头
	if err := writeWAVHeader(f, sampleRate, channels, bitDepth, dataSize); err != nil {
		return fmt.Errorf("failed to write wav header: %w", err)
	}

	// 写入音频数据
	if err := writeWAVData(f, samples, bitDepth); err != nil {
		return fmt.Errorf("failed to write wav data: %w", err)
	}

	// 确保文件已完全写入
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync wav file: %w", err)
	}

	// 如果目标格式是WAV，直接复制文件
	if format == "wav" {
		data, err := os.ReadFile(wavPath)
		if err != nil {
			return fmt.Errorf("failed to read wav file: %w", err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		return nil
	}

	// 根据格式选择适当的编码器和参数
	var cmd *exec.Cmd
	switch format {
	case "mp3":
		cmd = exec.Command("ffmpeg", "-y",
			"-f", "wav",
			"-i", wavPath,
			"-c:a", "libmp3lame",
			"-b:a", "320k", // 使用高比特率
			"-ar", strconv.Itoa(sampleRate),
			"-ac", strconv.Itoa(channels),
			path)
	case "ogg":
		cmd = exec.Command("ffmpeg", "-y",
			"-f", "wav",
			"-i", wavPath,
			"-c:a", "libvorbis",
			"-q:a", "10", // 使用高质量设置
			"-ar", strconv.Itoa(sampleRate),
			"-ac", strconv.Itoa(channels),
			path)
	default:
		cmd = exec.Command("ffmpeg", "-y",
			"-f", "wav",
			"-i", wavPath,
			"-acodec", "pcm_s"+strconv.Itoa(bitDepth)+"le",
			"-ar", strconv.Itoa(sampleRate),
			"-ac", strconv.Itoa(channels),
			"-f", format,
			path)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert wav to %s: %w", format, err)
	}

	return nil
}
