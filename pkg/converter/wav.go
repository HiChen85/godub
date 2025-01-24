package converter

import (
	"encoding/binary"
	"io"
)

// WAV文件头结构
type wavHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // 文件大小 - 8
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 for mono, 2 for stereo
	SampleRate    uint32  // 采样率
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8, 16, 24, 32
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // 数据大小
}

// writeWAVHeader 写入WAV文件头
func writeWAVHeader(w io.Writer, sampleRate, channels, bitDepth int, dataSize int) error {
	header := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1,
		NumChannels:   uint16(channels),
		SampleRate:    uint32(sampleRate),
		BitsPerSample: uint16(bitDepth),
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(dataSize),
	}

	header.ByteRate = uint32(sampleRate * channels * bitDepth / 8)
	header.BlockAlign = uint16(channels * bitDepth / 8)
	header.ChunkSize = 36 + header.Subchunk2Size

	return binary.Write(w, binary.LittleEndian, &header)
}

// writeWAVData 写入WAV数据
func writeWAVData(w io.Writer, samples []float64, bitDepth int) error {
	for _, sample := range samples {
		switch bitDepth {
		case 8:
			value := uint8((sample + 1.0) * 127.5)
			if err := binary.Write(w, binary.LittleEndian, value); err != nil {
				return err
			}
		case 16:
			value := int16(sample * 32767.0)
			if err := binary.Write(w, binary.LittleEndian, value); err != nil {
				return err
			}
		case 24:
			value := int32(sample * 8388607.0)
			bytes := []byte{
				byte(value),
				byte(value >> 8),
				byte(value >> 16),
			}
			if _, err := w.Write(bytes); err != nil {
				return err
			}
		case 32:
			value := int32(sample * 2147483647.0)
			if err := binary.Write(w, binary.LittleEndian, value); err != nil {
				return err
			}
		}
	}
	return nil
}
