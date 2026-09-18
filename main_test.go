package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"prismtranscode/internal/jobs"
	"strings"
	"testing"
	"time"
)

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	for _, body := range []string{`{} {}`, `{} garbage`, `{"unknown":1}`} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		if decode(httptest.NewRecorder(), r, &struct{}{}) == nil {
			t.Fatalf("accepted %q", body)
		}
	}
}

func TestInvalidUploadRemovesStagedFiles(t *testing.T) {
	a := testApp(t)
	imports := filepath.Join(a.cache, "imports")
	if err := os.Mkdir(imports, 0700); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("files", "test.wav")
	part.Write([]byte("synthetic invalid data"))
	writer.WriteField("options", `{"target":"not-a-format"}`)
	writer.Close()
	r := httptest.NewRequest("POST", "http://"+a.host+"/api/upload", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	entries, err := os.ReadDir(imports)
	if err != nil || len(entries) != 0 {
		t.Fatalf("leaked imports: %v %v", entries, err)
	}
}

func testApp(t *testing.T) *App {
	d := t.TempDir()
	return &App{token: "secret", host: "127.0.0.1:9876", cache: d, session: d, manager: jobs.New(nil, d, ""), quit: make(chan struct{}), lastSeen: time.Now()}
}
func TestAuthRejectsCrossOrigin(t *testing.T) {
	a := testApp(t)
	for _, v := range []struct{ host, origin, token string }{{"evil.test:9876", "", "secret"}, {a.host, "https://evil.test", "secret"}, {a.host, "", "bad"}} {
		r := httptest.NewRequest("GET", "http://"+a.host+"/api/queue", nil)
		r.Host = v.host
		r.Header.Set("Origin", v.origin)
		r.Header.Set("Authorization", "Bearer "+v.token)
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
}
func TestAuthenticatedRead(t *testing.T) {
	a := testApp(t)
	r := httptest.NewRequest("GET", "http://"+a.host+"/api/queue", nil)
	r.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	var s jobs.Snapshot
	if json.Unmarshal(w.Body.Bytes(), &s) != nil {
		t.Fatal("bad JSON")
	}
}
func TestMutationRequiresPOST(t *testing.T) {
	a := testApp(t)
	r := httptest.NewRequest("GET", "http://"+a.host+"/api/stop", nil)
	r.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
}
func TestCSVFormulaEscaping(t *testing.T) {
	for _, s := range []string{"=HYPERLINK(\"x\")", "+SUM(1,2)", "@bad", "-2"} {
		if !strings.HasPrefix(csvSafe(s), "'") {
			t.Fatal(s)
		}
	}
	if csvSafe("song.mp3") != "song.mp3" {
		t.Fatal("unnecessary escaping")
	}
}
