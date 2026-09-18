// Package instance protects each data directory with an OS-owned file lock.
// The OS releases the lock on process exit, including crashes. The lock file
// must not be unlinked, which would allow another process to lock a new inode.
package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func Acquire(dir string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(dir, "instance.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = lock(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("数据目录已被其他实例占用或无法锁定，请关闭另一个 PrismTranscode 后重试：%w", err)
	}
	var once sync.Once
	return func() { once.Do(func() { _ = f.Close() }) }, nil
}
