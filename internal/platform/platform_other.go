//go:build !windows

package platform

import (
	"errors"
	"os/exec"
	"runtime"
)

func HideConsole(cmd *exec.Cmd) {}
func BrowseFolder() (string, error) {
	return "", errors.New("此平台请直接在界面填写输出文件夹路径")
}
func BrowseFFmpeg() (string, error) {
	return "", errors.New("此平台请使用 --ffmpeg 参数指定组件")
}
func OpenURL(url string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", url).Start()
	}
	return exec.Command("xdg-open", url).Start()
}
func OpenFolder(path string) error { return OpenURL(path) }
func Alert(text string)            {}
