package media

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Target struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Ext       string `json:"ext"`
	Mux       string `json:"mux"`
	Encoder   string `json:"encoder"`
	Lossless  bool   `json:"lossless"`
	Available bool   `json:"available"`
}

var Targets = []Target{
	{ID: "mp3", Name: "MP3 · 通用音频", Kind: "audio", Ext: "mp3", Mux: "mp3", Encoder: "libmp3lame"},
	{ID: "m4a", Name: "M4A · AAC 音频", Kind: "audio", Ext: "m4a", Mux: "ipod", Encoder: "aac"},
	{ID: "aac", Name: "AAC · ADTS 音频", Kind: "audio", Ext: "aac", Mux: "adts", Encoder: "aac"},
	{ID: "flac", Name: "FLAC · 无损音频", Kind: "audio", Ext: "flac", Mux: "flac", Encoder: "flac", Lossless: true},
	{ID: "wav", Name: "WAV · PCM 音频", Kind: "audio", Ext: "wav", Mux: "wav", Encoder: "pcm_s16le", Lossless: true},
	{ID: "alac", Name: "M4A · Apple Lossless", Kind: "audio", Ext: "m4a", Mux: "ipod", Encoder: "alac", Lossless: true},
	{ID: "aiff", Name: "AIFF · PCM 音频", Kind: "audio", Ext: "aiff", Mux: "aiff", Encoder: "pcm_s16be", Lossless: true},
	{ID: "ogg", Name: "OGG · Vorbis 音频", Kind: "audio", Ext: "ogg", Mux: "ogg", Encoder: "libvorbis"},
	{ID: "opus", Name: "Opus · 高效音频", Kind: "audio", Ext: "opus", Mux: "opus", Encoder: "libopus"},
	{ID: "wma", Name: "WMA · 兼容音频", Kind: "audio", Ext: "wma", Mux: "asf", Encoder: "wmav2"},
	{ID: "ac3", Name: "AC-3 · 杜比数字音频", Kind: "audio", Ext: "ac3", Mux: "ac3", Encoder: "ac3"},
	{ID: "wv", Name: "WavPack · 无损音频", Kind: "audio", Ext: "wv", Mux: "wv", Encoder: "wavpack", Lossless: true},
	{ID: "copy_audio", Name: "原码提取 · 不重编码", Kind: "audio", Ext: "auto", Mux: "auto", Lossless: true},
	{ID: "mp4", Name: "MP4 · H.264 / AAC", Kind: "video", Ext: "mp4", Mux: "mp4", Encoder: "libx264"},
	{ID: "mp4_hevc", Name: "MP4 · H.265 / AAC", Kind: "video", Ext: "mp4", Mux: "mp4", Encoder: "libx265"},
	{ID: "mp4_av1", Name: "MP4 · AV1 / AAC", Kind: "video", Ext: "mp4", Mux: "mp4", Encoder: "libsvtav1"},
	{ID: "mkv", Name: "MKV · H.264 / AAC", Kind: "video", Ext: "mkv", Mux: "matroska", Encoder: "libx264"},
	{ID: "webm", Name: "WebM · VP9 / Opus", Kind: "video", Ext: "webm", Mux: "webm", Encoder: "libvpx-vp9"},
	{ID: "webm_av1", Name: "WebM · AV1 / Opus", Kind: "video", Ext: "webm", Mux: "webm", Encoder: "libsvtav1"},
	{ID: "mov_prores", Name: "MOV · ProRes 422 HQ / PCM", Kind: "video", Ext: "mov", Mux: "mov", Encoder: "prores_ks"},
	{ID: "mkv_ffv1", Name: "MKV · FFV1 / PCM 归档", Kind: "video", Ext: "mkv", Mux: "matroska", Encoder: "ffv1", Lossless: true},
	{ID: "avi", Name: "AVI · MPEG-4 / MP3", Kind: "video", Ext: "avi", Mux: "avi", Encoder: "mpeg4"},
	{ID: "mpeg", Name: "MPG · MPEG-2 / MP2", Kind: "video", Ext: "mpg", Mux: "mpeg", Encoder: "mpeg2video"},
	{ID: "ts", Name: "TS · H.264 / AAC", Kind: "video", Ext: "ts", Mux: "mpegts", Encoder: "libx264"},
	{ID: "remux_mp4", Name: "MP4 · 原码重封装", Kind: "video", Ext: "mp4", Mux: "mp4", Lossless: true},
	{ID: "remux_mkv", Name: "MKV · 原码重封装", Kind: "video", Ext: "mkv", Mux: "matroska", Lossless: true},
	{ID: "png", Name: "PNG · 无损静态图片", Kind: "image", Ext: "png", Mux: "image2", Encoder: "png", Lossless: true},
	{ID: "jpg", Name: "JPEG · 通用静态图片", Kind: "image", Ext: "jpg", Mux: "image2", Encoder: "mjpeg"},
	{ID: "webp", Name: "WebP · 静态图片", Kind: "image", Ext: "webp", Mux: "image2", Encoder: "libwebp"},
	{ID: "webp_lossless", Name: "WebP · 无损静态图片", Kind: "image", Ext: "webp", Mux: "image2", Encoder: "libwebp", Lossless: true},
	{ID: "avif", Name: "AVIF · 静态图片", Kind: "image", Ext: "avif", Mux: "avif", Encoder: "libaom-av1"},
	{ID: "jxl", Name: "JPEG XL · 无损静态图片", Kind: "image", Ext: "jxl", Mux: "image2", Encoder: "libjxl", Lossless: true},
	{ID: "tiff", Name: "TIFF · 无损静态图片", Kind: "image", Ext: "tiff", Mux: "image2", Encoder: "tiff", Lossless: true},
	{ID: "bmp", Name: "BMP · 静态图片", Kind: "image", Ext: "bmp", Mux: "image2", Encoder: "bmp", Lossless: true},
	{ID: "gif", Name: "GIF · 动图（最多 256 色）", Kind: "image", Ext: "gif", Mux: "gif", Encoder: "gif"},
	{ID: "srt", Name: "SRT · 文本字幕", Kind: "subtitle", Ext: "srt", Mux: "srt", Encoder: "srt"},
	{ID: "vtt", Name: "WebVTT · 文本字幕", Kind: "subtitle", Ext: "vtt", Mux: "webvtt", Encoder: "webvtt"},
	{ID: "ass", Name: "ASS · 文本字幕", Kind: "subtitle", Ext: "ass", Mux: "ass", Encoder: "ass"},
}

func Catalog(e *Engine) []Target {
	out := append([]Target(nil), Targets...)
	for j := range out {
		t := &out[j]
		t.Available = e != nil && (t.Encoder == "" || e.Encoders[t.Encoder]) && (t.Mux == "auto" || e.Muxers[t.Mux])
	}
	return out
}
func FindTarget(id string) (Target, bool) {
	for _, t := range Targets {
		if t.ID == id {
			return t, true
		}
	}
	return Target{}, false
}

type Options struct {
	Target          string  `json:"target"`
	InputFormat     string  `json:"input_format"`
	Quality         string  `json:"quality"`
	AutoCopy        bool    `json:"auto_copy"`
	KeepMetadata    bool    `json:"keep_metadata"`
	Verify          bool    `json:"verify"`
	Checksum        bool    `json:"checksum"`
	AudioBitrate    int     `json:"audio_bitrate"`
	SampleRate      int     `json:"sample_rate"`
	BitDepth        int     `json:"bit_depth"`
	Channels        int     `json:"channels"`
	AudioStream     int     `json:"audio_stream"`
	SubtitleStream  int     `json:"subtitle_stream"`
	AllAudio        bool    `json:"all_audio"`
	Resolution      int     `json:"resolution"`
	FPS             int     `json:"fps"`
	Start           float64 `json:"start"`
	Duration        float64 `json:"duration"`
	Hardware        string  `json:"hardware"`
	Fallback        bool    `json:"fallback"`
	Threads         int     `json:"threads"`
	Collision       string  `json:"collision"`
	PreserveFolders bool    `json:"preserve_folders"`
	Suffix          string  `json:"suffix"`
}

func DefaultOptions() Options {
	return Options{Target: "mp3", InputFormat: "auto", Quality: "high", AutoCopy: true, KeepMetadata: true, Verify: true, Checksum: true, AudioStream: -1, SubtitleStream: -1, Hardware: "software", Fallback: true, Threads: 2, Collision: "rename"}
}
func (o Options) Validate() error {
	if _, ok := FindTarget(o.Target); !ok {
		return errors.New("未知输出格式")
	}
	if !ValidInputFormat(o.InputFormat) {
		return errors.New("不允许的输入格式")
	}
	switch o.Quality {
	case "high", "balanced", "small":
	default:
		return errors.New("未知质量预设")
	}
	switch o.Hardware {
	case "software", "nvenc", "qsv", "amf":
	default:
		return errors.New("未知硬件编码器")
	}
	if o.AudioBitrate < 0 || o.AudioBitrate > 512 || o.SampleRate < 0 || o.SampleRate > 384000 || o.Start < 0 || o.Duration < 0 || o.Start > 1e9 || o.Duration > 1e9 {
		return errors.New("音频参数或裁剪时间超出范围")
	}
	if o.SampleRate > 0 && o.SampleRate < 8000 {
		return errors.New("采样率至少为 8000 Hz")
	}
	if o.BitDepth != 0 && o.BitDepth != 16 && o.BitDepth != 24 && o.BitDepth != 32 {
		return errors.New("位深仅支持自动 / 16 / 24 / 32 bit")
	}
	if o.Channels != 0 && o.Channels != 1 && o.Channels != 2 && o.Channels != 6 && o.Channels != 8 {
		return errors.New("声道数仅支持原始 / 1 / 2 / 6 / 8")
	}
	if o.Threads < 1 || o.Threads > 16 || o.FPS < 0 || o.FPS > 240 || o.AudioStream < -1 || o.SubtitleStream < -1 {
		return errors.New("线程数、帧率或轨道编号不正确")
	}
	switch o.Resolution {
	case 0, 480, 720, 1080, 1440, 2160:
	default:
		return errors.New("未知分辨率限制")
	}
	if o.Collision != "rename" && o.Collision != "skip" {
		return errors.New("仅支持自动编号或跳过；不提供覆盖原文件")
	}
	if len(o.Suffix) > 80 || strings.ContainsAny(o.Suffix, "/\\:*?\"<>|\x00\r\n") {
		return errors.New("输出后缀含非法路径字符或过长")
	}
	return nil
}

type Plan struct {
	Args     []string `json:"args"`
	Ext      string   `json:"extension"`
	Mode     string   `json:"mode"`
	Warnings []string `json:"warnings"`
	Kind     string   `json:"kind"`
	Target   string   `json:"target"`
	Lossless bool     `json:"lossless"`
	Hardware bool     `json:"hardware"`
}

func (q *Plan) add(args ...string) { q.Args = append(q.Args, args...) }
func (q *Plan) warn(s string)      { q.Warnings = append(q.Warnings, s) }
func copyMux(codec string) (string, string) {
	switch codec {
	case "mp3":
		return "mp3", "mp3"
	case "aac", "alac":
		return "m4a", "ipod"
	case "flac":
		return "flac", "flac"
	case "opus":
		return "opus", "opus"
	case "vorbis":
		return "ogg", "ogg"
	case "ac3":
		return "ac3", "ac3"
	case "eac3":
		return "eac3", "eac3"
	case "dts":
		return "dts", "dts"
	case "wavpack":
		return "wv", "wv"
	case "wmav1", "wmav2":
		return "wma", "asf"
	}
	if strings.HasPrefix(codec, "pcm_") {
		return "wav", "wav"
	}
	return "mka", "matroska"
}
func codecFor(t string) string {
	switch t {
	case "mp3":
		return "mp3"
	case "m4a", "aac":
		return "aac"
	case "flac":
		return "flac"
	case "alac":
		return "alac"
	case "opus":
		return "opus"
	case "ogg":
		return "vorbis"
	case "wv":
		return "wavpack"
	case "wma":
		return "wmav2"
	case "ac3":
		return "ac3"
	}
	return ""
}
func noAudioTransform(o Options) bool {
	return o.Start == 0 && o.Duration == 0 && o.SampleRate == 0 && o.BitDepth == 0 && o.Channels == 0 && o.AudioBitrate == 0
}

func Build(e *Engine, i Info, source string, o Options) (Plan, error) {
	p := Plan{Warnings: []string{}, Mode: "transcode", Target: o.Target}
	if err := o.Validate(); err != nil {
		return p, err
	}
	if e == nil {
		return p, errors.New("转换引擎未安装")
	}
	t, _ := FindTarget(o.Target)
	p.Ext = t.Ext
	p.Kind = t.Kind
	p.Lossless = t.Lossless
	if t.Mux != "auto" && !e.Muxers[t.Mux] {
		return p, fmt.Errorf("当前引擎不支持 %s 容器，请安装完整引擎", t.Mux)
	}
	p.add("-hide_banner", "-nostdin", "-y", "-loglevel", "warning", "-progress", "pipe:1", "-nostats", "-threads", strconv.Itoa(o.Threads), "-filter_threads", "1", "-filter_complex_threads", "1")
	format := o.InputFormat
	if i.NCM {
		format = "auto"
	}
	p.Args = append(p.Args, InputArgs(source, format)...)
	if o.Start > 0 {
		p.add("-ss", strconv.FormatFloat(o.Start, 'f', 6, 64))
	}
	if o.Duration > 0 {
		p.add("-t", strconv.FormatFloat(o.Duration, 'f', 6, 64))
	}
	if i.Duration > 0 && o.Start >= i.Duration {
		return p, errors.New("开始时间已经超过媒体总时长")
	}
	if o.KeepMetadata {
		p.add("-map_metadata", "0")
		for _, key := range []string{"title", "artist", "album", "date", "track", "genre"} {
			if value := i.Tags[key]; value != "" {
				p.add("-metadata", key+"="+value)
			}
		}
	} else {
		p.add("-map_metadata", "-1")
	}
	if o.Start > 0 || o.Duration > 0 || !o.KeepMetadata {
		p.add("-map_chapters", "-1")
	}
	if i.HDR && (t.Kind == "video" || t.Kind == "image") && t.ID != "mkv_ffv1" && !strings.HasPrefix(t.ID, "remux_") {
		return p, errors.New("检测到 HDR：此版本不做未经验证的 HDR→SDR 或动态元数据重编码。请选择原码重封装或 FFV1 归档")
	}
	if strings.HasPrefix(t.ID, "remux_") {
		if _, ok := i.Video(); !ok || i.Kind == "image" {
			return p, errors.New("重封装需要视频文件")
		}
		if !noAudioTransform(o) || o.Resolution != 0 || o.FPS != 0 {
			return p, errors.New("原码重封装不能同时改变采样率、分辨率、帧率或裁剪范围")
		}
		if t.ID == "remux_mp4" {
			for _, s := range i.Streams {
				if s.Type == "video" && !strings.Contains("|h264|hevc|av1|mpeg4|vp9|mjpeg|png|", "|"+s.Codec+"|") {
					return p, fmt.Errorf("%s 视频不能安全原码封装为 MP4，请选 MKV 或转码", s.Codec)
				}
				if s.Type == "audio" && !strings.Contains("|aac|mp3|ac3|eac3|alac|", "|"+s.Codec+"|") {
					return p, fmt.Errorf("%s 音频不在本项目的 MP4 原码兼容表内，请选 MKV", s.Codec)
				}
				if s.Type == "subtitle" && s.Codec != "mov_text" {
					return p, errors.New("字幕不兼容 MP4 原码封装，请改用 MKV，或先单独提取字幕")
				}
			}
		}
		p.add("-map", "0", "-c", "copy")
		p.Mode = "stream-copy"
		p.warn("原码重封装保留所有可封装轨道；不兼容时报告失败，不偷偷丢弃轨道。")
	} else if t.Kind == "audio" {
		a, ok := i.Audio(o.AudioStream)
		if !ok {
			return p, errors.New("输入没有所选音频轨道，不能转换成音频")
		}
		p.add("-map", fmt.Sprintf("0:%d", a.Index), "-vn", "-sn", "-dn")
		if t.ID == "copy_audio" {
			if !noAudioTransform(o) {
				return p, errors.New("原码提取不能同时改变音频参数或裁剪")
			}
			p.Ext, t.Mux = copyMux(a.Codec)
			p.add("-c:a", "copy")
			p.Mode = "stream-copy"
		} else if o.AutoCopy && noAudioTransform(o) && codecFor(t.ID) == a.Codec {
			p.add("-c:a", "copy")
			p.Mode = "stream-copy"
			p.Lossless = true
			p.warn("已匹配原始编码，跳过重新压缩；质量预设不会改写原始码率。")
		} else {
			if err := audioEncode(&p, e, a, o, t.ID); err != nil {
				return p, err
			}
		}
		if t.ID == "mp3" {
			p.add("-id3v2_version", "3")
		}
		// Cover art is intentionally not remapped as video here. Exported as a sidecar by the runner.
		for _, s := range i.Streams {
			if s.Disposition["attached_pic"] == 1 {
				p.warn("含封面图片：音频保留通用文字标签；内嵌封面本版不保证迁移。可使用图片输出单独处理。")
			}
		}
	} else if t.Kind == "video" {
		v, ok := i.Video()
		if !ok || i.Kind == "image" {
			return p, errors.New("此输出需要视频输入；静态图片转视频不在本版范围")
		}
		p.add("-map", fmt.Sprintf("0:%d", v.Index))
		a, hasA := i.Audio(o.AudioStream)
		if o.AllAudio {
			p.add("-map", "0:a?")
		} else if hasA {
			p.add("-map", fmt.Sprintf("0:%d", a.Index))
		}
		p.add("-sn", "-dn")
		if len(i.Streams) > 2 {
			p.warn("普通视频转码不迁移字幕、附件和额外视频流；需要全部保留请选原码重封装。")
		}
		if t.ID == "mkv_ffv1" && (o.Resolution != 0 || o.FPS != 0 || o.SampleRate != 0 || o.Channels != 0 || o.BitDepth != 0) {
			return p, errors.New("FFV1 归档预设不允许修改分辨率、帧率或音频采样参数")
		}
		filters := []string{}
		if o.Resolution > 0 {
			w := o.Resolution * 16 / 9
			filters = append(filters, fmt.Sprintf("scale=w='min(iw,%d)':h='min(ih,%d)':force_original_aspect_ratio=decrease:force_divisible_by=2:flags=lanczos", w, o.Resolution))
		}
		if o.FPS > 0 {
			filters = append(filters, fmt.Sprintf("fps=%d", o.FPS))
		}
		if t.ID != "mkv_ffv1" && o.Resolution == 0 && (v.Width%2 != 0 || v.Height%2 != 0) {
			filters = append(filters, "pad=ceil(iw/2)*2:ceil(ih/2)*2")
			p.warn("输入为奇数尺寸，补齐一像素边缘以兼容视频编码器。")
		}
		if len(filters) > 0 {
			p.add("-vf", strings.Join(filters, ","))
		}
		if err := videoEncode(&p, e, v, o, t); err != nil {
			return p, err
		}
		if hasA || o.AllAudio {
			atype := "m4a"
			switch t.ID {
			case "webm", "webm_av1":
				atype = "opus"
			case "mov_prores", "mkv_ffv1":
				atype = "wav"
			case "avi":
				atype = "mp3"
			case "mpeg":
				atype = "mp2"
			}
			if o.AllAudio && atype == "wav" {
				p.add("-c:a", "pcm_f64le")
				p.warn("多音轨归档统一使用 64-bit 浮点 PCM，以容纳各轨道解码样本；不自动改变声道或采样率。")
			} else if err := audioEncode(&p, e, a, o, atype); err != nil {
				return p, err
			}
		}
		if t.ID == "mp4_hevc" {
			p.add("-tag:v", "hvc1")
		}
	} else if t.Kind == "image" {
		v, ok := i.Video()
		if !ok {
			return p, errors.New("输入没有可转换的画面")
		}
		p.add("-map", fmt.Sprintf("0:%d", v.Index), "-an", "-sn", "-dn")
		if !e.Encoders[t.Encoder] {
			return p, fmt.Errorf("缺少图片编码器 %s，请安装完整引擎", t.Encoder)
		}
		if t.ID == "gif" {
			fps := o.FPS
			if fps == 0 {
				fps = 15
			}
			width := 960
			if o.Resolution > 0 {
				width = o.Resolution * 16 / 9
			}
			p.add("-vf", fmt.Sprintf("fps=%d,scale='min(iw,%d)':-1:flags=lanczos,split[s0][s1];[s0]palettegen=stats_mode=diff[p];[s1][p]paletteuse=dither=sierra2_4a", fps, width), "-loop", "0")
			p.warn("GIF 使用调色板生成与抖动，颜色最多 256 种；不能做到原色无损。长视频建议设置输出时长。")
		} else {
			p.add("-frames:v", "1", "-c:v", t.Encoder)
			if i.Kind == "video" || i.Container == "gif" || i.Container == "apng" {
				p.warn("当前输出为静态图片，只取选定开始时间的一帧；动图请选择 GIF。")
			}
			switch t.ID {
			case "jpg":
				q := "2"
				if o.Quality == "balanced" {
					q = "4"
				}
				if o.Quality == "small" {
					q = "7"
				}
				p.add("-q:v", q)
				p.warn("JPEG 不保留透明度；ICC/EXIF 等图片元数据不保证完整迁移。")
			case "webp", "webp_lossless":
				p.add("-compression_level", "6")
				if t.ID == "webp_lossless" {
					p.add("-lossless", "1")
				} else {
					p.add("-quality", quality(o, "95", "85", "70"))
				}
			case "avif":
				p.add("-crf", quality(o, "18", "28", "38"), "-b:v", "0", "-cpu-used", "4", "-still-picture", "1")
				p.warn("此 AVIF 预设不保留透明通道；有透明背景请选 PNG 或无损 WebP。")
			case "jxl":
				p.add("-distance", "0", "-effort", "7")
			}
			if o.Resolution > 0 {
				p.add("-vf", fmt.Sprintf("scale='min(iw,%d)':-1:flags=lanczos", o.Resolution*16/9))
			}
			if t.Mux == "image2" {
				p.add("-update", "1")
			}
		}
	} else if t.Kind == "subtitle" {
		s, ok := i.Subtitle(o.SubtitleStream)
		if !ok {
			return p, errors.New("输入没有所选文本字幕轨道")
		}
		if strings.Contains(s.Codec, "pgs") || strings.Contains(s.Codec, "dvd") || s.Codec == "dvb_subtitle" {
			return p, errors.New("图像字幕需要 OCR，不能冒充文本字幕直接转换")
		}
		if o.Start > 0 || o.Duration > 0 {
			return p, errors.New("本版字幕转换不支持裁剪，请保持开始时间和时长为零")
		}
		p.add("-map", fmt.Sprintf("0:%d", s.Index), "-c:s", t.Encoder, "-vn", "-an", "-dn")
		p.warn("字幕转为 SRT/VTT 时复杂字体、排版和特效可能丢失。")
	}
	if t.Mux == "mp4" || t.Mux == "ipod" || t.Mux == "mov" {
		p.add("-movflags", "+faststart")
	}
	p.add("-threads", strconv.Itoa(o.Threads), "-f", t.Mux)
	return p, nil
}
func quality(o Options, high, balanced, small string) string {
	switch o.Quality {
	case "small":
		return small
	case "balanced":
		return balanced
	}
	return high
}
func audioEncode(p *Plan, e *Engine, s Stream, o Options, id string) error {
	encoder := ""
	switch id {
	case "mp3":
		encoder = "libmp3lame"
	case "m4a", "aac":
		encoder = "aac"
	case "flac":
		encoder = "flac"
	case "alac":
		encoder = "alac"
	case "wav":
		encoder = "pcm_s16le"
	case "aiff":
		encoder = "pcm_s16be"
	case "ogg":
		encoder = "libvorbis"
	case "opus":
		encoder = "libopus"
	case "wma":
		encoder = "wmav2"
	case "ac3":
		encoder = "ac3"
	case "wv":
		encoder = "wavpack"
	case "mp2":
		encoder = "mp2"
	}
	if encoder == "" || !e.Encoders[encoder] {
		return fmt.Errorf("缺少音频编码器 %s，请安装完整引擎", encoder)
	}
	depth := o.BitDepth
	if depth == 0 {
		depth = s.Depth()
		if depth == 0 {
			depth = 24
		}
	}
	floatInput := strings.Contains(s.SampleFmt, "flt") || strings.Contains(s.SampleFmt, "dbl")
	if id == "wav" || id == "aiff" {
		end := "le"
		if id == "aiff" {
			end = "be"
		}
		if o.BitDepth == 0 && floatInput {
			encoder = "pcm_f32" + end
			if strings.Contains(s.SampleFmt, "dbl") {
				encoder = "pcm_f64" + end
			}
		} else {
			if depth > 32 {
				depth = 32
			}
			if depth <= 16 {
				depth = 16
			} else if depth <= 24 {
				depth = 24
			}
			encoder = fmt.Sprintf("pcm_s%d%s", depth, end)
		}
	}
	if !e.Encoders[encoder] {
		return fmt.Errorf("当前引擎缺少 %s", encoder)
	}
	p.add("-c:a", encoder)
	switch id {
	case "mp3", "m4a", "aac", "opus", "wma", "ac3", "mp2":
		bitrate := o.AudioBitrate
		if bitrate == 0 {
			v := quality(o, "320", "192", "128")
			if id == "opus" {
				v = quality(o, "192", "128", "96")
			}
			if id == "wma" {
				v = quality(o, "192", "128", "96")
			}
			if id == "ac3" {
				v = "448"
			}
			bitrate, _ = strconv.Atoi(v)
		}
		if id == "mp3" && (bitrate > 320 || bitrate < 8) {
			return errors.New("MP3 码率须在 8–320 kbps 范围")
		}
		if id == "wma" && bitrate > 192 {
			return errors.New("WMA 兼容预设最大 192 kbps")
		}
		p.add("-b:a", strconv.Itoa(bitrate)+"k")
		if id == "opus" {
			p.add("-vbr", "on", "-compression_level", "10")
		}
	case "ogg":
		p.add("-q:a", quality(o, "8", "6", "4"))
	case "flac", "alac":
		if depth > 24 {
			depth = 24
			p.warn("此 FLAC/ALAC 预设最高 24-bit，输入的更高位深将量化；严格保留请选择 WAV 或 WavPack。")
		}
		if floatInput {
			p.warn("浮点解码样本将量化为整数无损格式；不是原始压缩文件的数学逆恢复。保留浮点样本请选择 WAV。")
		}
		if depth <= 16 {
			p.add("-sample_fmt", "s16")
		} else {
			p.add("-sample_fmt", "s32", "-bits_per_raw_sample", strconv.Itoa(depth))
		}
		if id == "flac" {
			p.add("-compression_level", "8")
		}
	case "wv":
		p.add("-compression_level", "4")
	}
	channels := o.Channels
	if (id == "mp3" || id == "wma" || id == "mp2") && (channels > 2 || channels == 0 && s.Channels > 2) {
		channels = 2
		p.warn("目标编码仅支持本项目的单声道/立体声预设，已将多声道下混为立体声。")
	}
	if channels > 0 {
		p.add("-ac", strconv.Itoa(channels))
	}
	rate := o.SampleRate
	originalRate, _ := strconv.Atoi(s.SampleRate)
	if id == "opus" && rate == 0 && originalRate != 48000 {
		rate = 48000
		p.warn("Opus 输出使用 48 kHz 时间基准 / 采样率。")
	}
	if (id == "mp3" || id == "aac" || id == "m4a" || id == "ac3" || id == "wma" || id == "mp2") && rate == 0 && originalRate > 48000 {
		rate = 48000
		p.warn("为兼容目标编码，高采样率输入重采样为 48 kHz。")
	}
	if id == "opus" && rate != 0 && rate != 8000 && rate != 12000 && rate != 16000 && rate != 24000 && rate != 48000 {
		return errors.New("Opus 支持的采样率为 8000/12000/16000/24000/48000 Hz")
	}
	if (id == "mp3" || id == "wma" || id == "ac3" || id == "mp2") && rate > 48000 {
		return errors.New("此目标编码的兼容预设最高支持 48 kHz；请降低采样率或选择 WAV/FLAC")
	}
	filters := []string{}
	if rate > 0 {
		p.add("-ar", strconv.Itoa(rate))
		if rate != originalRate {
			if e.Soxr {
				filters = append(filters, fmt.Sprintf("aresample=%d:resampler=soxr:precision=28", rate))
			} else {
				filters = append(filters, fmt.Sprintf("aresample=%d:resampler=swr:filter_size=64:exact_rational=1", rate))
				p.warn("当前引擎不含 SoXR，本次使用 SWR 重采样；需要 SoXR 请安装完整引擎。")
			}
		}
	}
	if o.BitDepth > 0 && o.BitDepth < s.Depth() && (id == "wav" || id == "aiff" || id == "flac" || id == "alac") {
		sampleFmt := "s32"
		if o.BitDepth == 16 {
			sampleFmt = "s16"
		}
		filters = append(filters, fmt.Sprintf("aresample=osf=%s:dither_method=triangular:output_sample_bits=%d", sampleFmt, o.BitDepth))
		p.warn("降低整数位深时使用三角抖动，减少量化失真。")
	}
	if len(filters) > 0 {
		p.add("-af", strings.Join(filters, ","))
	}
	return nil
}
func videoEncode(p *Plan, e *Engine, v Stream, o Options, t Target) error {
	enc := t.Encoder
	if !e.Encoders[enc] {
		return fmt.Errorf("当前引擎缺少 %s，请安装完整引擎", enc)
	}
	hwBase := ""
	switch enc {
	case "libx264":
		hwBase = "h264"
	case "libx265":
		hwBase = "hevc"
	case "libsvtav1":
		hwBase = "av1"
	}
	if o.Hardware != "software" {
		if hwBase == "" {
			return errors.New("此目标只开放软件编码，请将编码设备改为 CPU")
		}
		enc = hwBase + "_" + o.Hardware
		if !e.Encoders[enc] {
			if o.Fallback {
				p.warn("所选硬件编码器未编入引擎，已使用软件编码。 ")
				enc = t.Encoder
			} else {
				return fmt.Errorf("引擎没有 %s", enc)
			}
		} else {
			p.Hardware = true
		}
	}
	p.add("-c:v", enc)
	if p.Hardware {
		q := quality(o, "18", "23", "28")
		switch o.Hardware {
		case "nvenc":
			p.add("-preset", "p6", "-rc", "vbr", "-cq", q, "-b:v", "0")
		case "qsv":
			p.add("-global_quality", q, "-preset", "slow")
		case "amf":
			p.add("-quality", "quality", "-rc", "cqp", "-qp_i", q, "-qp_p", q)
		}
		p.warn("硬件编码优先速度；相同数值不代表与软件编码相同的画质。")
	} else {
		switch enc {
		case "libx264":
			p.add("-preset", "slow", "-crf", quality(o, "18", "23", "28"))
		case "libx265":
			p.add("-preset", "slow", "-crf", quality(o, "20", "26", "31"), "-x265-params", fmt.Sprintf("pools=%d:frame-threads=1", o.Threads))
		case "libsvtav1":
			p.add("-preset", "6", "-crf", quality(o, "22", "30", "38"), "-svtav1-params", fmt.Sprintf("lp=%d", o.Threads))
		case "libvpx-vp9":
			p.add("-crf", quality(o, "24", "32", "40"), "-b:v", "0", "-row-mt", "1", "-cpu-used", "2")
		case "prores_ks":
			p.add("-profile:v", "3", "-pix_fmt", "yuv422p10le")
			p.warn("ProRes 422 HQ 是高质量中间编码，不是数学无损，也不保留 Alpha 通道。")
		case "ffv1":
			p.add("-level", "3", "-coder", "1", "-context", "1", "-g", "1", "-slicecrc", "1")
			p.warn("FFV1 归档避免有损视频重压缩；字幕/附件和 HDR 动态元数据不保证迁移，完整封装保留请选原码重封装。")
		case "mpeg4", "mpeg2video":
			p.add("-q:v", quality(o, "2", "4", "7"))
		}
	}
	if enc != "ffv1" && enc != "prores_ks" {
		pix := "yuv420p"
		if (t.Encoder == "libx265" || t.Encoder == "libsvtav1") && (strings.Contains(v.PixFmt, "10") || strings.Contains(v.PixFmt, "12")) {
			pix = "yuv420p10le"
		}
		p.add("-pix_fmt", pix)
		if strings.Contains(v.PixFmt, "444") || strings.Contains(v.PixFmt, "422") || strings.Contains(v.PixFmt, "rgb") {
			p.warn("此交付预设使用 4:2:0 色度采样；需要保留原始色度精度请使用 FFV1 归档或原码重封装。")
		}
	}
	return nil
}
