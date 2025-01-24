package stream

import (
	"io"

	"github.com/HiChen85/godub/pkg/audio"
	"github.com/pkg/errors"
)

// AudioStream 表示一个音频流
type AudioStream struct {
	// 音频参数
	sampleRate int
	channels   int
	bitDepth   int

	// 缓冲区大小（以样本为单位）
	bufferSize int

	// 输入输出
	reader io.Reader
	writer io.Writer

	// 缓冲区
	buffer []float64
}

// NewAudioStream 创建一个新的音频流
func NewAudioStream(reader io.Reader, writer io.Writer, sampleRate, channels, bitDepth, bufferSize int) (*AudioStream, error) {
	if reader == nil && writer == nil {
		return nil, errors.New("at least one of reader or writer must be provided")
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
	if bufferSize <= 0 {
		return nil, errors.New("buffer size must be positive")
	}

	return &AudioStream{
		sampleRate: sampleRate,
		channels:   channels,
		bitDepth:   bitDepth,
		bufferSize: bufferSize,
		reader:     reader,
		writer:     writer,
		buffer:     make([]float64, bufferSize),
	}, nil
}

// Read 从音频流中读取数据
func (s *AudioStream) Read(p []float64) (n int, err error) {
	if s.reader == nil {
		return 0, errors.New("stream is not readable")
	}

	// 计算需要读取的字节数
	bytesPerSample := s.bitDepth / 8
	bytesToRead := len(p) * bytesPerSample
	bytes := make([]byte, bytesToRead)

	// 从底层reader读取字节
	n, err = s.reader.Read(bytes)
	if err != nil {
		return 0, err
	}

	// 将字节转换为float64样本
	samplesRead := n / bytesPerSample
	for i := 0; i < samplesRead; i++ {
		var sample int32
		switch s.bitDepth {
		case 8:
			sample = int32(bytes[i]) - 128
			p[i] = float64(sample) / 127.0
		case 16:
			sample = int32(int16(bytes[i*2]) | int16(bytes[i*2+1])<<8)
			p[i] = float64(sample) / 32767.0
		case 24:
			sample = int32(bytes[i*3]) | int32(bytes[i*3+1])<<8 | int32(bytes[i*3+2])<<16
			if sample&0x800000 != 0 {
				sample |= ^0xffffff
			}
			p[i] = float64(sample) / 8388607.0
		case 32:
			sample = int32(bytes[i*4]) | int32(bytes[i*4+1])<<8 | int32(bytes[i*4+2])<<16 | int32(bytes[i*4+3])<<24
			p[i] = float64(sample) / 2147483647.0
		}
	}

	return samplesRead, nil
}

// Write 写入数据到音频流
func (s *AudioStream) Write(p []float64) (n int, err error) {
	if s.writer == nil {
		return 0, errors.New("stream is not writable")
	}

	// 计算需要写入的字节数
	bytesPerSample := s.bitDepth / 8
	bytes := make([]byte, len(p)*bytesPerSample)

	// 将float64样本转换为字节
	for i := 0; i < len(p); i++ {
		// 限制样本值在有效范围内
		sample := p[i]
		switch s.bitDepth {
		case 8:
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}
			intSample := int32(sample * 127.0)
			bytes[i] = byte(intSample + 128)
		case 16:
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}
			intSample := int32(sample * 32767.0)
			bytes[i*2] = byte(intSample)
			bytes[i*2+1] = byte(intSample >> 8)
		case 24:
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}
			intSample := int32(sample * 8388607.0)
			bytes[i*3] = byte(intSample)
			bytes[i*3+1] = byte(intSample >> 8)
			bytes[i*3+2] = byte(intSample >> 16)
		case 32:
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}
			intSample := int32(sample * 2147483647.0)
			bytes[i*4] = byte(intSample)
			bytes[i*4+1] = byte(intSample >> 8)
			bytes[i*4+2] = byte(intSample >> 16)
			bytes[i*4+3] = byte(intSample >> 24)
		}
	}

	// 写入字节到底层writer
	n, err = s.writer.Write(bytes)
	if err != nil {
		return 0, err
	}

	return n / bytesPerSample, nil
}

// Process 处理音频流数据
func (s *AudioStream) Process(processor func([]float64) error) error {
	if s.reader == nil {
		return errors.New("stream is not readable")
	}

	for {
		// 读取数据到缓冲区
		n, err := s.Read(s.buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 处理缓冲区中的数据
		if err := processor(s.buffer[:n]); err != nil {
			return err
		}

		// 如果有writer，写入处理后的数据
		if s.writer != nil {
			if _, err := s.Write(s.buffer[:n]); err != nil {
				return err
			}
		}
	}

	return nil
}

// ToSegment 将音频流转换为音频段
func (s *AudioStream) ToSegment() (*audio.AudioSegment, error) {
	if s.reader == nil {
		return nil, errors.New("stream is not readable")
	}

	var samples []float64
	buffer := make([]float64, s.bufferSize)

	for {
		n, err := s.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		samples = append(samples, buffer[:n]...)
	}

	return audio.NewAudioSegment(samples, s.sampleRate, s.channels, s.bitDepth)
}

// FromSegment 从音频段创建音频流
func FromSegment(segment *audio.AudioSegment, writer io.Writer, bufferSize int) (*AudioStream, error) {
	if segment == nil {
		return nil, errors.New("segment cannot be nil")
	}

	stream, err := NewAudioStream(
		nil,
		writer,
		segment.SampleRate(),
		segment.Channels(),
		segment.BitDepth(),
		bufferSize,
	)
	if err != nil {
		return nil, err
	}

	samples := segment.Samples()
	if _, err := stream.Write(samples); err != nil {
		return nil, err
	}

	return stream, nil
}
