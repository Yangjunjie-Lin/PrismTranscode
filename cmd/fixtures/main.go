// Generate synthetic input files for browser/CLI integration tests.
package main

import (
	"flag"
	"os"
	"path/filepath"
	"prismtranscode/internal/testfixture"
)

func main() {
	out := flag.String("out", "test-inputs", "fixture output directory")
	flag.Parse()
	if e := os.MkdirAll(*out, 0755); e != nil {
		panic(e)
	}
	for _, format := range []string{"mp3", "flac"} {
		b, e := os.ReadFile(filepath.Join("testdata", "tone."+format))
		if e != nil {
			panic(e)
		}
		if e = os.WriteFile(filepath.Join(*out, "测试_"+format+".ncm"), testfixture.Wrap(b, format, 2048, true), 0644); e != nil {
			panic(e)
		}
	}
	_ = os.WriteFile(filepath.Join(*out, "损坏样例.ncm"), []byte("not-an-ncm-container"), 0644)
}
