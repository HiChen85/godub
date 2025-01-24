package audio

import (
	"time"

	"github.com/pkg/errors"
)

// AudioSegment 表示一个音频片段
type AudioSegment struct {
	// 音频数据
	samples []float64
	// 采样率 (Hz)
	sampleRate int
	// 声道数
	channels int
	// 位深度
	bitDepth int
	// 音频时长
	duration time.Duration
}

// NewAudioSegment 创建一个新的音频段
func NewAudioSegment(samples []float64, sampleRate, channels, bitDepth int) (*AudioSegment, error) {
	if len(samples) == 0 {
		return nil, errors.New("samples cannot be empty")
	}
	if sampleRate <= 0 {
		return nil, errors.New("sample rate must be positive")
	}
	if channels <= 0 {
		return nil, errors.New("channels must be positive")
	}
	if bitDepth <= 0 {
		return nil, errors.New("bit depth must be positive")
	}

	duration := time.Duration(float64(len(samples)) / float64(sampleRate*channels) * float64(time.Second))

	return &AudioSegment{
		samples:    samples,
		sampleRate: sampleRate,
		channels:   channels,
		bitDepth:   bitDepth,
		duration:   duration,
	}, nil
}

// Duration 返回音频段的时长
func (a *AudioSegment) Duration() time.Duration {
	return a.duration
}

// SampleRate 返回采样率
func (a *AudioSegment) SampleRate() int {
	return a.sampleRate
}

// Channels 返回声道数
func (a *AudioSegment) Channels() int {
	return a.channels
}

// BitDepth 返回位深度
func (a *AudioSegment) BitDepth() int {
	return a.bitDepth
}

// Samples 返回音频样本数据
func (a *AudioSegment) Samples() []float64 {
	return a.samples
}

// Slice 切片音频段
func (a *AudioSegment) Slice(start, end time.Duration) (*AudioSegment, error) {
	if start < 0 || end > a.duration || start >= end {
		return nil, errors.New("invalid slice range")
	}

	startSample := int(float64(start) * float64(a.sampleRate*a.channels) / float64(time.Second))
	endSample := int(float64(end) * float64(a.sampleRate*a.channels) / float64(time.Second))

	return NewAudioSegment(
		a.samples[startSample:endSample],
		a.sampleRate,
		a.channels,
		a.bitDepth,
	)
}
