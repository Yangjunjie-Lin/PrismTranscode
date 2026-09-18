//go:build !windows

package platform

import "errors"

func BrowseFiles() ([]string, error) {
	return nil, errors.New("此平台请使用路径导入或拖放文件")
}
