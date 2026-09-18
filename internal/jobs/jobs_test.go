package jobs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"prismtranscode/internal/media"
	"testing"
)

func TestForcedNCMRejectsWrongHeader(t *testing.T) {
	p := filepath.Join(t.TempDir(), "not-ncm.ncm")
	if err := os.WriteFile(p, []byte("Not a NetEase header"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := Inspect(context.Background(), nil, p, t.TempDir(), "ncm")
	if err == nil {
		t.Fatal("forcing NCM must not ignore signature")
	}
}
func TestScanHonorsRecursionAndOutputExclusion(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	out := filepath.Join(root, "outputs")
	for _, p := range []string{sub, out} {
		if e := os.Mkdir(p, 0700); e != nil {
			t.Fatal(e)
		}
	}
	for _, p := range []string{filepath.Join(root, "a.mp3"), filepath.Join(root, "b.docx"), filepath.Join(root, ".hidden.mp3"), filepath.Join(sub, "c.flac"), filepath.Join(out, "d.mp4")} {
		os.WriteFile(p, []byte("fixture"), 0600)
	}
	p, e := Scan(root, false, out)
	if e != nil || len(p) != 1 {
		t.Fatalf("shallow %v %v", p, e)
	}
	p, e = Scan(root, true, out)
	if e != nil || len(p) != 2 {
		t.Fatalf("recursive %v %v", p, e)
	}
}
func TestQueuePersistsAndRemoveLeavesOriginal(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "file.wav")
	os.WriteFile(p, []byte("fixture"), 0600)
	state := filepath.Join(root, "queue.json")
	m := New(nil, root, state)
	n, e := m.Add(context.Background(), []string{p, p}, "", media.DefaultOptions())
	if e != nil || n != 1 {
		t.Fatalf("dedupe %d %v", n, e)
	}
	m.Save()
	restored := New(nil, root, state)
	q := restored.Snapshot()
	if len(q.Jobs) != 1 {
		t.Fatal("queue not restored")
	}
	o := media.DefaultOptions()
	o.Target = "flac"
	if e = restored.Configure([]string{q.Jobs[0].ID}, o); e != nil {
		t.Fatal(e)
	}
	if restored.Snapshot().Jobs[0].Options.Target != "flac" {
		t.Fatal("options not saved")
	}
	if e = restored.Remove([]string{q.Jobs[0].ID}); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(p); e != nil {
		t.Fatal("original was removed")
	}
}
func TestInterruptedQueueRestoresCancelled(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "queue.json")
	b, _ := json.Marshal(Snapshot{Jobs: []Job{{ID: "a", State: "running"}, {ID: "b", State: "waiting"}, {ID: "c", State: "completed"}}})
	os.WriteFile(state, b, 0600)
	q := New(nil, root, state).Snapshot()
	if q.Jobs[0].State != "cancelled" || q.Jobs[1].State != "cancelled" || q.Jobs[2].State != "completed" {
		t.Fatal("invalid restored states")
	}
}
