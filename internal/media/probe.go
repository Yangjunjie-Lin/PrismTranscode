package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"prismtranscode/internal/ncm"
	"strconv"
	"strings"
	"time"
)

type Stream struct {
	Index          int               `json:"index"`
	Codec          string            `json:"codec_name"`
	Type           string            `json:"codec_type"`
	SampleRate     string            `json:"sample_rate"`
	Channels       int               `json:"channels"`
	Layout         string            `json:"channel_layout"`
	Bits           int               `json:"bits_per_sample"`
	RawBits        string            `json:"bits_per_raw_sample"`
	SampleFmt      string            `json:"sample_fmt"`
	Width          int               `json:"width"`
	Height         int               `json:"height"`
	PixFmt         string            `json:"pix_fmt"`
	FPS            string            `json:"avg_frame_rate"`
	ColorSpace     string            `json:"color_space"`
	ColorTransfer  string            `json:"color_transfer"`
	ColorPrimaries string            `json:"color_primaries"`
	Duration       string            `json:"duration"`
	Tags           map[string]string `json:"tags"`
	Disposition    map[string]int    `json:"disposition"`
}
type Info struct {
	Path      string            `json:"path"`
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`
	Container string            `json:"container"`
	Size      int64             `json:"size"`
	Duration  float64           `json:"duration"`
	Bitrate   string            `json:"bitrate"`
	Streams   []Stream          `json:"streams"`
	Tags      map[string]string `json:"tags"`
	Warnings  []string          `json:"warnings"`
	NCM       bool              `json:"ncm"`
	HDR       bool              `json:"hdr"`
}

func (s Stream) Depth() int {
	if n, _ := strconv.Atoi(s.RawBits); n > 0 {
		return n
	}
	if s.Bits > 0 {
		return s.Bits
	}
	if strings.Contains(s.SampleFmt, "s16") {
		return 16
	}
	if strings.Contains(s.SampleFmt, "s32") {
		return 32
	}
	return 0
}
func (i Info) Audio(index int) (Stream, bool) {
	var first *Stream
	for _, s := range i.Streams {
		if s.Type == "audio" {
			ss := s
			if first == nil {
				first = &ss
			}
			if index >= 0 && s.Index == index {
				return s, true
			}
		}
	}
	if index < 0 && first != nil {
		return *first, true
	}
	return Stream{}, false
}
func (i Info) Video() (Stream, bool) {
	for _, s := range i.Streams {
		if s.Type == "video" && s.Disposition["attached_pic"] != 1 {
			return s, true
		}
	}
	return Stream{}, false
}
func (i Info) Subtitle(index int) (Stream, bool) {
	for _, s := range i.Streams {
		if s.Type == "subtitle" && (index < 0 || s.Index == index) {
			return s, true
		}
	}
	return Stream{}, false
}
func ValidInputFormat(f string) bool {
	switch f {
	case "", "auto", "ncm", "mp3", "flac", "wav", "aac", "ogg", "mov", "matroska", "avi", "mpegts", "mpeg", "srt", "webvtt", "ass":
		return true
	}
	return false
}
func IsNCM(path string) bool {
	f, e := os.Open(path)
	if e != nil {
		return false
	}
	defer f.Close()
	var b [8]byte
	_, e = io.ReadFull(f, b[:])
	return e == nil && string(b[:]) == "CTENFDAM"
}

// Prepare decrypts an NCM to a session-private, bounded-memory spool file.
// All other files are passed by path and never duplicated in memory.
func Prepare(ctx context.Context, path, cache string) (string, *ncm.Metadata, func(), error) {
	if !IsNCM(path) {
		return path, nil, func() {}, nil
	}
	c, e := ncm.Open(path)
	if e != nil {
		return "", nil, nil, e
	}
	defer c.Close()
	f, e := os.CreateTemp(cache, "ncm-*.media")
	if e != nil {
		return "", nil, nil, e
	}
	name := f.Name()
	clean := func() { os.Remove(name) }
	e = c.Extract(ctx, f, nil)
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		clean()
		return "", nil, nil, e
	}
	if _, e = ncm.Detect(name); e != nil {
		clean()
		return "", nil, nil, e
	}
	return name, &c.Metadata, clean, nil
}
func Probe(ctx context.Context, e *Engine, path, format string) (Info, error) {
	var out Info
	if e == nil {
		return out, errors.New("请先安装 FFmpeg 引擎")
	}
	if !ValidInputFormat(format) {
		return out, errors.New("不支持的强制输入格式")
	}
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() {
		return out, errors.New("输入不是可读取的普通文件")
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	args := []string{"-v", "error", "-show_format", "-show_streams", "-of", "json"}
	args = append(args, InputArgs(path, format)...)
	b, err := Capture(ctx, e.FFprobe, args...)
	if err != nil {
		return out, err
	}
	var raw struct {
		Streams []Stream `json:"streams"`
		Format  struct {
			FormatName string            `json:"format_name"`
			Duration   string            `json:"duration"`
			BitRate    string            `json:"bit_rate"`
			Tags       map[string]string `json:"tags"`
		} `json:"format"`
	}
	if err = json.Unmarshal(b, &raw); err != nil {
		return out, fmt.Errorf("无法读取媒体信息：%w", err)
	}
	if len(raw.Streams) == 0 {
		return out, errors.New("文件内没有可识别的音频、视频、图片或字幕流")
	}
	out = Info{Path: path, Name: filepath.Base(path), Size: st.Size(), Streams: raw.Streams, Container: raw.Format.FormatName, Bitrate: raw.Format.BitRate, Tags: raw.Format.Tags, Warnings: []string{}}
	out.Duration, _ = strconv.ParseFloat(raw.Format.Duration, 64)
	if v, ok := out.Video(); ok {
		out.Kind = "video"
		out.HDR = v.ColorTransfer == "smpte2084" || v.ColorTransfer == "arib-std-b67"
		if imageContainer(out.Container) || stillImageBrand(raw.Format.Tags["major_brand"]) {
			out.Kind = "image"
		}
	} else if _, ok := out.Audio(-1); ok {
		out.Kind = "audio"
	} else {
		out.Kind = "subtitle"
	}
	if out.HDR {
		out.Warnings = append(out.Warnings, "检测到 HDR；普通有损转码将被阻止。请使用原码重封装或 FFV1 归档，避免未经处理的色彩失真。")
	}
	for _, s := range out.Streams {
		if strings.Contains(s.Codec, "dvd_subtitle") || strings.Contains(s.Codec, "pgs") {
			out.Warnings = append(out.Warnings, "包含图像字幕，不能直接转成 SRT/VTT 文本。")
		}
	}
	return out, nil
}
func stillImageBrand(brand string) bool {
	switch strings.TrimSpace(strings.ToLower(brand)) {
	case "avif", "heic", "heix", "hevc", "hevx", "mif1":
		return true
	}
	return false
}
func imageContainer(f string) bool {
	for _, v := range []string{"image2", "png_pipe", "jpeg_pipe", "webp_pipe", "bmp_pipe", "tiff_pipe", "jxl_pipe", "gif", "apng", "avif"} {
		if strings.Contains(f, v) {
			return true
		}
	}
	return false
}
func KnownExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return strings.Contains("|ncm|mp3|flac|wav|wave|m4a|aac|alac|aiff|aif|ogg|opus|wma|ape|wv|amr|ac3|eac3|dts|mka|caf|mp2|mp4|mkv|mov|m4v|webm|avi|flv|wmv|asf|ts|mts|m2ts|mpg|mpeg|vob|3gp|3g2|mxf|rm|rmvb|ogv|hevc|h264|h265|av1|jpg|jpeg|png|webp|avif|jxl|bmp|tif|tiff|gif|apng|heic|heif|jp2|j2k|tga|exr|svg|srt|vtt|ass|ssa|", "|"+strings.TrimPrefix(ext, ".")+"|")
}
