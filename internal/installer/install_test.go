package installer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecksum(t *testing.T) {
	h := strings.Repeat("a", 64)
	if x, e := Checksum(h+"  wanted.zip\n", "wanted.zip"); e != nil || x != h {
		t.Fatal(x, e)
	}
	if _, e := Checksum(h+"  other.zip", "wanted.zip"); e == nil {
		t.Fatal("wrong archive accepted")
	}
	if _, e := Checksum("invalid hash", ""); e == nil {
		t.Fatal("invalid hash accepted")
	}
}
func TestURLAllowlist(t *testing.T) {
	for _, s := range []string{"http://github.com/x", "https://github.com.evil.test/x", "file:///a", "https://evil.test/x"} {
		if safeURL(s) {
			t.Fatal(s)
		}
	}
	if !safeURL(fullURL) {
		t.Fatal("official distributor rejected")
	}
}
func testZIP(t *testing.T, entries map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.zip")
	f, _ := os.Create(p)
	z := zip.NewWriter(f)
	for name, data := range entries {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		w.Write([]byte(data))
	}
	z.Close()
	f.Close()
	return p
}
func TestExtractBothBinaries(t *testing.T) {
	p := testZIP(t, map[string]string{"foo/bin/ffmpeg.exe": "MZfake", "foo/bin/ffprobe.exe": "MZfake", "foo/LICENSE": "license", "../../outside.txt": "bad"})
	d := filepath.Join(t.TempDir(), "stage")
	if e := Extract(p, d); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(d, "ffprobe.exe")); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(d, "outside.txt")); e == nil {
		t.Fatal("unexpected file extracted")
	}
}
func TestRejectMissingOrInvalidPE(t *testing.T) {
	for _, m := range []map[string]string{{"ffmpeg.exe": "MZfake"}, {"ffmpeg.exe": "notPE", "ffprobe.exe": "MZfake"}} {
		if e := Extract(testZIP(t, m), t.TempDir()); e == nil {
			t.Fatal("invalid engine accepted")
		}
	}
}
