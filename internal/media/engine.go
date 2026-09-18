// Package media contains detection, validated conversion planning and execution.
// FFmpeg is invoked directly with argument arrays, never through a shell.
package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"prismtranscode/internal/platform"
	"runtime"
	"strings"
	"time"
)

type Engine struct {
	FFmpeg   string          `json:"ffmpeg"`
	FFprobe  string          `json:"ffprobe"`
	Version  string          `json:"version"`
	Encoders map[string]bool `json:"encoders"`
	Muxers   map[string]bool `json:"muxers"`
	Filters  map[string]bool `json:"filters"`
	Soxr     bool            `json:"soxr"`
}

func Discover(configured, appDir string) (*Engine, error) {
	if strings.TrimSpace(configured) != "" {
		return Select(configured)
	}
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	candidates := []string{configured, filepath.Join(appDir, "tools", name)}
	if p, err := exec.LookPath(name); err == nil {
		candidates = append(candidates, p)
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if s, err := os.Stat(p); err == nil && s.IsDir() {
			p = filepath.Join(p, name)
		}
		if s, err := os.Stat(p); err != nil || !s.Mode().IsRegular() {
			continue
		}
		probe := filepath.Join(filepath.Dir(p), strings.Replace(name, "ffmpeg", "ffprobe", 1))
		if s, err := os.Stat(probe); err != nil || !s.Mode().IsRegular() {
			continue
		}
		e, err := Load(p, probe)
		if err == nil {
			return e, nil
		}
	}
	return nil, errors.New("未找到可用的 FFmpeg + FFprobe；请先安装引擎，或选择同时含有两个程序的 bin 文件夹")
}

// Select loads only the explicitly chosen installation; it never silently falls
// back to a different engine found on PATH.
func Select(path string) (*Engine, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("请选择 FFmpeg 或包含 FFmpeg 与 FFprobe 的文件夹")
	}
	if s, err := os.Stat(path); err == nil && s.IsDir() {
		name := "ffmpeg"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		path = filepath.Join(path, name)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	name := "ffprobe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return Load(path, filepath.Join(filepath.Dir(path), name))
}
func Load(ffmpeg, ffprobe string) (*Engine, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	b, err := Capture(ctx, ffmpeg, "-version")
	if err != nil {
		return nil, err
	}
	e := &Engine{FFmpeg: ffmpeg, FFprobe: ffprobe, Version: strings.SplitN(string(b), "\n", 2)[0], Soxr: strings.Contains(string(b), "--enable-libsoxr")}
	b, err = Capture(ctx, ffprobe, "-version")
	if err != nil || !bytes.Contains(b, []byte("ffprobe version")) {
		return nil, errors.New("ffprobe 无法运行")
	}
	for _, r := range []struct {
		flag string
		dst  *map[string]bool
	}{{"-encoders", &e.Encoders}, {"-muxers", &e.Muxers}, {"-filters", &e.Filters}} {
		b, err = Capture(ctx, ffmpeg, "-hide_banner", r.flag)
		if err != nil {
			return nil, err
		}
		*r.dst = map[string]bool{}
		for _, line := range strings.Split(string(b), "\n") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				for _, n := range strings.Split(f[1], ",") {
					(*r.dst)[n] = true
				}
			}
		}
	}
	return e, nil
}

// boundedBuffer keeps subprocess logs bounded, even for damaged media.
type boundedBuffer struct {
	b     bytes.Buffer
	limit int
}

func (w *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	left := w.limit - w.b.Len()
	if left > 0 {
		if len(p) > left {
			p = p[:left]
		}
		w.b.Write(p)
	}
	return n, nil
}
func Capture(ctx context.Context, program string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, program, args...)
	platform.HideConsole(cmd)
	out := &boundedBuffer{limit: 8 << 20}
	errout := &boundedBuffer{limit: 64 << 10}
	cmd.Stdout = out
	cmd.Stderr = errout
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("引擎执行失败：%v\n%s", err, errout.b.String())
	}
	return out.b.Bytes(), nil
}
func InputArgs(path, format string) []string {
	a := []string{"-protocol_whitelist", "file,pipe"}
	if format != "" && format != "auto" && format != "ncm" {
		a = append(a, "-f", format)
	}
	return append(a, "-i", path)
}
