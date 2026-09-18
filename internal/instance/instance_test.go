package instance

import "testing"

func TestExclusiveDirectoryLockAndRelease(t *testing.T) {
	dir := t.TempDir()
	release, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if second, err := Acquire(dir); err == nil {
		second()
		t.Fatal("second instance accepted")
	}
	other, err := Acquire(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other()
	release()
	release() // Cleanup is idempotent.
	again, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	again()
}
