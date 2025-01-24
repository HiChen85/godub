package stream

import (
	"bytes"
	"io"
	"testing"

	"github.com/HiChen85/godub/pkg/audio"
)

func TestNewAudioStream(t *testing.T) {
	tests := []struct {
		name        string
		reader      io.Reader
		writer      io.Writer
		sampleRate  int
		channels    int
		bitDepth    int
		bufferSize  int
		expectError bool
	}{
		{
			name:        "Valid Stream with Reader",
			reader:      bytes.NewBuffer([]byte{}),
			writer:      nil,
			sampleRate:  44100,
			channels:    2,
			bitDepth:    16,
			bufferSize:  1024,
			expectError: false,
		},
		{
			name:        "Valid Stream with Writer",
			reader:      nil,
			writer:      bytes.NewBuffer([]byte{}),
			sampleRate:  44100,
			channels:    2,
			bitDepth:    16,
			bufferSize:  1024,
			expectError: false,
		},
		{
			name:        "Invalid - No Reader or Writer",
			reader:      nil,
			writer:      nil,
			sampleRate:  44100,
			channels:    2,
			bitDepth:    16,
			bufferSize:  1024,
			expectError: true,
		},
		{
			name:        "Invalid Sample Rate",
			reader:      bytes.NewBuffer([]byte{}),
			writer:      nil,
			sampleRate:  0,
			channels:    2,
			bitDepth:    16,
			bufferSize:  1024,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := NewAudioStream(
				tt.reader,
				tt.writer,
				tt.sampleRate,
				tt.channels,
				tt.bitDepth,
				tt.bufferSize,
			)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if stream.sampleRate != tt.sampleRate {
				t.Errorf("expected sample rate %d, got %d", tt.sampleRate, stream.sampleRate)
			}
			if stream.channels != tt.channels {
				t.Errorf("expected channels %d, got %d", tt.channels, stream.channels)
			}
			if stream.bitDepth != tt.bitDepth {
				t.Errorf("expected bit depth %d, got %d", tt.bitDepth, stream.bitDepth)
			}
			if stream.bufferSize != tt.bufferSize {
				t.Errorf("expected buffer size %d, got %d", tt.bufferSize, stream.bufferSize)
			}
		})
	}
}

func TestAudioStreamReadWrite(t *testing.T) {
	// 创建测试数据
	testSamples := []float64{0.0, 0.5, -0.5, 1.0, -1.0}
	sampleRate := 44100
	channels := 1
	bitDepth := 16
	bufferSize := 1024

	// 创建一个缓冲区作为writer
	buf := bytes.NewBuffer([]byte{})

	// 创建输出流
	outStream, err := NewAudioStream(nil, buf, sampleRate, channels, bitDepth, bufferSize)
	if err != nil {
		t.Fatalf("failed to create output stream: %v", err)
	}

	// 写入测试数据
	n, err := outStream.Write(testSamples)
	if err != nil {
		t.Fatalf("failed to write samples: %v", err)
	}
	if n != len(testSamples) {
		t.Errorf("expected to write %d samples, wrote %d", len(testSamples), n)
	}

	// 创建输入流
	inStream, err := NewAudioStream(bytes.NewReader(buf.Bytes()), nil, sampleRate, channels, bitDepth, bufferSize)
	if err != nil {
		t.Fatalf("failed to create input stream: %v", err)
	}

	// 读取数据
	readSamples := make([]float64, len(testSamples))
	n, err = inStream.Read(readSamples)
	if err != nil {
		t.Fatalf("failed to read samples: %v", err)
	}
	if n != len(testSamples) {
		t.Errorf("expected to read %d samples, read %d", len(testSamples), n)
	}

	// 比较原始数据和读取的数据
	for i := 0; i < len(testSamples); i++ {
		if !almostEqual(testSamples[i], readSamples[i], 0.0001) {
			t.Errorf("sample %d: expected %f, got %f", i, testSamples[i], readSamples[i])
		}
	}
}

func TestAudioStreamProcess(t *testing.T) {
	// 创建测试数据
	testSamples := []float64{0.0, 0.25, -0.25, 0.5, -0.5}
	sampleRate := 44100
	channels := 1
	bitDepth := 16
	bufferSize := 1024

	// 创建输入缓冲区
	inBuf := bytes.NewBuffer([]byte{})
	outBuf := bytes.NewBuffer([]byte{})

	// 创建输入流并写入测试数据
	inStream, err := NewAudioStream(nil, inBuf, sampleRate, channels, bitDepth, bufferSize)
	if err != nil {
		t.Fatalf("failed to create input stream: %v", err)
	}
	if _, err := inStream.Write(testSamples); err != nil {
		t.Fatalf("failed to write test samples: %v", err)
	}

	// 打印写入的字节数据
	t.Logf("Written bytes: %v", inBuf.Bytes())

	// 创建处理流
	stream, err := NewAudioStream(bytes.NewReader(inBuf.Bytes()), outBuf, sampleRate, channels, bitDepth, bufferSize)
	if err != nil {
		t.Fatalf("failed to create processing stream: %v", err)
	}

	// 定义处理函数（这里简单地将所有样本值翻倍）
	processor := func(samples []float64) error {
		t.Logf("Processing samples: %v", samples)
		for i := range samples {
			samples[i] *= 2.0
		}
		t.Logf("Processed samples: %v", samples)
		return nil
	}

	// 处理数据
	if err := stream.Process(processor); err != nil {
		t.Fatalf("failed to process stream: %v", err)
	}

	// 打印处理后的字节数据
	t.Logf("Processed bytes: %v", outBuf.Bytes())

	// 读取处理后的数据
	processedStream, err := NewAudioStream(bytes.NewReader(outBuf.Bytes()), nil, sampleRate, channels, bitDepth, bufferSize)
	if err != nil {
		t.Fatalf("failed to create output stream: %v", err)
	}

	processedSamples := make([]float64, len(testSamples))
	if _, err := processedStream.Read(processedSamples); err != nil {
		t.Fatalf("failed to read processed samples: %v", err)
	}

	t.Logf("Final samples: %v", processedSamples)

	// 验证处理结果
	for i := 0; i < len(testSamples); i++ {
		expected := testSamples[i] * 2.0
		if !almostEqual(expected, processedSamples[i], 0.0001) {
			t.Errorf("sample %d: expected %f, got %f", i, expected, processedSamples[i])
		}
	}
}

func TestAudioStreamToFromSegment(t *testing.T) {
	// 创建测试数据
	testSamples := []float64{0.0, 0.5, -0.5, 1.0, -1.0}
	sampleRate := 44100
	channels := 1
	bitDepth := 16

	// 创建音频段
	segment, err := audio.NewAudioSegment(testSamples, sampleRate, channels, bitDepth)
	if err != nil {
		t.Fatalf("failed to create audio segment: %v", err)
	}

	// 创建输出缓冲区
	buf := bytes.NewBuffer([]byte{})

	// 从音频段创建流
	_, err = FromSegment(segment, buf, 1024)
	if err != nil {
		t.Fatalf("failed to create stream from segment: %v", err)
	}

	// 创建新的流用于读取数据
	readStream, err := NewAudioStream(bytes.NewReader(buf.Bytes()), nil, sampleRate, channels, bitDepth, 1024)
	if err != nil {
		t.Fatalf("failed to create read stream: %v", err)
	}

	// 将流转换回音频段
	newSegment, err := readStream.ToSegment()
	if err != nil {
		t.Fatalf("failed to convert stream to segment: %v", err)
	}

	// 验证参数
	if newSegment.SampleRate() != segment.SampleRate() {
		t.Errorf("expected sample rate %d, got %d", segment.SampleRate(), newSegment.SampleRate())
	}
	if newSegment.Channels() != segment.Channels() {
		t.Errorf("expected channels %d, got %d", segment.Channels(), newSegment.Channels())
	}
	if newSegment.BitDepth() != segment.BitDepth() {
		t.Errorf("expected bit depth %d, got %d", segment.BitDepth(), newSegment.BitDepth())
	}

	// 验证样本数据
	originalSamples := segment.Samples()
	newSamples := newSegment.Samples()
	if len(newSamples) != len(originalSamples) {
		t.Errorf("expected %d samples, got %d", len(originalSamples), len(newSamples))
	}

	for i := 0; i < len(originalSamples); i++ {
		if !almostEqual(originalSamples[i], newSamples[i], 0.0001) {
			t.Errorf("sample %d: expected %f, got %f", i, originalSamples[i], newSamples[i])
		}
	}
}

// almostEqual 比较两个浮点数是否近似相等
func almostEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
