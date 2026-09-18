// PrismTranscode: a localhost-only, offline-first media conversion workbench.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"prismtranscode/internal/installer"
	"prismtranscode/internal/instance"
	"prismtranscode/internal/jobs"
	"prismtranscode/internal/media"
	"prismtranscode/internal/platform"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed web/*
var assets embed.FS

const Version = "2.1.0-beta.1"

type Settings struct {
	Engine string `json:"engine"`
	Output string `json:"output_dir"`
}
type InstallStatus struct {
	Busy    bool    `json:"busy"`
	Phase   string  `json:"phase"`
	Percent float64 `json:"percent"`
	Error   string  `json:"error,omitempty"`
}
type App struct {
	mu                                  sync.Mutex
	manager                             *jobs.Manager
	token, host, cache, session, appDir string
	settings                            Settings
	installation                        InstallStatus
	installCancel                       context.CancelFunc
	quit                                chan struct{}
	once                                sync.Once
	lastSeen                            time.Time
	importMu                            sync.Mutex
	saveMu                              sync.Mutex
}

func reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func failure(w http.ResponseWriter, e error) { reply(w, 400, map[string]string{"error": e.Error()}) }
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		return errors.New("请求必须只包含一个 JSON 对象")
	}
	return nil
}
func (a *App) auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Host != a.host || r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "http://"+a.host {
			reply(w, 403, map[string]string{"error": "仅允许本地同源访问"})
			return
		}
		t := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(t), []byte(a.token)) != 1 {
			reply(w, 403, map[string]string{"error": "会话失效，请重新打开软件"})
			return
		}
		a.mu.Lock()
		a.lastSeen = time.Now()
		a.mu.Unlock()
		fn(w, r)
	}
}
func (a *App) post(fn http.HandlerFunc) http.HandlerFunc {
	return a.auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			reply(w, 405, map[string]string{"error": "需要 POST"})
			return
		}
		fn(w, r)
	})
}
func (a *App) save() {
	a.saveMu.Lock()
	defer a.saveMu.Unlock()
	a.mu.Lock()
	b, _ := json.MarshalIndent(a.settings, "", "  ")
	a.mu.Unlock()
	f, e := os.CreateTemp(a.cache, "settings-*.tmp")
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
		_ = os.Rename(p, filepath.Join(a.cache, "settings.json"))
	}
}
func (a *App) shutdown() {
	a.once.Do(func() {
		a.manager.Stop()
		a.mu.Lock()
		if a.installCancel != nil {
			a.installCancel()
		}
		a.mu.Unlock()
		close(a.quit)
	})
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/"+a.token+"/", func(w http.ResponseWriter, r *http.Request) {
		if r.Host != a.host || r.URL.Path != "/app/"+a.token+"/" {
			http.NotFound(w, r)
			return
		}
		b, _ := assets.ReadFile("web/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		_, _ = w.Write([]byte(strings.ReplaceAll(string(b), "__TOKEN__", a.token)))
	})
	webFS, _ := fs.Sub(assets, "web")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(webFS))))
	mux.HandleFunc("/api/config", a.auth(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		s := a.settings
		is := a.installation
		a.mu.Unlock()
		e := a.manager.Engine()
		reply(w, 200, map[string]any{"name": "流光转码 · PrismTranscode", "version": Version, "engine": e, "targets": media.Catalog(e), "defaults": media.DefaultOptions(), "settings": s, "platform": runtime.GOOS, "installation": is})
	}))
	mux.HandleFunc("/api/queue", a.auth(func(w http.ResponseWriter, r *http.Request) { reply(w, 200, a.manager.Snapshot()) }))
	mux.HandleFunc("/api/add", a.post(func(w http.ResponseWriter, r *http.Request) {
		a.importMu.Lock()
		defer a.importMu.Unlock()
		var q struct {
			Paths     []string      `json:"paths"`
			Folder    string        `json:"folder"`
			Recursive bool          `json:"recursive"`
			Options   media.Options `json:"options"`
		}
		q.Options = media.DefaultOptions()
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		if q.Folder != "" {
			a.mu.Lock()
			out := a.settings.Output
			a.mu.Unlock()
			p, e := jobs.Scan(q.Folder, q.Recursive, out)
			if e != nil {
				failure(w, e)
				return
			}
			q.Paths = append(q.Paths, p...)
		}
		n, e := a.manager.Add(r.Context(), q.Paths, q.Folder, q.Options)
		if e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]int{"added": n})
	}))
	mux.HandleFunc("/api/browse-files", a.post(func(w http.ResponseWriter, r *http.Request) {
		p, e := platform.BrowseFiles()
		if e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]any{"paths": p})
	}))
	mux.HandleFunc("/api/browse-folder", a.post(func(w http.ResponseWriter, r *http.Request) {
		p, e := platform.BrowseFolder()
		if e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]string{"path": p})
	}))
	mux.HandleFunc("/api/upload", a.post(a.upload))
	mux.HandleFunc("/api/configure", a.post(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			IDs     []string      `json:"ids"`
			Options media.Options `json:"options"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		if e := a.manager.Configure(q.IDs, q.Options); e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/remove", a.post(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			IDs []string `json:"ids"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		if e := a.manager.Remove(q.IDs); e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/start", a.post(func(w http.ResponseWriter, r *http.Request) {
		a.importMu.Lock()
		defer a.importMu.Unlock()
		var q struct {
			IDs     []string `json:"ids"`
			Output  string   `json:"output_dir"`
			Workers int      `json:"workers"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		a.mu.Lock()
		installing := a.installation.Busy
		a.mu.Unlock()
		if installing {
			failure(w, errors.New("引擎安装期间不能开始转换，请完成或取消安装后再试"))
			return
		}
		if e := a.manager.Start(q.IDs, q.Output, q.Workers); e != nil {
			failure(w, e)
			return
		}
		a.mu.Lock()
		a.settings.Output = q.Output
		a.mu.Unlock()
		a.save()
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/stop", a.post(func(w http.ResponseWriter, r *http.Request) {
		a.manager.Stop()
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/plan", a.post(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID string `json:"id"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		p, e := a.manager.Preview(r.Context(), q.ID)
		if e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, p)
	}))
	mux.HandleFunc("/api/engine", a.post(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			Path   string `json:"path"`
			Browse bool   `json:"browse"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		if a.manager.Snapshot().Busy {
			failure(w, errors.New("转换中不能更换引擎"))
			return
		}
		if q.Browse {
			p, e := platform.BrowseFolder()
			if e != nil {
				failure(w, e)
				return
			}
			if p == "" {
				reply(w, 200, map[string]bool{"cancelled": true})
				return
			}
			q.Path = p
		}
		e, err := media.Select(q.Path)
		if err != nil {
			failure(w, err)
			return
		}
		if err = a.manager.SetEngine(e); err != nil {
			failure(w, err)
			return
		}
		a.mu.Lock()
		a.settings.Engine = e.FFmpeg
		a.mu.Unlock()
		a.save()
		reply(w, 200, e)
	}))
	mux.HandleFunc("/api/install", a.post(a.install))
	mux.HandleFunc("/api/install-status", a.auth(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		s := a.installation
		a.mu.Unlock()
		reply(w, 200, s)
	}))
	mux.HandleFunc("/api/cancel-install", a.post(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		if a.installCancel != nil {
			a.installCancel()
		}
		a.mu.Unlock()
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/open-output", a.post(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			Path string `json:"path"`
		}
		if e := decode(w, r, &q); e != nil {
			failure(w, e)
			return
		}
		path, e := filepath.Abs(q.Path)
		if e != nil {
			failure(w, e)
			return
		}
		s, e := os.Stat(path)
		if e != nil {
			failure(w, e)
			return
		}
		if !s.IsDir() {
			path = filepath.Dir(path)
		}
		if e = platform.OpenFolder(path); e != nil {
			failure(w, e)
			return
		}
		reply(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/export", a.auth(a.export))
	mux.HandleFunc("/api/shutdown", a.post(func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]bool{"ok": true}); a.shutdown() }))
	return mux
}
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	a.importMu.Lock()
	defer a.importMu.Unlock()
	if a.manager.Snapshot().Busy {
		failure(w, errors.New("转换中不能导入"))
		return
	}
	// MultipartReader streams to disk; no ParseMultipartForm or whole-file RAM buffer.
	r.Body = http.MaxBytesReader(w, r.Body, 8<<30)
	mr, e := r.MultipartReader()
	if e != nil {
		failure(w, e)
		return
	}
	options := media.DefaultOptions()
	paths := []string{}
	staged := []string{}
	committed := false
	defer func() {
		if !committed {
			for _, dir := range staged {
				_ = os.RemoveAll(dir)
			}
		}
	}()
	count := 0
	// Upload copies are stored persistently so queue entries survive application restarts.
	for {
		part, e := mr.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			failure(w, e)
			return
		}
		if part.FormName() == "options" {
			b, e := io.ReadAll(io.LimitReader(part, 65536))
			part.Close()
			if e != nil || json.Unmarshal(b, &options) != nil {
				failure(w, errors.New("导入参数错误"))
				return
			}
			continue
		}
		if part.FileName() == "" {
			part.Close()
			continue
		}
		count++
		if count > 200 {
			failure(w, errors.New("一次拖放最多 200 个文件；大批量请用文件夹导入"))
			return
		}
		dir, e := os.MkdirTemp(filepath.Join(a.cache, "imports"), "import-")
		if e != nil {
			failure(w, e)
			return
		}
		staged = append(staged, dir)
		name := filepath.Base(strings.ReplaceAll(part.FileName(), "\\", "/"))
		if name == "." || name == "" || name == ".." {
			name = "upload.media"
		}
		path := filepath.Join(dir, name)
		out, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			failure(w, e)
			return
		}
		_, e = io.Copy(out, part)
		ce := out.Close()
		part.Close()
		if e == nil {
			e = ce
		}
		if e != nil {
			os.RemoveAll(dir)
			failure(w, e)
			return
		}
		paths = append(paths, path)
	}
	if e := options.Validate(); e != nil {
		failure(w, e)
		return
	}
	n, e := a.manager.Add(r.Context(), paths, "", options)
	// Add can partially succeed before cancellation; keep referenced inputs.
	committed = n > 0
	if e != nil {
		failure(w, e)
		return
	}
	reply(w, 200, map[string]int{"added": n})
}
func (a *App) install(w http.ResponseWriter, r *http.Request) {
	var q struct {
		Source string `json:"source"`
	}
	if e := decode(w, r, &q); e != nil {
		failure(w, e)
		return
	}
	if runtime.GOOS != "windows" {
		failure(w, errors.New("自动安装仅支持 Windows；其他平台请配置本地 FFmpeg"))
		return
	}
	if a.manager.Snapshot().Busy {
		failure(w, errors.New("请先停止任务"))
		return
	}
	a.mu.Lock()
	if a.installation.Busy {
		a.mu.Unlock()
		failure(w, errors.New("安装正在进行"))
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.installCancel = cancel
	a.installation = InstallStatus{Busy: true, Phase: "准备下载"}
	a.mu.Unlock()
	go func() {
		defer cancel()
		p, e := installer.Install(ctx, q.Source, a.cache, func(s string, p float64) {
			a.mu.Lock()
			a.installation.Phase = s
			a.installation.Percent = p
			a.mu.Unlock()
		})
		if e == nil {
			engine, err := media.Select(p)
			e = err
			if err == nil {
				e = a.manager.SetEngine(engine)
				if e == nil {
					a.mu.Lock()
					a.settings.Engine = engine.FFmpeg
					a.mu.Unlock()
					a.save()
				}
			}
		}
		a.mu.Lock()
		a.installation.Busy = false
		a.installCancel = nil
		if e != nil {
			a.installation.Error = e.Error()
			a.installation.Phase = "安装失败"
		} else {
			a.installation.Phase = "引擎已就绪"
		}
		a.mu.Unlock()
	}()
	reply(w, 200, map[string]bool{"started": true})
}
func csvSafe(s string) string {
	if s != "" && strings.ContainsAny(string(s[0]), "=+-@\t\r") {
		return "'" + s
	}
	return s
}
func (a *App) export(w http.ResponseWriter, r *http.Request) {
	s := a.manager.Snapshot()
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=PrismTranscode-results.csv")
		w.Write([]byte{0xef, 0xbb, 0xbf})
		c := csv.NewWriter(w)
		c.Write([]string{"输入文件", "识别格式", "输出格式", "状态", "输出路径", "处理方式", "SHA256", "验证", "错误"})
		for _, j := range s.Jobs {
			out, mode, sha, verify := "", "", "", ""
			if j.Result != nil {
				out = j.Result.Output
				mode = j.Result.Mode
				sha = j.Result.SHA256
				verify = j.Result.Verification
			}
			row := []string{j.Input, j.Info.Container, j.Options.Target, j.State, out, mode, sha, verify, j.Error}
			for n := range row {
				row[n] = csvSafe(row[n])
			}
			c.Write(row)
		}
		c.Flush()
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=PrismTranscode-results.json")
	reply(w, 200, s)
}
func main() {
	noBrowser := flag.Bool("no-browser", false, "Do not launch a browser")
	data := flag.String("data-dir", "", "Settings and queue directory")
	engine := flag.String("ffmpeg", "", "FFmpeg executable or bin directory")
	listen := flag.String("listen", "127.0.0.1:0", "Loopback address only")
	flag.Parse()
	host, _, err := net.SplitHostPort(*listen)
	if err != nil || host != "127.0.0.1" {
		fmt.Fprintln(os.Stderr, "--listen 只允许 127.0.0.1")
		os.Exit(1)
	}
	if *data == "" {
		base, e := os.UserConfigDir()
		if e != nil {
			base = os.TempDir()
		}
		*data = filepath.Join(base, "PrismTranscode")
	}
	if err = os.MkdirAll(*data, 0700); err != nil {
		platform.Alert(err.Error())
		os.Exit(1)
	}
	release, err := instance.Acquire(*data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if !*noBrowser {
			platform.Alert(err.Error())
		}
		os.Exit(1)
	}
	defer release()
	os.MkdirAll(filepath.Join(*data, "imports"), 0700)
	logFile, e := os.OpenFile(filepath.Join(*data, "application.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	session, err := os.MkdirTemp(*data, "session-")
	if err != nil {
		platform.Alert(err.Error())
		os.Exit(1)
	}
	defer os.RemoveAll(session)
	exe, _ := os.Executable()
	appDir := filepath.Dir(exe)
	a := &App{cache: *data, session: session, appDir: appDir, quit: make(chan struct{}), lastSeen: time.Now()}
	b, _ := os.ReadFile(filepath.Join(*data, "settings.json"))
	_ = json.Unmarshal(b, &a.settings)
	if *engine != "" {
		a.settings.Engine = *engine
	}
	eng, _ := media.Discover(a.settings.Engine, appDir)
	a.manager = jobs.New(eng, session, filepath.Join(*data, "queue.json"))
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		platform.Alert(err.Error())
		os.Exit(1)
	}
	a.host = listener.Addr().String()
	secret := make([]byte, 32)
	if _, err = rand.Read(secret); err != nil {
		panic(err)
	}
	a.token = hex.EncodeToString(secret)
	server := &http.Server{Handler: a.routes(), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 64 << 10}
	url := "http://" + a.host + "/app/" + a.token + "/"
	fmt.Println(url)
	_ = os.WriteFile(filepath.Join(*data, "last-session-url.txt"), []byte(url), 0600)
	go func() {
		if e := server.Serve(listener); e != nil && e != http.ErrServerClosed {
			log.Println(e)
			a.shutdown()
		}
	}()
	if !*noBrowser {
		if e = platform.OpenURL(url); e != nil {
			platform.Alert("请在浏览器打开：\n" + url)
		}
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signals:
		a.shutdown()
	case <-a.quit:
	}
	a.manager.Stop()
	a.manager.Wait()
	a.manager.Save()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
