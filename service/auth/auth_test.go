package auth

import (
	"testing"
	"time"
)

var l = NewUserLock(2 * time.Second)

func TestLock(t *testing.T) {
	t.Log("Locking john")
	l.Lock("john")
	if !l.IsLocked("john") {
		t.Errorf("John should be locked")
	}
	t.Log("Locking jack")
	l.Lock("jack")
	t.Log("Sleep for 2 seconds and check...")
	time.Sleep(2 * time.Second)
	if l.IsLocked("jack") {
		t.Errorf("After 2 seconds, jack shouldn't be locked")
	}
	if l.IsLocked("daniel") {
		t.Errorf("daniel has never been locked, he should not be locked")
	}
}
