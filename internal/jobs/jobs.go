// Package jobs owns the persistent queue and bounded worker pool.
package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"prismtranscode/internal/media"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Job struct {
	ID       string         `json:"id"`
	Input    string         `json:"input"`
	Relative string         `json:"relative"`
	Info     media.Info     `json:"info"`
	Options  media.Options  `json:"options"`
	State    string         `json:"state"`
	Error    string         `json:"error,omitempty"`
	Progress media.Progress `json:"progress"`
	Result   *media.Result  `json:"result,omitempty"`
	Created  string         `json:"created"`
}
type Snapshot struct {
	Jobs      []Job  `json:"jobs"`
	Busy      bool   `json:"busy"`
	OutputDir string `json:"output_dir"`
	Workers   int    `json:"workers"`
}
type Manager struct {
	mu               sync.Mutex
	jobs             []*Job
	busy             bool
	out              string
	workers          int
	engine           *media.Engine
	cache, statePath string
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	persistMu        sync.Mutex
}

func New(e *media.Engine, cache, statePath string) *Manager {
	m := &Manager{engine: e, cache: cache, statePath: statePath, jobs: []*Job{}}
	b, err := os.ReadFile(statePath)
	if err == nil && len(b) < 64<<20 {
		var s Snapshot
		if json.Unmarshal(b, &s) == nil {
			m.out = s.OutputDir
			for _, j := range s.Jobs {
				jj := j
				if jj.State == "running" || jj.State == "waiting" {
					jj.State = "cancelled"
					jj.Error = "上次任务中断，可重新开始"
				}
				m.jobs = append(m.jobs, &jj)
			}
		}
	}
	return m
}
func (m *Manager) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Snapshot{Jobs: []Job{}, Busy: m.busy, OutputDir: m.out, Workers: m.workers}
	for _, j := range m.jobs {
		s.Jobs = append(s.Jobs, *j)
	}
	return s
}
func (m *Manager) Save() {
	if m.statePath == "" {
		return
	}
	m.persistMu.Lock()
	defer m.persistMu.Unlock()
	s := m.Snapshot()
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return
	}
	f, e := os.CreateTemp(filepath.Dir(m.statePath), "queue-*.tmp")
	if e != nil {
		return
	}
	p := f.Name()
	defer os.Remove(p)
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	f.Close()
	if e == nil {
		_ = os.Rename(p, m.statePath)
	}
}
func (m *Manager) SetEngine(e *media.Engine) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return errors.New("转换中不能更换引擎")
	}
	m.engine = e
	return nil
}
func (m *Manager) Engine() *media.Engine { m.mu.Lock(); defer m.mu.Unlock(); return m.engine }
func randomID() string                   { var b [12]byte; _, _ = rand.Read(b[:]); return hex.EncodeToString(b[:]) }
func (m *Manager) Add(ctx context.Context, paths []string, root string, o media.Options) (int, error) {
	if err := o.Validate(); err != nil {
		return 0, err
	}
	count := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if p, e := filepath.EvalSymlinks(abs); e == nil {
			abs = p
		}
		st, err := os.Stat(abs)
		if err != nil || !st.Mode().IsRegular() {
			continue
		}
		m.mu.Lock()
		if m.busy {
			m.mu.Unlock()
			return count, errors.New("请先停止转换再添加文件")
		}
		if len(m.jobs) >= 5000 {
			m.mu.Unlock()
			return count, errors.New("单个队列最多 5000 个文件")
		}
		exists := false
		for _, j := range m.jobs {
			if j.Input == abs || runtime.GOOS == "windows" && strings.EqualFold(j.Input, abs) {
				exists = true
				break
			}
		}
		if exists {
			m.mu.Unlock()
			continue
		}
		j := &Job{ID: randomID(), Input: abs, State: "ready", Options: o, Created: time.Now().Format(time.RFC3339), Info: media.Info{Name: filepath.Base(abs), Path: abs, Size: st.Size(), Kind: "unknown", Warnings: []string{}}}
		if root != "" {
			rel, e := filepath.Rel(root, filepath.Dir(abs))
			if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
				j.Relative = rel
			}
		}
		m.jobs = append(m.jobs, j)
		engine := m.engine
		m.mu.Unlock()
		count++
		if engine != nil {
			info, _, clean, e := Inspect(ctx, engine, abs, m.cache, o.InputFormat)
			if clean != nil {
				clean()
			}
			m.mu.Lock()
			if e != nil {
				j.State = "failed"
				j.Error = e.Error()
			} else {
				j.Info = info
			}
			m.mu.Unlock()
		}
	}
	m.Save()
	return count, nil
}

// Inspect probes decrypted bytes, not NCM's extension or metadata claims.
func Inspect(ctx context.Context, e *media.Engine, path, cache, format string) (media.Info, string, func(), error) {
	var empty media.Info
	if format == "ncm" && !media.IsNCM(path) {
		return empty, "", nil, errors.New("手动指定了 NCM，但文件头不是受支持的 NCM 格式")
	}
	source, meta, clean, err := media.Prepare(ctx, path, cache)
	if err != nil {
		return empty, "", nil, err
	}
	if meta != nil {
		format = "auto"
	}
	info, err := media.Probe(ctx, e, source, format)
	if err != nil {
		clean()
		return empty, "", nil, err
	}
	info.Path = path
	info.Name = filepath.Base(path)
	if st, err := os.Stat(path); err == nil {
		info.Size = st.Size()
	}
	if meta != nil {
		info.NCM = true
		info.Container = "NCM → " + info.Container
		info.Warnings = append(info.Warnings, meta.Warnings...)
		if info.Tags == nil {
			info.Tags = map[string]string{}
		}
		if meta.Title != "" {
			info.Tags["title"] = meta.Title
		}
		if meta.Artist != "" {
			info.Tags["artist"] = meta.Artist
		}
		if meta.Album != "" {
			info.Tags["album"] = meta.Album
		}
		if len(meta.Cover) > 0 {
			info.Warnings = append(info.Warnings, "NCM 内嵌封面本版不迁移；保留标题、歌手、专辑文字标签。")
		}
	}
	return info, source, clean, nil
}
func Scan(root string, recursive bool, exclude string) ([]string, error) {
	root, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	st, e := os.Stat(root)
	if e != nil || !st.IsDir() {
		return nil, errors.New("输入路径不是文件夹")
	}
	out := []string{}
	visited := 0
	e = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited++
		if visited > 100000 {
			return errors.New("扫描超过 100000 个目录项，请缩小输入范围")
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			if !recursive || strings.HasPrefix(d.Name(), ".") || exclude != "" && path == exclude {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if media.KnownExtension(path) {
			out = append(out, path)
			if len(out) > 5000 {
				return errors.New("扫描结果超过 5000 个文件，请分批导入")
			}
		}
		return nil
	})
	return out, e
}
func (m *Manager) Configure(ids []string, o media.Options) error {
	if e := o.Validate(); e != nil {
		return e
	}
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return errors.New("转换中不能修改队列")
	}
	set := idSet(ids)
	for _, j := range m.jobs {
		if set[j.ID] {
			j.Options = o
			if j.State == "completed" || j.State == "skipped" {
				j.State = "ready"
				j.Result = nil
			}
		}
	}
	m.mu.Unlock()
	m.Save()
	return nil
}
func (m *Manager) Remove(ids []string) error {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return errors.New("请先停止转换")
	}
	set := idSet(ids)
	keep := []*Job{}
	for _, j := range m.jobs {
		if !set[j.ID] {
			keep = append(keep, j)
		}
	}
	m.jobs = keep
	m.mu.Unlock()
	m.Save()
	return nil
}
func idSet(ids []string) map[string]bool {
	s := map[string]bool{}
	for _, v := range ids {
		s[v] = true
	}
	return s
}
func (m *Manager) Start(ids []string, out string, workers int) error {
	if out == "" {
		return errors.New("请选择输出目录")
	}
	var err error
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}
	if workers < 1 || workers > 4 {
		return errors.New("并行任务数须为 1–4")
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return errors.New("已有任务在运行")
	}
	if m.engine == nil {
		m.mu.Unlock()
		return errors.New("请先安装转换引擎")
	}
	set := idSet(ids)
	selected := []*Job{}
	for _, j := range m.jobs {
		if set[j.ID] && j.State != "completed" {
			selected = append(selected, j)
		}
	}
	if len(selected) == 0 {
		m.mu.Unlock()
		return errors.New("请选择待处理文件；已完成文件请先重新应用参数")
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.busy = true
	m.out = out
	m.workers = workers
	engine := m.engine
	for _, j := range selected {
		j.State = "waiting"
		j.Error = ""
		j.Result = nil
		j.Progress = media.Progress{Phase: "等待", Percent: 0}
	}
	m.wg.Add(1)
	m.mu.Unlock()
	m.Save()
	go func() {
		defer m.wg.Done()
		defer cancel()
		queue := make(chan *Job, len(selected))
		for _, j := range selected {
			queue <- j
		}
		close(queue)
		var wg sync.WaitGroup
		for n := 0; n < workers; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range queue {
					m.run(ctx, engine, j, out)
				}
			}()
		}
		wg.Wait()
		m.mu.Lock()
		m.busy = false
		m.cancel = nil
		m.mu.Unlock()
		m.Save()
	}()
	return nil
}
func (m *Manager) run(ctx context.Context, e *media.Engine, j *Job, out string) {
	m.mu.Lock()
	j.State = "running"
	j.Progress = media.Progress{Phase: "读取与识别", Percent: 0}
	o := j.Options
	m.mu.Unlock()
	finishError := func(err error) {
		m.mu.Lock()
		if errors.Is(err, context.Canceled) {
			j.State = "cancelled"
			j.Progress.Phase = "已停止"
		} else if errors.Is(err, media.ErrSkipped) {
			j.State = "skipped"
			j.Progress.Phase = "已跳过"
		} else {
			j.State = "failed"
			j.Progress.Phase = "失败"
		}
		j.Error = err.Error()
		m.mu.Unlock()
		m.Save()
	}
	if ctx.Err() != nil {
		finishError(ctx.Err())
		return
	}
	info, source, clean, err := Inspect(ctx, e, j.Input, m.cache, o.InputFormat)
	if err != nil {
		finishError(err)
		return
	}
	defer clean()
	m.mu.Lock()
	j.Info = info
	m.mu.Unlock()
	if o.PreserveFolders && j.Relative != "" {
		out = filepath.Join(out, j.Relative)
	}
	base := strings.TrimSuffix(filepath.Base(j.Input), filepath.Ext(j.Input))
	r, err := media.Execute(ctx, e, info, source, out, base, o, func(p media.Progress) { m.mu.Lock(); j.Progress = p; m.mu.Unlock() })
	m.mu.Lock()
	j.Result = &r
	m.mu.Unlock()
	if err != nil {
		finishError(err)
		return
	}
	m.mu.Lock()
	j.State = "completed"
	j.Progress = media.Progress{Phase: "完成", Percent: 100}
	m.mu.Unlock()
	m.Save()
}
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()
}
func (m *Manager) Wait() { m.wg.Wait() }
func (m *Manager) Preview(ctx context.Context, id string) (media.Plan, error) {
	m.mu.Lock()
	var j *Job
	for _, v := range m.jobs {
		if v.ID == id {
			vv := *v
			j = &vv
			break
		}
	}
	e := m.engine
	m.mu.Unlock()
	if j == nil {
		return media.Plan{}, errors.New("任务不存在")
	}
	if e == nil {
		return media.Plan{}, errors.New("引擎未安装")
	}
	i, source, clean, err := Inspect(ctx, e, j.Input, m.cache, j.Options.InputFormat)
	if err != nil {
		return media.Plan{}, err
	}
	defer clean()
	p, err := media.Build(e, i, source, j.Options)
	if err == nil {
		p.Args = append([]string{e.FFmpeg}, p.Args...)
		p.Args = append(p.Args, fmt.Sprintf("<OUTPUT>.%s", p.Ext))
	}
	return p, err
}
