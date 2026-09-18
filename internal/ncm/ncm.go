// Package ncm reads the conventional CTENFDAM NCM container.
// Protocol reference: https://github.com/taurusxin/ncmdump (MIT).
// This is an independent Go implementation, with bounded header parsing,
// strict key padding checks, streaming audio extraction, and cancellation.
package ncm

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var coreKey = []byte("hzHRAmso5kInbaxW")
var metaKey = []byte{0x23, 0x31, 0x34, 0x6c, 0x6a, 0x6b, 0x5f, 0x21, 0x5c, 0x5d, 0x26, 0x30, 0x55, 0x3c, 0x27, 0x28}

const MaxHeader = 32 << 20

type Metadata struct {
	Title      string   `json:"title"`
	Artist     string   `json:"artist"`
	Album      string   `json:"album"`
	Format     string   `json:"format"`
	DurationMS float64  `json:"duration_ms"`
	Bitrate    int64    `json:"bitrate"`
	Cover      []byte   `json:"-"`
	Warnings   []string `json:"warnings,omitempty"`
}

type Container struct {
	File     *os.File
	Metadata Metadata
	Offset   int64
	Size     int64
	stream   [256]byte
}

func readBytes(r io.Reader, n uint32, max uint32) ([]byte, error) {
	if n > max {
		return nil, fmt.Errorf("NCM 头部长度异常：%d 字节", n)
	}
	b := make([]byte, int(n))
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, fmt.Errorf("NCM 文件不完整：%w", err)
	}
	return b, nil
}
func readU32(r io.Reader) (uint32, error) {
	var b [4]byte
	_, e := io.ReadFull(r, b[:])
	return binary.LittleEndian.Uint32(b[:]), e
}
func decryptECB(data, key []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%16 != 0 {
		return nil, errors.New("AES 数据长度不正确")
	}
	cipher, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	out := make([]byte, len(data))
	for i := 0; i < len(data); i += 16 {
		cipher.Decrypt(out[i:i+16], data[i:i+16])
	}
	pad := int(out[len(out)-1])
	if pad < 1 || pad > 16 || pad > len(out) {
		return nil, errors.New("AES 填充不正确，文件可能损坏或格式不支持")
	}
	for _, v := range out[len(out)-pad:] {
		if int(v) != pad {
			return nil, errors.New("AES 填充校验失败")
		}
	}
	return out[:len(out)-pad], nil
}

func Open(path string) (_ *Container, err error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer func() {
		if err != nil {
			f.Close()
		}
	}()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, errors.New("输入必须是普通文件")
	}
	magic, e := readBytes(f, 10, 10)
	if e != nil {
		return nil, e
	}
	if string(magic[:8]) != "CTENFDAM" {
		return nil, errors.New("不是支持的 NCM 文件：文件头不匹配")
	}
	keyLen, e := readU32(f)
	if e != nil {
		return nil, errors.New("缺少 NCM 密钥长度")
	}
	keyData, e := readBytes(f, keyLen, 1<<20)
	if e != nil {
		return nil, e
	}
	for i := range keyData {
		keyData[i] ^= 0x64
	}
	plain, e := decryptECB(keyData, coreKey)
	if e != nil {
		return nil, fmt.Errorf("NCM 密钥解码失败：%w", e)
	}
	if len(plain) <= 17 || !bytes.Equal(plain[:17], []byte("neteasecloudmusic")) {
		return nil, errors.New("不支持的 NCM 密钥标识")
	}
	key := plain[17:]
	var box [256]byte
	for i := range box {
		box[i] = byte(i)
	}
	last := 0
	for i := 0; i < 256; i++ {
		j := (int(box[i]) + last + int(key[i%len(key)])) & 255
		box[i], box[j] = box[j], box[i]
		last = j
	}
	c := &Container{File: f, Size: st.Size()}
	for i := 0; i < 256; i++ {
		j := (i + 1) & 255
		c.stream[i] = box[(int(box[j])+int(box[(int(box[j])+j)&255]))&255]
	}
	metaLen, e := readU32(f)
	if e != nil {
		return nil, errors.New("缺少 NCM 元数据长度")
	}
	meta, e := readBytes(f, metaLen, MaxHeader)
	if e != nil {
		return nil, e
	}
	if len(meta) > 0 {
		if e = c.parseMetadata(meta); e != nil {
			c.Metadata.Warnings = append(c.Metadata.Warnings, "歌曲信息未能读取；继续提取音频："+e.Error())
		}
	}
	// CRC32 (4), cover version (1), allocated cover length (4), actual cover length (4).
	tail, e := readBytes(f, 13, 13)
	if e != nil {
		return nil, e
	}
	allocated := binary.LittleEndian.Uint32(tail[5:9])
	actual := binary.LittleEndian.Uint32(tail[9:13])
	if actual > allocated || allocated > MaxHeader {
		return nil, errors.New("NCM 封面区长度异常")
	}
	offset, e := f.Seek(0, io.SeekCurrent)
	if e != nil {
		return nil, e
	}
	if offset+int64(allocated) >= st.Size() {
		return nil, errors.New("NCM 没有完整音频数据")
	}
	cover, e := readBytes(f, actual, MaxHeader)
	if e != nil {
		return nil, e
	}
	c.Metadata.Cover = cover
	c.Offset = offset + int64(allocated)
	_, e = f.Seek(c.Offset, io.SeekStart)
	if e != nil {
		return nil, e
	}
	return c, nil
}
func (c *Container) Close() error { return c.File.Close() }

func (c *Container) parseMetadata(b []byte) error {
	for i := range b {
		b[i] ^= 0x63
	}
	prefix := []byte("163 key(Don't modify):")
	if !bytes.HasPrefix(b, prefix) {
		return errors.New("未知元数据标识")
	}
	raw, e := base64.StdEncoding.DecodeString(strings.TrimSpace(string(b[len(prefix):])))
	if e != nil {
		return e
	}
	data, e := decryptECB(raw, metaKey)
	if e != nil {
		return e
	}
	if !bytes.HasPrefix(data, []byte("music:")) {
		return errors.New("未知元数据类型")
	}
	var m map[string]json.RawMessage
	if e = json.Unmarshal(data[6:], &m); e != nil {
		return e
	}
	getS := func(k string) string { var v string; _ = json.Unmarshal(m[k], &v); return v }
	c.Metadata.Title = getS("musicName")
	c.Metadata.Album = getS("album")
	c.Metadata.Format = strings.ToLower(getS("format"))
	_ = json.Unmarshal(m["duration"], &c.Metadata.DurationMS)
	_ = json.Unmarshal(m["bitrate"], &c.Metadata.Bitrate)
	var artists [][]json.RawMessage
	if json.Unmarshal(m["artist"], &artists) == nil {
		var names []string
		for _, a := range artists {
			if len(a) > 0 {
				var n string
				if json.Unmarshal(a[0], &n) == nil && n != "" {
					names = append(names, n)
				}
			}
		}
		c.Metadata.Artist = strings.Join(names, " / ")
	}
	return nil
}

// Extract uses fixed-size memory irrespective of input audio size.
func (c *Container) Extract(ctx context.Context, w io.Writer, progress func(float64)) error {
	if _, e := c.File.Seek(c.Offset, io.SeekStart); e != nil {
		return e
	}
	buf := make([]byte, 128*1024)
	var offset int64
	total := c.Size - c.Offset
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		n, e := c.File.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				buf[i] ^= c.stream[(offset+int64(i))&255]
			}
			nn, we := w.Write(buf[:n])
			if we != nil {
				return we
			}
			if nn != n {
				return io.ErrShortWrite
			}
			offset += int64(n)
			if progress != nil {
				progress(float64(offset) / float64(total))
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
	}
	if offset != total {
		return errors.New("NCM 音频读取不完整")
	}
	return nil
}

func Detect(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	var b [12]byte
	n, e := f.Read(b[:])
	if e != nil && e != io.EOF {
		return "", e
	}
	if n >= 4 && string(b[:4]) == "fLaC" {
		return "flac", nil
	}
	if n >= 3 && string(b[:3]) == "ID3" {
		return "mp3", nil
	}
	// MPEG sync, non-reserved version, Layer III, supported bitrate and sample-rate fields.
	if n >= 4 && b[0] == 255 && b[1]&0xe0 == 0xe0 && b[1]&0x18 != 0x08 && b[1]&0x06 == 0x02 && b[2]>>4 != 0 && b[2]>>4 != 15 && b[2]&0x0c != 0x0c {
		return "mp3", nil
	}
	return "", errors.New("解码后的音频不是可识别的 MP3 或 FLAC；文件可能损坏或使用了不支持的格式")
}
