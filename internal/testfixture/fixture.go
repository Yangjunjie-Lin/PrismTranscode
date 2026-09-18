// Package testfixture builds synthetic NCM containers for offline tests.
// The shipped test tones are generated sine waves, not commercial music.
package testfixture

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
)

func EncryptECB(data, key []byte) []byte {
	cipher, e := aes.NewCipher(key)
	if e != nil {
		panic(e)
	}
	pad := 16 - len(data)%16
	data = append(append([]byte{}, data...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	out := make([]byte, len(data))
	for i := 0; i < len(data); i += 16 {
		cipher.Encrypt(out[i:i+16], data[i:i+16])
	}
	return out
}
func Wrap(audio []byte, format string, padding int, withMetadata bool) []byte {
	key := []byte("local-test-key-for-ncm-mp3-batch-2026")
	var out bytes.Buffer
	out.WriteString("CTENFDAM")
	out.Write([]byte{1, 0x70})
	keyEnc := EncryptECB(append([]byte("neteasecloudmusic"), key...), []byte("hzHRAmso5kInbaxW"))
	for i := range keyEnc {
		keyEnc[i] ^= 0x64
	}
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(keyEnc)))
	out.Write(keyEnc)
	var meta []byte
	if withMetadata {
		m, _ := json.Marshal(map[string]interface{}{"musicName": "测试歌曲・テスト", "artist": [][]interface{}{{"Test Artist", 123}}, "album": "Synthetic tones", "format": format, "bitrate": 192000, "duration": 1200})
		mk := []byte{0x23, 0x31, 0x34, 0x6c, 0x6a, 0x6b, 0x5f, 0x21, 0x5c, 0x5d, 0x26, 0x30, 0x55, 0x3c, 0x27, 0x28}
		enc := EncryptECB(append([]byte("music:"), m...), mk)
		meta = []byte("163 key(Don't modify):" + base64.StdEncoding.EncodeToString(enc))
		for i := range meta {
			meta[i] ^= 0x63
		}
	}
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(meta)))
	out.Write(meta)
	out.Write(make([]byte, 4))
	out.WriteByte(1)
	cover := []byte{} // The padding fixture specifically checks allocated > actual.
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(cover)+padding))
	_ = binary.Write(&out, binary.LittleEndian, uint32(len(cover)))
	out.Write(cover)
	out.Write(make([]byte, padding))
	box := make([]int, 256)
	for i := range box {
		box[i] = i
	}
	j := 0
	for i := range box {
		j = (box[i] + j + int(key[i%len(key)])) % 256
		box[i], box[j] = box[j], box[i]
	}
	// Independently compute each output byte instead of reusing the decoder.
	for i, v := range audio {
		p := (i + 1) % 256
		k := box[(box[p]+box[(box[p]+p)%256])%256]
		out.WriteByte(v ^ byte(k))
	}
	return out.Bytes()
}
