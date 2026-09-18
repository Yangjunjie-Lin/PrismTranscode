// Package installer downloads an external engine only after explicit user action.
// It verifies distributor SHA-256, stages both binaries, and keeps a receipt.
package installer

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxDownload int64 = 768 << 20
const fullURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
const checksumURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/checksums.sha256"
const essentialsURL = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"

type Receipt struct {
	Source      string `json:"source"`
	Archive     string `json:"archive"`
	SHA256      string `json:"sha256"`
	InstalledAt string `json:"installed_at"`
}

func allowedHost(h string) bool {
	return h == "github.com" || h == "api.github.com" || h == "www.gyan.dev" || h == "gyan.dev" || h == "release-assets.githubusercontent.com" || h == "objects.githubusercontent.com" || h == "github-releases.githubusercontent.com"
}
func safeURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && u.Scheme == "https" && allowedHost(u.Hostname())
}
func Checksum(text, name string) (string, error) {
	for _, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		hash := f[0]
		if len(hash) != 64 {
			continue
		}
		if _, e := hex.DecodeString(hash); e != nil {
			continue
		}
		if name != "" && (len(f) < 2 || strings.TrimPrefix(f[len(f)-1], "*") != name) {
			continue
		}
		return strings.ToLower(hash), nil
	}
	return "", errors.New("没有找到与引擎包匹配的 SHA-256；不会安装未校验文件")
}
func Install(ctx context.Context, source, cache string, update func(string, float64)) (string, error) {
	archive, checks, name := fullURL, checksumURL, "ffmpeg-master-latest-win64-gpl.zip"
	if source == "essentials" {
		archive = essentialsURL
		checks = essentialsURL + ".sha256"
		name = ""
	} else if source != "full" {
		return "", errors.New("未知引擎来源")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) > 8 || !safeURL(r.URL.String()) {
			return errors.New("拒绝非白名单或非 HTTPS 下载重定向")
		}
		return nil
	}}
	get := func(u string) (*http.Response, error) {
		if !safeURL(u) {
			return nil, errors.New("不允许的下载地址")
		}
		req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
		if e != nil {
			return nil, e
		}
		req.Header.Set("User-Agent", "PrismTranscode/2.1.0-beta.1")
		r, e := client.Do(req)
		if e != nil {
			return nil, e
		}
		if r.StatusCode != 200 {
			r.Body.Close()
			return nil, fmt.Errorf("引擎服务器 HTTP %d", r.StatusCode)
		}
		return r, nil
	}
	update("读取发行方 SHA-256", 1)
	resp, e := get(checks)
	if e != nil {
		return "", e
	}
	data, e := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	resp.Body.Close()
	if e != nil {
		return "", e
	}
	expected, e := Checksum(string(data), name)
	if e != nil {
		return "", e
	}
	resp, e = get(archive)
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()
	if resp.ContentLength > MaxDownload {
		return "", errors.New("引擎包超过安全大小限制")
	}
	stage, e := os.MkdirTemp(cache, "engine-download-")
	if e != nil {
		return "", e
	}
	defer os.RemoveAll(stage)
	zipPath := filepath.Join(stage, "engine.zip")
	f, e := os.Create(zipPath)
	if e != nil {
		return "", e
	}
	hash := sha256.New()
	writer := io.MultiWriter(f, hash)
	reader := io.LimitReader(resp.Body, MaxDownload+1)
	buf := make([]byte, 256<<10)
	var done int64
	last := time.Time{}
	for {
		if e = ctx.Err(); e != nil {
			break
		}
		n, re := reader.Read(buf)
		if n > 0 {
			_, e = writer.Write(buf[:n])
			done += int64(n)
			if e != nil {
				break
			}
		}
		if time.Since(last) > 250*time.Millisecond {
			p := 10.0
			if resp.ContentLength > 0 {
				p = 5 + 85*float64(done)/float64(resp.ContentLength)
			}
			update(fmt.Sprintf("下载引擎 %.1f MB", float64(done)/(1<<20)), p)
			last = time.Now()
		}
		if re == io.EOF {
			break
		}
		if re != nil {
			e = re
			break
		}
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return "", e
	}
	if done > MaxDownload {
		return "", errors.New("引擎包超过安全大小限制")
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return "", errors.New("SHA-256 不匹配，已丢弃下载。发行方更新期间可重试")
	}
	update("校验通过，解压两个引擎程序", 93)
	target := filepath.Join(cache, "engines", actual[:16])
	if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
		return "", e
	}
	content := filepath.Join(stage, "content")
	if e = Extract(zipPath, content); e != nil {
		return "", e
	}
	receipt := Receipt{Source: source, Archive: archive, SHA256: actual, InstalledAt: time.Now().Format(time.RFC3339)}
	b, _ := json.MarshalIndent(receipt, "", "  ")
	if e = os.WriteFile(filepath.Join(content, "engine-receipt.json"), b, 0644); e != nil {
		return "", e
	}
	if _, e = os.Stat(filepath.Join(target, "ffmpeg.exe")); e == nil {
		return filepath.Join(target, "ffmpeg.exe"), nil
	}
	if e = os.Rename(content, target); e != nil {
		return "", e
	}
	update("引擎安装完成", 100)
	return filepath.Join(target, "ffmpeg.exe"), nil
}
func Extract(path, dest string) error {
	z, e := zip.OpenReader(path)
	if e != nil {
		return e
	}
	defer z.Close()
	if e = os.MkdirAll(dest, 0700); e != nil {
		return e
	}
	found := map[string]bool{}
	for _, f := range z.File {
		normal := strings.ReplaceAll(f.Name, "\\", "/")
		base := strings.ToLower(filepath.Base(normal))
		wantExe := base == "ffmpeg.exe" || base == "ffprobe.exe"
		wantDoc := base == "license" || base == "license.txt" || base == "readme.txt" || base == "license.md"
		if !wantExe && !wantDoc {
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 || f.FileInfo().IsDir() {
			return errors.New("引擎压缩包含不安全的文件类型")
		}
		if found[base] {
			return errors.New("引擎压缩包含重复文件")
		}
		found[base] = true
		limit := int64(512 << 20)
		if wantDoc {
			limit = 4 << 20
		}
		if f.UncompressedSize64 > uint64(limit) {
			return errors.New("解压文件超过大小限制")
		}
		src, e := f.Open()
		if e != nil {
			return e
		}
		out, e := os.OpenFile(filepath.Join(dest, base), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
		if e != nil {
			src.Close()
			return e
		}
		n, e := io.Copy(out, io.LimitReader(src, limit+1))
		src.Close()
		ce := out.Close()
		if e == nil {
			e = ce
		}
		if e != nil {
			return e
		}
		if n > limit {
			return errors.New("解压文件超过大小限制")
		}
		if wantExe {
			h, e := os.Open(filepath.Join(dest, base))
			if e != nil {
				return e
			}
			var b [2]byte
			_, e = io.ReadFull(h, b[:])
			h.Close()
			if e != nil || string(b[:]) != "MZ" {
				return errors.New("引擎文件不是 Windows 可执行文件")
			}
		}
	}
	if !found["ffmpeg.exe"] || !found["ffprobe.exe"] {
		return errors.New("压缩包必须同时包含 ffmpeg.exe 和 ffprobe.exe")
	}
	return nil
}
