//go:build windows

package platform

import (
	"fmt"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

// BrowseFiles returns paths directly, avoiding browser uploads for large videos.
func BrowseFiles() ([]string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	buf := make([]uint16, 65536)
	filter := syscall.StringToUTF16("Media files / all files")
	filter = append(filter, syscall.StringToUTF16("*.*")...)
	filter = append(filter, 0)
	o := openFilename{File: &buf[0], MaxFile: uint32(len(buf)), Filter: &filter[0], FilterIndex: 1, Title: ptr("选择音频、视频、图片或字幕（可多选）"), Flags: 0x80000 | 0x1000 | 0x800 | 0x200 | 0x4 | 0x8}
	o.Size = uint32(unsafe.Sizeof(o))
	r, _, _ := commdlg.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&o)))
	if r == 0 {
		c, _, _ := commdlg.NewProc("CommDlgExtendedError").Call()
		if c != 0 {
			return nil, fmt.Errorf("文件选择器错误：0x%x", c)
		}
		return nil, nil
	}
	parts := []string{}
	start := 0
	for i, v := range buf {
		if v == 0 {
			if i == start {
				break
			}
			parts = append(parts, syscall.UTF16ToString(buf[start:i]))
			start = i + 1
		}
	}
	if len(parts) <= 1 {
		return parts, nil
	}
	out := make([]string, 0, len(parts)-1)
	for _, n := range parts[1:] {
		out = append(out, filepath.Join(parts[0], n))
	}
	return out, nil
}
