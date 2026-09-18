package ncm

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"prismtranscode/internal/testfixture"
	"testing"
)

func tempNCM(t *testing.T, b []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "样例_日本語.ncm")
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestExtractAcrossChunkBoundaries(t *testing.T) {
	audio := make([]byte, 350017)
	copy(audio, []byte("ID3"))
	for i := 3; i < len(audio); i++ {
		audio[i] = byte(i * 37)
	}
	p := tempNCM(t, testfixture.Wrap(audio, "mp3", 12345, true))
	c, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	var out bytes.Buffer
	if e = c.Extract(context.Background(), &out, nil); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(audio, out.Bytes()) {
		t.Fatal("stream mismatch")
	}
	if c.Metadata.Title != "测试歌曲・テスト" {
		t.Fatal(c.Metadata)
	}
	if c.Metadata.Artist != "Test Artist" {
		t.Fatal(c.Metadata.Artist)
	}
}
func TestMissingMetadataStillExtracts(t *testing.T) {
	audio := []byte("fLaCsynthetic-stream")
	p := tempNCM(t, testfixture.Wrap(audio, "flac", 0, false))
	c, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	var out bytes.Buffer
	if e = c.Extract(context.Background(), &out, nil); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(out.Bytes(), audio) {
		t.Fatal("mismatch")
	}
}
func TestCorruptMetadataWarnsNotAbort(t *testing.T) {
	raw := testfixture.Wrap([]byte("ID3some-data"), "mp3", 0, true)
	keySize := int(binary.LittleEndian.Uint32(raw[10:14]))
	metaStart := 14 + keySize + 4
	raw[metaStart] ^= 0x12
	c, e := Open(tempNCM(t, raw))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	if len(c.Metadata.Warnings) == 0 {
		t.Fatal("expected metadata warning")
	}
}
func TestRejectMalformedHeaders(t *testing.T) {
	good := testfixture.Wrap([]byte("ID3test-audio"), "mp3", 0, true)
	huge := append([]byte{}, good...)
	binary.LittleEndian.PutUint32(huge[10:14], 0xffffffff)
	wrongKey := append([]byte{}, good...)
	wrongKey[14] ^= 0xff
	for name, data := range map[string][]byte{"empty": {}, "bad_magic": []byte("not an ncm file"), "truncated": good[:30], "huge_key_length": huge, "bad_key": wrongKey} {
		t.Run(name, func(t *testing.T) {
			if c, e := Open(tempNCM(t, data)); e == nil {
				c.Close()
				t.Fatal("should reject")
			}
		})
	}
}
func TestRejectCoverLengthOverflow(t *testing.T) {
	b := testfixture.Wrap([]byte("ID3test-data"), "mp3", 0, false)
	keyLen := int(binary.LittleEndian.Uint32(b[10:14]))
	tail := 14 + keyLen + 4
	binary.LittleEndian.PutUint32(b[tail+5:tail+9], 1)
	binary.LittleEndian.PutUint32(b[tail+9:tail+13], 2)
	if c, e := Open(tempNCM(t, b)); e == nil {
		c.Close()
		t.Fatal("must reject actual > allocated")
	}
}
func TestCancellation(t *testing.T) {
	c, e := Open(tempNCM(t, testfixture.Wrap([]byte("ID3audio"), "mp3", 0, true)))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var b bytes.Buffer
	if e = c.Extract(ctx, &b, nil); e != context.Canceled {
		t.Fatalf("got %v", e)
	}
	if b.Len() != 0 {
		t.Fatal("wrote canceled data")
	}
}
func TestAudioDetection(t *testing.T) {
	for name, b := range map[string][]byte{"mp3": {'I', 'D', '3', 3, 0, 0, 0, 0, 0, 0}, "flac": []byte("fLaCdata")} {
		p := tempNCM(t, b)
		got, e := Detect(p)
		if e != nil || got != name {
			t.Fatalf("%s %v", got, e)
		}
	}
	if _, e := Detect(tempNCM(t, []byte("something else"))); e == nil {
		t.Fatal("accepted unknown audio")
	}
}
