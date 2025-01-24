package effects

import (
	"math"
	"time"

	"github.com/HiChen85/godub/pkg/audio"
)

// FadeIn 实现音频淡入效果
func FadeIn(segment *audio.AudioSegment, duration time.Duration) (*audio.AudioSegment, error) {
	samples := segment.Samples()
	sampleRate := segment.SampleRate()
	channels := segment.Channels()
	fadeLength := int(float64(duration) * float64(sampleRate*channels) / float64(time.Second))

	if fadeLength > len(samples) {
		fadeLength = len(samples)
	}

	newSamples := make([]float64, len(samples))
	copy(newSamples, samples)

	// 应用淡入效果
	for i := 0; i < fadeLength; i++ {
		factor := float64(i) / float64(fadeLength)
		newSamples[i] = samples[i] * factor
	}

	return audio.NewAudioSegment(newSamples, sampleRate, channels, segment.BitDepth())
}

// FadeOut 实现音频淡出效果
func FadeOut(segment *audio.AudioSegment, duration time.Duration) (*audio.AudioSegment, error) {
	samples := segment.Samples()
	sampleRate := segment.SampleRate()
	channels := segment.Channels()
	fadeLength := int(float64(duration) * float64(sampleRate*channels) / float64(time.Second))

	if fadeLength > len(samples) {
		fadeLength = len(samples)
	}

	newSamples := make([]float64, len(samples))
	copy(newSamples, samples)

	// 应用淡出效果
	startIndex := len(samples) - fadeLength
	for i := 0; i < fadeLength; i++ {
		factor := float64(fadeLength-i) / float64(fadeLength)
		newSamples[startIndex+i] = samples[startIndex+i] * factor
	}

	return audio.NewAudioSegment(newSamples, sampleRate, channels, segment.BitDepth())
}

// Normalize 标准化音频音量
func Normalize(segment *audio.AudioSegment) (*audio.AudioSegment, error) {
	samples := segment.Samples()
	sampleRate := segment.SampleRate()
	channels := segment.Channels()

	// 找到最大振幅
	maxAmp := 0.0
	for _, sample := range samples {
		abs := math.Abs(sample)
		if abs > maxAmp {
			maxAmp = abs
		}
	}

	// 如果最大振幅为0，直接返回原始音频
	if maxAmp == 0 {
		return segment, nil
	}

	// 标准化样本
	newSamples := make([]float64, len(samples))
	for i, sample := range samples {
		newSamples[i] = sample / maxAmp
	}

	return audio.NewAudioSegment(newSamples, sampleRate, channels, segment.BitDepth())
}

// AdjustVolume 调整音频音量
func AdjustVolume(segment *audio.AudioSegment, dB float64) (*audio.AudioSegment, error) {
	samples := segment.Samples()
	sampleRate := segment.SampleRate()
	channels := segment.Channels()

	// 将dB转换为振幅倍数
	factor := math.Pow(10, dB/20.0)

	// 调整样本振幅
	newSamples := make([]float64, len(samples))
	for i, sample := range samples {
		newSamples[i] = sample * factor
	}

	return audio.NewAudioSegment(newSamples, sampleRate, channels, segment.BitDepth())
}
