//go:build windows

package platform

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"
)

var shell = syscall.NewLazyDLL("shell32.dll")
var ole = syscall.NewLazyDLL("ole32.dll")
var commdlg = syscall.NewLazyDLL("comdlg32.dll")
var user = syscall.NewLazyDLL("user32.dll")

func ptr(s string) *uint16      { p, _ := syscall.UTF16PtrFromString(s); return p }
func HideConsole(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
func OpenURL(url string) error {
	r, _, _ := shell.NewProc("ShellExecuteW").Call(0, uintptr(unsafe.Pointer(ptr("open"))), uintptr(unsafe.Pointer(ptr(url))), 0, 0, 1)
	if r <= 32 {
		return fmt.Errorf("无法打开界面（Windows 错误 %d）", r)
	}
	return nil
}
func OpenFolder(path string) error { return OpenURL(path) }
func Alert(s string) {
	user.NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(ptr(s))), uintptr(unsafe.Pointer(ptr("流光转码 · PrismTranscode"))), 0x10)
}

type browseInfo struct {
	Owner, Root     uintptr
	Display, Title  *uint16
	Flags           uint32
	Callback, Param uintptr
	Image           int32
}

func BrowseFolder() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := ole.NewProc("CoInitializeEx").Call(0, 2)
	if uint32(hr) > 1 {
		return "", fmt.Errorf("无法打开文件夹选择器：0x%x", hr)
	}
	defer ole.NewProc("CoUninitialize").Call()
	display := make([]uint16, 32768)
	bi := browseInfo{Display: &display[0], Title: ptr("选择文件夹"), Flags: 0x51}
	pidl, _, _ := shell.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil
	}
	defer ole.NewProc("CoTaskMemFree").Call(pidl)
	buf := make([]uint16, 32768)
	r, _, _ := shell.NewProc("SHGetPathFromIDListW").Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	if r == 0 {
		return "", errors.New("请选择本地文件夹")
	}
	return syscall.UTF16ToString(buf), nil
}

type openFilename struct {
	Size                         uint32
	Owner, Instance              uintptr
	Filter, CustomFilter         *uint16
	MaxCustomFilter, FilterIndex uint32
	File                         *uint16
	MaxFile                      uint32
	FileTitle                    *uint16
	MaxFileTitle                 uint32
	InitialDir, Title            *uint16
	Flags                        uint32
	FileOffset, FileExtension    uint16
	DefaultExt                   *uint16
	CustData, Hook               uintptr
	TemplateName                 *uint16
	Reserved                     uintptr
	ReservedInt, FlagsEx         uint32
}

func BrowseFFmpeg() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	buf := make([]uint16, 32768)
	filter := syscall.StringToUTF16("FFmpeg executable")
	filter = append(filter, syscall.StringToUTF16("ffmpeg.exe")...)
	filter = append(filter, 0)
	o := openFilename{File: &buf[0], MaxFile: uint32(len(buf)), Filter: &filter[0], FilterIndex: 1, Title: ptr("选择可信来源的 ffmpeg.exe"), Flags: 0x80000 | 0x1000 | 0x800 | 0x4 | 0x8}
	o.Size = uint32(unsafe.Sizeof(o))
	r, _, _ := commdlg.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&o)))
	if r == 0 {
		code, _, _ := commdlg.NewProc("CommDlgExtendedError").Call()
		if code != 0 {
			return "", fmt.Errorf("文件选择器错误：0x%x", code)
		}
		return "", nil
	}
	return syscall.UTF16ToString(buf), nil
}
