package stream

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// AudioFormat 音频格式
type AudioFormat string

const (
	FormatOGG AudioFormat = "ogg"
	FormatWAV AudioFormat = "wav"
	FormatMP3 AudioFormat = "mp3"
)

// AudioParams 音频参数
type AudioParams struct {
	SampleRate int // 采样率
	Channels   int // 声道数
	BitDepth   int // 位深度
}

// AudioProcessor 音频处理器接口
type AudioProcessor interface {
	// Process 处理音频数据，返回处理后的音频帧和参数
	Process(data []byte, params AudioParams) ([][]byte, error)
}

// FFmpegProcessor 基于 FFmpeg 的音频处理器
type FFmpegProcessor struct {
	inputFormat AudioFormat
}

// NewFFmpegProcessor 创建新的 FFmpeg 处理器
func NewFFmpegProcessor(format AudioFormat) *FFmpegProcessor {
	return &FFmpegProcessor{
		inputFormat: format,
	}
}

// Process 实现 AudioProcessor 接口
func (p *FFmpegProcessor) Process(data []byte, params AudioParams) ([][]byte, error) {
	// 创建命令管道用于音频转换
	ffmpeg := exec.Command("ffmpeg",
		"-f", string(p.inputFormat), // 输入格式
		"-i", "pipe:0", // 从标准输入读取
		"-ar", strconv.Itoa(params.SampleRate),
		"-ac", strconv.Itoa(params.Channels),
		"-acodec", fmt.Sprintf("pcm_s%dle", params.BitDepth),
		"-f", "wav",
		"pipe:1",
	)

	// 设置标准输入输出
	ffmpeg.Stdin = bytes.NewReader(data)
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// 启动命令
	if err := ffmpeg.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// 跳过 WAV 文件头（44 字节）
	header := make([]byte, 44)
	if _, err := io.ReadFull(stdout, header); err != nil {
		ffmpeg.Process.Kill()
		return nil, fmt.Errorf("failed to read wav header: %w", err)
	}

	// 计算帧大小（每个采样的字节数 * 声道数）
	frameSize := (params.BitDepth / 8) * params.Channels

	// 读取音频数据并按帧存储
	var frames [][]byte
	buffer := make([]byte, frameSize*1024) // 每次读取 1024 个采样
	for {
		n, err := stdout.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			ffmpeg.Process.Kill()
			return nil, fmt.Errorf("failed to read audio data: %w", err)
		}

		// 确保读取的数据是完整的帧
		frameCount := n / frameSize
		if frameCount > 0 {
			frame := make([]byte, frameCount*frameSize)
			copy(frame, buffer[:frameCount*frameSize])
			frames = append(frames, frame)
		}
	}

	// 等待命令完成
	if err := ffmpeg.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg process failed: %w", err)
	}

	return frames, nil
}
