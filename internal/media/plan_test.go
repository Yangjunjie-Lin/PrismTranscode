package media

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeEngine() *Engine {
	e := &Engine{Encoders: map[string]bool{}, Muxers: map[string]bool{}, Filters: map[string]bool{}, Soxr: true}
	for _, t := range Targets {
		e.Encoders[t.Encoder] = true
		e.Muxers[t.Mux] = true
	}
	for _, c := range []string{"pcm_f32le", "pcm_f64le", "pcm_s24le", "pcm_s32le", "pcm_s24be", "pcm_s32be", "pcm_f32be", "pcm_f64be", "mp2"} {
		e.Encoders[c] = true
	}
	return e
}
func audioInfo() Info {
	return Info{Kind: "audio", Duration: 3, Streams: []Stream{{Index: 0, Type: "audio", Codec: "mp3", SampleRate: "44100", Channels: 2, SampleFmt: "fltp"}}, Tags: map[string]string{}, Warnings: []string{}}
}
func videoInfo() Info {
	i := audioInfo()
	i.Kind = "video"
	i.Streams = append(i.Streams, Stream{Index: 1, Type: "video", Codec: "h264", Width: 320, Height: 180, PixFmt: "yuv420p"})
	return i
}
func TestCatalogUnique(t *testing.T) {
	ids := map[string]bool{}
	for _, v := range Targets {
		if ids[v.ID] {
			t.Fatal("duplicate", v.ID)
		}
		ids[v.ID] = true
	}
	if len(ids) != 38 {
		t.Fatalf("want 38 presets got %d", len(ids))
	}
}
func TestStreamCopyDefault(t *testing.T) {
	o := DefaultOptions()
	p, e := Build(fakeEngine(), audioInfo(), "source.mp3", o)
	if e != nil || p.Mode != "stream-copy" {
		t.Fatal(p, e)
	}
	o.SampleRate = 48000
	p, e = Build(fakeEngine(), audioInfo(), "source.mp3", o)
	if e != nil || p.Mode != "transcode" || !strings.Contains(strings.Join(p.Args, " "), "soxr:precision=28") {
		t.Fatal(p, e)
	}
}
func TestNoShellInterpolation(t *testing.T) {
	o := DefaultOptions()
	path := "song ' ; & $(bad).mp3"
	p, e := Build(fakeEngine(), audioInfo(), path, o)
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, a := range p.Args {
		if a == path {
			count++
		}
	}
	if count != 1 {
		t.Fatal("path is not one unchanged argument")
	}
}
func TestRejectUnsafeOptions(t *testing.T) {
	for _, change := range []func(*Options){func(o *Options) { o.Target = "../../bad" }, func(o *Options) { o.InputFormat = "concat" }, func(o *Options) { o.Suffix = "../bad" }, func(o *Options) { o.Start = -1 }, func(o *Options) { o.Threads = 500 }, func(o *Options) { o.Collision = "overwrite" }, func(o *Options) { o.BitDepth = 17 }} {
		o := DefaultOptions()
		change(&o)
		if o.Validate() == nil {
			t.Fatal(o)
		}
	}
}
func TestHDRFailClosed(t *testing.T) {
	i := videoInfo()
	i.HDR = true
	o := DefaultOptions()
	o.Target = "mp4"
	if _, e := Build(fakeEngine(), i, "hdr.mp4", o); e == nil {
		t.Fatal("HDR must not silently lose color")
	}
	o.Target = "remux_mkv"
	if _, e := Build(fakeEngine(), i, "hdr.mp4", o); e != nil {
		t.Fatal(e)
	}
}
func TestIncompatibleRemux(t *testing.T) {
	i := videoInfo()
	i.Streams[0].Codec = "vorbis"
	o := DefaultOptions()
	o.Target = "remux_mp4"
	if _, e := Build(fakeEngine(), i, "v.mkv", o); e == nil {
		t.Fatal("should reject incompatible MP4 audio")
	}
}
func TestNoAudioReject(t *testing.T) {
	i := Info{Kind: "image", Streams: []Stream{{Index: 0, Type: "video", Codec: "png"}}}
	if _, e := Build(fakeEngine(), i, "image.png", DefaultOptions()); e == nil {
		t.Fatal("should reject no audio")
	}
}
func TestMissingCodec(t *testing.T) {
	e := fakeEngine()
	delete(e.Encoders, "libsvtav1")
	o := DefaultOptions()
	o.Target = "mp4_av1"
	if _, err := Build(e, videoInfo(), "a.mp4", o); err == nil {
		t.Fatal("missing AV1 must fail")
	}
}
func TestSWRFallbackDisclosed(t *testing.T) {
	e := fakeEngine()
	e.Soxr = false
	o := DefaultOptions()
	o.SampleRate = 48000
	p, err := Build(e, audioInfo(), "a.mp3", o)
	if err != nil || len(p.Warnings) == 0 || !strings.Contains(strings.Join(p.Args, " "), "resampler=swr") {
		t.Fatal(p, err)
	}
}
func TestSubtitleImageReject(t *testing.T) {
	o := DefaultOptions()
	o.Target = "srt"
	i := Info{Streams: []Stream{{Index: 0, Type: "subtitle", Codec: "hdmv_pgs_subtitle"}}}
	if _, e := Build(fakeEngine(), i, "sub.mkv", o); e == nil {
		t.Fatal("image subtitle must fail")
	}
}
func TestCommitNeverOverwrites(t *testing.T) {
	d := t.TempDir()
	src := filepath.Join(d, "temp")
	os.WriteFile(src, []byte("new"), 0600)
	dest := filepath.Join(d, "name.mp3")
	os.WriteFile(dest, []byte("original"), 0600)
	out, _, e := commit(context.Background(), src, d, "name", "mp3", "rename")
	if e != nil || out == dest {
		t.Fatal(out, e)
	}
	b, _ := os.ReadFile(dest)
	if string(b) != "original" {
		t.Fatal("overwritten")
	}
	if _, _, e = commit(context.Background(), src, d, "name", "mp3", "skip"); e != ErrSkipped {
		t.Fatal(e)
	}
}
func TestCleanReservedFilename(t *testing.T) {
	for _, s := range []string{"CON", "x/y\\z", "...", "a\x00b"} {
		got := cleanStem(s)
		if got == "" || strings.ContainsAny(got, "/\\\x00") {
			t.Fatal(got)
		}
	}
}

func TestStillImageBrands(t *testing.T) {
	for _, brand := range []string{"avif", "heic", "mif1", "avif "} {
		if !stillImageBrand(brand) {
			t.Fatal(brand)
		}
	}
	for _, brand := range []string{"isom", "mp42", "qt  ", "avis"} {
		if stillImageBrand(brand) {
			t.Fatal(brand)
		}
	}
}
func TestExplicitEngineDoesNotFallback(t *testing.T) {
	if _, e := Select(filepath.Join(t.TempDir(), "missing")); e == nil {
		t.Fatal("explicit invalid engine path must fail")
	}
}
