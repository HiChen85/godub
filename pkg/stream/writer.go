package stream

import (
	"fmt"
	"os"
)

// WriteWAVFile 将音频帧写入 WAV 文件
func WriteWAVFile(frames [][]byte, params AudioParams, path string) error {
	// 创建输出文件
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// 计算数据总大小
	var dataSize int
	for _, frame := range frames {
		dataSize += len(frame)
	}

	// 写入 WAV 文件头
	header := makeWAVHeader(params.SampleRate, params.Channels, params.BitDepth, dataSize)
	if _, err := f.Write(header); err != nil {
		return fmt.Errorf("failed to write wav header: %w", err)
	}

	// 写入音频数据
	for _, frame := range frames {
		if _, err := f.Write(frame); err != nil {
			return fmt.Errorf("failed to write audio frame: %w", err)
		}
	}

	return nil
}

// makeWAVHeader 生成 WAV 文件头
func makeWAVHeader(sampleRate, channels, bitDepth, dataSize int) []byte {
	header := make([]byte, 44)

	// RIFF 头
	copy(header[0:4], []byte("RIFF"))
	// 文件大小（数据大小 + 36）
	putLE(header[4:8], uint32(dataSize+36))
	// WAVE 标识
	copy(header[8:12], []byte("WAVE"))
	// fmt 块
	copy(header[12:16], []byte("fmt "))
	// fmt 块大小（16）
	putLE(header[16:20], uint32(16))
	// 音频格式（1 表示 PCM）
	putLE(header[20:22], uint16(1))
	// 声道数
	putLE(header[22:24], uint16(channels))
	// 采样率
	putLE(header[24:28], uint32(sampleRate))
	// 字节率
	putLE(header[28:32], uint32(sampleRate*channels*bitDepth/8))
	// 块对齐
	putLE(header[32:34], uint16(channels*bitDepth/8))
	// 位深度
	putLE(header[34:36], uint16(bitDepth))
	// data 块
	copy(header[36:40], []byte("data"))
	// 数据大小
	putLE(header[40:44], uint32(dataSize))

	return header
}

// putLE 写入小端序数据
func putLE(b []byte, v interface{}) {
	switch val := v.(type) {
	case uint16:
		b[0] = byte(val)
		b[1] = byte(val >> 8)
	case uint32:
		b[0] = byte(val)
		b[1] = byte(val >> 8)
		b[2] = byte(val >> 16)
		b[3] = byte(val >> 24)
	}
}
