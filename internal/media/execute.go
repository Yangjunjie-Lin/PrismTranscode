package media

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"prismtranscode/internal/platform"
	"strconv"
	"strings"
	"time"
)

type Progress struct {
	Phase   string  `json:"phase"`
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
}
type Result struct {
	Output       string   `json:"output"`
	Size         int64    `json:"size"`
	SHA256       string   `json:"sha256,omitempty"`
	Mode         string   `json:"mode"`
	Verified     bool     `json:"verified"`
	Verification string   `json:"verification"`
	Elapsed      float64  `json:"elapsed_seconds"`
	Engine       string   `json:"engine"`
	Warnings     []string `json:"warnings"`
	Command      []string `json:"command"`
	Log          string   `json:"log,omitempty"`
	OutputInfo   Info     `json:"output_info"`
}

var ErrSkipped = errors.New("目标同名文件已存在，按设置跳过")

func Execute(ctx context.Context, e *Engine, i Info, source, outDir, stem string, o Options, update func(Progress)) (Result, error) {
	started := time.Now()
	r := Result{Engine: e.Version, Warnings: []string{}}
	p, err := Build(e, i, source, o)
	if err != nil {
		return r, err
	}
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return r, err
	}
	stem = cleanStem(stem) + o.Suffix
	if o.Collision == "skip" {
		if _, err = os.Lstat(filepath.Join(outDir, stem+"."+p.Ext)); err == nil {
			return r, ErrSkipped
		}
	}
	f, err := os.CreateTemp(outDir, ".prism-output-*."+p.Ext)
	if err != nil {
		return r, err
	}
	tmp := f.Name()
	f.Close()
	defer os.Remove(tmp)
	progress := func(phase string, v float64, speed string) {
		if update != nil {
			if v < 0 {
				v = 0
			}
			if v > 100 {
				v = 100
			}
			update(Progress{phase, v, speed})
		}
	}
	duration := i.Duration - o.Start
	if o.Duration > 0 && (duration <= 0 || o.Duration < duration) {
		duration = o.Duration
	}
	args := append(append([]string(nil), p.Args...), tmp)
	r.Command = append([]string{e.FFmpeg}, args...)
	progress("正在转换", 1, "")
	log, err := runFFmpeg(ctx, e.FFmpeg, args, duration, func(v float64, s string) { progress("正在转换", 2+v*.80, s) })
	if err != nil && p.Hardware && o.Fallback && ctx.Err() == nil {
		p.warn("硬件编码执行失败，已自动重试软件编码。硬件错误：" + truncate(log, 1600))
		o.Hardware = "software"
		prior := append([]string(nil), p.Warnings...)
		p, err = Build(e, i, source, o)
		if err == nil {
			p.Warnings = append(prior, p.Warnings...)
			args = append(append([]string(nil), p.Args...), tmp)
			r.Command = append([]string{e.FFmpeg}, args...)
			log, err = runFFmpeg(ctx, e.FFmpeg, args, duration, func(v float64, s string) { progress("CPU 回退转换", 2+v*.80, s) })
		}
	}
	r.Log = log
	r.Mode = p.Mode
	r.Warnings = append(r.Warnings, p.Warnings...)
	r.Warnings = append(r.Warnings, i.Warnings...)
	if err != nil {
		if ctx.Err() != nil {
			return r, ctx.Err()
		}
		return r, fmt.Errorf("转换失败：%w\n%s", err, truncate(log, 5000))
	}
	if ctx.Err() != nil {
		return r, ctx.Err()
	}
	progress("检查输出结构", 84, "")
	info, err := Probe(ctx, e, tmp, "auto")
	if err != nil {
		return r, fmt.Errorf("输出文件无法重新识别：%w", err)
	}
	if p.Kind == "audio" {
		if _, ok := info.Audio(-1); !ok {
			return r, errors.New("输出校验失败：没有音频流")
		}
	} else if p.Kind == "video" || p.Kind == "image" {
		if _, ok := info.Video(); !ok {
			return r, errors.New("输出校验失败：没有画面流")
		}
	} else {
		if _, ok := info.Subtitle(-1); !ok {
			return r, errors.New("输出校验失败：没有字幕流")
		}
	}
	if p.Kind == "audio" || p.Kind == "video" {
		if duration > 0 && info.Duration > 0 {
			delta := duration - info.Duration
			if delta < 0 {
				delta = -delta
			}
			tolerance := 2.0
			if duration*.02 > tolerance {
				tolerance = duration * .02
			}
			if delta > tolerance {
				return r, fmt.Errorf("输出时长异常：预期 %.3f 秒，实际 %.3f 秒", duration, info.Duration)
			}
		}
	}
	r.Verification = "ffprobe 结构与时长检查"
	if o.Verify && p.Kind != "subtitle" {
		progress("完整解码校验", 88, "")
		detect := "explode"
		if v, ok := info.Video(); ok && v.Codec == "tiff" {
			// FFmpeg 7.x's TIFF decoder rejects valid ancillary ResolutionUnit
			// tag 296 in explode mode, including its own encoder's output.
			// Keep full-frame decoding and -xerror, with explicit integrity checks.
			detect = "crccheck+bitstream+buffer"
			r.Warnings = append(r.Warnings, "TIFF 使用完整像素解码及 CRC/位流/缓冲区错误检查；不将解码器不支持的附属标签直接视为像素损坏。")
		}
		va := []string{"-hide_banner", "-nostdin", "-v", "error", "-xerror", "-err_detect", detect, "-threads", strconv.Itoa(o.Threads)}
		va = append(va, InputArgs(tmp, "auto")...)
		va = append(va, "-map", "0:v?", "-map", "0:a?", "-f", "null", "-")
		if _, err = Capture(ctx, e.FFmpeg, va...); err != nil {
			return r, fmt.Errorf("输出完整解码校验失败：%w", err)
		}
		r.Verified = true
		r.Verification = "ffprobe 结构/时长 + 全文件音视频解码"
	} else if o.Verify {
		r.Verified = true
		r.Verification = "文本字幕重新解析（不等于样式一致性校验）"
	}
	if o.Checksum {
		progress("计算 SHA-256", 96, "")
		r.SHA256, err = HashFile(ctx, tmp)
		if err != nil {
			return r, err
		}
	}
	st, err := os.Stat(tmp)
	if err != nil || st.Size() == 0 {
		return r, errors.New("输出为空")
	}
	r.Size = st.Size()
	if err = ctx.Err(); err != nil {
		return r, err
	}
	path, atomic, err := commit(ctx, tmp, outDir, stem, p.Ext, o.Collision)
	if err != nil {
		return r, err
	}
	if !atomic {
		r.Warnings = append(r.Warnings, "目标文件系统不支持硬链接，使用排他创建写入完成文件；不会覆盖已有文件。")
	}
	r.Output = path
	info.Path = path
	info.Name = filepath.Base(path)
	r.OutputInfo = info
	r.Elapsed = time.Since(started).Seconds()
	progress("已完成", 100, "")
	return r, nil
}
func runFFmpeg(ctx context.Context, program string, args []string, duration float64, progress func(float64, string)) (string, error) {
	cmd := exec.CommandContext(ctx, program, args...)
	platform.HideConsole(cmd)
	stderr := &boundedBuffer{limit: 128 << 10}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", err
	}
	scan := bufio.NewScanner(stdout)
	scan.Buffer(make([]byte, 4096), 1<<20)
	pct := 0.0
	speed := ""
	for scan.Scan() {
		line := scan.Text()
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch k {
		case "out_time_us":
			n, _ := strconv.ParseFloat(v, 64)
			if duration > 0 {
				pct = n / 1e6 / duration * 100
			}
		case "speed":
			speed = v
		case "progress":
			if v == "end" {
				pct = 100
			}
			if progress != nil {
				progress(pct, speed)
			}
		}
	}
	se := scan.Err()
	if se != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	err = cmd.Wait()
	if err == nil {
		err = se
	}
	return stderr.b.String(), err
}
func HashFile(ctx context.Context, path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 256<<10)
	for {
		if e = ctx.Err(); e != nil {
			return "", e
		}
		n, re := f.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
		}
		if re == io.EOF {
			break
		}
		if re != nil {
			return "", re
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[len(r)-n:])
	}
	return s
}
func cleanStem(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune("/\\:*?\"<>|", r) {
			return '_'
		}
		return r
	}, s)
	s = strings.Trim(s, " .")
	if s == "" {
		s = "converted"
	}
	r := []rune(s)
	if len(r) > 150 {
		s = string(r[:150])
	}
	switch strings.ToUpper(strings.Split(s, ".")[0]) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		s = "_" + s
	}
	return s
}

// A hard link on the same volume atomically commits without replacing a path.
// FAT/network fallback uses O_EXCL and deletes only its own partial output.
func commit(ctx context.Context, tmp, dir, stem, ext, collision string) (string, bool, error) {
	for n := 0; n < 10000; n++ {
		base := stem
		if n > 0 {
			base = fmt.Sprintf("%s (%d)", stem, n)
		}
		path := filepath.Join(dir, base+"."+ext)
		err := os.Link(tmp, path)
		if err == nil {
			return path, true, nil
		}
		if _, e := os.Lstat(path); e == nil {
			if collision == "skip" {
				return "", false, ErrSkipped
			}
			continue
		}
		out, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if errors.Is(e, os.ErrExist) {
			if collision == "skip" {
				return "", false, ErrSkipped
			}
			continue
		}
		if e != nil {
			return "", false, e
		}
		in, e := os.Open(tmp)
		if e == nil {
			buf := make([]byte, 256<<10)
			for {
				if e = ctx.Err(); e != nil {
					break
				}
				n, re := in.Read(buf)
				if n > 0 {
					_, e = out.Write(buf[:n])
					if e != nil {
						break
					}
				}
				if re == io.EOF {
					break
				}
				if re != nil {
					e = re
					break
				}
			}
			in.Close()
		}
		if e == nil {
			e = out.Sync()
		}
		ce := out.Close()
		if e == nil {
			e = ce
		}
		if e != nil {
			os.Remove(path)
			return "", false, e
		}
		return path, false, nil
	}
	return "", false, errors.New("同名输出超过 10000 个，请更换文件名或目录")
}
