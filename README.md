# Godub

Goudub 是一个 Go 语言实现的音频处理库，是 Python [pydub](https://github.com/jiaaro/pydub) 库的 Go 语言版本。

## 功能特性

- 音频文件的读取和写入（支持多种格式：wav, mp3, ogg, flv, etc.）
- 音频切片和拼接
- 音频效果处理（淡入淡出、音量调节等）
- 音频格式转换
- 音频分析工具

## 安装

```bash
go get github.com/haichen-zhang/goudub
```

## 依赖要求

- Go 1.21 或更高版本
- FFmpeg（用于音频格式转换和某些高级功能）

## API 文档

### 音频加载

```go
// 从文件加载音频
sound, err := audio.FromFile("test.mp3")

// 从特定格式加载
sound, err := audio.FromMP3("test.mp3")
sound, err := audio.FromWAV("test.wav")
sound, err := audio.FromOGG("test.ogg")
sound, err := audio.FromFLV("test.flv")
```

### 音频切片

```go
// 切片音频（参数单位为毫秒）
segment, err := sound.Slice(0 * time.Second, 30 * time.Second)
```

### 音频效果

```go
// 淡入效果
processed, err := effects.FadeIn(sound, 2 * time.Second)

// 淡出效果
processed, err := effects.FadeOut(sound, 3 * time.Second)

// 音量标准化
processed, err := effects.Normalize(sound)

// 调整音量（单位：分贝）
processed, err := effects.AdjustVolume(sound, 6.0)  // 增加6dB
processed, err := effects.AdjustVolume(sound, -6.0) // 降低6dB
```

### 音频导出

```go
// 导出为MP3
err := converter.SaveAudioFile(sound.Samples(), sound.SampleRate(), sound.Channels(), sound.BitDepth(), "output.mp3", "mp3")
```

## 示例

### 基本音频处理

```go
package main

import (
	"github.com/HiChen85/godub/pkg/converter"
	"time"
	"github.com/HiChen85/godub/pkg/audio"
	"github.com/HiChen85/godub/pkg/effects"
)

func main() {
	// 加载音频文件
	sound, err := audio.FromFile("input.mp3")
	if err != nil {
		panic(err)
	}

	// 处理音频
	processed, err := effects.FadeIn(sound, 2*time.Second) // 2秒淡入
	if err != nil {
		panic(err)
	}

	processed, err = effects.FadeOut(processed, 3*time.Second) // 3秒淡出
	if err != nil {
		panic(err)
	}

	// 导出处理后的音频
	err = converter.SaveAudioFile(
		processed.Samples(),
		processed.SampleRate(),
		processed.Channels(),
		processed.BitDepth(),
		"output.mp3",
		"mp3",
	)
	if err != nil {
		panic(err)
	}
}
```

### 音频切片和拼接

```go
package main

import (
	"github.com/HiChen85/godub/pkg/converter"
	"time"
	"github.com/HiChen85/godub/pkg/audio"
)

func main() {
	// 加载音频文件
	sound, err := audio.FromFile("song.mp3")
	if err != nil {
		panic(err)
	}

	// 获取前30秒
	first30s, err := sound.Slice(0, 30*time.Second)
	if err != nil {
		panic(err)
	}

	// 导出切片
	err = converter.SaveAudioFile(
		first30s.Samples(),
		first30s.SampleRate(),
		first30s.Channels(),
		first30s.BitDepth(),
		"first_30s.mp3",
		"mp3",
	)
	if err != nil {
		panic(err)
	}
}
```

## 许可证

MIT License
